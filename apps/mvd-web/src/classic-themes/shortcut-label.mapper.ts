import type { Shortcut } from '../workspace-snapshot'
import type { ChromeLayout } from './chrome-layout.policy'

const macKeys: Record<string, string> = {
  Up:    '↑',
  Down:  '↓',
  Left:  '←',
  Right: '→',
  Del:   '⌦',
  Enter: '↩',
  Esc:   '⎋',
  Tab:   '⇥',
}

/**
 * A shortcut as the family writes it: "Ctrl+Shift+S" on Windows, "⇧⌘S" on a Mac, where the
 * modifiers come first in the Mac's own order (⌥ ⇧ ⌘). Ctrl is shown as ⌘ on a Mac.
 */
export function shortcutLabel (shortcut: Shortcut, style: ChromeLayout['shortcutStyle']): string {
  if (style === 'windows') {
    const parts = [shortcut.ctrl && 'Ctrl', shortcut.alt && 'Alt', shortcut.shift && 'Shift', shortcut.key]

    return parts.filter(Boolean).join('+')
  }
  const key = macKeys[shortcut.key] ?? shortcut.key

  return `${shortcut.alt ? '⌥' : ''}${shortcut.shift ? '⇧' : ''}${shortcut.ctrl ? '⌘' : ''}${key}`
}
