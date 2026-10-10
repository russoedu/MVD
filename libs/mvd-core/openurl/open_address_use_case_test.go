package openurl

import "testing"

func TestEachSystemOpensAnAddressItsOwnWay(t *testing.T) {
	const address = "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
	cases := map[string]string{"windows": "rundll32", "darwin": "open", "linux": "xdg-open", "freebsd": "xdg-open"}
	for goos, want := range cases {
		name, args := command(goos, address)
		if name != want || args[len(args)-1] != address {
			t.Errorf("%s: %s %v", goos, name, args)
		}
	}
}

func TestOnlyWebAddressesAreOpened(t *testing.T) {
	for _, bad := range []string{"", "file:///etc/passwd", "calc.exe", "javascript:alert(1)", "https://", `C:\Windows\notepad.exe`, "ms-settings:"} {
		if err := Open(bad); err == nil {
			t.Errorf("Open(%q) should be refused", bad)
		}
	}
}
