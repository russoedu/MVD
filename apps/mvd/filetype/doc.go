// Package filetype makes the system know what a .mvd file is: it names it a download list,
// gives it the icon of the app, and opens it with MVD when it is double clicked. Windows and
// Linux are told at every start (it is cheap, and it follows the program when it moves);
// macOS is told by the Info.plist of the bundle (see apps/mvd/macbundle).
package filetype

// Extension is the end of the name of a download list.
const Extension = ".mvd"

// ProgramID names the kind of file to Windows.
const ProgramID = "MVD.DownloadList"

// Description is what a file manager calls the kind of file.
const Description = "MVD download list"

// MimeType is the kind of file to Linux.
const MimeType = "application/x-mvd-list"
