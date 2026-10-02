# YouTube Playlist Downloader (Go + yt-dlp)

A lightweight, zero-setup, concurrent Go application that automatically reads playlist URLs from `downloads.conf`, reads settings from `setup.conf`, self-diagnoses and installs missing dependencies, and downloads playlists in parallel with the best available video and audio quality.

---

## ⚡ Features & Self-Installation

* **Auto-Dependency Installation**: On launch, the app checks for `yt-dlp`, `ffmpeg`, and a JavaScript engine (`deno`). If any dependency is missing, it automatically downloads and extracts pre-built binaries for your operating system into `./bin/` before running.
* **Cross-Platform**: Works out of the box on **Windows**, **macOS** (Intel & Apple Silicon), and **Linux**.
* **Parallel Downloads**: Downloads multiple playlists concurrently using Go goroutines and a worker semaphore pool.
* **YouTube 403 Bypass**: Pre-configured with IPv4 enforcement and JS runtime options to prevent HTTP 403 Forbidden errors.

---

## 🚀 Quick Start

Simply run the main application. It will detect your system, install any missing dependencies, and start downloading your playlists:

```powershell
go run main.go
```

Or run the compiled executable directly:
```powershell
.\downloader.exe
```

---

## ⚙️ Configuration Files

### 1. `setup.conf`
Defines output paths, download quality, format merging, concurrency limits, and extra flags for `yt-dlp`.

```ini
# setup.conf - Downloader Configuration

# Directory where downloaded videos will be saved
output_dir=../DJ/new

# Quality setting for yt-dlp (-f option)
quality=bestvideo+bestaudio/best

# Container format to merge video and audio streams into (e.g. mp4, mkv)
merge_output_format=mp4

# Output filename template for yt-dlp (-o option)
output_template=%(title)s.%(ext)s

# Number of playlists to download in parallel
max_concurrent_downloads=3

# Extra flags passed to yt-dlp (space separated)
# -4 enforces IPv4 (prevents YouTube 403 Forbidden errors)
# --js-runtimes deno,node specifies JS runtimes for deciphering
extra_args=-4 --js-runtimes deno,node
```

### 2. `downloads.conf`
Contains the list of YouTube playlist URLs to download (one URL per line). Empty lines and lines starting with `#` are ignored.

```txt
https://youtube.com/playlist?list=PLYPcrcIixkLEaupMLBEt3GaajHGVTHeF9
https://youtube.com/playlist?list=PLsc4x0rSyZsNF6WV5rBk2M41W12ox0nhq
```

---

## 🛠️ Building & Releasing

```powershell
# Build for your OS
go build -o downloader.exe main.go
```

GitHub Releases are automatically created via GitHub Actions on every new tag push (e.g., `v1.0.0`).
