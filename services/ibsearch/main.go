package ibsearch

import (
	"strings"
	"time"

	"net/http"

	"encoding/json"

	"bytes"
	"io/ioutil"

	"fmt"

	"math/rand"

	"github.com/zet4/catsbutnotreally/services"
)

type Config struct {
	Query string `json:"query"`
	Key   string `json:"key"`
}

type Image struct {
	Server string `json:"server"`
	Path   string `json:"path"`
}

func (img *Image) String() string {
	return fmt.Sprintf("https://%s.ibsear.ch/%s", img.Server, img.Path)
}

func init() {
	services.Index["ibsearch"] = func(arguments json.RawMessage) (filename string, file *bytes.Reader, err error) {
		var config Config
		if err := json.Unmarshal([]byte(arguments), &config); err != nil {
			return "", nil, err
		}

		timeout := time.Duration(60 * time.Second)
		client := http.Client{
			Timeout: timeout,
		}

		req, err := http.NewRequest("GET", fmt.Sprintf("https://ibsear.ch/api/v1/images.json?q=%s", config.Query), nil)
		req.Header.Set("X-IbSearch-Key", config.Key)
		req.Header.Set("User-Agent", fmt.Sprintf("Discord Image Webhook Bot (https://github.com/zet4/catsbutnotreally, v%s)", "0.1"))

		resp, err := client.Do(req)
		if err != nil {
			return
		}

		var images []Image

		if err = json.NewDecoder(resp.Body).Decode(&images); err != nil {
			return
		}

		if err = resp.Body.Close(); err != nil {
			return
		}

		if len(images) == 0 {
			return "", nil, fmt.Errorf("Result array of images for '%s' is empty", config.Query)
		}

		resp, err = client.Get(images[rand.Intn(len(images))].String())
		if err != nil {
			return
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}

		tokens := strings.Split(resp.Request.URL.String(), "/")

		if err = resp.Body.Close(); err != nil {
			return
		}

		filename = tokens[len(tokens)-1]
		file = bytes.NewReader(data)

		return
	}
}
