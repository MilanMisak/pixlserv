package main

// TODO - refactor out a resizing function

import (
	"image"
	"image/draw"
	//"log"

	"github.com/nfnt/resize"
)

const (
	CROPPING_MODE_EXACT     = "e"
	CROPPING_MODE_ALL       = "a"
	CROPPING_MODE_PART      = "p"
	CROPPING_MODE_KEEPSCALE = "k"

	GRAVITY_NORTH      = "n"
	GRAVITY_NORTH_EAST = "ne"
	GRAVITY_EAST       = "e"
	GRAVITY_SOUTH_EAST = "se"
	GRAVITY_SOUTH      = "s"
	GRAVITY_SOUTH_WEST = "sw"
	GRAVITY_WEST       = "w"
	GRAVITY_NORTH_WEST = "nw"
	GRAVITY_CENTER     = "c"

	DEFAULT_CROPPING_MODE = CROPPING_MODE_EXACT
	DEFAULT_GRAVITY       = GRAVITY_NORTH_WEST
)

func isValidCroppingMode(str string) bool {
	return str == CROPPING_MODE_EXACT || str == CROPPING_MODE_ALL || str == CROPPING_MODE_PART || str == CROPPING_MODE_KEEPSCALE
}

func isValidGravity(str string) bool {
	return str == GRAVITY_NORTH || str == GRAVITY_NORTH_EAST || str == GRAVITY_EAST || str == GRAVITY_SOUTH_EAST || str == GRAVITY_SOUTH || str == GRAVITY_SOUTH_WEST || str == GRAVITY_WEST || str == GRAVITY_NORTH_WEST || str == GRAVITY_CENTER
}

func transformCropAndResize(img image.Image, parameters Params) (imgNew image.Image) {
	width := parameters.width
	height := parameters.height
	gravity := parameters.gravity

	imgWidth := img.Bounds().Dx()
	imgHeight := img.Bounds().Dy()

	// Resize and crop
	switch parameters.cropping {
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

		topLeftPoint := calculateTopLeftPointFromGravity(gravity, croppedRect.Dx(), croppedRect.Dy(), imgWidth, imgHeight)
		imgDraw := image.NewRGBA(croppedRect)

		draw.Draw(imgDraw, croppedRect, img, topLeftPoint, draw.Src)
		imgNew = resize.Resize(uint(width), uint(height), imgDraw, resize.Bilinear)
	case CROPPING_MODE_KEEPSCALE:
		// If passed in dimensions are bigger use those of the image
		if width > imgWidth {
			width = imgWidth
		}
		if height > imgHeight {
			height = imgHeight
		}

		croppedRect := image.Rect(0, 0, width, height)
		topLeftPoint := calculateTopLeftPointFromGravity(gravity, width, height, imgWidth, imgHeight)
		imgDraw := image.NewRGBA(croppedRect)

		draw.Draw(imgDraw, croppedRect, img, topLeftPoint, draw.Src)
		imgNew = imgDraw.SubImage(croppedRect)
	}

	return
}

func calculateTopLeftPointFromGravity(gravity string, width, height int, imgWidth, imgHeight int) image.Point {
	// Assuming width <= imgWidth && height <= imgHeight
	switch gravity {
	case GRAVITY_NORTH:
		return image.Point{(imgWidth - width) / 2, 0}
	case GRAVITY_NORTH_EAST:
		return image.Point{imgWidth - width, 0}
	case GRAVITY_EAST:
		return image.Point{imgWidth - width, (imgHeight - height) / 2}
	case GRAVITY_SOUTH_EAST:
		return image.Point{imgWidth - width, imgHeight - height}
	case GRAVITY_SOUTH:
		return image.Point{(imgWidth - width) / 2, imgHeight - height}
	case GRAVITY_SOUTH_WEST:
		return image.Point{0, imgHeight - height}
	case GRAVITY_WEST:
		return image.Point{0, (imgHeight - height) / 2}
	case GRAVITY_NORTH_WEST:
		return image.Point{0, 0}
	case GRAVITY_CENTER:
		return image.Point{(imgWidth - width) / 2, (imgHeight - height) / 2}
	}
	panic("This point should not be reached")
}
