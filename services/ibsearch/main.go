package ibsearch

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/zet4/catsbutnotreally/services"
)

// Config Config from source's optional arguments
type Config struct {
	Query string `json:"query"`
	Key   string `json:"key"`
}

// Image Image from ibsearch api
type Image struct {
	Server string `json:"server"`
	Path   string `json:"path"`
	Tags   string `json:"tags"`
	ID     string `json:"id"`
}

func (img *Image) String() string {
	return fmt.Sprintf("https://%s.ibsear.ch/%s", img.Server, img.Path)
}

var (
	client *http.Client
)

func escapeTags(tags string) string {
	return strings.Replace(tags, "_", "\\_", -1)
}

func boldenMatches(tags string, query []string) string {
	temp := tags
	for _, tag := range query {
		if strings.Contains(tags, tag) {
			temp = strings.Replace(temp, tag, fmt.Sprintf("**%s**", tag), 1)
		}
	}
	return temp
}

func parseQuery(query string) (result []string) {
	for _, s := range strings.Fields(query) {
		// Ignore - tags
		if strings.HasPrefix(s, "-") {
			continue
		}

		// Ignore random and size meta tags.
		if strings.HasPrefix(s, "random") || strings.HasPrefix(s, "size") {
			continue
		}

		parsed := strings.Trim(s, "()|")
		if len(parsed) > 0 {
			result = append(result, parsed)
		}
	}
	return
}

func getIbsearch(host string) func(arguments json.RawMessage) func() (image string, customfields services.CustomFields, err error) {
	return func(arguments json.RawMessage) func() (image string, customfields services.CustomFields, err error) {
		var config Config
		if err := json.Unmarshal([]byte(arguments), &config); err != nil {
			return func() (string, services.CustomFields, error) { return "", nil, err }
		}
		parsedQuery := parseQuery(config.Query)

		return func() (image string, customfields services.CustomFields, err error) {
			req, err := http.NewRequest("GET", fmt.Sprintf("https://%s/api/v1/images.json?q=%s", host, config.Query), nil)
			req.Header.Set("X-IbSearch-Key", config.Key)
			req.Header.Set("User-Agent", fmt.Sprintf("Discord Image Webhook Bot (https://github.com/zet4/catsbutnotreally, v%s)", "0.2"))

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

			imageObj := images[rand.Intn(len(images))]

			fields := make(services.CustomFields)
			fields[fmt.Sprintf("Source: https://%s/images/%s", host, imageObj.ID)] = fmt.Sprintf("Tags: %s", escapeTags(boldenMatches(imageObj.Tags, parsedQuery)))

			return imageObj.String(), fields, nil
		}
	}
}

func init() {
	timeout := time.Duration(20 * time.Second)
	client = &http.Client{
		Timeout: timeout,
	}

	services.Index["ibsearch"] = getIbsearch("ibsear.ch")
	services.Index["ibsearchxxx"] = getIbsearch("ibsearch.xxx")
}
