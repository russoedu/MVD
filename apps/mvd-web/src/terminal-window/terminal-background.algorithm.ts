/** How many rows must end in the same colour before the page trusts it as the screen's own. */
const agreeing = 3

const noColour = new Set(['', 'transparent', 'rgba(0, 0, 0, 0)'])

/** The colour of the last cell of a row, or undefined when the row is empty or the cell has none. */
function lastCellBackground (row: Element): string | undefined {
  const last = row.lastElementChild

  if (last === null) {
    return undefined
  }

  const colour = getComputedStyle(last).backgroundColor

  return noColour.has(colour) ? undefined : colour
}

/**
 * The background the program has drawn at the right edge of the lower half of the terminal:
 * the colour most of those rows end in. The terminal is a whole number of cells, so a strip is
 * left beside its last column and under its last row; the page paints that strip with this
 * colour so it looks like part of the screen.
 *
 * It takes what most rows agree on, not any one cell: a highlighted menu or a button is one
 * cell with a colour of its own. A theme with no background draws none, so nothing agrees and
 * the strip keeps the terminal's own.
 */
export function terminalBackground (terminal: ParentNode): string | undefined {
  const rows = terminal.querySelector('.xterm-rows')?.children

  if (rows === undefined) {
    return undefined
  }

  const lower = [...rows].slice(Math.floor(rows.length / 2))
  const counts = new Map<string, number>()

  for (const row of lower) {
    const colour = lastCellBackground(row)

    if (colour === undefined) {
      continue
    }

    counts.set(colour, (counts.get(colour) ?? 0) + 1)
  }

  const [best, count] = [...counts].sort((a, b) => b[1] - a[1])[0] ?? [undefined, 0]

  return count >= agreeing ? best : undefined
}
