package urllist

// Begin asks to start the downloads; an empty list has nothing to start.
func (l *List) Begin() Outcome {
	if len(l.URLs()) == 0 {
		return None
	}

	return Start
}

// ReviewLinks asks to open the editor on the links; an empty list has nothing to review.
func (l *List) ReviewLinks() Outcome {
	if len(l.URLs()) == 0 {
		return None
	}

	return Review
}

// OpenPreferences asks for the preferences.
func (*List) OpenPreferences() Outcome { return Prefs }

// OpenSongs asks for the screen to paste songs.
func (*List) OpenSongs() Outcome { return Songs }

// Reset empties the list.
func (l *List) Reset() { l.Text = "" }

// Leave asks to quit: a list with links asks first, an empty one quits at once.
func (l *List) Leave() Outcome {
	if len(l.URLs()) > 0 {
		l.ConfirmQuit = true

		return None
	}

	return Quit
}

// Answer is the person's reply to the question to discard the list: yes quits, anything else keeps
// the list.
func (l *List) Answer(yes bool) Outcome {
	if yes {
		return Quit
	}
	l.ConfirmQuit = false

	return None
}
