package treenodes

import (
	"github.com/dave/frizz/stores"
	"github.com/gopherjs/vecty"
)

type Var struct {
	*Node
	path, file, name string
}

func NewVar(app *stores.App, path, file, name string) *Var {
	return &Var{
		Node: &Node{
			app: app,
		},
		path: path,
		file: file,
		name: name,
	}
}

func (v *Var) Render() vecty.ComponentOrHTML {

	var children []vecty.MarkupOrChild
	/*
		for _, o := range ... {
			...
		}
	*/

	return v.Body(
		vecty.Text(v.name + " (var)"),
	).Children(
		children...,
	).Build()
}
