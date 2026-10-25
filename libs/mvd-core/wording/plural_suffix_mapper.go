// Package wording has the small pieces of wording the screens share.
package wording

import "fmt"

// Plural is the suffix a count takes in English: "" for one, "s" for any other number.
func Plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// Counted writes a count with its noun: "1 file", "4 files".
func Counted(n int, word string) string {
	return fmt.Sprintf("%d %s%s", n, word, Plural(n))
}
