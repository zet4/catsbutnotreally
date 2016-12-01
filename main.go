package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"math/rand"

	"github.com/robfig/cron"

	"github.com/zet4/catsbutnotreally/services"

	_ "github.com/zet4/catsbutnotreally/services/ibsearch"
	_ "github.com/zet4/catsbutnotreally/services/randomcat"

	"net/http"
	_ "net/http/pprof"
)

const (
	VERSION = "0.1"
)

var (
	config Config
)

func init() {
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Fatalln("Error opening config file", err.Error())
	}

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&config); err != nil {
		log.Fatalln("Error parsing config file", err.Error())
	}
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

		filename, file, err := service(source.OptionalArguments)
		if err != nil {
			log.Printf("Error occurred while trying to run '%s': %s\n", source.Service, err.Error())
			return
		}
		message := Message{
			File:     file,
			Filename: filename,
		}
		if d.Avatar != nil {
			message.Avatar = *d.Avatar
		}
		if d.Username != nil {
			message.Username = *d.Username
		}

		err = postWebhook(d.Webhook, message)
		if err != nil {
			log.Printf("Error occurred while trying to send image retrieved from %s: %s", source.Service, err.Error())
			return
		}

		log.Printf("Sending a picture from '%s': %s", source.Service, filename)
	}
}

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	c := cron.New()
	for _, v := range config.Destinations {
		c.AddFunc(v.Cron, work(v))
	}
	c.Start()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
