import { lazy, Suspense, useCallback, type ComponentProps } from 'react'
import type { ClassicApp } from '../classic-app'
import { sendTerminalInput, TerminalWindow, type TerminalLook } from '../terminal-window'
import { askWindow } from '../window-actions'
import type { WorkspaceIntent } from '../workspace-snapshot'
import { useLivePreferences } from './live-preferences.hook'
import { useLiveSnapshot } from './live-snapshot.hook'
import { useWindowStyle } from './window-style.hook'

type ClassicLookProps = Omit<ComponentProps<typeof ClassicApp>, 'theme'> & { style: string }

// Loaded only once a classic style is chosen, so the terminal look does not carry the themes.
const ClassicLook = lazy(async () => {
  const [{ ClassicApp: App }, { classicThemeOf }] = await Promise.all([import('../classic-app'), import('../classic-themes')])

  function Look ({ style, ...rest }: ClassicLookProps) {
    return <App {...rest} theme={classicThemeOf(style)} />
  }

  return { default: Look }
})

/** Esc, as the terminal reads it: it closes the screen the terminal shows. */
const escape = '\u{1B}'

/**
 * MVD's window, in the style the preferences say: the terminal look, or a classic system's window
 * drawn from what the program sends. The terminal stays mounted either way, so MVD keeps running
 * behind a classic window and shows through it for the screens the themes do not draw yet. The
 * preferences are a window of their own (see PreferencesApp); while it is open this one cannot be used.
 */
export function AppWindow () {
  const chosen = useWindowStyle()
  const style = chosen?.style ?? null
  const scheme = chosen?.scheme || undefined
  const [snapshot, send] = useLiveSnapshot()
  const [prefs] = useLivePreferences()
  const classic = style !== null && style !== 'dos' && snapshot !== null
  let look: TerminalLook = 'page'
  if (classic) look = snapshot.covered ? 'inset' : 'behind'

  const onIntent = useCallback((intent: WorkspaceIntent) => {
    if (intent.kind === 'window') {
      askWindow(intent.action)

      return
    }
    send(intent)
  }, [send])

  return (
    <div className='app-window' inert={prefs !== null}>
      <TerminalWindow look={look} />
      {classic && (
        <div className='app-window-classic'>
          <Suspense>
            <ClassicLook style={style} scheme={scheme} snapshot={snapshot} onIntent={onIntent} onCloseCovered={() => sendTerminalInput(escape)} />
          </Suspense>
        </div>
      )}
    </div>
  )
}
