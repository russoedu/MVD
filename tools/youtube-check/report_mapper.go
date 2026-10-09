package main

import (
	"fmt"
	"strings"
)

// maxListed is how many songs a report names; the rest are counted.
const maxListed = 25

// reportOf is the verdict as a Markdown text, for the issue and the agent: what failed,
// and for each song that differs what was recorded and what YouTube gives now.
func reportOf(playlist string, verdict Verdict, maxDifferent float64) string {
	var b strings.Builder
	state := "still match"
	if verdict.Failed {
		state = "CHANGED"
	}
	fmt.Fprintf(&b, "## YouTube check: the answers %s\n\n", state)
	fmt.Fprintf(&b, "Playlist: %s\n\n", playlist)
	fmt.Fprintf(&b, "- Songs with a recorded answer: %d\n", verdict.Known)
	fmt.Fprintf(&b, "- Answering differently now: %d (allowed: up to %.0f%%)\n", len(verdict.Different), maxDifferent*100)
	fmt.Fprintf(&b, "- New in the playlist, not recorded yet: %d\n", len(verdict.New))
	fmt.Fprintf(&b, "- Recorded but no longer in the playlist: %d\n", verdict.Gone)
	if verdict.Reason != "" {
		fmt.Fprintf(&b, "\n**%s.**\n", verdict.Reason)
	}
	if len(verdict.Different) > 0 {
		b.WriteString("\n| Song | Recorded | Now | Why (the lookup's own words) |\n|---|---|---|---|\n")
		for i, o := range verdict.Different {
			if i == maxListed {
				fmt.Fprintf(&b, "| … and %d more | | | |\n", len(verdict.Different)-maxListed)
				break
			}
			fmt.Fprintf(&b, "| %s (`%s`) | %s | %s | %s |\n", cell(o.Title), o.ID, shown(o.Expected), shown(o.Found), cell(o.Why))
		}
	}

	return b.String()
}

func shown(id string) string {
	if id == "" {
		return "none"
	}

	return "`" + id + "`"
}

// cell makes text safe inside a table cell.
func cell(text string) string {
	text = strings.NewReplacer("|", "/", "\r", " ", "\n", " ").Replace(text)
	if len(text) > 160 {
		text = text[:160] + "…"
	}

	return text
}
