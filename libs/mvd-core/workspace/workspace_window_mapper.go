package workspace

import (
	"path/filepath"
	"strconv"
	"strings"

	"youtube-downloader/libs/mvd-core/dialog"
	"youtube-downloader/libs/mvd-core/run/downloadsession"
	"youtube-downloader/libs/mvd-core/run/downloadsnapshot"
	"youtube-downloader/libs/mvd-core/songlist"
	"youtube-downloader/libs/mvd-core/wording"
	"youtube-downloader/libs/mvd-core/workspace/addsource"
	"youtube-downloader/libs/mvd-core/workspace/windowsnapshot"
)

// Window is what the workspace shows, for a classic themed window to draw: the same menus,
// buttons, lists and words the terminal look draws, as roles and tones.
func (m Model) Window() windowsnapshot.Snapshot {
	s := m.dl
	snap := windowsnapshot.Snapshot{
		Title:    m.windowTitle(),
		Menus:    m.windowMenus(),
		Toolbar:  m.windowButtons(),
		Status:   m.windowStatus(),
		Activity: m.windowActivity(),
		Question: m.windowQuestion(),
		Covered:  m.windowCovered(),
	}
	if s.State == nil {
		return snap
	}
	snap.Sources = windowSources(s)
	snap.Songs = windowSongs(s)
	snap.Details = windowDetails(s)
	if m.addOpen {
		snap.Adding = m.windowAdding()
	}

	return snap
}

// windowTitle is the name of the window: the app and the list that is open.
func (m Model) windowTitle() string {
	name := "New list"
	if path := m.session.Path(); path != "" {
		name = filepath.Base(path)
	}

	return "MVD - Music Video Downloader - " + name
}

// windowMenus are the menus of the bar, with what can be chosen now and each entry's keys.
func (m Model) windowMenus() []windowsnapshot.Menu {
	menus := m.bar.Menus()
	headings := make([]string, len(menus))
	for i, menu := range menus {
		headings[i] = menu.Name
	}
	letters := windowsnapshot.Mnemonics(headings)
	out := make([]windowsnapshot.Menu, 0, len(menus))
	for i, menu := range menus {
		labels := make([]string, len(menu.Entries))
		for j, entry := range menu.Entries {
			labels[j] = entry.Name
		}
		entryLetters := windowsnapshot.Mnemonics(labels)
		entries := []windowsnapshot.MenuEntry{}
		separator := false
		for j, entry := range menu.Entries {
			if entry.Separator {
				separator = len(entries) > 0
				continue
			}
			entries = append(entries, windowsnapshot.MenuEntry{
				ID: entry.Name, Label: entry.Name, Mnemonic: entryLetters[j],
				Shortcut: windowsnapshot.ShortcutOf(entry.Shortcut), Enabled: m.bar.Enabled(entry.Name),
				SeparatorBefore: separator,
			})
			separator = false
		}
		out = append(out, windowsnapshot.Menu{Label: menu.Name, Mnemonic: letters[i], Entries: entries})
	}

	return out
}

// windowButtons are the buttons of the toolbar: the first that can be used is the default one,
// and one that cannot says why. While songs wait for the person to choose them, the window leads
// with the button the details name, since it has no column of buttons beside the playlists.
func (m Model) windowButtons() []windowsnapshot.Button {
	buttons := m.buttons()
	out := make([]windowsnapshot.Button, 0, len(buttons)+1)
	accept := m.canAcceptAll()
	if accept {
		out = append(out, windowsnapshot.Button{ID: entryDownloadAll, Label: acceptAllLabel, Role: windowsnapshot.RoleDefault, Enabled: true})
	}
	for i, b := range buttons {
		role := windowsnapshot.RoleOther
		if !accept && buttonKind(buttons, i) == dialog.PillPrimary {
			role = windowsnapshot.RoleDefault
		}
		one := windowsnapshot.Button{ID: b.entry, Label: b.label, Role: role, Enabled: !b.disabled}
		if b.disabled {
			one.DisabledReason = m.whyDisabled()
		}
		out = append(out, one)
	}

	return out
}

// whyDisabled is why a button of the toolbar cannot be used now.
func (m Model) whyDisabled() string {
	if m.life.Busy() {
		return "One moment..."
	}
	if state, waiting := m.toolsWaiting(); waiting {
		return state.Message
	}

	return "Add a source first."
}

// windowStatus is the line at the bottom: what was just said, or how the run is going. The
// terminal look's key hints (move, pane, follow) are its own: a classic window has its menus and
// their shortcuts instead.
func (m Model) windowStatus() windowsnapshot.Status {
	return windowsnapshot.Status{Summary: m.statusSummary(), Hints: []windowsnapshot.KeyHint{}}
}

// statusSummary is the notice, the warning or the news of the workspace, else the counts of the run.
func (m Model) statusSummary() string {
	for _, said := range []string{m.notice, m.caution, m.done} {
		if said != "" {
			return said
		}
	}
	s := m.dl
	if s.State == nil {
		return ""
	}
	header := downloadsnapshot.HeaderOf(s)
	var parts []string
	for _, c := range header.Counters {
		if c.Count > 0 {
			parts = append(parts, c.Label+" "+strconv.Itoa(c.Count))
		}
	}
	if header.Right != "" {
		parts = append(parts, header.Right)
	}

	return strings.Join(parts, " · ")
}

// windowActivity is what MVD is doing, for the theme's pointer: starting or stopping keeps the
// person waiting; reading the playlists goes on behind.
func (m Model) windowActivity() windowsnapshot.Activity {
	f := m.facts()
	switch {
	case m.life.Busy():
		return windowsnapshot.ActivityWaiting
	case m.running() && !f.AllListed:
		return windowsnapshot.ActivityBackground
	}

	return windowsnapshot.ActivityIdle
}

// windowCovered is the screen open over the window that only the terminal look draws so far.
func (m Model) windowCovered() *windowsnapshot.Covered {
	// The questions and notes are dialogs of the theme (see windowQuestion); only the address being
	// written in its row is the terminal's still.
	if !m.inline.on || m.windowQuestion() != nil {
		return nil
	}

	return &windowsnapshot.Covered{Title: "Change the address"}
}

// windowSources are the sources of the list.
func windowSources(s downloadsession.Session) windowsnapshot.Sources {
	rows := downloadsnapshot.PlaylistRowsOf(s)
	out := windowsnapshot.Sources{Heading: wording.Counted(len(rows), "source"), Rows: []windowsnapshot.SourceRow{}, Selected: s.SelPlaylist}
	for i, row := range rows {
		out.Rows = append(out.Rows, windowsnapshot.SourceRow{ID: strconv.Itoa(i), Label: row.Name, Detail: row.Count, Tone: windowTone(row.Glyph.Tone)})
	}

	return out
}

// windowSongs are the songs shown, in the order shown.
func windowSongs(s downloadsession.Session) windowsnapshot.Songs {
	ids := s.VisibleEntries()
	out := windowsnapshot.Songs{Heading: wording.Counted(len(ids), "song"), Rows: []windowsnapshot.SongRow{}, Selected: s.SelEntry}
	for _, id := range ids {
		en := s.State.Entry(id)
		if en == nil {
			continue
		}
		row := downloadsnapshot.SongRowOf(s, en)
		out.Rows = append(out.Rows, windowsnapshot.SongRow{
			ID: strconv.Itoa(row.ID), Ticked: row.Ticked, Artist: row.Artist, Title: row.Name,
			Progress: windowProgress(row), Status: row.Status, Tone: windowTone(row.Tone),
		})
	}

	return out
}

// windowProgress is how far a song is: measured while it comes in, unknown while it is being
// found or put together, none before it begins.
func windowProgress(row downloadsnapshot.SongRow) windowsnapshot.Progress {
	switch {
	case row.Progress.Visible:
		return windowsnapshot.Part(row.Progress.Percent / 100)
	case row.Glyph.Kind == downloadsnapshot.GlyphResolving || row.Glyph.Kind == downloadsnapshot.GlyphBusy:
		return windowsnapshot.Unknown()
	}

	return windowsnapshot.Progress{}
}

// windowDetails are the lines of the details, as plain sentences.
func windowDetails(s downloadsession.Session) windowsnapshot.Details {
	out := windowsnapshot.Details{Heading: downloadsnapshot.OutputTitle(s), Lines: []string{}}
	for _, line := range downloadsnapshot.DetailsOf(s) {
		out.Lines = append(out.Lines, line.Text)
	}

	return out
}

// windowTone is a tone of the lists as the classic themes name it.
func windowTone(tone songlist.Tone) windowsnapshot.Tone {
	switch tone {
	case songlist.ToneDim:
		return windowsnapshot.ToneDim
	case songlist.ToneSuccess:
		return windowsnapshot.ToneGood
	case songlist.ToneError:
		return windowsnapshot.ToneBad
	case songlist.ToneHighlight, songlist.ToneFocus:
		return windowsnapshot.ToneBusy
	}

	return windowsnapshot.ToneNormal
}

// addFields are the fields of the window for adding music, as the window names them.
var addFields = map[addsource.Field]string{
	addsource.FieldURL:    windowsnapshot.FieldURL,
	addsource.FieldArtist: windowsnapshot.FieldArtist,
	addsource.FieldAlbum:  windowsnapshot.FieldAlbum,
	addsource.FieldSong:   windowsnapshot.FieldSong,
}

// windowAdding is the window for adding music, as it is now.
func (m Model) windowAdding() *windowsnapshot.AddMusic {
	add := m.add
	mode := windowsnapshot.ModeURL
	if add.Mode == addsource.ModeSearch {
		mode = windowsnapshot.ModeSearch
	}
	modeLetters := windowsnapshot.Mnemonics(addsource.Modes)
	out := &windowsnapshot.AddMusic{
		Heading: addsource.Heading, Mode: mode, Problem: add.Problem,
		Modes: []windowsnapshot.AddMode{
			{Mode: windowsnapshot.ModeURL, Label: addsource.Modes[0], Mnemonic: modeLetters[0]},
			{Mode: windowsnapshot.ModeSearch, Label: addsource.Modes[1], Mnemonic: modeLetters[1]},
		},
		Help: windowHelp(addsource.Help(add.Mode)),
	}
	var labels []string
	for _, f := range add.Fields() {
		if addsource.IsText(f) {
			labels = append(labels, f.Label())
		}
	}
	// The letters of the fields and the Add button are chosen after the modes', so none repeats.
	letters := windowsnapshot.Mnemonics(append(append(append([]string{}, addsource.Modes...), labels...), addsource.Buttons[0]))[len(addsource.Modes):]
	for _, f := range add.Fields() {
		if addsource.IsText(f) {
			out.Fields = append(out.Fields, windowsnapshot.AddField{Field: addFields[f], Label: f.Label(), Mnemonic: letters[len(out.Fields)], Text: add.Text(f)})
		}
	}
	out.Buttons = []windowsnapshot.Button{
		{ID: "add", Label: addsource.Buttons[0], Mnemonic: letters[len(letters)-1], Role: windowsnapshot.RoleDefault, Enabled: true},
		{ID: "cancel", Label: addsource.Buttons[1], Role: windowsnapshot.RoleCancel, Enabled: true},
	}

	return out
}

// windowHelp is the help of the window as items: the empty lines and the lead-in go, and the
// bullets lose their mark.
func windowHelp(lines []string) []string {
	var out []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "•") {
			out = append(out, strings.Join(strings.Fields(strings.TrimPrefix(line, "•")), " "))
		} else if line != "" && !strings.HasSuffix(line, ":") {
			out = append(out, line)
		}
	}

	return out
}
