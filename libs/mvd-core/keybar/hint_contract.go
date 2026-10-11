// Package keybar is the key bar every screen shows at its bottom: the hints it lists, the
// hint under a column, the key press a hint stands for, and the typed hints of a page.
package keybar

// Hint is one entry of a key bar: a key and what it does.
type Hint struct{ Key, Desc string }
