import { resolveAssetUrl } from '@/api/serverConfig'

/**
 * Returns the best available avatar URL for a user object.
 * Priority: explicit avatar_url → Gravatar URL supplied by backend → null
 * Resolves server-relative /uploads/... paths to absolute URLs in desktop mode.
 */
export function avatarUrl(user) {
  if (!user) return null
  const url = user.avatar_url || user.gravatar_url || null
  return resolveAssetUrl(url)
}
