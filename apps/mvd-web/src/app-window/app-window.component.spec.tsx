import { act, fireEvent, render, screen, within } from '@testing-library/react'
import { sampleWorkspace } from '../theme-preview/sample-workspace.content'
import { AppWindow } from './app-window.component'
import { PreferencesApp } from './preferences-app.component'

jest.mock('../terminal-window', () => ({
  TerminalWindow:    ({ look }: { look: string }) => <main data-testid='terminal' data-look={look} />,
  sendTerminalInput: jest.fn(),
}))

/** The program, as the page sees it: events it sends, and the ones the page sent it. */
function fakeProgram () {
  const listeners = new Map<string, ((...data: unknown[]) => void)[]>()
  const sent: [string, unknown[]][] = []
  Reflect.set(globalThis, 'runtime', {
    EventsOn: (name: string, callback: (...data: unknown[]) => void) => {
      listeners.set(name, [...listeners.get(name) ?? [], callback])

      return () => listeners.set(name, (listeners.get(name) ?? []).filter(each => each !== callback))
    },
    EventsEmit: (name: string, ...data: unknown[]) => {
      sent.push([name, data])
    },
  })

  return {
    sent,
    send: (name: string, value: unknown) => act(() => {
      const named = listeners.get(name) ?? []
      for (const listener of named) listener(value)
    }),
  }
}

const preferences = {
  title:   'Preferences',
  tabs:    ['Output', 'Theme'],
  tab:     0,
  fields:  [],
  colours: { background: '', accent: '#ff2a6d', focus: '', highlight: '', success: '', error: '', dim: '', text: '', selected: '' },
}

const look = (): string | undefined => screen.getByTestId('terminal').dataset.look

describe('AppWindow', () => {
  // The classic look and the preferences are loaded on demand; loading them once first keeps the
  // first test from waiting on it.
  beforeAll(async () => {
    await Promise.all([import('../classic-app'), import('../preferences-window'), import('../classic-themes')])
  })

  beforeEach(() => {
    if (!('CSS' in globalThis)) Reflect.set(globalThis, 'CSS', { escape: (text: string) => text, supports: () => false })
    // jsdom has no ResizeObserver, which the dialog of a covered screen watches itself with.
    if (!('ResizeObserver' in globalThis)) Reflect.set(globalThis, 'ResizeObserver', class { observe (): void {} disconnect (): void {} })
  })

  afterEach(() => Reflect.deleteProperty(globalThis, 'runtime'))

  it('asks the program for the style and stays the terminal look while it is the DOS one', () => {
    const program = fakeProgram()
    render(<AppWindow />)
    expect(program.sent.map(([name]) => name)).toContain('mvd:style-request')

    program.send('mvd:style', 'dos')
    program.send('mvd:snapshot', JSON.stringify(sampleWorkspace))
    expect(look()).toBe('page')
    expect(document.querySelector('.classic-window')).toBeNull()
  })

  it('draws the classic window from what the program sends, with the terminal running behind it', async () => {
    const program = fakeProgram()
    render(<AppWindow />)
    program.send('mvd:style', JSON.stringify({ style: 'windows-95', scheme: 'brick' }))
    program.send('mvd:snapshot', JSON.stringify(sampleWorkspace))

    // The classic look is loaded on demand, the first time slowly.
    expect(await screen.findByRole('button', { name: '+ Add source' }, { timeout: 5000 })).toBeTruthy()
    expect(look()).toBe('behind')
    expect(document.documentElement.dataset.scheme).toBe('brick')

    fireEvent.click(screen.getByRole('button', { name: '+ Add source' }))
    const intents = program.sent.filter(([name]) => name === 'mvd:intent').map(([, data]) => JSON.parse(String(data[0])) as unknown)
    expect(intents).toContainEqual({ kind: 'button', id: 'add-source' })
  })

  it('sends a menu shortcut as the menu entry, and a plain letter as a key', async () => {
    const program = fakeProgram()
    render(<AppWindow />)
    program.send('mvd:style', 'windows-xp')
    program.send('mvd:snapshot', JSON.stringify(sampleWorkspace))
    await screen.findByRole('button', { name: '+ Add source' }, { timeout: 5000 })

    fireEvent.keyDown(document.body, { key: 'l', ctrlKey: true })
    fireEvent.keyDown(document.body, { key: 'd' })
    const intents = program.sent.filter(([name]) => name === 'mvd:intent').map(([, data]) => JSON.parse(String(data[0])) as unknown)
    expect(intents).toEqual([{ kind: 'menu', id: 'add-source' }, { kind: 'key', key: 'd' }])
  })

  it('shows the terminal in a dialog of the theme for a screen the themes do not draw yet', async () => {
    const program = fakeProgram()
    render(<AppWindow />)
    program.send('mvd:style', 'mac-os-9')
    program.send('mvd:snapshot', JSON.stringify({ ...sampleWorkspace, covered: { title: 'Preferences' } }))

    expect(await screen.findByRole('region', { name: 'Preferences' }, { timeout: 5000 })).toBeTruthy()
    expect(look()).toBe('inset')
  })

  it('cannot be used while the preferences are open in their own window', () => {
    const program = fakeProgram()
    render(<AppWindow />)
    program.send('mvd:style', 'dos')
    expect(document.querySelector('.app-window')?.hasAttribute('inert')).toBe(false)

    program.send('mvd:prefs', JSON.stringify(preferences))
    expect(document.querySelector('.app-window')?.hasAttribute('inert')).toBe(true)
    expect(screen.queryByRole('dialog')).toBeNull()

    program.send('mvd:prefs', 'null')
    expect(document.querySelector('.app-window')?.hasAttribute('inert')).toBe(false)
  })

  it('asks the quit question as an alert of the theme, with Keep downloading the safe answer', async () => {
    const program = fakeProgram()
    render(<AppWindow />)
    program.send('mvd:style', JSON.stringify({ style: 'mac-os-9', scheme: 'default' }))
    program.send('mvd:snapshot', JSON.stringify({
      ...sampleWorkspace,
      question: {
        id:      'quit',
        title:   'Quit MVD?',
        tone:    'caution',
        lines:   ['59 songs are still downloading.', 'Quit and stop them?'],
        buttons: [{ id: 'quit', label: 'Quit', role: 'other', enabled: true }, { id: 'keep', label: 'Keep downloading', role: 'default', enabled: true }],
      },
    }))

    const alert = await screen.findByRole('alertdialog', { name: 'Quit MVD?' }, { timeout: 5000 })
    expect(look()).toBe('behind')
    // A Mac puts the default button last.
    expect(within(alert).getAllByRole('button').map(button => button.textContent)).toEqual(['Quit', 'Keep downloading'])
    fireEvent.click(within(alert).getByRole('button', { name: 'Quit' }))
    const intents = program.sent.filter(([name]) => name === 'mvd:intent').map(([, data]) => JSON.parse(String(data[0])) as unknown)
    expect(intents).toContainEqual({ kind: 'question-button', id: 'quit' })
  })
})

describe('PreferencesApp', () => {
  beforeAll(async () => {
    await Promise.all([import('../preferences-window'), import('../classic-themes')])
  })

  afterEach(() => Reflect.deleteProperty(globalThis, 'runtime'))

  it('fills its own window with the preferences, in the window style, and sends what the person does', async () => {
    const program = fakeProgram()
    render(<PreferencesApp />)
    expect(program.sent.map(([name]) => name)).toContain('mvd:style-request')
    program.send('mvd:style', 'dos')
    program.send('mvd:prefs', JSON.stringify(preferences))

    expect(await screen.findByRole('dialog', { name: 'Preferences' }, { timeout: 5000 })).toBeTruthy()
    expect(document.querySelector('.preferences-own-window')).not.toBeNull()
    expect(document.documentElement.dataset.family).toBe('dos')

    // The Close button at the bottom, not the one of the title bar.
    fireEvent.click(screen.getAllByRole('button', { name: 'Close' }).find(button => button.classList.contains('classic-button')) as HTMLElement)
    fireEvent.keyDown(screen.getByRole('dialog'), { key: 'Escape' })
    const intents = program.sent.filter(([name]) => name === 'mvd:prefs-intent').map(([, data]) => JSON.parse(String(data[0])) as unknown)
    expect(intents).toEqual([{ kind: 'prefs-close' }, { kind: 'prefs-close' }])
  })

  it('shows nothing until the program has the preferences open', () => {
    const program = fakeProgram()
    const { container } = render(<PreferencesApp />)
    program.send('mvd:style', 'dos')

    expect(container.children).toHaveLength(0)
  })
})
