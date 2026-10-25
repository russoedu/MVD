// Package folderbrowser is renderer-neutral: it holds what the folder picker does, apart from how it
// is drawn. Which folder is shown, which sub-folder the cursor is on, the folder being made and what
// went wrong are its state; moving, opening, going up and making a folder are its intents. The disk
// is behind the Disk interface.
package folderbrowser
