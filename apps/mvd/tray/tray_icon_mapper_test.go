package tray

import (
	"bytes"
	"encoding/binary"
	"image/png"
	"testing"
)

func TestTheEmbeddedLogoIsASquarePNGOfTheIconSize(t *testing.T) {
	img, err := png.Decode(bytes.NewReader(logoPNG))
	if err != nil {
		t.Fatal(err)
	}
	if b := img.Bounds(); b.Dx() != iconSize || b.Dy() != iconSize {
		t.Fatalf("size = %v, want %dx%d", b, iconSize, iconSize)
	}
}

func TestWindowsGetsAnICOHoldingTheSamePNGAndEveryoneElseThePNG(t *testing.T) {
	if !bytes.Equal(trayIcon("linux"), logoPNG) || !bytes.Equal(trayIcon("darwin"), logoPNG) {
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
	if !bytes.Equal(ico[entry.Offset:], logoPNG) {
		t.Error("the payload is not the PNG")
	}
}
