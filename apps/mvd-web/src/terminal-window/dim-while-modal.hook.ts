import { useEffect } from 'react'

/** The event the program sends when a popup opens or closes over the terminal. */
const modalEvent = 'mvd:modal'

interface EventsRuntime {
  EventsOn?: (name: string, callback: (open: unknown) => void) => (() => void) | undefined
}

/**
 * Dims the page around the terminal while the program has a popup open (the preferences,
 * the About window), as the program dims the terminal itself. The page only has to wear
 * the `modal-open` class; the style does the rest.
 */
export function useDimWhileModal (): void {
  useEffect(() => {
    const events = Reflect.get(globalThis, 'runtime') as EventsRuntime | undefined
    const stop = events?.EventsOn?.(modalEvent, open => {
      document.body.classList.toggle('modal-open', open === true)
    })

    return () => {
      stop?.()
      document.body.classList.remove('modal-open')
    }
  }, [])
}
