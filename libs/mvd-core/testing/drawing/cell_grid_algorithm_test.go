package drawing

import (
	"testing"
)

func TestGridReadsStylesAndWidths(t *testing.T) {
	g := gridOf("\x1b[1;38;2;255;0;127mA\x1b[0m\x1b[48;5;236m界\x1b[0m!", 6)
	if len(g) != 1 || len(g[0]) != 6 {
		t.Fatalf("grid %v", g)
	}
	if c := g[0][0]; c.Text != "A" || c.FG != "#ff007f" || c.Attrs != "b" {
		t.Errorf("cell 0 %v", c)
	}
	if c := g[0][1]; c.Text != "界" || c.BG != "c236" {
		t.Errorf("cell 1 %v", c)
	}
	if c := g[0][2]; c.Text != "" || c.BG != "c236" {
		t.Errorf("the second cell of a wide character %v", c)
	}
	if c := g[0][3]; c.Text != "!" || c.BG != "" {
		t.Errorf("cell 3 %v", c)
	}
	if d := SameCells("\x1b[31mx\x1b[0m", "\x1b[38;5;1mx\x1b[0m", 1, 5); len(d) != 0 {
		t.Errorf("the same colour written two ways should match: %v", d)
	}
}
