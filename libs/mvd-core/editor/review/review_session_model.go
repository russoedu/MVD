package review

import (
	"time"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/playlisteditor"
	"youtube-downloader/libs/mvd-core/playlistfile"
	"youtube-downloader/libs/mvd-core/runstate"
)

// Mode is what the editor is doing with what the person gives it.
type Mode int

const (
	Browsing Mode = iota
	// AskPrevious asks whether to bring the review that was saved last time.
	AskPrevious
	// AskDownload asks what to do with the songs nobody has looked at.
	AskDownload
	// Input is a one-line question: an address, a file, or a reason.
	Input
)

// inputPurpose is what a one-line question is for.
type inputPurpose int

const (
	inputReplace inputPurpose = iota
	inputSave
	inputLoad
	inputNote
)

// noticeFor is how long a notice stays on the bottom line.
const noticeFor = 4 * time.Second

// Session is the playlist editor: the songs of a plan, what was proposed for each, and what the
// person decides about them, before anything is downloaded. It is a value; a renderer reads its
// exported fields and changes it only through its methods.
type Session struct {
	// Rows are the songs, with what was proposed and decided for each.
	Rows []playlistfile.Entry
	// Selected is the index of the song the cursor is on.
	Selected int
	// States are the states of the songs of a live review, parallel to Rows.
	States []engine.EntryState

	// Planning is true while a plan-only run is still working; PlanState is that run's.
	Planning  bool
	PlanState *runstate.State

	Mode Mode
	// Prompt and InputText are the one-line question that is open, and what has been typed.
	Prompt, InputText string
	// Picking is true while the system's file window is up.
	Picking bool
	// Asks counts the one-line questions opened, so a renderer can tell a new one from the one
	// it already shows.
	Asks int
	// Saved is the review of last time, offered while planning.
	Saved []playlistfile.Entry

	// Version grows with every decision of a live review, so a host can tell when to save them.
	Version int

	// Notice is the message on the bottom line, until NoticeUntil.
	Notice      string
	NoticeUntil time.Time

	opts    Options
	live    Live
	planner Planner
	// decisions are what the person decided on a live review, by song, kept from one refresh to
	// the next (a live review rebuilds its rows from the run).
	decisions map[string]playlistfile.Entry
	// previous is what is brought along when planning ends.
	previous []playlistfile.Entry
	planDone bool // the run has finished while a question was waiting
	purpose  inputPurpose
	// noteFor is the row a reason is being asked for.
	noteFor string
}

// NewFromRows opens the editor on rows that are already planned.
func NewFromRows(opts Options, rows []playlistfile.Entry) Session {
	return Session{opts: opts, Rows: rows}
}

// NewLive opens the editor on a run that holds its songs for review. brought are the decisions
// kept from an earlier session, which are put on the songs they belong to.
func NewLive(opts Options, live Live, brought []playlistfile.Entry) Session {
	s := Session{opts: opts, live: live, decisions: map[string]playlistfile.Entry{}}
	for _, b := range brought {
		if b.Decision != playlistfile.DecisionNone || b.Downloaded {
			s.decisions[b.Key()] = b
		}
	}
	s.Refresh()

	return s
}

// NewPlanning opens the editor while a plan-only run works. brought is what the person asked to
// bring from files; saved is the review of last time, which they are asked about when there is
// nothing else to bring.
func NewPlanning(opts Options, run PlanRun, brought, saved []playlistfile.Entry) Session {
	s := Session{
		opts:      opts,
		Planning:  true,
		PlanState: runstate.New(run.Sources()),
		previous:  brought,
	}
	s.planner, _ = run.(Planner)
	if len(brought) == 0 && len(saved) > 0 {
		s.Saved = saved
		s.Mode = AskPrevious
	}

	return s
}

// IsLive reports whether the editor shows a run as it goes.
func (s Session) IsLive() bool { return s.live != nil }

// Current is the song the cursor is on.
func (s Session) Current() (playlistfile.Entry, bool) {
	if s.Selected < 0 || s.Selected >= len(s.Rows) {
		return playlistfile.Entry{}, false
	}

	return s.Rows[s.Selected], true
}

// Flash shows text on the bottom line for a moment.
func (s *Session) Flash(text string) {
	s.Notice = text
	s.NoticeUntil = time.Now().Add(noticeFor)
}

// NoticeShown is the message on the bottom line, when it is still to be shown at now.
func (s Session) NoticeShown(now time.Time) (string, bool) {
	if s.Notice != "" && !s.NoticeUntil.IsZero() && now.Before(s.NoticeUntil) {
		return s.Notice, true
	}

	return "", false
}

// autosave keeps the review where the next session can find it.
func (s Session) autosave() {
	if s.opts.SessionFile == "" || s.Planning {
		return
	}
	_ = playlistfile.Save(s.opts.SessionFile, playlistfile.FromSongs(s.Rows))
}

// Unreviewed is how many songs nobody has looked at.
func (s Session) Unreviewed() int { return playlisteditor.Unreviewed(s.Rows) }
