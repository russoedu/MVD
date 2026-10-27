import { useEffect, useState } from 'react'
import { onProgramEvent, sendToProgram } from './program-events.client'

/** The events of the window style (see apps/mvd/appwindow/classic_window_client.go). */
const styleEvent = 'mvd:style'
const styleRequestEvent = 'mvd:style-request'

/**
 * The window style the settings have: 'dos' for the terminal look or a classic theme's id, or
 * null until the program has said. The page asks when it starts, as it may start after the
 * program told it, and follows each change made in the preferences.
 */
export function useWindowStyle (): string | null {
  const [style, setStyle] = useState<string | null>(null)

  useEffect(() => {
    const stop = onProgramEvent(styleEvent, value => {
      if (typeof value === 'string') setStyle(value)
    })
    sendToProgram(styleRequestEvent)

    return stop
  }, [])

  return style
}
