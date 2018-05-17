package views

import (
	"github.com/dave/frizz/ed/stores"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/gopherjs/vecty/prop"
)

type Empty struct {
	vecty.Core
	app *stores.App
}

func NewEmpty(app *stores.App) *Empty {
	v := &Empty{
		app: app,
	}
	return v
}

func (v *Empty) Mount() {
	v.app.Watch(v, func(done chan struct{}) {
		defer close(done)
		// Things that happen on every refresh
	})
	// Things that happen once at initialisation
}

func (v *Empty) Unmount() {
	v.app.Delete(v)
}

func (v *Empty) Render() vecty.ComponentOrHTML {
	return elem.Div(
		vecty.Markup(
			prop.ID("empty"),
			vecty.Class("empty"),
		),
	)
}
