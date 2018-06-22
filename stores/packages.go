package stores

import (
	"errors"
	"net/http"
	"sync"

	"honnef.co/go/js/dom"

	"fmt"

	"encoding/json"

	"strings"

	"encoding/gob"

	"sort"

	"go/ast"
	"go/token"

	"github.com/dave/flux"
	"github.com/dave/frizz/actions"
	"github.com/dave/frizz/models"
	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server/frizz/gotypes"
	"github.com/dave/jsgo/server/frizz/messages"
)

func init() {
	gotypes.RegisterTypesGob()
}

func NewPackageStore(a *App) *PackageStore {
	s := &PackageStore{
		app:           a,
		sourceHashes:  map[string]string{},
		objectsHashes: map[string]string{},
		source:        map[string]map[string]string{},
		objects:       map[string]map[string]map[string]gotypes.Object{},
		types:         map[string]map[string]gotypes.Type{},
		packageNames:  map[string]string{},
	}
	return s
}

type PackageStore struct {
	app *App

	path string
	tags []string // tags at last package download

	sourceHashes  map[string]string
	objectsHashes map[string]string

	source  map[string]map[string]string                    // package -> file -> contents
	objects map[string]map[string]map[string]gotypes.Object // package -> file -> name -> object
	types   map[string]map[string]gotypes.Type              // package -> name -> type

	packageNames map[string]string

	index *messages.PackageIndex

	mutex sync.Mutex
	wait  sync.WaitGroup
}

func (s *PackageStore) ObjectsInFile(path, file string) map[string]gotypes.Object {
	files, ok := s.objects[path]
	if !ok {
		return map[string]gotypes.Object{}
	}
	objects, ok := files[file]
	if !ok {
		return map[string]gotypes.Object{}
	}
	return objects
}

// ResolveType resolves Reference, Named, Pointer or Interface types to their underlying types. Tries
// to return one of: Basic, Array, Slice, Struct, Tuple, Signature, Map, or Chan. If the interface cannot
// be resolved or data == nil, it may return *Interface. If the reference can't be resolved, it may
// return *Reference.
func (s *PackageStore) ResolveType(t gotypes.Type, path, file string, data ast.Expr) gotypes.Type {
	var depth int
	for {
		if depth > MaxResolveTypeDepth {
			// sanity check for recursive types (shouldn't happen?)
			panic("past max depth in ResolveType")
		}
		depth++
		var resolved gotypes.Type
		switch t := t.(type) {
		case *gotypes.Reference:
			resolved = s.resolveReference(t)
		case *gotypes.Named:
			resolved = s.resolveNamed(t)
		case *gotypes.Pointer:
			resolved = s.resolvePointer(t)
		case *gotypes.Interface:
			resolved = s.resolveTypeFromExpr(path, file, data)
		default:
			return t
		}
		if resolved == nil {
			// if we don't successfully manage to resolve a type, return the previous type
			return t
		} else {
			// if we resolved the type, recurse
			t = resolved
		}
	}
}

const MaxResolveTypeDepth = 100

func (s *PackageStore) resolveNamed(t *gotypes.Named) gotypes.Type {
	return t.Type
}

func (s *PackageStore) resolvePointer(t *gotypes.Pointer) gotypes.Type {
	return t.Elem
}

func (s *PackageStore) resolveReference(t *gotypes.Reference) gotypes.Type {
	pkg, ok := s.types[t.Path]
	if !ok {
		return nil
	}
	return pkg[t.Name]
}

func (s *PackageStore) resolveTypeFromExpr(path, file string, e ast.Expr) gotypes.Type {
	if e == nil {
		return nil
	}
	switch e := e.(type) {
	case *ast.BadExpr:
	case *ast.Ident:
		// TODO: Search packages that are dot-imported?
		return s.resolveReference(&gotypes.Reference{Identifier: gotypes.Identifier{Path: path, Name: e.Name}})
	case *ast.Ellipsis:
	case *ast.BasicLit:
		switch e.Kind {
		case token.INT:
			return gotypes.Typ[gotypes.UntypedInt]
		case token.FLOAT:
			return gotypes.Typ[gotypes.UntypedFloat]
		case token.IMAG:
			return gotypes.Typ[gotypes.UntypedComplex]
		case token.CHAR:
			return gotypes.Typ[gotypes.UntypedRune]
		case token.STRING:
			return gotypes.Typ[gotypes.UntypedString]
		}
	case *ast.FuncLit:
	case *ast.CompositeLit:
		return s.resolveTypeFromExpr(path, file, e.Type)
	case *ast.ParenExpr:
	case *ast.SelectorExpr:
		if x, ok := e.X.(*ast.Ident); ok && x.Obj == nil {
			// if X is an ident and Obj == nil -> selector is of the form <path>.<Name>
			return s.resolveReference(&gotypes.Reference{
				Identifier: gotypes.Identifier{
					Path: s.app.Data.Import(path, file, x.Name),
					Name: e.Sel.Name,
				},
			})
		}
	case *ast.IndexExpr:
	case *ast.SliceExpr:
	case *ast.TypeAssertExpr:
	case *ast.CallExpr:
	case *ast.StarExpr:
	case *ast.UnaryExpr:
	case *ast.BinaryExpr:
	case *ast.KeyValueExpr:
	case *ast.ArrayType:
	case *ast.StructType:
	case *ast.FuncType:
	case *ast.InterfaceType:
	case *ast.MapType:
	case *ast.ChanType:
	}
	return nil
}

func (s *PackageStore) SortedObjectsInFile(path, file string) []gotypes.Object {
	var objects []gotypes.Object
	for _, o := range s.objects[path][file] {
		objects = append(objects, o)
	}
	// all in same file -> don't need to compare package
	sort.Slice(objects, func(i, j int) bool { return objects[i].Object().Name < objects[j].Object().Name })
	return objects
}

func (s *PackageStore) SortedSourceFiles(path string) []string {
	var files []string
	for f := range s.source[path] {
		files = append(files, f)
	}
	sort.Strings(files)
	return files
}

func (s *PackageStore) Source() map[string]map[string]string {
	return s.source
}

func (s *PackageStore) SortedSourcePackages() []string {
	var paths []string
	for p := range s.source {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	return paths
}

func (s *PackageStore) PackageName(path string) string {
	return s.packageNames[path]
}

func (s *PackageStore) DisplayPath(path string) string {
	parts := strings.Split(path, "/")
	guessed := parts[len(parts)-1]
	name := s.packageNames[path]
	suffix := ""
	if guessed != name && name != "" {
		suffix = " (" + name + ")"
	}
	return path + suffix
}

func (s *PackageStore) DisplayName(path string) string {
	if s.packageNames[path] != "" {
		return s.packageNames[path]
	}
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

func (s *PackageStore) Path() string {
	return s.path
}

func (s *PackageStore) SourceHashes() map[string]string {
	return s.sourceHashes
}

func (s *PackageStore) ObjectsHashes() map[string]string {
	return s.objectsHashes
}

// Fresh is true if current cache matches the previously downloaded archives
func (s *PackageStore) Fresh() bool {

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// if index is nil, either the page has just loaded or we're in the middle of an update
	if s.index == nil {
		return false
	}

	// TODO: tags

	// first check that all indexed packages are in the cache at the right versions.
	for path, item := range s.index.Source {
		cached, ok := s.sourceHashes[path]
		if !ok {
			return false
		}
		if cached != item.Hash {
			return false
		}
	}

	for path, item := range s.index.Objects {
		cached, ok := s.objectsHashes[path]
		if !ok {
			return false
		}
		if cached != item.Hash {
			return false
		}
	}

	return true
}

func (s *PackageStore) Handle(payload *flux.Payload) bool {
	switch action := payload.Action.(type) {
	case *actions.Load:
		location := strings.Trim(dom.GetWindow().Location().Pathname, "/")
		s.app.Dispatch(&actions.GetPackageStart{Path: location})
	case *actions.GetPackageStart:
		s.app.Log("getting package")
		s.app.Dispatch(&actions.Dial{
			Open: func() flux.ActionInterface {
				return &actions.GetPackageOpen{Path: action.Path}
			},
			Message: func(m interface{}) flux.ActionInterface {
				return &actions.GetPackageMessage{Path: action.Path, Message: m}
			},
			Close: func() flux.ActionInterface {
				return &actions.GetPackageClose{Path: action.Path}
			},
		})
		payload.Notify()
	case *actions.GetPackageOpen:
		s.app.Dispatch(&actions.Send{
			Message: messages.GetPackages{
				Path:    action.Path,
				Tags:    s.app.Tags.Tags(),
				Source:  s.SourceHashes(),
				Objects: s.ObjectsHashes(),
			},
		})
	case *actions.GetPackageMessage:
		switch message := action.Message.(type) {
		default:
			return s.app.HandleGenericStatusMessage(payload, message)
		case messages.Source:
			s.wait.Add(1)
			go func() {
				defer s.wait.Done()
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
					s.mutex.Lock()
					defer s.mutex.Unlock()
					s.source[sp.Path] = sp.Files
					s.sourceHashes[sp.Path] = message.Hash
				}()
				getwait.Wait()
				s.app.Log(message.Path)
			}()
			return true
		case messages.Objects:
			s.wait.Add(1)
			go func() {
				defer s.wait.Done()
				var getwait sync.WaitGroup
				getwait.Add(1)
				go func() {
					defer getwait.Done()
					resp, err := http.Get(fmt.Sprintf("%s://%s/%s.%s.objects.gob", config.Protocol[config.Pkg], config.Host[config.Pkg], message.Path, message.Hash))
					if err != nil {
						s.app.Fail(err)
						return
					}
					var op models.ObjectPack
					if err := gob.NewDecoder(resp.Body).Decode(&op); err != nil {
						s.app.Fail(err)
						return
					}
					s.mutex.Lock()
					defer s.mutex.Unlock()
					s.packageNames[op.Path] = op.Name
					s.objects[op.Path] = op.Objects
					// types
					s.types[op.Path] = map[string]gotypes.Type{}
					for _, objects := range op.Objects {
						for name, object := range objects {
							tn, ok := object.(*gotypes.TypeName)
							if !ok {
								continue
							}
							s.types[op.Path][name] = tn.Type
						}
					}
					s.objectsHashes[op.Path] = message.Hash
				}()
				getwait.Wait()
				s.app.Log(message.Path)
			}()
			return true
		case messages.PackageIndex:
			s.path = message.Path
			s.tags = message.Tags
			s.index = &message
		}
	case *actions.GetPackageClose:
		s.wait.Wait()
		if !s.Fresh() {
			s.app.Fail(errors.New("websocket closed but package not fully updated"))
			return true
		}
		s.app.LogHidef("done")
		payload.Notify()
	}
	return true
}
