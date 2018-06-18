package treenodes

import (
	"fmt"

	"github.com/dave/frizz/stores"
	"github.com/dave/jsgo/server/frizz/gotypes"
	"github.com/gopherjs/vecty"
)

// Decl is a gotypes.Var or gotypes.Const
type Decl struct {
	*Node
	path, file, name string
}

func NewDecl(app *stores.App, path, file, name string) *Decl {
	return &Decl{
		Node: &Node{
			app: app,
		},
		path: path,
		file: file,
		name: name,
	}
}

func (v *Decl) Render() vecty.ComponentOrHTML {

	ob := v.app.Packages.ObjectsInFile(v.path, v.file)[v.name]

	// determine type
	var t gotypes.Type
	switch ob := ob.(type) {
	case *gotypes.Var:
		t = ob.Type
	case *gotypes.Const:
		t = ob.Type
	}

	t = v.app.Packages.ResolveType(t)

	if _, ok := t.(*gotypes.Interface); ok {
		d := v.app.Data.Expr(ob)
		t = v.app.Packages.ResolveTypeFromExpr(v.path, v.file, d)
		t = v.app.Packages.ResolveType(t)
	}

	var children []vecty.MarkupOrChild
	/*
		for _, o := range ... {
			...
		}
	*/

	return v.Body(
		vecty.Text(v.name + fmt.Sprintf(" (%T)", t)),
	).Children(
		children...,
	).Build()
}
