/**
 * Utility for client-side image compression
 */
export async function compressImageFile(
  file: File,
  maxWidth = 1600,
  maxHeight = 1600,
  quality = 0.8
): Promise<{ compressedFile: File; originalSize: number; compressedSize: number }> {
  const originalSize = file.size

  // If not an image, return original file
  if (!file.type.startsWith('image/')) {
    return { compressedFile: file, originalSize, compressedSize: originalSize }
  }

  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.readAsDataURL(file)
    reader.onload = (event) => {
      const img = new Image()
      img.src = event.target?.result as string
      img.onload = () => {
        let width = img.width
        let height = img.height

        // Calculate aspect ratio downscaling
        if (width > height) {
          if (width > maxWidth) {
            height = Math.round((height * maxWidth) / width)
            width = maxWidth
          }
        } else {
          if (height > maxHeight) {
            width = Math.round((width * maxHeight) / height)
            height = maxHeight
          }
        }

        const canvas = document.createElement('canvas')
        canvas.width = width
        canvas.height = height

        const ctx = canvas.getContext('2d')
        if (!ctx) {
          resolve({ compressedFile: file, originalSize, compressedSize: originalSize })
          return
        }

        // Fill white background for transparent PNG converted to JPEG
        ctx.fillStyle = '#FFFFFF'
        ctx.fillRect(0, 0, width, height)
        ctx.drawImage(img, 0, 0, width, height)

        // Convert to compressed jpeg blob
        canvas.toBlob(
          (blob) => {
            if (!blob) {
              resolve({ compressedFile: file, originalSize, compressedSize: originalSize })
              return
            }

            // Only use compressed if it is actually smaller than original
            if (blob.size < originalSize) {
              const newFileName = file.name.replace(/\.[^/.]+$/, '') + '.jpg'
              const compressedFile = new File([blob], newFileName, {
                type: 'image/jpeg',
                lastModified: Date.now(),
              })
              resolve({
                compressedFile,
                originalSize,
                compressedSize: blob.size,
              })
            } else {
              resolve({
                compressedFile: file,
                originalSize,
                compressedSize: originalSize,
              })
            }
          },
          'image/jpeg',
          quality
        )
      }
      img.onerror = (err) => reject(err)
    }
    reader.onerror = (err) => reject(err)
  })
}

export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}
