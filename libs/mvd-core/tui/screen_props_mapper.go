package tui

// hintsOf turns the key hints of a screen into the typed hints of a key bar.
func hintsOf[T ~struct{ Key, Label string }](hints []keyHint) []T {
	out := make([]T, 0, len(hints))
	for _, h := range hints {
		out = append(out, T(struct{ Key, Label string }{h.key, h.desc}))
	}
	return out
}
