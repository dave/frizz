package data

import (
	"go/types"

	"github.com/dave/frizz/gotypes"
)

func ConvertType(t types.Type, stack *[]types.Type) gotypes.Type {
	for _, stacked := range *stack {
		if t == stacked {
			return gotypes.Circular("circular reference")
		}
	}
	*stack = append(*stack, t)
	defer func() {
		*stack = (*stack)[:len(*stack)-1]
	}()
	switch t := t.(type) {
	case *types.Basic:
		return &gotypes.Basic{
			Kind: gotypes.BasicKind(t.Kind()),
			Info: gotypes.BasicInfo(t.Info()),
			Name: t.Name(),
		}
	case *types.Array:
		return &gotypes.Array{
			Len:  t.Len(),
			Elem: ConvertType(t.Elem(), stack),
		}
	case *types.Slice:
		return &gotypes.Slice{
			Elem: ConvertType(t.Elem(), stack),
		}
	case *types.Struct:
		var fields []*gotypes.Var
		var tags []string
		for i := 0; i < t.NumFields(); i++ {
			fields = append(fields, ConvertVar(t.Field(i), stack))
			tags = append(tags, t.Tag(i))
		}
		return &gotypes.Struct{
			Fields: fields,
			Tags:   tags,
		}
	case *types.Pointer:
		return &gotypes.Pointer{
			Elem: ConvertType(t.Elem(), stack),
		}
	case *types.Tuple:
		var vars []*gotypes.Var
		for i := 0; i < t.Len(); i++ {
			vars = append(vars, ConvertVar(t.At(i), stack))
		}
		return &gotypes.Tuple{
			Vars: vars,
		}
	case *types.Signature:
		return &gotypes.Signature{
			Recv:     ConvertVar(t.Recv(), stack),
			Params:   ConvertType(t.Params(), stack).(*gotypes.Tuple),
			Results:  ConvertType(t.Results(), stack).(*gotypes.Tuple),
			Variadic: t.Variadic(),
		}
	case *types.Interface:
		var methods []*gotypes.Func
		var embeddeds []*gotypes.Named
		var allMethods []*gotypes.Func
		for i := 0; i < t.NumExplicitMethods(); i++ {
			methods = append(methods, ConvertFunc(t.ExplicitMethod(i), stack))
		}
		for i := 0; i < t.NumEmbeddeds(); i++ {
			embeddeds = append(embeddeds, ConvertType(t.Embedded(i), stack).(*gotypes.Named))
		}
		for i := 0; i < t.NumMethods(); i++ {
			allMethods = append(allMethods, ConvertFunc(t.Method(i), stack))
		}
		return &gotypes.Interface{
			Methods:    methods,
			Embeddeds:  embeddeds,
			AllMethods: allMethods,
		}
	case *types.Map:
		return &gotypes.Map{
			Key:  ConvertType(t.Key(), stack),
			Elem: ConvertType(t.Elem(), stack),
		}
	case *types.Chan:
		return &gotypes.Chan{
			Dir:  gotypes.ChanDir(t.Dir()),
			Elem: ConvertType(t.Elem(), stack),
		}
	case *types.Named:
		var methods []*gotypes.Func
		for i := 0; i < t.NumMethods(); i++ {
			methods = append(methods, ConvertFunc(t.Method(i), stack))
		}
		var path string
		if t.Obj().Pkg() != nil {
			path = t.Obj().Pkg().Path()
		}
		return &gotypes.Named{
			Obj: &gotypes.TypeName{
				Obj: gotypes.Obj{
					Pkg:  path,
					Name: t.Obj().Name(),
					Typ:  ConvertType(t.Obj().Type(), stack),
				},
			},
			Type:    ConvertType(t.Underlying(), stack),
			Methods: methods,
		}
	}
	// notest
	return nil
}

func ConvertFunc(f *types.Func, stack *[]types.Type) *gotypes.Func {
	if f == nil {
		// notest
		return nil
	}
	var path string
	if f.Pkg() != nil {
		path = f.Pkg().Path()
	}
	return &gotypes.Func{
		Obj: gotypes.Obj{
			Pkg:  path,
			Name: f.Name(),
			Typ:  ConvertType(f.Type(), stack),
		},
	}
}

func ConvertVar(v *types.Var, stack *[]types.Type) *gotypes.Var {
	if v == nil {
		return nil
	}
	var path string
	if v.Pkg() != nil {
		path = v.Pkg().Path()
	}
	return &gotypes.Var{
		Obj: gotypes.Obj{
			Pkg:  path,
			Name: v.Name(),
			Typ:  ConvertType(v.Type(), stack),
		},
		Anonymous: v.Anonymous(),
		IsField:   v.IsField(),
	}
}
