package prefspopup

import "testing"

func thumbAt(thumb []bool) (first, size int) {
	first = -1
	for i, on := range thumb {
		if on {
			if first < 0 {
				first = i
			}
			size++
		}
	}
	return first, size
}

func TestAPageThatFitsHasNoScrollbarThumb(t *testing.T) {
	if first, size := thumbAt(scrollbarThumb(10, 10, 0)); size != 0 {
		t.Errorf("a page that fits showed a thumb at %d", first)
	}
}

func TestTheThumbStartsAtTheTopAndEndsAtTheBottom(t *testing.T) {
	if first, _ := thumbAt(scrollbarThumb(40, 10, 0)); first != 0 {
		t.Errorf("at the top the thumb starts at %d", first)
	}
	first, size := thumbAt(scrollbarThumb(40, 10, 30))
	if first+size != 10 {
		t.Errorf("at the bottom the thumb ends at %d, want 10", first+size)
	}
}

func TestTheThumbIsTallerForAPageThatMostlyShows(t *testing.T) {
	_, small := thumbAt(scrollbarThumb(100, 10, 0))
	_, large := thumbAt(scrollbarThumb(14, 10, 0))
	if small < 1 || large <= small {
		t.Errorf("thumb sizes %d and %d", small, large)
	}
}

func TestTheThumbMovesDownWithTheScroll(t *testing.T) {
	prev := -1
	for scroll := 0; scroll <= 30; scroll += 5 {
		first, _ := thumbAt(scrollbarThumb(40, 10, scroll))
		if first < prev {
			t.Errorf("the thumb went up from %d to %d at scroll %d", prev, first, scroll)
		}
		prev = first
	}
}
