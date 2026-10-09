package songid

import "youtube-downloader/libs/mvd-core/official"

// Identifier names the song of an upload with the help of two databases. Either may be
// nil; a database that fails is skipped, since the answer is only a help.
type Identifier struct {
	MusicBrainz *MusicBrainzClient
	ITunes      *ITunesClient
}

// NewIdentifier returns an identifier that asks the real MusicBrainz and iTunes.
func NewIdentifier() *Identifier {
	return &Identifier{MusicBrainz: NewMusicBrainzClient(), ITunes: NewITunesClient()}
}

// Identify returns the artist and the title a database gives the song of an upload,
// or false when none does with the upload's own words behind it. The title is
// returned without the mix and credit terms a search leaves out.
func (i *Identifier) Identify(uploadTitle, channel string) (artist, title string, ok bool) {
	for _, guess := range guessesFrom(uploadTitle, channel) {
		if i.MusicBrainz != nil {
			if found, err := i.MusicBrainz.Recordings(guess.Artist, guess.Title); err == nil {
				if identity, hit := firstThatNamesTheUpload(found, uploadTitle, channel); hit {
					return identity.Artist, official.SearchTitle(identity.Title), true
				}
			}
		}
		if i.ITunes != nil {
			if found, err := i.ITunes.Songs(guess.Artist + " " + guess.Title); err == nil {
				if identity, hit := firstThatNamesTheUpload(found, uploadTitle, channel); hit {
					return identity.Artist, official.SearchTitle(identity.Title), true
				}
			}
		}
	}
	return "", "", false
}

func firstThatNamesTheUpload(found []Identity, uploadTitle, channel string) (Identity, bool) {
	for _, identity := range found {
		if namesTheUpload(identity, uploadTitle, channel) {
			return identity, true
		}
	}
	return Identity{}, false
}
