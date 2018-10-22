package views

import (
	"fmt"

	"github.com/dave/dst"
	"github.com/dave/dst/dstutil"
	"github.com/dave/frizz/actions"
	"github.com/dave/frizz/stores"
	"github.com/dave/jsgo/server/frizz/gotypes"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/gopherjs/vecty/event"
	"github.com/gopherjs/vecty/prop"
)

type Int struct {
	vecty.Core
	root  gotypes.Object
	app   *stores.App
	Data  dst.Expr `vecty:"prop"`
	input *vecty.HTML
}

func NewInt(app *stores.App, root gotypes.Object, data dst.Expr) *Int {
	v := &Int{
		app:  app,
		root: root,
		Data: data,
	}
	return v
}

/*
func (v *Int) Mount() {
	v.app.Watch(v, func(done chan struct{}) {
		defer close(done)
		// Things that happen on every refresh
	})
	// Things that happen once at initialisation
}

func (v *Int) Unmount() {
	v.app.Delete(v)
}
*/

func (v *Int) Render() vecty.ComponentOrHTML {

	var value string
	switch data := v.Data.(type) {
	case *dst.BasicLit:
		value = data.Value
	}

	v.input = elem.Input(
		vecty.Markup(
			prop.Type(prop.TypeNumber),
			vecty.Class("form-control"),
			prop.Value(value),
			event.KeyPress(func(ev *vecty.Event) {
				if ev.Get("keyCode").Int() == 13 {
					ev.Call("preventDefault")
					v.save(ev)
				}
			}),
		),
	)

	return elem.Form(
		elem.Div(
			vecty.Markup(
				vecty.Class("form-group"),
			),
			v.input,
		),
		elem.Button(
			vecty.Markup(
				prop.Type(prop.TypeSubmit),
				vecty.Class("btn", "btn-primary"),
				event.Click(v.save).PreventDefault(),
			),
			vecty.Text("Submit"),
		),
	)
}

func (v *Int) save(*vecty.Event) {
	value := v.input.Node().Get("value").Int()
	v.app.Dispatch(&actions.UserMutatedValue{
		Root: v.root,
		Change: func(c *dstutil.Cursor) bool {
			if c.Node() != v.Data {
				return true
			}
			c.Node().(*dst.BasicLit).Value = fmt.Sprintf("%#v", value)
			return true
		},
	})
}
