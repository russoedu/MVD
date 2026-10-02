<p align="center">
  <img src="assets/logo.svg" alt="MVD - Music Video Downloader" width="480">
</p>

# YouTube Playlist Downloader (Go + yt-dlp)

A lightweight, zero-setup, concurrent Go application that automatically reads playlist URLs from `downloads.conf`, reads settings from `setup.conf`, self-diagnoses and installs missing dependencies, and downloads playlists in parallel with the best available video and audio quality.

---

## ⚡ Features & Self-Installation

* **Auto-Dependency Installation**: On launch, the app checks for `yt-dlp`, `ffmpeg`, and a JavaScript engine (`deno`). If any dependency is missing, it automatically downloads and extracts pre-built binaries for your operating system into `./bin/` before running.
* **Cross-Platform**: Works out of the box on **Windows**, **macOS** (Intel & Apple Silicon), and **Linux**.
* **Parallel Downloads**: Downloads multiple playlists concurrently using Go goroutines and a worker semaphore pool.
* **YouTube 403 Bypass**: Pre-configured with IPv4 enforcement and JS runtime options to prevent HTTP 403 Forbidden errors.
* **Official Music Video Mode**: Optionally swaps auto-generated "`<Artist> - Topic`" audio tracks for the official music video that YouTube links from the description's **Music** card.
* **Full Screen Interface**: A fixed terminal UI shows every playlist and entry with its state, live progress of the running downloads, global counters (queue, running, done, official, duplicates, failed) and the yt-dlp output of whatever you select. Failed entries can be retried from the screen. Pipes and CI get a plain log instead.

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

## 🖥️ The Interface

```
╭─ MVD · Music Video Downloader ──────────────────────────────────────── 00:12:41 ─╮
│ Queue 82   ⠋ Running 3   ✓ Done 57   ⇄ Official 41   ≡ Dup 3   ✗ Failed 2        │
╰──────────────────────────────────────────────────────────────────────────────────╯
╭─ Playlists ───────────────── 3 ─╮╭─ 90s UK Dance Hits › 07 Stardust ──── follow ─╮
│▸ 90s UK Dance Hits    23/90  ⠋  ││[official] hRvrj_diWYQ -> 5dYAQEAyE1E          │
│  Ibiza Classics       44/44  ✓  ││[youtube] 5dYAQEAyE1E: Downloading webpage     │
│  Trance Anthems        0/61  ·  ││[info] Downloading format 137+251              │
╰─────────────────────────────────╯│[download] Destination: ...                    │
╭─ Entries ────────────────── 90 ─╮│                                               │
│✓ 01 You Don't Know Me        ⇄  ││                                               │
│✓ 02 Don't Call Me Baby       ⇄  ││                                               │
│✓ 03 Modjo - Lady                ││                                               │
│✗ 04 Bob Sinclar - I Feel   ERR  ││                                               │
│≡ 05 Bob Sinclar (Radio)    dup  ││                                               │
│⟲ 06 The Bucketheads       res.  ││                                               │
│⠋ 07 Stardust               34%  ││ ████████░░░░░░░░░░░  34.2% of 112.4MiB at ... │
│· 08 Daft Punk                   ││                                               │
╰─────────────────────────────────╯╰───────────────────────────────────────────────╯
 ↑↓ move  tab pane  ⏎ follow  f failed  r retry  l full log  ? help  q quit
```

* **Header**: global counters and elapsed time. `FINISHED` appears when the queue is empty.
* **Playlists**: one row per URL with `done/total`. The selected playlist drives the entries pane.
* **Entries**: every track of the selected playlist. Glyphs: `·` queued, `⟲` resolving the official video, `⠋` downloading, `⚙` merging, `✓` done, `≡` duplicate skipped, `✗` failed. The right hand tag shows `⇄` when the official video replaced the art track, the download percentage, `dup`, `res.`, `merge` or `ERR`.
* **Output**: the yt-dlp log of the selected entry with a progress bar, speed and ETA. Select a playlist to see its combined log.
* **Keys**: `↑ ↓` move, `tab` switch pane, `enter` toggle follow mode (the selection jumps to whatever starts downloading), `f` show failed entries only, `r` retry the selected failed entry (or every failed entry of the playlist when the playlists pane is focused), `l` full screen output, `?` help, `ctrl+l` redraw the screen, `q` quit.

The screen stays up when everything is finished so the counters and failures remain visible. Every line is also appended to `mvd.log` (see `log_file`).

Terminals narrower than 100 columns show only the lists; press `l` for the output. When stdout is not a terminal, or with `--no-tui` (or `MVD_NO_TUI=1`), the app prints a plain log instead and exits when done.

Emoji in playlist and video titles are not drawn, because terminals disagree on their width and one wrong guess shifts the whole layout. The interface uses Unicode box drawing and status glyphs, so use a terminal with a font that has them (Windows Terminal, iTerm2, GNOME Terminal, kitty, VS Code and most others are fine). A terminal that does not answer colour queries can add a five second pause at start-up; `--no-tui` avoids it.

Downloads are scheduled per entry: `max_concurrent_downloads` is the number of videos in flight across all playlists, filled in playlist order.

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

# Number of videos to download in parallel
max_concurrent_downloads=3

# Replace auto-generated "- Topic" tracks with the official music video
download_official_music_video=false

# Full output log ("off" to disable)
log_file=mvd.log

# Extra flags passed to yt-dlp (space separated)
# -4 enforces IPv4 (prevents YouTube 403 Forbidden errors)
# --js-runtimes deno,node specifies JS runtimes for deciphering
extra_args=-4 --js-runtimes deno,node
```

#### Official music video mode

Many playlists contain auto-generated uploads from `<Artist> - Topic` channels: a still image with the audio track, whose description ends with *"Auto-generated by YouTube"*. Below that description YouTube shows a **Music** card that links to the official video of the song.

With `download_official_music_video=true` the app, for every playlist:

1. Lists the playlist with `yt-dlp --flat-playlist` (nothing is downloaded yet).
2. For every entry uploaded by a `- Topic` channel, crawls the watch page (and, if the page comes back stripped down, the same `youtubei/v1/next` call the page makes) and reads the video linked from the **Music** card. Links written in the description text are used as a fallback.
3. Confirms through YouTube's oEmbed endpoint that the linked video is not another auto-generated track and still exists.
4. Downloads the official video instead of the art track. Entries from normal channels, and tracks without an official video, are downloaded unchanged. If two tracks of a playlist point to the same official video (radio edit and extended mix, for example) it is downloaded once.

Videos are always downloaded one by one, so `%(playlist_title)s`, `%(playlist_index)s` and the other playlist fields of `output_template` are filled in from the playlist listing and files land exactly where a playlist download would put them.

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
go build -o downloader.exe .

# Run the tests
go test ./...

# Plain log output (no full screen interface)
.\downloader.exe --no-tui
```

GitHub Releases are automatically created via GitHub Actions on every new tag push (e.g., `v1.0.0`).
