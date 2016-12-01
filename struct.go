package main

import "bytes"
import "encoding/json"

type Config struct {
	Destinations []*Destination `json:"destinations"`
}

type Destination struct {
	Webhook  string    `json:"webhook"`
	Username *string   `json:"username"`
	Avatar   *string   `json:"avatar"`
	Cron     string    `json:"cron"`
	Sources  []*Source `json:"sources"`
}

type Source struct {
	Service string `json:"service"`
	Chance  int    `json:"chance"`

	OptionalArguments json.RawMessage `json:"arguments"`
}

type Message struct {
	Username string
	Avatar   string
	Content  string
	Filename string
	File     *bytes.Reader
}
