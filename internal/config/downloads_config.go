package config

import (
	"bufio"
	"os"
	"strings"
)

// LoadDownloads reads downloads.conf and returns the playlist URLs, one per
// line. Empty lines and lines starting with # are ignored.
func LoadDownloads(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		urls = append(urls, line)
	}

	return urls, scanner.Err()
}
