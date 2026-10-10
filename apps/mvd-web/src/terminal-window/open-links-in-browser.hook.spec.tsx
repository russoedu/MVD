import { act } from 'react'
import { createRoot } from 'react-dom/client'
import { useOpenLinksInBrowser } from './open-links-in-browser.hook'

function Probe () {
  useOpenLinksInBrowser()

  return null
}

describe('useOpenLinksInBrowser', () => {
  const opened: string[] = []
  let unmount: () => void

  beforeEach(() => {
    opened.length = 0
    Reflect.set(globalThis, 'IS_REACT_ACT_ENVIRONMENT', true)
    Reflect.set(globalThis, 'runtime', { BrowserOpenURL: (address: string) => { opened.push(address) } })
    const root = createRoot(document.createElement('div'))
    act(() => root.render(<Probe />))
    unmount = () => act(() => root.unmount())
  })

  afterEach(() => {
    unmount()
    Reflect.deleteProperty(globalThis, 'runtime')
  })

  it('answers the link question and opens the address the way xterm asks for it', () => {
    const question = 'Do you want to navigate to https://youtu.be/x?\n\nWARNING: This link could potentially be dangerous'
    expect(confirm(question)).toBe(true)

    const page = open() as unknown as { opener: unknown, location: { href: string } }
    page.opener = null
    page.location.href = 'https://youtu.be/x'

    expect(opened).toEqual(['https://youtu.be/x'])
  })

  it('does not open an address that is not a web address', () => {
    const page = open() as unknown as { location: { href: string } }
    page.location.href = 'file:///C:/Windows/System32/calc.exe'

    expect(opened).toEqual([])
  })
})
