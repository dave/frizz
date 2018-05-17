package views

import (
	"github.com/dave/frizz/ed/models"
	"github.com/dave/frizz/ed/stores"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/gopherjs/vecty/prop"
)

type Tree struct {
	vecty.Core
	app *stores.App
}

func NewTree(app *stores.App) *Tree {
	v := &Tree{
		app: app,
	}
	return v
}

func (v *Tree) Mount() {
	v.app.Watch(v, func(done chan struct{}) {
		defer close(done)
		// Things that happen on every refresh
	})
	// Things that happen once at initialisation
}

func (v *Tree) Unmount() {
	v.app.Delete(v)
}

func (v *Tree) Render() vecty.ComponentOrHTML {
	extView := GetExternalViewFunc(models.Id{"github.com/dave/frizz/ed/stores/ext", "View"})

	return elem.Div(
		vecty.Markup(
			prop.ID("tree"),
			vecty.Class("tree"),
		),
		extView(v.app, nil),
	)
}
