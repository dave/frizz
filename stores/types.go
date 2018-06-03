package stores

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"

	"strings"

	"path/filepath"

	"github.com/dave/flux"
	"github.com/dave/frizz/actions"
	"github.com/dave/services/srcimporter"
)

func NewTypesStore(a *App) *TypesStore {
	s := &TypesStore{
		app:   a,
		types: map[string]map[string]map[string]*types.Named{},
		data:  map[string]map[string]map[string]types.Object{},
	}
	return s
}

type TypesStore struct {
	app   *App
	types map[string]map[string]map[string]*types.Named // path->file->name
	data  map[string]map[string]map[string]types.Object // path->file->name (*types.Var or *types.Const)
}

func (s *TypesStore) Handle(payload *flux.Payload) bool {
	switch action := payload.Action.(type) {
	case *actions.SourceChanged:

		// Parse for types
		fset := token.NewFileSet()
		bctx := srcimporter.NewBuildContext(s.app.Source.Source(), s.app.Tags.Tags())
		fileNames := map[string]*ast.File{}
		files := []*ast.File{}

		s.app.Log("Scanning source...")

		for name, contents := range s.app.Source.Source()[s.app.Source.Path()] {
			if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}
			match, err := bctx.MatchFile(filepath.Join("gopath", "src", s.app.Source.Path()), name)
			if err != nil {
				s.app.Fail(err)
				return true
			}
			if !match {
				continue
			}

			f, err := parser.ParseFile(fset, name, []byte(contents), 0)
			if err != nil {
				s.app.Fail(err)
				return true
			}
			fileNames[name] = f
			files = append(files, f)
		}
		/*
			imports := map[string]bool{
				s.app.Source.Path(): true,
			}
			for _, f := range files {
				for _, i := range f.Imports {
					s, _ := strconv.Unquote(i.Path.Value)
					imports[s] = true
				}
			}
		*/
		packages := map[string]*types.Package{}
		importer := srcimporter.New(bctx, fset, packages)
		importer.Filter = func(path string) bool {
			//return imports[path]
			return true
		}
		importer.Callback = func(path string) {
			s.app.Log(path)
		}
		tc := &types.Config{
			IgnoreFuncBodies: true,
			Importer:         importer,
			Error: func(err error) {
				// ignore errors
			},
		}
		ti := &types.Info{
			//Types: map[ast.Expr]types.TypeAndValue{},
			Defs: map[*ast.Ident]types.Object{},
		}

		s.app.Log("Checking types...")

		/*
			n := strings.TrimSuffix(files[0].Name.Name, "_test")
			p := types.NewPackage(s.app.Source.Path(), n)
			c := types.NewChecker(tc, fset, p, ti)
			for n, f := range fileNames {
				fmt.Println(n)
				c.Files([]*ast.File{f})
			}
		*/

		p, _ := tc.Check(s.app.Source.Path(), fset, files, ti) // ignore errors
		packages[s.app.Source.Path()] = p

		var typesCount, dataCount int

		for _, p := range packages {
			if p == nil {
				continue
			}
			for _, name := range p.Scope().Names() {
				v := p.Scope().Lookup(name)
				if v == nil {
					continue
				}
				if !v.Exported() {
					continue
				}
				path, file, name := v.Pkg().Path(), fset.File(v.Pos()).Name(), v.Name()
				switch v := v.(type) {
				case *types.TypeName:
					if s.types[path] == nil {
						s.types[path] = map[string]map[string]*types.Named{}
					}
					if s.types[path][file] == nil {
						s.types[path][file] = map[string]*types.Named{}
					}
					s.types[path][file][name] = v.Type().(*types.Named)
					typesCount++
				case *types.Var, *types.Const:
					if s.data[path] == nil {
						s.data[path] = map[string]map[string]types.Object{}
					}
					if s.data[path][file] == nil {
						s.data[path][file] = map[string]types.Object{}
					}
					s.data[path][file][name] = v
					dataCount++
				}
			}
		}
		s.app.LogHidef("Found %d types and %d variables", typesCount, dataCount)

	default:
		_ = action
	}
	return true
}
