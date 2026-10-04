package main

// dialogCommand is a program that shows a folder chooser and prints the choice.
type dialogCommand struct {
	name string
	args []string
	// env is added to the program's environment. The starting folder travels here (or
	// as a separate argument), never inside the script text, so a folder name can
	// never be read as code.
	env []string
}
