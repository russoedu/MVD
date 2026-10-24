import { useEffect } from 'react'
import { cellAt, shiftClickSequence } from './shift-click.algorithm'
import { readGrid } from './terminal-grid.client'
import { sendTerminalInput } from './terminal-socket.client'

const nothing = (): void => {}

/**
 * A click with Shift held is xterm's to select text with, so it never reaches the program, which wants it
 * to tick a range of songs. This catches it before xterm does and sends the program the click, with
 * Shift, as a terminal would have.
 */
export function useForwardShiftClicks (host: React.RefObject<HTMLElement | null>): void {
  useEffect(() => {
    const element = host.current

    if (element === null) {
      return nothing
    }

    const forward = (event: MouseEvent): void => {
      if (!event.shiftKey || event.button !== 0) return

      const grid = readGrid(element)
      const cell = grid === null ? null : cellAt({ x: event.clientX, y: event.clientY }, grid)
      if (cell === null) return

      event.preventDefault()
      event.stopPropagation()
      sendTerminalInput(shiftClickSequence(cell))
    }

    element.addEventListener('mousedown', forward, { capture: true })

    return () => element.removeEventListener('mousedown', forward, { capture: true })
  }, [host])
}
