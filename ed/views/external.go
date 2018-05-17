package views

import (
	"sync"

	"github.com/dave/frizz/ed/models"
	"github.com/dave/frizz/ed/stores"
	"github.com/gopherjs/vecty"
)

var externalViewFuncsM sync.RWMutex
var externalViewFuncs = map[models.Id]ViewFunc{}

type ViewFunc func(app *stores.App, options map[string]interface{}) vecty.Component

func RegisterExternalViewFunc(id models.Id, v ViewFunc) {
	externalViewFuncsM.Lock()
	defer externalViewFuncsM.Unlock()
	externalViewFuncs[id] = v
}

func GetExternalViewFunc(id models.Id) ViewFunc {
	externalViewFuncsM.RLock()
	defer externalViewFuncsM.RUnlock()
	return externalViewFuncs[id]
}
