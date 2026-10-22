import type { PressedKey } from './browser-shortcut.policy'

const clipboardKeys = new Set(['c', 'v', 'x'])
const letter = /^[a-z]$/i
const punctuation = /^[,.;/\-=[\]\']$/

/**
 * The bytes to send the program for a chord that a terminal cannot spell with a control
 * character: Ctrl+Shift+S, Ctrl+, and every Cmd chord on a Mac, where Cmd takes the place
 * of Ctrl. It is the "CSI u" form (the key's code and its modifiers), which the program
 * reads as the same chord; null when xterm already sends the chord or the page leaves it
 * to the browser.
 */
export function chordSequence (pressed: PressedKey): string | null {
  if (pressed.altKey || (!pressed.ctrlKey && !pressed.metaKey)) return null

  const key = pressed.key.toLowerCase()
  const isLetter = letter.test(key)
  if (!isLetter && !punctuation.test(key)) return null
  if (isLetter && !pressed.shiftKey && !pressed.metaKey) return null // Ctrl+letter is a control character
  if (clipboardKeys.has(key) && !pressed.shiftKey) return null // the browser keeps copy, paste and cut

  const shiftBit = pressed.shiftKey && isLetter ? 1 : 0
  const ctrlBit = 4

  return `\u001B[${key.codePointAt(0)};${1 + shiftBit + ctrlBit}u`
}
