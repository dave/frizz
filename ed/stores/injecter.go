package stores

import (
	"go/types"

	"honnef.co/go/js/dom"

	"context"

	"strconv"

	"github.com/dave/flux"
	"github.com/dave/frizz/ed/actions"
	"github.com/dave/play/stores/builderjs"
	"github.com/gopherjs/gopherjs/compiler"
	"github.com/gopherjs/gopherjs/js"
)

func NewInjectorStore(a *App) *InjectorStore {
	s := &InjectorStore{
		app: a,
	}
	return s
}

type InjectorStore struct {
	app *App
}

func (s *InjectorStore) Handle(payload *flux.Payload) bool {
	switch action := payload.Action.(type) {
	case *actions.Inject:
		if err := s.inject(); err != nil {
			s.app.Fail(err)
			return true
		}
	default:
		_ = action
	}
	return true
}

func (s *InjectorStore) inject() error {

	path := "github.com/dave/frizz/alerter"

	if _, ok := js.Global.Get("$packages").Get(path).Interface().(map[string]interface{}); ok {
		// this package is already loaded
		return nil
	}

	source := map[string]map[string]string{
		path: {
			"alerter.go": `package alerter

func init() {
	println("alerter init")
}`,
		},
	}
	a, err := builderjs.BuildPackage(
		path,
		source,
		[]string{},
		[]*compiler.Archive{},
		false,
		map[string]*compiler.Archive{},
		map[string]*types.Package{},
	)
	if err != nil {
		return err
	}
	code, _, err := builderjs.GetPackageCode(
		context.Background(),
		a,
		false,
		false,
	)
	if err != nil {
		return err
	}

	doc := dom.GetWindow().Document()
	head := doc.GetElementsByTagName("head")[0].(*dom.HTMLHeadElement)

	script := doc.CreateElement("script")
	script.SetInnerHTML(string(code) + "$packages[" + strconv.Quote(path) + "].$init();")
	head.AppendChild(script)

	return nil
}
