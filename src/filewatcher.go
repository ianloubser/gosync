package main

import (
	"crypto/md5"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
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
func getMD5(filePath string) (*FileHash, error) {
	var fileHash = FileHash{}

	f, err := os.Open(filePath)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	fileInfo, _ := f.Stat()
	fileHash.Filename = fileInfo.Name()
	fileHash.Size = fileInfo.Size()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Println(err)
	}

	fileHash.MD5 = h.Sum(nil)
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

func getCanonicalFileKey(path string) (string, error) {

	path, _ = filepath.Abs(path)
	newPath := strings.Replace(path, "\\", "/", -1)
	dirList := filepath.SplitList(newPath)

	if len(dirList) == 2 {
		path := dirList[1]
		return filepath.FromSlash(path), nil
	}

	return newPath, nil
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
				if !event.FileInfo.IsDir() {
					// This is not a directory, so send the file event to incoming event
					eventPool.incomingEvent <- event
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
