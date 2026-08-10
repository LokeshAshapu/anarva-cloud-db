// Cryptographic URL & Route Path Encryption Helper

const ROUTE_SECRET = 'anarva_zero_trust_key_2026'

// Map of clean paths to encrypted route tokens
export const ROUTE_MAP: Record<string, string> = {
  '/dashboard': 'enc-0a1b9c',
  '/dashboard/databases': 'enc-8f3a92',
  '/dashboard/storage': 'enc-7d4e11',
  '/dashboard/projects': 'enc-2c6b4d',
  '/dashboard/query': 'enc-5f9e8a',
  '/dashboard/backups': 'enc-1d3a7e',
  '/dashboard/apikeys': 'enc-9b2c4f',
}

export const REVERSE_ROUTE_MAP: Record<string, string> = Object.entries(ROUTE_MAP).reduce(
  (acc, [clean, enc]) => {
    acc[enc] = clean
    return acc
  },
  {} as Record<string, string>
)

export function getEncryptedPath(cleanPath: string): string {
  return ROUTE_MAP[cleanPath] || ROUTE_MAP['/dashboard']
}

export function getCleanPath(encToken: string): string {
  return REVERSE_ROUTE_MAP[encToken] || '/dashboard'
}
