import { act } from 'react'
import { createRoot } from 'react-dom/client'
import { useFollowTerminalBackground } from './follow-terminal-background.hook'

function Probe ({ host }: { host: React.RefObject<HTMLElement | null> }) {
  useFollowTerminalBackground(host)

  return null
}

function drawnWith (background: string): HTMLElement {
  const main = document.createElement('main')
  const rows = document.createElement('div')
  rows.className = 'xterm-rows'
  main.append(rows)
  paint(main, background)

  return main
}

function paint (main: HTMLElement, background: string): void {
  const rows = main.querySelector('.xterm-rows')
  const row = document.createElement('div')
  const span = document.createElement('span')
  span.style.backgroundColor = background
  row.append(span)
  rows?.replaceChildren(row)
}

describe('useFollowTerminalBackground', () => {
  beforeEach(() => {
    Reflect.set(globalThis, 'IS_REACT_ACT_ENVIRONMENT', true)
    jest.useFakeTimers()
  })

  afterEach(() => {
    jest.useRealTimers()
  })

  it('makes the strip beside the terminal the colour the program drew, and follows a new theme', async () => {
    const main = drawnWith('#2a2a2a')
    const root = createRoot(document.createElement('div'))
    act(() => root.render(<Probe host={{ current: main }} />))
    expect(main.style.getPropertyValue('--terminal-background')).toBe('rgb(42, 42, 42)')

    paint(main, '#101820')
    await act(async () => {
      await Promise.resolve()
      jest.advanceTimersByTime(200)
    })
    expect(main.style.getPropertyValue('--terminal-background')).toBe('rgb(16, 24, 32)')

    act(() => root.unmount())
  })

  it('gives the page its own colour back when it goes', () => {
    const main = drawnWith('#2a2a2a')
    const root = createRoot(document.createElement('div'))
    act(() => root.render(<Probe host={{ current: main }} />))
    act(() => root.unmount())

    expect(main.style.getPropertyValue('--terminal-background')).toBe('')
  })
})
