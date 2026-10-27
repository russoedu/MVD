/** Wails' window runtime, which only the program's window has. */
interface WindowRuntime {
  WindowGetSize?: () => Promise<{ w: number, h: number }>
  WindowSetSize?: (width: number, height: number) => void
}

const runtime = (): WindowRuntime | undefined => Reflect.get(globalThis, 'runtime') as WindowRuntime | undefined

/** The size of MVD's window in screen points, or null outside the program. */
export async function windowSize (): Promise<{ width: number, height: number } | null> {
  const size = await runtime()?.WindowGetSize?.()

  return size ? { width: size.w, height: size.h } : null
}

/** Gives MVD's window a size, in screen points; outside the program it does nothing. */
export function resizeWindow (width: number, height: number): void {
  runtime()?.WindowSetSize?.(Math.round(width), Math.round(height))
}
