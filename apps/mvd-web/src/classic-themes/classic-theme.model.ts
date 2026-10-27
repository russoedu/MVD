/**
 * The chrome families: themes in one family share the structure of their window frame, menus,
 * dialogs and controls, and differ in colours, sizes and fonts.
 */
export type ChromeFamily = 'windows-classic' | 'windows-themed' | 'mac-classic' | 'aqua' | 'dos'

export type ClassicThemeId =
  | 'dos' |
  'windows-3-1' |
  'windows-95' |
  'windows-xp' |
  'windows-vista' |
  'mac-os-7' |
  'mac-os-9' |
  'aqua'

export interface ClassicTheme {
  id:     ClassicThemeId
  /** The name the theme picker shows. */
  label:  string
  family: ChromeFamily
}

/** The classic themes, Windows first and then the Mac, each side oldest first. */
export const classicThemes: readonly ClassicTheme[] = [
  { id: 'windows-3-1', label: 'Windows 3.1', family: 'windows-classic' },
  { id: 'windows-95', label: 'Windows 95', family: 'windows-classic' },
  { id: 'windows-xp', label: 'Windows XP', family: 'windows-themed' },
  { id: 'windows-vista', label: 'Windows Vista', family: 'windows-themed' },
  { id: 'mac-os-7', label: 'Mac OS 7', family: 'mac-classic' },
  { id: 'mac-os-9', label: 'Mac OS 9', family: 'mac-classic' },
  { id: 'aqua', label: 'Mac OS X 10.0', family: 'aqua' },
]

/**
 * MVD's own terminal look, for the windows the page draws over the terminal (the preferences): a
 * box-drawn frame in the colours of the terminal's theme. The main window stays the terminal.
 */
export const dosTheme: ClassicTheme = { id: 'dos', label: 'MVD terminal (DOS)', family: 'dos' }

/** The theme of a window style as the settings name it: the terminal look, or a classic one. */
export function windowThemeOf (style: string): ClassicTheme {
  return style === 'dos' ? dosTheme : classicThemeOf(style)
}

/** The theme with this id, or Windows 95 when the id names none. */
export function classicThemeOf (id: string | null | undefined): ClassicTheme {
  return classicThemes.find(theme => theme.id === id) ?? classicThemes[1]
}
