package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/zet4/catsbutnotreally/services"
)

func postWebhook(url string, displayable services.Displayable) (err error) {
	contenttype, body, err := displayable.GetRequestData()
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", url, body.(io.Reader))
	req.Header.Set("Content-Type", contenttype)
	req.Header.Set("User-Agent", fmt.Sprintf("Discord Image Webhook Bot (https://github.com/zet4/catsbutnotreally, v%s)", VERSION))

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	body = nil

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	resp.Body.Close()
	return nil
}
