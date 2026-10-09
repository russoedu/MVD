package songfile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	unsafeName = regexp.MustCompile(`[<>:"/\|?*\x00-\x1f]+`)
	spaces     = regexp.MustCompile(`\s+`)
)

// maxNameRunes bounds a list's file name.
const maxNameRunes = 60

// Save stores a pasted list of songs in dir and returns the path of the new
// file, which a download list can then name. The list is called what its first
// line says when that line is a "# comment", else after the time. A CSV is
// stored as .csv, anything else as .txt. A text without songs is an error and
// stores nothing.
func Save(dir, text string, now time.Time) (string, error) {
	text = strings.ReplaceAll(strings.TrimPrefix(text, "\ufeff"), "\r\n", "\n")
	if len(Parse(text).Songs) == 0 {
		return "", errors.New("no songs to save: write one song per line as \"Artist - Title\"")
	}

	ext := ".txt"
	if IsCSV(text) {
		ext = ".csv"
	}
	name := listName(text)
	if name == "" {
		name = "Songs " + now.Format("2006-01-02 1504")
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	for n := 1; ; n++ {
		file := name + ext
		if n > 1 {
			file = fmt.Sprintf("%s (%d)%s", name, n, ext)
		}
		path := filepath.Join(dir, file)
		out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if os.IsExist(err) {
			continue
		}
		if err != nil {
			return "", err
		}
		_, err = out.WriteString(strings.TrimSpace(text) + "\n")
		if closeErr := out.Close(); err == nil {
			err = closeErr
		}
		if err != nil {
			_ = os.Remove(path)
			return "", err
		}
		return path, nil
	}
}

// listName is the name a "# name" first line gives, made safe for a file name.
func listName(text string) string {
	first, _, _ := strings.Cut(strings.TrimSpace(text), "\n")
	name, ok := strings.CutPrefix(strings.TrimSpace(first), "#")
	if !ok {
		return ""
	}
	name = spaces.ReplaceAllString(unsafeName.ReplaceAllString(name, " "), " ")
	name = strings.Trim(strings.TrimSpace(name), ". ")
	if runes := []rune(name); len(runes) > maxNameRunes {
		name = strings.TrimSpace(string(runes[:maxNameRunes]))
	}
	return name
}
