package stores

import (
	"github.com/dave/flux"
)

func NewEmptyStore(a *App) *EmptyStore {
	s := &EmptyStore{
		app: a,
	}
	return s
}

type EmptyStore struct {
	app *App
}

func (s *EmptyStore) Handle(payload *flux.Payload) bool {
	switch action := payload.Action.(type) {
	default:
		_ = action
	}
	return true
}
