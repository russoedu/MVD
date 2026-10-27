// Package runcontrol is renderer-neutral: it decides what the person can do with the list given
// what the run is doing. Whether a run has finished, whether it only holds the list, whether the
// list can be edited or links added, how many songs would be lost by closing, which entries of the Run and Edit menus are enabled,
// and why the downloads cannot begin. It is made from facts
// the workspace gathers, so any renderer asks the same questions and gets the same answers.
package runcontrol
