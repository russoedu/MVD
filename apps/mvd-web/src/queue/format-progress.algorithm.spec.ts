import { formatBytes, formatEta, formatSpeed } from './format-progress.algorithm'

describe('formatBytes', () => {
  it('scales to the largest unit that keeps the number small', () => {
    expect(formatBytes(512)).toBe('512 B')
    expect(formatBytes(1536)).toBe('1.5 KB')
    expect(formatBytes(15 * 1024 * 1024)).toBe('15 MB')
    expect(formatBytes(3 * 1024 ** 3)).toBe('3.0 GB')
  })

  it('is empty when the size is unknown', () => {
    expect(formatBytes(0)).toBe('')
    expect(formatBytes(-1)).toBe('')
    expect(formatBytes(NaN)).toBe('')
  })
})

describe('formatSpeed', () => {
  it('adds the rate, and stays empty when unknown', () => {
    expect(formatSpeed(2 * 1024 * 1024)).toBe('2.0 MB/s')
    expect(formatSpeed(0)).toBe('')
  })
})

describe('formatEta', () => {
  it('shows minutes and seconds, then hours', () => {
    expect(formatEta(0)).toBe('0:00')
    expect(formatEta(65)).toBe('1:05')
    expect(formatEta(3723)).toBe('1:02:03')
  })

  it('is empty for the server\'s "unknown" of -1', () => {
    expect(formatEta(-1)).toBe('')
  })
})
