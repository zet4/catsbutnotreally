package main

import (
	"bytes"
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
	_ "net/http/pprof"
	"time"
)

const (
	// VERSION Version of the project
	VERSION = "0.2"
)

var (
	client = &http.Client{Timeout: 60 * time.Second}
	runner *cron.Cron
)

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
		return &services.Embeded{
			Username: username,
			Avatar:   avatar,
			Image:    image,
			Fields:   fields,
		}, nil
	}

	return nil, fmt.Errorf("'%s' is not a valid display type", display)
}

func work(d *Destination) func() {
	var pool []int
	var total int
	for idx, v := range d.Sources {
		for x := 0; x < v.Chance; x++ {
			pool = append(pool, idx)
		}
		total += v.Chance
	}

	return func() {
		source := d.Sources[pool[rand.Intn(total)]]
		service, ok := services.Index[source.Service]
		if !ok {
			log.Printf("Service '%s' is not registred/available.\n", source.Service)
			return
		}

		image, fields, err := service(source.OptionalArguments)
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

		log.Printf("Sending a picture from '%s': %s", source.Service, image)
	}
}

func main() {
	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	reloadChan := WatchConfig("config.json")

	go func() {
		var err error
		for {
			<-reloadChan
			if runner != nil {
				runner.Stop()
				runner = nil
			}
			runner = cron.New()
			for _, v := range config.Destinations {
				err = runner.AddFunc(v.Cron, work(v))
				if err != nil {
					log.Println("Error occurred while trying to add destination:", err.Error())
				}
			}
			runner.Start()
		}
	}()

	reloadChan <- true

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
