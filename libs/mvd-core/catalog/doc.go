// Package catalog is renderer-neutral: it knows what music exists. Given the music a person asked for by
// name (see package musicsearch) it names the songs: every song of an artist, the songs of an album, or
// the one song, with the artist, album, title and length as a music catalogue spells them. It asks the
// public catalogues that need no key (iTunes, Deezer), one after the other, and keeps the first that
// knows. Finding the video of each song is not its business (see package trackfind).
package catalog
