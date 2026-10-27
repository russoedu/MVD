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
    // jsdom draws every element as a block unless told, which would read "File" as "F ile".
    if (!document.querySelector('style[data-inline-letters]')) {
      document.head.insertAdjacentHTML('beforeend', '<style data-inline-letters>.classic-mnemonic { display: inline }</style>')
    }
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

  it('puts the control-menu box of Windows 3.1 and the close box of a classic Mac on the left, the rest on the right', () => {
    render(<ThemePreview />)
    choose('Windows 3.1')
    expect(controls()).toEqual(['Close', 'Minimise', 'Maximise'])

    choose('Mac OS 7.6')
    expect(controls()).toEqual(['Close', 'Maximise'])

    choose('Mac OS 9')
    expect(controls()).toEqual(['Close', 'Maximise', 'Minimise'])
  })

  it('draws the window for adding music on a classic Mac as a plain box with its heading inside', () => {
    render(<ThemePreview />)
    choose('Mac OS 9')
    fireEvent.click(screen.getByRole('button', { name: '+ Add source' }))

    expect(screen.getByRole('dialog', { name: 'Add music' })).toBeTruthy()
    expect(document.querySelector('.classic-dialog[data-style="plain"] .classic-title-bar')).toBeNull()
    expect(within(screen.getByRole('dialog')).getAllByRole('button').map(button => button.textContent)).toEqual(['Cancel', 'Add'])
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
  it('underlines the keyboard letters always on Windows 95, after Alt on Vista, and never on a Mac', () => {
    render(<ThemePreview />)
    choose('Windows 95')
    expect(document.documentElement.dataset.mnemonics).toBe('shown')
    expect(screen.getByRole('button', { name: 'File' }).querySelector('.classic-mnemonic')?.textContent).toBe('F')

    choose('Windows Vista')
    expect(document.documentElement.dataset.mnemonics).toBe('hidden')
    fireEvent.keyDown(document.body, { key: 'Alt', code: 'AltLeft', altKey: true })
    expect(document.documentElement.dataset.mnemonics).toBe('shown')
    fireEvent.pointerDown(document.body)
    expect(document.documentElement.dataset.mnemonics).toBe('hidden')

    choose('Mac OS X 10.0')
    fireEvent.keyDown(document.body, { key: 'Alt', code: 'AltLeft', altKey: true })
    expect(document.documentElement.dataset.mnemonics).toBe('hidden')
  })

  it('opens a menu with Alt and its letter, and picks an entry with its letter', () => {
    render(<ThemePreview />)
    choose('Windows 95')
    fireEvent.keyDown(document.body, { key: 'f', code: 'KeyF', altKey: true })
    const menu = screen.getByRole('menu', { name: 'File' })

    fireEvent.keyDown(within(menu).getAllByRole('menuitem')[0], { key: 's', code: 'KeyS' })
    expect(screen.getByText('{"kind":"menu","id":"save"}')).toBeTruthy()
    expect(screen.queryByRole('menu')).toBeNull()
  })

  it('leaves Alt and a letter to the Mac, which has no keyboard letters', () => {
    render(<ThemePreview />)
    choose('Mac OS X 10.0')
    fireEvent.keyDown(document.body, { key: 'f', code: 'KeyF', altKey: true })

    expect(screen.queryByRole('menu')).toBeNull()
  })

  it('reaches the choices of the window for adding music with Alt and their letter', () => {
    render(<ThemePreview />)
    choose('Windows 95')
    fireEvent.click(screen.getByRole('button', { name: '+ Add source' }))
    fireEvent.keyDown(screen.getByRole('textbox', { name: 'URL' }), { key: 'm', code: 'KeyM', altKey: true })

    expect(screen.getByRole('textbox', { name: 'Artist' })).toBeTruthy()
    expect(screen.queryByRole('menu')).toBeNull()
  })

  it("shows the theme's busy pointers for the activity the snapshot names", () => {
    render(<ThemePreview />)
    fireEvent.change(screen.getByRole('combobox', { name: 'Pointer' }), { target: { value: 'waiting' } })

    expect(document.querySelector<HTMLElement>('.classic-window')?.dataset.activity).toBe('waiting')
  })

  it("opens a source's own menu with a right click, to remove it or move it", () => {
    render(<ThemePreview />)
    fireEvent.contextMenu(screen.getByRole('option', { name: /90s UK Dance Hits/ }))
    const menu = screen.getByRole('menu', { name: 'Source' })
    expect(within(menu).getAllByRole('menuitem').map(item => item.textContent)).toEqual(['Remove source', 'Move up', 'Move down'])

    fireEvent.click(within(menu).getByRole('menuitem', { name: 'Remove source' }))
    expect(screen.getByText('{"kind":"menu","id":"delete"}')).toBeTruthy()
  })
})
