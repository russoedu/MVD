import { isMnemonic, pressedLetter, splitMnemonic } from './mnemonic.mapper'

describe('splitMnemonic', () => {
  it('cuts the label around the first place the letter appears, in either case', () => {
    expect(splitMnemonic('Accept all songs', 'c')).toEqual({ before: 'A', letter: 'c', after: 'cept all songs' })
    expect(splitMnemonic('Platform URL', 'u')).toEqual({ before: 'Platform ', letter: 'U', after: 'RL' })
  })

  it('gives nothing when there is no letter, or the label does not have it', () => {
    expect(splitMnemonic('Cancel')).toBeNull()
    expect(splitMnemonic('Cancel', 'x')).toBeNull()
  })
})

describe('pressedLetter', () => {
  it('reads the letter the key writes, lower case', () => {
    expect(pressedLetter({ key: 'F', code: 'KeyF' })).toBe('f')
    expect(pressedLetter({ key: 'é', code: 'Digit2' })).toBe('é')
  })

  it("falls back to the key's place when Alt turns it into a symbol or a dead key", () => {
    expect(pressedLetter({ key: '©', code: 'KeyG' })).toBe('g')
    expect(pressedLetter({ key: 'Dead', code: 'KeyE' })).toBe('e')
  })

  it('gives nothing for keys that are not letters or digits', () => {
    expect(pressedLetter({ key: 'Alt', code: 'AltLeft' })).toBeNull()
    expect(pressedLetter({ key: 'ArrowDown', code: 'ArrowDown' })).toBeNull()
  })
})

describe('isMnemonic', () => {
  it('matches the letter in either case', () => {
    expect(isMnemonic('f', 'F')).toBe(true)
    expect(isMnemonic('f', 'e')).toBe(false)
    expect(isMnemonic(null, 'f')).toBe(false)
    expect(isMnemonic('f')).toBe(false)
  })
})
