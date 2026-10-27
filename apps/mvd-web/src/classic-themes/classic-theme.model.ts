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
  'aqua' |
  'mac-os-x-10-4'

/** A colour scheme of a theme, as the system offered it: the first is the system's own. */
export interface ColourScheme {
  id:    string
  label: string
}

export interface ClassicTheme {
  id:       ClassicThemeId
  /** The name the theme picker shows. */
  label:    string
  family:   ChromeFamily
  /**
   * Its colour schemes, the same as libs/mvd-core/config's WindowSchemes; each but the first is
   * a set of colours under :root[data-scheme=...] in the theme's style.
   */
  schemes?: readonly ColourScheme[]
}

const scheme = (id: string, label: string): ColourScheme => ({ id, label })

/** The classic themes, Windows first and then the Mac, each side oldest first. */
export const classicThemes: readonly ClassicTheme[] = [
  {
    id:      'windows-3-1',
    label:   'Windows 3.1',
    family:  'windows-classic',
    schemes: [scheme('default', 'Windows Default'), scheme('hot-dog-stand', 'Hot Dog Stand'), scheme('ocean', 'Ocean'), scheme('monochrome', 'Monochrome')],
  },
  {
    id:      'windows-95',
    label:   'Windows 95',
    family:  'windows-classic',
    schemes: [scheme('standard', 'Windows Standard'), scheme('high-contrast-black', 'High Contrast Black'), scheme('high-contrast-white', 'High Contrast White'), scheme('rainy-day', 'Rainy Day'), scheme('brick', 'Brick')],
  },
  {
    id:      'windows-xp',
    label:   'Windows XP',
    family:  'windows-themed',
    schemes: [scheme('blue', 'Default (blue)'), scheme('olive-green', 'Olive Green'), scheme('silver', 'Silver')],
  },
  {
    id:      'windows-vista',
    label:   'Windows Vista',
    family:  'windows-themed',
    schemes: [scheme('aero', 'Windows Aero'), scheme('high-contrast-black', 'High Contrast Black')],
  },
  {
    id:      'mac-os-7',
    label:   'Mac OS 7.6',
    family:  'mac-classic',
    schemes: [scheme('standard', 'Standard'), scheme('black-and-white', 'Black & White'), scheme('teal', 'Teal'), scheme('rose', 'Rose')],
  },
  {
    id:      'mac-os-9',
    label:   'Mac OS 9',
    family:  'mac-classic',
    schemes: [scheme('default', 'Mac OS Default'), scheme('bondi-blue', 'Bondi Blue'), scheme('grape', 'Grape'), scheme('tangerine', 'Tangerine'), scheme('graphite', 'Graphite')],
  },
  { id: 'aqua', label: 'Mac OS X 10.0', family: 'aqua', schemes: [scheme('blue', 'Blue'), scheme('graphite', 'Graphite')] },
  { id: 'mac-os-x-10-4', label: 'Mac OS X 10.4', family: 'aqua', schemes: [scheme('blue', 'Blue'), scheme('graphite', 'Graphite')] },
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
