//go:build !windows

package console

import "fmt"

// ShowInfo prints the message to the terminal: these systems show a notification
// instead, from the caller.
func ShowInfo(message string) {
	fmt.Println(message)
}
