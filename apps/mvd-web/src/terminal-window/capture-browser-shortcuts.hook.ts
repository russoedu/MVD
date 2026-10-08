import { useEffect } from 'react'
import { takesOverBrowserShortcut } from './browser-shortcut.policy'

/** Where the terminal reads the keyboard: xterm's hidden input. */
const terminalInput = (): HTMLElement | null => document.querySelector<HTMLElement>('.xterm-helper-textarea')

/** Places that use the keyboard for themselves: the About window and form fields. */
const ownsKeyboard = (target: Element | null): boolean => target?.closest('dialog, input, textarea, select, [contenteditable]') != null

/**
 * Keeps the browser's own shortcuts (Ctrl+P prints, Ctrl+S saves the page...)
 * from firing while the window is focused, and hands the keys to the terminal
 * interface instead, even when the terminal itself does not have the focus
 * (after a click on the header, say). The terminal takes the focus back
 * whenever the window gains it.
 */
export function useCaptureBrowserShortcuts (): void {
  useEffect(() => {
    const focusTerminal = (): void => {
      if (document.querySelector('dialog[open]')) return
      terminalInput()?.focus()
    }

    const onKeyDown = (event: KeyboardEvent): void => {
      if (!takesOverBrowserShortcut(event)) return

      const target = event.target instanceof Element ? event.target : null
      if (target?.closest('.xterm')) return // the terminal already keeps the key from the browser
      if (ownsKeyboard(target)) return

      event.preventDefault()
      const input = terminalInput()
      if (!input) return
      input.focus()
      input.dispatchEvent(new KeyboardEvent('keydown', {
        key:        event.key,
        code:       event.code,
        // xterm builds Ctrl+letter from keyCode, which has no replacement in `key`.
        // eslint-disable-next-line unicorn/prefer-keyboard-event-key
        keyCode:    event.keyCode,
        ctrlKey:    event.ctrlKey,
        metaKey:    event.metaKey,
        bubbles:    true,
        cancelable: true,
      }))
    }

    globalThis.addEventListener('keydown', onKeyDown, { capture: true })
    window.addEventListener('focus', focusTerminal)

    return () => {
      globalThis.removeEventListener('keydown', onKeyDown, true)
      window.removeEventListener('focus', focusTerminal)
    }
  }, [])
}
