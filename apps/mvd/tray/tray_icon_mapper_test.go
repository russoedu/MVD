package tray

import (
	"bytes"
	"encoding/binary"
	"image/png"
	"testing"
)

func TestWindowsGetsAnICOWithTheSizesTheTrayNeeds(t *testing.T) {
	ico := trayIcon("windows")

	var header struct{ Reserved, Type, Count uint16 }
	if err := binary.Read(bytes.NewReader(ico), binary.LittleEndian, &header); err != nil {
		t.Fatal(err)
	}
	if header.Reserved != 0 || header.Type != 1 || header.Count < 2 {
		t.Fatalf("header = %+v, want an icon with several images", header)
	}

	sizes := map[int]bool{}
	for i := 0; i < int(header.Count); i++ {
		var entry struct {
			Width, Height, Colours, Reserved uint8
			Planes, BitsPerPixel             uint16
			Bytes, Offset                    uint32
		}
		if err := binary.Read(bytes.NewReader(ico[6+16*i:]), binary.LittleEndian, &entry); err != nil {
			t.Fatal(err)
		}
		if int(entry.Offset+entry.Bytes) > len(ico) {
			t.Errorf("image %d runs past the end of the file", i)
		}
		sizes[int(entry.Width)] = true
	}
	for _, want := range []int{16, 32} {
		if !sizes[want] {
			t.Errorf("no %d px image; has %v", want, sizes)
		}
	}
}

func TestMacAndLinuxGetAPNGOfTheSizeTheirTrayShows(t *testing.T) {
	for goos, wantHeight := range map[string]int{"darwin": 44, "linux": 44} {
		img, err := png.Decode(bytes.NewReader(trayIcon(goos)))
		if err != nil {
			t.Fatalf("%s: %v", goos, err)
		}
		if got := img.Bounds().Dy(); got != wantHeight {
			t.Errorf("%s: %d px tall, want %d", goos, got, wantHeight)
		}
	}
}

func TestAnyOtherSystemGetsTheLinuxIcon(t *testing.T) {
	if !bytes.Equal(trayIcon("freebsd"), trayIcon("linux")) {
		t.Error("other systems should get the PNG a Linux tray gets")
	}
}
