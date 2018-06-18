package views

import (
	"github.com/dave/frizz/actions"
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
				vecty.Class("navbar-nav", "mr-auto"),
			),
			v.renderPackageDropdown(),
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

func (v *Menu) renderPackageDropdown() *vecty.HTML {
	var packageItems []vecty.MarkupOrChild
	packageItems = append(packageItems,
		vecty.Markup(
			vecty.Class("dropdown-menu"),
			vecty.Property("aria-labelledby", "packageDropdown"),
		),
	)
	for _, path := range v.app.Packages.SortedSourcePackages() {
		path := path
		packageItems = append(packageItems,
			elem.Anchor(
				vecty.Markup(
					vecty.Class("dropdown-item"),
					vecty.ClassMap{
						"disabled": path == v.app.Page.CurrentPackage(),
					},
					prop.Href(""),
					event.Click(func(e *vecty.Event) {
						v.app.Dispatch(&actions.UserChangedPackage{
							Path: path,
						})
					}).PreventDefault(),
				),
				vecty.Text(v.app.Packages.DisplayPath(path)),
			),
		)
	}
	packageItems = append(packageItems,
		elem.Div(
			vecty.Markup(
				vecty.Class("dropdown-divider"),
			),
		),
		elem.Anchor(
			vecty.Markup(
				vecty.Class("dropdown-item"),
				prop.Href(""),
				event.Click(func(e *vecty.Event) {
					v.app.LogHide("TODO")
					//v.app.Dispatch(&actions.ModalOpen{Modal: models.AddPackageModal})
				}).PreventDefault(),
			),
			vecty.Text("Add package"),
		),
		elem.Anchor(
			vecty.Markup(
				vecty.Class("dropdown-item"),
				prop.Href(""),
				event.Click(func(e *vecty.Event) {
					v.app.LogHide("TODO")
					//v.app.Dispatch(&actions.ModalOpen{Modal: models.LoadPackageModal})
				}).PreventDefault(),
			),
			vecty.Text("Load package"),
		),
		elem.Anchor(
			vecty.Markup(
				vecty.Class("dropdown-item"),
				prop.Href(""),
				event.Click(func(e *vecty.Event) {
					v.app.LogHide("TODO")
					//v.app.Dispatch(&actions.ModalOpen{Modal: models.RemovePackageModal})
				}).PreventDefault(),
			),
			vecty.Text("Remove package"),
		),
	)

	classes := vecty.Class("nav-item", "dropdown", "d-none")
	if len(v.app.Packages.SortedSourcePackages()) > 0 {
		classes = vecty.Class("nav-item", "dropdown")
	}

	return elem.ListItem(
		vecty.Markup(
			classes,
		),
		elem.Anchor(
			vecty.Markup(
				prop.ID("packageDropdown"),
				prop.Href(""),
				vecty.Class("nav-link", "dropdown-toggle"),
				vecty.Property("role", "button"),
				vecty.Data("toggle", "dropdown"),
				vecty.Property("aria-haspopup", "true"),
				vecty.Property("aria-expanded", "false"),
				event.Click(func(ev *vecty.Event) {}).PreventDefault(),
			),
			vecty.Text(v.app.Packages.DisplayName(v.app.Page.CurrentPackage())),
		),
		elem.Div(
			packageItems...,
		),
	)
}
