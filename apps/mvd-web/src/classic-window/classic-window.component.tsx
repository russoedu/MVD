import { useLayoutEffect, useState, type ReactNode } from 'react'
import { layoutOf, type ClassicTheme, type ChromeLayout, type WindowControl } from '../classic-themes'
import type { WorkspaceIntent, WorkspaceSnapshot } from '../workspace-snapshot'
import { MenuBar } from './menu-bar.component'
import { StatusLine } from './status-line.component'
import { TitleBar } from './title-bar.component'
import { ActionsToolbar } from './actions-toolbar.component'

interface ClassicWindowProps {
  theme:    ClassicTheme
  snapshot: WorkspaceSnapshot
  onIntent: (intent: WorkspaceIntent) => void
  /** The panes inside the window. */
  children: ReactNode
  /**
   * What is drawn over the window's inside, such as a dialog, given the element to draw it in and
   * the family's layout.
   */
  overlay?: (container: Element, layout: ChromeLayout) => ReactNode
}

/**
 * MVD's main window in a classic theme: its frame, title bar, menu bar, toolbar and status line,
 * around the panes it is given. The theme's tokens are set on the page's root, so menus that open
 * outside the window wear the theme too.
 */
export function ClassicWindow ({ theme, snapshot, onIntent, children, overlay }: ClassicWindowProps) {
  const layout = layoutOf(theme.family)
  const [inside, setInside] = useState<HTMLDivElement | null>(null)

  // Before the browser lays the panes out: Chrome styles a scroll bar's parts when it creates the
  // scroll bar, and does not restyle them when the theme changes later.
  useLayoutEffect(() => {
    const root = document.documentElement
    root.dataset.theme = theme.id
    root.dataset.family = theme.family

    return () => {
      delete root.dataset.theme
      delete root.dataset.family
    }
  }, [theme])

  const onWindow = (action: WindowControl) => onIntent({ kind: 'window', action })

  return (
    <div className='classic-window' data-family={theme.family} data-busy={snapshot.busy || undefined}>
      <TitleBar title={snapshot.title} layout={layout} onWindow={onWindow} />
      <MenuBar menus={snapshot.menus} layout={layout} onEntry={id => onIntent({ kind: 'menu', id })} />
      <ActionsToolbar buttons={snapshot.toolbar} onButton={id => onIntent({ kind: 'button', id })} />
      {/* Drawn again for each theme, so the scroll bars are made in the new theme's style. */}
      <div key={theme.id} ref={setInside} className='classic-window-inside'>
        <div className='classic-window-panes'>{children}</div>
        {inside && overlay?.(inside, layout)}
      </div>
      <StatusLine status={snapshot.status} layout={layout} />
    </div>
  )
}
