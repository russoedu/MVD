import { render } from '@testing-library/react'
import { classicThemeOf, dosTheme, type ClassicTheme } from '../classic-themes'
import { useThemeOnRoot } from './theme-root.hook'

function Wearer ({ theme, scheme }: { theme: ClassicTheme, scheme?: string }) {
  useThemeOnRoot(theme, scheme)

  return null
}

const root = document.documentElement

describe('useThemeOnRoot', () => {
  it('keeps the window\'s theme when a dialog opened before it closes (the style changed in the preferences)', () => {
    const win31 = classicThemeOf('windows-3-1')
    // The preferences are open in the terminal look; the style becomes Windows 3.1 in them...
    const prefs = render(<Wearer theme={dosTheme} />)
    prefs.rerender(<Wearer theme={win31} />)
    // ...so the classic window opens behind them, then the preferences close.
    const window = render(<Wearer theme={win31} scheme='hot-dog-stand' />)
    prefs.unmount()

    expect(root.dataset.theme).toBe('windows-3-1')
    expect(root.dataset.family).toBe('windows-classic')
    expect(root.dataset.scheme).toBe('hot-dog-stand')

    window.unmount()
    expect(root.dataset.theme).toBeUndefined()
    expect(root.dataset.scheme).toBeUndefined()
  })

  it('wears the newest window\'s theme, and the one before once it goes', () => {
    const window = render(<Wearer theme={classicThemeOf('mac-os-9')} />)
    const dialog = render(<Wearer theme={dosTheme} />)
    expect(root.dataset.theme).toBe('dos')

    dialog.unmount()
    expect(root.dataset.theme).toBe('mac-os-9')
    window.unmount()
  })
})
