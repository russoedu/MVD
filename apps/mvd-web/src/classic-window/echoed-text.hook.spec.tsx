import { act, renderHook } from '@testing-library/react'
import { useEchoedText } from './echoed-text.hook'

describe('useEchoedText', () => {
  beforeEach(() => {
    Reflect.set(globalThis, 'IS_REACT_ACT_ENVIRONMENT', true)
  })

  it('keeps every key typed before the program has answered the first ones', () => {
    const send = jest.fn()
    const { result, rerender } = renderHook(({ text }) => useEchoedText(text, send), { initialProps: { text: '' } })

    act(() => { for (const typed of ['T', 'Th', 'Thr']) result.current[1](typed) })
    expect(result.current[0]).toBe('Thr')

    // The answers arrive late, oldest first: they must not put the text back.
    for (const answer of ['T', 'Th', 'Thr']) {
      rerender({ text: answer })
      expect(result.current[0]).toBe('Thr')
    }
    expect(send.mock.calls.slice(0, 3)).toEqual([['T'], ['Th'], ['Thr']])
  })

  it('takes what the program changes by itself, such as an emptied field', () => {
    const { result, rerender } = renderHook(({ text }) => useEchoedText(text, jest.fn()), { initialProps: { text: '' } })
    act(() => { result.current[1]('abc') })
    rerender({ text: 'abc' })
    rerender({ text: '' })

    expect(result.current[0]).toBe('')
  })
})
