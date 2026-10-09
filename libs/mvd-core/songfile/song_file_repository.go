package songfile

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// maxFileBytes bounds what Load reads; a list of songs is far smaller.
const maxFileBytes = 8 << 20

// Load reads the songs of a file (see Parse). The list is titled after the file.
func Load(path string) (List, error) {
	file, err := os.Open(path)
	if err != nil {
		return List{}, err
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(io.LimitReader(file, maxFileBytes))
	if err != nil {
		return List{}, err
	}
	list := Parse(string(data))
	list.Title = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	if len(list.Songs) == 0 {
		return List{}, fmt.Errorf("no songs found in %s: write one song per line as \"Artist - Title\", or use a CSV with title and artist columns", filepath.Base(path))
	}
	return list, nil
}
