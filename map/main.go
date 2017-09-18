package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/ilya-pirogov/sonarmap/map/alivechecker"
	"github.com/ilya-pirogov/sonarmap/map/config"
	"github.com/ilya-pirogov/sonarmap/map/flushcache"
	"github.com/ilya-pirogov/sonarmap/map/nmea-handler"
	"github.com/ilya-pirogov/sonarmap/map/rpc"
	"github.com/ilya-pirogov/sonarmap/map/sdcard"
)

var currentSd = sdcard.New(config.Current.SCid)

var logger = log.New(os.Stdout, "SonarMap [main]: ", log.LstdFlags|log.LUTC)

func main() {
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, os.Interrupt)

	go nmea_handler.StartNmeaHandling()
	go currentSd.Watch()
	go flushcache.StartWatch(currentSd)
	go alivechecker.StartWatch(currentSd)
	go rpc.Start(currentSd)

	for {
		select {
		case <-sigC:
			println("\rInterrupted. Exiting...")
			return
		}
	}
}
