// Package downloadlist keeps the download list the person has open: a .mvd file (see
// playlistfile) that is written as the downloads go, so a list left half way is taken up
// another day.
//
// A Session is the open list. Inputs says what a run is to be given to carry on with it:
// the addresses nobody has looked at and the songs that are not downloaded yet.
// Watch writes the list from an engine while it runs.
package downloadlist
