import { fireEvent, render, screen, within } from '@testing-library/react'
import { ThemePreview } from './theme-preview.route'

function choose (label: string): void {
  fireEvent.click(screen.getByRole('button', { name: label }))
}

function controls (): (string | null)[] {
  const bar = document.querySelector('.classic-window > .classic-title-bar')
  if (!bar) throw new Error('no title bar')

  return [...bar.querySelectorAll('button')].map(button => button.getAttribute('aria-label'))
}

describe('ThemePreview', () => {
  beforeEach(() => {
    Reflect.set(globalThis, 'IS_REACT_ACT_ENVIRONMENT', true)
    // jsdom has no CSS object, which React Aria reads.
    if (!('CSS' in globalThis)) Reflect.set(globalThis, 'CSS', { escape: (text: string) => text, supports: () => false })
  })

  it('sets the theme and its family on the page root, so menus outside the window wear it too', () => {
    render(<ThemePreview />)
    choose('Windows Vista')

    expect(document.documentElement.dataset.theme).toBe('windows-vista')
    expect(document.documentElement.dataset.family).toBe('windows-themed')
  })

  it('puts the window buttons on the right for Windows and on the left, close first, for Mac OS X', () => {
    render(<ThemePreview />)
    choose('Windows 95')
    expect(controls()).toEqual(['Minimise', 'Maximise', 'Close'])
    expect(document.querySelector('.classic-title-bar')?.firstElementChild?.className).toBe('classic-title')

    choose('Mac OS X 10.0')
    expect(controls()).toEqual(['Close', 'Minimise', 'Maximise'])
    expect(document.querySelector('.classic-title-bar')?.firstElementChild?.className).toBe('classic-window-controls')
  })

  it('writes the key hints the way each family does', () => {
    render(<ThemePreview />)
    choose('Windows 95')
    expect(screen.getByText('Ctrl+L')).toBeTruthy()

    choose('Mac OS X 10.0')
    expect(screen.getByText('⌘L')).toBeTruthy()
  })

  it('opens the window for adding music, with Add first on Windows and last on a Mac', () => {
    render(<ThemePreview />)
    choose('Windows 95')
    fireEvent.click(screen.getByRole('button', { name: '+ Add source' }))
    let dialog = screen.getByRole('dialog')
    expect(within(dialog).getAllByRole('button').map(button => button.textContent).filter(text => text === 'Add' || text === 'Cancel')).toEqual(['Add', 'Cancel'])

    fireEvent.click(within(dialog).getByRole('button', { name: 'Cancel' }))
    expect(screen.queryByRole('dialog')).toBeNull()

    choose('Mac OS X 10.0')
    fireEvent.click(screen.getByRole('button', { name: '+ Add source' }))
    dialog = screen.getByRole('dialog')
    expect(within(dialog).getAllByRole('button').map(button => button.textContent)).toEqual(['Cancel', 'Add'])
    expect(document.querySelector('.classic-sheet')).not.toBeNull()
  })

  it('shows the search fields when Music search is chosen, and refuses an empty search with the reason', () => {
    render(<ThemePreview />)
    fireEvent.click(screen.getByRole('button', { name: '+ Add source' }))
    fireEvent.click(screen.getByRole('radio', { name: 'Music search' }))

    expect(screen.getByRole('textbox', { name: 'Artist' })).toBeTruthy()
    expect(screen.getByRole('textbox', { name: 'Album' })).toBeTruthy()
    fireEvent.submit(document.querySelector('.classic-dialog-body') as HTMLFormElement)
    expect(screen.getByRole('alert').textContent).toBe('Write the artist')
  })
})
