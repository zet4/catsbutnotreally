package services

import "encoding/json"

// CustomFields Container for custom fields
type CustomFields map[string]string

var (
	// Index Represents a list of services
	Index = make(map[string]func(arguments json.RawMessage) (image string, customfields CustomFields, err error))
)
