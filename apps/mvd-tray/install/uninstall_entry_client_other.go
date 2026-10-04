//go:build !windows

package install

// registerUninstallEntry does nothing here: only Windows lists programs in a settings page.
func registerUninstallEntry(installTarget, string) error { return nil }

// RemoveUninstallEntries does nothing here, for the same reason.
func RemoveUninstallEntries() error { return nil }
