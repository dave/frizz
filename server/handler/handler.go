package handler

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"strings"
	"time"

	pathpkg "path"

	"errors"

	"context"

	"sync"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/storage"
	"github.com/dave/frizz/config"
	"github.com/dave/frizz/server/assets"
	"github.com/dave/frizz/server/messages"
	"github.com/dave/jsgo/server/queue"
	"github.com/dave/jsgo/server/store"
	"github.com/dave/patsy"
	"github.com/dave/patsy/vos"
	"github.com/dave/services"
	"github.com/dave/services/cachefileserver"
	"github.com/dave/services/gcsdatabase"
	"github.com/dave/services/gcsfileserver"
	"github.com/dave/services/getter/cache"
	"github.com/dave/services/gitfetcher"
	"github.com/dave/services/localdatabase"
	"github.com/dave/services/localfileserver"
	"github.com/dave/services/localresolverfetcher"
	"github.com/gorilla/websocket"
	"github.com/shurcooL/httpgzip"
	"gopkg.in/src-d/go-billy.v4"
)

func New(shutdown chan struct{}) *Handler {
	var c *cache.Cache
	var fileserver services.Fileserver
	var database services.Database
	if config.LOCAL {
		fileserver = localfileserver.New(config.LocalFileserverTempDir, config.Sites)
		database = localdatabase.New(config.LocalFileserverTempDir)
		fetcherResolver := localresolverfetcher.New()
		c = cache.New(
			database,
			fetcherResolver,
			fetcherResolver,
			config.HintsKind,
		)
	} else {
		storageClient, err := storage.NewClient(context.Background())
		if err != nil {
			panic(err)
		}

		datastoreClient, err := datastore.NewClient(context.Background(), config.ProjectID)
		if err != nil {
			panic(err)
		}

		database = gcsdatabase.New(datastoreClient)
		fileserver = gcsfileserver.New(storageClient, config.Buckets)
		c = cache.New(
			database,
			gitfetcher.New(
				cachefileserver.New(1024*1024*1042, 100*1024*1024),
				fileserver,
				config.GitSaveTimeout,
				config.GitCloneTimeout,
				config.GitMaxObjects,
				config.GitBucket,
			),
			nil,
			config.HintsKind,
		)
	}
	h := &Handler{
		Cache:      c,
		mux:        http.NewServeMux(),
		shutdown:   shutdown,
		Queue:      queue.New(config.MaxConcurrentRequests, config.MaxQueue),
		Waitgroup:  &sync.WaitGroup{},
		Fileserver: fileserver,
		Database:   database,
	}
	h.mux.HandleFunc("/", h.Page)
	h.mux.HandleFunc("/_script.js", h.Script)
	h.mux.HandleFunc("/_script.js.map", h.Script)
	//h.mux.HandleFunc("/_info/", h.InfoHandler)
	h.mux.HandleFunc("/_ws/", h.Socket)
	h.mux.HandleFunc("/favicon.ico", h.Icon)
	h.mux.HandleFunc("/_ah/health", h.HealthCheck)
	if config.LOCAL {
		dir, err := patsy.Dir(vos.Os(), "github.com/dave/frizz/server/assets/static/")
		if err != nil {
			panic(err)
		}
		h.mux.Handle("/_local/", http.FileServer(http.Dir(dir)))
	}
	return h
}

type Handler struct {
	Cache      *cache.Cache
	Fileserver services.Fileserver
	Database   services.Database
	Waitgroup  *sync.WaitGroup
	Queue      *queue.Queue
	mux        *http.ServeMux
	shutdown   chan struct{}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *Handler) sendAndStoreError(ctx context.Context, send func(messages.Message), path string, err error, req *http.Request) {
	h.storeError(ctx, err, req)
	h.sendError(send, err)
}

func (h *Handler) sendError(send func(messages.Message), err error) {
	send(messages.Error{
		Message: err.Error(),
	})
}

func (h *Handler) storeError(ctx context.Context, err error, req *http.Request) {

	fmt.Println(err)

	if err == queue.TooManyItemsQueued {
		// If the server is getting flooded by a DOS, this will prevent database flooding
		return
	}

	// ignore errors when logging an error
	store.StoreError(ctx, h.Database, store.Error{
		Time:  time.Now(),
		Error: err.Error(),
		Ip:    req.Header.Get("X-Forwarded-For"),
	})

}

func (h *Handler) Icon(w http.ResponseWriter, req *http.Request) {
	if err := serveStatic(req.URL.Path, w, req, "image/x-icon"); err != nil {
		http.Error(w, "error serving static file", 500)
	}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func serveStatic(name string, w http.ResponseWriter, req *http.Request, mimeType string) error {
	var file billy.File
	var err error
	file, err = assets.Assets.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, req)
			return nil
		}
		http.Error(w, fmt.Sprintf("error opening %s", name), 500)
		return nil
	}
	defer file.Close()

	w.Header().Set("Cache-Control", "public,max-age=31536000,immutable")
	if mimeType == "" {
		w.Header().Set("Content-Type", mime.TypeByExtension(pathpkg.Ext(req.URL.Path)))
	} else {
		w.Header().Set("Content-Type", mimeType)
	}

	_, noCompress := file.(httpgzip.NotWorthGzipCompressing)
	gzb, isGzb := file.(httpgzip.GzipByter)

	if isGzb && !noCompress && strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		if err := writeWithTimeout(w, gzb.GzipBytes()); err != nil {
			http.Error(w, fmt.Sprintf("error streaming gzipped %s", name), 500)
			return err
		}
	} else {
		if err := streamWithTimeout(w, file); err != nil {
			http.Error(w, fmt.Sprintf("error streaming %s", name), 500)
			return err
		}
	}
	return nil

}

func streamWithTimeout(w io.Writer, r io.Reader) error {
	c := make(chan error, 1)
	go func() {
		_, err := io.Copy(w, r)
		c <- err
	}()
	select {
	case err := <-c:
		if err != nil {
			return err
		}
		return nil
	case <-time.After(config.WriteTimeout):
		return errors.New("timeout")
	}
}

func writeWithTimeout(w io.Writer, b []byte) error {
	return streamWithTimeout(w, bytes.NewBuffer(b))
}
