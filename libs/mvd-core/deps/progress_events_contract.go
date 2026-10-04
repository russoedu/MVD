package deps

// EventKind says what an Event reports.
type EventKind int

const (
	// EventMissing: some tools are not installed. Names lists them and Dir is where
	// they are about to be put.
	EventMissing EventKind = iota
	// EventDownloading: Name is being fetched from URL.
	EventDownloading
	// EventInstalled: Name is ready.
	EventInstalled
	// EventFailed: Name could not be installed; Err says why. Installing carries on
	// with the next tool.
	EventFailed
	// EventUpToDate: Name was checked and is the newest build there is, or a newer one
	// is not new enough yet to be worth fetching. Detail says which.
	EventUpToDate
	// EventUpdated: Name was replaced by a newer build. Detail is its version.
	EventUpdated
	// EventUpdateFailed: Name could not be checked or updated (no connection, GitHub
	// limiting requests, a bad download). The copy that was there is untouched.
	EventUpdateFailed
)

// Event is one step of making the tools available or keeping them current.
type Event struct {
	Kind   EventKind
	Name   string
	Names  []string
	URL    string
	Dir    string
	Detail string
	Err    error
}

// Reporter receives the steps as they happen, on the goroutine doing the work.
type Reporter func(Event)
