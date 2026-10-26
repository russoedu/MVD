package windowsnapshot

import (
	"encoding/json"
	"testing"
)

func TestProgressIsWrittenAsANumberUnknownOrNull(t *testing.T) {
	for want, p := range map[string]Progress{"0.3400": Part(0.34), `"unknown"`: Unknown(), "null": {}} {
		data, err := json.Marshal(p)
		if err != nil || string(data) != want {
			t.Fatalf("Marshal(%+v) = %s, %v; want %s", p, data, err, want)
		}
		var back Progress
		if err := json.Unmarshal(data, &back); err != nil || back != p {
			t.Fatalf("Unmarshal(%s) = %+v, %v; want %+v", data, back, err, p)
		}
	}
}
