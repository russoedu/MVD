import type { ButtonSnapshot } from '../workspace-snapshot'
import type { ChromeFamily, ClassicTheme, ClassicThemeId } from './classic-theme.model'

export type WindowControl = 'minimise' | 'maximise' | 'close'

/** Where a theme puts things, and how its dialogs, menus and shortcuts behave. */
export interface ChromeLayout {
  /** Where the window's title sits in its title bar. */
  titleAlign:       'left' | 'center'
  /**
   * The window's buttons on each side of the title, left to right. Windows 3.1 has its
   * control-menu box on the left (it closes the window here); a classic Mac has its close box on
   * the left and its zoom (and, from Mac OS 8, collapse) box on the right.
   */
  controlsLeft:     readonly WindowControl[]
  controlsRight:    readonly WindowControl[]
  /** What a double click on the title bar does: Windows maximises, a Mac minimises or collapses, Mac OS 7.0 nothing. */
  titleDoubleClick: WindowControl | null
  /** Windows puts the default button first ([OK] [Cancel]); a Mac puts it last ([Cancel] [OK]). */
  defaultButton:    'first' | 'last'
  /** How shortcuts are written: Ctrl+L, or ⌘L. */
  shortcutStyle:    'windows' | 'mac'
  /**
   * A dialog is a window of its own with a title bar (Windows), a plain framed box with no title
   * bar (a classic Mac's modal dialog), or a sheet that comes down from the title bar (Mac OS X).
   */
  dialogStyle:      'window' | 'plain' | 'sheet'
  /**
   * How a menu stays open: until a choice or a click elsewhere, or only while the mouse button
   * is held down, as on Mac OS 7.
   */
  menuStyle:        'sticky' | 'hold'
  /**
   * When the keyboard letters are underlined: always (Windows 3.1 and 95), once Alt is pressed
   * (Windows 2000 onwards), or never, since a Mac has none.
   */
  mnemonics:        'always' | 'on-alt' | 'none'
  /**
   * Whether the window is scaled by whole screen pixels only, so a bitmap font and one-pixel
   * bevels stay sharp on a screen scaled by 125 % or 150 %.
   */
  pixelSnap:        boolean
}

const windows: Omit<ChromeLayout, 'mnemonics' | 'pixelSnap'> = {
  titleAlign:       'left',
  controlsLeft:     [],
  controlsRight:    ['minimise', 'maximise', 'close'],
  titleDoubleClick: 'maximise',
  defaultButton:    'first',
  shortcutStyle:    'windows',
  dialogStyle:      'window',
  menuStyle:        'sticky',
}

const mac: Omit<ChromeLayout, 'controlsLeft' | 'controlsRight' | 'titleDoubleClick' | 'dialogStyle' | 'menuStyle' | 'pixelSnap'> = {
  titleAlign:    'center',
  defaultButton: 'last',
  shortcutStyle: 'mac',
  mnemonics:     'none',
}

const families: Record<ChromeFamily, ChromeLayout> = {
  'windows-classic': { ...windows, mnemonics: 'always', pixelSnap: true },
  'windows-themed':  { ...windows, mnemonics: 'on-alt', pixelSnap: false },
  'mac-classic':     {
    ...mac,
    controlsLeft:     ['close'],
    controlsRight:    ['maximise', 'minimise'],
    titleDoubleClick: 'minimise',
    dialogStyle:      'plain',
    menuStyle:        'sticky',
    pixelSnap:        true,
  },
  // MVD's terminal look: a frame of lines, the title on its top edge and a close mark at the right.
  'dos': {
    ...windows,
    titleAlign:    'center',
    controlsRight: ['close'],
    mnemonics:     'none',
    pixelSnap:     false,
  },
  'aqua': {
    ...mac,
    controlsLeft:     ['close', 'minimise', 'maximise'],
    controlsRight:    [],
    titleDoubleClick: 'minimise',
    dialogStyle:      'sheet',
    menuStyle:        'sticky',
    pixelSnap:        false,
  },
}

/** Where a theme differs from the rest of its family. */
const themes: Partial<Record<ClassicThemeId, Partial<ChromeLayout>>> = {
  // A centred title, the control-menu box on the left and ▼ ▲ on the right; no close button.
  'windows-3-1': { titleAlign: 'center', controlsLeft: ['close'], controlsRight: ['minimise', 'maximise'] },
  // Mac OS 7.0: no collapse box yet (WindowShade came with 7.5), and menus open only while held.
  'mac-os-7':    { controlsRight: ['maximise'], titleDoubleClick: null, menuStyle: 'hold' },
}

/** The layout of a theme: its family's, with what the theme does its own way. */
export function layoutOf (theme: Pick<ClassicTheme, 'id' | 'family'>): ChromeLayout {
  return { ...families[theme.family], ...themes[theme.id] }
}

/**
 * The buttons of a dialog in the order the family places them: on Windows the default button,
 * then the others, then Cancel; on a Mac the others, then Cancel, then the default button.
 */
export function orderDialogButtons (buttons: readonly ButtonSnapshot[], layout: ChromeLayout): ButtonSnapshot[] {
  const byRole = (role: ButtonSnapshot['role']): ButtonSnapshot[] => buttons.filter(button => button.role === role)
  const others = byRole('other')
  if (layout.defaultButton === 'first') return [...byRole('default'), ...others, ...byRole('cancel')]

  return [...others, ...byRole('cancel'), ...byRole('default')]
}
