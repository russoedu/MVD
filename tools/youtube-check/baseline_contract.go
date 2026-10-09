package main

// Song is a song of the playlist and the answer recorded for it.
type Song struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	// Official is the video id the lookup gave when the baseline was recorded, or empty
	// when it found none (the song has no official video, or it is one itself).
	Official string `json:"official"`
}

// Baseline is what the lookup answered for the playlist when someone last looked at the
// answers and found them right.
type Baseline struct {
	Playlist string `json:"playlist"`
	Songs    []Song `json:"songs"`
}

// Outcome is the answer of the lookup now for one song of the playlist.
type Outcome struct {
	ID       string
	Title    string
	Found    string
	Why      string
	Expected string
	// Known says the song is in the baseline, so Expected means something.
	Known bool
}
