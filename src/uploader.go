package main

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/radovskyb/watcher"
)

// describes what the sync task needs to do in batch
type TaskOperation uint32

const (
	Create TaskOperation = iota
	Delete
)

// this differs from the SyncPool
type Sync struct {
	queue []SyncTask
}

type SyncTask struct {
	taskType   TaskOperation
	eventBatch []watcher.Event
}

func removeFiles(config *Configuration, events []watcher.Event) {
	creds := credentials.NewStaticCredentials(config.AccessKeyID, config.SecretAccessKey, "")

	removeObjects := make([]s3manager.BatchDeleteObject, len(events))

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

	uploadObjects := make([]s3manager.BatchUploadObject, len(events))

	for i, event := range events {
		file, err := os.Open(event.Path)
		if err != nil {
			log.Printf("Could not load the file to upload, %s", event.Path)
		} else {
			// sanitize the path to a valid key descriptor
			// keyStr := generateKeyFromPath(event.Path)
			uploadObjects[i] = s3manager.BatchUploadObject{
				Object: &s3manager.UploadInput{
					Bucket:  aws.String(config.BucketName),
					Key:     aws.String(event.Path),
					Body:    file,
					Tagging: aws.String("some md5 file hash"),
				},
			}
			// defer file.Close()
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
		log.Println("Failed batch upload of objects", err)
	}
}

func existsOnS3(config *Configuration, canonicalPath string, md5Hash string) (bool, error) {
	creds := credentials.NewStaticCredentials(config.AccessKeyID, config.SecretAccessKey, "")
	sess := session.New(&aws.Config{
		Region:           aws.String(config.BucketRegion),
		Endpoint:         aws.String(config.BucketEndpoint),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      creds,
	})

	fileInput := &s3.GetObjectTaggingInput{
		Bucket: aws.String(config.BucketName),
		Key:    aws.String(canonicalPath),
	}

	svc := s3.New(sess)

	result, err := svc.GetObjectTagging(fileInput)
	if err != nil {
		log.Println("Some Issue here")
	}

	log.Printf("Checked on s3 for %s", canonicalPath)
	log.Printf("Result: %s", result)

	return false, nil
}

func syncFile(config *Configuration, event watcher.Event) {
	if event.Op == watcher.Remove { //event.Op == watcher.Rename || event.Op == watcher.Move ||
		// we need some more fanciness reacting on moved and renamed files
		eventPool.incomingEvent <- event
	} else if event.Op == watcher.Create || event.Op == watcher.Write {
		// TODO: an md5 check against current database
		// if found, err := exists(event.Path); err == nil && found {

		// canonicalPath, _ := getCanonicalFileKey(event.Path)
		// fileHash, _ := getFileHash(event.Path)
		// md5Hash := hex.EncodeToString(fileHash.MD5)

		// exist, _ := existsOnS3(config, canonicalPath, md5Hash)
		// if !exist {
		// 	eventPool.incomingEvent <- event
		// 	log.Printf("Could not find file on S3")
		// }

		// we store the md5 hash as the key in cache
		// if _, found := fileCache.Get(md5Hash); found {
		// 	// TODO: we need to do something if the file hash was found
		// 	exist, _ := existsOnS3(config, canonicalPath, md5Hash)
		// 	if !exist {
		// 		eventPool.incomingEvent <- event
		// 	}
		// } else {
		// 	exist, _ := existsOnS3(config, canonicalPath, md5Hash)
		// 	if !exist {
		// 		eventPool.incomingEvent <- event
		// 	}
		// 	fileCache.Set(md5Hash, canonicalPath, cache.DefaultExpiration)
		eventPool.incomingEvent <- event
		// }
	}
}
