import { useLayoutEffect } from 'react'
import { pixelSnapZoom } from '../classic-themes'

/**
 * Zooms the page so one CSS pixel is a whole number of screen pixels while a bitmap theme is
 * shown, and follows the screen's scale when it changes (another monitor, or the browser's zoom).
 * The zoom goes on the page's root so the menus that open outside the window match it, and is
 * also given as --pixel-zoom for the parts placed by screen position to undo.
 */
export function usePixelSnap (enabled: boolean): void {
  useLayoutEffect(() => {
    if (!enabled) return
    const root = document.documentElement
    let scale: MediaQueryList | undefined

    const snap = (): void => {
      const ratio = globalThis.devicePixelRatio
      const zoom = String(pixelSnapZoom(ratio))
      root.style.setProperty('zoom', zoom)
      root.style.setProperty('--pixel-zoom', zoom)
      scale?.removeEventListener('change', snap)
      scale = globalThis.matchMedia?.(`(resolution: ${ratio}dppx)`)
      scale?.addEventListener('change', snap)
    }
    snap()

    return () => {
      scale?.removeEventListener('change', snap)
      root.style.removeProperty('zoom')
      root.style.removeProperty('--pixel-zoom')
    }
  }, [enabled])
}
