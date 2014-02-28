package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
)

// Reads an image at the given path, returns an image instance,
// format string and an error
func loadImage(imagePath string) (image.Image, string, error) {
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

func saveImage(img image.Image, format string, imagePath string) error {
	// Open file for writing, overwrite if it already exists
	writer, err := os.Create(LOCAL_IMAGES_PATH + "/" + imagePath)
	defer writer.Close()

	if err != nil {
		return err
	}

	return writeImage(img, format, writer)
}

// Writes a given image of the given format to the given destination.
// Returns error.
func writeImage(img image.Image, format string, w io.Writer) error {
	if format == "jpeg" {
		return jpeg.Encode(w, img, nil)
	} else {
		return png.Encode(w, img)
	}
}

// Checks if an image file exists at the given path.
func imageExists(imagePath string) bool {
	if _, err := os.Stat(LOCAL_IMAGES_PATH + "/" + imagePath); os.IsNotExist(err) {
		return false
	}
	return true
}
