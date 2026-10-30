import { useEffect } from 'react'

const nothing = (): void => {}

/** The input xterm reads keys from. */
const keyboardInput = '.xterm-helper-textarea'

/**
 * Puts the keyboard in the terminal whenever nothing else has it: when the page opens and
 * when the window is activated again. Without it, keys do nothing until the first click.
 * A popup (the About window) or any field that has focus keeps it.
 */
export function focusTerminal (host: ParentNode): void {
  const active = document.activeElement
  const nothingHasFocus = [null, document.body, document.documentElement].includes(active as HTMLElement | null)

  if (!nothingHasFocus) {
    return
  }

  host.querySelector<HTMLElement>(keyboardInput)?.focus()
}

/**
 * Keeps the keyboard in the terminal: puts focus on its input when the page or the window gets
 * focus again and nothing else holds it.
 *
 * @param host The element that holds the terminal.
 * @param enabled False while another window has the keyboard. The default is true.
 * @returns Nothing. The listeners are removed when the component unmounts.
 */
export function useFocusTerminal (host: React.RefObject<HTMLElement | null>, enabled = true): void {
  useEffect(() => {
    const element = host.current

    if (element === null || !enabled) {
      return nothing
    }

    const focus = (): void => focusTerminal(element)
    // The terminal draws its input a moment after the page, so look again once it has.
    const first = setTimeout(focus, 0)
    const second = setTimeout(focus, 400)
    globalThis.addEventListener('focus', focus)
    document.addEventListener('visibilitychange', focus)

    return () => {
      clearTimeout(first)
      clearTimeout(second)
      globalThis.removeEventListener('focus', focus)
      document.removeEventListener('visibilitychange', focus)
    }
  }, [host, enabled])
}
