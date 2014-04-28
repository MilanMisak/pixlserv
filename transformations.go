package main

// TODO - refactor out a resizing function

import (
	"image"
	"image/draw"
	//"log"

	"github.com/nfnt/resize"
)

const (
	// CroppingModeExact crops an image exactly to given dimensions
	CroppingModeExact = "e"
	// CroppingModeAll crops an image so that all of it is displayed in a frame of at most given dimensions
	CroppingModeAll = "a"
	// CroppingModePart crops an image so that it fills a frame of given dimensions
	CroppingModePart = "p"
	// CroppingModeKeepScale crops an image so that it fills a frame of given dimensions, keeps scale
	CroppingModeKeepScale = "k"

	GravityNorth     = "n"
	GravityNorthEast = "ne"
	GravityEast      = "e"
	GravitySouthEast = "se"
	GravitySouth     = "s"
	GravitySouthWest = "sw"
	GravityWest      = "w"
	GravityNorthWest = "nw"
	GravityCenter    = "c"

	DefaultCroppingMode = CroppingModeExact
	DefaultGravity      = GravityNorthWest
)

func isValidCroppingMode(str string) bool {
	return str == CroppingModeExact || str == CroppingModeAll || str == CroppingModePart || str == CroppingModeKeepScale
}

func isValidGravity(str string) bool {
	return str == GravityNorth || str == GravityNorthEast || str == GravityEast || str == GravitySouthEast || str == GravitySouth || str == GravitySouthWest || str == GravityWest || str == GravityNorthWest || str == GravityCenter
}

func transformCropAndResize(img image.Image, parameters Params) (imgNew image.Image) {
	width := parameters.width
	height := parameters.height
	gravity := parameters.gravity

	imgWidth := img.Bounds().Dx()
	imgHeight := img.Bounds().Dy()

	// Resize and crop
	switch parameters.cropping {
	case CroppingModeExact:
		imgNew = resize.Resize(uint(width), uint(height), img, resize.Bilinear)
	case CroppingModeAll:
		if float32(width)*(float32(imgHeight)/float32(imgWidth)) > float32(height) {
			// Keep height
			imgNew = resize.Resize(0, uint(height), img, resize.Bilinear)
		} else {
			// Keep width
			imgNew = resize.Resize(uint(width), 0, img, resize.Bilinear)
		}
	case CroppingModePart:
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
	case CroppingModeKeepScale:
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
	case GravityNorth:
		return image.Point{(imgWidth - width) / 2, 0}
	case GravityNorthEast:
		return image.Point{imgWidth - width, 0}
	case GravityEast:
		return image.Point{imgWidth - width, (imgHeight - height) / 2}
	case GravitySouthEast:
		return image.Point{imgWidth - width, imgHeight - height}
	case GravitySouth:
		return image.Point{(imgWidth - width) / 2, imgHeight - height}
	case GravitySouthWest:
		return image.Point{0, imgHeight - height}
	case GravityWest:
		return image.Point{0, (imgHeight - height) / 2}
	case GravityNorthWest:
		return image.Point{0, 0}
	case GravityCenter:
		return image.Point{(imgWidth - width) / 2, (imgHeight - height) / 2}
	}
	panic("This point should not be reached")
}
