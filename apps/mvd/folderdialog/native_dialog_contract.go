package folderdialog

// dialogKind is what a file window asks the person to choose.
type dialogKind int

const (
	pickFolder dialogKind = iota
	openFile
	saveFile
)

// dialogRequest is what a native file window is asked to show: start is the folder it opens in,
// name the file name offered by a save window, and extension (without its dot) and label the
// kind of file it shows first.
type dialogRequest struct {
	kind                          dialogKind
	start, name, extension, label string
}

// title is what the window says at its top.
func (r dialogRequest) title() string {
	switch r.kind {
	case openFile:
		return "Open a file"
	case saveFile:
		return "Save as"
	}

	return "Choose a folder"
}
