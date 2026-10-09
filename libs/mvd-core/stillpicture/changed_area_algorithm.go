package stillpicture

const (
	// registrationTolerance is how many pixels a frame may have slipped and still be the
	// same picture, as scanmate's tolerance forgives a misaligned scan: without it every
	// edge would be a change after compression noise or a one-pixel drift.
	registrationTolerance = 1
	// changeThreshold is how much a pixel's structure must differ, as a ratio of the
	// local background, to count as changed.
	changeThreshold = 0.18
)

// changedShare is the share of a frame, from 0 to 1, that differs from another one:
// the pixels whose structure matches none of the other frame's pixels within the
// tolerance. Both frames are structures (see structureOf) of the same size.
func changedShare(a, b greyFrame) float64 {
	if a.width != b.width || a.height != b.height || len(a.value) == 0 {
		return 0
	}
	changed := 0
	for y := 0; y < a.height; y++ {
		for x := 0; x < a.width; x++ {
			here := a.value[y*a.width+x]
			best := -1.0
			for dy := -registrationTolerance; dy <= registrationTolerance; dy++ {
				for dx := -registrationTolerance; dx <= registrationTolerance; dx++ {
					other := b.value[clamp(y+dy, b.height)*b.width+clamp(x+dx, b.width)]
					d := here - other
					if d < 0 {
						d = -d
					}
					if best < 0 || d < best {
						best = d
					}
				}
			}
			if best > changeThreshold {
				changed++
			}
		}
	}
	return float64(changed) / float64(len(a.value))
}
