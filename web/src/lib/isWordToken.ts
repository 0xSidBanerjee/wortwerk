/** True when the token surface is tappable (not punctuation-only). */
export function isWordToken(text: string): boolean {
  if (text === '') return false
  for (const char of text) {
    if (/[\p{L}\p{N}]/u.test(char)) return true
  }
  return false
}
