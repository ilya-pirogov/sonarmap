package main

import (
    "log"
    "os"
    "os/signal"

    "sonarmap/alivechecker"
    "sonarmap/config"
    "sonarmap/flushcache"
    "sonarmap/rpc"
    "sonarmap/sdcard"
)

var currentSd = sdcard.New(config.Current.SCid)

var logger = log.New(os.Stdout, "SonarMap [main]: ", log.LstdFlags | log.LUTC)

func main() {
    sigC := make(chan os.Signal, 1)
    signal.Notify(sigC, os.Interrupt)

    go currentSd.Watch()
    go flushcache.StartWatch(currentSd)
    go alivechecker.StartWatch(currentSd)
    go rpc.Start(currentSd)

    for {
        select {
        case <- sigC:
            println("\rInterrupted. Exiting...")
            return
        }
    }
}
