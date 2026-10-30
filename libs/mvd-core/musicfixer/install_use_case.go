package musicfixer

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

//go:embed all:bundle
var bundled embed.FS

// ErrNotBundled says this build of MVD carries no music-fixer for its system.
var ErrNotBundled = errors.New("this build of MVD has no music-fixer for this system")

// Install puts the bundled music-fixer in dataDir and returns the path of its program. A folder
// already made from the same bundle is kept as it is.
func Install(dataDir string) (string, error) {
	zipName := fmt.Sprintf("bundle/music-fixer-%s-%s.zip", runtime.GOOS, runtime.GOARCH)
	data, err := bundled.ReadFile(zipName)
	if err != nil {
		return "", ErrNotBundled
	}
	return unpack(data, dataDir)
}

func unpack(data []byte, dataDir string) (string, error) {
	sum := sha256.Sum256(data)
	stamp := hex.EncodeToString(sum[:])
	exe := filepath.Join(dataDir, "music-fixer", "music-fixer")
	if runtime.GOOS == "windows" {
		exe += ".exe"
	}
	stampFile := filepath.Join(dataDir, "music-fixer", ".bundle")
	if old, err := os.ReadFile(stampFile); err == nil && string(old) == stamp {
		if _, err := os.Stat(exe); err == nil {
			return exe, nil
		}
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}
	if err := os.RemoveAll(filepath.Join(dataDir, "music-fixer")); err != nil {
		return "", err
	}
	for _, f := range zr.File {
		target := filepath.Join(dataDir, filepath.FromSlash(f.Name))
		if !strings.HasPrefix(target, filepath.Clean(dataDir)+string(os.PathSeparator)) {
			return "", fmt.Errorf("music-fixer bundle has a file outside its folder: %s", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return "", err
			}
			continue
		}
		if err := writeEntry(f, target); err != nil {
			return "", err
		}
	}
	if _, err := os.Stat(exe); err != nil {
		return "", errors.New("music-fixer bundle has no program in it")
	}
	return exe, os.WriteFile(stampFile, []byte(stamp), 0o644)
}

func writeEntry(f *zip.File, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	in, err := f.Open()
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode()|0o600)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}
