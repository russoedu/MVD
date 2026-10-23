import type { PressedKey } from './browser-shortcut.policy'

const clipboardKeys = new Set(['c', 'v', 'x'])
/** Letters whose Shift chords open the browser's developer tools. */
const developerToolsKeys = new Set(['c', 'i', 'j'])
const letter = /^[a-z]$/i
const punctuation = /^[,.;/\-=[\]']$/

/**
 * The bytes to send the program for a chord that a terminal cannot spell with a control
 * character: Ctrl+Shift+S, Ctrl+, and every Cmd chord on a Mac, where Cmd takes the place
 * of Ctrl. It is the "CSI u" form (the key's code and its modifiers), which the program
 * reads as the same chord; null when xterm already sends the chord or the page leaves it
 * to the browser. Copy and cut are only sent when the page has no text selected (see the hook).
 */
export function chordSequence (pressed: PressedKey): string | null {
  if (pressed.altKey || (!pressed.ctrlKey && !pressed.metaKey)) return null

  const key = pressed.key.toLowerCase()
  const isLetter = letter.test(key)
  if (!isLetter && !punctuation.test(key)) return null
  if (pressed.shiftKey && developerToolsKeys.has(key)) return null // the browser keeps the developer tools
  // Ctrl+letter is a control character, except copy, cut and paste: the program asks the machine's
  // clipboard itself, which the page cannot reach.
  if (isLetter && !pressed.shiftKey && !pressed.metaKey && !clipboardKeys.has(key)) return null

  const shiftBit = isLetter && pressed.shiftKey ? 1 : 0
  const ctrlBit = 4

  return `\u{1B}[${key.codePointAt(0)};${1 + shiftBit + ctrlBit}u`
}
