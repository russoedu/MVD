// Package wording has the small pieces of wording the screens share.
package wording

// Plural is the suffix a count takes in English: "" for one, "s" for any other number.
func Plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
