import type { Mnemonic } from '../workspace-snapshot'

/** A label cut around its keyboard letter, so the letter can be underlined. */
export interface MnemonicParts {
  before: string
  letter: string
  after:  string
}

/** The label cut around the first place its keyboard letter appears, in either case, or null when it has none. */
export function splitMnemonic (label: string, mnemonic?: Mnemonic): MnemonicParts | null {
  if (!mnemonic) return null
  const at = label.toLowerCase().indexOf(mnemonic.toLowerCase())
  if (at === -1) return null

  return { before: label.slice(0, at), letter: label.slice(at, at + mnemonic.length), after: label.slice(at + mnemonic.length) }
}

/** The pressed key as a keyboard letter, lower case, or null when it is not a letter or a digit. */
export function pressedLetter (key: { key: string, code: string }): string | null {
  if (/^[\p{L}\p{N}]$/u.test(key.key)) return key.key.toLowerCase()
  // With Alt, some layouts give a symbol instead of the letter: the key's place still names it.
  const place = /^(?:Key|Digit)(\w)$/.exec(key.code)

  return place ? place[1].toLowerCase() : null
}

/** Whether a pressed letter (from pressedLetter) is this keyboard letter. */
export function isMnemonic (letter: string | null, mnemonic?: Mnemonic): boolean {
  return letter !== null && mnemonic?.toLowerCase() === letter
}
