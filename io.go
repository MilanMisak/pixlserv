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
func readImage(imagePath string) (image.Image, string, error) {
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

func writeImage(img image.Image, format string, w io.Writer) error {
	if format == "jpeg" {
		return jpeg.Encode(w, img, nil)
	} else {
		return png.Encode(w, img)
	}
}
