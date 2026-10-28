import { Browser, Events, Window } from '@wailsio/runtime'

/**
 * The page talks to the program through a small `runtime` object (events, the window's size,
 * opening an address), the shape the terminal socket of treactui also asks for. Wails 3 has
 * its runtime as a module, so this builds that object from it.
 */
interface ProgramRuntime {
  EventsOn:       (name: string, callback: (...data: unknown[]) => void) => () => void
  EventsEmit:     (name: string, ...data: unknown[]) => void
  BrowserOpenURL: (url: string) => void
  WindowGetSize:  () => Promise<{ w: number, h: number }>
  WindowSetSize:  (width: number, height: number) => void
}

/** Only a page that Wails serves has a program behind it; Wails serves it from its own origin. */
function isProgramPage (place: Pick<Location, 'hostname' | 'protocol'>): boolean {
  return place.hostname === 'wails.localhost' || place.protocol === 'wails:'
}

/** Gives the page the `runtime` object when the program serves it; elsewhere there is none. */
export function installProgramRuntime (place: Pick<Location, 'hostname' | 'protocol'> = globalThis.location): void {
  if (!isProgramPage(place) || Reflect.has(globalThis, 'runtime')) return

  const runtime: ProgramRuntime = {
    EventsOn:       (name, callback) => Events.On(name, event => { callback(event.data) }),
    EventsEmit:     (name, ...data) => { void Events.Emit(name, data[0]) },
    BrowserOpenURL: url => { void Browser.OpenURL(url) },
    WindowGetSize:  async () => {
      const { width, height } = await Window.Size()

      return { w: width, h: height }
    },
    WindowSetSize: (width, height) => { void Window.SetSize(width, height) },
  }
  Object.defineProperty(globalThis, 'runtime', { value: runtime, configurable: true })
}
