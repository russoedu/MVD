package loading

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestRenderNamesWhatIsLoading(t *testing.T) {
	got := ansi.Strip(Render(Progress{Playlists: []Playlist{
		{Name: "Rock", Songs: 12, Listed: true},
		{Name: "Pop"},
	}}, 80, 20))

	for _, want := range []string{"Loading playlists", "1 of 2", "✓ Rock", "12 songs", "· Pop"} {
		if !strings.Contains(got, want) {
			t.Errorf("screen lacks %q:\n%s", want, got)
		}
	}
	if n := len(strings.Split(got, "\n")); n != 20 {
		t.Errorf("the screen is %d lines, want 20", n)
	}
}

func TestRenderWhilePreparingHasNoBar(t *testing.T) {
	got := ansi.Strip(Render(Progress{Detail: "Getting ready…"}, 80, 10))
	if strings.Contains(got, "░") || !strings.Contains(got, "Getting ready…") {
		t.Errorf("unexpected screen:\n%s", got)
	}
}
