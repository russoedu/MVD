package question

import "golang.org/x/sys/windows"

// What a message box returns for the button that was pressed.
const (
	idCancel = 2
	idYes    = 6
	idNo     = 7
)

// Ask puts a question to the person in a message box and returns their choice.
// With two choices it is Yes and No; with three it is Yes, No and Cancel, and the box
// says which button means what. Closing the box counts as the last choice, leaving
// things as they are.
func Ask(title, question string, choices []string) Answer {
	flags := uint32(windows.MB_YESNO)
	text := question
	if len(choices) == 3 {
		flags = windows.MB_YESNOCANCEL
		text += "\n\nYes: " + choices[0] + "      No: " + choices[1] + "      Cancel: " + choices[2]
	}
	body, err := windows.UTF16PtrFromString(text)
	if err != nil {
		return AnswerLeave
	}
	caption, err := windows.UTF16PtrFromString(title)
	if err != nil {
		return AnswerLeave
	}

	pressed, _ := windows.MessageBox(0, body, caption, flags|windows.MB_ICONQUESTION|windows.MB_SETFOREGROUND)
	switch pressed {
	case idYes:
		return AnswerFirst
	case idNo:
		if len(choices) == 3 {
			return AnswerSecond
		}
	}

	return AnswerLeave
}
