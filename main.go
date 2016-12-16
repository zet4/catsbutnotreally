package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"math/rand"

	"github.com/robfig/cron"

	"github.com/zet4/catsbutnotreally/services"

	_ "github.com/zet4/catsbutnotreally/services/ibsearch"
	_ "github.com/zet4/catsbutnotreally/services/randomcat"

	"fmt"
	"net/http"
	"time"
)

const (
	// VERSION Version of the project
	VERSION = "0.4"
)

var (
	client = &http.Client{Timeout: 60 * time.Second}
	runner *cron.Cron

	webapp *WebApp
)

var (
	// Flags
	configFile = ""
)

func flags() {
	flag.StringVar(&configFile, "config", "config.json", "Config file for catsbutnotreally")
	flag.Parse()
}

func main() {
	flags()
	reloadChan, reloadedChan := WatchConfig(configFile)

	go func() {
		var err error
		for {
			<-reloadChan
			if runner != nil {
				runner.Stop()
				runner = nil
			}
			if webapp != nil {
				webapp.Close()
			}
			runner = cron.New()
			for _, v := range config.Destinations {
				err = runner.AddFunc(v.Cron, work(v))
				if err != nil {
					log.Println("Error occurred while trying to add destination:", err.Error())
				}
			}
			runner.Start()
			webapp = WebAppFromConfig(config)
			reloadedChan <- struct{}{}
		}
	}()

	reloadChan <- struct{}{}
	<-reloadedChan

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func work(d *Destination) func() {
	var pool []int
	var total int
	sourceServices := make(map[*Source]func() (image string, customfields services.CustomFields, err error), 0)
	for idx, v := range d.Sources {
		for x := 0; x < v.Chance; x++ {
			pool = append(pool, idx)
		}
		total += v.Chance

		service, ok := services.Index[v.Service]
		if !ok {
			log.Printf("Service '%s' is not registred/available.\n", v.Service)
			continue
		}
		sourceServices[v] = service(v.OptionalArguments)
	}

	return func() {
		source := d.Sources[pool[rand.Intn(total)]]
		service, ok := sourceServices[source]
		if !ok {
			return
		}

		image, fields, err := service()
		if err != nil {
			log.Printf("Error occurred while trying to run '%s': %s\n", source.Service, err.Error())
			return
		}

		message, err := createMessage(*d.Username, *d.Avatar, source.Service, image, source.Display, &fields)
		if err != nil {
			log.Printf("Error occurred while preparing message: %s", err.Error())
			return
		}

		err = postWebhook(d.Webhook, message)
		if err != nil {
			log.Printf("Error occurred while trying to send image retrieved from %s: %s", source.Service, err.Error())
			return
		}

		if webapp != nil {
			webapp.Add(d, source, image)
		}

		log.Printf("Sending a picture from '%s': %s", source.Service, image)
	}
}

func createMessage(username, avatar, service, image, display string, fields *services.CustomFields) (result services.Displayable, err error) {
	if display == "simple" || display == "" {
		resp, err := client.Get(image)
		if err != nil {
			return nil, err
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		tokens := strings.Split(resp.Request.URL.String(), "/")

		if err = resp.Body.Close(); err != nil {
			return nil, err
		}

		return &services.Simple{
			Username: username,
			Avatar:   avatar,
			File:     bytes.NewReader(data),
			Filename: tokens[len(tokens)-1],
		}, nil
	} else if display == "embed" {
		return &services.Embedded{
			Username: username,
			Avatar:   avatar,
			Image:    image,
			Fields:   fields,
		}, nil
	}

	return nil, fmt.Errorf("'%s' is not a valid display type", display)
}
