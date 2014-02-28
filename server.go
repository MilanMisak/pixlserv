package main

import (
	"bytes"
	"log"
	"net/http"

	"github.com/codegangsta/martini"
)

const (
	LOCAL_IMAGES_PATH = "local-images"
)

func main() {
	// Set up logging
	log.SetPrefix("[pixlserv] ")
	log.SetFlags(0) // Remove the timestamp

	// Initialise the cache
	err := cacheInit()

	if err != nil {
		log.Println("Cache initialisation failed:", err)
		return
	}

	// Run the server
	m := martini.Classic()
	m.Get("/image/:parameters/**", func(params martini.Params) (int, string) {
		parameters, err := parseParameters(params["parameters"])
		if err != nil {
			return http.StatusBadRequest, err.Error()
		}
		log.Println("Parameters:")
		log.Println(parameters)
		baseImagePath := params["_1"]

		// Check if the image with the given parameters already exists
		// and return it
		fullImagePath, _ := createFilePath(baseImagePath, parameters)
		if fileExistsInCache(fullImagePath) {
			img, format, err := loadImage(fullImagePath)
			if err == nil {
				var buffer bytes.Buffer
				writeImage(img, format, &buffer)

				return http.StatusOK, buffer.String()
			}
		}

		// Load the original image and process it
		if !imageExists(baseImagePath) {
			return http.StatusNotFound, "Image not found: " + baseImagePath
		} else {
			img, format, err := loadImage(baseImagePath)
			if err != nil {
				return http.StatusInternalServerError, err.Error()
			}

			imgNew := transformCropAndResize(img, parameters)

			var buffer bytes.Buffer
			err = writeImage(imgNew, format, &buffer)
			if err != nil {
				log.Println("Writing an image to the response failed:", err)
			}

			// TODO - these should be done asynchronously to speed up the request
            err = saveImage(imgNew, format, fullImagePath)
			if err != nil {
				log.Println("Saving an image to cache failed:", err)
			}
			addToCache(fullImagePath)  // &buffer

			return http.StatusOK, buffer.String()
		}
	})
	m.Run()
}
