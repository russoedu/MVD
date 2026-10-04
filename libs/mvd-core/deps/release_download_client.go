package deps

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
)

// downloadFile fetches a release asset over HTTPS into dest, one chunk at a time, so a
// build of a couple of hundred megabytes never sits in memory, and returns the SHA-256
// of what it wrote as hex. A failed or partial download leaves no file behind.
func downloadFile(url, dest string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP status %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	hash := sha256.New()
	_, err = io.Copy(io.MultiWriter(out, hash), resp.Body)
	if closeErr := out.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(dest)

		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
