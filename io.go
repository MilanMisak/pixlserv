package main

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"io"
)

// Writes a given image of the given format to the given destination.
// Returns error.
func writeImage(img image.Image, format string, w io.Writer) error {
	if format == "png" {
		return png.Encode(w, img)
	}
	return jpeg.Encode(w, img, nil)
}

func readImage(data []byte, format string) (image.Image, error) {
	reader := bytes.NewReader(data)
	if format == "png" {
		return png.Decode(reader)
	}
	return jpeg.Decode(reader)
}
