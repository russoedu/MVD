import { act } from 'react'
import { createRoot } from 'react-dom/client'
import { Ticker } from './ticker.component'

describe('Ticker', () => {
  let container: HTMLDivElement
  let unmount: () => void

  beforeEach(() => {
    Reflect.set(globalThis, 'IS_REACT_ACT_ENVIRONMENT', true)
    container = document.createElement('div')
    const root = createRoot(container)
    act(() => root.render(<Ticker text='I want my MTV' />))
    unmount = () => act(() => root.unmount())
  })

  afterEach(() => unmount())

  const moving = () => container.querySelector('[aria-hidden="true"]')!

  it('is read once as plain text, and the moving copy is hidden from screen readers', () => {
    expect(container.querySelector('.visually-hidden')?.textContent).toBe('I want my MTV')
    expect(moving().textContent).toBe('I want my MTV')
  })

  it('moves at first and stays still after a few passes', () => {
    const ticker = container.querySelector<HTMLElement>('.page-header-ticker')!
    expect(ticker.dataset.settled).toBe('false')

    for (let pass = 0; pass < 3; pass++) {
      act(() => { moving().dispatchEvent(new Event('animationiteration', { bubbles: true })) })
    }

    expect(ticker.dataset.settled).toBe('true')
  })
})
