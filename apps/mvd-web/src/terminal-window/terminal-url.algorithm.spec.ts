import { terminalUrl } from './terminal-url.algorithm'

describe('terminalUrl', () => {
  it('uses ws on a page served over http', () => {
    expect(terminalUrl({ protocol: 'http:', host: '127.0.0.1:8421' })).toBe('ws://127.0.0.1:8421/term')
  })

  it('uses wss on a page served over https', () => {
    expect(terminalUrl({ protocol: 'https:', host: 'mvd.example' })).toBe('wss://mvd.example/term')
  })
})
