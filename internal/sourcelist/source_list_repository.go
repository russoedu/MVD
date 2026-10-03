// Package sourcelist stores the list of playlist/video URLs to download, in
// the app-data folder, one URL per line.
package sourcelist

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Load reads the URL list. Blank lines and # comments are ignored. A missing
// file yields an empty list, not an error.
func Load(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		urls = append(urls, line)
	}
	return urls, scanner.Err()
}

// Save writes the URLs one per line, creating the parent directory. Blank
// entries are dropped.
func Save(path string, urls []string) error {
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	var b strings.Builder
	for _, u := range urls {
		if u = strings.TrimSpace(u); u != "" {
			b.WriteString(u)
			b.WriteByte('\n')
		}
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// Clear removes the saved list file; a missing file is not an error.
func Clear(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
