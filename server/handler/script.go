package handler

import (
	"net/http"

	"bytes"
	"strings"

	"io"

	"errors"

	"github.com/dave/jsgo/config"
	gbuild "github.com/gopherjs/gopherjs/build"
	"github.com/gopherjs/gopherjs/compiler"
	"github.com/gopherjs/gopherjs/compiler/prelude"
	"github.com/neelance/sourcemap"
)

func (h *Handler) Script(w http.ResponseWriter, req *http.Request) {
	if err := h.handleScript(w, req); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func (h *Handler) handleScript(w http.ResponseWriter, req *http.Request) error {
	if !config.DEV {
		return errors.New("script only available in dev mode")
	}

	isPkg := strings.HasSuffix(req.URL.Path, ".js")
	isMap := strings.HasSuffix(req.URL.Path, ".js.map")

	options := &gbuild.Options{
		Quiet:          true,
		CreateMapFile:  true,
		MapToLocalDisk: true,
		BuildTags:      []string{"jsgo", "dev"},
	}

	if config.LOCAL {
		options.BuildTags = append(options.BuildTags, "local")
	}

	// If we're going to be serving our special files, make sure there's a Go command in this folder.
	s := gbuild.NewSession(options)
	pkg, err := gbuild.Import("github.com/dave/frizz/ed", 0, s.InstallSuffix(), options.BuildTags)
	if err != nil {
		return err
	}

	switch {
	case isPkg:
		buf := new(bytes.Buffer)
		err := func() error {
			archive, err := s.BuildPackage(pkg)
			if err != nil {
				return err
			}

			sourceMapFilter := &compiler.SourceMapFilter{Writer: buf}
			m := &sourcemap.Map{File: "_script.js"}
			sourceMapFilter.MappingCallback = gbuild.NewMappingCallback(m, options.GOROOT, options.GOPATH, options.MapToLocalDisk)

			deps, err := compiler.ImportDependencies(archive, s.BuildImportPath)
			if err != nil {
				return err
			}
			if err := WriteProgramCode(deps, sourceMapFilter); err != nil {
				return err
			}

			mapBuf := new(bytes.Buffer)
			m.WriteTo(mapBuf)
			buf.WriteString("//# sourceMappingURL=_script.js.map\n")
			lastMap = mapBuf.Bytes()
			return nil
		}()
		if err != nil {
			return err
		}
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Content-Type", "application/javascript")
		if _, err := io.Copy(w, buf); err != nil {
			return err
		}

	case isMap:
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Content-Type", "application/javascript")
		if _, err := io.Copy(w, bytes.NewBuffer(lastMap)); err != nil {
			return err
		}
	}
	return nil
}

var lastMap []byte

func WriteProgramCode(pkgs []*compiler.Archive, w *compiler.SourceMapFilter) error {
	mainPkg := pkgs[len(pkgs)-1]
	minify := mainPkg.Minified

	if _, err := w.Write([]byte("\"use strict\";\n")); err != nil {
		return err
	}
	//if _, err := w.Write([]byte("\"use strict\";\n(function() {\n\n")); err != nil {
	//	return err
	//}
	preludeJS := prelude.Prelude
	if minify {
		preludeJS = prelude.Minified
	}
	if _, err := io.WriteString(w, preludeJS); err != nil {
		return err
	}
	if _, err := w.Write([]byte("\n")); err != nil {
		return err
	}

	// write packages
	for _, pkg := range pkgs {
		dceSelection := make(map[*compiler.Decl]struct{})
		for _, d := range pkg.Declarations {
			dceSelection[d] = struct{}{}
		}
		if err := compiler.WritePkgCode(pkg, dceSelection, minify, w); err != nil {
			return err
		}
	}

	if _, err := w.Write([]byte("$synthesizeMethods();\nvar $mainPkg = $packages[\"" + string(mainPkg.ImportPath) + "\"];\n$packages[\"runtime\"].$init();\n$go($mainPkg.$init, []);\n$flushConsole();\n")); err != nil {
		return err
	}

	return nil
}
