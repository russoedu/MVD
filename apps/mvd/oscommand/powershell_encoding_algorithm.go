package oscommand

import (
	"encoding/base64"
	"unicode/utf16"
)

// EncodePowerShell is the form -EncodedCommand takes: the script as UTF-16LE, in
// base64. It avoids every quoting rule between this program and PowerShell.
func EncodePowerShell(script string) string {
	units := utf16.Encode([]rune(script))
	raw := make([]byte, 0, len(units)*2)
	for _, u := range units {
		raw = append(raw, byte(u), byte(u>>8))
	}

	return base64.StdEncoding.EncodeToString(raw)
}
