package titlematch

import "strings"

// Mentions reports whether a text names a name or a phrase as whole words,
// whatever the case, accents and punctuation ("Beyonce" is in "BEYONCÉ - Halo").
// An initialism is also found written together: "M/A/R/R/S" is in "Pump Up The
// Volume MARRS".
func Mentions(text, name string) bool {
	phrase := Normalize(name)
	if phrase == "" {
		return false
	}
	spelled := Normalize(text)
	if ContainsWords(spelled, phrase) {
		return true
	}

	letters := strings.Fields(phrase)
	for _, letter := range letters {
		if len([]rune(letter)) != 1 {
			return false
		}
	}
	if len(letters) < 3 {
		return false
	}
	return ContainsWords(spelled, strings.Join(letters, ""))
}
