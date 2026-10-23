import { act } from 'react'
import { createRoot } from 'react-dom/client'
import { WindowControls } from './window-controls.component'

const sent: unknown[][] = []

describe('WindowControls', () => {
  let unmount: () => void
  let host: HTMLElement

  beforeEach(() => {
    sent.length = 0
    Reflect.set(globalThis, 'IS_REACT_ACT_ENVIRONMENT', true)
    Reflect.set(globalThis, 'runtime', { EventsEmit: (...data: unknown[]) => { sent.push(data) } })
    host = document.createElement('div')
    const root = createRoot(host)
    act(() => root.render(<WindowControls />))
    unmount = () => act(() => root.unmount())
  })

  afterEach(() => {
    unmount()
    Reflect.deleteProperty(globalThis, 'runtime')
  })

  it('asks the program to minimise, maximise and close the window', () => {
    const buttons = [...host.querySelectorAll('button')]
    expect(buttons.map(button => button.getAttribute('aria-label'))).toEqual(['Minimise', 'Maximise', 'Close'])
    for (const button of buttons) act(() => button.click())

    expect(sent).toEqual([['mvd:window', 'minimise'], ['mvd:window', 'maximise'], ['mvd:window', 'close']])
  })
})
