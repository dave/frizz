package views

import (
	"go/ast"

	"github.com/dave/frizz/stores"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/gopherjs/vecty/prop"
)

type Int struct {
	vecty.Core
	app  *stores.App
	Data ast.Expr `vecty:"prop"`
}

func NewInt(app *stores.App, data ast.Expr) *Int {
	v := &Int{
		app:  app,
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
	case *ast.BasicLit:
		value = data.Value
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
					prop.Type(prop.TypeNumber),
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
