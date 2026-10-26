/**
 * The zoom that makes one CSS pixel a whole number of screen pixels: on a screen scaled by 125 %
 * the window is drawn at 80 % so a pixel is one screen pixel again, and at 150 % it is drawn at
 * 133 % so a pixel is two. A screen already scaled by a whole number needs none.
 */
export function pixelSnapZoom (devicePixelRatio: number): number {
  if (Number.isNaN(devicePixelRatio) || devicePixelRatio <= 0) return 1

  return Math.max(1, Math.round(devicePixelRatio)) / devicePixelRatio
}
