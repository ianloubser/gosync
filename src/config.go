package main

import (
	"encoding/json"
	"log"
	"os"
)

type Configuration struct {
	// interesting... the json decoded wont 'export' any values if variable names don't start with upper case
	// e.g. nothing starting with lower case letter will be parse back from json
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	BucketRegion    string
	BucketEndpoint  string
	ScanInterval    int64
	BatchSyncSize   int
	InitialSync     bool
	LogFile         string
	Paths           []string
}

func readConfig(filePath string) Configuration {
	// Read the config file specified by 'filePath', and return the object
	file, readErr := os.Open(filePath)

	// make sure the file open is closed after return
	defer file.Close()

	if readErr != nil {
		log.Fatalf("Failed to read config: '%s'", readErr)
	}
	decoder := json.NewDecoder(file)

	configuration := Configuration{}
	decodeErr := decoder.Decode(&configuration)
	if decodeErr != nil {
		log.Fatalln("error:", decodeErr)
	}

	return configuration
}
