import { useLayoutEffect } from 'react'
import { Button, Dialog, Modal, ModalOverlay, Tab, TabList, TabPanel, Tabs } from 'react-aria-components'
import { layoutOf, orderDialogButtons, type ClassicTheme } from '../classic-themes'
import { TitleBar, useThemeOnRoot } from '../classic-window'
import type { ButtonSnapshot } from '../workspace-snapshot'
import { PreferenceField } from './preference-field.component'
import type { PreferencesIntent, PreferencesSnapshot, TerminalColours } from './preferences-window.contract'

interface PreferencesWindowProps {
  theme:    ClassicTheme
  /** The theme's colour scheme; none for its first, the system's own. */
  scheme?:  string
  prefs:    PreferencesSnapshot
  onIntent: (intent: PreferencesIntent) => void
}

/** The buttons at the bottom: Save keeps the settings, Close leaves them. */
const buttons: ButtonSnapshot[] = [
  { id: 'save', label: 'Save', role: 'default', enabled: true },
  { id: 'close', label: 'Close', role: 'cancel', enabled: true },
]

/** Wears the terminal look's colours on the page's root while MVD's own look draws the window. */
function useTerminalColours (theme: ClassicTheme, colours: TerminalColours): void {
  useLayoutEffect(() => {
    if (theme.family !== 'dos') return
    const root = document.documentElement
    const entries = Object.entries(colours) as [keyof TerminalColours, string][]
    const names = entries.map(([name]) => name)
    for (const [name, colour] of entries) {
      // An empty colour is the terminal's own, which the style's default stands in for.
      if (colour === '') root.style.removeProperty(`--dos-${name}`)
      else root.style.setProperty(`--dos-${name}`, colour)
    }

    return () => {
      for (const name of names) root.style.removeProperty(`--dos-${name}`)
    }
  }, [theme, colours])
}

/**
 * The preferences, as a window of the theme: its tabs, the settings of the tab shown, what
 * happened last, and Save and Close where the system puts them. It draws what the program sends
 * and sends back each change; the program keeps the settings and runs what a button asks (the
 * folder chooser, moving or removing the app).
 */
export function PreferencesWindow ({ theme, scheme, prefs, onIntent }: PreferencesWindowProps) {
  useThemeOnRoot(theme, scheme)
  useTerminalColours(theme, prefs.colours)
  const layout = layoutOf(theme)
  const close = (): void => onIntent({ kind: 'prefs-close' })

  return (
    <ModalOverlay className='preferences-overlay' isOpen isDismissable={false} onOpenChange={open => { if (!open) close() }}>
      <Modal className='preferences-window' data-family={theme.family}>
        <Dialog className='preferences-dialog' aria-label={prefs.title}>
          <TitleBar title={prefs.title} layout={layout} controls={['close']} onWindow={close} />
          <div className='preferences-body' inert={prefs.busy}>
            {prefs.welcome && <p className='preferences-welcome'>{prefs.welcome}</p>}
            <Tabs
              className='preferences-tabs'
              selectedKey={String(prefs.tab)}
              onSelectionChange={key => onIntent({ kind: 'prefs-tab', index: Number(key) })}
            >
              <TabList className='preferences-tab-list' aria-label='Pages'>
                {prefs.tabs.map((tab, index) => <Tab key={tab} id={String(index)} className='preferences-tab'>{tab}</Tab>)}
              </TabList>
              <TabPanel id={String(prefs.tab)} className='preferences-page'>
                {prefs.fields.map((field, index) => field.hidden ? null : <PreferenceField key={field.label} field={field} index={index} onIntent={onIntent} />)}
              </TabPanel>
            </Tabs>
            <div className='preferences-footer'>
              <p className='preferences-status' data-bad={prefs.bad || undefined} role='status'>{prefs.status}</p>
              <div className='preferences-buttons'>
                {orderDialogButtons(buttons, layout).map(button => (
                  <Button
                    key={button.id}
                    className='classic-button'
                    data-role={button.role}
                    onPress={() => onIntent(button.id === 'save' ? { kind: 'prefs-save' } : { kind: 'prefs-close' })}
                  >
                    {button.label}
                  </Button>
                ))}
              </div>
            </div>
          </div>
        </Dialog>
      </Modal>
    </ModalOverlay>
  )
}
