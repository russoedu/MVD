import { act } from 'react'
import { createRoot } from 'react-dom/client'
import { useDimWhileModal } from './dim-while-modal.hook'

const nothing = (): void => {}
const stopListening = (): void => {}

function Probe () {
  useDimWhileModal()

  return null
}

describe('useDimWhileModal', () => {
  let send: (open: unknown) => void = nothing
  let unmount: () => void

  beforeEach(() => {
    Reflect.set(globalThis, 'IS_REACT_ACT_ENVIRONMENT', true)
    Reflect.set(globalThis, 'runtime', {
      EventsOn: (_name: string, callback: (open: unknown) => void) => {
        send = callback

        return stopListening
      },
    })
    const root = createRoot(document.createElement('div'))
    act(() => root.render(<Probe />))
    unmount = () => act(() => root.unmount())
  })

  afterEach(() => {
    unmount()
    Reflect.deleteProperty(globalThis, 'runtime')
  })

  it('puts the page behind a popup while one is open and brings it back after', () => {
    send(true)
    expect(document.body.classList.contains('modal-open')).toBe(true)
    send(false)
    expect(document.body.classList.contains('modal-open')).toBe(false)
  })

  it('takes the dimming away when the page goes', () => {
    send(true)
    unmount()
    expect(document.body.classList.contains('modal-open')).toBe(false)
    unmount = nothing
  })
})
