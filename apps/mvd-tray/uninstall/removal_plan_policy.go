package uninstall

import (
	"fmt"
	"path/filepath"
	"strings"

	"youtube-downloader/apps/mvd-tray/install"
)

// appDataFolderName is the last part of the app-data folder (appdir.Dir ends in it on
// every system). A folder that is called anything else is never treated as the app's.
const appDataFolderName = "mvd"

// planInput is everything that decides what an uninstall removes.
type planInput struct {
	GOOS string
	// Dev is true for a build made on a developer's machine, whose program is left alone.
	Dev bool
	// Exe is the program that is running.
	Exe string
	// AppDir is the app-data folder and AppDirEntries are the names directly inside it.
	AppDir        string
	AppDirEntries []string
	// DeletePreferences also removes the app-data folder.
	DeletePreferences bool
	// Keep are folders that hold the person's own files (the downloads and the log
	// folder). Nothing that is or contains one of them is ever removed.
	Keep      []string
	Footprint install.Footprint
	Exists    func(path string) bool
}

// removalPlan is what an uninstall removes, and nothing else.
type removalPlan struct {
	// Program is the running program. It goes first, and on Windows it can only be moved
	// aside, because the system will not delete a program that is running.
	Program string
	// Folders are the app's own install folders and macOS bundles, removed whole.
	Folders []string
	// Files are other copies of the program, shortcuts and menu entries.
	Files []string
	// Registry is true when the Settings > Apps entry is removed.
	Registry bool
	// AppData is the app-data folder, and AppDataEntries what is removed from it. The
	// folder itself goes last, and only if it is empty by then.
	AppData        string
	AppDataEntries []string
	// Kept explains what was left alone although it would otherwise have been removed.
	Kept []string
}

// buildPlan decides what to remove. It only ever names:
//   - the running program, unless this is a developer's build;
//   - the install places the installer knows about, when something is there: the whole
//     folder (or bundle) for the app's own, only the program for a shared one;
//   - the shortcuts and menu entries the installer writes;
//   - with DeletePreferences, what is inside the app-data folder;
//
// and it leaves out anything that is or contains a Keep folder, because removing it
// would remove the person's files. Something that merely sits inside a Keep folder (the
// app's own data under a downloads folder set to the home folder, say) is the app's.
func buildPlan(in planInput) removalPlan {
	var plan removalPlan
	same := func(a, b string) bool { return samePath(in.GOOS, a, b) }

	if !in.Dev && in.Exe != "" {
		plan.Program = in.Exe
	}

	seenFiles := map[string]bool{}
	addFile := func(path string) {
		key := normalise(in.GOOS, path)
		if seenFiles[key] || (plan.Program != "" && same(plan.Program, path)) {
			return
		}
		seenFiles[key] = true
		plan.Files = append(plan.Files, path)
	}

	for _, place := range in.Footprint.Places {
		holdsExe := in.Exe != "" && within(in.GOOS, place.Folder, in.Exe)
		if !holdsExe && !in.Exists(place.Program) {
			continue
		}
		if place.Shared {
			addFile(place.Program)

			continue
		}
		if keeps := holding(in.GOOS, in.Keep, place.Folder); keeps != "" {
			plan.Kept = append(plan.Kept, fmt.Sprintf("%s holds your files (%s), so only the program in it is removed", place.Folder, keeps))
			addFile(place.Program)

			continue
		}
		plan.Folders = append(plan.Folders, place.Folder)
	}

	for _, entry := range in.Footprint.Entries {
		if in.Exists(entry) {
			addFile(entry)
		}
	}
	plan.Registry = in.GOOS == "windows"

	if in.DeletePreferences && safeAppData(in.AppDir) {
		plan.AppData = in.AppDir
		for _, name := range in.AppDirEntries {
			entry := filepath.Join(in.AppDir, name)
			if keeps := holding(in.GOOS, in.Keep, entry); keeps != "" {
				plan.Kept = append(plan.Kept, fmt.Sprintf("%s holds your files (%s), so it is kept", entry, keeps))

				continue
			}
			plan.AppDataEntries = append(plan.AppDataEntries, entry)
		}
	}

	return plan
}

// safeAppData reports whether dir is plausibly the app-data folder: an absolute path
// whose last part is the app's folder name, with something above it.
func safeAppData(dir string) bool {
	if dir == "" || !filepath.IsAbs(dir) {
		return false
	}
	dir = filepath.Clean(dir)

	return strings.EqualFold(filepath.Base(dir), appDataFolderName) && filepath.Dir(dir) != dir
}

// confirmation words what is about to happen, so that the person sees every path before
// anything is removed.
func confirmation(plan removalPlan, appDir string, deletePreferences bool) string {
	var lines []string
	if plan.Program != "" {
		lines = append(lines, plan.Program)
	}
	lines = append(lines, plan.Folders...)
	lines = append(lines, plan.Files...)

	text := "MVD will be removed from this computer:\n\n"
	for _, line := range lines {
		text += "  " + line + "\n"
	}
	if plan.Registry {
		text += "  its entry in Settings > Apps\n"
	}
	text += "\n"
	if deletePreferences {
		text += fmt.Sprintf("Your preferences, your list and the downloaded tools in %s will be deleted too.\n", appDir)
	} else {
		text += fmt.Sprintf("Your preferences, your list and the downloaded tools in %s are kept.\n", appDir)
	}
	text += "Your downloaded videos, and the folder they are in, are never touched."

	return text
}

// samePath reports whether a and b are the same path. Windows paths ignore case.
func samePath(goos, a, b string) bool {
	return normalise(goos, a) == normalise(goos, b)
}

func normalise(goos, path string) string {
	path = filepath.Clean(path)
	if goos == "windows" {
		return strings.ToLower(path)
	}

	return path
}

// within reports whether path is dir or sits somewhere inside it.
func within(goos, dir, path string) bool {
	dir, path = normalise(goos, dir), normalise(goos, path)
	if dir == path {
		return true
	}

	return strings.HasPrefix(path, strings.TrimSuffix(dir, string(filepath.Separator))+string(filepath.Separator))
}

// holding returns the first of keeps that is path or sits inside it, which removing path
// would remove, or an empty string when there is none.
func holding(goos string, keeps []string, path string) string {
	for _, keep := range keeps {
		if keep != "" && within(goos, path, keep) {
			return keep
		}
	}

	return ""
}
