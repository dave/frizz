package treenodes

import (
	"github.com/dave/frizz/stores"
	"github.com/gopherjs/vecty"
)

type Const struct {
	*Node
	path, file, name string
}

func NewConst(app *stores.App, path, file, name string) *Const {
	return &Const{
		Node: &Node{
			app: app,
		},
		path: path,
		file: file,
		name: name,
	}
}

func (v *Const) Render() vecty.ComponentOrHTML {

	return v.Body(
		vecty.Text(v.name + " (const)"),
	).Build()
}
