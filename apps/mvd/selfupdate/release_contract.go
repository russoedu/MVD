// Package selfupdate tells the person when a newer MVD is published, and (once they say yes)
// brings it in. It never downloads anything on its own and never blocks the start of the app.
package selfupdate

// Release is a published version of MVD.
type Release struct {
	// Version is the number alone, such as 0.7.0.
	Version string
	// Page is the address of the release on GitHub, where every build can be downloaded.
	Page   string
	Assets []Asset
}

// Asset is one file of a release.
type Asset struct {
	Name string
	URL  string
	// Digest is the checksum GitHub states for the file, such as sha256:ab12...; empty when none.
	Digest string
}
