interface ErrorBody {
  error?:  string
  fields?: Record<string, string>
}

/** A failed API call. `fields` names what is wrong with each input, by its JSON key. */
export class ApiError extends Error {
  constructor (
    message: string,
    readonly status: number,
    readonly fields: Record<string, string> = {},
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

/** Calls the MVD API with a JSON body and returns the JSON answer. */
export async function requestJson<T> (method: string, path: string, body?: unknown): Promise<T> {
  const response = await fetch(path, {
    method,
    headers: { 'Content-Type': 'application/json' },
    body:    body === undefined ? undefined : JSON.stringify(body),
  })
  let payload = {} as T & ErrorBody
  try {
    payload = await response.json() as T & ErrorBody
  } catch {
    // An error from a proxy or a crash has no JSON body; the status says enough.
  }
  if (!response.ok) {
    throw new ApiError(payload.error ?? `The server answered ${response.status}`, response.status, payload.fields)
  }

  return payload
}
