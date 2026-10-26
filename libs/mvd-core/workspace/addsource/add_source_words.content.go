package addsource

// Heading is the title of the window.
const Heading = "Add music"

// URLHelp is what the window says under the field for a link: what kinds are taken.
var URLHelp = []string{
	"Paste one or more links, separated by spaces or on separate lines. These work:",
	"",
	"  • YouTube playlist      youtube.com/playlist?list=…",
	"  • YouTube video         youtube.com/watch?v=…  or  youtu.be/…",
	"  • Spotify playlist      open.spotify.com/playlist/…",
	"  • Apple Music playlist  music.apple.com/…/playlist/…",
	"  • A list of songs on your computer: a .txt file with one “Artist - Title” per line, or a .csv or .tsv",
}

// SearchHelp is what the window says for a search by name.
var SearchHelp = []string{
	"Write what you want and MVD looks for the songs and their videos:",
	"",
	"  • An artist only: every song of the artist it can find",
	"  • Artist and album: the songs of that album",
	"  • Artist and song, or artist, album and song: that one song",
	"",
	"An album or a song always needs its artist.",
}

// Help is what the window says for the mode.
func Help(mode Mode) []string {
	if mode == ModeSearch {
		return SearchHelp
	}

	return URLHelp
}

// Buttons are the two buttons of the window.
var Buttons = []string{"Add", "Cancel"}
