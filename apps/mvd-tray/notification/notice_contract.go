package notification

// Notice is something to tell the person without taking over their screen.
type Notice struct {
	Title string
	Text  string
	// Failure draws it as an error rather than information.
	Failure bool
}
