/** Rows that should be mounted: `start` is the first, `end` is one past the last. */
export interface RowRange {
  start: number
  end:   number
}

/** Prefix sums of the row heights: offsets[i] is where row i starts, the last is the total. */
export function rowOffsets (heights: readonly number[]): number[] {
  const offsets = [0]
  let sum = 0
  for (const height of heights) {
    sum += height
    offsets.push(sum)
  }

  return offsets
}

/** The row that contains the vertical position y (the last row for anything past the end). */
function rowAt (offsets: readonly number[], y: number): number {
  let low = 0
  let high = offsets.length - 2
  while (low < high) {
    const middle = (low + high + 1) >> 1
    if (offsets[middle] <= y) low = middle
    else high = middle - 1
  }

  return low
}

/**
 * The rows to mount for a scroll position: those that touch the viewport plus `overscan`
 * rows on each side. A scroll position past the end is treated as the end, so a list that
 * just got shorter still shows its last rows.
 */
export function visibleRange (
  offsets: readonly number[],
  scrollTop: number,
  viewportHeight: number,
  overscan: number,
): RowRange {
  const count = offsets.length - 1
  if (count <= 0) return { start: 0, end: 0 }
  const total = offsets[count]
  const viewport = Math.max(viewportHeight, 1)
  const top = Math.min(Math.max(scrollTop, 0), Math.max(total - viewport, 0))
  const first = rowAt(offsets, top)
  const last = rowAt(offsets, Math.max(top + viewport - 1, top))

  return {
    start: Math.max(first - overscan, 0),
    end:   Math.min(last + 1 + overscan, count),
  }
}
