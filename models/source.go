package models

import "github.com/dave/jsgo/server/frizz/gotypes"

// SourcePack is the structure of the json persisted to the src.jsgo.io bucket, so use json tags to get
// lower case field names.
type SourcePack struct {
	Path  string            `json:"path"`
	Files map[string]string `json:"files"`
}

type ObjectPack struct {
	Path    string                               // package path
	Name    string                               // package name
	Objects map[string]map[string]gotypes.Object // filename->name->object
}
