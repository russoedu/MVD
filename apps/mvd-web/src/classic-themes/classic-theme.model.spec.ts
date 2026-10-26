import { classicThemeOf, classicThemes } from './classic-theme.model'

describe('classicThemeOf', () => {
  it('finds a theme by its id and falls back to the first theme', () => {
    expect(classicThemeOf('aqua').family).toBe('aqua')
    expect(classicThemeOf('windows-3.1')).toBe(classicThemes[0])
    expect(classicThemeOf(null)).toBe(classicThemes[0])
  })
})
