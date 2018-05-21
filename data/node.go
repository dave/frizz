package data

type Node struct {
	Type   Type
	Value  interface{}
	Fields map[string]*Node
	Map    map[interface{}]*Node
	Slice  []*Node
}

type Type interface {
	isType()
}

type Named struct {
	Package string
	Name    string
	Type    Type
}

type Struct struct {
	Fields []Field
}

type Field struct {
	Name string
	Type Type
	Tag  string
}

type Map struct {
	Key   Type
	Value Type
}

type Slice struct {
	Value Type
}

type Array struct {
	Length uint64
	Value  Type
}

type Alias struct {
	Alias Type
}

type Pointer struct {
	Type Type
}

type Builtin int

const (
	Bool Builtin = iota
	Uint8
	Uint16
	Uint32
	Uint64
	Int8
	Int16
	Int32
	Int64
	Float32
	Float64
	Complex64
	Complex128
	String
	Int
	Uint
	Uintptr
	Byte
	Rune
)

func (Named) isType()   {}
func (Struct) isType()  {}
func (Map) isType()     {}
func (Slice) isType()   {}
func (Array) isType()   {}
func (Alias) isType()   {}
func (Pointer) isType() {}
func (Builtin) isType() {}
