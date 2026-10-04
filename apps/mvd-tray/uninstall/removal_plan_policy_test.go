package uninstall

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"youtube-downloader/apps/mvd-tray/install"
)

func everything(string) bool { return true }

func nothing(string) bool { return false }

func windowsFootprint(root string) install.Footprint {
	user := filepath.Join(root, "local", "Programs", "MVD")
	system := filepath.Join(root, "pf", "MVD")

	return install.Footprint{
		Places: []install.Place{
			{Folder: system, Program: filepath.Join(system, "mvd.exe")},
			{Folder: user, Program: filepath.Join(user, "mvd.exe")},
		},
		Entries: []string{filepath.Join(root, "menu", "MVD.lnk"), filepath.Join(root, "all", "MVD.lnk")},
	}
}

func TestTheInstalledFolderAndShortcutsAreRemovedAndThePreferencesAreKept(t *testing.T) {
	root := t.TempDir()
	foot := windowsFootprint(root)
	user := foot.Places[1]

	plan := buildPlan(planInput{GOOS: "windows", Exe: user.Program, AppDir: filepath.Join(root, "roaming", "mvd"), Footprint: foot, Exists: everything})

	if plan.Program != user.Program {
		t.Errorf("program = %s", plan.Program)
	}
	if !reflect.DeepEqual(plan.Folders, []string{foot.Places[0].Folder, user.Folder}) {
		t.Errorf("folders = %v", plan.Folders)
	}
	if !reflect.DeepEqual(plan.Files, foot.Entries) {
		t.Errorf("files = %v", plan.Files)
	}
	if !plan.Registry {
		t.Error("the Settings > Apps entry should go on Windows")
	}
	if plan.AppData != "" || len(plan.AppDataEntries) != 0 {
		t.Errorf("preferences must be kept unless asked: %s %v", plan.AppData, plan.AppDataEntries)
	}
}

func TestOnlyTheProgramIsRemovedWhenItRunsFromADownloadsFolder(t *testing.T) {
	root := t.TempDir()
	foot := windowsFootprint(root)
	downloads := filepath.Join(root, "Downloads")
	exe := filepath.Join(downloads, "mvd (1).exe")

	plan := buildPlan(planInput{GOOS: "windows", Exe: exe, Footprint: foot, Exists: nothing})

	if plan.Program != exe || len(plan.Folders) != 0 || len(plan.Files) != 0 {
		t.Errorf("plan = %+v", plan)
	}
	for _, path := range append(append([]string{plan.Program}, plan.Folders...), plan.Files...) {
		if path == downloads {
			t.Error("the Downloads folder must never be removed")
		}
	}
}

func TestAPlaceWithNothingInItIsLeftOut(t *testing.T) {
	root := t.TempDir()
	foot := windowsFootprint(root)

	plan := buildPlan(planInput{GOOS: "windows", Exe: filepath.Join(root, "elsewhere", "mvd.exe"), Footprint: foot, Exists: nothing})

	if len(plan.Folders) != 0 || len(plan.Files) != 0 {
		t.Errorf("plan = %+v", plan)
	}
}

func TestADevelopersBuildIsNotRemoved(t *testing.T) {
	plan := buildPlan(planInput{GOOS: "linux", Dev: true, Exe: "/tmp/go-build/exe", Exists: everything})

	if plan.Program != "" {
		t.Errorf("program = %s", plan.Program)
	}
}

func TestOnLinuxTheSharedBinFolderIsNeverRemovedOnlyTheProgramAndTheMenuEntry(t *testing.T) {
	root := t.TempDir()
	bin := filepath.Join(root, ".local", "bin")
	program := filepath.Join(bin, "mvd")
	desktop := filepath.Join(root, "apps", "mvd.desktop")
	foot := install.Footprint{Places: []install.Place{{Folder: bin, Program: program, Shared: true}}, Entries: []string{desktop}}

	plan := buildPlan(planInput{GOOS: "linux", Exe: program, Footprint: foot, Exists: everything})

	if plan.Program != program || len(plan.Folders) != 0 {
		t.Errorf("plan = %+v", plan)
	}
	if !reflect.DeepEqual(plan.Files, []string{desktop}) {
		t.Errorf("files = %v", plan.Files)
	}
}

func TestOnLinuxAnotherCopyInTheBinFolderIsRemovedToo(t *testing.T) {
	root := t.TempDir()
	program := filepath.Join(root, ".local", "bin", "mvd")
	foot := install.Footprint{Places: []install.Place{{Folder: filepath.Dir(program), Program: program, Shared: true}}}

	plan := buildPlan(planInput{GOOS: "linux", Exe: filepath.Join(root, "Downloads", "mvd"), Footprint: foot, Exists: everything})

	if !reflect.DeepEqual(plan.Files, []string{program}) {
		t.Errorf("files = %v", plan.Files)
	}
}

func TestAMacBundleIsRemovedAsAWhole(t *testing.T) {
	root := t.TempDir()
	bundle := filepath.Join(root, "Applications", "MVD.app")
	program := filepath.Join(bundle, "Contents", "MacOS", "mvd")
	foot := install.Footprint{Places: []install.Place{{Folder: bundle, Program: program, Bundle: true}}}

	plan := buildPlan(planInput{GOOS: "darwin", Exe: program, Footprint: foot, Exists: everything})

	if plan.Program != program || !reflect.DeepEqual(plan.Folders, []string{bundle}) || plan.Registry {
		t.Errorf("plan = %+v", plan)
	}
}

func TestPreferencesAreRemovedFromTheAppDataFolderWhenAsked(t *testing.T) {
	root := t.TempDir()
	appDir := filepath.Join(root, "mvd")

	plan := buildPlan(planInput{
		GOOS: "linux", Exe: filepath.Join(root, "x"), AppDir: appDir, DeletePreferences: true,
		AppDirEntries: []string{"config.conf", "list.txt", "bin"}, Exists: nothing,
	})

	want := []string{filepath.Join(appDir, "config.conf"), filepath.Join(appDir, "list.txt"), filepath.Join(appDir, "bin")}
	if plan.AppData != appDir || !reflect.DeepEqual(plan.AppDataEntries, want) {
		t.Errorf("plan = %+v", plan)
	}
}

func TestAFolderThatIsNotCalledMvdIsNeverTreatedAsAppData(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{root, filepath.Join(root, "Documents"), "", "relative/mvd", string(filepath.Separator)} {
		plan := buildPlan(planInput{GOOS: "linux", AppDir: dir, DeletePreferences: true, AppDirEntries: []string{"a"}, Exists: nothing})

		if plan.AppData != "" || len(plan.AppDataEntries) != 0 {
			t.Errorf("%q was treated as app data: %+v", dir, plan)
		}
	}
}

func TestDownloadsInsideTheAppDataFolderAreKept(t *testing.T) {
	root := t.TempDir()
	appDir := filepath.Join(root, "mvd")
	downloads := filepath.Join(appDir, "videos")

	plan := buildPlan(planInput{
		GOOS: "linux", AppDir: appDir, DeletePreferences: true, Keep: []string{filepath.Join(downloads, "music")},
		AppDirEntries: []string{"config.conf", "videos"}, Exists: nothing,
	})

	if !reflect.DeepEqual(plan.AppDataEntries, []string{filepath.Join(appDir, "config.conf")}) {
		t.Errorf("entries = %v", plan.AppDataEntries)
	}
	if len(plan.Kept) != 1 || !strings.Contains(plan.Kept[0], "videos") {
		t.Errorf("kept = %v", plan.Kept)
	}
}

func TestAnInstallFolderThatHoldsTheDownloadsLosesOnlyItsProgram(t *testing.T) {
	root := t.TempDir()
	foot := windowsFootprint(root)
	user := foot.Places[1]
	downloads := filepath.Join(user.Folder, "videos")

	plan := buildPlan(planInput{GOOS: "windows", Exe: user.Program, Keep: []string{downloads}, Footprint: foot, Exists: func(path string) bool { return path == user.Program }})

	for _, folder := range plan.Folders {
		if folder == user.Folder {
			t.Errorf("the folder that holds the downloads is in the removal set: %v", plan.Folders)
		}
	}
	if len(plan.Kept) == 0 {
		t.Error("the person should be told why the folder stays")
	}
}

func TestNothingInThePlanIsOrHoldsAKeptFolder(t *testing.T) {
	root := t.TempDir()
	keep := filepath.Join(root, "Music Videos")
	appDir := filepath.Join(root, "mvd")
	foot := windowsFootprint(root)

	plan := buildPlan(planInput{
		GOOS: "windows", Exe: foot.Places[1].Program, AppDir: appDir, DeletePreferences: true, Keep: []string{keep},
		AppDirEntries: []string{"config.conf", "bin"}, Footprint: foot, Exists: everything,
	})

	all := append([]string{plan.Program, plan.AppData}, plan.Folders...)
	all = append(append(all, plan.Files...), plan.AppDataEntries...)
	for _, path := range all {
		if path == "" {
			continue
		}
		if within("windows", path, keep) {
			t.Errorf("%s is or holds the kept folder %s", path, keep)
		}
	}
}

func TestTheAppsOwnDataIsStillRemovedWhenTheDownloadsFolderIsAboveIt(t *testing.T) {
	root := t.TempDir()
	appDir := filepath.Join(root, "AppData", "mvd")
	foot := windowsFootprint(root)

	plan := buildPlan(planInput{
		GOOS: "windows", Exe: foot.Places[1].Program, AppDir: appDir, DeletePreferences: true, Keep: []string{root},
		AppDirEntries: []string{"config.conf"}, Footprint: foot, Exists: everything,
	})

	if len(plan.AppDataEntries) != 1 || len(plan.Folders) != 2 || len(plan.Kept) != 0 {
		t.Errorf("plan = %+v", plan)
	}
}

func TestWindowsPathsAreComparedWithoutCase(t *testing.T) {
	if !within("windows", filepath.Join("C:", "Users", "A", "MVD"), filepath.Join("c:", "users", "a", "mvd", "mvd.exe")) {
		t.Error("the folder should contain the program whatever the case")
	}
	if within("linux", "/a/MVD", "/a/mvd/x") {
		t.Error("Linux paths are case sensitive")
	}
	if within("linux", "/a/mvd", "/a/mvdx/y") {
		t.Error("a folder whose name merely starts the same is not inside")
	}
}

func TestTheConfirmationNamesEveryPathAndSaysWhatIsKept(t *testing.T) {
	plan := removalPlan{Program: "/p/mvd", Folders: []string{"/p/MVD"}, Files: []string{"/m/MVD.lnk"}, Registry: true}

	kept := confirmation(plan, "/data/mvd", false)
	deleted := confirmation(plan, "/data/mvd", true)

	for _, want := range []string{"/p/mvd", "/p/MVD", "/m/MVD.lnk", "Settings > Apps", "never touched"} {
		if !strings.Contains(kept, want) {
			t.Errorf("missing %q in:\n%s", want, kept)
		}
	}
	if !strings.Contains(kept, "/data/mvd are kept") || !strings.Contains(deleted, "/data/mvd will be deleted") {
		t.Errorf("kept:\n%s\ndeleted:\n%s", kept, deleted)
	}
}
