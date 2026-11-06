/**
 * Watermark Generator for Attendance Photos
 * Matches the visual design and layout of Flutter Mobile App (`face_verification_page.dart`).
 */

export interface AttendanceWatermarkOptions {
  timeText?: string
  branchName?: string
  addressText?: string
  targetWidth?: number
  cropSquare?: boolean
  quality?: number
}

/**
 * Format Date into Indonesian Locale: "EEEE, dd MMMM yyyy HH:mm:ss"
 * Example: "Senin, 28 September 2026 08:30:15"
 */
export function formatAttendanceDateTime(date: Date = new Date()): string {
  const days = ['Minggu', 'Senin', 'Selasa', 'Rabu', 'Kamis', 'Jumat', 'Sabtu']
  const months = [
    'Januari',
    'Februari',
    'Maret',
    'April',
    'Mei',
    'Juni',
    'Juli',
    'Agustus',
    'September',
    'Oktober',
    'November',
    'Desember',
  ]

  const dayName = days[date.getDay()]
  const day = String(date.getDate()).padStart(2, '0')
  const monthName = months[date.getMonth()]
  const year = date.getFullYear()
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')

  return `${dayName}, ${day} ${monthName} ${year} ${hours}:${minutes}:${seconds}`
}

/**
 * Try reverse geocoding via OpenStreetMap Nominatim API (with timeout & fallback)
 */
export async function reverseGeocodeAddress(
  lat: number,
  lon: number,
  timeoutMs: number = 3000,
): Promise<string> {
  const fallbackStr = `Lat: ${lat.toFixed(6)}, Long: ${lon.toFixed(6)}`
  try {
    const controller = new AbortController()
    const timer = setTimeout(() => controller.abort(), timeoutMs)

    const response = await fetch(
      `https://nominatim.openstreetmap.org/reverse?format=json&lat=${lat}&lon=${lon}&zoom=18&addressdetails=1`,
      {
        headers: {
          'Accept-Language': 'id,en',
        },
        signal: controller.signal,
      },
    )
    clearTimeout(timer)

    if (!response.ok) return fallbackStr
    const data = await response.json()
    if (data && data.display_name) {
      return data.display_name
    }
    return fallbackStr
  } catch {
    return fallbackStr
  }
}

/**
 * Helper to split text into lines that fit within max width in canvas context
 */
function wrapText(
  ctx: CanvasRenderingContext2D,
  text: string,
  maxWidth: number,
): string[] {
  const words = text.split(' ')
  const lines: string[] = []
  let currentLine = ''

  for (const word of words) {
    if (!word) continue
    const testLine = currentLine ? `${currentLine} ${word}` : word
    const metrics = ctx.measureText(testLine)
    if (metrics.width > maxWidth && currentLine) {
      lines.push(currentLine)
      currentLine = word
    } else {
      currentLine = testLine
    }
  }
  if (currentLine) {
    lines.push(currentLine)
  }
  return lines.length > 0 ? lines : [text]
}

/**
 * Draws an icon with text on canvas
 */
function drawIcon(
  ctx: CanvasRenderingContext2D,
  type: 'time' | 'branch' | 'location',
  x: number,
  y: number,
  size: number,
  color: string = '#ffffff',
) {
  ctx.save()
  ctx.fillStyle = color
  ctx.strokeStyle = color
  ctx.lineWidth = Math.max(1.5, size * 0.12)
  ctx.lineCap = 'round'
  ctx.lineJoin = 'round'

  const cx = x + size / 2
  const cy = y - size / 2.5

  if (type === 'time') {
    // Clock Icon
    const radius = size * 0.44
    ctx.beginPath()
    ctx.arc(cx, cy, radius, 0, Math.PI * 2)
    ctx.stroke()

    // Hands
    ctx.beginPath()
    ctx.moveTo(cx, cy)
    ctx.lineTo(cx, cy - radius * 0.55)
    ctx.moveTo(cx, cy)
    ctx.lineTo(cx + radius * 0.45, cy)
    ctx.stroke()
  } else if (type === 'branch') {
    // Storefront / Building Icon
    const w = size * 0.85
    const h = size * 0.75
    const startX = cx - w / 2
    const startY = cy - h / 2

    // Roof / Top
    ctx.beginPath()
    ctx.strokeRect(startX, startY, w, h)

    // Door
    const dw = w * 0.35
    const dh = h * 0.5
    ctx.strokeRect(cx - dw / 2, startY + h - dh, dw, dh)
  } else if (type === 'location') {
    // Location Pin Icon
    const r = size * 0.35
    const topY = cy - size * 0.2
    ctx.beginPath()
    ctx.arc(cx, topY, r, Math.PI * 0.8, Math.PI * 2.2, false)
    ctx.lineTo(cx, cy + size * 0.45)
    ctx.closePath()
    ctx.stroke()

    // Inner dot
    ctx.beginPath()
    ctx.arc(cx, topY, r * 0.35, 0, Math.PI * 2)
    ctx.fill()
  }

  ctx.restore()
}

/**
 * Burns attendance watermark onto canvas and returns Base64 JPEG data URL
 */
export async function burnAttendanceWatermark(
  imageSource: string | HTMLImageElement | HTMLVideoElement | File | Blob,
  options: AttendanceWatermarkOptions = {},
): Promise<string> {
  const {
    timeText = formatAttendanceDateTime(new Date()),
    branchName = 'Cabang Kantor Pusat',
    addressText = 'Lokasi Presensi Karyawan',
    targetWidth = 720,
    cropSquare = true,
    quality = 0.8,
  } = options

  return new Promise((resolve, reject) => {
    const processImage = (img: HTMLImageElement | HTMLVideoElement) => {
      try {
        const naturalW =
          img instanceof HTMLImageElement ? img.naturalWidth || img.width : img.videoWidth
        const naturalH =
          img instanceof HTMLImageElement ? img.naturalHeight || img.height : img.videoHeight

        if (!naturalW || !naturalH) {
          throw new Error('Dimensi gambar tidak valid.')
        }

        const canvas = document.createElement('canvas')
        const ctx = canvas.getContext('2d')
        if (!ctx) throw new Error('Canvas 2D context tidak didukung.')

        let srcX = 0
        let srcY = 0
        let srcW = naturalW
        let srcH = naturalH

        if (cropSquare) {
          const cropSize = Math.min(naturalW, naturalH)
          srcX = (naturalW - cropSize) / 2
          srcY = (naturalH - cropSize) / 2
          srcW = cropSize
          srcH = cropSize
        }

        const targetW = targetWidth
        const targetH = cropSquare ? targetWidth : Math.round((naturalH / naturalW) * targetWidth)

        canvas.width = targetW
        canvas.height = targetH

        // 1. Draw cropped image
        ctx.drawImage(img, srcX, srcY, srcW, srcH, 0, 0, targetW, targetH)

        // 2. Setup typography & dimensions
        const padding = Math.round(targetW * 0.035) // e.g. ~25px on 720w
        const fontSize = Math.max(13, Math.round(targetW * 0.026)) // ~18px
        const iconSize = Math.round(fontSize * 1.1)
        const lineHeight = Math.round(fontSize * 1.55)

        ctx.font = `500 ${fontSize}px 'Poppins', 'Inter', -apple-system, BlinkMacSystemFont, sans-serif`

        const maxTextW = targetW - padding * 2 - iconSize - 12
        const wrappedAddressLines = wrapText(ctx, addressText, maxTextW)

        const totalTextLines = 2 + wrappedAddressLines.length
        const boxHeight = totalTextLines * lineHeight + padding * 2
        const boxY = targetH - boxHeight - Math.round(padding * 0.6)

        // 3. Draw Watermark Background Bar (Semi-transparent black)
        ctx.fillStyle = 'rgba(0, 0, 0, 0.55)'
        ctx.fillRect(0, boxY, targetW, boxHeight + Math.round(padding * 0.6))

        // 4. Render Text & Icons
        let currentY = boxY + padding + fontSize

        // Line 1: Date & Time
        drawIcon(ctx, 'time', padding, currentY, iconSize, '#ffffff')
        ctx.fillStyle = '#ffffff'
        ctx.fillText(timeText, padding + iconSize + 10, currentY)
        currentY += lineHeight

        // Line 2: Branch Name
        drawIcon(ctx, 'branch', padding, currentY, iconSize, '#ffffff')
        ctx.fillStyle = '#ffffff'
        ctx.fillText(branchName, padding + iconSize + 10, currentY)
        currentY += lineHeight

        // Line 3+: Location / Address
        drawIcon(ctx, 'location', padding, currentY, iconSize, '#ffffff')
        ctx.fillStyle = '#ffffff'
        const firstLine = wrappedAddressLines[0] || addressText
        ctx.fillText(firstLine, padding + iconSize + 10, currentY)

        for (let i = 1; i < wrappedAddressLines.length; i++) {
          const lineText = wrappedAddressLines[i]
          if (lineText) {
            currentY += lineHeight
            ctx.fillText(lineText, padding + iconSize + 10, currentY)
          }
        }

        // Return Data URL
        const dataUrl = canvas.toDataURL('image/jpeg', quality)
        resolve(dataUrl)
      } catch (error) {
        reject(error)
      }
    }

    if (imageSource instanceof HTMLVideoElement || imageSource instanceof HTMLImageElement) {
      processImage(imageSource)
      return
    }

    const img = new Image()
    img.crossOrigin = 'anonymous'
    img.onload = () => processImage(img)
    img.onerror = (err) => reject(err)

    if (typeof imageSource === 'string') {
      img.src = imageSource
    } else {
      const url = URL.createObjectURL(imageSource)
      img.src = url
    }
  })
}

/**
 * Downloads attendance photo with watermark burned onto it.
 * Falls back to raw photo download if canvas/CORS fails.
 */
export async function downloadWatermarkedAttendancePhoto(
  imageUrl: string,
  options: AttendanceWatermarkOptions,
  filename: string,
): Promise<void> {
  try {
    let sourceToBurn: string | Blob = imageUrl

    // Try fetching as blob to bypass cross-origin canvas taint when possible
    try {
      const resp = await fetch(imageUrl, { mode: 'cors' })
      if (resp.ok) {
        sourceToBurn = await resp.blob()
      }
    } catch {
      // Ignore and use imageUrl directly
    }

    const watermarkedDataUrl = await burnAttendanceWatermark(sourceToBurn, options)
    const link = document.createElement('a')
    link.href = watermarkedDataUrl
    link.download = filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  } catch (error) {
    console.warn('Gagal mem-burn watermark untuk download, fallback ke foto asli:', error)
    const fallbackLink = document.createElement('a')
    fallbackLink.href = imageUrl
    fallbackLink.download = filename
    fallbackLink.target = '_blank'
    document.body.appendChild(fallbackLink)
    fallbackLink.click()
    document.body.removeChild(fallbackLink)
  }
}

