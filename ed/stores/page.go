package stores

import (
	"github.com/dave/flux"
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

	splitSizes []float64
}

func (s *PageStore) SplitSizes() []float64 {
	return s.splitSizes
}

func (s *PageStore) Handle(payload *flux.Payload) bool {
	switch action := payload.Action.(type) {
	default:
		_ = action
	}
	return true
}
