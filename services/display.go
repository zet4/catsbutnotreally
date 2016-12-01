package services

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"strings"
)

// Displayable Interface for image responses
type Displayable interface {
	GetRequestData() (contenttype string, body interface{}, err error)
}

type Simple struct {
	Displayable

	Username, Avatar, Content, Filename string
	File                                *bytes.Reader
}

func addField(body *multipart.Writer, field, value string) (err error) {
	if value == "" {
		return nil
	}
	writer, err := body.CreateFormField(field)
	if err != nil {
		return
	}
	_, err = io.Copy(writer, strings.NewReader(value))
	if err != nil {
		return
	}
	return nil
}

func (s *Simple) GetRequestData() (contenttype string, body interface{}, err error) {
	bodybuffer := &bytes.Buffer{}
	bodywriter := multipart.NewWriter(bodybuffer)

	addField(bodywriter, "username", s.Username)
	if err != nil {
		return
	}

	addField(bodywriter, "avatar_url", s.Avatar)
	if err != nil {
		return
	}

	writer, err := bodywriter.CreateFormFile("file", s.Filename)
	if err != nil {
		return
	}

	_, err = io.Copy(writer, s.File)
	if err != nil {
		return
	}

	s.File = nil

	err = bodywriter.Close()
	if err != nil {
		return
	}
	return bodywriter.FormDataContentType(), bytes.NewBuffer(bodybuffer.Bytes()), nil
}

type Embeded struct {
	Displayable

	Username, Avatar, Content, Image string

	Fields *CustomFields
}

type field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type embed struct {
	Image struct {
		URL string `json:"url"`
	} `json:"image"`
	Fields []*field `json:"fields"`
}

func (e *Embeded) GetRequestData() (contenttype string, body interface{}, err error) {
	var message struct {
		Username string   `json:"username"`
		Avatar   string   `json:"avatar_url"`
		Content  string   `json:"content"`
		Embeds   []*embed `json:"embeds"`
	}

	message.Username = e.Username

	message.Avatar = e.Avatar

	message.Content = e.Content

	message.Embeds = make([]*embed, 1)

	message.Embeds[0] = &embed{}

	message.Embeds[0].Image.URL = e.Image

	for k, v := range *e.Fields {
		message.Embeds[0].Fields = append(message.Embeds[0].Fields, &field{Name: k, Value: v, Inline: false})
	}

	data, err := json.Marshal(message)

	return "application/json", bytes.NewBuffer(data), nil
}
