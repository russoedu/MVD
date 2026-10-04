<p align="center">
  <img src="assets/logo.svg" alt="MVD - Music Video Downloader" width="480">
</p>

# YouTube Playlist Downloader (Go + yt-dlp)

A lightweight, zero-setup, concurrent Go application with interactive terminal screens: paste a list of playlists/videos, tune settings in a preferences screen, and watch a live download dashboard. It self-diagnoses and installs missing dependencies, borrows your browser's YouTube cookies automatically, downloads in parallel, and keeps its config and list in your OS preferences folder.

---

## ⚡ Features & Self-Installation

* **Auto-Dependency Installation**: On launch, the app checks for `yt-dlp`, `ffmpeg`, and a JavaScript engine (`deno`). If any dependency is missing, it automatically downloads and extracts pre-built binaries for your operating system into `./bin/` before running.
* **Cross-Platform**: Works out of the box on **Windows**, **macOS** (Intel & Apple Silicon), and **Linux**.
* **Parallel Downloads**: Downloads multiple playlists concurrently using Go goroutines and a worker semaphore pool.
* **YouTube 403 Bypass**: Pre-configured with IPv4 enforcement and JS runtime options to prevent HTTP 403 Forbidden errors.
* **Official Music Video Mode**: Optionally swaps auto-generated "`<Artist> - Topic`" audio tracks for the official music video that YouTube links from the description's **Music** card.
* **Automatic browser cookies**: Finds a browser you're signed into YouTube with and uses its cookies to clear bot checks and `429` errors, with no configuration.
* **Auto-retry**: Retries one-off failures at once and rate-limited ones in a sweep after the backlog finishes; never retries permanently gone videos.
* **Full Screen Interface**: A fixed terminal UI shows every playlist and entry with its state, live progress of the running downloads, global counters (queue, running, done, official, duplicates, failed) and the yt-dlp output of whatever you select. Failed entries can be retried from the screen. Pipes and CI get a plain log instead.

---

## 🚀 Quick Start

Build and run; on first launch it creates a default config in your preferences folder and opens the **preferences** screen:

```powershell
go build -o mvd.exe ./apps/mvd-cli
.\mvd.exe
```

Then: paste your playlist/video URLs on the **list** screen (one per line), press `Ctrl+S` to start, and the **download** dashboard takes over. Everything is kept in the app-data folder (see below) — there are no config files next to the binary.

For an unattended/headless run (pipes, CI, cron) it downloads the saved list with a plain log instead of the screens:

```powershell
.\mvd.exe --no-tui
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

Downloads are scheduled per entry: `max_concurrent_downloads` is the number of videos in flight across all playlists, filled in playlist order. Each video also fetches `concurrent_fragments` fragments in parallel.

## 🌐 The browser app (`mvd-tray`)

`apps/mvd-tray` is a second front end for the same engine. It runs until you quit it from its tray icon (or press Ctrl+C in the console), serves a page on `http://127.0.0.1:8421` and opens it in your browser. Paste links into the page at any time, including while it is downloading, and watch the queue fill and progress live. It reads the same `config.conf` as the terminal app.

```powershell
npx nx run mvd-tray:build          # builds the React page, embeds it, builds the binary
./dist/apps/mvd-tray/mvd-tray      # flags: -addr 127.0.0.1:8421  -no-browser  -no-tray
npx nx run mvd-tray:dev            # development: Vite on :4200 proxying /api to the Go app
```

**Tools and updates.** The app keeps its own copy of yt-dlp and ffmpeg in a `bin` folder inside the app-data folder (`%AppData%\mvd\bin` on Windows, `~/Library/Application Support/mvd/bin` on macOS, `~/.config/mvd/bin` on Linux), whatever is on your `PATH`, so it can keep them current without touching anything you installed yourself. A folder you can always write to, so no administrator rights are involved, and the same place on every start. deno is added only if you have neither deno nor node.

The first run downloads them from their official GitHub releases (on Windows about 310 MB: yt-dlp 17 MB, ffmpeg 200 MB, deno 93 MB), checks each against the SHA-256 that GitHub publishes for it, and says what it is doing in a notification. After that it looks for newer versions in the background each time it starts, which never holds up the page or a download, and again when a download fails (at most every ten minutes), because an out-of-date yt-dlp is the usual reason a video that worked yesterday does not today. If it replaced something after a failure it queues the failed downloads again and tells you. A new yt-dlp replaces the old one as soon as it is published. ffmpeg is replaced only by a build at least 30 days newer, because its builds are republished daily and are 200 MB. The old program is moved aside rather than deleted, so an update works even while a download is using it, and a check or download that fails changes nothing.

Limits: ffmpeg is kept current on Windows only (the project that publishes the builds has no macOS build and ships Linux ones in a format this does not unpack yet, so those fetch a pinned 4.4.1 once). deno is fetched once and never updated. The terminal app keeps using `./bin` and the tools on your `PATH`, and does not update them.

Starting it a second time opens the running one instead. The server only answers to `localhost`: a request is refused unless its Host is a loopback name, any Origin is a loopback page, and anything that changes state is `application/json`, so a web page on another site cannot read your queue or add to it.

The page is `apps/mvd-web` (React); the HTTP API is `libs/mvd-server`: `GET /api/state`, `GET /api/events` (server-sent snapshots), `POST /api/sources`, `POST /api/entries/{id}/retry`, `POST /api/playlists/{index}/retry`. The **Settings** tab edits the same `config.conf` as the terminal app: folders (with a Browse button that opens the operating system's own folder chooser: PowerShell on Windows, `osascript` on macOS, `zenity` or `kdialog` on Linux; if none is present you type the path), quality, file format and name template, how many downloads run at once, cookies, retries and the log. Raw yt-dlp arguments and the cookie file path are not shown and are never changed by saving. The running downloads keep the settings they started with, so the page tells you to restart MVD for a change to reach them. `GET`/`PUT /api/settings` and `POST /api/folders/pick` back this tab. The tray icon opens the page when clicked and has **Open MVD** and **Quit**; where there is no system tray (a server, a bare window manager) it says so and runs until Ctrl+C, and `-no-tray` does that on purpose. The icon is drawn in code, so there is no image to ship. On macOS the tray needs a C toolchain to build (it is native Cocoa), which is why the app is built on a runner of each OS; Windows and Linux build without one. The Windows release is a windowed program, so no console window opens next to the icon; a start-up error is shown in a message box instead, and when it is started from a terminal it also prints there. That comes from `-H=windowsgui` in the `build-native` target of `apps/mvd-tray/project.json`, which is edited by hand (mnci cannot pass the flag yet), and from `libs/mvd-core/procwindow`, which keeps yt-dlp and the folder chooser from opening console windows of their own. Plain `nx run mvd-tray:build` and `go run` stay console builds, so you see the logs.

## ⚙️ Configuration

Config and the download list live in your OS preferences folder, created on first run — nothing sits next to the binary:

| OS | Folder |
|---|---|
| Windows | `%AppData%\mvd\` |
| macOS | `~/Library/Application Support/mvd/` |
| Linux | `~/.config/mvd/` |

It holds `config.conf` (settings), `list.txt` (your URLs) and `cookies.txt` (the exported browser cookies, private — keep it safe). You normally never touch these by hand; edit everything in the app. A legacy `setup.conf`/`downloads.conf` next to the binary is imported once on first run.

### Screens and keys

- **List** — paste/type URLs, one per line. `Ctrl+S` start · `Ctrl+P` preferences · `Ctrl+R` reset · `Ctrl+Q`/`Esc` quit. (Ctrl here because Return makes a new line.)
- **Preferences** — `↑↓` move · `Enter` edit/toggle · `a` advanced · `s` save · `Esc` cancel. Booleans toggle on Enter; quality/merge/cookies open a radio selector; the output and log folders open a folder navigator (`↑↓` move, `→` open, `←` up, `n` new folder, `Enter` choose, `Esc` cancel).
- **Advanced** — raw extra yt-dlp args, parallel fragments and auto-retry. `s` save · `Esc` back.
- **Download** — the live dashboard (see above); `q` returns to the list.

> Keys are bare single letters where you aren't typing; `Ctrl` is used only on the list editor. A terminal can't receive the Cmd key on macOS, so `Ctrl` is used on every platform.

### `config.conf` keys (for reference)

`output_dir`, `video_quality` (best/2160p/1440p/1080p/720p/480p), `audio_quality` (best/high/medium/low), `raw_format` (raw `-f` override), `merge_output_format`, `output_template`, `max_concurrent_downloads`, `concurrent_fragments` (`off` to disable), `download_official_music_video`, `auto_retry`, `cookies_from_browser` (`all`/`off`/a browser name), `cookies_file`, `create_log_file`, `log_dir`, `extra_args`.

### Browser cookies

YouTube rate limits heavy use and answers with `Sign in to confirm you're not a bot` or `HTTP Error 429`. The way around it is to run as a signed-in user by borrowing a browser's cookies. **This is on by default and needs no configuration**: at start the app tries every installed browser and uses the first one with a live YouTube login, saving it to `cookies_file` (default `cookies.txt`, ignored by git, keep it private: it holds your session). Later runs reuse that file; delete it to refresh. Every yt-dlp run and the official video resolver use it.

To pin one browser, set its name; to turn cookies off, set `off`:

```ini
cookies_from_browser=firefox   # or: edge, chrome, brave, chromium, opera, vivaldi, firefox:default
```

You can also drop your own cookie file (exported with a browser extension) at `cookies_file`; an existing file is reused as-is.

**Caveat:** Chrome and Edge 127+ encrypt their cookies (App-Bound Encryption) and yt-dlp often cannot read them, even with the browser closed. **Firefox is the reliable source.** If auto mode finds nothing, sign in to YouTube in Firefox, or export a cookie file manually.

### Auto-retry

Failed downloads are retried automatically (`auto_retry=on` by default). A one-off glitch is retried immediately; a rate-limited failure (`429`, bot check) is retried in a single sweep after the whole backlog finishes, once a cooldown lets the limit window reset; a permanent failure (private, removed, geo-blocked) is never retried. The header and summary show a **Retried** count. Set `auto_retry=off` to fail and move on instead.

### Official music video mode

Many playlists contain auto-generated uploads from `<Artist> - Topic` channels: a still image with the audio track, whose description ends with *"Auto-generated by YouTube"*. Below that description YouTube shows a **Music** card that links to the official video of the song.

With `download_official_music_video=true` the app, for every playlist:

1. Lists the playlist with `yt-dlp --flat-playlist` (nothing is downloaded yet).
2. For every entry uploaded by a `- Topic` channel, crawls the watch page (and, if the page comes back stripped down, the same `youtubei/v1/next` call the page makes) and reads the video linked from the **Music** card. Links written in the description text are used as a fallback.
3. If the page came back without the card (YouTube serves a stripped page to clients it distrusts), fetches it again through yt-dlp itself, so the browser cookies and yt-dlp's bot-check workarounds apply.
4. Confirms through YouTube's oEmbed endpoint that the linked video is not another auto-generated track and still exists.
5. Downloads the official video instead of the art track. Entries from normal channels, and tracks without an official video, are downloaded unchanged. If two tracks of a playlist point to the same official video (radio edit and extended mix, for example) it is downloaded once. When the official video exists but cannot be downloaded (blocked, removed, private), the original art track is downloaded instead and the log says so.

Videos are always downloaded one by one, so `%(playlist_title)s`, `%(playlist_index)s` and the other playlist fields of `output_template` are filled in from the playlist listing and files land exactly where a playlist download would put them.

---

## 🗂️ Code Layout

The repository is an [mnci](https://github.com/russoedu/MoNecromanCi) (Nx) workspace with one Go module at the root: `apps/` holds the programs (`apps/mvd-cli`, the terminal app; `apps/mvd-tray`, the browser app, with its React page in `apps/mvd-web`) and `libs/` the code they share (`libs/mvd-core`, the engine; `libs/mvd-server`, what the browser app adds), so each front end reuses the engine instead of copying it. `npx nx run-many -t test,build` builds and tests all of it.

The code follows vertical feature slices: `apps/mvd-cli/main.go` only wires things together, and every folder under `libs/mvd-core/` is one slice that owns one outcome. A slice is flat, and each file is named `<name>_<role>.go` so the role says what the file does (`use_case` coordinates an operation, `policy` is a reusable decision, `algorithm` is pure computation, `mapper` converts representations, `contract` is data crossing a boundary, `client` talks to an external service, `store` holds runtime state, `repository` persists, `handler` adapts a transport such as the keyboard, `config` and `enum` are what they say).

| Slice | Outcome |
|---|---|
| `libs/mvd-core/appdir` | Locate the OS app-data folder and the Downloads folder. |
| `libs/mvd-core/config` | Load/create/save the config; compile quality presets to a yt-dlp `-f`. |
| `libs/mvd-core/sourcelist` | Load/save/clear the saved URL list. |
| `libs/mvd-core/runner` | Assemble a ready-to-run engine (cookies, resolver, options) from a config. |
| `libs/mvd-core/deps` | Make yt-dlp, ffmpeg and a JavaScript runtime available, check what it downloads against the published checksum, and keep the ones the app owns up to date. |
| `libs/mvd-core/ytdlp` | Run yt-dlp: list a playlist, download one video with captured output, decode progress lines, render the output template, export cookies, dump pages. |
| `libs/mvd-core/cookies` | Acquire a YouTube cookie file by trying the installed browsers and keeping the first with a live login. |
| `libs/mvd-core/official` | Find the official music video of an auto-generated art track by crawling the watch page. |
| `libs/mvd-core/engine` | Download every playlist: queue, worker pool, duplicate detection, retries, the `mvd.log` file, and the events every renderer consumes. |
| `libs/mvd-core/runstate` | Mirror engine events into a state renderers can draw, plus human readable sizes and times. |
| `libs/mvd-core/plain` | Print the run as a plain log (pipes, CI, `--no-tui`). |
| `libs/mvd-core/tui` | The interactive screens: list, preferences, advanced, folder picker and the download dashboard. |
| `libs/mvd-server/snapshot` | Turn the run state into a versioned, JSON-friendly snapshot for browsers. |
| `libs/mvd-server/session` | Own the one long-lived engine: start it on the first URLs, accept more while it runs, publish each change. |
| `libs/mvd-server/settings` | Read, validate and save the common settings in `config.conf`, keeping the advanced ones. |
| `libs/mvd-server/api` | The localhost HTTP API and its request guard. |

Dependencies point one way: `main` → `engine` → `ytdlp`; `main` → `official`, injected into the engine through a small port interface so the engine never imports it; `tui` and `plain` → `runstate` → `engine`. No two slices import each other. Tests sit next to the file they test; the `ytdlp` and `engine` test binaries double as a stub `yt-dlp`, so the suite runs on every platform without shell scripts.

## 🛠️ Building & Releasing

```powershell
# Build for your OS
go build -o downloader.exe ./apps/mvd-cli

# Run the tests (Nx runs each project from its own folder; a bare `go test ./...`
# from the root would also walk node_modules once `npm install` has run)
npx nx run-many -t test

# Plain log output (no full screen interface)
.\downloader.exe --no-tui
```

GitHub Releases are automatically created via GitHub Actions on every new tag push (e.g., `v1.0.0`).

Two things release, separately. The terminal app is built by `release.yml` on every `v*` tag as before (it tests only `mvd-cli` and `libs`, because `mvd-tray` cannot compile until its React page has been built into `apps/mvd-tray/web`). The browser app is released by `ci.yml` (mnci): a push to `main` versions `mvd-tray` from conventional commits, tags it and attaches a zip per OS built on a runner of that OS (it is a `--cgo` app, so it is not cross-compiled). Nx therefore leaves `mvd-tray` out of the cross-compiling verify job and builds it in the `native` job instead.
