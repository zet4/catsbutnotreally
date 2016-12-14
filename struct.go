package main

import "encoding/json"

// Config Base structure of a json config file.
type Config struct {
	Destinations []*Destination `json:"destinations"`

	WebAddress       string `json:"web_address"`
	EnableStatistics bool   `json:"enable_statistics"`
	EnablePPRof      bool   `json:"enable_pprof"`
}

// Destination Defines a single webhook destination's configuration.
type Destination struct {
	Name        string `json:"name"`
	Description string `json:"description"`

	Webhook  string    `json:"webhook"`
	Username *string   `json:"username"`
	Avatar   *string   `json:"avatar"`
	Cron     string    `json:"cron"`
	Sources  []*Source `json:"sources"`
}

//Source Defines a single source's configuration.
type Source struct {
	Name        string `json:"name"`
	Description string `json:"description"`

	Service string `json:"service"`
	Chance  int    `json:"chance"`
	Display string `json:"display"`

	OptionalArguments json.RawMessage `json:"arguments"`
}
