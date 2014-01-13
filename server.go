package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"

	"github.com/codegangsta/martini"
)

const (
	LOCAL_IMAGES_PATH = "local-images"
)

func main() {
	m := martini.Classic()
	m.Get("/image/**", func(params martini.Params) (int, string) {
		imagePath := params["_1"]

		if _, err := os.Stat(LOCAL_IMAGES_PATH + "/" + imagePath); os.IsNotExist(err) {
			return http.StatusNotFound, "Image not found: " + imagePath
		} else {
			reader, err := os.Open(LOCAL_IMAGES_PATH + "/" + imagePath)
			if err != nil {
				return http.StatusInternalServerError, "Cannot open image"
			}
			image, format, err := image.Decode(reader)
			if err != nil {
				return http.StatusInternalServerError, "Cannot decode image"
			}

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

func log(message string) {
	fmt.Println("[pixlserv] " + message)
}
