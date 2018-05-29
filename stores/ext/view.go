package ext

import (
	"github.com/dave/frizz/stores"
	"github.com/dave/frizz/views"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
)

func init() {
	views.RegisterExternalViewFunc(
		viewId,
		func(app *stores.App, options map[string]interface{}) vecty.Component {
			return NewView(app)
		},
	)
}

type View struct {
	vecty.Core
	app   *stores.App
	store *Store
}

func NewView(app *stores.App) *View {
	v := &View{
		app:   app,
		store: app.ExternalStore(storeId).(*Store),
	}
	return v
}

func (v *View) Mount() {
	v.app.Watch(v, func(done chan struct{}) {
		defer close(done)
		// Things that happen on every refresh
	})
	// Things that happen once at initialisation
}

func (v *View) Unmount() {
	v.app.Delete(v)
}

func (v *View) Render() vecty.ComponentOrHTML {
	return elem.Div(
		vecty.Text("external component"),
	)
}
