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
		{"download-entries", 120, 36, download(func(m *Model) { m.session.Pane = paneEntries; m.session.SelEntry = 4 })},
		{"download-lists", 120, 36, download(func(m *Model) { m.session.Pane = paneLists })},
		{"download-emoji-playlist", 120, 36, download(func(m *Model) { m.session.Pane = paneLists; m.session.SelPlaylist = 1 })},
		{"download-failed-playlist", 120, 36, download(func(m *Model) { m.session.Pane = paneLists; m.session.SelPlaylist = 2 })},
		{"download-full-log", 120, 36, download(func(m *Model) { m.session.FullLog = true; m.session.Pane = paneOutput; m.session.OutScroll = 5 })},
		{"download-filter-failed", 120, 36, download(func(m *Model) { m.session.Filter = onlyFailed(); m.session.Pane = paneEntries })},
		{"download-confirm-quit", 120, 36, download(func(m *Model) { m.session.ConfirmQuit = true })},
		{"download-help", 120, 36, download(func(m *Model) { m.session.ShowHelp = true })},
		{"download-add-links", 120, 36, download(func(m *Model) {
			ta := textarea.New()
			ta.Placeholder = "Paste or type playlist / video URLs, one per line"
			ta.ShowLineNumbers = false
			ta.SetWidth(m.addBoxWidth())
			ta.SetHeight(8)
			ta.SetValue("https://youtube.com/playlist?list=A\nhttps://youtube.com/watch?v=B")
			ta.Focus()
			m.session.Adding, m.links = true, ta
		})},
		{"download-add-links-narrow", 70, 20, download(func(m *Model) {
			m.width, m.height = 70, 20
			ta := textarea.New()
			ta.Placeholder = "Paste or type playlist / video URLs, one per line"
			ta.ShowLineNumbers = false
			ta.SetWidth(m.addBoxWidth())
			ta.SetHeight(8)
			ta.Focus()
			m.session.Adding, m.links = true, ta
		})},
		{"download-help-narrow", 70, 20, download(func(m *Model) { m.width, m.height = 70, 20; m.session.ShowHelp = true })},
		{"download-narrow", 80, 24, download(func(m *Model) { m.width, m.height = 80, 24; m.session.Pane = paneEntries; m.session.SelEntry = 2 })},
	}
	for _, c := range cases {
		drawing.Check(t, "testdata", c.name, c.w, c.draw(t))
	}
}
