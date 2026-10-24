package musicsource

import "testing"

func page(cards string) []byte {
	return []byte(`<html><script>var ytInitialData = {"contents":{"panel":{"horizontalCardListRenderer":{"cards":[` + cards + `]}}}};</script></html>`)
}

const oneCard = `{"videoAttributeViewModel":{"title":"I Swear (Extended Club Mix)","subtitle":"Bill Power","secondarySubtitle":{"content":"I Swear"}}}`

func TestTheMusicCardNamesTheSongOfAPlainVideo(t *testing.T) {
	artist, title, ok := SongFromWatchPage(page(oneCard))
	if !ok || artist != "Bill Power" || title != "I Swear (Extended Club Mix)" {
		t.Errorf("got %q, %q, %v", artist, title, ok)
	}
}

func TestTheSameCardWrittenTwiceIsOneSong(t *testing.T) {
	if _, _, ok := SongFromWatchPage(page(oneCard + "," + oneCard)); !ok {
		t.Error("a card repeated in the page is still one song")
	}
}

func TestSeveralSongsOrNoCardNameNothing(t *testing.T) {
	two := oneCard + `,{"videoAttributeViewModel":{"title":"Another","subtitle":"Someone"}}`
	if _, _, ok := SongFromWatchPage(page(two)); ok {
		t.Error("a compilation's card cannot say which song the upload is")
	}
	if _, _, ok := SongFromWatchPage(page("")); ok {
		t.Error("no card, no song")
	}
	if _, _, ok := SongFromWatchPage([]byte("<html>no data</html>")); ok {
		t.Error("a page without data names nothing")
	}
	if _, _, ok := SongFromWatchPage(page(`{"videoAttributeViewModel":{"title":"Only a title"}}`)); ok {
		t.Error("a card without an artist names nothing")
	}
}

func TestTheMusicCardAlsoNamesTheAlbum(t *testing.T) {
	artist, title, album, ok := SongCardFromWatchPage(page(oneCard))
	if !ok || artist != "Bill Power" || title != "I Swear (Extended Club Mix)" || album != "I Swear" {
		t.Errorf("got %q, %q, %q, %v", artist, title, album, ok)
	}
	noAlbum := `{"videoAttributeViewModel":{"title":"Song","subtitle":"Band"}}`
	if _, _, album, ok := SongCardFromWatchPage(page(noAlbum)); !ok || album != "" {
		t.Errorf("a card with no album gave %q, %v", album, ok)
	}
}
