package install

// uninstallEntry is what Windows lists for the app in Settings > Apps.
type uninstallEntry struct {
	DisplayName     string
	DisplayVersion  string
	Publisher       string
	InstallLocation string
	DisplayIcon     string
	UninstallString string
}

// uninstallEntryFor describes the program installed at target. The uninstall command is
// the program itself with -uninstall, so Windows needs nothing else from the install.
func uninstallEntryFor(target installTarget, version string) uninstallEntry {
	return uninstallEntry{
		DisplayName:     "MVD",
		DisplayVersion:  version,
		Publisher:       "MVD",
		InstallLocation: target.Folder,
		DisplayIcon:     target.Program,
		UninstallString: `"` + target.Program + `" -uninstall`,
	}
}

// strings are the text values of the entry, by the name Windows reads them under.
func (e uninstallEntry) strings() map[string]string {
	return map[string]string{
		"DisplayName":     e.DisplayName,
		"DisplayVersion":  e.DisplayVersion,
		"Publisher":       e.Publisher,
		"InstallLocation": e.InstallLocation,
		"DisplayIcon":     e.DisplayIcon,
		"UninstallString": e.UninstallString,
	}
}

// numbers are the switches of the entry: the app is neither modified nor repaired from
// Settings, only removed.
func (e uninstallEntry) numbers() map[string]uint32 {
	return map[string]uint32{"NoModify": 1, "NoRepair": 1}
}
