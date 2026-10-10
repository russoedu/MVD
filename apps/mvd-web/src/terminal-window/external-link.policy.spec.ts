import { isLinkQuestion, isWebAddress } from './external-link.policy'

describe('isWebAddress', () => {
  it('accepts http and https addresses', () => {
    expect(isWebAddress('https://www.youtube.com/watch?v=8DNQRtmIMxk')).toBe(true)
    // eslint-disable-next-line unicorn/prefer-https -- plain http is accepted on purpose
    expect(isWebAddress('http://example.com')).toBe(true)
  })

  it('refuses anything that could start another program', () => {
    expect(isWebAddress('file:///C:/Windows/System32/calc.exe')).toBe(false)
    expect(isWebAddress('javascript:alert(1)')).toBe(false)
    expect(isWebAddress('calc.exe')).toBe(false)
    expect(isWebAddress('')).toBe(false)
  })
})

describe('isLinkQuestion', () => {
  it('recognises the question xterm asks before following a link', () => {
    expect(isLinkQuestion('Do you want to navigate to https://a.b?\n\nWARNING: This link could potentially be dangerous')).toBe(true)
  })

  it('leaves other questions alone', () => {
    expect(isLinkQuestion('Delete everything?')).toBe(false)
    expect(isLinkQuestion(undefined)).toBe(false)
  })
})
