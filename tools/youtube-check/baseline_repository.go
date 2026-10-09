package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
)

// loadBaseline reads the recorded answers; a file that is not there is an empty baseline.
func loadBaseline(path string) (Baseline, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Baseline{}, nil
	}
	if err != nil {
		return Baseline{}, err
	}
	var baseline Baseline
	if err := json.Unmarshal(data, &baseline); err != nil {
		return Baseline{}, fmt.Errorf("%s is not a baseline: %w", path, err)
	}

	return baseline, nil
}

// saveBaseline writes the answers sorted by song, so that a change shows in a diff as
// the songs whose answer changed and nothing else.
func saveBaseline(path string, baseline Baseline) error {
	sort.Slice(baseline.Songs, func(i, j int) bool { return baseline.Songs[i].ID < baseline.Songs[j].ID })
	data, err := json.MarshalIndent(baseline, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// baselineFrom records the answers of this run.
func baselineFrom(playlist string, outcomes []Outcome) Baseline {
	baseline := Baseline{Playlist: playlist}
	for _, o := range outcomes {
		baseline.Songs = append(baseline.Songs, Song{ID: o.ID, Title: o.Title, Official: o.Found})
	}

	return baseline
}
