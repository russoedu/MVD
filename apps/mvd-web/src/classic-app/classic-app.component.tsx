import type { ReactNode } from 'react'
import { AddMusicWindow } from '../add-music-window'
import { layoutOf, type ClassicTheme } from '../classic-themes'
import { DetailsPane, SongsTable, SourcesPane } from '../classic-panes'
import { ClassicWindow } from '../classic-window'
import type { WorkspaceIntent, WorkspaceSnapshot } from '../workspace-snapshot'
import { CoveredScreen } from './covered-screen.component'
import { useKeyboardIntents } from './keyboard-intents.hook'

interface ClassicAppProps {
  theme:           ClassicTheme
  snapshot:        WorkspaceSnapshot
  onIntent:        (intent: WorkspaceIntent) => void
  /** Closes a screen only the terminal draws, as Esc does there; nil where none is drawn. */
  onCloseCovered?: () => void
}

/**
 * MVD's main window in a classic theme, whole: the window with its sources, songs and details,
 * the window for adding music when it is open, and a dialog the terminal shows through for a
 * screen the themes do not draw yet. It draws what the snapshot says and sends back what the
 * person does; MVD decides the rest.
 */
export function ClassicApp ({ theme, snapshot, onIntent, onCloseCovered }: ClassicAppProps) {
  const { covered, adding } = snapshot
  useKeyboardIntents(snapshot.menus, onIntent, !covered && !adding)

  return (
    <>
      <ClassicWindow
        theme={theme}
        snapshot={snapshot}
        onIntent={onIntent}
        overlay={(container, layout): ReactNode => adding && (
          <AddMusicWindow
            adding={adding}
            layout={layout}
            container={container}
            onMode={mode => onIntent({ kind: 'adding-mode', mode })}
            onText={(field, text) => onIntent({ kind: 'adding-text', field, text })}
            onButton={id => onIntent({ kind: 'adding-button', id })}
          />
        )}
      >
        <SourcesPane sources={snapshot.sources} onSelect={index => onIntent({ kind: 'select-source', index })} />
        <SongsTable
          songs={snapshot.songs}
          onSelect={index => onIntent({ kind: 'select-song', index })}
          onTick={(index, ticked) => onIntent({ kind: 'tick-song', index, ticked })}
        />
        <DetailsPane details={snapshot.details} />
      </ClassicWindow>
      {covered && <CoveredScreen title={covered.title} layout={layoutOf(theme)} onClose={() => onCloseCovered?.()} />}
    </>
  )
}
