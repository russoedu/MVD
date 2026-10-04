// Package deps makes sure yt-dlp, ffmpeg and a JavaScript runtime are
// available, downloading prebuilt binaries into ./bin when they are not.
package deps

// Dep describes where to get one dependency for a platform.
type Dep struct {
	Name     string
	URL      string
	IsZip    bool
	FileName string
}

// platformDeps decides which prebuilt binaries fit an OS and architecture.
func platformDeps(targetOS, targetArch string) map[string]Dep {
	deps := make(map[string]Dep)

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
	deps["yt-dlp"] = Dep{Name: "yt-dlp", URL: ytDlpURL, IsZip: false, FileName: ytDlpName}

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
	deps["ffmpeg"] = Dep{Name: "ffmpeg", URL: ffmpegURL, IsZip: true, FileName: ffmpegName}

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
	deps["deno"] = Dep{Name: "deno", URL: denoURL, IsZip: true, FileName: denoName}

	return deps
}
