/**
 * The background the program has drawn at the bottom right of the terminal: the colour of
 * the last cell of the lowest row that has any. The terminal is a whole number of cells, so
 * a strip is left beside its last column and under its last row; the page paints that strip
 * with this colour so it looks like part of the screen.
 *
 * Rows come from the bottom up and a cell with no colour of its own (transparent) is passed
 * over, so a blank line under the screen does not make the answer nothing.
 */
export function terminalBackground (terminal: ParentNode): string | undefined {
  const rows = terminal.querySelector('.xterm-rows')?.children

  if (rows === undefined) {
    return undefined
  }

  for (let row = rows.length - 1; row >= 0; row--) {
    const last = rows[row]?.lastElementChild

    if (last === null || last === undefined) {
      continue
    }

    const colour = getComputedStyle(last).backgroundColor

    if (colour !== '' && colour !== 'transparent' && colour !== 'rgba(0, 0, 0, 0)') {
      return colour
    }
  }

  return undefined
}
