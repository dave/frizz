package treenodes

import (
	"fmt"

	"github.com/dave/dst"
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
	path, file string
	name       string         // name is just a title - e.g. "#1" for slice fields
	root       gotypes.Object // the exported package-level object that the node is part of
	typ        gotypes.Type
	data       dst.Expr
}

func NewObj(app *stores.App, path, file string, root gotypes.Object, name string, typ gotypes.Type, data dst.Expr) *Obj {
	return &Obj{
		Node: &Node{
			app: app,
		},
		path: path,
		file: file,
		root: root,
		name: name,
		typ:  typ,
		data: data,
	}
}

func (v *Obj) Render() vecty.ComponentOrHTML {

	typ := v.app.Packages.ResolveType(v.typ, v.path, v.file, v.data)

	// childrenForNode also does ResolveType but will be a noop if already done.
	children := childrenForNode(v.app, v.path, v.file, v.root, typ, v.data)

	var data string
	if bl, ok := v.data.(*dst.BasicLit); ok {
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
						Root: v.root,
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

func childrenForNode(app *stores.App, path, file string, root gotypes.Object, typ gotypes.Type, data dst.Expr) []vecty.MarkupOrChild {
	type named struct {
		name string
		typ  gotypes.Type
		data dst.Expr
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
		case *dst.CompositeLit:
			elType := app.Packages.ResolveType(elem, "", "", nil)
			for key, el := range data.Elts {
				switch el := el.(type) {
				case *dst.KeyValueExpr:
					var name string
					switch key := el.Key.(type) {
					case *dst.Ident:
						name = key.Name
					case *dst.BasicLit:
						name = key.Value
					default:
						panic(fmt.Sprintf("TODO: collection el key of type %T\n", el.Key))
					}
					fields = append(fields, named{name, elType, el.Value})
				case *dst.CompositeLit:
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
		case *dst.CompositeLit:
			for _, el := range data.Elts {
				switch el := el.(type) {
				case *dst.KeyValueExpr:
					var name string
					switch key := el.Key.(type) {
					case *dst.Ident:
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
		children = append(children, NewObj(app, path, file, root, field.name, field.typ, field.data))
	}
	return children
}

type hasElem interface {
	Element() gotypes.Type
}
