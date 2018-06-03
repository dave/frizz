package stores

import (
	"errors"
	"net/http"
	"sync"

	"honnef.co/go/js/dom"

	"fmt"

	"encoding/json"

	"strings"

	"github.com/dave/flux"
	"github.com/dave/frizz/actions"
	"github.com/dave/frizz/models"
	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server/frizz/messages"
)

func NewSourceStore(a *App) *SourceStore {
	s := &SourceStore{
		app:    a,
		cache:  map[string]CacheItem{},
		source: map[string]map[string]string{},
	}
	return s
}

type SourceStore struct {
	app *App

	path string
	tags []string // tags at last source download

	// cache (path -> item) of archives
	cache map[string]CacheItem

	source map[string]map[string]string

	// index (path -> item) of the previously received update
	index messages.SourceIndex

	wait sync.WaitGroup
}

type CacheItem struct {
	Path  string
	Hash  string
	Files map[string]string
}

func (s *SourceStore) Source() map[string]map[string]string {
	return s.source
}

func (s *SourceStore) Cache() map[string]CacheItem {
	return s.cache
}

func (s *SourceStore) Path() string {
	return s.path
}

func (s *SourceStore) CacheStrings() map[string]string {
	hashes := map[string]string{}
	for path, item := range s.cache {
		hashes[path] = item.Hash
	}
	return hashes
}

// Fresh is true if current cache matches the previously downloaded archives
func (s *SourceStore) Fresh() bool {
	// if index is nil, either the page has just loaded or we're in the middle of an update
	if s.index == nil {
		return false
	}

	// TODO: tags

	// first check that all indexed packages are in the cache at the right versions. This would fail
	// if there was an error while downloading one of the archive files.
	for path, item := range s.index {
		cached, ok := s.cache[path]
		if !ok {
			return false
		}
		if cached.Hash != item.Hash {
			return false
		}
	}

	return true
}

func (s *SourceStore) Handle(payload *flux.Payload) bool {
	switch action := payload.Action.(type) {
	case *actions.Load:
		location := strings.Trim(dom.GetWindow().Location().Pathname, "/")
		s.app.Dispatch(&actions.GetSourceStart{Path: location})
	case *actions.GetSourceStart:
		s.app.Log("getting source")
		s.app.Dispatch(&actions.Dial{
			Open: func() flux.ActionInterface { return &actions.GetSourceOpen{Path: action.Path} },
			Message: func(m interface{}) flux.ActionInterface {
				return &actions.GetSourceMessage{Path: action.Path, Message: m}
			},
			Close: func() flux.ActionInterface { return &actions.GetSourceClose{} },
		})
		payload.Notify()
	case *actions.GetSourceOpen:
		message := messages.GetSource{
			Path:  action.Path,
			Tags:  s.app.Tags.Tags(),
			Cache: s.CacheStrings(),
		}
		s.app.Dispatch(&actions.Send{
			Message: message,
		})
	case *actions.GetSourceMessage:
		switch message := action.Message.(type) {
		case messages.Source:
			s.wait.Add(1)
			go func() {
				defer s.wait.Done()
				c := CacheItem{
					Path: message.Path,
					Hash: message.Hash,
				}
				var getwait sync.WaitGroup
				getwait.Add(1)
				go func() {
					defer getwait.Done()
					resp, err := http.Get(fmt.Sprintf("%s://%s/%s.%s.json", config.Protocol[config.Pkg], config.Host[config.Pkg], message.Path, message.Hash))
					if err != nil {
						s.app.Fail(err)
						return
					}
					var sp models.SourcePack
					if err := json.NewDecoder(resp.Body).Decode(&sp); err != nil {
						s.app.Fail(err)
						return
					}
					c.Files = sp.Files
				}()
				getwait.Wait()
				s.cache[message.Path] = c
				s.source[message.Path] = c.Files
				s.app.Log(message.Path)
			}()
			return true
		case messages.SourceIndex:
			s.path = action.Path
			s.index = message
		}
	case *actions.GetSourceClose:
		s.wait.Wait()
		if !s.Fresh() {
			s.app.Fail(errors.New("websocket closed but source not fully updated"))
			return true
		}
		var downloaded, unchanged int
		for _, v := range s.index {
			if v.Unchanged {
				unchanged++
			} else {
				downloaded++
			}
		}
		if downloaded == 0 && unchanged == 0 {
			s.app.Log()
		} else if downloaded > 0 && unchanged > 0 {
			s.app.LogHidef("%d downloaded, %d unchanged", downloaded, unchanged)
		} else if downloaded > 0 {
			s.app.LogHidef("%d downloaded", downloaded)
		} else if unchanged > 0 {
			s.app.LogHidef("%d unchanged", unchanged)
		}
		s.app.Dispatch(&actions.SourceChanged{})
	}
	return true
}
