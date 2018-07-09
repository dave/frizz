package views

import (
	"github.com/dave/frizz/stores"
	"github.com/dave/frizz/views/editors"
	"github.com/dave/jsgo/server/frizz/gotypes"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/gopherjs/vecty/prop"
)

type Editor struct {
	vecty.Core
	app *stores.App
}

func NewEditor(app *stores.App) *Editor {
	v := &Editor{
		app: app,
	}
	return v
}

func (v *Editor) Mount() {
	v.app.Watch(v, func(done chan struct{}) {
		defer close(done)
		// Things that happen on every refresh
	})
	// Things that happen once at initialisation
}

func (v *Editor) Unmount() {
	v.app.Delete(v)
}

func (v *Editor) Render() vecty.ComponentOrHTML {

	if v.app.Editor.Root() == nil {
		return elem.Div()
	}

	var editor vecty.MarkupOrChild
	switch t := v.app.Editor.Type().(type) {
	case *gotypes.Basic:
		switch t.Kind {
		case gotypes.Int:
			editor = views.NewInt(v.app, v.app.Editor.Root(), v.app.Editor.Data())
		case gotypes.String:
			editor = views.NewString(v.app, v.app.Editor.Root(), v.app.Editor.Data())
		}
	}

	return elem.Div(
		vecty.Markup(
			prop.ID("editor"),
			vecty.Class("editor"),
		),
		elem.Heading1(
			vecty.Text(v.app.Editor.Name()),
		),
		editor,
	)
}
