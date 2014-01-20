package main

import (
	"image"
	"image/draw"
	"strconv"

	"github.com/nfnt/resize"
)

const (
	CROPPING_MODE_EXACT     = "e"
	CROPPING_MODE_ALL       = "a"
	CROPPING_MODE_PART      = "p"
	CROPPING_MODE_KEEPSCALE = "k"

	DEFAULT_CROPPING_MODE = CROPPING_MODE_EXACT
)

func isValidCroppingMode(str string) bool {
	return str == CROPPING_MODE_EXACT || str == CROPPING_MODE_ALL || str == CROPPING_MODE_PART || str == CROPPING_MODE_KEEPSCALE
}

func transformCropAndResize(img image.Image, parameters map[string]string) (imgNew image.Image) {
	width, _ := strconv.Atoi(parameters[PARAMETER_WIDTH])
	height, _ := strconv.Atoi(parameters[PARAMETER_HEIGHT])

	imgWidth := img.Bounds().Dx()
	imgHeight := img.Bounds().Dy()

	// Resize and crop
	switch parameters[PARAMETER_CROPPING] {
	case CROPPING_MODE_EXACT:
		imgNew = resize.Resize(uint(width), uint(height), img, resize.Bilinear)
	case CROPPING_MODE_ALL:
		if float32(width)*(float32(imgHeight)/float32(imgWidth)) > float32(height) {
			// Keep height
			imgNew = resize.Resize(0, uint(height), img, resize.Bilinear)
		} else {
			// Keep width
			imgNew = resize.Resize(uint(width), 0, img, resize.Bilinear)
		}
	case CROPPING_MODE_PART:
		// Use the top left part of the image for now
		var croppedRect image.Rectangle
		if float32(width)*(float32(imgHeight)/float32(imgWidth)) > float32(height) {
			// Whole width displayed
			newHeight := int((float32(imgWidth) / float32(width)) * float32(height))
			croppedRect = image.Rect(0, 0, imgWidth, newHeight)
		} else {
			// Whole height displayed
			newWidth := int((float32(imgHeight) / float32(height)) * float32(width))
			croppedRect = image.Rect(0, 0, newWidth, imgHeight)
		}

		imgDraw := image.NewRGBA(croppedRect)
		// TODO - gravity
		draw.Draw(imgDraw, croppedRect, img, image.Point{0, 0}, draw.Src)
		imgNew = resize.Resize(uint(width), uint(height), imgDraw, resize.Bilinear)
	case CROPPING_MODE_KEEPSCALE:
		// TODO - implement
		imgNew = resize.Resize(uint(width), uint(height), img, resize.Bilinear)
	}

	return
}
