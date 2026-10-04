package main

import (
	"flag"
	"fmt"
	"os"

	"youtube-downloader/apps/mvd-tray/macbundle"
)

// writebundle builds MVD.app from an already built program. The release build runs it
// before making the .dmg, so the app installer and the .dmg share one bundle layout.
func main() {
	program := flag.String("program", "", "the built mvd program to put in the bundle")
	version := flag.String("version", "dev", "the version for the Info.plist")
	icon := flag.String("icon", "", "an .icns file for the bundle (optional)")
	out := flag.String("out", "", "where to write the MVD.app bundle")
	flag.Parse()

	if *program == "" || *out == "" {
		_, _ = fmt.Fprintln(os.Stderr, "usage: writebundle -program <file> -out <MVD.app> [-version <v>] [-icon <file.icns>]")
		os.Exit(2)
	}
	if err := macbundle.Write(*out, *program, *version, *icon); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "writebundle:", err)
		os.Exit(1)
	}
}
