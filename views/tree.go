package views

import (
	"github.com/dave/frizz/stores"
	"github.com/dave/frizz/views/treenodes"
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
	//extView := GetExternalViewFunc(models.Id{"github.com/dave/frizz/stores/ext", "View"})
	//extView(v.app, nil),

	nodes := []vecty.MarkupOrChild{
		vecty.Markup(
			prop.ID("tree"),
			vecty.Class("tree"),
		),
	}
	for _, path := range v.app.Packages.SortedSourcePackages() {
		name := v.app.Packages.PackageName(path)
		nodes = append(nodes, treenodes.NewPackage(v.app, path, name))
	}

	return elem.UnorderedList(
		nodes...,
	)
}
