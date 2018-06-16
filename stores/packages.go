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

	packageNames map[string]string

	index *messages.PackageIndex

	mutex sync.Mutex
	wait  sync.WaitGroup
}

func (s *PackageStore) SourcePackages() []string {
	var paths []string
	for p := range s.source {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	return paths
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
		var downloaded, unchanged int
		for _, v := range s.index.Source {
			if v.Unchanged {
				unchanged++
			} else {
				downloaded++
			}
		}
		for _, v := range s.index.Objects {
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
		payload.Notify()
	}
	return true
}
