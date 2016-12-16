package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/pprof"
	"sync/atomic"
	"time"

	"net"

	"sync"

	rice "github.com/GeertJohan/go.rice"
	"github.com/zet4/catsbutnotreally/utils"
)

type Drop struct {
	Value       uint64 `json:"value"`
	Description string `json:"description,omitempty"`
	LastImage   string `json:"last_image,omitempty"`
}

// Bucket Container for per destination count of images
type Bucket struct {
	Total       uint64 `json:"total"`
	Unknown     uint64 `json:"unknown"`
	Description string `json:"description,omitempty"`
	LastImage   string `json:"last_image,omitempty"`

	Sources map[string]*Drop `json:"sources"`
}

// Statistics Container for global count of images
type Statistics struct {
	sync.RWMutex

	Total     uint64 `json:"total"`
	Unknown   uint64 `json:"unknown"`
	LastImage string `json:"last_image,omitempty"`

	Buckets map[string]*Bucket `json:"destinations"`
}

// Add Appends a stats count atomicly
func (s *Statistics) Add(dest *Destination, source *Source, image string) {
	s.Lock()
	defer s.Unlock()

	atomic.AddUint64(&s.Total, 1)
	s.LastImage = image

	b, ok := s.Buckets[dest.Name]
	if !ok {
		atomic.AddUint64(&s.Unknown, 1)
		return
	}
	atomic.AddUint64(&b.Total, 1)
	b.LastImage = image

	ss, ok := b.Sources[source.Name]
	if !ok {
		atomic.AddUint64(&b.Unknown, 1)
		return
	}
	atomic.AddUint64(&ss.Value, 1)
	ss.LastImage = image
}

// StatisticsFromConfig construct Stats object from config file
func StatisticsFromConfig(cfg *Config) *Statistics {
	stats := Statistics{Buckets: make(map[string]*Bucket)}
	for _, bucket := range cfg.Destinations {
		if bucket.Name != "" {
			temp := Bucket{Sources: make(map[string]*Drop), Description: bucket.Description}
			for _, source := range bucket.Sources {
				if source.Name != "" {
					temp.Sources[source.Name] = &Drop{Description: source.Description}
				}
			}
			stats.Buckets[bucket.Name] = &temp
		}
	}
	return &stats
}

type WebApp struct {
	Stats    *Statistics
	Serve    *utils.StoppableListener
	Broker   *utils.Broker
	Listener net.Listener
}

func (w *WebApp) Close() {
	w.Serve.Stop <- true
	w.Broker.Stop()

	var alive int
	/* Wait at most 5 seconds for the clients to disconnect */
	for i := 0; i < 2; i++ {
		/* Get the number of clients still connected */
		alive = w.Serve.ConnCount.Get()
		if alive == 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}

	alive = w.Serve.ConnCount.Get()
	if alive > 0 {
		log.Printf("Server stopped after 2 seconds with %d client(s) still connected.", alive)
	} else {
		log.Println("Server stopped gracefully.")
	}

	w.Listener.Close()

	w.Serve = nil
	w.Stats = nil
	w.Broker = nil
	w.Listener = nil
}

func WebAppFromConfig(cfg *Config) *WebApp {
	if config.WebAddress != "" && (config.EnableStatistics || config.EnablePPRof) {
		var err error
		webapp = &WebApp{}

		mux := http.NewServeMux()
		if config.EnableStatistics {
			webapp.Stats = StatisticsFromConfig(config)
			webapp.Broker = utils.NewBroker([]byte("stats"), func() []byte { b, _ := json.Marshal(webapp.Stats); return b })
			mux.Handle("/events", webapp.Broker)
			mux.Handle("/", http.FileServer(rice.MustFindBox("web").HTTPBox()))
			mux.Handle("/go", http.HandlerFunc(utils.GoStatisticsHandler))
		}
		if config.EnablePPRof {
			mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
			mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
			mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
			mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
			mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
		}

		webapp.Listener, err = net.Listen("tcp", config.WebAddress)
		if err != nil {
			log.Printf("Error occurred while tryin to listen on '%s': %s\n", config.WebAddress, err.Error())
		}
		webapp.Serve = utils.Handle(webapp.Listener)

		go func() {
			http.Serve(webapp.Serve, mux)
		}()

		return webapp
	}
	return nil
}

// Add increments statistics and broadcasts to all clients.
func (w *WebApp) Add(dest *Destination, source *Source, image string) {
	w.Stats.Add(dest, source, image)
	b, _ := json.Marshal(w.Stats)

	w.Broker.Notifier <- b
}
