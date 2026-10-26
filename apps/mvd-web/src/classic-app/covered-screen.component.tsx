import { useLayoutEffect, useRef } from 'react'
import type { ChromeLayout } from '../classic-themes'
import { TitleBar } from '../classic-window'

interface CoveredScreenProps {
  title:   string
  layout:  ChromeLayout
  onClose: () => void
}

/** The places of the terminal while it is inset, as the page's style reads them. */
const insets = ['top', 'left', 'width', 'height'] as const

/**
 * A dialog of the theme for a screen only the terminal look draws so far (the preferences, About,
 * a question): its body is a hole the terminal is shown through, kept in step with the dialog's
 * size. Its close button closes the screen as Esc does.
 */
export function CoveredScreen ({ title, layout, onClose }: CoveredScreenProps) {
  const bodyRef = useRef<HTMLDivElement>(null)

  useLayoutEffect(() => {
    const element = bodyRef.current
    if (!element) return
    const root = document.documentElement
    const place = (): void => {
      // The page may be zoomed to keep a bitmap theme sharp: the box is measured on the screen,
      // and the terminal is placed in the page's own pixels.
      const zoom = Number(getComputedStyle(root).getPropertyValue('--pixel-zoom')) || 1
      const box = element.getBoundingClientRect()
      const values = { top: box.top / zoom, left: box.left / zoom, width: box.width / zoom, height: box.height / zoom }
      for (const inset of insets) root.style.setProperty(`--terminal-inset-${inset}`, `${values[inset]}px`)
    }
    place()
    // The dialog's own size, and the window's, which moves the dialog it is centred in.
    const watch = new ResizeObserver(place)
    watch.observe(element)
    watch.observe(root)

    return () => {
      watch.disconnect()
      for (const inset of insets) root.style.removeProperty(`--terminal-inset-${inset}`)
    }
  }, [])

  return (
    <div className='classic-covered-overlay'>
      <section className='classic-covered' aria-label={title}>
        <TitleBar title={title} layout={layout} controls={['close']} onWindow={onClose} />
        <div ref={bodyRef} className='classic-covered-body' />
      </section>
    </div>
  )
}
