import { installProgramRuntime } from './install-program-runtime.client'

const mockOn = jest.fn()
const mockEmit = jest.fn()

jest.mock('@wailsio/runtime', () => ({
  Events:  { On: (...args: unknown[]) => mockOn(...args), Emit: (...args: unknown[]) => mockEmit(...args) },
  Browser: { OpenURL: jest.fn() },
  Window:  { Size: async () => ({ width: 1100, height: 760 }), SetSize: jest.fn() },
}))

interface Runtime {
  EventsOn:   (name: string, callback: (data: unknown) => void) => void
  EventsEmit: (name: string, data: string) => void
}

afterEach(() => {
  Reflect.deleteProperty(globalThis, 'runtime')
  jest.clearAllMocks()
})

describe('installProgramRuntime', () => {
  it('leaves a page the program does not serve without a runtime', () => {
    installProgramRuntime({ hostname: 'localhost', protocol: 'http:' })

    expect(Reflect.has(globalThis, 'runtime')).toBe(false)
  })

  it('gives the page that Wails serves a runtime that hands the event data on and sends events', () => {
    installProgramRuntime({ hostname: 'wails.localhost', protocol: 'http:' })
    const runtime = Reflect.get(globalThis, 'runtime') as Runtime
    const seen: unknown[] = []
    runtime.EventsOn('mvd:style', data => { seen.push(data) })
    const [, callback] = mockOn.mock.calls[0] as [string, (event: { data: unknown }) => void]
    callback({ data: 'dos' })
    runtime.EventsEmit('mvd:window', 'close')

    expect(seen).toEqual(['dos'])
    expect(mockEmit).toHaveBeenCalledWith('mvd:window', 'close')
  })
})
