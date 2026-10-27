package windowsnapshot

// The questions and notes MVD shows over its window: the question before quitting while songs
// download, a question with a line to write in (open, save as, a link), About MVD and the keyboard
// shortcuts. A theme draws each as its system's dialog.

// The questions, by what they are.
const (
	QuestionQuit      = "quit"
	QuestionAsk       = "ask"
	QuestionAbout     = "about"
	QuestionShortcuts = "shortcuts"
)

// QuestionTone is the mark a dialog wears: a caution for a choice that loses something, a note
// for what is only said.
type QuestionTone string

// The tones.
const (
	ToneCaution QuestionTone = "caution"
	ToneNote    QuestionTone = "note"
	ToneAsk     QuestionTone = "ask"
)

// QuestionRow is a line of a table in a dialog: the keys and what they do.
type QuestionRow struct {
	Keys string `json:"keys"`
	Does string `json:"does"`
}

// QuestionSection is a heading of the table and its rows.
type QuestionSection struct {
	Heading string        `json:"heading"`
	Rows    []QuestionRow `json:"rows"`
}

// QuestionField is the line a question asks to write in, with what is written in it so far.
type QuestionField struct {
	Label string `json:"label"`
	Text  string `json:"text"`
}

// Question is a dialog over the window. A button intent sends back the id of the button pressed;
// a text intent what is written in the field.
type Question struct {
	ID       string            `json:"id"`
	Title    string            `json:"title"`
	Tone     QuestionTone      `json:"tone"`
	Lines    []string          `json:"lines,omitempty"`
	Sections []QuestionSection `json:"sections,omitempty"`
	Field    *QuestionField    `json:"field,omitempty"`
	Buttons  []Button          `json:"buttons"`
}

// The intents of a question.
const (
	IntentQuestionButton IntentKind = "question-button"
	IntentQuestionText   IntentKind = "question-text"
)
