package data

import (
	"go/parser"
	"go/token"
	"testing"

	"go/ast"
	"go/types"

	"encoding/json"

	"bytes"
	"fmt"

	"strings"

	"sort"

	"regexp"

	"github.com/dave/frizz/gotypes"
)

func TestConvertType(t *testing.T) {
	type spec struct {
		path     string
		code     string
		expected string
	}
	tests := map[string]spec{
		"simple": {
			"foo",
			`
				package foo
	
				type Foo int64
			`,
			`Foo: *gotypes.Basic: {"Kind":6,"Info":2,"Name":"int64"}`,
		},
		"ignore non-global": {
			"foo",
			`
				package foo
	
				type Foo int64
				
				func f() {
					type Bar int64
				}
			`,
			`Foo: *gotypes.Basic: {"Kind":6,"Info":2,"Name":"int64"}`,
		},
		"two types": {
			"foo",
			`
				package foo
	
				type Foo int64
				type Bar int64
			`,
			`Foo: *gotypes.Basic: {"Kind":6,"Info":2,"Name":"int64"}
			Bar: *gotypes.Basic: {"Kind":6,"Info":2,"Name":"int64"}`,
		},
		"alias": {
			"foo",
			`
				package foo
	
				type Foo int
				type Bar Foo
			`,
			`Foo: *gotypes.Basic: {"Kind":2,"Info":2,"Name":"int"}
			Bar: *gotypes.Basic: {"Kind":2,"Info":2,"Name":"int"}`,
		},
		"complex": {
			"foo",
			`
				package foo
	
				type Foo struct {
					bar string
				}
			`,
			`Foo: *gotypes.Struct: {"Fields":[{"Pkg":"foo","Name":"bar","Typ":{"Kind":17,"Info":32,"Name":"string"},"Anonymous":false,"IsField":true}],"Tags":[""]}`,
		},
	}
	for name, test := range tests {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, "foo.go", []byte(test.code), 0)
		if err != nil {
			t.Fatal(err)
		}
		tc := types.Config{}
		info := &types.Info{
			Types: map[ast.Expr]types.TypeAndValue{},
			Defs:  map[*ast.Ident]types.Object{},
		}

		p, err := tc.Check(test.path, fset, []*ast.File{f}, info)
		if err != nil {
			t.Fatal(err)
		}

		var defs []*types.Named
		for _, v := range info.Defs {
			if v == nil {
				continue
			}
			if v.Parent() != p.Scope() {
				continue
			}
			tn, ok := v.(*types.TypeName)
			if !ok {
				continue
			}
			n, ok := tn.Type().(*types.Named)
			if !ok {
				t.Fatalf("%s, got %T", name, v)
			}
			defs = append(defs, n)
		}
		sort.Slice(defs, func(i, j int) bool { return defs[i].Obj().Pos() < defs[j].Obj().Pos() })
		var globals []*gotypes.Named
		for _, v := range defs {
			globals = append(globals, ConvertType(v).(*gotypes.Named))
		}
		buf := &bytes.Buffer{}
		for _, g := range globals {
			b, err := json.Marshal(g.Type)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Fprintf(buf, "%s: %T: %s\n", g.Obj.Name, g.Type, string(b))
		}
		if strings.TrimSpace(buf.String()) != indent.ReplaceAllString(test.expected, "") {
			t.Fatalf("%s, got:\n%s", name, buf.String())
		}
	}
}

var indent = regexp.MustCompile(`(?m)^\s*`)
