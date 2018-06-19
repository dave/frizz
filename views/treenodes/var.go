package treenodes

import (
	"fmt"

	"go/ast"

	"github.com/dave/frizz/stores"
	"github.com/dave/jsgo/server/frizz/gotypes"
	"github.com/gopherjs/vecty"
)

type Var struct {
	*Node
	path, file string
	tvar       *gotypes.Var
	data       ast.Expr
}

func NewVar(app *stores.App, path, file string, tvar *gotypes.Var, data ast.Expr) *Var {
	return &Var{
		Node: &Node{
			app: app,
		},
		path: path,
		file: file,
		tvar: tvar,
		data: data,
	}
}

func (v *Var) Render() vecty.ComponentOrHTML {

	typ := v.app.Packages.ResolveType(v.tvar.Type)

	if _, ok := typ.(*gotypes.Interface); ok {
		// If the type of the decl is an interface, try to extract the actual type from the data
		typ = v.app.Packages.ResolveTypeFromExpr(v.path, v.file, v.data)
	}

	children := childrenForNode(v.app, v.path, v.file, typ, v.data)

	return v.Body(
		vecty.Text(v.tvar.Name + fmt.Sprintf(" (%T)", typ)),
	).Children(
		children...,
	).Build()
}
