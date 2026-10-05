import { useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react'
import type { ReactNode } from 'react'
import { rowOffsets, visibleRange } from './visible-range.algorithm'

export interface WindowedRow {
  key:    string
  height: number
}

interface WindowedRowsProps<Row extends WindowedRow> {
  rows:       readonly Row[]
  renderRow:  (row: Row) => ReactNode
  className?: string
  /** Rows kept mounted beyond each edge of the viewport. */
  overscan?:  number
}

/** Used until the container has been measured (and where there is no layout at all). */
const DEFAULT_VIEWPORT = 600

/**
 * A scrolling list that mounts only the rows near the viewport. The rest are not in the
 * page at all: a spacer gives the scrollbar the full height, and rows are placed by their
 * offsets, so thousands of rows cost the same as a screenful.
 */
export function WindowedRows<Row extends WindowedRow> ({ rows, renderRow, className, overscan = 8 }: WindowedRowsProps<Row>) {
  const containerRef = useRef<HTMLDivElement>(null)
  const [scrollTop, setScrollTop] = useState(0)
  const [viewport, setViewport] = useState(DEFAULT_VIEWPORT)

  const offsets = useMemo(() => rowOffsets(rows.map(row => row.height)), [rows])
  const range = visibleRange(offsets, scrollTop, viewport, overscan)

  useLayoutEffect(() => {
    const element = containerRef.current
    if (!element) return
    const measure = () => setViewport(element.clientHeight || DEFAULT_VIEWPORT)
    measure()
    if (typeof ResizeObserver === 'undefined') return
    const observer = new ResizeObserver(measure)
    observer.observe(element)

    return () => observer.disconnect()
  }, [])

  useEffect(() => {
    const element = containerRef.current
    if (!element) return
    let frame = 0
    const onScroll = () => {
      if (frame !== 0) return
      frame = requestAnimationFrame(() => {
        frame = 0
        setScrollTop(element.scrollTop)
      })
    }
    element.addEventListener('scroll', onScroll, { passive: true })

    return () => {
      element.removeEventListener('scroll', onScroll)
      if (frame !== 0) cancelAnimationFrame(frame)
    }
  }, [])

  const mounted: ReactNode[] = []
  for (let index = range.start; index < range.end; index += 1) {
    const row = rows[index]
    mounted.push(
      <div
        key={row.key}
        className='windowed-row'
        role='listitem'
        style={{ top: offsets[index], height: row.height }}
      >
        {renderRow(row)}
      </div>,
    )
  }

  return (
    <div ref={containerRef} className={className} role='list'>
      <div className='windowed-rows' style={{ height: offsets[rows.length] }}>
        {mounted}
      </div>
    </div>
  )
}
