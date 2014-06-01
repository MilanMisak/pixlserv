package main

// TODO - refactor out a resizing function

import (
	"image"
	"image/color"
	"image/draw"
	"log"

	"code.google.com/p/freetype-go/freetype"
	"code.google.com/p/freetype-go/freetype/truetype"
	"github.com/nfnt/resize"
)

// Transformation specifies parameters and a watermark to be used when transforming an image
type Transformation struct {
	params    *Params
	watermark *Watermark
	texts     []*Text
}

// Watermark specifies a watermark to be applied to an image
type Watermark struct {
	imagePath string
	x, y      int
}

// Text specifies a text overlay to be applied to an image
type Text struct {
	content    string
	x, y, size int
	font       *truetype.Font
	color      color.Color
}

// FontMetrics defines font metrics for a Text struct as rounded up integers
type FontMetrics struct {
	width, height, ascent, descent float64
}

func (t *Text) GetFontMetrics(scale int) FontMetrics {
	// Adapted from: https://code.google.com/p/plotinum/

	// Converts truetype.FUnit to float64
	fUnit2Float64 := float64(t.size) / float64(t.font.FUnitsPerEm())

	width := 0
	prev, hasPrev := truetype.Index(0), false
	for _, rune := range t.content {
		index := t.font.Index(rune)
		if hasPrev {
			width += int(t.font.Kerning(t.font.FUnitsPerEm(), prev, index))
		}
		width += int(t.font.HMetric(t.font.FUnitsPerEm(), index).AdvanceWidth)
		prev, hasPrev = index, true
	}
	widthFloat := float64(width) * fUnit2Float64 * float64(scale)

	bounds := t.font.Bounds(t.font.FUnitsPerEm())
	height := float64(bounds.YMax-bounds.YMin) * fUnit2Float64 * float64(scale)
	ascent := float64(bounds.YMax) * fUnit2Float64 * float64(scale)
	descent := float64(bounds.YMin) * fUnit2Float64 * float64(scale)

	return FontMetrics{widthFloat, height, ascent, descent}
}

func transformCropAndResize(img image.Image, transformation *Transformation) (imgNew image.Image) {
	parameters := transformation.params
	width := parameters.width
	height := parameters.height
	gravity := parameters.gravity
	scale := parameters.scale

	imgWidth := img.Bounds().Dx()
	imgHeight := img.Bounds().Dy()

	// Scaling factor
	if parameters.cropping != CroppingModeKeepScale {
		width *= scale
		height *= scale
	}

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

	// Filters
	if parameters.filter == FilterGrayScale {
		bounds := imgNew.Bounds()
		w, h := bounds.Max.X, bounds.Max.Y
		gray := image.NewGray(bounds)
		for x := 0; x < w; x++ {
			for y := 0; y < h; y++ {
				oldColor := imgNew.At(x, y)
				grayColor := color.GrayModel.Convert(oldColor)
				gray.Set(x, y, grayColor)
			}
		}
		imgNew = gray
	}

	if transformation.watermark != nil {
		w := transformation.watermark

		var watermarkSrcScaled image.Image
		var watermarkBounds image.Rectangle

		// Try to load a scaled watermark first
		if scale > 1 {
			scaledPath, err := constructScaledPath(w.imagePath, scale)
			if err != nil {
				log.Println("Error:", err)
				return
			}

			watermarkSrc, _, err := loadImage(scaledPath)
			if err != nil {
				log.Println("Error: could not load a watermark", err)
			} else {
				watermarkBounds = watermarkSrc.Bounds()
				watermarkSrcScaled = watermarkSrc
			}
		}

		if watermarkSrcScaled == nil {
			watermarkSrc, _, err := loadImage(w.imagePath)
			if err != nil {
				log.Println("Error: could not load a watermark", err)
				return
			}
			watermarkBounds = image.Rect(0, 0, watermarkSrc.Bounds().Max.X*scale, watermarkSrc.Bounds().Max.Y*scale)
			watermarkSrcScaled = resize.Resize(uint(watermarkBounds.Max.X), uint(watermarkBounds.Max.Y), watermarkSrc, resize.Bilinear)
		}

		bounds := imgNew.Bounds()

		// Make sure we have a transparent watermark if possible
		watermark := image.NewRGBA(watermarkBounds)
		draw.Draw(watermark, watermarkBounds, watermarkSrcScaled, watermarkBounds.Min, draw.Src)

		wX := w.x * scale
		wY := w.y * scale
		if wX < 0 {
			wX += bounds.Dx() - watermarkBounds.Dx()
		}
		if wY < 0 {
			wY += bounds.Dy() - watermarkBounds.Dy()
		}
		watermarkRect := image.Rect(wX, wY, watermarkBounds.Dx()+wX, watermarkBounds.Dy()+wY)
		finalImage := image.NewRGBA(bounds)
		draw.Draw(finalImage, bounds, imgNew, bounds.Min, draw.Src)
		draw.Draw(finalImage, watermarkRect, watermark, watermarkBounds.Min, draw.Over)
		imgNew = finalImage.SubImage(bounds)
	}

	if transformation.texts != nil {
		bounds := imgNew.Bounds()
		rgba := image.NewRGBA(bounds)
		draw.Draw(rgba, bounds, imgNew, image.ZP, draw.Src)

		dpi := float64(72) // Multiply this by scale for a baaad time

		c := freetype.NewContext()
		c.SetDPI(dpi)
		c.SetClip(rgba.Bounds())
		c.SetDst(rgba)

		for _, text := range transformation.texts {
			size := float64(text.size * scale)

			c.SetSrc(image.NewUniform(text.color))
			c.SetFont(text.font)
			c.SetFontSize(size)

			fontMetrics := text.GetFontMetrics(scale)
			x := text.x * scale
			y := text.y*scale + int(c.PointToFix32(fontMetrics.ascent)>>8)
			if x < 0 {
				x += bounds.Dx() - int(c.PointToFix32(fontMetrics.width)>>8)
			}
			if y < 0 {
				y += bounds.Dy() - int(c.PointToFix32(fontMetrics.height)>>8)
			}

			_, err := c.DrawString(text.content, freetype.Pt(x, y))
			if err != nil {
				log.Println("Error adding text:", err)
				return
			}
		}

		imgNew = rgba
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
