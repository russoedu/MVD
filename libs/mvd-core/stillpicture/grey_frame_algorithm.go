package stillpicture

import "image"

// greyFrame is a frame as brightness values, row by row.
type greyFrame struct {
	width, height int
	value         []float64
}

// greyOf converts any image to brightness values, weighting the channels the way the
// eye does.
func greyOf(img image.Image) greyFrame {
	bounds := img.Bounds()
	frame := greyFrame{width: bounds.Dx(), height: bounds.Dy()}
	frame.value = make([]float64, frame.width*frame.height)
	for y := 0; y < frame.height; y++ {
		for x := 0; x < frame.width; x++ {
			r, g, b, _ := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			frame.value[y*frame.width+x] = (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 257
		}
	}
	return frame
}

// cutSheet cuts a storyboard sheet into its frames, left to right and top to bottom,
// up to count of them.
func cutSheet(sheet image.Image, width, height, rows, columns, count int) []greyFrame {
	bounds := sheet.Bounds()
	var frames []greyFrame
	for row := 0; row < rows; row++ {
		for column := 0; column < columns; column++ {
			if len(frames) == count {
				return frames
			}
			x, y := column*width, row*height
			if x+width > bounds.Dx() || y+height > bounds.Dy() {
				return frames
			}
			frames = append(frames, greyOf(subImage(sheet, bounds.Min.X+x, bounds.Min.Y+y, width, height)))
		}
	}
	return frames
}

// subImage is the part of an image at a position and size.
func subImage(img image.Image, x, y, width, height int) image.Image {
	if sub, ok := img.(interface {
		SubImage(image.Rectangle) image.Image
	}); ok {
		return sub.SubImage(image.Rect(x, y, x+width, y+height))
	}
	return img
}
