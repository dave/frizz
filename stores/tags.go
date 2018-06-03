package stores

import (
	"github.com/dave/flux"
)

func NewTagsStore(a *App) *TagsStore {
	s := &TagsStore{
		app: a,
	}
	return s
}

type TagsStore struct {
	app  *App
	tags []string
}

func (s *TagsStore) Tags() []string {
	return s.tags
}

func (s *TagsStore) Handle(payload *flux.Payload) bool {
	switch action := payload.Action.(type) {
	default:
		_ = action
	}
	return true
}
