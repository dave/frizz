package data

import (
	"go/types"

	"github.com/dave/frizz/gotypes"
)

func ConvertType(t types.Type) gotypes.Type {
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
			Elem: ConvertType(t.Elem()),
		}
	case *types.Slice:
		return &gotypes.Slice{
			Elem: ConvertType(t.Elem()),
		}
	case *types.Struct:
		var fields []*gotypes.Var
		var tags []string
		for i := 0; i < t.NumFields(); i++ {
			fields = append(fields, ConvertVar(t.Field(i)))
			tags = append(tags, t.Tag(i))
		}
		return &gotypes.Struct{
			Fields: fields,
			Tags:   tags,
		}
	case *types.Pointer:
		return &gotypes.Pointer{
			Elem: ConvertType(t.Elem()),
		}
	case *types.Tuple:
		var vars []*gotypes.Var
		for i := 0; i < t.Len(); i++ {
			vars = append(vars, ConvertVar(t.At(i)))
		}
		return &gotypes.Tuple{
			Vars: vars,
		}
	case *types.Signature:
		return &gotypes.Signature{
			Recv:     ConvertVar(t.Recv()),
			Params:   ConvertType(t.Params()).(*gotypes.Tuple),
			Results:  ConvertType(t.Results()).(*gotypes.Tuple),
			Variadic: t.Variadic(),
		}
	case *types.Interface:
		var methods []*gotypes.Func
		var embeddeds []*gotypes.Named
		var allMethods []*gotypes.Func
		for i := 0; i < t.NumExplicitMethods(); i++ {
			methods = append(methods, ConvertFunc(t.ExplicitMethod(i)))
		}
		for i := 0; i < t.NumEmbeddeds(); i++ {
			embeddeds = append(embeddeds, ConvertType(t.Embedded(i)).(*gotypes.Named))
		}
		for i := 0; i < t.NumMethods(); i++ {
			allMethods = append(allMethods, ConvertFunc(t.Method(i)))
		}
		return &gotypes.Interface{
			Methods:    methods,
			Embeddeds:  embeddeds,
			AllMethods: allMethods,
		}
	case *types.Map:
		return &gotypes.Map{
			Key:  ConvertType(t.Key()),
			Elem: ConvertType(t.Elem()),
		}
	case *types.Chan:
		return &gotypes.Chan{
			Dir:  gotypes.ChanDir(t.Dir()),
			Elem: ConvertType(t.Elem()),
		}
	case *types.Named:
		var methods []*gotypes.Func
		for i := 0; i < t.NumMethods(); i++ {
			methods = append(methods, ConvertFunc(t.Method(i)))
		}
		return &gotypes.Named{
			Obj: &gotypes.TypeName{
				Obj: gotypes.Obj{
					Pkg:  t.Obj().Pkg().Path(),
					Name: t.Obj().Name(),
					// TODO: What to do here? Circular reference breaks json encoding.
					// Typ:  ConvertType(t.Obj().Type()),
				},
			},
			Type:    ConvertType(t.Underlying()),
			Methods: methods,
		}
	}
	return nil
}

func ConvertFunc(f *types.Func) *gotypes.Func {
	return &gotypes.Func{
		Obj: gotypes.Obj{
			Pkg:  f.Pkg().Path(),
			Name: f.Name(),
			Typ:  ConvertType(f.Type()),
		},
	}
}

func ConvertVar(v *types.Var) *gotypes.Var {
	return &gotypes.Var{
		Obj: gotypes.Obj{
			Pkg:  v.Pkg().Path(),
			Name: v.Name(),
			Typ:  ConvertType(v.Type()),
		},
		Anonymous: v.Anonymous(),
		IsField:   v.IsField(),
	}
}
