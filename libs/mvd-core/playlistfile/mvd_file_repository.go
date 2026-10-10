package playlistfile

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Extension is the extension of a plan file, with its dot.
const Extension = ".mvd"

// version is the number in the header of a file this code writes and reads.
const version = 1

// File is a whole plan file.
type File struct {
	Created time.Time
	Entries []Entry
}

type header struct {
	MVD     int       `json:"mvd"`
	Created time.Time `json:"created"`
}

// Write writes f to w, compressed.
func Write(w io.Writer, f File) error {
	zw := gzip.NewWriter(w)
	enc := json.NewEncoder(zw)
	if err := enc.Encode(header{MVD: version, Created: f.Created}); err != nil {
		return err
	}
	for _, e := range f.Entries {
		if err := enc.Encode(e); err != nil {
			return err
		}
	}
	return zw.Close()
}

// Read reads a file written by Write.
func Read(r io.Reader) (File, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return File{}, errors.New("this is not an MVD file")
	}
	defer func() { _ = zr.Close() }()

	scanner := bufio.NewScanner(zr)
	scanner.Buffer(make([]byte, 0, 1024*1024), 16*1024*1024)

	var f File
	first := true
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		if first {
			first = false
			var h header
			if err := json.Unmarshal(line, &h); err != nil || h.MVD == 0 {
				return File{}, errors.New("this is not an MVD file")
			}
			if h.MVD > version {
				return File{}, fmt.Errorf("this file was saved by a newer MVD (format %d); update MVD to open it", h.MVD)
			}
			f.Created = h.Created
			continue
		}
		var e Entry
		if err := json.Unmarshal(line, &e); err != nil {
			return File{}, fmt.Errorf("cannot read a song of the file: %w", err)
		}
		f.Entries = append(f.Entries, e)
	}
	if err := scanner.Err(); err != nil {
		return File{}, errors.New("the file is damaged: " + err.Error())
	}
	if first {
		return File{}, errors.New("this is not an MVD file")
	}
	return f, nil
}

// Save writes f to path through a temporary file, so a file that already exists is
// never left half written.
func Save(path string, f File) error {
	if f.Created.IsZero() {
		f.Created = time.Now().UTC()
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".mvd-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	if err := Write(tmp, f); err != nil {
		_ = tmp.Close()
		_ = os.Remove(name)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(name)
		return err
	}
	if err := os.Rename(name, path); err != nil {
		_ = os.Remove(name)
		return err
	}
	return nil
}

// Load reads the file at path.
func Load(path string) (File, error) {
	file, err := os.Open(path)
	if err != nil {
		return File{}, err
	}
	defer func() { _ = file.Close() }()
	return Read(file)
}
