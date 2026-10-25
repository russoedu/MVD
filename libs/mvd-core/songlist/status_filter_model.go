package songlist

import "youtube-downloader/libs/mvd-core/runstate"

// Filter is which groups of statuses the person hid. It is a value: a change gives a new filter,
// so a copy of a screen keeps its own.
type Filter struct{ hidden map[string]bool }

// HiddenFilter is a filter that hides the given groups.
func HiddenFilter(groups ...string) Filter {
	hidden := make(map[string]bool, len(groups))
	for _, g := range groups {
		hidden[g] = true
	}

	return Filter{hidden: hidden}
}

// Active says whether some of the statuses are hidden.
func (f Filter) Active() bool { return len(f.hidden) > 0 }

// HiddenCount is how many groups of statuses are hidden.
func (f Filter) HiddenCount() int { return len(f.hidden) }

// Hidden says whether the songs of a group are hidden.
func (f Filter) Hidden(group string) bool { return f.hidden[group] }

// Matches says whether a song passes the filter.
func (f Filter) Matches(en *runstate.Entry) bool { return !f.hidden[GroupOf(en)] }

// Toggled shows or hides the songs of a group of statuses.
func (f Filter) Toggled(group string) Filter {
	hidden := make(map[string]bool, len(f.hidden)+1)
	for k, v := range f.hidden {
		hidden[k] = v
	}
	if hidden[group] {
		delete(hidden, group)
	} else {
		hidden[group] = true
	}

	return Filter{hidden: hidden}
}

// Only shows the songs of one group of statuses and hides the rest; when that is what the filter
// already is, it shows them all.
func (f Filter) Only(group string) Filter {
	hidden := map[string]bool{}
	for _, g := range Groups {
		if g.Key != group {
			hidden[g.Key] = true
		}
	}
	if len(f.hidden) == len(hidden) && !f.hidden[group] {
		return Filter{}
	}

	return Filter{hidden: hidden}
}
