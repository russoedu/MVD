package download

import (
	"testing"

	"charm.land/bubbles/v2/textarea"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/testing/drawing"
)

// TestTheScreenKeepsItsLook compares the screen in the states the screen test walks
// through with the reference drawings of the hand written screen it replaced.
func TestTheScreenKeepsItsLook(t *testing.T) {
	// The download screen, in the states the screens test walks through.
	download := func(setup func(m *Model)) func(t *testing.T) string {
		return func(t *testing.T) string {
			look.Apply(config.DefaultThemeColors())
			m := sampleModel(t)
			m.width, m.height = 120, 36
			setup(&m)
			return m.View().Content
		}
	}
	cases := []struct {
		name string
		w, h int
		draw func(t *testing.T) string
	}{
		{"download-entries", 120, 36, download(func(m *Model) { m.focus = paneEntries; m.selEntry = 4 })},
		{"download-lists", 120, 36, download(func(m *Model) { m.focus = paneLists })},
		{"download-emoji-playlist", 120, 36, download(func(m *Model) { m.focus = paneLists; m.selPlaylist = 1 })},
		{"download-failed-playlist", 120, 36, download(func(m *Model) { m.focus = paneLists; m.selPlaylist = 2 })},
		{"download-full-log", 120, 36, download(func(m *Model) { m.fullLog = true; m.focus = paneOutput; m.outScroll = 5 })},
		{"download-filter-failed", 120, 36, download(func(m *Model) { m.statusHidden = onlyFailed(); m.focus = paneEntries })},
		{"download-confirm-quit", 120, 36, download(func(m *Model) { m.confirmQuit = true })},
		{"download-help", 120, 36, download(func(m *Model) { m.showHelp = true })},
		{"download-add-links", 120, 36, download(func(m *Model) {
			ta := textarea.New()
			ta.Placeholder = "Paste or type playlist / video URLs, one per line"
			ta.ShowLineNumbers = false
			ta.SetWidth(m.addBoxWidth())
			ta.SetHeight(8)
			ta.SetValue("https://youtube.com/playlist?list=A\nhttps://youtube.com/watch?v=B")
			ta.Focus()
			m.adding, m.links = true, ta
		})},
		{"download-add-links-narrow", 70, 20, download(func(m *Model) {
			m.width, m.height = 70, 20
			ta := textarea.New()
			ta.Placeholder = "Paste or type playlist / video URLs, one per line"
			ta.ShowLineNumbers = false
			ta.SetWidth(m.addBoxWidth())
			ta.SetHeight(8)
			ta.Focus()
			m.adding, m.links = true, ta
		})},
		{"download-help-narrow", 70, 20, download(func(m *Model) { m.width, m.height = 70, 20; m.showHelp = true })},
		{"download-narrow", 80, 24, download(func(m *Model) { m.width, m.height = 80, 24; m.focus = paneEntries; m.selEntry = 2 })},
	}
	for _, c := range cases {
		drawing.Check(t, "testdata", c.name, c.w, c.draw(t))
	}
}
