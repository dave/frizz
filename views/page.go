package views

import (
	"github.com/dave/frizz/actions"
	"github.com/dave/frizz/stores"
	"github.com/dave/splitter"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/gopherjs/vecty/prop"
)

type Page struct {
	vecty.Core
	app *stores.App

	split *splitter.Split
	tree  *Tree
}

func NewPage(app *stores.App) *Page {
	v := &Page{
		app: app,
	}
	return v
}

func (v *Page) Mount() {
	v.app.Watch(v, func(done chan struct{}) {
		defer close(done)
		sizes := v.app.Page.SplitSizes()
		if v.split.Changed(sizes) {
			v.split.SetSizes(sizes)
		}
	})

	v.split = splitter.New("split")
	v.split.Init(
		js.S{"#left", "#right"},
		js.M{
			"sizes": v.app.Page.SplitSizes(),
			"onDragEnd": func() {
				v.app.Dispatch(&actions.UserChangedSplitSizes{
					SplitSizes: v.split.GetSizes(),
				})
			},
		},
	)
}

func (v *Page) Unmount() {
	v.app.Delete(v)
}

const Styles = `
	html, body {
		height: 100%;
	}
	#left {
		display: flex;
		flex-flow: column;
		height: 100%;
	}
	.menu {
		min-height: 56px;
	}
	.tree, .empty-panel {
		flex: 1;
		width: 100%;
	}
	.empty-panel {
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.split {
		height: 100%;
		width: 100%;
	}
	.gutter {
		height: 100%;
		background-color: #eee;
		background-repeat: no-repeat;
		background-position: 50%;
	}
	.gutter.gutter-horizontal {
		float: left;
		cursor: col-resize;
		background-image:  url('data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUAAAAeCAYAAADkftS9AAAAIklEQVQoU2M4c+bMfxAGAgYYmwGrIIiDjrELjpo5aiZeMwF+yNnOs5KSvgAAAABJRU5ErkJggg==')
	}
	.gutter.gutter-vertical {
		cursor: row-resize;
		background-image:  url('data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAB4AAAAFAQMAAABo7865AAAABlBMVEVHcEzMzMzyAv2sAAAAAXRSTlMAQObYZgAAABBJREFUeF5jOAMEEAIEEFwAn3kMwcB6I2AAAAAASUVORK5CYII=')
	}
	.split {
		-webkit-box-sizing: border-box;
		-moz-box-sizing: border-box;
		box-sizing: border-box;
	}
	.split, .gutter.gutter-horizontal {
		float: left;
	}
	.octicon {
		display: inline-block;
		vertical-align: text-top;
		fill: currentColor;
	}
`

func (v *Page) Render() vecty.ComponentOrHTML {
	return elem.Body(
		NewMenu(v.app),
		elem.Div(
			vecty.Markup(
				vecty.Class("container-fluid", "p-0", "split", "split-horizontal"),
			),
			v.renderTree(),
			v.renderContent(),
		),
	)
}

func (v *Page) renderTree() *vecty.HTML {

	v.tree = NewTree(v.app)

	var emptyDisplay, loadingDisplay, emptyMessageDisplay string
	emptyDisplay = "none"

	return elem.Div(
		vecty.Markup(
			prop.ID("left"),
			vecty.Class("split"),
		),
		v.tree,
		elem.Div(
			vecty.Markup(
				vecty.Class("empty-panel"),
				vecty.Style("display", emptyDisplay),
			),
			elem.Span(
				vecty.Markup(
					vecty.Style("display", loadingDisplay),
				),
				vecty.Text("Loading..."),
			),
			elem.Span(
				vecty.Markup(
					vecty.Style("display", emptyMessageDisplay),
				),
				vecty.Text("Empty"),
			),
		),
	)
}

func (v *Page) renderContent() *vecty.HTML {
	return elem.Div(
		vecty.Markup(
			prop.ID("right"),
			vecty.Class("split", "split-vertical"),
		),
		elem.Div(
			vecty.Markup(
				prop.ID("content-holder"),
			),
		),
	)
}
