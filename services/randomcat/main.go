package randomcat

import (
	"time"

	"net/http"

	"encoding/json"

	"github.com/zet4/catsbutnotreally/services"
)

func init() {
	services.Index["randomcat"] = func(arguments json.RawMessage) func() (image string, customfields services.CustomFields, err error) {
		return func() (image string, customfields services.CustomFields, err error) {
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
			fields := make(services.CustomFields)

			return msg.File, fields, nil
		}
	}
}
