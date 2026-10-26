/** Wails' events between the page and the program; outside the program there are none. */
interface EventsRuntime {
  EventsOn?:   (name: string, callback: (...data: unknown[]) => void) => (() => void) | undefined
  EventsEmit?: (name: string, ...data: unknown[]) => void
}

const runtime = (): EventsRuntime | undefined => Reflect.get(globalThis, 'runtime') as EventsRuntime | undefined

/** Calls back with the first value of each event of this name the program sends; returns how to stop. */
export function onProgramEvent (name: string, callback: (value: unknown) => void): () => void {
  const stop = runtime()?.EventsOn?.(name, (...data) => callback(data[0]))

  return () => stop?.()
}

/** Sends the program an event. */
export function sendToProgram (name: string, ...data: unknown[]): void {
  runtime()?.EventsEmit?.(name, ...data)
}
