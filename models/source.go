package models

// SourcePack is the structure of the json persisted to the src.jsgo.io bucket, so use json tags to get
// lower case field names.
type SourcePack struct {
	Path  string            `json:"path"`
	Files map[string]string `json:"files"`
}
