import { describeResult } from './describe-add-result.algorithm'

describe('describeResult', () => {
  it('names each outcome that happened', () => {
    expect(describeResult({ added: ['a', 'b'], duplicates: ['c'], rejected: [] })).toBe('2 added, 1 already queued')
    expect(describeResult({ added: [], duplicates: [], rejected: ['x'] })).toBe('1 not a web link')
    expect(describeResult({ added: [], duplicates: [], rejected: [] })).toBe('')
  })
})
