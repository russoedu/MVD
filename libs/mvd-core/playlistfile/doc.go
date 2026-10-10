// Package playlistfile reads and writes .mvd files: a playlist that has been planned and
// reviewed, saved so it can be finished later, in small chunks.
//
// A .mvd file is gzip-compressed JSON lines. The first line is a header
// ({"mvd":1,"created":...}); each of the others is one song with what was proposed for
// it, what the person decided and whether it has been downloaded.
package playlistfile
