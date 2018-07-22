package main

import (
	"github.com/wordwizzard/playin_go_web/host"
	"github.com/wordwizzard/playin_go_web/tag"
	"os"
	"os/signal"
	"log"
)

// TODO: Look into creating a go function/loop for the web service so we can continue with other tasks from the main loop onset.

// The Web Service routine...
func WebService() {
	tag.Info("Start Web Service")

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	tag.Info("Initialize Interface")
	source := host.Server()
	tag.Info("Interface Initialized")

	<- quit
	if err := source.Shutdown(nil); err != nil {
		log.Fatal("ERROR - Socket Disconnected")
	}
	tag.Info("Web Service Suspended")
}

func main() {
	tag.Info("Start Program")
	WebService()
}
