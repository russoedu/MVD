package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

// Config holds settings read from setup.conf
type Config struct {
	OutputDir              string
	Quality                string
	MergeOutputFormat      string
	OutputTemplate         string
	MaxConcurrentDownloads int
	ExtraArgs              []string
	// DownloadOfficialMusicVideo replaces auto-generated "- Topic" art
	// tracks with the official music video linked from their description.
	DownloadOfficialMusicVideo bool
}

// defaultConfig returns fallback settings if setup.conf lacks them
func defaultConfig() Config {
	return Config{
		OutputDir:              "./downloads",
		Quality:                "bestvideo+bestaudio/best",
		MergeOutputFormat:      "mp4",
		OutputTemplate:         "%(playlist_title,playlist)s/%(playlist_index)02d - %(title)s.%(ext)s",
		MaxConcurrentDownloads: 3,
		ExtraArgs:              []string{"-4", "--js-runtimes", "deno,node"},
	}
}

// setupEnvironmentPaths ensures local ./bin and ./bin/<os> directories are on PATH
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

type DepInfo struct {
	Name     string
	URL      string
	IsZip    bool
	FileName string
}

func getPlatformDeps(targetOS, targetArch string) map[string]DepInfo {
	deps := make(map[string]DepInfo)

	// 1. yt-dlp
	var ytDlpURL string
	switch targetOS {
	case "windows":
		ytDlpURL = "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp.exe"
	case "darwin":
		ytDlpURL = "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_macos"
	default: // linux
		if targetArch == "arm64" {
			ytDlpURL = "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux_aarch64"
		} else {
			ytDlpURL = "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp"
		}
	}
	ytDlpName := "yt-dlp"
	if targetOS == "windows" {
		ytDlpName = "yt-dlp.exe"
	}
	deps["yt-dlp"] = DepInfo{Name: "yt-dlp", URL: ytDlpURL, IsZip: false, FileName: ytDlpName}

	// 2. ffmpeg
	var ffmpegURL string
	switch targetOS {
	case "windows":
		ffmpegURL = "https://github.com/ffbinaries/ffbinaries-prebuilt/releases/download/v4.4.1/ffmpeg-4.4.1-win-64.zip"
	case "darwin":
		ffmpegURL = "https://github.com/ffbinaries/ffbinaries-prebuilt/releases/download/v4.4.1/ffmpeg-4.4.1-osx-64.zip"
	default: // linux
		ffmpegURL = "https://github.com/ffbinaries/ffbinaries-prebuilt/releases/download/v4.4.1/ffmpeg-4.4.1-linux-64.zip"
	}
	ffmpegName := "ffmpeg"
	if targetOS == "windows" {
		ffmpegName = "ffmpeg.exe"
	}
	deps["ffmpeg"] = DepInfo{Name: "ffmpeg", URL: ffmpegURL, IsZip: true, FileName: ffmpegName}

	// 3. deno
	var denoURL string
	switch targetOS {
	case "windows":
		denoURL = "https://github.com/denoland/deno/releases/latest/download/deno-x86_64-pc-windows-msvc.zip"
	case "darwin":
		if targetArch == "arm64" {
			denoURL = "https://github.com/denoland/deno/releases/latest/download/deno-aarch64-apple-darwin.zip"
		} else {
			denoURL = "https://github.com/denoland/deno/releases/latest/download/deno-x86_64-apple-darwin.zip"
		}
	default: // linux
		denoURL = "https://github.com/denoland/deno/releases/latest/download/deno-x86_64-unknown-linux-gnu.zip"
	}
	denoName := "deno"
	if targetOS == "windows" {
		denoName = "deno.exe"
	}
	deps["deno"] = DepInfo{Name: "deno", URL: denoURL, IsZip: true, FileName: denoName}

	return deps
}

func downloadBytes(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

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

func autoInstallDependencies(missing []string) {
	fmt.Printf("\n[!] Missing dependency/dependencies detected: %s\n", strings.Join(missing, ", "))
	fmt.Printf("[+] Automatically downloading dependencies for [%s/%s] into ./bin...\n\n", runtime.GOOS, runtime.GOARCH)

	destDir := "./bin"
	if err := os.MkdirAll(destDir, 0755); err != nil {
		fmt.Printf("Error creating ./bin directory: %v\n", err)
		return
	}

	platformDeps := getPlatformDeps(runtime.GOOS, runtime.GOARCH)

	for _, depName := range missing {
		dep, ok := platformDeps[depName]
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

func ensureDependencies() {
	setupEnvironmentPaths()

	var missing []string

	// Check yt-dlp
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		missing = append(missing, "yt-dlp")
	}

	// Check ffmpeg
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		missing = append(missing, "ffmpeg")
	}

	// Check JS runtime (deno or node)
	_, denoErr := exec.LookPath("deno")
	_, nodeErr := exec.LookPath("node")
	if denoErr != nil && nodeErr != nil {
		missing = append(missing, "deno")
	}

	if len(missing) > 0 {
		autoInstallDependencies(missing)
	}
}

// loadSetupConfig reads setup.conf and parses key=value pairs
func loadSetupConfig(filePath string) (Config, error) {
	cfg := defaultConfig()

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Warning: '%s' not found. Using default configurations.\n", filePath)
			return cfg, nil
		}
		return cfg, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)

		switch key {
		case "output_dir":
			if val != "" {
				cfg.OutputDir = val
			}
		case "quality":
			if val != "" {
				cfg.Quality = val
			}
		case "merge_output_format":
			if val != "" {
				cfg.MergeOutputFormat = val
			}
		case "output_template":
			if val != "" {
				cfg.OutputTemplate = val
			}
		case "max_concurrent_downloads":
			if n, err := strconv.Atoi(val); err == nil && n > 0 {
				cfg.MaxConcurrentDownloads = n
			}
		case "extra_args":
			if val != "" {
				cfg.ExtraArgs = strings.Fields(val)
			}
		case "download_official_music_video":
			cfg.DownloadOfficialMusicVideo = parseBool(val)
		}
	}

	return cfg, scanner.Err()
}

// parseBool accepts the usual spellings of a boolean config value.
func parseBool(val string) bool {
	switch strings.ToLower(strings.TrimSpace(val)) {
	case "1", "true", "yes", "on", "y":
		return true
	}
	return false
}

// loadDownloads reads downloads.conf and returns list of playlist URLs
func loadDownloads(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
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

func main() {
	// 1. Auto-check & download any missing dependencies
	ensureDependencies()

	setupFile := "setup.conf"
	downloadsFile := "downloads.conf"

	// 2. Verify yt-dlp location
	ytDlpPath, err := exec.LookPath("yt-dlp")
	if err != nil {
		fmt.Println("Error: 'yt-dlp' executable could not be found or installed.")
		os.Exit(1)
	}
	fmt.Printf("Found yt-dlp at: %s\n", ytDlpPath)

	// 3. Load settings from setup.conf
	cfg, err := loadSetupConfig(setupFile)
	if err != nil {
		fmt.Printf("Error reading %s: %v\n", setupFile, err)
		os.Exit(1)
	}

	fmt.Println("\n--- Downloader Configuration ---")
	fmt.Printf("Output Directory:         %s\n", cfg.OutputDir)
	fmt.Printf("Quality:                  %s\n", cfg.Quality)
	fmt.Printf("Merge Output Format:      %s\n", cfg.MergeOutputFormat)
	fmt.Printf("Output Template:          %s\n", cfg.OutputTemplate)
	fmt.Printf("Max Concurrent Downloads: %d\n", cfg.MaxConcurrentDownloads)
	fmt.Printf("Extra Arguments:          %v\n", cfg.ExtraArgs)
	fmt.Printf("Official Music Video:     %v\n", cfg.DownloadOfficialMusicVideo)
	fmt.Println("--------------------------------")

	// 4. Ensure output directory exists
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory '%s': %v\n", cfg.OutputDir, err)
		os.Exit(1)
	}

	// 5. Load URLs from downloads.conf
	urls, err := loadDownloads(downloadsFile)
	if err != nil {
		fmt.Printf("Error reading %s: %v\n", downloadsFile, err)
		os.Exit(1)
	}

	if len(urls) == 0 {
		fmt.Printf("No valid playlist URLs found in '%s'.\n", downloadsFile)
		return
	}

	fmt.Printf("Found %d playlist(s) to process (%d parallel worker(s)).\n\n", len(urls), cfg.MaxConcurrentDownloads)

	// 6. Download playlists in parallel with limited concurrency
	var successful int64
	var failed int64

	outPattern := filepath.Join(cfg.OutputDir, cfg.OutputTemplate)
	sem := make(chan struct{}, cfg.MaxConcurrentDownloads)
	var wg sync.WaitGroup
	var logMutex sync.Mutex

	logMsg := func(format string, a ...interface{}) {
		logMutex.Lock()
		defer logMutex.Unlock()
		fmt.Printf(format, a...)
	}

	var resolver *OfficialResolver
	var stats videoStats
	if cfg.DownloadOfficialMusicVideo {
		resolver = newOfficialResolver(logMsg)
	}

	for i, url := range urls {
		wg.Add(1)
		sem <- struct{}{} // Acquire worker slot

		go func(idx int, playlistURL string) {
			defer wg.Done()
			defer func() { <-sem }() // Release worker slot

			logMsg("[PLAYLIST %d/%d] Starting download: %s\n", idx+1, len(urls), playlistURL)

			var err error
			if cfg.DownloadOfficialMusicVideo {
				err = downloadPlaylistOfficial(ytDlpPath, cfg, outPattern, playlistURL, resolver, logMsg, &stats)
			} else {
				err = runYtDlp(ytDlpPath, buildDownloadArgs(cfg, outPattern, playlistURL))
			}

			if err != nil {
				logMsg("\n[ERROR] Playlist #%d (%s) failed: %v\n\n", idx+1, playlistURL, err)
				atomic.AddInt64(&failed, 1)
			} else {
				logMsg("\n[SUCCESS] Playlist #%d completed successfully!\n\n", idx+1)
				atomic.AddInt64(&successful, 1)
			}
		}(i, url)
	}

	wg.Wait()

	// 7. Print summary
	fmt.Println("==================================================")
	fmt.Println("Download Summary:")
	fmt.Printf("Total Playlists: %d\n", len(urls))
	fmt.Printf("Successful:      %d\n", successful)
	fmt.Printf("Failed:          %d\n", failed)
	if cfg.DownloadOfficialMusicVideo {
		fmt.Printf("Videos downloaded:        %d\n", atomic.LoadInt64(&stats.downloaded))
		fmt.Printf("  replaced by official:   %d\n", atomic.LoadInt64(&stats.replaced))
		fmt.Printf("  kept original:          %d\n", atomic.LoadInt64(&stats.keptOriginal))
		fmt.Printf("  skipped duplicates:     %d\n", atomic.LoadInt64(&stats.duplicates))
		fmt.Printf("Videos failed:            %d\n", atomic.LoadInt64(&stats.failed))
	}
	fmt.Println("==================================================")
}

// videoStats counts per-video outcomes in official music video mode.
type videoStats struct {
	downloaded   int64
	replaced     int64
	keptOriginal int64
	duplicates   int64
	failed       int64
}

// buildDownloadArgs assembles the yt-dlp command line. The trailing
// arguments (flags and the URL to download) are appended as given.
func buildDownloadArgs(cfg Config, outPattern string, trailing ...string) []string {
	args := []string{
		"-f", cfg.Quality,
		"-o", outPattern,
	}
	if cfg.MergeOutputFormat != "" {
		args = append(args, "--merge-output-format", cfg.MergeOutputFormat)
	}
	if len(cfg.ExtraArgs) > 0 {
		args = append(args, cfg.ExtraArgs...)
	}
	return append(args, trailing...)
}

func runYtDlp(ytDlpPath string, args []string) error {
	cmd := exec.Command(ytDlpPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// downloadPlaylistOfficial lists the playlist, swaps every auto-generated
// art track for the official music video linked from its description, and
// downloads the videos one by one so playlist folder/index fields keep
// working in the output template.
func downloadPlaylistOfficial(ytDlpPath string, cfg Config, outPattern, playlistURL string, resolver *OfficialResolver, logMsg func(string, ...interface{}), stats *videoStats) error {
	entries, err := listPlaylistEntries(ytDlpPath, playlistURL, cfg.ExtraArgs)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return fmt.Errorf("playlist is empty or could not be listed")
	}
	logMsg("    Playlist %q has %d entries, resolving official videos...\n", entries[0].PlaylistTitle, len(entries))

	seen := make(map[string]bool, len(entries))
	var failures int
	for _, e := range entries {
		targetID := e.ID
		label := "original"

		if isAutoGenerated(e) {
			if official, reason := resolver.Resolve(e.ID); official != "" {
				targetID = official
				label = "official"
				logMsg("    [%02d] %s -> %s (%s)\n", e.PlaylistIndex, e.ID, official, reason)
			} else {
				logMsg("    [%02d] %s kept as is (%s)\n", e.PlaylistIndex, e.ID, reason)
			}
		}

		if seen[targetID] {
			logMsg("    [%02d] %s already downloaded for this playlist, skipping duplicate\n", e.PlaylistIndex, targetID)
			atomic.AddInt64(&stats.duplicates, 1)
			continue
		}
		seen[targetID] = true

		args := buildDownloadArgs(cfg, applyPlaylistFields(outPattern, e), "--no-playlist", "https://www.youtube.com/watch?v="+targetID)

		if err := runYtDlp(ytDlpPath, args); err != nil {
			logMsg("    [ERROR] %s (%s for entry %02d) failed: %v\n", targetID, label, e.PlaylistIndex, err)
			atomic.AddInt64(&stats.failed, 1)
			failures++
			continue
		}
		atomic.AddInt64(&stats.downloaded, 1)
		if label == "official" {
			atomic.AddInt64(&stats.replaced, 1)
		} else {
			atomic.AddInt64(&stats.keptOriginal, 1)
		}
	}

	if failures > 0 {
		return fmt.Errorf("%d of %d videos failed", failures, len(entries))
	}
	return nil
}
