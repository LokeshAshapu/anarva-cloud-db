// Web Crypto API SHA-256 Digest Hashing Helper for Zero-Trust Password Encryption

export async function hashPassword(plainText: string): Promise<string> {
  if (!plainText) return ''
  const encoder = new TextEncoder()
  const data = encoder.encode(plainText)
  const hashBuffer = await crypto.subtle.digest('SHA-256', data)
  const hashArray = Array.from(new Uint8Array(hashBuffer))
  return hashArray.map((b) => b.toString(16).padStart(2, '0')).join('')
}
