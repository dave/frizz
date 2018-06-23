package views

import (
	"go/ast"

	"strconv"

	"github.com/dave/frizz/stores"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/gopherjs/vecty/prop"
)

type String struct {
	vecty.Core
	app  *stores.App
	Data ast.Expr `vecty:"prop"`
}

func NewString(app *stores.App, data ast.Expr) *String {
	v := &String{
		app:  app,
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
			),
			vecty.Text("Submit"),
		),
	)
}
