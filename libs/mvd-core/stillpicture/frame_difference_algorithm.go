package stillpicture

import "image"

const (
	sampleWidth  = 32
	sampleHeight = 24
)

// frameDifference is how much two frames differ, from 0 (the same picture) to 255:
// the mean absolute difference of their brightness, each sampled on a 32x24 grid so
// that the size of the frames and the noise of the compression do not count.
func frameDifference(a, b image.Image) float64 {
	var total float64
	for y := 0; y < sampleHeight; y++ {
		for x := 0; x < sampleWidth; x++ {
			d := brightnessAt(a, x, y) - brightnessAt(b, x, y)
			if d < 0 {
				d = -d
			}
			total += d
		}
	}
	return total / (sampleWidth * sampleHeight)
}

// brightnessAt is the brightness (0 to 255) of the frame at a point of the sample grid.
func brightnessAt(img image.Image, x, y int) float64 {
	bounds := img.Bounds()
	px := bounds.Min.X + (2*x+1)*bounds.Dx()/(2*sampleWidth)
	py := bounds.Min.Y + (2*y+1)*bounds.Dy()/(2*sampleHeight)
	r, g, b, _ := img.At(px, py).RGBA()
	// The 16-bit channels of RGBA, weighted the way the eye does.
	return (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 257
}
