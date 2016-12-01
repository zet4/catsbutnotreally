package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

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

func postWebhook(url string, message Message) (err error) {
	body := &bytes.Buffer{}
	bodywriter := multipart.NewWriter(body)

	addField(bodywriter, "username", message.Username)
	if err != nil {
		return
	}

	addField(bodywriter, "avatar_url", message.Avatar)
	if err != nil {
		return
	}

	addField(bodywriter, "content", message.Content)
	if err != nil {
		return
	}

	writer, err := bodywriter.CreateFormFile("file", message.Filename)
	if err != nil {
		return
	}

	_, err = io.Copy(writer, message.File)
	if err != nil {
		return
	}

	message.File = nil

	err = bodywriter.Close()
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body.Bytes()))
	req.Header.Set("Content-Type", bodywriter.FormDataContentType())
	req.Header.Set("User-Agent", fmt.Sprintf("Discord Image Webhook Bot (https://github.com/zet4/catsbutnotreally, v%s)", VERSION))

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	body = nil
	ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return nil
}
