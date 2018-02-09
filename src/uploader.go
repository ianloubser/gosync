package main

import (
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/radovskyb/watcher"
)

type SyncPool struct {
	// store all the events queued for S3 sync to allow for buffering
	incomingEvent  chan watcher.Event
	queue          []watcher.Event
	lastPoolUpdate time.Time
	delay          *time.Timer
	lastSync       time.Time
	batchSize      int64
	lastSize       int
}

func removeFiles(config *Configuration, events []watcher.Event) {
	creds := credentials.NewStaticCredentials(config.AccessKeyID, config.SecretAccessKey, "")

	removeObjects := make([]s3manager.BatchDeleteObject, config.BatchSyncSize)

	for i, event := range events {
		removeObjects[i] = s3manager.BatchDeleteObject{
			Object: &s3.DeleteObjectInput{
				Bucket: aws.String(config.BucketName),
				Key:    aws.String(event.Path),
			},
		}
	}

	// initialize the session connection
	sess := session.New(&aws.Config{
		Region:           aws.String(config.BucketRegion),
		Endpoint:         aws.String(config.BucketEndpoint),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      creds,
	})

	// start the uploader instance
	purger := s3manager.NewBatchDelete(sess)
	iter := &s3manager.DeleteObjectsIterator{Objects: removeObjects}

	if err := purger.Delete(aws.BackgroundContext(), iter); err != nil {
		log.Panicln("Failed batch delete of objects", err)
	}
}

func uploadFiles(config *Configuration, events []watcher.Event) {
	creds := credentials.NewStaticCredentials(config.AccessKeyID, config.SecretAccessKey, "")

	uploadObjects := make([]s3manager.BatchUploadObject, config.BatchSyncSize)

	for i, event := range events {
		file, err := os.Open(event.Path)
		if err != nil {
			log.Printf("Could not load the file to upload, %s", event.Path)
		} else {
			uploadObjects[i] = s3manager.BatchUploadObject{
				Object: &s3manager.UploadInput{
					Bucket:  aws.String(config.BucketName),
					Key:     aws.String(event.Path),
					Body:    file,
					Tagging: aws.String("some md5 file hash"),
				},
			}
		}
	}

	// initialize the session connection
	sess := session.New(&aws.Config{
		Region:           aws.String(config.BucketRegion),
		Endpoint:         aws.String(config.BucketEndpoint),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      creds,
	})

	// start the uploader instance
	uploader := s3manager.NewUploader(sess)
	iter := &s3manager.UploadObjectsIterator{Objects: uploadObjects}

	if err := uploader.UploadWithIterator(aws.BackgroundContext(), iter); err != nil {
		log.Panicln("Failed batch upload of objects", err)
	}
}

func syncFile(config *Configuration, event watcher.Event) {
	if event.Op == watcher.Rename || event.Op == watcher.Move || event.Op == watcher.Remove {
		// we need some more fanciness reacting on moved and renamed files
	} else if event.Op == watcher.Create || event.Op == watcher.Write {
		// TODO: an md5 check against current database
		// if found, err := exists(event.Path); err == nil && found {
		// fileHash, _ := getFileHash(event.Path)

		syncPool.incomingEvent <- event
	}
}

// fileInput := &s3.GetObjectTaggingInput{
// 	Bucket: aws.String(config.BucketName),
// 	Key:    aws.String(event.Path),
// }

// svc := s3.New(sess)

// result, err := svc.GetObjectTagging(input)
// if err != nil {
// 	log.Println("Some Issue here")
// }
