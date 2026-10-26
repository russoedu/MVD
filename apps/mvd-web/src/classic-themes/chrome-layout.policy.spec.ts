import type { ButtonSnapshot } from '../workspace-snapshot'
import { layoutOf, orderDialogButtons } from './chrome-layout.policy'

const buttons: ButtonSnapshot[] = [
  { id: 'cancel', label: 'Cancel', role: 'cancel', enabled: true },
  { id: 'add', label: 'Add', role: 'default', enabled: true },
  { id: 'help', label: 'Help', role: 'other', enabled: true },
]

describe('orderDialogButtons', () => {
  it('puts the default button first on Windows and Cancel last', () => {
    const ids = orderDialogButtons(buttons, layoutOf('windows-classic')).map(button => button.id)

    expect(ids).toEqual(['add', 'help', 'cancel'])
  })

  it('puts the default button last on a Mac, right after Cancel', () => {
    const ids = orderDialogButtons(buttons, layoutOf('aqua')).map(button => button.id)

    expect(ids).toEqual(['help', 'cancel', 'add'])
  })
})

describe('layoutOf', () => {
  it('puts the Mac window buttons on the left, close first, and the Windows ones on the right, close last', () => {
    expect(layoutOf('aqua')).toMatchObject({ controlsSide: 'left', controls: ['close', 'minimise', 'maximise'] })
    expect(layoutOf('windows-themed')).toMatchObject({ controlsSide: 'right', controls: ['minimise', 'maximise', 'close'] })
  })

  it('underlines keyboard letters always on Windows 95, after Alt on Vista, never on a Mac, and keeps only bitmap themes sharp', () => {
    expect(layoutOf('windows-classic')).toMatchObject({ mnemonics: 'always', pixelSnap: true })
    expect(layoutOf('windows-themed')).toMatchObject({ mnemonics: 'on-alt', pixelSnap: false })
    expect(layoutOf('aqua')).toMatchObject({ mnemonics: 'none', pixelSnap: false })
  })
})
