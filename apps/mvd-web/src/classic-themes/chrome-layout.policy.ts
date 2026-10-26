import type { ButtonSnapshot } from '../workspace-snapshot'
import type { ChromeFamily } from './classic-theme.model'

export type WindowControl = 'minimise' | 'maximise' | 'close'

/** Where a family puts things, and how its dialogs and shortcuts read. */
export interface ChromeLayout {
  /** Where the window's title sits in its title bar. */
  titleAlign:       'left' | 'center'
  /** Which side of the title bar holds the window's buttons, and their order from left to right. */
  controlsSide:     'left' | 'right'
  controls:         readonly WindowControl[]
  /** What a double click on the title bar does: Windows maximises, Mac OS X 10.0 minimises. */
  titleDoubleClick: WindowControl
  /** Windows puts the default button first ([OK] [Cancel]); a Mac puts it last ([Cancel] [OK]). */
  defaultButton:    'first' | 'last'
  /** How shortcuts are written: Ctrl+L, or ⌘L. */
  shortcutStyle:    'windows' | 'mac'
  /** A dialog is a window of its own, or a sheet that comes down from the window's title bar. */
  dialogStyle:      'window' | 'sheet'
  /**
   * When the keyboard letters are underlined: always (Windows 95), once Alt is pressed (Windows
   * 2000 onwards), or never, since a Mac has none.
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
  controlsSide:     'right',
  controls:         ['minimise', 'maximise', 'close'],
  titleDoubleClick: 'maximise',
  defaultButton:    'first',
  shortcutStyle:    'windows',
  dialogStyle:      'window',
}

const layouts: Record<ChromeFamily, ChromeLayout> = {
  'windows-classic': { ...windows, mnemonics: 'always', pixelSnap: true },
  'windows-themed':  { ...windows, mnemonics: 'on-alt', pixelSnap: false },
  'aqua':            {
    titleAlign:       'center',
    controlsSide:     'left',
    controls:         ['close', 'minimise', 'maximise'],
    titleDoubleClick: 'minimise',
    defaultButton:    'last',
    shortcutStyle:    'mac',
    dialogStyle:      'sheet',
    mnemonics:        'none',
    pixelSnap:        false,
  },
}

/** The layout of a chrome family. */
export function layoutOf (family: ChromeFamily): ChromeLayout {
  return layouts[family]
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
