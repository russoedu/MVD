import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { WindowedRows } from './windowed-rows.component'
import type { WindowedRow } from './windowed-rows.component'

const ROW = 40
const rows = (count: number): WindowedRow[] => Array.from({ length: count }, (_, i) => ({ key: `r${i}`, height: ROW }))
const renderRow = (row: WindowedRow) => <span>{row.key}</span>

beforeEach(() => {
  jest.spyOn(HTMLElement.prototype, 'clientHeight', 'get').mockReturnValue(400)
})

afterEach(() => {
  jest.restoreAllMocks()
})

async function scrollTo (container: HTMLElement, top: number) {
  container.scrollTop = top
  fireEvent.scroll(container)
}

describe('WindowedRows', () => {
  it('mounts only the rows near the viewport out of thousands, inside a spacer of the full height', () => {
    render(<WindowedRows rows={rows(5000)} renderRow={renderRow} overscan={4} />)

    // 400px / 40px = 10 rows, 4 of overscan below
    expect(screen.getAllByRole('listitem')).toHaveLength(14)
    expect(screen.getByText('r0')).toBeTruthy()
    expect(screen.queryByText('r14')).toBeNull()
    expect(screen.getByRole('list').firstElementChild?.getAttribute('style')).toContain(`height: ${5000 * ROW}px`)
  })

  it('places each row at its offset', () => {
    render(<WindowedRows rows={rows(100)} renderRow={renderRow} />)

    const third = screen.getByText('r2').parentElement
    expect(third?.style.top).toBe(`${2 * ROW}px`)
    expect(third?.style.height).toBe(`${ROW}px`)
  })

  it('mounts the rows for the new position after a scroll and unmounts the ones it left', async () => {
    render(<WindowedRows rows={rows(5000)} renderRow={renderRow} overscan={2} />)

    await scrollTo(screen.getByRole('list'), 3000 * ROW)

    await waitFor(() => { expect(screen.getByText('r3000')).toBeTruthy() })
    expect(screen.queryByText('r0')).toBeNull()
    expect(screen.getAllByRole('listitem').length).toBeLessThan(20)
  })

  it('brings rows back, intact, when scrolling to the top again', async () => {
    render(<WindowedRows rows={rows(5000)} renderRow={renderRow} />)
    const list = screen.getByRole('list')

    await scrollTo(list, 4000 * ROW)
    await waitFor(() => { expect(screen.getByText('r4000')).toBeTruthy() })
    await scrollTo(list, 0)

    await waitFor(() => { expect(screen.getByText('r0')).toBeTruthy() })
    expect(screen.queryByText('r4000')).toBeNull()
  })

  it('keeps showing the last rows when the list gets shorter than the scroll position', async () => {
    const { rerender } = render(<WindowedRows rows={rows(5000)} renderRow={renderRow} overscan={0} />)
    await scrollTo(screen.getByRole('list'), 4990 * ROW)
    await waitFor(() => { expect(screen.getByText('r4990')).toBeTruthy() })

    rerender(<WindowedRows rows={rows(30)} renderRow={renderRow} overscan={0} />)

    await waitFor(() => { expect(screen.getByText('r29')).toBeTruthy() })
  })

  it('renders nothing for an empty list', () => {
    render(<WindowedRows rows={[]} renderRow={renderRow} />)

    expect(screen.queryAllByRole('listitem')).toHaveLength(0)
  })
})
