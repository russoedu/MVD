import { useLayoutEffect } from 'react'
import type { ClassicTheme } from '../classic-themes'

/**
 * Wears a theme on the page's root, so what opens outside a window (its menus, a dialog) wears it
 * too, and puts back what the root wore before when it goes: a dialog over a themed window leaves
 * the window's theme in place. It runs before the browser lays the page out: Chrome styles a scroll
 * bar's parts when it creates the scroll bar, and does not restyle them when the theme changes
 * later.
 */
export function useThemeOnRoot (theme: ClassicTheme): void {
  useLayoutEffect(() => {
    const root = document.documentElement
    const before = { theme: root.dataset.theme, family: root.dataset.family }
    root.dataset.theme = theme.id
    root.dataset.family = theme.family

    return () => {
      for (const [key, value] of Object.entries(before)) {
        if (value === undefined) delete root.dataset[key]
        else root.dataset[key] = value
      }
    }
  }, [theme])
}
