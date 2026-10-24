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

describe('terminalBackground', () => {
  it('is the colour of the last cell of the lowest row', () => {
    const terminal = terminalWith([['#111111'], ['#222222', '#2a2a2a'], ['#333333', '#444444']])

    expect(terminalBackground(terminal)).toBe('rgb(68, 68, 68)')
  })

  it('passes over rows with nothing drawn in them', () => {
    const terminal = terminalWith([['#2a2a2a'], [], []])

    expect(terminalBackground(terminal)).toBe('rgb(42, 42, 42)')
  })

  it('passes over a cell with no colour of its own', () => {
    const terminal = terminalWith([['#2a2a2a'], ['transparent']])

    expect(terminalBackground(terminal)).toBe('rgb(42, 42, 42)')
  })

  it('is nothing before the terminal has drawn anything', () => {
    expect(terminalBackground(document.createElement('div'))).toBeUndefined()
    expect(terminalBackground(terminalWith([[], []]))).toBeUndefined()
  })
})
