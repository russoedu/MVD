import type { ReactNode } from 'react'
import type { ChromeLayout, WindowControl } from '../classic-themes'

interface TitleBarProps {
  title:     ReactNode
  layout:    ChromeLayout
  /** The buttons to show, when not all of the family's: a dialog has only Close. */
  controls?: readonly WindowControl[]
  onWindow:  (action: WindowControl) => void
}

const names: Record<WindowControl, string> = {
  minimise: 'Minimise',
  maximise: 'Maximise',
  close:    'Close',
}

const glyphs: Record<WindowControl, string> = {
  minimise: '_',
  maximise: '□',
  close:    '×',
}

/** The title bar: the window's name and its buttons, where the family puts them. */
export function TitleBar ({ title, layout, controls: shown, onWindow }: TitleBarProps) {
  const wanted = shown ?? layout.controls
  const controls = (
    <div className='classic-window-controls'>
      {layout.controls.filter(control => wanted.includes(control)).map(control => (
        <button
          key={control}
          type='button'
          className='classic-window-control'
          data-control={control}
          aria-label={names[control]}
          title={names[control]}
          onClick={() => onWindow(control)}
        >
          <span aria-hidden='true'>{glyphs[control]}</span>
        </button>
      ))}
    </div>
  )

  return (
    <div
      className='classic-title-bar'
      data-align={layout.titleAlign}
      onDoubleClick={event => {
        if ((event.target as HTMLElement).closest('button')) return
        onWindow(layout.titleDoubleClick)
      }}
    >
      {layout.controlsSide === 'left' && controls}
      {typeof title === 'string' ? <span className='classic-title'>{title}</span> : title}
      {layout.controlsSide === 'right' && controls}
    </div>
  )
}
