/**
 * Normalizes photo and image URLs across the client installation.
 * Handles:
 * Handles Google Drive links, local uploads, and relative image paths.
 */
export const resolvePhotoUrl = (path: string | null | undefined): string => {
  if (!path || !path.trim()) return ''

  let url = path.trim()

  // FileReader previews are already complete browser URLs. Rewriting them as
  // legacy Branding storage paths makes a freshly selected image look broken.
  if (url.startsWith('data:image/') || url.startsWith('blob:')) {
    return url
  }

  // 2. Google Drive preview link converter
  if (url.includes('drive.google.com/file/d/')) {
    const fileIdMatch = url.match(/\/d\/([a-zA-Z0-9_-]+)/)
    if (fileIdMatch && fileIdMatch[1]) {
      return `https://drive.google.com/file/d/${fileIdMatch[1]}/preview`
    }
  }

  // 3. Absolute URLs (http/https) that are valid and not misdirected
  if (url.startsWith('http://') || url.startsWith('https://')) {
    return url
  }

  // 4. Local Super App Uploads (/uploads/... or uploads/...)
  const cleanLocalPath = url.startsWith('/') ? url : `/${url}`
  if (cleanLocalPath.startsWith('/uploads/') || cleanLocalPath.startsWith('/api/v1/uploads/')) {
    return cleanLocalPath
  }

  // 5. Fallback for legacy remote Branding storage paths (e.g. "shops/xxx.jpg", "banners/yyy.jpg")
  const stripped = url.replace(/^\/+/, '')
  return `/uploads/${stripped}`
}
