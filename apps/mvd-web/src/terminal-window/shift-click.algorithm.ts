/** The terminal's drawn area on the page, and how big each of its cells is. */
export interface CellGrid {
  left:       number
  top:        number
  cellWidth:  number
  cellHeight: number
  cols:       number
  rows:       number
}

/** A cell of the terminal, counted from 1 as the mouse reports of a terminal are. */
export interface Cell {
  col: number
  row: number
}

/** The cell under a point of the page, or null when the point is outside the terminal. */
export function cellAt (point: { x: number, y: number }, grid: CellGrid): Cell | null {
  if (grid.cellWidth <= 0 || grid.cellHeight <= 0) return null

  const col = Math.floor((point.x - grid.left) / grid.cellWidth) + 1
  const row = Math.floor((point.y - grid.top) / grid.cellHeight) + 1
  if (col < 1 || row < 1 || col > grid.cols || row > grid.rows) return null

  return { col, row }
}

/**
 * The bytes for a left click with Shift held, in the form a terminal reports the mouse (SGR): the
 * press and the release. xterm keeps a click with Shift for itself, to select text, and never
 * reports it, so the page spells it for the program: that is how it ticks a range of songs.
 */
export function shiftClickSequence (cell: Cell): string {
  const shift = 4

  return `\u{1B}[<${shift};${cell.col};${cell.row}M\u{1B}[<${shift};${cell.col};${cell.row}m`
}
