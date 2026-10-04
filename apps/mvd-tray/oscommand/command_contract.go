package oscommand

// Command is a program to run that shows something to the person (a folder chooser, a
// notification, a question) or makes a shortcut, with everything it needs.
type Command struct {
	Name string
	Args []string
	// Env is added to the program's environment. The words and paths travel here (or as
	// separate arguments), never inside the script text, so they can never be read as
	// code.
	Env []string
}
