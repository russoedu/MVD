import './window-controls.style.css'

/** What the buttons ask of the window; the program does it (see appwindow). */
type WindowAction = 'minimise' | 'maximise' | 'close'

interface EventsRuntime {
  EventsEmit?: (name: string, ...data: unknown[]) => void
}

/** The event the program listens to for the buttons of the title bar. */
const windowEvent = 'mvd:window'

function ask (action: WindowAction): void {
  const events = Reflect.get(globalThis, 'runtime') as EventsRuntime | undefined
  events?.EventsEmit?.(windowEvent, action)
}

/**
 * The buttons of the window, in the 90s style, for a window with no frame of the system's:
 * the header is its title bar, and these are minimise, maximise and close.
 */
export function WindowControls () {
  return (
    <div className='window-controls'>
      <button type='button' className='window-control' aria-label='Minimise' title='Minimise' onClick={() => ask('minimise')}>_</button>
      <button type='button' className='window-control' aria-label='Maximise' title='Maximise' onClick={() => ask('maximise')}>□</button>
      <button type='button' className='window-control window-control-close' aria-label='Close' title='Close' onClick={() => ask('close')}>x</button>
    </div>
  )
}
