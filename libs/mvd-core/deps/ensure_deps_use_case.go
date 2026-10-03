package deps

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Ensure puts ./bin on PATH, checks for yt-dlp, ffmpeg and a JS runtime
// (deno or node) and downloads whatever is missing into ./bin.
func Ensure() {
	setupEnvironmentPaths()

	var missing []string

	if _, err := exec.LookPath("yt-dlp"); err != nil {
		missing = append(missing, "yt-dlp")
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		missing = append(missing, "ffmpeg")
	}
	_, denoErr := exec.LookPath("deno")
	_, nodeErr := exec.LookPath("node")
	if denoErr != nil && nodeErr != nil {
		missing = append(missing, "deno")
	}

	if len(missing) > 0 {
		autoInstall(missing)
	}
}

// setupEnvironmentPaths puts ./bin and ./bin/<os> at the front of PATH.
func setupEnvironmentPaths() {
	absBin, err := filepath.Abs("./bin")
	if err == nil {
		pathEnv := os.Getenv("PATH")
		osPathSep := string(os.PathListSeparator)
		osBinPath := filepath.Join(absBin, runtime.GOOS)

		newPath := absBin + osPathSep + osBinPath + osPathSep + pathEnv
		os.Setenv("PATH", newPath)
	}
}

func autoInstall(missing []string) {
	fmt.Printf("\n[!] Missing dependency/dependencies detected: %s\n", strings.Join(missing, ", "))
	fmt.Printf("[+] Automatically downloading dependencies for [%s/%s] into ./bin...\n\n", runtime.GOOS, runtime.GOARCH)

	destDir := "./bin"
	if err := os.MkdirAll(destDir, 0755); err != nil {
		fmt.Printf("Error creating ./bin directory: %v\n", err)
		return
	}

	deps := platformDeps(runtime.GOOS, runtime.GOARCH)

	for _, depName := range missing {
		dep, ok := deps[depName]
		if !ok {
			continue
		}

		destFile := filepath.Join(destDir, dep.FileName)
		fmt.Printf("    Downloading %s from %s...\n", dep.Name, dep.URL)

		data, err := downloadBytes(dep.URL)
		if err != nil {
			fmt.Printf("    [!] Error downloading %s: %v\n", dep.Name, err)
			continue
		}

		if dep.IsZip {
			fmt.Printf("    Extracting %s to %s...\n", dep.FileName, destFile)
			err = extractZipFile(data, dep.FileName, destFile)
			if err != nil {
				fmt.Printf("    [!] Error extracting %s: %v\n", dep.Name, err)
				continue
			}
		} else {
			fmt.Printf("    Writing binary %s...\n", destFile)
			err = os.WriteFile(destFile, data, 0755)
			if err != nil {
				fmt.Printf("    [!] Error writing %s: %v\n", dep.Name, err)
				continue
			}
		}

		os.Chmod(destFile, 0755)
		fmt.Printf("    [OK] Successfully installed %s!\n\n", dep.Name)
	}

	// Refresh PATH
	setupEnvironmentPaths()
}

// extractZipFile writes the first entry of the archive named targetFileName
// to destPath.
func extractZipFile(data []byte, targetFileName, destPath string) error {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}

	for _, f := range r.File {
		baseName := filepath.Base(f.Name)
		if strings.EqualFold(baseName, targetFileName) || strings.EqualFold(f.Name, targetFileName) {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			out, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				return err
			}
			defer out.Close()

			_, err = io.Copy(out, rc)
			return err
		}
	}

	return fmt.Errorf("file '%s' not found inside zip archive", targetFileName)
}
