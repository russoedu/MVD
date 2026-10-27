import type { Shortcut } from '../workspace-snapshot'

/** The browser's names of the keys a shortcut may name another way. */
const keyNames: Record<string, string> = {
  del:   'delete',
  up:    'arrowup',
  down:  'arrowdown',
  left:  'arrowleft',
  right: 'arrowright',
  esc:   'escape',
  space: ' ',
}

interface Pressed {
  key:      string
  ctrlKey:  boolean
  metaKey:  boolean
  shiftKey: boolean
  altKey:   boolean
}

/**
 * Whether a key press is the shortcut: the same key, in either case, with the same modifiers. Ctrl
 * is ⌘ on a Mac, so either counts for it.
 */
export function matchesShortcut (pressed: Pressed, shortcut: Shortcut): boolean {
  const wanted = shortcut.key.toLowerCase()
  const key = pressed.key.toLowerCase()
  const sameKey = key === (keyNames[wanted] ?? wanted)

  return sameKey &&
    (pressed.ctrlKey || pressed.metaKey) === (shortcut.ctrl ?? false) &&
    pressed.shiftKey === (shortcut.shift ?? false) &&
    pressed.altKey === (shortcut.alt ?? false)
}

/** Whether a key press is a plain letter or digit, with no modifier but Shift. */
export function isPlainCharacter (pressed: Pressed): boolean {
  return /^[\p{L}\p{N}]$/u.test(pressed.key) && !pressed.ctrlKey && !pressed.metaKey && !pressed.altKey
}
