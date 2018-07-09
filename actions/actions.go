package actions

import (
	"go/ast"

	"github.com/dave/flux"
	"github.com/dave/jsgo/server/frizz/gotypes"
	"github.com/dave/services"
	"golang.org/x/tools/go/ast/astutil"
)

type Load struct{}

type UserChangedSplitSizes struct {
	SplitSizes []float64
}

type Inject struct{}

type Send struct{ Message services.Message }
type Dial struct {
	Open    func() flux.ActionInterface
	Message func(interface{}) flux.ActionInterface
	Close   func() flux.ActionInterface
}

type GetPackageStart struct{ Path string }
type GetPackageOpen struct{ Path string }
type GetPackageMessage struct {
	Path    string
	Message interface{}
}
type GetPackageClose struct{ Path string }

type UserChangedPackage struct{ Path string }

type UserClickedNode struct {
	Root gotypes.Object
	Name string
	Type gotypes.Type
	Data ast.Expr
}

type UserMutatedValue struct {
	Root   gotypes.Object
	Change astutil.ApplyFunc
}
