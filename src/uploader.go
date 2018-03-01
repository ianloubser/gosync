package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func removeFiles(config *Configuration, events []FileSync) {
	creds := credentials.NewStaticCredentials(config.AccessKeyID, config.SecretAccessKey, "")

	removeObjects := make([]s3manager.BatchDeleteObject, len(events))

	for i, event := range events {
		removeObjects[i] = s3manager.BatchDeleteObject{
			Object: &s3.DeleteObjectInput{
				Bucket: aws.String(config.BucketName),
				Key:    aws.String(event.fileKey),
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

// The does the actual batch upload of the file
func uploadFiles(config *Configuration, events []FileSync) {
	creds := credentials.NewStaticCredentials(config.AccessKeyID, config.SecretAccessKey, "")

	uploadObjects := make([]s3manager.BatchUploadObject, len(events))

	for i, event := range events {
		file, err := os.Open(event.fileEvent.Path)
		if err != nil {
			log.Printf("Could not load the file to upload, %s", event.fileEvent.Path)
		} else {
			// sanitize the path to a valid key descriptor
			uploadObjects[i] = s3manager.BatchUploadObject{
				Object: &s3manager.UploadInput{
					Bucket:  aws.String(config.BucketName),
					Key:     aws.String(event.fileKey),
					Body:    file,
					Tagging: aws.String(event.fileMD5),
				},
			}
			defer file.Close()
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
		if multierr, ok := err.(s3manager.MultiUploadFailure); ok {
			// Process error and its associated uploadID
			log.Println("Error:", multierr.Code(), multierr.Message(), multierr.UploadID())
		} else {
			// Process error generically
			log.Println("Error:", err.Error())
		}
	}
}

func existsOnS3(config *Configuration, canonicalPath string, md5Hash string) (bool, error) {
	creds := credentials.NewStaticCredentials(config.AccessKeyID, config.SecretAccessKey, "")

	sess := session.New(&aws.Config{
		Region:      aws.String(config.BucketRegion),
		Endpoint:    aws.String(config.BucketEndpoint),
		Credentials: creds,
	})

	fileInput := &s3.HeadObjectInput{
		Bucket: aws.String(config.BucketName),
		Key:    aws.String(canonicalPath),
	}

	svc := s3.New(sess)

	log.Println("Make HEAD call")

	// result, err := svc.HeadObject(fileInput)
	req, resp := svc.HeadObjectRequest(fileInput)

	err := req.Send()
	if err == nil { // resp is now filled
		fmt.Println(resp)
	}

	if err != nil {
		log.Println("Some Issue here")
	}

	log.Printf("Checked on s3 for %s", canonicalPath)
	log.Printf("Result: %s", resp)

	return false, nil
}
