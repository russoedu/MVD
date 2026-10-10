// Package audiofix finds downloads whose sound plays on one side only: some uploads carry
// a stereo track with one silent channel (a mono recording put on the right channel),
// which the file's metadata does not show. The file is never changed; the caller looks
// for another version of the song.
package audiofix
