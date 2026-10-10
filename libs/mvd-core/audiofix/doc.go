// Package audiofix repairs a download whose sound comes out of one side only: some
// uploads carry a stereo track with one silent channel (a mono recording put on the
// right channel), which the file's metadata does not show. The audible channel is copied
// to both sides, so the song plays in the middle.
package audiofix
