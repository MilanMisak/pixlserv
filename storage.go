package main

import (
	"fmt"
	"image"
	"os"

	//"launchpad.net/goamz/s3"
)

var (
	storage Storage = nil
)

type Storage interface {
	loadImage(imagePath string) (image.Image, string, error)

	saveImage(img image.Image, format string, imagePath string) error

	imageExists(imagePath string) bool
}

func storageInit() {
	storage = new(LocalStorage)
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
}

func (s *LocalStorage) loadImage(imagePath string) (image.Image, string, error) {
	reader, err := os.Open(LOCAL_IMAGES_PATH + "/" + imagePath)
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
	writer, err := os.Create(LOCAL_IMAGES_PATH + "/" + imagePath)
	defer writer.Close()

	if err != nil {
		return err
	}

	return writeImage(img, format, writer)
}

func (s *LocalStorage) imageExists(imagePath string) bool {
	if _, err := os.Stat(LOCAL_IMAGES_PATH + "/" + imagePath); os.IsNotExist(err) {
		return false
	}
	return true
}
