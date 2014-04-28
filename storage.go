package main

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	"strings"

	// 	"launchpad.net/goamz/aws"
	// 	"launchpad.net/goamz/s3"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

const (
	localPathEnvVar = "PIXLSERV_LOCAL_PATH"
	s3BucketEnvVar  = "PIXLSERV_S3_BUCKET"

	defaultLocalPath = "local-images"
)

var (
	storageImpl storage
)

type storage interface {
	init() error

	loadImage(imagePath string) (image.Image, string, error)

	saveImage(img image.Image, format string, imagePath string) error

	imageExists(imagePath string) bool
}

func storageInit() {
	// TODO - return error
	// 	storage = new(localStorage)
	storageImpl = new(s3Storage)
	err := storageImpl.init()
	if err != nil {
		log.Println("Storage could not be initialised:")
		log.Println(err)
	}
}

func storageCleanUp() {
}

func loadImage(imagePath string) (image.Image, string, error) {
	return storageImpl.loadImage(imagePath)
}

func saveImage(img image.Image, format string, imagePath string) error {
	return storageImpl.saveImage(img, format, imagePath)
}

func imageExists(imagePath string) bool {
	return storageImpl.imageExists(imagePath)
}

// localStorage is a storage implementation using local disk
type localStorage struct {
	path string
}

func (s *localStorage) init() error {
	path := os.Getenv(localPathEnvVar)
	if path == "" {
		path = defaultLocalPath
	}
	s.path = path
	return nil
}

func (s *localStorage) loadImage(imagePath string) (image.Image, string, error) {
	reader, err := os.Open(s.path + "/" + imagePath)
	defer reader.Close()

	if err != nil {
		return nil, "", fmt.Errorf("image not found: %q", imagePath)
	}
	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, "", fmt.Errorf("cannot decode image: %q", imagePath)
	}
	return img, format, nil
}

func (s *localStorage) saveImage(img image.Image, format string, imagePath string) error {
	// Open file for writing, overwrite if it already exists
	writer, err := os.Create(s.path + "/" + imagePath)
	defer writer.Close()

	if err != nil {
		return err
	}

	return writeImage(img, format, writer)
}

func (s *localStorage) imageExists(imagePath string) bool {
	if _, err := os.Stat(s.path + "/" + imagePath); os.IsNotExist(err) {
		return false
	}
	return true
}

// s3Storage is a storage implementation using Amazon S3
type s3Storage struct {
	bucket *s3.Bucket
}

func (s *s3Storage) init() error {
	auth, err := aws.EnvAuth()
	if err != nil {
		return err
	}

	bucketName := os.Getenv(s3BucketEnvVar)
	if bucketName == "" {
		return fmt.Errorf("%s not set", s3BucketEnvVar)
	}

	conn := s3.New(auth, aws.EUWest)
	s.bucket = conn.Bucket(bucketName)

	return nil
}

func (s *s3Storage) loadImage(imagePath string) (image.Image, string, error) {
	data, err := s.bucket.Get(imagePath)
	if err != nil {
		return nil, "", err
	}

	format := strings.TrimLeft(filepath.Ext(imagePath), ".")
	image, err := readImage(data, format)
	if err != nil {
		return nil, "", err
	}

	return image, format, nil
}

func (s *s3Storage) saveImage(img image.Image, format string, imagePath string) error {
	var buffer bytes.Buffer
	err := writeImage(img, format, &buffer)
	if err != nil {
		return err
	}

	return s.bucket.Put(imagePath, buffer.Bytes(), "image/"+format, s3.PublicRead)
}

func (s *s3Storage) imageExists(imagePath string) bool {
	resp, err := s.bucket.List(imagePath, "/", "", 10)
	if err != nil {
		log.Printf("Error while listing S3 bucket: %s\n", err.Error())
		return false
	}
	if resp == nil {
		log.Println("Error while listing S3 bucket: empty response")
	}

	for _, element := range resp.Contents {
		if element.Key == imagePath {
			return true
		}
	}

	return false
}
