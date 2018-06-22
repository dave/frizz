package treenodes

import (
	"fmt"

	"go/ast"

	"github.com/dave/frizz/stores"
	"github.com/dave/jsgo/server/frizz/gotypes"
	"github.com/gopherjs/vecty"
)

type Obj struct {
	*Node
	path, file string
	obj        gotypes.Obj
	data       ast.Expr
}

func NewObj(app *stores.App, path, file string, obj gotypes.Obj, data ast.Expr) *Obj {
	return &Obj{
		Node: &Node{
			app: app,
		},
		path: path,
		file: file,
		obj:  obj,
		data: data,
	}
}

func (v *Obj) Render() vecty.ComponentOrHTML {

	typ := v.app.Packages.ResolveType(v.obj.Type, v.path, v.file, v.data)

	// childrenForNode also does ResolveType but will be a noop if already done.
	children := childrenForNode(v.app, v.path, v.file, typ, v.data)

	return v.Body(
		vecty.Text(v.obj.Name + fmt.Sprintf(" (%T)", typ)),
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
					default:
						panic(fmt.Sprintf("TODO: key of type %T\n", el.Key))
					}
					fieldData[name] = el.Value
				default:
					panic(fmt.Sprintf("TODO: el of type %T\n", el))
				}
			}
		default:
			panic(fmt.Sprintf("TODO: data of type %T\n", data))
		}
		for _, field := range typ.Fields {
			dataItem := fieldData[field.Name]
			children = append(children, NewObj(app, path, file, field.Obj, dataItem))
		}
	}
	return children
}
