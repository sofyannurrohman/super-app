import { unpack_cascade, run_cascade, cluster_detections } from './pico'

let faceClassifyRegion: any = null
let isModelLoading = false

async function loadFaceModel(): Promise<void> {
  if (faceClassifyRegion) return
  if (isModelLoading) {
    // Wait a brief moment for the model to finish loading if already initiated
    return new Promise((resolve) => {
      const checkInterval = setInterval(() => {
        if (faceClassifyRegion) {
          clearInterval(checkInterval)
          resolve()
        }
      }, 50)
    })
  }

  isModelLoading = true
  try {
    const res = await fetch('/models/facefinder')
    if (!res.ok) {
      throw new Error(`Failed to fetch cascade file: ${res.statusText}`)
    }
    const buffer = await res.arrayBuffer()
    const bytes = new Uint8Array(buffer)
    faceClassifyRegion = unpack_cascade(bytes)
  } catch (error) {
    console.error('Error loading face detection model:', error)
  } finally {
    isModelLoading = false
  }
}

function rgbaToGrayscale(rgba: Uint8ClampedArray, nrows: number, ncols: number): Uint8Array {
  const gray = new Uint8Array(nrows * ncols)
  for (let r = 0; r < nrows; ++r) {
    for (let c = 0; c < ncols; ++c) {
      const idx = r * ncols + c
      // standard grayscale conversion formula: 0.299*R + 0.587*G + 0.114*B
      gray[idx] = Math.round(
        (299 * rgba[4 * idx + 0]! + 587 * rgba[4 * idx + 1]! + 114 * rgba[4 * idx + 2]!) / 1000
      )
    }
  }
  return gray
}

/**
 * Detects face in a base64 or source image, crops it with padding, resizes to 300x300, and returns compressed base64 JPEG.
 * Falls back to compressing the full resized image if no face is detected.
 */
export async function detectAndCropFace(imageSrc: string): Promise<string> {
  await loadFaceModel()

  return new Promise((resolve) => {
    const img = new Image()
    img.crossOrigin = 'anonymous'
    img.src = imageSrc
    
    img.onload = () => {
      try {
        // If the model failed to load, go straight to fallback
        if (!faceClassifyRegion) {
          console.warn('Face detection model not available, using fallback compression')
          resolve(getFallbackCompressedImage(img))
          return
        }

        // Scale down large images for faster processing in browser
        const maxDim = 600
        let w = img.width
        let h = img.height
        if (w > maxDim || h > maxDim) {
          if (w > h) {
            h = Math.round((h * maxDim) / w)
            w = maxDim
          } else {
            w = Math.round((w * maxDim) / h)
            h = maxDim
          }
        }

        const canvas = document.createElement('canvas')
        canvas.width = w
        canvas.height = h
        const ctx = canvas.getContext('2d')
        if (!ctx) {
          resolve(getFallbackCompressedImage(img))
          return
        }
        ctx.drawImage(img, 0, 0, w, h)

        const imgData = ctx.getImageData(0, 0, w, h)
        const gray = rgbaToGrayscale(imgData.data, h, w)

        const picoImage = {
          pixels: gray,
          nrows: h,
          ncols: w,
          ldim: w,
        }

        const params = {
          shiftfactor: 0.1, // Move detection window by 10%
          minsize: 50,      // Minimum face size
          maxsize: 1000,    // Maximum face size
          scalefactor: 1.1, // Scale window size by 10% per step
        }

        // Run face detection cascade
        let dets = run_cascade(picoImage, faceClassifyRegion, params)
        // Group multiple overlapping detection boxes
        dets = cluster_detections(dets, 0.2) // IoU overlap threshold

        // Find the detection box with the highest score
        let bestDet = null
        for (const det of dets) {
          if (det[3] > 8.0) { // Detection confidence threshold (8.0-10.0 works well)
            if (!bestDet || det[3] > bestDet[3]) {
              bestDet = det
            }
          }
        }

        if (bestDet) {
          const [r, c, s] = bestDet // r=row center (Y), c=col center (X), s=size

          // Crop with padding around the face (45% padding)
          const padding = s * 0.45
          const rawX = c - s / 2 - padding
          const rawY = r - s / 2 - padding
          const rawSize = s + padding * 2

          // Bound coordinates to image edges
          const x = Math.max(0, rawX)
          const y = Math.max(0, rawY)
          const size = Math.min(w - x, h - y, rawSize)

          const cropCanvas = document.createElement('canvas')
          cropCanvas.width = 300
          cropCanvas.height = 300
          const cropCtx = cropCanvas.getContext('2d')
          
          if (cropCtx) {
            // Draw cropped part of the original high-res image
            const scaleX = img.width / w
            const scaleY = img.height / h
            cropCtx.drawImage(
              img,
              x * scaleX,
              y * scaleY,
              size * scaleX,
              size * scaleY,
              0,
              0,
              300,
              300
            )

            // Compress to JPEG with 70% quality (ideal for storage vs visibility)
            const croppedBase64 = cropCanvas.toDataURL('image/jpeg', 0.70)
            resolve(croppedBase64)
            return
          }
        }
      } catch (err) {
        console.error('Error during face detection processing:', err)
      }

      // Fallback if no face was found or an error occurred
      resolve(getFallbackCompressedImage(img))
    }

    img.onerror = () => {
      console.error('Failed to load image for face detection')
      resolve(imageSrc) // Return original if load failed
    }
  })
}

function getFallbackCompressedImage(img: HTMLImageElement): string {
  const fallbackCanvas = document.createElement('canvas')
  const fallbackSize = 400
  let fw = img.width
  let fh = img.height
  
  if (fw > fallbackSize || fh > fallbackSize) {
    if (fw > fh) {
      fh = Math.round((fh * fallbackSize) / fw)
      fw = fallbackSize
    } else {
      fw = Math.round((fw * fallbackSize) / fh)
      fh = fallbackSize
    }
  }

  fallbackCanvas.width = fw
  fallbackCanvas.height = fh
  const fallbackCtx = fallbackCanvas.getContext('2d')
  if (fallbackCtx) {
    fallbackCtx.drawImage(img, 0, 0, fw, fh)
    return fallbackCanvas.toDataURL('image/jpeg', 0.70)
  }
  
  return img.src
}
