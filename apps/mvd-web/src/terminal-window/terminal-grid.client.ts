import type { CellGrid } from './shift-click.algorithm'

/** What xterm draws its cells into, and the text it measures one cell with. */
const screenElement = '.xterm-screen'
const measureElement = '.xterm-char-measure-element'
const rowElements = '.xterm-rows > div'

/** Where the terminal is on the page and how big its cells are; null before it has drawn. */
export function readGrid (host: ParentNode): CellGrid | null {
  const screen = host.querySelector(screenElement)
  const measure = host.querySelector(measureElement)
  const rows = host.querySelectorAll(rowElements)
  const length = measure?.textContent?.length ?? 0

  const empty = rows.length === 0

  if (empty || length === 0 || screen === null || measure === null) return null

  const box = screen.getBoundingClientRect()
  const cellWidth = measure.getBoundingClientRect().width / length

  return {
    left:       box.left,
    top:        box.top,
    cellWidth,
    cellHeight: box.height / rows.length,
    cols:       Math.round(box.width / cellWidth),
    rows:       rows.length,
  }
}
