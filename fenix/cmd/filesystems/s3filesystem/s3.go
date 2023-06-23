package s3filesystem

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/wtran29/fenix/fenix/cmd/filesystems"
)

type S3 struct {
	Key      string
	Secret   string
	Region   string
	Endpoint string
	Bucket   string
}

func (s *S3) getCredentials() *credentials.Credentials {
	client := credentials.NewStaticCredentials(s.Key, s.Secret, "")
	return client
}

func (s *S3) Put(filename, folder string) error {
	client := s.getCredentials()
	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint:    &s.Endpoint,
		Region:      &s.Region,
		Credentials: client,
	}))

	uploader := s3manager.NewUploader(sess)

	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	fileInfo, err := f.Stat()
	if err != nil {
		return err
	}

	var size = fileInfo.Size()

	buffer := make([]byte, size)
	_, err = f.Read(buffer)
	if err != nil {
		return err
	}

	fileBytes := bytes.NewReader(buffer)
	fileType := http.DetectContentType(buffer)

	log.Println("Content-Type", fileType)
	log.Println("Key", fmt.Sprintf("%s/%s", folder, path.Base(filename)))

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(s.Bucket),
		Key:         aws.String(fmt.Sprintf("%s/%s", folder, path.Base(filename))),
		Body:        fileBytes,
		ACL:         aws.String("public-read"),
		ContentType: aws.String(fileType),
		Metadata: map[string]*string{
			"Key": aws.String("MetadataValue"),
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *S3) List(prefix string) ([]filesystems.Listing, error) {
	var listing []filesystems.Listing

	if prefix == "/" {
		prefix = ""
	}
	client := s.getCredentials()
	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint:    &s.Endpoint,
		Region:      &s.Region,
		Credentials: client,
	}))
	svc := s3.New(sess)
	input := &s3.ListObjectsInput{
		Bucket: aws.String(s.Bucket),
		Prefix: aws.String(prefix),
	}

	result, err := svc.ListObjects(input)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case s3.ErrCodeNoSuchBucket:
				fmt.Println(s3.ErrCodeNoSuchBucket, awsErr.Error())
			default:
				fmt.Println(awsErr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return nil, err
	}

	for _, v := range result.Contents {
		b := float64(*v.Size)
		kb := b / 1024
		mb := kb / 1024
		current := filesystems.Listing{
			Etag:         *v.ETag,
			LastModified: *v.LastModified,
			Key:          *v.Key,
			Size:         mb,
		}
		listing = append(listing, current)
	}
	return listing, nil
}

func (s *S3) Delete(itemsToDel []string) bool {
	client := s.getCredentials()
	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint:    &s.Endpoint,
		Region:      &s.Region,
		Credentials: client,
	}))
	svc := s3.New(sess)

	for _, item := range itemsToDel {
		input := &s3.DeleteObjectsInput{
			Bucket: aws.String(s.Bucket),
			Delete: &s3.Delete{
				Objects: []*s3.ObjectIdentifier{
					{
						Key: aws.String(item),
					},
				},
				Quiet: aws.Bool(false),
			},
		}

		_, err := svc.DeleteObjects(input)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				switch awsErr.Code() {
				default:
					fmt.Println("Amazon delete error:", awsErr.Error())
					return false
				}
			} else {
				fmt.Println("Other delete error:", err)
				return false
			}
		}
	}
	return true
}

func (s *S3) Get(destination string, items ...string) error {
	client := s.getCredentials()
	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint:    &s.Endpoint,
		Region:      &s.Region,
		Credentials: client,
	}))

	for _, item := range items {
		err := func() error {
			file, err := os.Create(fmt.Sprintf("%s/%s", destination, item))
			if err != nil {
				return err
			}
			defer file.Close()

			downloader := s3manager.NewDownloader(sess)
			_, err = downloader.Download(file, &s3.GetObjectInput{
				Bucket: aws.String(s.Bucket),
				Key:    aws.String(item),
			})
			if err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			return err
		}
	}
	return nil
}
