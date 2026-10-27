import { classicThemeOf } from './classic-theme.model'

describe('classicThemeOf', () => {
  it('finds a theme by its id and falls back to Windows 95', () => {
    expect(classicThemeOf('aqua').family).toBe('aqua')
    expect(classicThemeOf('mac-os-9').family).toBe('mac-classic')
    expect(classicThemeOf('mac-os-x-10-4')).toMatchObject({ family: 'aqua', label: 'Mac OS X 10.4' })
    expect(classicThemeOf('windows-3.1').id).toBe('windows-95')
    expect(classicThemeOf(null).id).toBe('windows-95')
  })
})
