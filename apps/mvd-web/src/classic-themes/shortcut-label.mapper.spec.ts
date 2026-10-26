import { shortcutLabel } from './shortcut-label.mapper'

describe('shortcutLabel', () => {
  it('writes Windows shortcuts with Ctrl, Alt and Shift joined by plus signs', () => {
    expect(shortcutLabel({ ctrl: true, shift: true, key: 'S' }, 'windows')).toBe('Ctrl+Shift+S')
    expect(shortcutLabel({ key: 'F1' }, 'windows')).toBe('F1')
  })

  it('writes Mac shortcuts with symbols, shift before command, and Ctrl as command', () => {
    expect(shortcutLabel({ ctrl: true, shift: true, key: 'S' }, 'mac')).toBe('⇧⌘S')
    expect(shortcutLabel({ ctrl: true, key: 'L' }, 'mac')).toBe('⌘L')
    expect(shortcutLabel({ shift: true, key: 'Up' }, 'mac')).toBe('⇧↑')
  })
})
