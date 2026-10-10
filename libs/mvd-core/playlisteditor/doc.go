// Package playlisteditor holds what a person does to a plan before it is downloaded:
// turning a plan into rows, deciding what each row should download, keeping the
// decisions across runs and files, and the optional log of why a decision was changed
// (only in a local build made for testing: see DebugBuild).
//
// The rows are playlistfile entries, so what the editor holds is what a .mvd file saves.
package playlisteditor
