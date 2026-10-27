import { isPlainCharacter, matchesShortcut } from './shortcut-match.policy'

const press = (key: string, more: Partial<{ ctrlKey: boolean, metaKey: boolean, shiftKey: boolean, altKey: boolean }> = {}) =>
  ({ key, ctrlKey: false, metaKey: false, shiftKey: false, altKey: false, ...more })

describe('matchesShortcut', () => {
  it('matches the key in either case with exactly its modifiers', () => {
    expect(matchesShortcut(press('S', { ctrlKey: true, shiftKey: true }), { ctrl: true, shift: true, key: 'S' })).toBe(true)
    expect(matchesShortcut(press('s', { ctrlKey: true }), { ctrl: true, shift: true, key: 'S' })).toBe(false)
    expect(matchesShortcut(press('l', { ctrlKey: true, altKey: true }), { ctrl: true, key: 'L' })).toBe(false)
  })

  it('takes ⌘ for Ctrl, and the browser names of Del and the arrows', () => {
    expect(matchesShortcut(press('l', { metaKey: true }), { ctrl: true, key: 'L' })).toBe(true)
    expect(matchesShortcut(press('Delete'), { key: 'Del' })).toBe(true)
    expect(matchesShortcut(press('ArrowUp', { shiftKey: true }), { shift: true, key: 'Up' })).toBe(true)
  })
})

describe('isPlainCharacter', () => {
  it('is a letter or a digit with no Ctrl, ⌘ or Alt', () => {
    expect(isPlainCharacter(press('d'))).toBe(true)
    expect(isPlainCharacter(press('D', { shiftKey: true }))).toBe(true)
    expect(isPlainCharacter(press('d', { ctrlKey: true }))).toBe(false)
    expect(isPlainCharacter(press('Enter'))).toBe(false)
  })
})
