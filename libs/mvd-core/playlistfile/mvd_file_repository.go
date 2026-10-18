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
const version = 2

// File is a download list: what is to be downloaded, in order, and how far each song has
// got, so the list can be put down and taken up again another day.
type File struct {
	Created time.Time
	Items   []Item
}

// Item is one line of the list: an address (a playlist, a video, a link of another
// service or a file of songs) and, once it has been looked at, its songs.
type Item struct {
	// Source is the line as the engine takes it.
	Source string  `json:"source"`
	Songs  []Entry `json:"songs,omitempty"`
}

// Songs returns every song of the list, in order.
func (f File) Songs() []Entry {
	var songs []Entry
	for _, item := range f.Items {
		songs = append(songs, item.Songs...)
	}
	return songs
}

// Sources returns the line of each item, in order.
func (f File) Sources() []string {
	sources := make([]string, len(f.Items))
	for i, item := range f.Items {
		sources[i] = item.Source
	}
	return sources
}

// FromSongs builds a list from songs, one item per playlist, in the order the
// playlists first appear.
func FromSongs(songs []Entry) File {
	var f File
	at := map[string]int{}
	for _, s := range songs {
		i, ok := at[s.PlaylistURL]
		if !ok {
			i = len(f.Items)
			at[s.PlaylistURL] = i
			f.Items = append(f.Items, Item{Source: s.PlaylistURL})
		}
		f.Items[i].Songs = append(f.Items[i].Songs, s)
	}
	return f
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
	for _, item := range f.Items {
		if err := enc.Encode(item); err != nil {
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
			if h.MVD != version {
				return File{}, fmt.Errorf("this file was saved by another version of MVD (format %d, this one reads %d)", h.MVD, version)
			}
			f.Created = h.Created
			continue
		}
		var item Item
		if err := json.Unmarshal(line, &item); err != nil {
			return File{}, fmt.Errorf("cannot read a line of the list: %w", err)
		}
		f.Items = append(f.Items, item)
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
