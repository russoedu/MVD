import { lazy, Suspense, type ComponentProps } from 'react'
import type { PreferencesWindow } from '../preferences-window'
import { useLivePreferences } from './live-preferences.hook'
import { useWindowStyle } from './window-style.hook'

type PreferencesLookProps = Omit<ComponentProps<typeof PreferencesWindow>, 'theme' | 'ownWindow'> & { style: string }

// Loaded when the window opens, as it is the only thing the window shows.
const PreferencesLook = lazy(async () => {
  const [{ PreferencesWindow: Window }, { windowThemeOf }] = await Promise.all([import('../preferences-window'), import('../classic-themes')])

  function Look ({ style, ...rest }: PreferencesLookProps) {
    return <Window {...rest} theme={windowThemeOf(style)} ownWindow />
  }

  return { default: Look }
})

/**
 * The page of the preferences window: the program opens a window of its own for the preferences and
 * this page fills it, in the window style the settings have. It draws what the program sends and
 * sends back what the person does; the program closes the window when the preferences close.
 */
export function PreferencesApp () {
  const chosen = useWindowStyle()
  const [prefs, sendPrefs] = useLivePreferences()
  if (prefs === null || chosen === null) return null

  return (
    <Suspense>
      <PreferencesLook style={chosen.style} scheme={chosen.scheme || undefined} prefs={prefs} onIntent={sendPrefs} />
    </Suspense>
  )
}
