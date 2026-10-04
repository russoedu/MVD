// Package deps makes sure yt-dlp, ffmpeg and a JavaScript runtime are available,
// downloading prebuilt binaries when they are not, and keeps the ones it owns up to
// date.
package deps

import "time"

// Source says where a tool's newest build is published: an asset of the latest release
// of a GitHub repository.
type Source struct {
	// Repo is "owner/name".
	Repo string
	// Asset is the file name of the build in the latest release.
	Asset string
}

// downloadURL is the address that always serves the newest build of the asset.
func (s Source) downloadURL() string {
	return "https://github.com/" + s.Repo + "/releases/latest/download/" + s.Asset
}

// Dep describes where to get one dependency for a platform.
type Dep struct {
	Name     string
	URL      string
	IsZip    bool
	FileName string
	// Source, when set, makes this a tool the app owns and keeps up to date: it always
	// has its own copy, whatever is on PATH, and replaces it when a newer build is
	// published. Without it the tool is fetched once from URL, only when nothing on
	// PATH provides it, and then left alone.
	Source *Source
	// MinAge is how much newer than the installed copy a published build must be before
	// it replaces it. Zero means any newer build does. It exists for builds that are
	// republished every day, which would otherwise be downloaded again at every start.
	MinAge time.Duration
}

// ffmpegMinAge is how long an ffmpeg build is kept before a newer one replaces it. Its
// builds are republished daily and are over 200 MB, and ffmpeg changes slowly enough
// that a month is soon enough.
const ffmpegMinAge = 30 * 24 * time.Hour

// platformDeps decides which prebuilt binaries fit an OS and architecture.
func platformDeps(targetOS, targetArch string) map[string]Dep {
	deps := make(map[string]Dep)

	// 1. yt-dlp: the standalone builds, which need no Python, from its own releases.
	ytDlpAsset := "yt-dlp_linux"
	switch {
	case targetOS == "windows" && targetArch == "arm64":
		ytDlpAsset = "yt-dlp_arm64.exe"
	case targetOS == "windows":
		ytDlpAsset = "yt-dlp.exe"
	case targetOS == "darwin":
		ytDlpAsset = "yt-dlp_macos"
	case targetArch == "arm64":
		ytDlpAsset = "yt-dlp_linux_aarch64"
	}
	ytDlpName := "yt-dlp"
	if targetOS == "windows" {
		ytDlpName = "yt-dlp.exe"
	}
	ytDlpSource := Source{Repo: "yt-dlp/yt-dlp", Asset: ytDlpAsset}
	deps["yt-dlp"] = Dep{Name: "yt-dlp", URL: ytDlpSource.downloadURL(), FileName: ytDlpName, Source: &ytDlpSource}

	// 2. ffmpeg: on Windows, yt-dlp's own builds, which it recommends and which are
	// republished daily. Elsewhere a pinned build that is fetched once: the same project
	// publishes Linux builds only as .tar.xz and has no macOS build.
	ffmpegName := "ffmpeg"
	if targetOS == "windows" {
		ffmpegName = "ffmpeg.exe"
	}
	switch {
	case targetOS == "windows":
		asset := "ffmpeg-master-latest-win64-gpl.zip"
		if targetArch == "arm64" {
			asset = "ffmpeg-master-latest-winarm64-gpl.zip"
		}
		source := Source{Repo: "yt-dlp/FFmpeg-Builds", Asset: asset}
		deps["ffmpeg"] = Dep{Name: "ffmpeg", URL: source.downloadURL(), IsZip: true, FileName: ffmpegName, Source: &source, MinAge: ffmpegMinAge}
	case targetOS == "darwin":
		deps["ffmpeg"] = Dep{Name: "ffmpeg", URL: "https://github.com/ffbinaries/ffbinaries-prebuilt/releases/download/v4.4.1/ffmpeg-4.4.1-osx-64.zip", IsZip: true, FileName: ffmpegName}
	default: // linux
		deps["ffmpeg"] = Dep{Name: "ffmpeg", URL: "https://github.com/ffbinaries/ffbinaries-prebuilt/releases/download/v4.4.1/ffmpeg-4.4.1-linux-64.zip", IsZip: true, FileName: ffmpegName}
	}

	// 3. deno, only when there is no JavaScript runtime at all. Fetched once.
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
