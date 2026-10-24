import { act } from 'react'
import { createRoot } from 'react-dom/client'
import { focusTerminal, useFocusTerminal } from './focus-terminal.hook'

function terminalIn (parent: HTMLElement): HTMLTextAreaElement {
  const input = document.createElement('textarea')
  input.className = 'xterm-helper-textarea'
  parent.append(input)

  return input
}

function Probe ({ host }: { host: React.RefObject<HTMLElement | null> }) {
  useFocusTerminal(host)

  return null
}

describe('focusTerminal', () => {
  afterEach(() => {
    document.body.replaceChildren()
  })

  it('gives the keyboard to the terminal when nothing has it', () => {
    const host = document.createElement('main')
    document.body.append(host)
    const input = terminalIn(host)

    focusTerminal(host)

    expect(document.activeElement).toBe(input)
  })

  it('leaves the keyboard where it is when a field already has it', () => {
    const host = document.createElement('main')
    const other = document.createElement('input')
    document.body.append(host, other)
    terminalIn(host)
    other.focus()

    focusTerminal(host)

    expect(document.activeElement).toBe(other)
  })

  it('does nothing before the terminal has drawn its input', () => {
    const host = document.createElement('main')
    document.body.append(host)

    expect(() => focusTerminal(host)).not.toThrow()
    expect(document.activeElement).toBe(document.body)
  })
})

describe('useFocusTerminal', () => {
  beforeEach(() => {
    Reflect.set(globalThis, 'IS_REACT_ACT_ENVIRONMENT', true)
    jest.useFakeTimers()
  })

  afterEach(() => {
    jest.useRealTimers()
    document.body.replaceChildren()
  })

  it('focuses the terminal once it appears, and again when the window is activated', () => {
    const host = document.createElement('main')
    document.body.append(host)
    const root = createRoot(document.createElement('div'))
    act(() => root.render(<Probe host={{ current: host }} />))
    const input = terminalIn(host)

    act(() => { jest.advanceTimersByTime(500) })
    expect(document.activeElement).toBe(input)

    input.blur()
    act(() => { dispatchEvent(new Event('focus')) })
    expect(document.activeElement).toBe(input)

    act(() => root.unmount())
  })
})
