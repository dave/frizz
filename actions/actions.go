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

type GetSourceStart struct{ Path string }
type GetSourceOpen struct{ Path string }
type GetSourceMessage struct {
	Path    string
	Message interface{}
}
type GetSourceClose struct{}

type SourceChanged struct{}
