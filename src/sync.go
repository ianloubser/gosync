package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/radovskyb/watcher"
)

type EventProcessPool struct {
	// store all the events queued for S3 sync to allow for buffering
	incomingEvent chan watcher.Event
	events        []watcher.Event
	delay         *time.Timer
	batchSize     int64
}

var eventPool EventProcessPool
var sync Sync

func configureLogging() (*os.File, error) {
	f, err := os.OpenFile("logfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Errorf("Could not initialise log file: %v", err)
	}
	return f, nil
}

func main() {
	// configure logging and run script
	// f, _ := configureLogging()
	// defer f.Close()
	// log.SetOutput(f)

	// make sure the ouput gets set first. So the log writes to a file.
	log.Println()
	log.Println("Starting gosyncs3...")

	config := readConfig("conf.json")

	// initialize these entries so memory gets preserved
	eventPool.incomingEvent = make(chan watcher.Event)

	// file event found listener
	go func() {
		for {
			newEvent := <-eventPool.incomingEvent
			eventPool.events = append(eventPool.events, newEvent)

			// do this so we can reset the callback
			if eventPool.delay != nil {
				eventPool.delay.Stop()
			}

			eventPool.delay = time.AfterFunc(time.Second*4, func() {
				// TODO: implement call to sync here
				log.Println("4 seconds have passed since last queue update")
			})

			// when a new fileEvent is found that should be synced, check pool size whether max is reached
			if len(eventPool.events) >= config.BatchSyncSize {
				// stop the sync timer callback
				eventPool.delay.Stop()

				// syncBatch := eventPool.events[0:config.BatchSyncSize]
				log.Println("Sync pool filled!!", newEvent)
				log.Println("Configured to batch sizes of", config.BatchSyncSize)
				if len(eventPool.events) > 0 {
					eventPool.events = eventPool.events[config.BatchSyncSize-1 : len(eventPool.events)-1]
				} else {
					eventPool.events = eventPool.events[0:0]
				}
			}
		}
	}()

	// this is the upload thread
	go func() {
		for {
			if len(sync.queue) > 0 {
				log.Println("Do some sync yo!!!")
			}
		}
	}()

	filewatcher(&config)
}
