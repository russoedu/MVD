import { fireEvent, render, screen, within } from '@testing-library/react'
import { classicThemeOf, dosTheme } from '../classic-themes'
import { PreferencesWindow } from './preferences-window.component'
import type { PreferencesIntent, PreferencesSnapshot } from './preferences-window.contract'

const prefs: PreferencesSnapshot = {
  title:  'Preferences',
  tabs:   ['Output', 'Download', 'Advanced', 'Theme'],
  tab:    0,
  fields: [
    { label: 'Videos Output Folder', kind: 'folder', value: '/music' },
    {
      label:   'Output Template',
      kind:    'template',
      value:   '%(title)s.%(ext)s',
      example: 'Part Time Lover.mp4',
      units:   [{ text: 'TITLE', block: true }, { text: ' ' }, { text: '-' }],
      blocks:  ['ARTIST', 'TITLE'],
    },
    { label: 'Retry Attempts', kind: 'number', value: '3', min: 1, max: 999, step: 1 },
    { label: 'Window style', kind: 'choice', value: 'dos', options: [{ value: 'dos', label: 'MVD terminal (DOS)' }, { value: 'aqua', label: 'Mac OS X 10.0' }] },
    { label: 'Accent (title)', kind: 'colour', value: '#ff2a6d' },
    { label: 'Theme', kind: 'choice', value: 'MVD Neon', hidden: true },
    { label: 'Uninstall MVD', kind: 'action', value: '' },
  ],
  colours: { background: '', accent: '#ff2a6d', focus: '#05d9e8', highlight: '#f9f871', success: '#01ff89', error: '#ff4365', dim: '#7a7aa8', text: '#e0e0ff', selected: '#2a1b5e' },
}

function open (theme = classicThemeOf('windows-95'), snapshot = prefs): PreferencesIntent[] {
  const sent: PreferencesIntent[] = []
  render(<PreferencesWindow theme={theme} prefs={snapshot} onIntent={intent => { sent.push(intent) }} />)

  return sent
}

describe('PreferencesWindow', () => {
  beforeEach(() => {
    if (!('CSS' in globalThis)) Reflect.set(globalThis, 'CSS', { escape: (text: string) => text, supports: () => false })
  })

  it('is a window of the theme with the tabs, the settings of the one shown, and Save before Close on Windows', () => {
    open()
    const dialog = screen.getByRole('dialog', { name: 'Preferences' })
    expect(within(dialog).getAllByRole('tab').map(tab => tab.textContent)).toEqual(prefs.tabs)
    expect(within(dialog).getByRole('textbox', { name: 'Videos Output Folder' })).toBeTruthy()
    const footer = within(dialog).getAllByRole('button').map(button => button.textContent).filter(text => text === 'Save' || text === 'Close')
    expect(footer).toEqual(['Save', 'Close'])
    expect(document.documentElement.dataset.theme).toBe('windows-95')
    expect(within(dialog).queryByText('Theme', { selector: '.preferences-label' })).toBeNull()
  })

  it('sends the tab chosen, a folder typed, Browse…, a block added and Save', () => {
    const sent = open()
    fireEvent.click(screen.getByRole('tab', { name: 'Theme' }))
    const folder = screen.getByRole('textbox', { name: 'Videos Output Folder' })
    fireEvent.change(folder, { target: { value: '/videos' } })
    fireEvent.blur(folder)
    fireEvent.click(screen.getByRole('button', { name: 'Browse…' }))
    fireEvent.click(screen.getByRole('button', { name: '+ ARTIST' }))
    fireEvent.click(screen.getByRole('button', { name: 'Take out TITLE' }))
    fireEvent.click(screen.getByRole('button', { name: 'Save' }))

    expect(sent).toEqual([
      { kind: 'prefs-tab', index: 3 },
      { kind: 'prefs-set', field: 0, text: '/videos' },
      { kind: 'prefs-activate', field: 0 },
      { kind: 'prefs-add-block', field: 1, index: 0 },
      { kind: 'prefs-remove-unit', field: 1, index: 0 },
      { kind: 'prefs-save' },
    ])
  })

  it('puts Save last on a Mac, and the terminal look in the colours of its theme', () => {
    open(classicThemeOf('mac-os-9'))
    const labels = screen.getAllByRole('button').map(button => button.textContent).filter(text => text === 'Save' || text === 'Close')
    expect(labels).toEqual(['Close', 'Save'])
  })

  it('wears the terminal look with its colours for the DOS style', () => {
    open(dosTheme)
    expect(document.documentElement.dataset.family).toBe('dos')
    expect(document.documentElement.style.getPropertyValue('--dos-accent')).toBe('#ff2a6d')
  })

  it('says what went wrong, and keeps the settings out of reach while the program asks', () => {
    open(classicThemeOf('windows-xp'), { ...prefs, status: 'Use a colour like #ff007f or #f07.', bad: true, busy: true })
    expect(screen.getByRole('status').textContent).toBe('Use a colour like #ff007f or #f07.')
    expect(document.querySelector('.preferences-body')?.hasAttribute('inert')).toBe(true)
  })
})
