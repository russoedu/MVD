package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"youtube-downloader/libs/mvd-core/config"
)

type configOutcome int

const (
	cfgNone configOutcome = iota
	cfgSave
	cfgCancel
	cfgAdvanced
)

// editor modes while a config item is being changed.
const (
	editNone = iota
	editText
	editNumber
	editRadio
	editFolder
)

const cfgItemCount = 10

var mergeOptions = []string{"mp4", "mkv", "webm"}
var cookieOptions = []string{"all", "off", "firefox", "chrome", "edge", "brave", "chromium", "opera", "vivaldi", "safari"}

// configModel is the preferences screen: a selectable list of settings, each
// edited in place by a toggle, number, text field or radio selector.
type configModel struct {
	cfg           config.Config
	cursor        int
	mode          int
	input         textinput.Model
	radioOpts     []string
	radioIdx      int
	folder        folderModel
	width, height int
}

func newConfigModel(cfg config.Config) configModel {
	ti := textinput.New()
	ti.Prompt = "> "
	return configModel{cfg: cfg, input: ti}
}

func (m configModel) init() tea.Cmd { return textinput.Blink }

func (m configModel) setSize(w, h int) configModel {
	m.width, m.height = w, h
	m.input.Width = max(10, w-8)
	m.folder = m.folder.setSize(w, h)
	return m
}

func (m configModel) update(msg tea.Msg) (configModel, tea.Cmd, configOutcome, config.Config) {
	k, isKey := msg.(tea.KeyMsg)

	if m.mode != editNone {
		cmd := m.updateEditor(msg, k, isKey)
		return m, cmd, cfgNone, m.cfg
	}

	if isKey {
		switch k.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < cfgItemCount-1 {
				m.cursor++
			}
		case "s", "ctrl+s":
			return m, nil, cfgSave, m.cfg
		case "a":
			return m, nil, cfgAdvanced, m.cfg
		case "esc":
			return m, nil, cfgCancel, m.cfg
		case "enter", " ":
			m.activate()
		}
	}
	return m, nil, cfgNone, m.cfg
}

// activate opens the editor for the selected item (or toggles a boolean).
func (m *configModel) activate() {
	switch m.cursor {
	case 6:
		m.cfg.DownloadOfficialMusicVideo = !m.cfg.DownloadOfficialMusicVideo
	case 8:
		m.cfg.CreateLogFile = !m.cfg.CreateLogFile
	case 1, 2, 3, 7:
		m.radioOpts, m.radioIdx = m.radioState()
		m.mode = editRadio
	case 5:
		m.beginInput(strconv.Itoa(m.cfg.MaxConcurrentDownloads))
		m.mode = editNumber
	case 0:
		m.folder = newFolderModel(m.cfg.OutputDir).setSize(m.width, m.height)
		m.mode = editFolder
	case 9:
		m.folder = newFolderModel(m.cfg.LogDir).setSize(m.width, m.height)
		m.mode = editFolder
	default: // 4 text
		m.beginInput(m.textValue())
		m.mode = editText
	}
}

func (m *configModel) beginInput(val string) {
	m.input.SetValue(val)
	m.input.CursorEnd()
	m.input.Focus()
}

func (m *configModel) updateEditor(msg tea.Msg, k tea.KeyMsg, isKey bool) tea.Cmd {
	if m.mode == editFolder {
		next, cmd, out := m.folder.update(msg)
		m.folder = next
		switch out {
		case folderChosen:
			if m.cursor == 0 {
				m.cfg.OutputDir = m.folder.dir
			} else {
				m.cfg.LogDir = m.folder.dir
			}
			m.mode = editNone
		case folderCancel:
			m.mode = editNone
		}
		return cmd
	}
	if isKey {
		switch k.String() {
		case "esc":
			m.mode = editNone
			return nil
		case "enter":
			m.commit()
			m.mode = editNone
			return nil
		}
		if m.mode == editRadio {
			switch k.String() {
			case "up", "k":
				if m.radioIdx > 0 {
					m.radioIdx--
				}
			case "down", "j":
				if m.radioIdx < len(m.radioOpts)-1 {
					m.radioIdx++
				}
			}
			return nil
		}
		if m.mode == editNumber {
			switch k.String() {
			case "up":
				m.input.SetValue(strconv.Itoa(m.numberValue() + 1))
				m.input.CursorEnd()
				return nil
			case "down":
				if n := m.numberValue(); n > 1 {
					m.input.SetValue(strconv.Itoa(n - 1))
					m.input.CursorEnd()
				}
				return nil
			}
		}
	}
	// This only reassigns the pointer-free copy's input; return a value model.
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return cmd
}

func (m configModel) numberValue() int {
	n, err := strconv.Atoi(strings.TrimSpace(m.input.Value()))
	if err != nil || n < 1 {
		return m.cfg.MaxConcurrentDownloads
	}
	return n
}

// commit writes the editor's value back into the config.
func (m *configModel) commit() {
	switch m.cursor {
	case 0:
		if v := strings.TrimSpace(m.input.Value()); v != "" {
			m.cfg.OutputDir = v
		}
	case 4:
		if v := strings.TrimSpace(m.input.Value()); v != "" {
			m.cfg.OutputTemplate = v
		}
	case 9:
		if v := strings.TrimSpace(m.input.Value()); v != "" {
			m.cfg.LogDir = v
		}
	case 5:
		if n, err := strconv.Atoi(strings.TrimSpace(m.input.Value())); err == nil && n > 0 {
			m.cfg.MaxConcurrentDownloads = n
		}
	case 1:
		m.cfg.VideoQuality = m.radioOpts[m.radioIdx]
	case 2:
		m.cfg.AudioQuality = m.radioOpts[m.radioIdx]
	case 3:
		m.cfg.MergeOutputFormat = m.radioOpts[m.radioIdx]
	case 7:
		applyCookieChoice(&m.cfg, m.radioOpts[m.radioIdx])
	}
}

func applyCookieChoice(cfg *config.Config, opt string) {
	switch opt {
	case "all":
		cfg.AutoCookies, cfg.CookiesFromBrowser = true, ""
	case "off":
		cfg.AutoCookies, cfg.CookiesFromBrowser = false, ""
	default:
		cfg.AutoCookies, cfg.CookiesFromBrowser = false, opt
	}
}

func (m configModel) textValue() string {
	switch m.cursor {
	case 0:
		return m.cfg.OutputDir
	case 4:
		return m.cfg.OutputTemplate
	case 9:
		return m.cfg.LogDir
	}
	return ""
}

// radioState returns the options and the current index for the selected item.
func (m configModel) radioState() ([]string, int) {
	switch m.cursor {
	case 1:
		return config.VideoPresets, indexOf(config.VideoPresets, m.cfg.VideoQuality)
	case 2:
		return config.AudioPresets, indexOf(config.AudioPresets, m.cfg.AudioQuality)
	case 3:
		return mergeOptions, indexOf(mergeOptions, m.cfg.MergeOutputFormat)
	case 7:
		return cookieOptions, indexOf(cookieOptions, currentCookieChoice(m.cfg))
	}
	return nil, 0
}

func currentCookieChoice(cfg config.Config) string {
	switch {
	case cfg.CookiesFromBrowser != "":
		return cfg.CookiesFromBrowser
	case cfg.AutoCookies:
		return "all"
	default:
		return "off"
	}
}

// --- view -------------------------------------------------------------------

var cfgLabels = []string{
	"Output Folder", "Video Quality", "Audio Quality", "Merge Format",
	"Output Template", "Max Concurrent Downloads", "Download Official Music Video",
	"Cookies from Browser", "Create Log File", "Log File Location",
}

func (m configModel) display(i int) string {
	c := m.cfg
	switch i {
	case 0:
		return c.OutputDir
	case 1:
		return c.VideoQuality
	case 2:
		return c.AudioQuality
	case 3:
		return c.MergeOutputFormat
	case 4:
		return c.OutputTemplate
	case 5:
		return strconv.Itoa(c.MaxConcurrentDownloads)
	case 6:
		return yesNo(c.DownloadOfficialMusicVideo)
	case 7:
		return cookieLabel(c)
	case 8:
		return yesNo(c.CreateLogFile)
	case 9:
		return c.LogDir
	}
	return ""
}

func (m configModel) view(width, height int) string {
	if m.mode == editFolder {
		return m.folder.view(width, height)
	}
	var rows []string
	labelW := 0
	for _, l := range cfgLabels {
		if len(l) > labelW {
			labelW = len(l)
		}
	}
	for i, label := range cfgLabels {
		cursor := "  "
		if i == m.cursor {
			cursor = styMagenta.Render("▸ ")
		}
		name := padRight(label, labelW)
		val := m.display(i)
		if i == m.cursor && m.mode != editNone {
			val = m.editorView()
		} else if i == m.cursor {
			name = stySelected.Render(name)
		}
		rows = append(rows, fmt.Sprintf("%s%s  %s", cursor, name, styCyan.Render(truncate(val, max(10, width-labelW-6)))))
	}

	body := strings.Join(rows, "\n")
	if m.cursor == 4 {
		body += "\n\n" + styDim.Render("Template uses yt-dlp fields, e.g. %(playlist_title)s/%(playlist_index)02d - %(title)s.%(ext)s")
	}

	return screenFrame(width, height, "MVD · Preferences", body, keyBar(width, m.hints()))
}

func (m configModel) hints() []keyHint {
	if m.mode != editNone {
		return []keyHint{{"enter", "apply"}, {"esc", "cancel"}}
	}
	return []keyHint{{"↑↓", "move"}, {"enter", "edit"}, {"a", "advanced"}, {"s", "save"}, {"esc", "cancel"}}
}

func (m configModel) editorView() string {
	switch m.mode {
	case editRadio:
		var parts []string
		for i, opt := range m.radioOpts {
			if i == m.radioIdx {
				parts = append(parts, styYellow.Render("("+opt+")"))
			} else {
				parts = append(parts, styDim.Render(opt))
			}
		}
		return strings.Join(parts, " ")
	default:
		return m.input.View()
	}
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func cookieLabel(cfg config.Config) string {
	switch {
	case cfg.CookiesFromBrowser != "":
		return cfg.CookiesFromBrowser
	case cfg.AutoCookies:
		return "all (try every browser)"
	default:
		return "off"
	}
}

func indexOf(opts []string, v string) int {
	for i, o := range opts {
		if o == v {
			return i
		}
	}
	return 0
}
