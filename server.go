package main

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"
	"strings"
	"strconv"

	"github.com/codegangsta/martini"
)

const (
	LOCAL_IMAGES_PATH = "local-images"
)

func main() {
	// Set up logging
	log.SetPrefix("[pixlserv] ")
	log.SetFlags(0)  // Removed the timestamp
	log.Println("Test")

	// Run the server
	m := martini.Classic()
	m.Get("/image/:parameters/**", func(params martini.Params) (int, string) {
		parameters, err := parseParameters(params["parameters"])
		if err != "" {
			return http.StatusBadRequest, err
		}
		log.Println("Parameters:")
		log.Println(parameters)
		imagePath := params["_1"]

		if _, err := os.Stat(LOCAL_IMAGES_PATH + "/" + imagePath); os.IsNotExist(err) {
			return http.StatusNotFound, "Image not found: " + imagePath
		} else {
			image, format, e := readImage(imagePath)
			if e != "" {
				return http.StatusInternalServerError, e
			}

			// TODO - magic

			var buffer bytes.Buffer
			if format == "jpeg" {
				jpeg.Encode(&buffer, image, nil)
			} else {
				png.Encode(&buffer, image)
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
	if err != nil {
		return nil, "", "Cannot open image"
	}
	image, format, err := image.Decode(reader)
	if err != nil {
		return nil, "", "Cannot decode image"
	}
	return image, format, ""
}

// Turns a string like "w_400,h_300" into a map[w:400 h:300]
// The second return value is an error message
// Also validates the parameters to make sure they have valid values
// w = width, h = height
func parseParameters(parametersStr string) (map[string]string, string) {
	parameters := make(map[string]string)
	parts := strings.Split(parametersStr, ",")
	for _, part := range parts {
		// TODO - validation
		keyAndValue := strings.SplitN(part, "_", 2)
		key := keyAndValue[0]
		value := keyAndValue[1]

		switch key {
		case "w", "h":
			value, err := strconv.Atoi(value)
			if err != nil {
				return nil, "Could not parse value for parameter: " + key
			}
			if value <= 0 {
				return nil, "Value must be > 0: " + key
			}
		}

		parameters[key] = value
	}
	return parameters, ""
}
