package treenodes

import (
	"fmt"

	"go/ast"

	"github.com/dave/frizz/actions"
	"github.com/dave/frizz/stores"
	"github.com/dave/jsgo/server/frizz/gotypes"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/gopherjs/vecty/event"
	"github.com/gopherjs/vecty/prop"
)

type Obj struct {
	*Node
	path, file, name string
	typ              gotypes.Type
	data             ast.Expr
}

func NewObj(app *stores.App, path, file, name string, typ gotypes.Type, data ast.Expr) *Obj {
	return &Obj{
		Node: &Node{
			app: app,
		},
		path: path,
		file: file,
		typ:  typ,
		data: data,
		name: name,
	}
}

func (v *Obj) Render() vecty.ComponentOrHTML {

	typ := v.app.Packages.ResolveType(v.typ, v.path, v.file, v.data)

	// childrenForNode also does ResolveType but will be a noop if already done.
	children := childrenForNode(v.app, v.path, v.file, typ, v.data)

	var data string
	if bl, ok := v.data.(*ast.BasicLit); ok {
		if len(bl.Value) > 10 {
			data = ": " + bl.Value[:10]
		} else {
			data = ": " + bl.Value
		}
	}

	return v.Body(
		elem.Anchor(
			vecty.Markup(
				prop.Href(""),
				event.Click(func(e *vecty.Event) {
					v.app.Dispatch(&actions.UserClickedNode{
						Path: v.path,
						File: v.file,
						Name: v.name,
						Type: v.typ,
						Data: v.data,
					})
				}).PreventDefault(),
			),
			vecty.Text(v.name),
		),
		vecty.Text(data),
	).Children(
		children...,
	).Build()
}

func childrenForNode(app *stores.App, path, file string, typ gotypes.Type, data ast.Expr) []vecty.MarkupOrChild {
	type named struct {
		name string
		typ  gotypes.Type
		data ast.Expr
	}
	var children []vecty.MarkupOrChild
	var fields []named
	switch typ := typ.(type) {
	case *gotypes.Basic:
		// no children in tree
		return nil
	case *gotypes.Slice, *gotypes.Array, *gotypes.Map:
		elem := typ.(hasElem).Element()
		switch data := data.(type) {
		case *ast.CompositeLit:
			elType := app.Packages.ResolveType(elem, "", "", nil)
			for key, el := range data.Elts {
				switch el := el.(type) {
				case *ast.KeyValueExpr:
					var name string
					switch key := el.Key.(type) {
					case *ast.Ident:
						name = key.Name
					case *ast.BasicLit:
						name = key.Value
					default:
						panic(fmt.Sprintf("TODO: collection el key of type %T\n", el.Key))
					}
					fields = append(fields, named{name, elType, el.Value})
				case *ast.CompositeLit:
					fields = append(fields, named{fmt.Sprint("#", key), elType, el})
				default:
					panic(fmt.Sprintf("TODO: collection el of type %T\n", el))
				}
			}
		default:
			panic(fmt.Sprintf("TODO: collection data of type %T\n", data))
		}
	case *gotypes.Struct:
		// make a map of name -> field
		typeFields := map[string]gotypes.Type{}
		for _, field := range typ.Fields {
			typeFields[field.Name] = field.Type
		}
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
						panic(fmt.Sprintf("TODO: struct el key of type %T\n", el.Key))
					}
					fields = append(fields, named{name, typeFields[name], el.Value})
				default:
					panic(fmt.Sprintf("TODO: struct el of type %T\n", el))
				}
			}
		default:
			panic(fmt.Sprintf("TODO: struct data of type %T\n", data))
		}
	default:
		panic(fmt.Sprintf("TODO: type of type %T\n", typ))
	}
	for _, field := range fields {
		children = append(children, NewObj(app, path, file, field.name, field.typ, field.data))
	}
	return children
}

type hasElem interface {
	Element() gotypes.Type
}
