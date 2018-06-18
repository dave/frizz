package treenodes

import (
	"github.com/dave/frizz/stores"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
)

type Node struct {
	vecty.Core
	app            *stores.App
	body, children []vecty.MarkupOrChild
}

/*
func (m *Node) Mount() {
	m.app.Watch(m, func(done chan struct{}) {
		defer close(done)
	})
}

func (m *Node) Unmount() {
	m.app.Delete(m)
}
*/

func (m *Node) Body(body ...vecty.MarkupOrChild) *Node {
	m.body = body
	return m
}

func (m *Node) Children(children ...vecty.MarkupOrChild) *Node {
	m.children = children
	return m
}

func (m *Node) Build() vecty.ComponentOrHTML {

	body := []vecty.MarkupOrChild{
		vecty.Markup(
			vecty.Class("tree-node-body"),
		),
	}
	body = append(body, m.body...)

	children := []vecty.MarkupOrChild{
		vecty.Markup(
			vecty.Class("tree-node-children"),
		),
	}
	children = append(children, m.children...)

	return elem.ListItem(
		//Plus(),
		elem.Div(
			body...,
		),
		elem.UnorderedList(
			children...,
		),
	)
}
