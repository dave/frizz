package data

import (
	"go/parser"
	"go/token"
	"testing"

	"bytes"
	"encoding/gob"
	"fmt"
	"go/ast"
	"go/types"
)

func TestParse(t *testing.T) {
	path := "foo"
	fset := token.NewFileSet()
	code := "package foo\n\ntype Foo int64\nvar Bar Foo = 1\n"
	f, err := parser.ParseFile(fset, "foo.go", []byte(code), 0)
	if err != nil {
		t.Fatal(err)
	}
	//c := loader.Config{Fset: fset}
	//c.CreateFromFiles("foo", f)
	//p, err := c.Load()
	//if err != nil {
	//	t.Fatal(f)
	//}
	//fmt.Printf("%#v\n", p)
	tc := types.Config{}
	info := &types.Info{
		Types: map[ast.Expr]types.TypeAndValue{},
	}
	p, err := tc.Check(path, fset, []*ast.File{f}, info)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v\n%#v\n", p, info)
	for k, v := range info.Types {
		switch k := k.(type) {
		case *ast.Ident:
			fmt.Println(k.Name, v.Type, v.Value)
		case *ast.BasicLit:
			fmt.Println(k.Value, k.Kind, v.Type, v.Value)
		}
	}
	buf := &bytes.Buffer{}
	err = gob.NewEncoder(buf).Encode(info)
	if err != nil {
		t.Fatal(err)
	}
}
