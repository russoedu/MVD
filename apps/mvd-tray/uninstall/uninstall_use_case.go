package uninstall

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"youtube-downloader/apps/mvd-tray/install"
	"youtube-downloader/apps/mvd-tray/question"
	"youtube-downloader/libs/mvd-server/api"
)

// Environment is what removing the app needs from the machine, so that every branch can
// be tested on temporary folders without a screen or a real installation.
type Environment struct {
	GOOS    string
	Version string
	// Exe is the running program.
	Exe string
	// AppDir is the app-data folder.
	AppDir string
	// Keep names the folders that hold the person's own files, read when asked.
	Keep      func() []string
	Footprint func() install.Footprint

	// Ask puts a question with the given choices, best one first, to the person.
	Ask func(title, text string, choices []string) question.Answer
	// Inform tells the person how it went.
	Inform func(text string)

	Exists  func(path string) bool
	ListDir func(path string) ([]string, error)
	// RemoveAll removes a folder and what is in it; Remove one file or an empty folder.
	RemoveAll func(path string) error
	Remove    func(path string) error
	// RemoveProgram removes the running program, which on Windows can only be moved
	// aside. leftover is where something stays that the person may delete (empty if the
	// system will clean it up).
	RemoveProgram func(path string) (leftover string, err error)
	// RemoveRegistration takes the program out of Settings > Apps.
	RemoveRegistration func() error
	// Quit ends the app. It is nil when the removal runs in a process that ends anyway.
	Quit func()
}

// Service removes the app. It satisfies api.Uninstaller.
type Service struct {
	Env Environment
}

const wizardTitle = "Remove MVD"

// Confirm shows the person what is about to be removed and returns the removal if they
// agree. With deletePreferences set (the page asked for it) there is one question that
// lists everything. With it nil there are two: whether to remove the app, then whether
// the stored preferences go too. The removal never touches the downloads.
func (s Service) Confirm(deletePreferences *bool) (func(), error) {
	env := s.Env
	names, _ := env.ListDir(env.AppDir)
	build := func(deleteToo bool) removalPlan {
		return buildPlan(planInput{
			GOOS: env.GOOS, Dev: env.Version == "dev", Exe: env.Exe,
			AppDir: env.AppDir, AppDirEntries: names, DeletePreferences: deleteToo,
			Keep: env.Keep(), Footprint: env.Footprint(), Exists: env.Exists,
		})
	}

	deleteToo := deletePreferences != nil && *deletePreferences
	plan := build(deleteToo)

	switch env.Ask(wizardTitle, confirmation(plan, env.AppDir, deleteToo), []string{"Remove MVD", "Keep it"}) {
	case question.AnswerFirst:
	case question.AnswerUnavailable:
		return nil, api.ErrNoDialog
	default:
		return nil, api.ErrUninstallDeclined
	}

	if deletePreferences == nil {
		switch env.Ask(wizardTitle, preferencesQuestion(env.AppDir), []string{"Delete them too", "Keep them"}) {
		case question.AnswerFirst:
			deleteToo = true
			plan = build(true)
		case question.AnswerUnavailable:
			return nil, api.ErrNoDialog
		}
	}

	return func() { s.remove(plan, deleteToo) }, nil
}

func preferencesQuestion(appDir string) string {
	return fmt.Sprintf("Also delete your preferences?\n\nThis removes your settings, your list and the tools MVD downloaded (yt-dlp and ffmpeg), all in:\n\n%s\n\n"+
		"Your downloaded videos, and the folder they are in, are never deleted.", appDir)
}

// remove carries out the plan, going on past anything that cannot be removed so that one
// problem does not leave the rest behind, then tells the person how it went and quits.
func (s Service) remove(plan removalPlan, deletedPreferences bool) {
	env := s.Env
	var problems []string
	note := func(path string, err error) {
		problems = append(problems, fmt.Sprintf("%s: %v%s", path, err, s.hint(path, err)))
	}

	if plan.Program != "" {
		leftover, err := env.RemoveProgram(plan.Program)
		if err != nil {
			note(plan.Program, err)
		}
		if leftover != "" {
			problems = append(problems, fmt.Sprintf("%s could not be deleted while MVD was running; you can delete it yourself.", leftover))
		}
	}
	for _, folder := range plan.Folders {
		if err := env.RemoveAll(folder); err != nil {
			note(folder, err)
		}
	}
	for _, file := range plan.Files {
		if err := env.Remove(file); err != nil && !errors.Is(err, fs.ErrNotExist) {
			note(file, err)
		}
	}
	if plan.Registry {
		if err := env.RemoveRegistration(); err != nil {
			problems = append(problems, fmt.Sprintf("the entry in Settings > Apps: %v", err))
		}
	}
	for _, entry := range plan.AppDataEntries {
		if err := env.RemoveAll(entry); err != nil {
			note(entry, err)
		}
	}
	if plan.AppData != "" {
		// Only an empty folder goes: anything left in it was kept on purpose.
		_ = env.Remove(plan.AppData)
	}

	message := "MVD has been removed."
	if deletedPreferences {
		message += " Your preferences were deleted."
	} else {
		message += fmt.Sprintf(" Your preferences were kept in %s.", env.AppDir)
	}
	message += " Your downloaded videos were not touched."
	if len(plan.Kept) > 0 {
		message += "\n\n" + strings.Join(plan.Kept, "\n")
	}
	if len(problems) > 0 {
		message += "\n\nSome things could not be removed:\n" + strings.Join(problems, "\n")
	}
	env.Inform(message)

	if env.Quit != nil {
		env.Quit()
	}
}

// hint says what to do about something that could not be removed, when the reason is a
// lack of permission. The app never asks for more rights than it has.
func (s Service) hint(path string, err error) string {
	if !errors.Is(err, fs.ErrPermission) {
		return ""
	}
	switch s.Env.GOOS {
	case "darwin":
		return fmt.Sprintf(" (drag %s to the Trash instead)", path)
	case "windows":
		return " (start MVD as an administrator to remove it)"
	}

	return " (remove it with your file manager, or as the account that put it there)"
}

var _ api.Uninstaller = Service{}
