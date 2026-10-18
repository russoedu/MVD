// Package playlistfile reads and writes .mvd files: a download list, what is to be downloaded and
// how far each song has got, saved so it can be finished later, in small chunks.
//
// A .mvd file is gzip-compressed JSON lines. The first line is a header
// ({"mvd":2,"created":...}); each of the others is one item of the list (an address) with, once it has been
// looked at, its songs: what was proposed for each, what the person decided and how far
// it has got.
package playlistfile
