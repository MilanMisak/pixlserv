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
	LOCAL_PATH_ENV_VAR = "PIXLSERV_LOCAL_PATH"
	S3_BUCKET_ENV_VAR  = "PIXLSERV_S3_BUCKET"

	DEFAULT_LOCAL_PATH = "local-images"
)

var (
	storage Storage = nil
)

type Storage interface {
	init() error

	loadImage(imagePath string) (image.Image, string, error)

	saveImage(img image.Image, format string, imagePath string) error

	imageExists(imagePath string) bool
}

func storageInit() {
	// TODO - return error
	// 	storage = new(LocalStorage)
	storage = new(S3Storage)
	err := storage.init()
	if err != nil {
		log.Println("Storage could not be initialised:")
		log.Println(err)
	}
}

func storageCleanUp() {
}

func loadImage(imagePath string) (image.Image, string, error) {
	return storage.loadImage(imagePath)
}

func saveImage(img image.Image, format string, imagePath string) error {
	return storage.saveImage(img, format, imagePath)
}

func imageExists(imagePath string) bool {
	return storage.imageExists(imagePath)
}

///// Local storage
type LocalStorage struct {
	path string
}

func (s *LocalStorage) init() error {
	path := os.Getenv(LOCAL_PATH_ENV_VAR)
	if path == "" {
		path = DEFAULT_LOCAL_PATH
	}
	s.path = path
	return nil
}

func (s *LocalStorage) loadImage(imagePath string) (image.Image, string, error) {
	reader, err := os.Open(s.path + "/" + imagePath)
	defer reader.Close()

	if err != nil {
		return nil, "", fmt.Errorf("Image not found: %q", imagePath)
	}
	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, "", fmt.Errorf("Cannot decode image: %q", imagePath)
	}
	return img, format, nil
}

func (s *LocalStorage) saveImage(img image.Image, format string, imagePath string) error {
	// Open file for writing, overwrite if it already exists
	writer, err := os.Create(s.path + "/" + imagePath)
	defer writer.Close()

	if err != nil {
		return err
	}

	return writeImage(img, format, writer)
}

func (s *LocalStorage) imageExists(imagePath string) bool {
	if _, err := os.Stat(s.path + "/" + imagePath); os.IsNotExist(err) {
		return false
	}
	return true
}

///// Amazon S3 storage
type S3Storage struct {
	bucket *s3.Bucket
}

func (s *S3Storage) init() error {
	auth, err := aws.EnvAuth()
	if err != nil {
		return err
	}

	bucketName := os.Getenv(S3_BUCKET_ENV_VAR)
	if bucketName == "" {
		return fmt.Errorf("%s not set", S3_BUCKET_ENV_VAR)
	}

	conn := s3.New(auth, aws.EUWest)
	s.bucket = conn.Bucket(bucketName)

	return nil
}

func (s *S3Storage) loadImage(imagePath string) (image.Image, string, error) {
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

func (s *S3Storage) saveImage(img image.Image, format string, imagePath string) error {
	var buffer bytes.Buffer
	err := writeImage(img, format, &buffer)
	if err != nil {
		return err
	}

	return s.bucket.Put(imagePath, buffer.Bytes(), "image/"+format, s3.PublicRead)
}

func (s *S3Storage) imageExists(imagePath string) bool {
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
