// Package songid says who sings a song and what it is called, by asking music
// databases (iTunes, Deezer, then MusicBrainz), when all there is to go by is the title of
// an upload and the channel that posted it. Playlist titles are untidy: the artist
// may be in front of the title, behind it, or in the brackets ("Tonight Is The Night
// (Le Click - Dance Mix)"), and a database's answer is taken only when the upload
// itself mentions both the artist and the song.
package songid
