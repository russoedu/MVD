import { rowOffsets, visibleRange } from './visible-range.algorithm'

const uniform = (count: number, height = 50) => rowOffsets(Array.from({ length: count }, () => height))

describe('rowOffsets', () => {
  it('is the running total of the heights, starting at zero', () => {
    expect(rowOffsets([10, 20, 30])).toEqual([0, 10, 30, 60])
    expect(rowOffsets([])).toEqual([0])
  })
})

describe('visibleRange', () => {
  it('mounts nothing for an empty list', () => {
    expect(visibleRange(rowOffsets([]), 0, 500, 5)).toEqual({ start: 0, end: 0 })
  })

  it('mounts everything when the rows are fewer than the viewport holds', () => {
    expect(visibleRange(uniform(3), 0, 500, 5)).toEqual({ start: 0, end: 3 })
  })

  it('mounts the top rows plus the overscan when scrolled to the top', () => {
    // 500px shows rows 0-9; two rows of overscan below
    expect(visibleRange(uniform(1000), 0, 500, 2)).toEqual({ start: 0, end: 12 })
  })

  it('follows the scroll position into the middle, with overscan on both sides', () => {
    // top at row 100, rows 100-109 visible
    expect(visibleRange(uniform(1000), 5000, 500, 3)).toEqual({ start: 97, end: 113 })
  })

  it('includes a row that is only partly visible at each edge', () => {
    // top is 25px into row 2; bottom is 25px into row 12
    expect(visibleRange(uniform(1000), 125, 500, 0)).toEqual({ start: 2, end: 13 })
  })

  it('stops at the last row when scrolled to the end', () => {
    expect(visibleRange(uniform(1000), 49500, 500, 4)).toEqual({ start: 986, end: 1000 })
  })

  it('treats a scroll position past the end as the end, for a list that just got shorter', () => {
    expect(visibleRange(uniform(20), 99999, 500, 0)).toEqual({ start: 10, end: 20 })
  })

  it('treats a negative scroll position as the top', () => {
    expect(visibleRange(uniform(100), -300, 200, 1)).toEqual({ start: 0, end: 5 })
  })

  it('works with rows of different heights', () => {
    // rows start at 0, 100, 130, 230, 260; 120-240 touches rows 1, 2 and 3
    const offsets = rowOffsets([100, 30, 100, 30, 100])
    expect(visibleRange(offsets, 120, 120, 0)).toEqual({ start: 1, end: 4 })
  })

  it('always mounts at least one row, even for a viewport that has no height yet', () => {
    expect(visibleRange(uniform(10), 0, 0, 0)).toEqual({ start: 0, end: 1 })
  })

  it('keeps a valid range when the number of rows changes between two calls', () => {
    const before = visibleRange(uniform(5000), 100000, 500, 5)
    const after = visibleRange(uniform(40), 100000, 500, 5)

    expect(before.end).toBeLessThanOrEqual(5000)
    expect(after.start).toBeGreaterThanOrEqual(0)
    expect(after.end).toBe(40)
    expect(after.start).toBeLessThan(after.end)
  })
})
