package official

// Mentions reports whether a text names a name or a phrase as whole words,
// whatever the case, accents and punctuation ("Beyonce" is in "BEYONCÉ - Halo").
func Mentions(text, name string) bool {
	phrase := normalize(name)
	return phrase != "" && containsWords(normalize(text), phrase)
}
