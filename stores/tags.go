package stores

import (
	"github.com/dave/flux"
)

func NewTagStore(a *App) *TagStore {
	s := &TagStore{
		app: a,
	}
	return s
}

type TagStore struct {
	app  *App
	tags []string
}

func (s *TagStore) Tags() []string {
	return s.tags
}

func (s *TagStore) Handle(payload *flux.Payload) bool {
	switch action := payload.Action.(type) {
	default:
		_ = action
	}
	return true
}
