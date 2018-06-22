package treenodes

import (
	"fmt"

	"go/ast"

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
	data := v.app.Data.Expr(ob)

	// determine type
	var typ gotypes.Type
	switch ob := ob.(type) {
	case *gotypes.Var:
		typ = ob.Type
	case *gotypes.Const:
		typ = ob.Type
	}

	typ = v.app.Packages.ResolveType(typ, v.path, v.file, data)

	children := childrenForNode(v.app, v.path, v.file, typ, data)

	return v.Body(
		vecty.Text(v.name + fmt.Sprintf(" (%T)", typ)),
	).Children(
		children...,
	).Build()
}

func childrenForNode(app *stores.App, path, file string, typ gotypes.Type, data ast.Expr) []vecty.MarkupOrChild {
	var children []vecty.MarkupOrChild
	switch typ := typ.(type) {
	case *gotypes.Struct:
		// make a map of name -> field
		fieldData := map[string]ast.Expr{}
		switch data := data.(type) {
		case *ast.CompositeLit:
			for _, el := range data.Elts {
				switch el := el.(type) {
				case *ast.KeyValueExpr:
					var name string
					switch key := el.Key.(type) {
					case *ast.Ident:
						name = key.Name
					}
					fieldData[name] = el.Value
				}
			}
		}
		for _, field := range typ.Fields {
			dataItem := fieldData[field.Name]
			children = append(children, NewVar(app, path, file, field, dataItem))
		}
	}
	return children
}
