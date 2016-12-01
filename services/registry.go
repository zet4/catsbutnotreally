package services

import "bytes"
import "encoding/json"

var (
	// Index Represents a list of services
	Index = make(map[string]func(arguments json.RawMessage) (filename string, file *bytes.Reader, err error))
)
