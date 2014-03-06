package main

import (
	"image"
	"image/jpeg"
	"image/png"
	"io"
)

// Writes a given image of the given format to the given destination.
// Returns error.
func writeImage(img image.Image, format string, w io.Writer) error {
	if format == "jpeg" {
		return jpeg.Encode(w, img, nil)
	} else {
		return png.Encode(w, img)
	}
}
