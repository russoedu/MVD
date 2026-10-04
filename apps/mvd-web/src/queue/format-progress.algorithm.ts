const UNITS = ['B', 'KB', 'MB', 'GB', 'TB']

/** 1536 -> "1.5 KB". Empty when the size is unknown. */
export function formatBytes (bytes: number): string {
  if (!Number.isFinite(bytes) || bytes <= 0) return ''
  let value = bytes
  let unit = 0
  while (value >= 1024 && unit < UNITS.length - 1) {
    value /= 1024
    unit += 1
  }

  return `${unit === 0 || value >= 10 ? Math.round(value) : value.toFixed(1)} ${UNITS[unit]}`
}

/** Bytes per second -> "2.4 MB/s". Empty when unknown. */
export function formatSpeed (bytesPerSecond: number): string {
  const size = formatBytes(bytesPerSecond)

  return size === '' ? '' : `${size}/s`
}

/** Seconds -> "1:05" or "1:02:03". Empty when unknown (the server sends -1). */
export function formatEta (seconds: number): string {
  if (!Number.isFinite(seconds) || seconds < 0) return ''
  const total = Math.round(seconds)
  const hours = Math.floor(total / 3600)
  const minutes = Math.floor((total % 3600) / 60)
  const rest = String(total % 60).padStart(2, '0')

  return hours > 0
    ? `${hours}:${String(minutes).padStart(2, '0')}:${rest}`
    : `${minutes}:${rest}`
}
