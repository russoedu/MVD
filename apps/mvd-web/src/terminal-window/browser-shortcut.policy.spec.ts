import type { PressedKey } from './browser-shortcut.policy'
import { takesOverBrowserShortcut } from './browser-shortcut.policy'

const press = (key: string, modifiers: Partial<PressedKey> = {}): PressedKey => ({
  key, ctrlKey: false, metaKey: false, altKey: false, shiftKey: false, ...modifiers,
})

describe('takesOverBrowserShortcut', () => {
  it('takes Ctrl+P (print) and Ctrl+S (save page) from the browser', () => {
    expect(takesOverBrowserShortcut(press('p', { ctrlKey: true }))).toBe(true)
    expect(takesOverBrowserShortcut(press('s', { ctrlKey: true }))).toBe(true)
    expect(takesOverBrowserShortcut(press('P', { ctrlKey: true }))).toBe(true)
  })

  it('takes the Cmd chords on a Mac', () => {
    expect(takesOverBrowserShortcut(press('p', { metaKey: true }))).toBe(true)
  })

  it('leaves plain keys alone', () => {
    expect(takesOverBrowserShortcut(press('p'))).toBe(false)
    expect(takesOverBrowserShortcut(press('Enter'))).toBe(false)
  })

  it('leaves copy, paste and cut to the browser', () => {
    for (const key of ['c', 'v', 'x']) {
      expect(takesOverBrowserShortcut(press(key, { ctrlKey: true }))).toBe(false)
    }
  })

  it('leaves Shift and Alt chords alone, such as the developer tools and Ctrl+Shift+M', () => {
    expect(takesOverBrowserShortcut(press('m', { ctrlKey: true, shiftKey: true }))).toBe(false)
    expect(takesOverBrowserShortcut(press('i', { ctrlKey: true, shiftKey: true }))).toBe(false)
    expect(takesOverBrowserShortcut(press('d', { ctrlKey: true, altKey: true }))).toBe(false)
  })

  it('ignores the modifier key on its own', () => {
    expect(takesOverBrowserShortcut(press('Control', { ctrlKey: true }))).toBe(false)
    expect(takesOverBrowserShortcut(press('Meta', { metaKey: true }))).toBe(false)
  })
})
