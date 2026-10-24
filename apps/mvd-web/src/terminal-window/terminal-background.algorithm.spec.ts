import { terminalBackground } from './terminal-background.algorithm'

/** A terminal as xterm draws it: a row per line, a span per run of cells with one colour. */
function terminalWith (rows: string[][]): HTMLElement {
  const terminal = document.createElement('div')
  const container = document.createElement('div')
  container.className = 'xterm-rows'

  for (const spans of rows) {
    const row = document.createElement('div')

    for (const background of spans) {
      const span = document.createElement('span')
      span.style.backgroundColor = background
      row.append(span)
    }

    container.append(row)
  }
  terminal.append(container)

  return terminal
}

const backgroundOf = (rows: string[][]): string | undefined => terminalBackground(terminalWith(rows))

const screen = '#2a2a2a'
const rowsOf = (count: number, spans: string[]): string[][] => Array.from({ length: count }, () => spans)
const blank = (count: number): string[][] => rowsOf(count, [])
const painted = (count: number, colour = screen): string[][] => rowsOf(count, ['#111111', colour])

describe('terminalBackground', () => {
  it('is the colour most of the lower rows end in', () => {
    expect(backgroundOf([...blank(2), ...painted(8)])).toBe('rgb(42, 42, 42)')
  })

  it('is not fooled by one highlighted cell, such as a menu that is lit', () => {
    const rows = [['#222222', '#ff87d7'], ...blank(7), ['#222222', '#ff87d7']]

    expect(backgroundOf(rows)).toBeUndefined()
  })

  it('looks at the lower half, so a lit bar at the top does not count', () => {
    const rows = [['#ff87d7'], ['#ff87d7'], ['#ff87d7'], ...blank(7)]

    expect(backgroundOf(rows)).toBeUndefined()
  })

  it('passes over a cell with no colour of its own', () => {
    const rows = [...painted(3), ...rowsOf(2, ['transparent']), ...painted(3)]

    expect(backgroundOf(rows)).toBe('rgb(42, 42, 42)')
  })

  it('takes the colour with the most rows when two are drawn', () => {
    expect(backgroundOf([...painted(2), ...painted(8, '#101820')])).toBe('rgb(16, 24, 32)')
  })

  it('is nothing before the terminal has drawn anything, or with a theme that has no background', () => {
    expect(terminalBackground(document.createElement('div'))).toBeUndefined()
    expect(backgroundOf([[], []])).toBeUndefined()
    expect(backgroundOf(rowsOf(10, ['transparent']))).toBeUndefined()
  })
})
