import { cellAt, shiftClickSequence, type CellGrid } from './shift-click.algorithm'

const grid: CellGrid = { left: 10, top: 20, cellWidth: 8, cellHeight: 18, cols: 100, rows: 30 }

describe('cellAt', () => {
  it('counts cells from 1, from the corner of the terminal', () => {
    expect(cellAt({ x: 10, y: 20 }, grid)).toEqual({ col: 1, row: 1 })
    expect(cellAt({ x: 10 + 8 * 4 + 3, y: 20 + 18 * 2 + 5 }, grid)).toEqual({ col: 5, row: 3 })
  })

  it('is nothing outside the terminal', () => {
    expect(cellAt({ x: 9, y: 25 }, grid)).toBeNull()
    expect(cellAt({ x: 20, y: 19 }, grid)).toBeNull()
    expect(cellAt({ x: 10 + 8 * 100, y: 25 }, grid)).toBeNull()
    expect(cellAt({ x: 20, y: 20 + 18 * 30 }, grid)).toBeNull()
  })

  it('is nothing before the size of a cell is known', () => {
    expect(cellAt({ x: 20, y: 25 }, { ...grid, cellWidth: 0 })).toBeNull()
  })
})

describe('shiftClickSequence', () => {
  it('presses and releases the left button with Shift', () => {
    expect(shiftClickSequence({ col: 5, row: 3 })).toBe('\u{1B}[<4;5;3M\u{1B}[<4;5;3m')
  })
})
