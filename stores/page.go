package stores

import (
	"github.com/dave/flux"
	"github.com/dave/frizz/actions"
)

func NewPageStore(app *App) *PageStore {
	s := &PageStore{
		app:        app,
		splitSizes: []float64{20, 80},
	}
	return s
}

type PageStore struct {
	app *App

	splitSizes     []float64
	currentPackage string
}

func (s *PageStore) CurrentPackage() string {
	return s.currentPackage
}

func (s *PageStore) SplitSizes() []float64 {
	return s.splitSizes
}

func (s *PageStore) Handle(payload *flux.Payload) bool {
	switch action := payload.Action.(type) {
	case *actions.GetPackageClose:
		payload.Wait(s.app.Packages)
		s.currentPackage = action.Path
		payload.Notify()
	default:
		_ = action
	}
	return true
}
