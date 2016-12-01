package main

import "encoding/json"

// Config Base structure of a json config file.
type Config struct {
	Destinations []*Destination `json:"destinations"`
}

// Destination Defines a single webhook destination's configuration.
type Destination struct {
	Webhook  string    `json:"webhook"`
	Username *string   `json:"username"`
	Avatar   *string   `json:"avatar"`
	Cron     string    `json:"cron"`
	Sources  []*Source `json:"sources"`
}

//Source Defines a single source's configuration.
type Source struct {
	Service string `json:"service"`
	Chance  int    `json:"chance"`
	Display string `json:"display"`

	OptionalArguments json.RawMessage `json:"arguments"`
}
