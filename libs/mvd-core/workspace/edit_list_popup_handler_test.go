package workspace

import (
	"testing"

	"youtube-downloader/libs/mvd-core/config"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/playlistfile"
)

func keys(p editPopup, names ...tea.KeyPressMsg) editPopup {
	for _, k := range names {
		p, _ = p.Update(k)
	}
	return p
}

func rune_(r rune) tea.KeyPressMsg { return tea.KeyPressMsg{Text: string(r), Code: r} }

func sources(p editPopup) []string {
	var out []string
	for _, it := range p.items {
		out = append(out, it.Source)
	}
	return out
}

func TestEditingMovesRemovesAndAddsAddresses(t *testing.T) {
	p := newEditPopup(playlistfile.File{Items: []playlistfile.Item{{Source: "a"}, {Source: "b"}, {Source: "c"}}})
	p = keys(p, tea.KeyPressMsg{Code: tea.KeyDown}, tea.KeyPressMsg{Code: tea.KeyUp, Mod: tea.ModShift})
	if got := sources(p); got[0] != "b" || got[1] != "a" || p.sel != 0 {
		t.Errorf("after move up %v sel %d", got, p.sel)
	}
	p = keys(p, rune_('d'))
	if got := sources(p); len(got) != 2 || got[0] != "a" {
		t.Errorf("after remove %v", got)
	}
	p = keys(p, rune_('a'))
	p.input.SetValue("new")
	p = keys(p, tea.KeyPressMsg{Code: tea.KeyEnter})
	if got := sources(p); len(got) != 3 || got[1] != "new" || p.sel != 1 {
		t.Errorf("after add %v sel %d", got, p.sel)
	}
}

func TestSongsMoveAmongTheSongsOfTheirOwnAddress(t *testing.T) {
	songs := []playlistfile.Entry{{Title: "one", Artist: "x"}, {Title: "two", Artist: "x"}}
	p := newEditPopup(playlistfile.File{Items: []playlistfile.Item{{Source: "a", Songs: songs}, {Source: "b"}}})
	p = keys(p, tea.KeyPressMsg{Code: tea.KeyDown}, tea.KeyPressMsg{Code: tea.KeyDown, Mod: tea.ModShift}, tea.KeyPressMsg{Code: tea.KeyDown, Mod: tea.ModShift})
	if p.items[0].Songs[1].Title != "one" || p.items[0].Songs[0].Title != "two" {
		t.Errorf("songs %+v", p.items[0].Songs)
	}
	if len(p.items) != 2 || p.items[1].Source != "b" {
		t.Errorf("items %v", sources(p))
	}
}

func TestTheEditReachesTheListOnlyWhenSaved(t *testing.T) {
	m := sized(New(Host{}, config.Config{}, nil))
	m.session.SetSources([]string{"a", "b"})
	next, _ := m.Update(tea.KeyPressMsg{Text: "e", Code: 'e'})
	m = next.(Model)
	if !m.editing {
		t.Fatal("not editing")
	}
	next, cmd := m.Update(tea.KeyPressMsg{Text: "d", Code: 'd'})
	m = next.(Model)
	if len(m.session.Sources()) != 2 {
		t.Fatal("the list changed before saving")
	}
	next, cmd = m.Update(tea.KeyPressMsg{Text: "s", Code: 's'})
	m = next.(Model)
	next, _ = m.Update(cmd())
	m = next.(Model)
	if m.editing || len(m.session.Sources()) != 1 {
		t.Errorf("editing %v sources %v", m.editing, m.session.Sources())
	}
}
