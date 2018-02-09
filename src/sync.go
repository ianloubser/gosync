package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/radovskyb/watcher"
)

var syncPool SyncPool

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
	log.Println("\nStarting gosyncs3...")
	config := readConfig("conf.json")

	// initialize these entries so memory gets preserved
	syncPool.incomingEvent = make(chan watcher.Event)

	// file event found listener
	go func() {
		for {
			newEvent := <-syncPool.incomingEvent
			syncPool.queue = append(syncPool.queue, newEvent)

			// do this so we can reset the callback
			if syncPool.delay != nil {
				syncPool.delay.Stop()
			}

			syncPool.delay = time.AfterFunc(time.Second*4, func() {
				// TODO: implement call to sync here
				log.Println("4 seconds have passed since last queue update")
			})

			// when a new fileEvent is found that should be synced, check pool size whether max is reached
			if len(syncPool.queue) >= config.BatchSyncSize {
				// stop the sync timer callback
				syncPool.delay.Stop()

				// syncBatch := syncPool.queue[0:config.BatchSyncSize]
				// go uploadFiles(config, syncBatch)
				log.Println("Sync pool filled!!", newEvent)
				log.Println("Configured to batch sizes of", config.BatchSyncSize)
				if len(syncPool.queue) > 0 {
					syncPool.queue = syncPool.queue[config.BatchSyncSize-1 : len(syncPool.queue)-1]
				} else {
					syncPool.queue = syncPool.queue[0:0]
				}
			}
		}
	}()

	// this is the upload thread
	go func() {
		for {
			// keep on checking for any new upload events that are queued
		}
	}()

	filewatcher(&config)
}
