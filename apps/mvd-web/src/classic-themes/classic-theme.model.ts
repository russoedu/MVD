/**
 * The chrome families: themes in one family share the structure of their window frame, menus,
 * dialogs and controls, and differ in colours, sizes and fonts.
 */
export type ChromeFamily = 'windows-classic' | 'windows-themed' | 'aqua'

export type ClassicThemeId = 'windows-95' | 'windows-vista' | 'aqua'

export interface ClassicTheme {
  id:     ClassicThemeId
  /** The name the theme picker shows. */
  label:  string
  family: ChromeFamily
}

/**
 * The classic themes built so far: the ones whose fonts are free to use (#369). Windows 3.1,
 * XP and Mac OS 7 and 9 follow once their fonts are cleared (see docs/architecture/classic-themes-plan.md).
 */
export const classicThemes: readonly ClassicTheme[] = [
  { id: 'windows-95', label: 'Windows 95', family: 'windows-classic' },
  { id: 'windows-vista', label: 'Windows Vista', family: 'windows-themed' },
  { id: 'aqua', label: 'Mac OS X 10.0', family: 'aqua' },
]

/** The theme with this id, or the first one when the id names none. */
export function classicThemeOf (id: string | null | undefined): ClassicTheme {
  return classicThemes.find(theme => theme.id === id) ?? classicThemes[0]
}
