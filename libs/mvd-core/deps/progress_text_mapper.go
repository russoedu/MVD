package deps

import (
	"fmt"
	"runtime"
	"strings"
)

// progressText is how a step is shown on the terminal; empty for a kind that is not shown.
func progressText(e Event) string {
	switch e.Kind {
	case EventMissing:
		return fmt.Sprintf("\n[!] Missing dependency/dependencies detected: %s\n", strings.Join(e.Names, ", ")) +
			fmt.Sprintf("[+] Automatically downloading dependencies for [%s/%s] into %s...\n\n", runtime.GOOS, runtime.GOARCH, e.Dir)
	case EventDownloading:
		return fmt.Sprintf("    Downloading %s from %s...\n", e.Name, e.URL)
	case EventInstalled:
		return fmt.Sprintf("    [OK] Successfully installed %s!\n\n", e.Name)
	case EventFailed:
		return fmt.Sprintf("    [!] Could not install %s: %v\n", e.Name, e.Err)
	case EventUpToDate:
		return fmt.Sprintf("    %s is up to date (%s)\n", e.Name, e.Detail)
	case EventUpdated:
		return fmt.Sprintf("    [OK] Updated %s to %s\n", e.Name, e.Detail)
	case EventUpdateFailed:
		return fmt.Sprintf("    [!] Could not update %s: %v\n", e.Name, e.Err)
	}

	return ""
}
