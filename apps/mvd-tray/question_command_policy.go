package main

import (
	"strconv"
	"strings"

	"youtube-downloader/apps/mvd-tray/oscommand"
)

// questionCommand picks the way to put a question with 2 or 3 choices to the person on
// goos (macOS and Linux; Windows has the message box), using only programs that has
// reports as installed. The first choice is the one to prefer and the last is the one
// that means "leave it". ok is false when there is no way to ask.
//
// The words travel as arguments, never inside a script.
func questionCommand(goos, title, text string, choices []string, has func(string) bool) (cmd oscommand.Command, ok bool) {
	if len(choices) < 2 || len(choices) > 3 {
		return oscommand.Command{}, false
	}

	switch goos {
	case "darwin":
		if !has("osascript") {
			return oscommand.Command{}, false
		}

		return oscommand.Command{Name: "osascript", Args: macQuestionArguments(title, text, choices)}, true
	case "linux":
		if len(choices) != 2 {
			return oscommand.Command{}, false
		}
		switch {
		case has("zenity"):
			return oscommand.Command{Name: "zenity", Args: []string{
				"--question", "--no-markup", "--title=" + title, "--text=" + text,
				"--ok-label=" + choices[0], "--cancel-label=" + choices[1],
			}}, true
		case has("kdialog"):
			return oscommand.Command{Name: "kdialog", Args: []string{
				"--title", title, "--yes-label", choices[0], "--no-label", choices[1], "--yesno", text,
			}}, true
		}
	}

	return oscommand.Command{}, false
}

// macQuestionArguments builds the osascript command line. AppleScript lays the buttons
// out in the order given and makes the last the default, so the choices are passed in
// reverse: the preferred one lands on the right, as the default. The title, the text and
// the button labels are arguments (argv items 1 and 2, then 3 onwards), not script text.
func macQuestionArguments(title, text string, choices []string) []string {
	buttons := make([]string, len(choices))
	for i := range choices {
		buttons[i] = "item " + strconv.Itoa(3+i) + " of argv"
	}
	preferred := "item " + strconv.Itoa(2+len(choices)) + " of argv"

	args := []string{
		"-e", "on run argv",
		"-e", "set theReply to display dialog (item 1 of argv) with title (item 2 of argv) buttons {" + strings.Join(buttons, ", ") + "} default button (" + preferred + ")",
		"-e", "return button returned of theReply",
		"-e", "end run",
		text, title,
	}
	for i := len(choices) - 1; i >= 0; i-- {
		args = append(args, choices[i])
	}

	return args
}

// parseAnswer turns what the question program did into the person's choice. On macOS the
// program prints the label of the button that was pressed and fails if the box was
// cancelled. On Linux the exit status is the answer: 0 for the first choice, 1 for the
// second, and anything else means the box could not be shown.
func parseAnswer(goos string, choices []string, exitCode int, stdout string) answer {
	switch goos {
	case "darwin":
		if exitCode != 0 {
			return answerLeave
		}
		label := strings.TrimSpace(stdout)
		switch {
		case len(choices) > 0 && label == choices[0]:
			return answerFirst
		case len(choices) == 3 && label == choices[1]:
			return answerSecond
		}

		return answerLeave
	case "linux":
		switch exitCode {
		case 0:
			return answerFirst
		case 1:
			return answerLeave
		}
	}

	return answerUnavailable
}
