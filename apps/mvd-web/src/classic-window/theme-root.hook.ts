import { useLayoutEffect } from 'react'
import type { ClassicTheme } from '../classic-themes'

/** A window's claim on the page's root: the theme it wears, and its colour scheme. */
interface Claim {
  theme:   ClassicTheme
  scheme?: string
}

/**
 * The windows that wear a theme now, oldest first. The root wears the newest one's: a dialog over
 * a window wears the dialog's, and once it goes the window's comes back, whichever of them was
 * opened or changed first.
 */
const claims: Claim[] = []

/** Puts the newest claim's theme on the root, or none when no window has one. */
function wearNewest (): void {
  const root = document.documentElement
  const [newest] = claims.slice(-1) as (Claim | undefined)[]
  if (newest) {
    root.dataset.theme = newest.theme.id
    root.dataset.family = newest.theme.family
    if (newest.scheme) root.dataset.scheme = newest.scheme
    else delete root.dataset.scheme
  } else {
    delete root.dataset.theme
    delete root.dataset.family
    delete root.dataset.scheme
  }
}

/**
 * Wears a theme, and its colour scheme, on the page's root while the window is open, so what opens
 * outside it (its menus, a dialog) wears them too. It runs before the browser lays the page out:
 * Chrome styles a scroll bar's parts when it creates the scroll bar, and does not restyle them when
 * the theme changes later.
 */
export function useThemeOnRoot (theme: ClassicTheme, scheme?: string): void {
  useLayoutEffect(() => {
    const claim: Claim = { theme, scheme }
    claims.push(claim)
    wearNewest()

    return () => {
      claims.splice(claims.indexOf(claim), 1)
      wearNewest()
    }
  }, [theme, scheme])
}
