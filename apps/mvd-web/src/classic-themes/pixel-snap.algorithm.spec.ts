import { pixelSnapZoom } from './pixel-snap.algorithm'

describe('pixelSnapZoom', () => {
  it('leaves a screen scaled by a whole number alone', () => {
    expect(pixelSnapZoom(1)).toBe(1)
    expect(pixelSnapZoom(2)).toBe(1)
  })

  it('makes a pixel one or two screen pixels on a screen scaled in between', () => {
    expect(pixelSnapZoom(1.25) * 1.25).toBeCloseTo(1)
    expect(pixelSnapZoom(1.5) * 1.5).toBeCloseTo(2)
    expect(pixelSnapZoom(1.75) * 1.75).toBeCloseTo(2)
  })

  it('never shrinks below one screen pixel, and ignores a ratio it cannot use', () => {
    expect(pixelSnapZoom(0.9) * 0.9).toBeCloseTo(1)
    expect(pixelSnapZoom(0)).toBe(1)
    expect(pixelSnapZoom(NaN)).toBe(1)
  })
})
