import { useEffect } from 'react'
import type { MenuSnapshot, WorkspaceIntent } from '../workspace-snapshot'
import { isPlainCharacter, matchesShortcut } from './shortcut-match.policy'

/** Places that use the keyboard for themselves: text fields and open menus. */
const ownsKeyboard = (target: EventTarget | null): boolean =>
  target instanceof Element && target.closest('input, textarea, select, [contenteditable], [role="menu"]') !== null

/**
 * The window's keyboard: a menu entry's shortcut chooses it, as the menu would, and a plain letter
 * goes to MVD (d accepts a song, x rejects it...), as in the terminal look. The keys are taken
 * before the lists see them, so a letter does not also jump in a list. Off while a dialog or a
 * covered screen has the keyboard.
 */
export function useKeyboardIntents (menus: MenuSnapshot[], onIntent: (intent: WorkspaceIntent) => void, enabled: boolean): void {
  useEffect(() => {
    if (!enabled) return
    const onKeyDown = (event: KeyboardEvent): void => {
      if (event.defaultPrevented || ownsKeyboard(event.target)) return
      const entry = menus.flatMap(menu => menu.entries).find(each => each.shortcut && matchesShortcut(event, each.shortcut))
      if (entry) {
        event.preventDefault()
        event.stopPropagation()
        if (entry.enabled) onIntent({ kind: 'menu', id: entry.id })

        return
      }
      if (isPlainCharacter(event)) {
        event.preventDefault()
        event.stopPropagation()
        onIntent({ kind: 'key', key: event.key })
      }
    }
    document.addEventListener('keydown', onKeyDown, { capture: true })

    return () => document.removeEventListener('keydown', onKeyDown, { capture: true })
  }, [menus, onIntent, enabled])
}
