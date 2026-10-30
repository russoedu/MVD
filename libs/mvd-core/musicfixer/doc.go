// Package musicfixer calls the music-fixer app (made with VideoCraw): it puts a high-quality
// copy of a song on a video that carries the same song in low quality, keeping the pictures.
// The app is started as a child process and spoken to by its command line, JSON events on
// standard output and "cancel" on standard input (VideoCraw's docs/exported-app-contract.md).
package musicfixer
