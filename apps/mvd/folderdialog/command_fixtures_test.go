package folderdialog

import (
	"encoding/base64"
	"testing"
	"unicode/utf16"
)

func installed(names ...string) func(string) bool {
	return func(name string) bool {
		for _, n := range names {
			if n == name {
				return true
			}
		}

		return false
	}
}

func decodePowerShell(t *testing.T, args []string) string {
	t.Helper()
	for i, a := range args {
		if a == "-EncodedCommand" && i+1 < len(args) {
			raw, err := base64.StdEncoding.DecodeString(args[i+1])
			if err != nil || len(raw)%2 != 0 {
				t.Fatalf("not UTF-16LE base64: %v", err)
			}
			units := make([]uint16, len(raw)/2)
			for j := range units {
				units[j] = uint16(raw[2*j]) | uint16(raw[2*j+1])<<8
			}

			return string(utf16.Decode(units))
		}
	}
	t.Fatal("no -EncodedCommand")

	return ""
}

func contains(list []string, want string) bool {
	for _, item := range list {
		if item == want {
			return true
		}
	}

	return false
}
