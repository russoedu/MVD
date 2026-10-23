package config

// ThemePreset is a named set of colours, in the spirit of the themes of a code editor.
type ThemePreset struct {
	Name   string
	Colors ThemeColors
}

// DefaultThemeName is the theme a new install starts with: the palette of the logo.
const DefaultThemeName = "MVD Neon"

// CustomTheme is the name of the theme the person makes: it starts as a copy of the one that
// was selected, and every colour of it can be changed.
const CustomTheme = "Create yours"

// ThemePresets are the themes that come with the app, in the order they are offered. The
// Background of a theme is "" for the terminal's own.
var ThemePresets = []ThemePreset{
	{DefaultThemeName, DefaultThemeColors()},
	{"Dark+", ThemeColors{Accent: "#569cd6", Focus: "#4ec9b0", Highlight: "#dcdcaa", Success: "#6a9955", Error: "#f44747", Dim: "#808080", Text: "#d4d4d4", Selected: "#264f78", Background: "#1e1e1e"}},
	{"Monokai", ThemeColors{Accent: "#f92672", Focus: "#66d9ef", Highlight: "#e6db74", Success: "#a6e22e", Error: "#ff5555", Dim: "#75715e", Text: "#f8f8f2", Selected: "#49483e", Background: "#272822"}},
	{"Dracula", ThemeColors{Accent: "#ff79c6", Focus: "#8be9fd", Highlight: "#f1fa8c", Success: "#50fa7b", Error: "#ff5555", Dim: "#6272a4", Text: "#f8f8f2", Selected: "#44475a", Background: "#282a36"}},
	{"One Dark", ThemeColors{Accent: "#c678dd", Focus: "#61afef", Highlight: "#e5c07b", Success: "#98c379", Error: "#e06c75", Dim: "#5c6370", Text: "#abb2bf", Selected: "#3e4451", Background: "#282c34"}},
	{"Nord", ThemeColors{Accent: "#88c0d0", Focus: "#81a1c1", Highlight: "#ebcb8b", Success: "#a3be8c", Error: "#bf616a", Dim: "#4c566a", Text: "#d8dee9", Selected: "#3b4252", Background: "#2e3440"}},
	{"Tokyo Night", ThemeColors{Accent: "#bb9af7", Focus: "#7aa2f7", Highlight: "#e0af68", Success: "#9ece6a", Error: "#f7768e", Dim: "#565f89", Text: "#c0caf5", Selected: "#283457", Background: "#1a1b26"}},
	{"GitHub Dark", ThemeColors{Accent: "#58a6ff", Focus: "#79c0ff", Highlight: "#d29922", Success: "#3fb950", Error: "#f85149", Dim: "#8b949e", Text: "#c9d1d9", Selected: "#1c2d41", Background: "#0d1117"}},
	{"Solarized Dark", ThemeColors{Accent: "#d33682", Focus: "#2aa198", Highlight: "#b58900", Success: "#859900", Error: "#dc322f", Dim: "#586e75", Text: "#93a1a1", Selected: "#073642", Background: "#002b36"}},
	{"Solarized Light", ThemeColors{Accent: "#d33682", Focus: "#268bd2", Highlight: "#b58900", Success: "#859900", Error: "#dc322f", Dim: "#93a1a1", Text: "#586e75", Selected: "#eee8d5", Background: "#fdf6e3"}},
	{"Light+", ThemeColors{Accent: "#0000ff", Focus: "#267f99", Highlight: "#795e26", Success: "#008000", Error: "#cd3131", Dim: "#8c8c8c", Text: "#333333", Selected: "#add6ff", Background: "#ffffff"}},
}

// ThemeNames are the names of the themes offered, ending with the one the person makes.
func ThemeNames() []string {
	names := make([]string, 0, len(ThemePresets)+1)
	for _, p := range ThemePresets {
		names = append(names, p.Name)
	}
	return append(names, CustomTheme)
}

// PresetColors are the colours of the theme called name, and whether there is such a theme.
func PresetColors(name string) (ThemeColors, bool) {
	for _, p := range ThemePresets {
		if p.Name == name {
			return p.Colors, true
		}
	}
	return ThemeColors{}, false
}

// settleTheme makes ThemeName agree with the colours: a name that is a theme of the app but
// whose colours were changed is the person's own.
func settleTheme(cfg *Config) {
	if colours, ok := PresetColors(cfg.ThemeName); ok && colours != cfg.Colors {
		cfg.ThemeName = CustomTheme
	}
	if _, ok := PresetColors(cfg.ThemeName); !ok {
		cfg.ThemeName = CustomTheme
	}
}
