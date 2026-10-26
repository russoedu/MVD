import type { ReactNode } from 'react'
import type { ChromeLayout, WindowControl } from '../classic-themes'

interface TitleBarProps {
  title:     ReactNode
  layout:    ChromeLayout
  /** The buttons to show, when not all of the theme's: a dialog has only Close. */
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

/** The title bar: the window's name and its buttons, on the sides the theme puts them. */
export function TitleBar ({ title, layout, controls: shown, onWindow }: TitleBarProps) {
  const group = (side: 'left' | 'right', controls: readonly WindowControl[]): ReactNode => {
    const wanted = controls.filter(control => !shown || shown.includes(control))
    if (wanted.length === 0) return null

    return (
      <div className='classic-window-controls' data-side={side}>
        {wanted.map(control => (
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
  }

  return (
    <div
      className='classic-title-bar'
      data-align={layout.titleAlign}
      onDoubleClick={event => {
        if (!layout.titleDoubleClick || (event.target as HTMLElement).closest('button')) return
        onWindow(layout.titleDoubleClick)
      }}
    >
      {group('left', layout.controlsLeft)}
      {typeof title === 'string' ? <span className='classic-title'>{title}</span> : title}
      {group('right', layout.controlsRight)}
    </div>
  )
}
