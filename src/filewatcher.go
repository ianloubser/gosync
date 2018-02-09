package main

import (
	"crypto/md5"
	"io"
	"log"
	"math"
	"os"
	"time"

	"github.com/radovskyb/watcher"
)

const fileChunk = 8192

//FileHash struct for MD5's
type FileHash struct {
	Filename string
	Size     int64
	MD5      []byte
}

//Get takes a file and returns the FileHash struct
func getFileHash(fileLocation string) (*FileHash, error) {
	var fileHash = FileHash{}

	file, err := os.Open(fileLocation)
	if err != nil {
		return &fileHash, err
	}

	defer file.Close()

	fileInfo, _ := file.Stat()
	fileHash.Filename = fileInfo.Name()
	fileHash.Size = fileInfo.Size()

	hash := md5.New()

	blocks := uint64(math.Ceil(float64(fileHash.Size) / float64(fileChunk)))

	for i := uint64(0); i < blocks; i++ {
		blocksize := int(math.Min(fileChunk, float64(fileHash.Size-int64(i*fileChunk))))
		buf := make([]byte, blocksize)

		file.Read(buf)
		io.WriteString(hash, string(buf)) // append into the hash
	}

	fileHash.MD5 = hash.Sum(nil)

	// get hex md5 string
	// hex.EncodeToString(fileHash.MD5)

	return &fileHash, nil
}

func exists(path string) (bool, error) {
	// utitlity function used to check whether a specified file or directory exists
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func filewatcher(config *Configuration) {
	w := watcher.New()

	// If SetMaxEvents is not set, the default is to send all events.
	// w.SetMaxEvents(1)

	// Only notify rename and move events.
	// w.FilterOps(watcher.Rename, watcher.Move, watcher.Create, watcher.Remove)

	// setup filewatcher goroutine
	go func() {
		for {
			select {
			case event := <-w.Event:
				// log.Println(event)
				if !event.FileInfo.IsDir() {
					// log.Println("File:", event.Path)
					// log.Println("Calling S3 sync on file", event.Op)
					syncFile(config, event)
				}
			case err := <-w.Error:
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}()

	for _, path := range config.Paths {
		if len(path) > 0 {
			if path[len(path)-1] == '*' {
				log.Printf("Watching a directory recursively: %s", path)
				// Watch test_folder recursively for changes.
				if err := w.AddRecursive(path[0 : len(path)-1]); err != nil {
					log.Printf("Failed adding recursive directory to watch '%s'", err)
				}
			} else {
				log.Printf("Watching a directory/file : %s", path)
				if err := w.Add(path); err != nil {
					log.Printf("Failed adding to watched files/folders '%s'", err)
				}
			}
		} else {
			log.Println("Invalid scan path, skipping")
		}
	}

	if len(w.WatchedFiles()) < 1 {
		log.Println("No directories or files to watch specified, quiting...")
		log.Fatalln("Please set the 'Paths' value in the conf.json file")
	}

	if config.InitialSync {
		log.Println("Performing initial sync step, this might take a while...")
		for path, f := range w.WatchedFiles() {
			// log.Println(path, f.Name())
			// uploadFiles(config, path)
			if !f.IsDir() {
				eventPool.incomingEvent <- watcher.Event{watcher.Create, path, f}
			}
		}
	} else {
		log.Println("Initial sync disabled, skipping this step...")
	}

	// Wait for watcher to start
	go func() {
		w.Wait()
	}()

	// Start the watching process - it'll check for changes every 100ms.
	if err := w.Start(time.Millisecond * (1000 * time.Duration(config.ScanInterval))); err != nil {
		log.Fatalln(err)
	}
}
