package ext

import (
	"github.com/dave/flux"
	"github.com/dave/frizz/ed/models"
	"github.com/dave/frizz/ed/stores"
)

var viewId = models.Id{"github.com/dave/frizz/ed/stores/ext", "View"}
var storeId = models.Id{"github.com/dave/frizz/ed/stores/ext", "Store"}

func init() {
	stores.RegisterExternalStoreFunc(
		storeId,
		func(a *stores.App) flux.StoreInterface {
			return NewStore(a)
		},
	)
}

func NewStore(app *stores.App) *Store {
	s := &Store{
		app: app,
	}
	return s
}

type Store struct {
	app *stores.App
}

func (s *Store) Handle(payload *flux.Payload) bool {
	switch action := payload.Action.(type) {
	default:
		_ = action
	}
	return true
}
