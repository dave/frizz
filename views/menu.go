package views

import (
	"github.com/dave/frizz/stores"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/gopherjs/vecty/event"
	"github.com/gopherjs/vecty/prop"
)

type Menu struct {
	vecty.Core
	app *stores.App
}

func NewMenu(app *stores.App) *Menu {
	v := &Menu{
		app: app,
	}
	return v
}

func (v *Menu) Render() vecty.ComponentOrHTML {

	return elem.Navigation(
		vecty.Markup(
			vecty.Class("menu", "navbar", "navbar-expand", "navbar-light", "bg-light"),
		),
		elem.UnorderedList(
			vecty.Markup(
				vecty.Class("navbar-nav", "ml-auto"),
			),
			elem.ListItem(
				vecty.Markup(
					vecty.Class("nav-item"),
				),
				elem.Span(
					vecty.Markup(
						vecty.Class("navbar-text"),
						vecty.Style("margin-right", "10px"),
						prop.ID("message"),
					),
					vecty.Text(""),
				),
			),
			elem.ListItem(
				vecty.Markup(
					vecty.Class("nav-item", "btn-group"),
				),
				elem.Button(
					vecty.Markup(
						vecty.Property("type", "button"),
						vecty.Class("btn", "btn-primary"),
						event.Click(func(e *vecty.Event) {
							// TODO: Run button
						}).PreventDefault(),
					),
					vecty.Text("..."),
				),
				elem.Button(
					vecty.Markup(
						vecty.Property("type", "button"),
						vecty.Data("toggle", "dropdown"),
						vecty.Property("aria-haspopup", "true"),
						vecty.Property("aria-expanded", "false"),
						vecty.Class("btn", "btn-primary", "dropdown-toggle", "dropdown-toggle-split"),
					),
					elem.Span(vecty.Markup(vecty.Class("sr-only")), vecty.Text("Options")),
				),
				elem.Div(
					vecty.Markup(
						vecty.Class("dropdown-menu", "dropdown-menu-right"),
					),
					elem.Anchor(
						vecty.Markup(
							vecty.Class("dropdown-item"),
							prop.Href(""),
							event.Click(func(e *vecty.Event) {
								// ...
							}).PreventDefault(),
						),
						vecty.Text("..."),
					),
					elem.Div(
						vecty.Markup(
							vecty.Class("dropdown-divider"),
						),
					),
				),
			),
		),
	)
}
