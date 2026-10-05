<p align="center">
  <img src="assets/mvd.gif" alt="MVD - Music Video Downloader" width="480">
</p>

# 📺 MVD — Music Video Downloader

> **"I want my MTV... and I want it stored locally, total overload style!"**

Remember the 90s? Back when music videos were an absolute cultural obsession, kids were brain-melted in front of MTV all night, and downloading a single crunchy 240p video clip on dial-up meant hostage-negotiating the family phone line for 48 straight hours?

**MVD (Music Video Downloader)** is built for that exact flavor of raw visual nostalgia—minus the ear-bleeding 56k modem noise. Written in Go and powered by `yt-dlp`, it’s designed to snatch back the golden, unfiltered era of music television with zero setup friction.

---

### 💡 The Concept

In today's soft world of background audio streams, sanitized algorithm feeds, and static cover art, the raw art of the *actual music video* got kicked to the curb. We built **MVD** to resurrect visual music with a vengeance.

Feed MVD a stack of YouTube video or playlist links, and instead of leaving you with boring, sanitized "Topic" uploads (those lifeless still-image audio tracks), it aggressively crawls the video metadata to hijack the **official, original music video**!

Whether you’re setting up an offline media wall, fueling a chaotic 90s house party, or hoarding visual culture before the copyright bots strike it into oblivion, MVD rips through your queue concurrently.

---

### 🎨 The Aesthetic: Unapologetic 90s Chaos

We didn't just write a tool; we embraced the total, gloriously ugly anarchy of early web aesthetics and terminal grit.

* **Iconically Ugly Branding:** Stare directly at that logo. Feast your eyes on toxic neon green and Barney-purple fonts stretched out in mismatched, unholy serif/sans-serif proportions. It looks like it was hacked together in Microsoft Paint on Windows 95 while blasting *Smells Like Teen Spirit* on loop—and that’s **exactly** why it rules.
* **Tray App Power (`mvd`):** Lurks silently in your system tray and serves a raw local web UI page in your browser, keeping your background downloads cranking without cluttering your desktop space.
* **TUI Madness (`mvd-tui`):** Modern Web3 rounded buttons and pastel design systems? Absolute trash. MVD serves up pure, hard-edged ASCII terminal UI (TUI) box-drawing energy. Run it straight in your terminal or render the full interactive TUI right inside a browser viewport via TReactUI.

---

### ⚡ Heavyweight Feature Arsenal

* 🎬 **Official Video Hijack:** Ruthlessly swaps out auto-generated `Artist - Topic` still-image uploads for the official, authentic music video linked inside YouTube's *Music* cards.
* 🛠 **Auto-Dependency Self-Installation:** Zero setup bullshit. On startup, MVD hunts down missing binaries for `yt-dlp`, `ffmpeg`, and JS engines (`deno`), verifying SHA-256 checksums and dropping them directly into your isolated app data folder—no admin rights or manual PATH hacking needed.
* 🚀 **Parallel Goroutine Power:** Multi-threaded download workers blast through massive playlists concurrently with live, per-track progress bars and raw log output.
* 🍪 **Automatic Cookie Hijacking:** Automatically detects signed-in YouTube cookies from installed local browsers (Firefox, Chrome, Edge) to dodge `403 Forbidden` errors, `429 Rate Limits`, and bot check walls without asking you to configure a thing.
* 🖥 **Tray Stealth + Browser TUI:** Runs as a native desktop tray utility driving a local browser session, or as a standalone, hardcore `mvd-tui` terminal executable.
* ⟲ **Relentless Auto-Retry Engine:** Instantly powers through temporary network glitches and sweeps back around for rate-limited downloads once cooldowns reset—ignoring dead, private, or geo-blocked tracks like a champ.
* 💻 **Cross-Platform & Self-Relocating:** Built for Windows, macOS (Universal/Apple Silicon/Intel), and Linux. Offers to cleanly drop itself into your OS application directory on first launch, complete with a built-in self-destruct uninstaller.

---

### 🕶️ Why?

Because modern music apps treat music videos like second-class garbage. MVD treats them like the sacred, radical 90s art form they are. Fire up your playlists, launch the app, and crank the visual volume!

### 🖹 The normal "README.md" (not this one)

[View the basic "README"](./NORMAL-README.md)
