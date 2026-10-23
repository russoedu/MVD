import type { PressedKey } from './browser-shortcut.policy'
import { chordSequence } from './chord-sequence.algorithm'

const press = (key: string, modifiers: Partial<PressedKey> = {}): PressedKey => ({
  key, ctrlKey: false, metaKey: false, altKey: false, shiftKey: false, ...modifiers,
})

describe('chordSequence', () => {
  it('spells Ctrl+Shift+S with its modifiers', () => {
    expect(chordSequence(press('S', { ctrlKey: true, shiftKey: true }))).toBe('\u{1B}[115;6u')
  })

  it('spells Ctrl+, which a terminal cannot', () => {
    expect(chordSequence(press(',', { ctrlKey: true }))).toBe('\u{1B}[44;5u')
  })

  it('takes Cmd for Ctrl on a Mac', () => {
    expect(chordSequence(press('s', { metaKey: true }))).toBe('\u{1B}[115;5u')
    expect(chordSequence(press(',', { metaKey: true }))).toBe('\u{1B}[44;5u')
  })

  it('leaves plain Ctrl+letter to xterm', () => {
    expect(chordSequence(press('s', { ctrlKey: true }))).toBeNull()
  })

  it('spells copy, cut and paste, which the program handles with the machine clipboard', () => {
    expect(chordSequence(press('c', { ctrlKey: true }))).toBe('[99;5u')
    expect(chordSequence(press('v', { metaKey: true }))).toBe('[118;5u')
    expect(chordSequence(press('x', { ctrlKey: true }))).toBe('[120;5u')
  })

  it('leaves the developer tools chords to the browser', () => {
    expect(chordSequence(press('I', { ctrlKey: true, shiftKey: true }))).toBeNull()
    expect(chordSequence(press('C', { ctrlKey: true, shiftKey: true }))).toBeNull()
  })

  it('ignores Alt chords and plain keys', () => {
    expect(chordSequence(press('s', { ctrlKey: true, altKey: true }))).toBeNull()
    expect(chordSequence(press('s'))).toBeNull()
  })
})
