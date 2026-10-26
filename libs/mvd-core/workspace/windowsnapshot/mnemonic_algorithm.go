package windowsnapshot

import "unicode"

// Mnemonics are the keyboard letters of a group of labels (the headings of a menu bar, the entries
// of one menu), one each and none twice, as Windows programs chose them: the first letter of a
// word when it is free, else any free letter or digit of the label, else none. An empty label
// gets "".
func Mnemonics(labels []string) []string {
	taken := map[rune]bool{}
	out := make([]string, len(labels))
	for i, label := range labels {
		r, ok := firstFree(label, taken, true)
		if !ok {
			r, ok = firstFree(label, taken, false)
		}
		if ok {
			out[i] = string(r)
			taken[unicode.ToLower(r)] = true
		}
	}

	return out
}

// firstFree is the first letter or digit of label not taken, only at the start of a word when
// wordStarts.
func firstFree(label string, taken map[rune]bool, wordStarts bool) (rune, bool) {
	start := true
	for _, r := range label {
		usable := r < unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsDigit(r))
		if usable && (start || !wordStarts) && !taken[unicode.ToLower(r)] {
			return r, true
		}
		start = !usable
	}

	return 0, false
}
