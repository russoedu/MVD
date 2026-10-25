// Package downloadsnapshot is renderer-neutral: it says what the download screen shows, in plain
// data. The counters of the header and what they count, the sentences of the details, the keys on
// offer, the rows of the playlists and the songs, the progress under the log. Each is made from a
// downloadsession.Session and carries words and tones, never colours, glyph characters or cells; a
// renderer decides how a tone, a kind of glyph or a bar looks.
package downloadsnapshot
