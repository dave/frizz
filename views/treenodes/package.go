package treenodes

import (
	"github.com/dave/frizz/stores"
	"github.com/gopherjs/vecty"
)

type Package struct {
	*Node
	path, name string
}

func NewPackage(app *stores.App, path, name string) *Package {
	return &Package{
		Node: &Node{
			app: app,
		},
		path: path,
		name: name,
	}
}

func (v *Package) Render() vecty.ComponentOrHTML {

	var children []vecty.MarkupOrChild
	for _, file := range v.app.Packages.SortedSourceFiles(v.path) {
		children = append(children, NewFile(v.app, v.path, file))
	}

	return v.Body(
		vecty.Text(v.name),
	).Children(
		children...,
	).Build()
}
