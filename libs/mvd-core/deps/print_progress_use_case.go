package deps

import "fmt"

// PrintProgress writes each step to the terminal.
func PrintProgress(e Event) {
	fmt.Print(progressText(e))
}
