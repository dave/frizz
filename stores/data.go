package stores

import (
	"go/ast"
	"strconv"

	"go/parser"
	"go/token"

	"github.com/dave/flux"
	"github.com/dave/frizz/actions"
	"github.com/dave/jsgo/server/frizz/gotypes"
)

func NewDataStore(a *App) *DataStore {
	s := &DataStore{
		app:     a,
		data:    map[gotypes.Object]ast.Expr{},
		imports: map[string]map[string]map[string]string{},
	}
	return s
}

type DataStore struct {
	app     *App
	data    map[gotypes.Object]ast.Expr
	imports map[string]map[string]map[string]string // package -> file -> alias -> import path
}

func (s *DataStore) Expr(ob gotypes.Object) ast.Expr {
	return s.data[ob]
}

func (s *DataStore) Import(path, file, alias string) string {
	return s.imports[path][file][alias]
}

func (s *DataStore) Handle(payload *flux.Payload) bool {
	switch action := payload.Action.(type) {
	case *actions.GetPackageClose:
		payload.Wait(s.app.Packages)
		s.app.Log("scanning data")
		fset := token.NewFileSet()
		s.imports[action.Path] = map[string]map[string]string{}
		for fname, contents := range s.app.Packages.Source()[action.Path] {

			objects := s.app.Packages.ObjectsInFile(action.Path, fname)
			decls := map[string]gotypes.Object{}
			for name, ob := range objects {
				switch ob.(type) {
				case *gotypes.Var, *gotypes.Const:
					decls[name] = ob
				}
			}
			if len(decls) == 0 {
				// if no vars/consts in file, continue
				continue
			}

			f, err := parser.ParseFile(fset, fname, contents, parser.ParseComments)
			if err != nil {
				s.app.Fail(err)
				return true
			}

			s.imports[action.Path][fname] = map[string]string{}
			for _, is := range f.Imports {
				if is.Name != nil && (is.Name.Name == "_" || is.Name.Name == ".") {
					continue
				}
				path, err := strconv.Unquote(is.Path.Value)
				if err != nil {
					s.app.Fail(err)
				}
				var name string
				if is.Name != nil {
					name = is.Name.Name
				} else {
					name = s.app.Packages.PackageName(path)
				}
				s.imports[action.Path][fname][name] = path
			}

			ast.Inspect(f, func(n ast.Node) bool {
				if n == nil {
					return false
				}
				switch n := n.(type) {
				case *ast.GenDecl:
					if n.Tok != token.VAR && n.Tok != token.CONST {
						return false
					}
					for _, spec := range n.Specs { // more than 1 spec if var ( ... )
						spec := spec.(*ast.ValueSpec) // var and const always have *ast.ValueSpec specs
						for i := 0; i < len(spec.Names); i++ {
							name := spec.Names[i].Name
							// look up name
							ob, ok := decls[name]
							if !ok {
								continue
							}
							if len(spec.Values) > 0 {
								// if just `var Foo string`, spec.Values == nil
								s.data[ob] = spec.Values[i]
							}
							s.app.Log(name)
							/*
								{
									buf := &bytes.Buffer{}
									if err := format.Node(buf, fset, value); err != nil {
										s.app.Fail(err)
										return true
									}
									fmt.Println(name, buf.String())
								}
							*/
						}
					}
				}
				return true
			})
		}
		s.app.LogHidef("done")
		payload.Notify()
	default:
		_ = action
	}
	return true
}
