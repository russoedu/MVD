package songid

import "youtube-downloader/libs/mvd-core/official"

// minTitleFit is how much of the database's title the upload's must carry.
const minTitleFit = 0.6

// namesTheUpload says whether a database's answer is the song of an upload: the
// upload (its title or its channel) mentions the artist, and its title says the
// song. Anything less is a database finding another song with a similar name.
func namesTheUpload(identity Identity, uploadTitle, channel string) bool {
	if !official.Mentions(uploadTitle+" "+channel, identity.Artist) {
		return false
	}
	return official.TitleSimilarity(identity.Title, uploadTitle, []string{identity.Artist}) >= minTitleFit
}
