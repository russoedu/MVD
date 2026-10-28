package download

import (
	"youtube-downloader/libs/mvd-core/run/downloadsnapshot"
)

// marqueeSubject is the text of the selected link and of the selected song joined, which is
// what the marquee scrolls: when it changes, the marquee starts again from its beginning.
func (m Model) marqueeSubject() string {
	subject := ""
	if pls := m.session.State.Playlists; m.session.SelPlaylist >= 0 && m.session.SelPlaylist < len(pls) {
		subject = downloadsnapshot.PlaylistName(m.session, pls[m.session.SelPlaylist])
	}
	if en := m.session.SelectedEntry(); en != nil {
		subject += "|" + songName(en)
	}

	return subject
}
