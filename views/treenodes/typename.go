package treenodes

import (
	"github.com/dave/frizz/stores"
	"github.com/gopherjs/vecty"
)

/*
	Not used right now - don't want to show types in the tree.
*/

type TypeName struct {
	*Node
	path, file, name string
}

func NewTypeName(app *stores.App, path, file, name string) *TypeName {
	return &TypeName{
		Node: &Node{
			app: app,
		},
		path: path,
		file: file,
		name: name,
	}
}

func (v *TypeName) Render() vecty.ComponentOrHTML {
	return v.Body(
		vecty.Text(v.name + " (type)"),
	).Build()
}
