package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
)

// iconSize is the edge of the generated icon in pixels. Operating systems scale it
// down to what the tray shows, and 64 stays sharp on high-DPI screens.
const iconSize = 64

// supersample is how many samples per pixel edge are averaged to smooth the edges.
const supersample = 4

// trayIcon returns the icon in the format the tray of goos expects: Windows wants an
// .ico, everything else takes a PNG.
func trayIcon(goos string) []byte {
	data := trayIconPNG()
	if goos == "windows" {
		return icoFromPNG(data, iconSize)
	}

	return data
}

// trayIconPNG draws a white download arrow on a blue rounded square.
func trayIconPNG() []byte {
	img := image.NewNRGBA(image.Rect(0, 0, iconSize, iconSize))
	for y := range iconSize {
		for x := range iconSize {
			img.SetNRGBA(x, y, pixel(x, y))
		}
	}

	var out bytes.Buffer
	_ = png.Encode(&out, img)

	return out.Bytes()
}

// pixel averages supersample x supersample points inside the pixel (x, y).
func pixel(x, y int) color.NRGBA {
	var r, g, b, a float64
	total := float64(supersample * supersample)
	for sy := range supersample {
		for sx := range supersample {
			u := (float64(x) + (float64(sx)+0.5)/supersample) / iconSize
			v := (float64(y) + (float64(sy)+0.5)/supersample) / iconSize
			switch {
			case inArrow(u, v):
				r, g, b, a = r+255, g+255, b+255, a+255
			case inRoundedSquare(u, v):
				r, g, b, a = r+59, g+130, b+246, a+255
			}
		}
	}
	if a == 0 {
		return color.NRGBA{}
	}

	// r, g and b were summed only over covered samples, so dividing by the covered
	// alpha gives the colour and dividing alpha by all samples gives the coverage.
	covered := a / 255

	return color.NRGBA{R: uint8(r / covered), G: uint8(g / covered), B: uint8(b / covered), A: uint8(a / total)}
}

// inRoundedSquare reports whether (u, v), both in 0..1, is inside the background.
func inRoundedSquare(u, v float64) bool {
	const radius = 0.22
	const margin = 0.02
	cx := clamp(u, margin+radius, 1-margin-radius)
	cy := clamp(v, margin+radius, 1-margin-radius)
	dx, dy := u-cx, v-cy

	return u >= margin && u <= 1-margin && v >= margin && v <= 1-margin && dx*dx+dy*dy <= radius*radius
}

// inArrow reports whether (u, v) is on the arrow: a stem, a head and a base bar.
func inArrow(u, v float64) bool {
	stem := u >= 0.42 && u <= 0.58 && v >= 0.2 && v <= 0.52
	bar := u >= 0.25 && u <= 0.75 && v >= 0.8 && v <= 0.87

	return stem || bar || inTriangle(u, v, 0.25, 0.5, 0.75, 0.5, 0.5, 0.78)
}

func inTriangle(px, py, ax, ay, bx, by, cx, cy float64) bool {
	side := func(x1, y1, x2, y2 float64) float64 { return (px-x2)*(y1-y2) - (x1-x2)*(py-y2) }
	d1, d2, d3 := side(ax, ay, bx, by), side(bx, by, cx, cy), side(cx, cy, ax, ay)
	negative := d1 < 0 || d2 < 0 || d3 < 0
	positive := d1 > 0 || d2 > 0 || d3 > 0

	return !negative || !positive
}

func clamp(value, low, high float64) float64 {
	return max(low, min(high, value))
}

// icoFromPNG wraps a PNG in a single-image .ico, which Windows has accepted since
// Vista. size is the PNG's edge in pixels (under 256).
func icoFromPNG(data []byte, size int) []byte {
	const headerSize, entrySize = 6, 16
	var out bytes.Buffer
	_ = binary.Write(&out, binary.LittleEndian, struct {
		Reserved, Type, Count uint16
	}{0, 1, 1})
	_ = binary.Write(&out, binary.LittleEndian, struct {
		Width, Height, Colours, Reserved uint8
		Planes, BitsPerPixel             uint16
		Bytes, Offset                    uint32
	}{uint8(size), uint8(size), 0, 0, 1, 32, uint32(len(data)), headerSize + entrySize})
	out.Write(data)

	return out.Bytes()
}
