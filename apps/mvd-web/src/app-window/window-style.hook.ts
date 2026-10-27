import { useEffect, useState } from 'react'
import { onProgramEvent, sendToProgram } from './program-events.client'

/** The events of the window style (see apps/mvd/appwindow/classic_window_client.go). */
const styleEvent = 'mvd:style'
const styleRequestEvent = 'mvd:style-request'

/** The window style of the settings and its colour scheme ("" for a style with none). */
export interface WindowStyle {
  style:  string
  scheme: string
}

/** What the program sends: the style and scheme as JSON, or (from an older program) the style alone. */
function readStyle (value: unknown): WindowStyle | null {
  if (typeof value !== 'string') return null
  try {
    const parsed = JSON.parse(value) as Partial<WindowStyle>
    if (typeof parsed.style === 'string') return { style: parsed.style, scheme: parsed.scheme ?? '' }
  } catch {
    // Not JSON: the style alone.
  }

  return { style: value, scheme: '' }
}

/**
 * The window style the settings have ('dos' for the terminal look or a classic theme's id) with
 * its colour scheme, or null until the program has said. The page asks when it starts, as it may
 * start after the program told it, and follows each change made in the preferences.
 */
export function useWindowStyle (): WindowStyle | null {
  const [style, setStyle] = useState<WindowStyle | null>(null)

  useEffect(() => {
    const stop = onProgramEvent(styleEvent, value => {
      const read = readStyle(value)
      if (read) setStyle(read)
    })
    sendToProgram(styleRequestEvent)

    return stop
  }, [])

  return style
}
