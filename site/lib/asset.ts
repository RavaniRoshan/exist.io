// Base path the site is served under (GitHub Pages project site).
// Local assets imported with absolute "/…" URLs must be prefixed with it.
export const basePath = process.env.NEXT_PUBLIC_BASE_PATH ?? '';

export function asset(path: `/${string}`): string {
  return `${basePath}${path}`;
}
