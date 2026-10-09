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

// Load reads the songs of a file. A .csv or .tsv with a header is read by its
// columns, anything else as "Artist - Title" lines. The list is titled after
// the file.
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
	text := string(data)

	list := List{Title: strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))}
	if ext := strings.ToLower(filepath.Ext(path)); ext == ".csv" || ext == ".tsv" {
		if songs, skipped, ok := ParseCSV(text); ok {
			list.Songs, list.Skipped = songs, skipped
		}
	}
	if list.Songs == nil {
		list.Songs, list.Skipped = ParseLines(text)
	}
	if len(list.Songs) == 0 {
		return List{}, fmt.Errorf("no songs found in %s: write one song per line as \"Artist - Title\", or use a CSV with title and artist columns", filepath.Base(path))
	}
	return list, nil
}
