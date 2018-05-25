package messages

import (
	"bytes"
	"encoding/gob"

	"github.com/dave/frizz/gotypes"
)

type Message interface{}

type Payload struct {
	Message Message
}

func init() {
	// Progress messages:
	gob.Register(Queueing{})
	gob.Register(Downloading{})

	// Data messages:
	gob.Register(Error{})
	gob.Register(GetComplete{})
	gob.Register(TypesComplete{})

	// Commands:
	gob.Register(Types{})

	// Initialise types in gotypes
	gotypes.RegisterTypes()
}

type Queueing struct {
	Position int
	Done     bool
}
type Downloading struct {
	Starting bool
	Message  string
	Done     bool
}
type Error struct {
	Message string
}
type Types struct {
	Path string
}
type GetComplete struct {
	Source map[string]map[string]string
}
type TypesComplete struct {
	Types []gotypes.Named
}

func Marshal(in Message) ([]byte, error) {
	p := Payload{in}
	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode(p); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Unmarshal(in []byte) (Message, error) {
	var p Payload
	if err := gob.NewDecoder(bytes.NewBuffer(in)).Decode(&p); err != nil {
		return nil, err
	}
	return p.Message, nil
}
