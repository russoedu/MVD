import { useEffect } from 'react'
import { terminalBackground } from './terminal-background.algorithm'

/** The custom property the page's style reads (see styles.css). */
const property = '--terminal-background'

const nothing = (): void => {}

/** How long the page waits after the terminal redraws before it looks at the colour again. */
const settleMs = 150

/**
 * Keeps the strip beside the terminal's last column and under its last row the colour of
 * the screen, whatever theme the program draws. The program only draws whole cells, so
 * the page watches what it drew and wears its colour.
 */
export function useFollowTerminalBackground (host: React.RefObject<HTMLElement | null>): void {
  useEffect(() => {
    const element = host.current

    if (element === null) {
      return nothing
    }

    let timer: ReturnType<typeof setTimeout> | undefined
    let shown: string | undefined

    const update = (): void => {
      timer = undefined
      const colour = terminalBackground(element)

      if (colour !== undefined && colour !== shown) {
        shown = colour
        element.style.setProperty(property, colour)
      }
    }

    // The terminal rewrites its rows all the time; one look after it settles is enough.
    const observer = new MutationObserver(() => {
      timer ??= setTimeout(update, settleMs)
    })
    observer.observe(element, { childList: true, subtree: true, attributes: true, attributeFilter: ['style'] })
    update()

    return () => {
      observer.disconnect()
      clearTimeout(timer)
      element.style.removeProperty(property)
    }
  }, [host])
}
