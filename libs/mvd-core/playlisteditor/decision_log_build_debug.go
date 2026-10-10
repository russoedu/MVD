//go:build mvddebug

package playlisteditor

// DebugBuild is true in a local build made for testing (go build -tags mvddebug). Only
// there does the editor ask why a decision was changed and write the answer to a file.
// The published builds are made without the tag, so the option does not exist in them.
const DebugBuild = true
