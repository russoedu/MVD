package stillpicture

// structureWindow is the side, in pixels, of the window whose mean is the local
// background of a pixel.
const structureWindow = 9

// structureOf divides every pixel of a frame by the mean of its surroundings, so what
// is left is the frame's structure whatever the light: a fade to black, a flash or a
// vignette change the background of every pixel together and cancel out. This is the
// "ink, not brightness" step of scanmate's pixel comparison.
func structureOf(frame greyFrame) greyFrame {
	out := greyFrame{width: frame.width, height: frame.height, value: make([]float64, len(frame.value))}
	half := structureWindow / 2
	for y := 0; y < frame.height; y++ {
		for x := 0; x < frame.width; x++ {
			var sum float64
			for dy := -half; dy <= half; dy++ {
				for dx := -half; dx <= half; dx++ {
					sum += frame.value[clamp(y+dy, frame.height)*frame.width+clamp(x+dx, frame.width)] + 1
				}
			}
			mean := sum / (structureWindow * structureWindow)
			out.value[y*frame.width+x] = (frame.value[y*frame.width+x] + 1) / mean
		}
	}
	return out
}

// clamp keeps an index inside 0..size-1, repeating the edge.
func clamp(i, size int) int {
	if i < 0 {
		return 0
	}
	if i >= size {
		return size - 1
	}
	return i
}
