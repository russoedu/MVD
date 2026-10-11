package keybar

// Of turns the hints of a screen into the typed hints of a page's key bar.
func Of[T ~struct{ Key, Label string }](hints []Hint) []T {
	out := make([]T, 0, len(hints))
	for _, h := range hints {
		out = append(out, T(struct{ Key, Label string }{h.Key, h.Desc}))
	}
	return out
}
