import { useLayoutEffect } from 'react'
import type { ChromeLayout } from '../classic-themes'

/**
 * Says on the page's root whether keyboard letters are underlined, so menus that open outside
 * the window follow it. On Vista they show once Alt is pressed and hide again on a click, as
 * Windows did from 2000 on; Windows 95 always shows them and a Mac never does.
 */
export function useMnemonicCues (mnemonics: ChromeLayout['mnemonics']): void {
  useLayoutEffect(() => {
    const root = document.documentElement
    const show = (shown: boolean): void => { root.dataset.mnemonics = shown ? 'shown' : 'hidden' }
    if (mnemonics !== 'on-alt') {
      show(mnemonics === 'always')

      return () => { delete root.dataset.mnemonics }
    }

    show(false)
    const onKey = (event: KeyboardEvent): void => { if (event.key === 'Alt' || event.key === 'F10') show(true) }
    const onPointer = (): void => show(false)
    document.addEventListener('keydown', onKey, { capture: true })
    document.addEventListener('pointerdown', onPointer, { capture: true })
    globalThis.addEventListener('blur', onPointer)

    return () => {
      document.removeEventListener('keydown', onKey, { capture: true })
      document.removeEventListener('pointerdown', onPointer, { capture: true })
      globalThis.removeEventListener('blur', onPointer)
      delete root.dataset.mnemonics
    }
  }, [mnemonics])
}
