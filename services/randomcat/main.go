package randomcat

import (
	"strings"
	"time"

	"net/http"

	"encoding/json"

	"bytes"
	"io/ioutil"

	"github.com/zet4/catsbutnotreally/services"
)

func init() {
	services.Index["randomcat"] = func(arguments json.RawMessage) (filename string, file *bytes.Reader, err error) {
		timeout := time.Duration(60 * time.Second)
		client := http.Client{
			Timeout: timeout,
		}
		resp, err := client.Get("http://random.cat/meow")
		if err != nil {
			return
		}

		var msg struct {
			File string `json:"file"`
		}

		err = json.NewDecoder(resp.Body).Decode(&msg)
		if err != nil {
			return
		}

		err = resp.Body.Close()
		if err != nil {
			return
		}

		resp, err = client.Get(msg.File)
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
