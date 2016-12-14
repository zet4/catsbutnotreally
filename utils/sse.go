package utils

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"
)

const patience time.Duration = time.Second * 1

var reloadMsg = []byte("reload")

// Broker is responsible for keeping a list of which clients (browsers) are currently attached
// and broadcasting events (messages) to those clients.
type Broker struct {
	// Events are pushed to this channel by the main events-gathering routine
	Notifier chan []byte

	// New client connections
	newClients chan chan []byte

	// Closed client connections
	closingClients chan chan []byte

	// Client connections registry
	clients map[chan []byte]bool

	// Event name
	Event []byte

	// Initial data
	Init func() []byte
}

// Start creates a goroutine loop that keeps server side events running.
func (broker *Broker) Start() {
	for {
		select {
		case s := <-broker.newClients:

			// A new client has connected.
			// Register their message channel
			broker.clients[s] = true
			log.Printf("Client added. %d registered clients", len(broker.clients))
		case s := <-broker.closingClients:

			// A client has dettached and we want to
			// stop sending them messages.
			delete(broker.clients, s)
			log.Printf("Removed client. %d registered clients", len(broker.clients))
		case event := <-broker.Notifier:

			// We got a new event from the outside!
			// Send event to all connected clients
			for clientMessageChan, _ := range broker.clients {
				select {
				case clientMessageChan <- event:
				case <-time.After(patience):
					log.Print("Skipping client.")
				}
			}
		}
	}
}

func (broker *Broker) Stop() {
	broker.Notifier <- reloadMsg
}

func (broker *Broker) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// Make sure that the writer supports flushing.
	//
	flusher, ok := rw.(http.Flusher)

	if !ok {
		http.Error(rw, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	// Each connection registers its own message channel with the Broker's connections registry
	messageChan := make(chan []byte)

	// If initial data is supplied send that data to new clients.
	if broker.Init != nil {
		go func() {
			messageChan <- broker.Init()
		}()
	}

	// Signal the broker that we have a new connection
	broker.newClients <- messageChan

	// Remove this client from the map of connected clients
	// when this handler exits.
	defer func() {
		broker.closingClients <- messageChan
	}()

	// Listen to connection close and un-register messageChan
	notify := rw.(http.CloseNotifier).CloseNotify()

	for {
		select {
		case <-notify:
			return
		case message := <-messageChan:
			// Write to the ResponseWriter
			// Server Sent Events compatible
			fmt.Fprintf(rw, "event: %s\ndata: %s\n\n", broker.Event, message)

			// Flush the data immediatly instead of buffering it for later.
			flusher.Flush()
			if bytes.Equal(message, reloadMsg) {
				return
			}
		}
	}
}

// NewBroker Creates a new instance of broker.
func NewBroker(event []byte, init func() []byte) *Broker {
	// Make a new Broker instance
	b := &Broker{
		Notifier:       make(chan []byte, 1),
		newClients:     make(chan chan []byte),
		closingClients: make(chan chan []byte),
		clients:        make(map[chan []byte]bool),
		Event:          event,
		Init:           init,
	}

	// Start processing events
	go b.Start()
	return b
}
