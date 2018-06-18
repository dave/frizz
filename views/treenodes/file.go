package treenodes

import (
	"github.com/dave/frizz/stores"
	"github.com/dave/jsgo/server/frizz/gotypes"
	"github.com/gopherjs/vecty"
)

type File struct {
	*Node
	path, file string
}

func NewFile(app *stores.App, path, file string) *File {
	return &File{
		Node: &Node{
			app: app,
		},
		path: path,
		file: file,
	}
}

func (v *File) Render() vecty.ComponentOrHTML {

	var children []vecty.MarkupOrChild
	for _, o := range v.app.Packages.SortedObjectsInFile(v.path, v.file) {
		switch o := o.(type) {
		case *gotypes.TypeName:
			//children = append(children, NewTypeName(v.app, v.path, v.file, o.Id().Name))
		case *gotypes.Var, *gotypes.Const:
			children = append(children, NewDecl(v.app, v.path, v.file, o.Id().Name))
			// TODO: more
		}
	}

	return v.Body(
		vecty.Text(v.file),
	).Children(
		children...,
	).Build()
}
