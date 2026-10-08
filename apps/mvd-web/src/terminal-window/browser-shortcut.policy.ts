/** The parts of a key press the policy looks at. */
export interface PressedKey {
  key:      string
  ctrlKey:  boolean
  metaKey:  boolean
  altKey:   boolean
  shiftKey: boolean
}

/** Copy, paste and cut keep working on the page, so the browser keeps them. */
const clipboardKeys = new Set(['c', 'v', 'x'])
const modifierKeys = new Set(['control', 'meta'])

/**
 * Whether the page takes this key press over from the browser: a Ctrl or Cmd
 * chord such as Ctrl+P (print) or Ctrl+S (save page), which the terminal
 * interface uses itself. Chords with Shift or Alt (the developer tools,
 * Ctrl+Shift+M that hands focus back to the page) stay with the browser.
 */
export function takesOverBrowserShortcut (pressed: PressedKey): boolean {
  if (!pressed.ctrlKey && !pressed.metaKey) return false
  if (pressed.shiftKey || pressed.altKey) return false

  const key = pressed.key.toLowerCase()

  return !modifierKeys.has(key) && !clipboardKeys.has(key)
}
