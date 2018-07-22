package sse

import (
	"fmt"
	"net/http"
	"encoding/json"
	"github.com/wordwizzard/playin_go_web/tag"
)

type Broker struct {
	Notifier 		chan []byte				// events are pushed here as by the gathering routine
	newClients		chan chan []byte		// registers the newly open HTTP clients
	closeClients	chan chan []byte		// closes connections of clients when notified
	clients 		map[chan []byte]bool	// holds all currently open channels
}

// Broker Factory
func NewServer() (broker *Broker) {
	// Instantiate a broker service, should spawn a new ServeHTTP() connection for each new connection
	broker = &Broker{
		Notifier:			make(chan [] byte, 1),
		newClients:    	 	make(chan chan [] byte),
		closeClients:		make(chan chan [] byte),
		clients:			make(map[chan [] byte] bool),
	}

	// Set it running listening and broadcasting events
	go broker.listen()
	return
}

// Implements the http.Handler interface it can be passed ro http.ListenAndServe to start listening on a network address
func (broker *Broker) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	// Always flush the flusher first as this will block on all channels if not open to new data
	flusher, ok := rw.(http.Flusher)
	if !ok {
		http.Error(rw, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Each connection registers its own message channel with the broker's connection registry
	messageChannel := make(chan []byte, 2)	// channel pip hold a max of 2 messages that can be flushed -> redundant as we ignore any above 1 later on...
	// signal the broker that we have a new connection
	broker.newClients <- messageChannel

	// notify the broker if our connection dies for some reason - catch failed connections
	defer func() {
		// remove the client form the connected clients when the handler exits
		broker.closeClients <- messageChannel
	}()

	notify := rw.(http.CloseNotifier).CloseNotify()
	go func() {
		<- notify
		broker.closeClients <- messageChannel
		tag.Info("HTTP connection just closed.")
	}()

	// Set for keep alive purposes and various client browser types
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	// This will cause a block on waiting messages broadcast on this connections messageChannel.
	for {
		message, open:= <- messageChannel

		if !open {
			break
		}
		// write to the ResponseWriter - must be Server sent events compatible
		fmt.Fprintf(rw, "data: %s\n\n", message)

		// flush the data immediately
		flusher.Flush()
	}
	// This loop will block on messageChannel waiting for the next event.
	// The never ending loop will keep the handler open and listening until the client closes the connection

}

// Listening method
func (broker *Broker) listen() {
	for {
		tag.Info("Listening for events...")
		select {

		case s := <-broker.newClients:

			tag.Info("New Client")
			broker.clients[s] = true

			// Handshake or Immediate client event created
			//broker.PushMessage(/* Handshake message can be included here */)

			tag.Info(fmt.Sprintf("New Client Served. %d Registered Clients", len(broker.clients)))

		case s := <-broker.closeClients:

			tag.Info("Close Client")
			delete(broker.clients, s)
			close(s)
			tag.Info(fmt.Sprintf("Client Removed. %d Registered Clients", len(broker.clients)))

		case event := <-broker.Notifier:

			tag.Info("New Event")

			for clientMessageChannel := range broker.clients {
				if len(clientMessageChannel) >= 1 {		// will block on spesific channel if there was a pushed message that was not flushed. -- Basically insurance against a web server crash.
					tag.Info(fmt.Sprintf("MessageChannel blocked. Current: %d  Maximum: %d", len(clientMessageChannel), cap(clientMessageChannel)))
					broker.closeClients <- clientMessageChannel // close the disappeared client
				} else {
					tag.Info(fmt.Sprintf("MessageChannel Open. Current: %d Maximum: %d", len(clientMessageChannel), cap(clientMessageChannel)))
					clientMessageChannel <- event	// push the message event to the visible clients.
				} // end if-else statement
			}	// end the for loop
			tag.Info(fmt.Sprintf("Broadcast Served %d Clients", len(broker.clients)))

		}	// end the select
	}	// end the for
}	// end the function

// Json Message Push Service
func (broker *Broker) PushMessage(message *[]byte) {
	jsonMessage, _ := json.Marshal(message)
	broker.Notifier <- []byte(jsonMessage)
}