package main

import (
	"bytes"
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
			img, format, err := readImage(imagePath)
			if err != nil {
				return http.StatusInternalServerError, err.Error()
			}

			// TODO - more transformations
			imgNew := transformCropAndResize(img, parameters)

			var buffer bytes.Buffer
			writeImage(imgNew, format, &buffer)

			// TODO - cache

			return http.StatusOK, buffer.String()
		}
	})
	m.Run()
}
