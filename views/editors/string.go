package views

import (
	"go/ast"
	"strconv"

	"github.com/dave/frizz/actions"
	"github.com/dave/frizz/stores"
	"github.com/dave/jsgo/server/frizz/gotypes"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/gopherjs/vecty/event"
	"github.com/gopherjs/vecty/prop"
	"golang.org/x/tools/go/ast/astutil"
)

type String struct {
	vecty.Core
	root gotypes.Object
	app  *stores.App
	Data ast.Expr `vecty:"prop"`
}

func NewString(app *stores.App, root gotypes.Object, data ast.Expr) *String {
	v := &String{
		app:  app,
		root: root,
		Data: data,
	}
	return v
}

/*
func (v *String) Mount() {
	v.app.Watch(v, func(done chan struct{}) {
		defer close(done)
		// Things that happen on every refresh
	})
	// Things that happen once at initialisation
}

func (v *String) Unmount() {
	v.app.Delete(v)
}
*/

func (v *String) Render() vecty.ComponentOrHTML {

	var value string
	switch data := v.Data.(type) {
	case *ast.BasicLit:
		value, _ = strconv.Unquote(data.Value)
	}

	return elem.Form(
		elem.Div(
			vecty.Markup(
				vecty.Class("form-group"),
			),
			/*
				elem.Label(
					vecty.Markup(
						prop.For("foo"),
					),
					vecty.Text("Foo"),
				),
			*/
			elem.Input(
				vecty.Markup(
					prop.Type(prop.TypeText),
					vecty.Class("form-control"),
					//prop.ID("foo"),
					prop.Value(value),
				),
			),
		),
		elem.Button(
			vecty.Markup(
				prop.Type(prop.TypeSubmit),
				vecty.Class("btn", "btn-primary"),
				event.Click(func(e *vecty.Event) {
					v.app.Dispatch(&actions.UserMutatedValue{
						Root: v.root,
						Change: func(c *astutil.Cursor) bool {
							if c.Node() != v.Data {
								return true
							}
							c.Node().(*ast.BasicLit).Value = `"FOO"`
							return true
						},
					})
				}).PreventDefault(),
			),
			vecty.Text("Submit"),
		),
	)
}
