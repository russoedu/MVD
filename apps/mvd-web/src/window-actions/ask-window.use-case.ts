/** What the buttons of a title bar ask of the window; the program does it (see appwindow). */
export type WindowAction = 'minimise' | 'maximise' | 'close'

interface EventsRuntime {
  EventsEmit?: (name: string, ...data: unknown[]) => void
}

/** The event the program listens to for the buttons of the title bar. */
const windowEvent = 'mvd:window'

/** Asks the program to minimise, maximise or close MVD's window. Outside the program it does nothing. */
export function askWindow (action: WindowAction): void {
  const events = Reflect.get(globalThis, 'runtime') as EventsRuntime | undefined
  events?.EventsEmit?.(windowEvent, action)
}
