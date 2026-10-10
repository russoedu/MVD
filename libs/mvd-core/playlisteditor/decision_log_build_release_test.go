//go:build !mvddebug

package playlisteditor

import "testing"

// Published builds are made without the mvddebug tag: the question about why a decision was
// changed must not exist in them.
func TestPublishedBuildsDoNotAskWhyADecisionWasChanged(t *testing.T) {
	if DebugBuild {
		t.Fatal("DebugBuild is on in a build made without the mvddebug tag")
	}
}
