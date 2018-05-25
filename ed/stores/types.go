package stores

import (
	"github.com/dave/flux"
	"github.com/dave/frizz/ed/actions"
	"github.com/dave/frizz/server/messages"
)

func NewTypesStore(a *App) *TypesStore {
	s := &TypesStore{
		app: a,
	}
	return s
}

type TypesStore struct {
	app *App
}

func (s *TypesStore) Handle(payload *flux.Payload) bool {
	switch action := payload.Action.(type) {
	case *actions.TypesStart:
		s.app.Log("getting types")
		s.app.Dispatch(&actions.Dial{
			Open:    func() flux.ActionInterface { return &actions.TypesOpen{} },
			Message: func(m interface{}) flux.ActionInterface { return &actions.TypesMessage{Message: m} },
			Close:   func() flux.ActionInterface { return &actions.TypesClose{} },
		})
		payload.Notify()
	case *actions.TypesOpen:
		message := messages.Types{
			Path: "github.com/dave/jstest",
		}
		s.app.Dispatch(&actions.Send{
			Message: message,
		})
	case *actions.TypesMessage:
		// nothing
	case *actions.TypesClose:
		// nothing
	default:
		_ = action
	}
	return true
}
