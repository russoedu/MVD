package windowsnapshot

import (
	"encoding/json"
	"strconv"
)

// Progress is how far a song is: a part from 0 to 1, unknown while it cannot be measured, or none.
type Progress struct {
	Part    float64
	Unknown bool
	Shown   bool
}

// Part is progress measured as a part from 0 to 1.
func Part(part float64) Progress { return Progress{Part: min(max(part, 0), 1), Shown: true} }

// Unknown is progress that cannot be measured yet.
func Unknown() Progress { return Progress{Unknown: true, Shown: true} }

// MarshalJSON writes the progress as the contract has it: a number, "unknown" or null.
func (p Progress) MarshalJSON() ([]byte, error) {
	switch {
	case !p.Shown:
		return []byte("null"), nil
	case p.Unknown:
		return []byte(`"unknown"`), nil
	}

	return []byte(strconv.FormatFloat(p.Part, 'f', 4, 64)), nil
}

// UnmarshalJSON reads what MarshalJSON writes.
func (p *Progress) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case "null":
		*p = Progress{}
		return nil
	case `"unknown"`:
		*p = Unknown()
		return nil
	}
	var part float64
	if err := json.Unmarshal(data, &part); err != nil {
		return err
	}
	*p = Part(part)

	return nil
}
