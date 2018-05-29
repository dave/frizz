package actions

import (
	"github.com/dave/flux"
	"github.com/dave/services"
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

type TypesStart struct{}
type TypesOpen struct{}
type TypesMessage struct{ Message interface{} }
type TypesClose struct{}
