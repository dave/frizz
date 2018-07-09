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
			ob := o.Object()
			data := v.app.Data.Expr(v.app.Packages.ObjectsInFile(v.path, v.file)[ob.Name])
			children = append(children, NewObj(v.app, v.path, v.file, o, ob.Name, ob.Type, data))
		}
	}

	return v.Body(
		vecty.Text(v.file),
	).Children(
		children...,
	).Build()
}
