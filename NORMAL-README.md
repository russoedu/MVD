<p align="center">
  <img src="assets/mvd-name.jpg" alt="MVD - Music Video Downloader" width="480">
</p>

# Music Video Downloader (Go + yt-dlp) - the boring description...

A lightweight, zero-setup, concurrent Go application with two front ends, a tray app with a browser page (`mvd`) and a terminal app (`mvd-tui`). The terminal app has interactive screens: paste a list of playlists/videos, tune settings in a preferences screen, and watch a live download dashboard. It self-diagnoses and installs missing dependencies, borrows your browser's YouTube cookies automatically, downloads in parallel, and keeps its config and list in your OS preferences folder.

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

## 🌐 The tray app (`mvd`)

`apps/mvd` builds the program `mvd`, the main app: a tray icon and a window of its own showing the terminal interface (the same screens as the terminal app, `mvd-tui`, described after this section), on the same engine. It runs until you quit it from its tray icon (or press Ctrl+C in the console), serves the window's page and its connection on `http://127.0.0.1:8421`, and opens the window at start. Paste links into it at any time, including while it is downloading (`a`).

```powershell
npx nx run mvd:build          # builds the React page, embeds it, builds the binary
./dist/apps/mvd/mvd           # flags: -addr 127.0.0.1:8421  -no-browser  -no-tray
npx nx run mvd:dev            # development: Vite on :4200 proxying /term to the Go app
```

**Tools and updates.** The app keeps its own copy of yt-dlp and ffmpeg in a `bin` folder inside the app-data folder (`%AppData%\mvd\bin` on Windows, `~/Library/Application Support/mvd/bin` on macOS, `~/.config/mvd/bin` on Linux), whatever is on your `PATH`, so it can keep them current without touching anything you installed yourself. A folder you can always write to, so no administrator rights are involved, and the same place on every start. deno is added only if you have neither deno nor node.

The first run downloads them from their official GitHub releases (on Windows about 310 MB: yt-dlp 17 MB, ffmpeg 200 MB, deno 93 MB), checks each against the SHA-256 that GitHub publishes for it, and says what it is doing in a notification. After that it looks for newer versions in the background each time it starts, which never holds up the page or a download, and again when a download fails (at most every ten minutes), because an out-of-date yt-dlp is the usual reason a video that worked yesterday does not today. If it replaced something after a failure it queues the failed downloads again and tells you. A new yt-dlp replaces the old one as soon as it is published. ffmpeg is replaced only by a build at least 30 days newer, because its builds are republished daily and are 200 MB. The old program is moved aside rather than deleted, so an update works even while a download is using it, and a check or download that fails changes nothing.

Limits: ffmpeg is kept current on Windows only (the project that publishes the builds has no macOS build and ships Linux ones in a format this does not unpack yet, so those fetch a pinned 4.4.1 once). deno is fetched once and never updated. The terminal app keeps using `./bin` and the tools on your `PATH`, and does not update them.

**Where it lives.** The first time a release of the tray app is started from somewhere else (your Downloads folder, say), it asks once whether to move itself to the place your system keeps programs. Yes copies it there, adds it to the menu, starts the copy and removes the old file. No leaves it where it is, and it never asks again. It does not ask when started with `-no-tray` (a terminal or script), from a development build, or from the installed place.

- **Windows:** `%LocalAppData%\Programs\MVD` with a Start menu shortcut, which needs no administrator rights. If the app was started as an administrator it offers a choice: *For everyone* (`C:\Program Files\MVD`, with a shortcut for every account) or *Just for me*. The app never asks Windows for administrator rights itself: an unsigned program that can start an administrator step is removed by Windows' antivirus (`Trojan:Win32/Bearfoos.A!ml`), so to install for everyone, right-click the app and choose Run as administrator. A signed build could offer that with one click.
- **macOS:** the release has an `mvd_<version>_macos_universal.dmg`. Open it and drag **MVD** onto the **Applications** link in the same window. The app is not notarized by Apple, so the first launch needs a right-click on MVD and **Open**; on newer macOS, if it still refuses, open System Settings, Privacy & Security and choose **Open Anyway**. An MVD started from anywhere else offers to move itself to `/Applications/MVD.app` (everyone) or `~/Applications/MVD.app` (just you).
- **Linux:** `~/.local/bin/mvd` with an entry in the applications menu.

The macOS and Linux paths, and the `.dmg`, are covered by unit tests and CI only; they have not been run on a real machine.

**Uninstalling.** Settings has a *Remove MVD* button, and `mvd -uninstall` does the same from a terminal. On Windows the app is also listed in Settings > Apps, whose Uninstall button runs `-uninstall`; if MVD is running, that one is asked to remove itself so its tray icon goes too. Whichever way it starts, MVD first shows a window on your computer listing exactly what it will remove, and nothing is removed unless you say yes there. You choose whether your preferences (settings, list and the tools MVD downloaded) go too. Your downloaded videos and their folder are never touched.

- **Windows:** the program, its install folder when that is MVD's own (`%LocalAppData%\Programs\MVD` or `C:\Program Files\MVD`; a copy running from Downloads loses only the program), the Start menu shortcuts and the Settings > Apps entry. The running program cannot delete itself, so it is moved into the temporary folder, which Windows empties. Removing the system-wide copy needs MVD to be started as an administrator; it never asks for those rights itself.
- **macOS:** the `MVD.app` it runs from, or the one in `/Applications` or `~/Applications`. If that is refused, MVD says so and you drag it to the Trash.
- **Linux:** `~/.local/bin/mvd` and the menu entry.

Starting it a second time opens the running one instead. The server only answers to `localhost`: a request is refused unless its Host is a loopback name, any Origin is a loopback page, and anything that changes state is `application/json`, so a web page on another site cannot read your queue or add to it.

**The terminal window.** The tray icon, and the start-up, open the terminal interface in a window of its own: the same screens as `apps/mvd-tui` (download list, preferences, downloads), served by [TReactUI](https://github.com/TReactUI/TReactUI) (`apps/mvd/terminalui`, at `/term`, with the page from `apps/mvd-web`). The window is a Chromium based browser's app mode (Edge or Chrome, found in the usual places; any other browser gets a normal tab). It has the mouse (click a key hint or a row, wheel to scroll) and paste, it is one program shared by every window, and a download run keeps going while no window is open. It reads and writes the same `config.conf` and `list.txt` as the terminal app. Keys: `a` adds links to a running download, and on the preferences `u` uninstalls and the folder settings open the system's own folder chooser. 

The page is `apps/mvd-web` (React, just `<TTY>` from [`@treactui/tty`](https://www.npmjs.com/package/@treactui/tty) filling the window); the terminal interface is `apps/mvd/terminalui`, which serves the app model of `libs/mvd-core/tui` at `/term` with [`tty-go`](https://github.com/TReactUI/TReactUI/tree/main/packages/tty-go). Besides `/term` and the page, the app answers `GET /api/ping` (so a second start finds the first) and `POST /api/uninstall` (so `mvd -uninstall` ends the running app); every route under `/api/` goes through a guard that refuses other sites. The tray icon opens the window when clicked and has **Open MVD** and **Quit**; where there is no system tray (a server, a bare window manager) it says so and runs until Ctrl+C, and `-no-tray` does that on purpose. The icons are drawn for each system (see below). On macOS the tray needs a C toolchain to build (it is native Cocoa), which is why the app is built on a runner of each OS; Windows and Linux build without one. The Windows release is a windowed program, so no console window opens next to the icon; a start-up error is shown in a message box instead, and when it is started from a terminal it also prints there. That comes from `-H=windowsgui` in the `build-native` target of `apps/mvd/project.json`, which is edited by hand (mnci cannot pass the flag yet), and from `libs/mvd-core/procwindow`, which keeps yt-dlp and the folder chooser from opening console windows of their own. Plain `nx run mvd:build` and `go run` stay console builds, so you see the logs.

## 🚀 Quick Start (terminal app)

Build and run; on first launch it creates a default config in your preferences folder and opens the **preferences** screen:

```powershell
go build -o mvd-tui.exe ./apps/mvd-tui
.\mvd-tui.exe
```

Then: paste your playlist/video URLs on the **list** screen (one per line), press `Ctrl+S` to start, and the **download** dashboard takes over. Everything is kept in the app-data folder (see below) — there are no config files next to the binary.

For an unattended/headless run (pipes, CI, cron) it downloads the saved list with a plain log instead of the screens:

```powershell
.\mvd-tui.exe --no-tui
```

---

## 🖥️ The terminal app interface (`mvd-tui`)

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

The repository is an [mnci](https://github.com/russoedu/MoNecromanCi) (Nx) workspace with one Go module at the root: `apps/` holds the programs (`apps/mvd-tui`, the terminal app; `apps/mvd`, the tray app, with its window's page in `apps/mvd-web`) and `libs/` the code they share (`libs/mvd-core`, the engine and the screens), so each front end reuses the engine instead of copying it. `npx nx run-many -t test,build` builds and tests all of it.

The code follows vertical feature slices: `apps/mvd-tui/main.go` only wires things together, and every folder under `libs/mvd-core/` is one slice that owns one outcome. A slice is flat, and each file is named `<name>_<role>.go` so the role says what the file does (`use_case` coordinates an operation, `policy` is a reusable decision, `algorithm` is pure computation, `mapper` converts representations, `contract` is data crossing a boundary, `client` talks to an external service, `store` holds runtime state, `repository` persists, `handler` adapts a transport such as the keyboard, `config` and `enum` are what they say).

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

Dependencies point one way: `main` → `engine` → `ytdlp`; `main` → `official`, injected into the engine through a small port interface so the engine never imports it; `tui` and `plain` → `runstate` → `engine`. No two slices import each other. Tests sit next to the file they test; the `ytdlp` and `engine` test binaries double as a stub `yt-dlp`, so the suite runs on every platform without shell scripts.

## 🛠️ Building & Releasing

```powershell
# Build for your OS
go build -o mvd-tui.exe ./apps/mvd-tui

# Run the tests (Nx runs each project from its own folder; a bare `go test ./...`
# from the root would also walk node_modules once `npm install` has run)
npx nx run-many -t test

# Plain log output (no full screen interface)
.\mvd-tui.exe --no-tui
```

GitHub Releases are automatically created via GitHub Actions on every new tag push (e.g., `v1.0.0`).

Both apps release from `ci.yml` (mnci): a push to `main` versions each from conventional commits, tags it (`mvd@x.y.z`, `mvd-tui@x.y.z`) and attaches its files to that GitHub Release. The files are named `<product>_<version>_<os>_<processor>.<type>` (`mvd_0.0.9_windows_amd64.zip`, `mvd_0.0.9_macos_universal.dmg`, `mvd-tui_0.1.0_linux_arm64.zip`), where the product is `mvd` for the desktop app (the project `mvd`) and `mvd-tui` for the terminal app, and macOS is always written `macos`. `tools/release-assets.cjs` is the one place that decides the names, and CI builds every file with a fixed version and fails if one breaks the rule. The desktop app is a `--cgo` app, so its zip is built on a runner of each OS (the macOS image holds one universal program); the terminal app is cross-compiled for the six platforms. Nx therefore leaves `mvd` out of the cross-compiling verify job and builds it in the `native` job instead. Releases up to 0.0.10 were published as `mvd-tray@x.y.z`; from the next one they are `mvd@x.y.z`. The file names did not change.

An older workflow, `release.yml`, still builds the terminal app on a `v*` tag under the name `youtube-downloader-<os>-<arch>`; it predates the mnci release and does not follow the naming above (it tests only `mvd-tui` and `libs`, because `mvd` cannot compile until its React page has been built into `apps/mvd/localserver/web`).
