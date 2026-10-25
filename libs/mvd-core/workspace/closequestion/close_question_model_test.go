package closequestion

import "testing"

func TestTheQuestionBeginsOnTheSafeAnswer(t *testing.T) {
	if s := Ask(); !s.On || s.Focus != AnswerKeep {
		t.Errorf("asked %+v", s)
	}
}

func TestKeysAnswerOrMoveTheFocus(t *testing.T) {
	cases := []struct {
		key  string
		from State
		want Outcome
	}{
		{"y", Ask(), QuitNow},
		{"n", Ask(), Dismissed},
		{"esc", Ask(), Dismissed},
		{"enter", State{On: true, Focus: AnswerQuit}, QuitNow},
		{"space", Ask(), Dismissed},
		{"x", Ask(), Undecided},
	}
	for _, c := range cases {
		if _, got := c.from.Key(c.key); got != c.want {
			t.Errorf("key %q from %+v: got %v, want %v", c.key, c.from, got, c.want)
		}
	}
	moved, out := Ask().Key("left")
	if out != Undecided || moved.Focus != AnswerQuit || !moved.On {
		t.Errorf("an arrow should move the focus and keep the question: %+v %v", moved, out)
	}
}

func TestAClickAnswersAndHoverMovesTheFocus(t *testing.T) {
	if s, out := Ask().Click(AnswerQuit); out != QuitNow || s.On {
		t.Errorf("click on quit: %+v %v", s, out)
	}
	if s := Ask().Hover(AnswerQuit); s.Focus != AnswerQuit || !s.On {
		t.Errorf("hover: %+v", s)
	}
}

func TestTheQuestionSaysHowManySongsWouldBeLost(t *testing.T) {
	if what, _ := Words(1); what != "1 song is still downloading." {
		t.Errorf("one song: %q", what)
	}
	if what, ask := Words(4); what != "4 songs are still downloading." || ask != "Quit and stop them?" {
		t.Errorf("four songs: %q / %q", what, ask)
	}
}
