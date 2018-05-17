package assets

import (
	"archive/zip"
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"

	"bytes"

	"io/ioutil"
	"path/filepath"

	"encoding/gob"

	"github.com/dave/frizz/config"
	"github.com/dave/patsy"
	"github.com/dave/patsy/vos"
	"github.com/gopherjs/gopherjs/compiler"
	billy "gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
)

var Assets = memfs.New()
var Archives map[string]map[bool]*compiler.Archive

func Init() {
	if err := loadAssets(Assets); err != nil {
		panic(err)
	}
}

func loadAssets(fs billy.Filesystem) error {

	var buf *bytes.Buffer

	if config.DEV {
		dir, err := patsy.Dir(vos.Os(), "github.com/dave/frizz/server/assets")
		if err != nil {
			return err
		}
		b, err := ioutil.ReadFile(filepath.Join(dir, config.AssetsFilename))
		if err != nil {
			return err
		}
		buf = bytes.NewBuffer(b)
	} else {
		ctx := context.Background()
		client, err := storage.NewClient(ctx)
		if err != nil {
			return err
		}
		defer client.Close()
		gcsReader, err := client.Bucket(config.PkgBucket).Object(config.AssetsFilename).NewReader(ctx)
		if err != nil {
			return err
		}
		fmt.Println("Getting assets from GCS...")
		buf = new(bytes.Buffer)
		if _, err := io.Copy(buf, gcsReader); err != nil {
			return err
		}
	}

	reader := bytes.NewReader(buf.Bytes())
	fmt.Println("Unzipping assets...")
	r, err := zip.NewReader(reader, int64(buf.Len()))
	if err != nil {
		return err
	}

	for _, zipFile := range r.File {
		if err := loadAssetFile(zipFile, fs); err != nil {
			return err
		}
	}
	return nil
}

func loadAssetFile(zipFile *zip.File, fs billy.Filesystem) error {
	if zipFile.Name == "/archives.gob" {
		// special case for archives gob file
		decompressed, err := zipFile.Open()
		if err != nil {
			return err
		}
		defer decompressed.Close()
		if err := gob.NewDecoder(decompressed).Decode(&Archives); err != nil {
			return err
		}
		return nil
	}
	fsFile, err := fs.Create(zipFile.Name)
	if err != nil {
		return err
	}
	defer fsFile.Close()
	decompressed, err := zipFile.Open()
	if err != nil {
		return err
	}
	defer decompressed.Close()
	if _, err := io.Copy(fsFile, decompressed); err != nil {
		return err
	}
	return nil
}
