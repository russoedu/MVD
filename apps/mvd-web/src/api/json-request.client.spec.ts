import { ApiError, requestJson } from './index'

const fetchMock = jest.fn()

beforeEach(() => {
  fetchMock.mockReset()
  Object.defineProperty(globalThis, 'fetch', { value: fetchMock, configurable: true, writable: true })
})

function respond (status: number, body: unknown) {
  fetchMock.mockResolvedValue({ ok: status < 400, status, json: () => (body === undefined ? Promise.reject(new Error('no body')) : Promise.resolve(body)) })
}

async function failure (call: Promise<unknown>): Promise<unknown> {
  try {
    await call
  } catch (error) {
    return error
  }

  return undefined
}

describe('requestJson', () => {
  it('sends the method and a JSON body, and returns the JSON answer', async () => {
    respond(200, { ok: true })

    const result = await requestJson<{ ok: boolean }>('PUT', '/api/x', { a: 1 })

    expect(result).toEqual({ ok: true })
    const [path, init] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect(path).toBe('/api/x')
    expect(init.method).toBe('PUT')
    expect(new Headers(init.headers).get('Content-Type')).toBe('application/json')
    expect(init.body).toBe('{"a":1}')
  })

  it('sends no body when there is none', async () => {
    respond(200, {})
    await requestJson('GET', '/api/x')
    expect((fetchMock.mock.calls[0] as [string, RequestInit])[1].body).toBeUndefined()
  })

  it('throws the server\'s own message and the field errors', async () => {
    respond(400, { error: 'the settings are not valid', fields: { outputDir: 'choose a folder' } })

    const error = await failure(requestJson('PUT', '/api/x', {}))

    expect(error).toBeInstanceOf(ApiError)
    expect((error as ApiError).message).toBe('the settings are not valid')
    expect((error as ApiError).status).toBe(400)
    expect((error as ApiError).fields).toEqual({ outputDir: 'choose a folder' })
  })

  it('falls back to the status when the error has no JSON body', async () => {
    respond(502, undefined)

    const error = await failure(requestJson('GET', '/api/x'))

    expect((error as ApiError).message).toBe('The server answered 502')
    expect((error as ApiError).fields).toEqual({})
  })
})
