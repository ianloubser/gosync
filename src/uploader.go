package main

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
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
			var fileMd5 *string
			fileMd5 = &event.fileMD5
			metadata := map[string]*string{
				"MD5": fileMd5,
			}

			uploadObjects[i] = s3manager.BatchUploadObject{
				Object: &s3manager.UploadInput{
					Bucket:   aws.String(config.BucketName),
					Key:      aws.String(event.fileKey),
					Body:     file,
					Metadata: metadata,
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

	log.Println("Loaded files into memory, starting upload")
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
		Region:           aws.String(config.BucketRegion),
		Endpoint:         aws.String(config.BucketEndpoint),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      creds,
	})

	fileInput := &s3.HeadObjectInput{
		Bucket: aws.String(config.BucketName),
		Key:    aws.String(canonicalPath),
	}

	svc := s3.New(sess)
	result, err := svc.HeadObject(fileInput)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NotFound":
				return false, nil
			default:
				log.Printf("Error Code: %s, Message: %s", aerr.Code(), aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Printf("Unknown AWS SDK error: %s", err.Error())
		}
		return false, nil
	}

	if md5Hash == *result.Metadata["Md5"] {
		return true, nil
	}

	return false, nil
}
