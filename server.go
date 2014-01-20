package main

import (
	"bytes"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/codegangsta/martini"
	"github.com/nfnt/resize"
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
		if err != "" {
			return http.StatusBadRequest, err
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

			// TODO - magic

			// The values have been validated
			width, _ := strconv.Atoi(parameters[PARAMETER_WIDTH])
			height, _ := strconv.Atoi(parameters[PARAMETER_HEIGHT])

			var imgNew image.Image
			// TODO - keep these as ints
			imgWidth := float32(img.Bounds().Dx())
			imgHeight := float32(img.Bounds().Dy())

			// Resize and crop
			switch parameters[PARAMETER_CROPPING] {
			case CROPPING_MODE_EXACT:
				imgNew = resize.Resize(uint(width), uint(height), img, resize.Bilinear)
			case CROPPING_MODE_ALL:
				if float32(width)*(imgHeight/imgWidth) > float32(height) {
					// Keep height
					imgNew = resize.Resize(0, uint(height), img, resize.Bilinear)
				} else {
					// Keep width
					imgNew = resize.Resize(uint(width), 0, img, resize.Bilinear)
				}
			case CROPPING_MODE_PART:
				// Use the top left part of the image for now
				var croppedRect image.Rectangle
				if float32(width)*(imgHeight/imgWidth) > float32(height) {
					// Whole width displayed
					newHeight := int((imgWidth / float32(width)) * float32(height))
					croppedRect = image.Rect(0, 0, int(imgWidth), newHeight)
				} else {
					// Whole height displayed
					newWidth := int((imgHeight / float32(height)) * float32(width))
					croppedRect = image.Rect(0, 0, newWidth, int(imgHeight))
				}

				imgDraw := image.NewRGBA(croppedRect)
				// TODO - gravity
				draw.Draw(imgDraw, croppedRect, img, image.Point{0, 0}, draw.Src)
				imgNew = resize.Resize(uint(width), uint(height), imgDraw, resize.Bilinear)
			case CROPPING_MODE_KEEPSCALE:
				// TODO
				imgNew = resize.Resize(uint(width), uint(height), img, resize.Bilinear)
			}

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
