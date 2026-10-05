package tray

import (
	"bytes"
	_ "embed"
	"encoding/binary"
)

// iconSize is the edge of the embedded icon in pixels. Operating systems scale it
// down to what the tray shows, and 128 stays sharp on high-DPI screens.
const iconSize = 128

// logoPNG is assets/logo.png scaled to iconSize. Replace it when the logo changes.
//
//go:embed tray_icon.png
var logoPNG []byte

// trayIcon returns the logo in the format the tray of goos expects: Windows wants an
// .ico, everything else takes a PNG.
func trayIcon(goos string) []byte {
	if goos == "windows" {
		return icoFromPNG(logoPNG, iconSize)
	}

	return logoPNG
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
