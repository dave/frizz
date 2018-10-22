package stores

import (
	"github.com/dave/dst"
	"github.com/dave/flux"
	"github.com/dave/frizz/actions"
	"github.com/dave/jsgo/server/frizz/gotypes"
)

type EditorStore struct {
	app  *App
	root gotypes.Object
	name string // name is not always a top-level name in file (can be deeply nested).
	typ  gotypes.Type
	data dst.Expr
}

func NewEditorStore(a *App) *EditorStore {
	s := &EditorStore{
		app: a,
	}
	return s
}

func (s *EditorStore) Root() gotypes.Object {
	return s.root
}

func (s *EditorStore) Name() string {
	return s.name
}

func (s *EditorStore) Type() gotypes.Type {
	return s.typ
}

func (s *EditorStore) Data() dst.Expr {
	return s.data
}

func (s *EditorStore) Handle(payload *flux.Payload) bool {
	switch action := payload.Action.(type) {
	case *actions.UserClickedNode:
		s.root = action.Root
		s.name = action.Name
		s.typ = action.Type
		s.data = action.Data
		payload.Notify()
	}
	return true
}
