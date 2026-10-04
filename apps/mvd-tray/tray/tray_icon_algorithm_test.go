package tray

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func decoded(t *testing.T, data []byte) image.Image {
	t.Helper()
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	return img
}

func at(img image.Image, u, v float64) color.NRGBA {
	b := img.Bounds()

	return color.NRGBAModel.Convert(img.At(int(u*float64(b.Dx())), int(v*float64(b.Dy())))).(color.NRGBA)
}

func TestTheIconIsASquarePNGWithTheArrowOnABlueRoundedSquare(t *testing.T) {
	img := decoded(t, trayIconPNG())

	if b := img.Bounds(); b.Dx() != iconSize || b.Dy() != iconSize {
		t.Fatalf("size = %v", b)
	}
	if got := at(img, 0.01, 0.01); got.A != 0 {
		t.Errorf("the corner should be transparent (rounded), got %+v", got)
	}
	if got := at(img, 0.1, 0.5); got != (color.NRGBA{R: 59, G: 130, B: 246, A: 255}) {
		t.Errorf("the background should be blue, got %+v", got)
	}
	for name, point := range map[string][2]float64{"stem": {0.5, 0.3}, "head": {0.5, 0.6}, "bar": {0.5, 0.83}} {
		if got := at(img, point[0], point[1]); got != (color.NRGBA{R: 255, G: 255, B: 255, A: 255}) {
			t.Errorf("the %s should be white, got %+v", name, got)
		}
	}
}

func TestEdgesAreSmoothedRatherThanJagged(t *testing.T) {
	img := decoded(t, trayIconPNG())
	partial := 0
	for y := range iconSize {
		for x := range iconSize {
			if a := at(img, float64(x)/iconSize, float64(y)/iconSize).A; a > 0 && a < 255 {
				partial++
			}
		}
	}
	if partial == 0 {
		t.Error("no partly transparent pixels: the corners are not anti-aliased")
	}
}

func TestWindowsGetsAnICOHoldingTheSamePNGAndEveryoneElseThePNG(t *testing.T) {
	pngData := trayIconPNG()

	if !bytes.Equal(trayIcon("linux"), pngData) || !bytes.Equal(trayIcon("darwin"), pngData) {
		t.Error("non-Windows platforms should get the PNG as is")
	}

	ico := trayIcon("windows")
	var header struct{ Reserved, Type, Count uint16 }
	if err := binary.Read(bytes.NewReader(ico), binary.LittleEndian, &header); err != nil {
		t.Fatal(err)
	}
	if header.Reserved != 0 || header.Type != 1 || header.Count != 1 {
		t.Errorf("header = %+v", header)
	}
	var entry struct {
		Width, Height, Colours, Reserved uint8
		Planes, BitsPerPixel             uint16
		Bytes, Offset                    uint32
	}
	if err := binary.Read(bytes.NewReader(ico[6:]), binary.LittleEndian, &entry); err != nil {
		t.Fatal(err)
	}
	if entry.Width != iconSize || entry.Height != iconSize || entry.Planes != 1 || entry.BitsPerPixel != 32 {
		t.Errorf("entry = %+v", entry)
	}
	if entry.Offset != 22 || int(entry.Offset+entry.Bytes) != len(ico) {
		t.Errorf("offset %d + bytes %d does not end the %d byte file", entry.Offset, entry.Bytes, len(ico))
	}
	if !bytes.Equal(ico[entry.Offset:], pngData) {
		t.Error("the payload is not the PNG")
	}
}
