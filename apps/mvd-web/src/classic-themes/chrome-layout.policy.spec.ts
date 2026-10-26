import type { ButtonSnapshot } from '../workspace-snapshot'
import { classicThemeOf } from './classic-theme.model'
import { layoutOf, orderDialogButtons } from './chrome-layout.policy'

const buttons: ButtonSnapshot[] = [
  { id: 'cancel', label: 'Cancel', role: 'cancel', enabled: true },
  { id: 'add', label: 'Add', role: 'default', enabled: true },
  { id: 'help', label: 'Help', role: 'other', enabled: true },
]

const of = (id: string) => layoutOf(classicThemeOf(id))

describe('orderDialogButtons', () => {
  it('puts the default button first on Windows and Cancel last', () => {
    const ids = orderDialogButtons(buttons, of('windows-95')).map(button => button.id)

    expect(ids).toEqual(['add', 'help', 'cancel'])
  })

  it('puts the default button last on a Mac, right after Cancel', () => {
    for (const id of ['mac-os-7', 'aqua']) {
      expect(orderDialogButtons(buttons, of(id)).map(button => button.id)).toEqual(['help', 'cancel', 'add'])
    }
  })
})

describe('layoutOf', () => {
  it('puts the window buttons where each system did', () => {
    expect(of('windows-95')).toMatchObject({ controlsLeft: [], controlsRight: ['minimise', 'maximise', 'close'] })
    expect(of('windows-3-1')).toMatchObject({ titleAlign: 'center', controlsLeft: ['close'], controlsRight: ['minimise', 'maximise'] })
    expect(of('mac-os-7')).toMatchObject({ controlsLeft: ['close'], controlsRight: ['maximise'] })
    expect(of('mac-os-9')).toMatchObject({ controlsLeft: ['close'], controlsRight: ['maximise', 'minimise'] })
    expect(of('aqua')).toMatchObject({ controlsLeft: ['close', 'minimise', 'maximise'], controlsRight: [] })
  })

  it('opens menus only while held on Mac OS 7, and collapses on a double click only from Mac OS 8', () => {
    expect(of('mac-os-7')).toMatchObject({ menuStyle: 'hold', titleDoubleClick: null })
    expect(of('mac-os-9')).toMatchObject({ menuStyle: 'sticky', titleDoubleClick: 'minimise' })
  })

  it('draws a classic Mac dialog as a plain box, Mac OS X as a sheet and Windows as a window', () => {
    expect(of('mac-os-9').dialogStyle).toBe('plain')
    expect(of('aqua').dialogStyle).toBe('sheet')
    expect(of('windows-xp').dialogStyle).toBe('window')
  })

  it('underlines keyboard letters always on early Windows, after Alt later, never on a Mac, and keeps bitmap themes sharp', () => {
    expect(of('windows-3-1')).toMatchObject({ mnemonics: 'always', pixelSnap: true })
    expect(of('windows-95')).toMatchObject({ mnemonics: 'always', pixelSnap: true })
    expect(of('windows-xp')).toMatchObject({ mnemonics: 'on-alt', pixelSnap: false })
    expect(of('mac-os-7')).toMatchObject({ mnemonics: 'none', pixelSnap: true })
    expect(of('aqua')).toMatchObject({ mnemonics: 'none', pixelSnap: false })
  })
})
