package main

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"

	"github.com/codegangsta/martini"
)

const (
	LOCAL_IMAGES_PATH = "local-images"
)

func main() {
	// Set up logging
	log.SetPrefix("[pixlserv] ")
	log.SetFlags(0) // Removed the timestamp

	// Run the server
	m := martini.Classic()
	m.Get("/image/:parameters/**", func(params martini.Params) (int, string) {
		parameters, err := parseParameters(params["parameters"])
		if err != nil {
			return http.StatusBadRequest, err.Error()
		}
		log.Println("Parameters:")
		log.Println(parameters)
		imagePath := params["_1"]

		if _, err := os.Stat(LOCAL_IMAGES_PATH + "/" + imagePath); os.IsNotExist(err) {
			return http.StatusNotFound, "Image not found: " + imagePath
		} else {
			img, format, e := readImage(imagePath)
			if e != "" {
				return http.StatusInternalServerError, e
			}

			// TODO - more transformations
			imgNew := transformCropAndResize(img, parameters)

			var buffer bytes.Buffer
			if format == "jpeg" {
				jpeg.Encode(&buffer, imgNew, nil)
			} else {
				png.Encode(&buffer, imgNew)
			}

			return http.StatusOK, buffer.String()
		}
	})
	m.Run()
}

// Reads an image at the given path, returns an image instance,
// format string and an error
func readImage(imagePath string) (image.Image, string, string) {
	reader, err := os.Open(LOCAL_IMAGES_PATH + "/" + imagePath)
	defer reader.Close()

	if err != nil {
		return nil, "", "Cannot open image"
	}
	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, "", "Cannot decode image"
	}
	return img, format, ""
}
