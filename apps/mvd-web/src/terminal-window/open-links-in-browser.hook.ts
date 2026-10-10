import { useEffect } from 'react'
import { isLinkQuestion, isWebAddress } from './external-link.policy'

interface WailsPage {
  runtime?: { BrowserOpenURL?: (url: string) => void }
}

/** Hands an address to Wails, which opens it in the person's default browser. */
function openInDefaultBrowser (address: string): void {
  if (!isWebAddress(address)) return
  (globalThis as WailsPage).runtime?.BrowserOpenURL?.(address)
}

/** Replaces a function of the browser page. */
function replace (name: 'confirm' | 'open', value: unknown): void {
  Object.defineProperty(globalThis, name, { value, configurable: true, writable: true })
}

/**
 * Sends the links of the terminal (the video addresses of the playlist editor)
 * straight to the person's default browser.
 *
 * xterm follows a link by asking "Do you want to navigate to...? WARNING..." with
 * confirm() and then window.open(), which the web view answers with its own box
 * and its own window. Both are replaced here: the question is answered yes, and
 * the address xterm writes into the "new window" goes to Wails' BrowserOpenURL.
 */
export function useOpenLinksInBrowser (): void {
  useEffect(() => {
    const originalConfirm = globalThis.confirm.bind(globalThis)
    const originalOpen = globalThis.open.bind(globalThis)

    replace('confirm', (message?: string): boolean => (isLinkQuestion(message) ? true : originalConfirm(message)))

    replace('open', (url?: string | URL): unknown => {
      // xterm opens a blank window and then sets the address of it.
      if (url === undefined || String(url) === '') {
        return { opener: null, location: { set href (address: string) { openInDefaultBrowser(address) } } }
      }
      openInDefaultBrowser(String(url))

      return null
    })

    return () => {
      replace('confirm', originalConfirm)
      replace('open', originalOpen)
    }
  }, [])
}
