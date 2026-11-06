import { ref, onUnmounted } from 'vue'
import { FilesetResolver, FaceLandmarker } from '@mediapipe/tasks-vision'

export type FaceStatusCode = 'LOADING' | 'NO_FACE' | 'MULTIPLE_FACES' | 'MASK_DETECTED' | 'VALID' | 'ERROR'

export function useFaceDetector() {
  const isModelLoading = ref(true)
  const isModelReady = ref(false)
  const faceCount = ref(0)
  const isMaskDetected = ref(false)
  const isFaceValid = ref(false)
  const statusCode = ref<FaceStatusCode>('LOADING')
  const statusMessage = ref('Memuat sistem deteksi wajah...')
  
  let landmarker: FaceLandmarker | null = null
  let animationFrameId: number | null = null
  let lastProcessTime = 0
  let isProcessing = false
  
  // Offscreen canvas for texture and mask analysis
  const sampleCanvas = document.createElement('canvas')
  const sampleCtx = sampleCanvas.getContext('2d', { willReadFrequently: true })

  // Initialize MediaPipe Face Landmarker
  const initDetector = async () => {
    try {
      isModelLoading.value = true
      statusMessage.value = 'Menyiapkan modul AI deteksi wajah...'
      
      const vision = await FilesetResolver.forVisionTasks(
        'https://cdn.jsdelivr.net/npm/@mediapipe/tasks-vision@latest/wasm'
      )
      
      landmarker = await FaceLandmarker.createFromOptions(vision, {
        baseOptions: {
          modelAssetPath: 'https://storage.googleapis.com/mediapipe-models/face_landmarker/face_landmarker/float16/1/face_landmarker.task',
          delegate: 'GPU'
        },
        runningMode: 'VIDEO',
        numFaces: 2,
        minFaceDetectionConfidence: 0.5,
        minFacePresenceConfidence: 0.5,
        minTrackingConfidence: 0.5
      })

      isModelReady.value = true
      isModelLoading.value = false
      statusCode.value = 'NO_FACE'
      statusMessage.value = 'Arahkan wajah ke kamera'
    } catch (err) {
      console.warn('Gagal memuat FaceLandmarker dengan GPU, mencoba CPU fallback...', err)
      try {
        const vision = await FilesetResolver.forVisionTasks(
          'https://cdn.jsdelivr.net/npm/@mediapipe/tasks-vision@latest/wasm'
        )
        landmarker = await FaceLandmarker.createFromOptions(vision, {
          baseOptions: {
            modelAssetPath: 'https://storage.googleapis.com/mediapipe-models/face_landmarker/face_landmarker/float16/1/face_landmarker.task',
            delegate: 'CPU'
          },
          runningMode: 'VIDEO',
          numFaces: 2,
          minFaceDetectionConfidence: 0.5,
          minFacePresenceConfidence: 0.5,
          minTrackingConfidence: 0.5
        })
        isModelReady.value = true
        isModelLoading.value = false
        statusCode.value = 'NO_FACE'
        statusMessage.value = 'Arahkan wajah ke kamera'
      } catch (fallbackErr) {
        console.error('Face detector initialization failed:', fallbackErr)
        isModelLoading.value = false
        isModelReady.value = false
        statusCode.value = 'ERROR'
        statusMessage.value = 'Gagal memuat sistem deteksi wajah'
      }
    }
  }

  /**
   * Analyze lower face (nose tip to chin, mouth region) to check for masks or face coverings.
   */
  const checkMaskOcclusion = (video: HTMLVideoElement, landmarks: any[]): boolean => {
    if (!sampleCtx || landmarks.length < 468) return false

    const vWidth = video.videoWidth || 640
    const vHeight = video.videoHeight || 480

    // Key landmarks
    const noseTip = landmarks[1]       // Nose tip
    const upperLip = landmarks[0] || landmarks[13]     // Upper lip top / middle
    const lowerLip = landmarks[17] || landmarks[14]    // Lower lip bottom / middle
    const mouthLeft = landmarks[61]    // Left mouth corner
    const mouthRight = landmarks[291]  // Right mouth corner
    const chin = landmarks[152]        // Chin bottom
    const leftCheek = landmarks[234]   // Left cheek
    const rightCheek = landmarks[454]  // Right cheek
    const forehead = landmarks[10]     // Forehead
    const leftEye = landmarks[33]      // Left eye
    const rightEye = landmarks[263]    // Right eye

    if (!noseTip || !upperLip || !lowerLip || !chin || !leftCheek || !rightCheek || !forehead) {
      return true // Landmarks missing = likely covered
    }

    // 1. Geometric proportions check
    const faceHeight = Math.abs(chin.y - forehead.y)
    const faceWidth = Math.abs(rightCheek.x - leftCheek.x)
    const mouthWidth = Math.abs(mouthRight.x - mouthLeft.x)
    const noseToChin = Math.abs(chin.y - noseTip.y)

    // Mask often compresses or severely distorts mouth width or distance relative to face
    if (faceWidth > 0 && mouthWidth / faceWidth < 0.12) {
      return true
    }
    if (faceHeight > 0 && noseToChin / faceHeight < 0.15) {
      return true
    }

    // 2. Forehead skin baseline sampling (control area not covered by mask)
    const foreheadX = Math.floor(Math.max(0, Math.min(vWidth - 10, forehead.x * vWidth - 5)))
    const foreheadY = Math.floor(Math.max(0, Math.min(vHeight - 10, forehead.y * vHeight)))
    sampleCanvas.width = 10
    sampleCanvas.height = 10
    sampleCtx.drawImage(video, foreheadX, foreheadY, 10, 10, 0, 0, 10, 10)
    const foreheadData = sampleCtx.getImageData(0, 0, 10, 10).data
    let fR = 0, fG = 0, fB = 0
    for (let i = 0; i < foreheadData.length; i += 4) {
      fR += foreheadData[i] ?? 0
      fG += foreheadData[i + 1] ?? 0
      fB += foreheadData[i + 2] ?? 0
    }
    const baselineSkinR = fR / 25
    const baselineSkinG = fG / 25
    const baselineSkinB = fB / 25
    const skinBrightness = (baselineSkinR + baselineSkinG + baselineSkinB) / 3

    // 3. Lower Face ROI (Mouth and chin area)
    const minX = Math.max(0, Math.min(mouthLeft.x, leftCheek.x) * vWidth)
    const maxX = Math.min(vWidth, Math.max(mouthRight.x, rightCheek.x) * vWidth)
    const minY = Math.max(0, (noseTip.y + 0.01) * vHeight)
    const maxY = Math.min(vHeight, chin.y * vHeight)

    const roiWidth = Math.max(10, Math.floor(maxX - minX))
    const roiHeight = Math.max(10, Math.floor(maxY - minY))

    sampleCanvas.width = 30
    sampleCanvas.height = 30
    sampleCtx.drawImage(video, minX, minY, roiWidth, roiHeight, 0, 0, 30, 30)
    const imgData = sampleCtx.getImageData(0, 0, 30, 30)
    const data = imgData.data

    let blueMaskPixels = 0
    let blackMaskPixels = 0
    let starkWhiteMaskPixels = 0
    const totalPixels = 30 * 30

    for (let i = 0; i < data.length; i += 4) {
      const r = data[i] ?? 0
      const g = data[i + 1] ?? 0
      const b = data[i + 2] ?? 0

      // a. Surgical Blue mask (distinct blue/cyan tint compared to red channel)
      if (b > r + 20 && g > r + 10 && b > 75) {
        blueMaskPixels++
      }

      // b. Black/Dark mask (very dark cloth when skin tone is not pitch black)
      if (r < 40 && g < 40 && b < 40 && skinBrightness > 55) {
        blackMaskPixels++
      }

      // c. Pure white surgical/KF94 mask (high brightness and much brighter than forehead skin)
      if (r > 205 && g > 205 && b > 205 && (r - baselineSkinR > 35 || skinBrightness < 160)) {
        starkWhiteMaskPixels++
      }
    }

    const blueRatio = blueMaskPixels / totalPixels
    const blackRatio = blackMaskPixels / totalPixels
    const whiteRatio = starkWhiteMaskPixels / totalPixels

    // If mask-specific colors cover significant portion of lower face
    if (blueRatio > 0.25) return true
    if (blackRatio > 0.40) return true
    if (whiteRatio > 0.45) return true

    return false
  }

  // Process a single video frame
  const processFrame = (video: HTMLVideoElement) => {
    if (!landmarker || !video || video.readyState < 2 || isProcessing) return

    const now = performance.now()
    // Throttle to ~10-15 FPS (every 70-100ms) to conserve CPU/GPU
    if (now - lastProcessTime < 75) return

    lastProcessTime = now
    isProcessing = true

    try {
      const results = landmarker.detectForVideo(video, now)
      
      if (!results || !results.faceLandmarks || results.faceLandmarks.length === 0) {
        faceCount.value = 0
        isMaskDetected.value = false
        isFaceValid.value = false
        statusCode.value = 'NO_FACE'
        statusMessage.value = 'Arahkan wajah ke kamera'
      } else if (results.faceLandmarks.length > 1) {
        faceCount.value = results.faceLandmarks.length
        isMaskDetected.value = false
        isFaceValid.value = false
        statusCode.value = 'MULTIPLE_FACES'
        statusMessage.value = 'Hanya 1 orang di depan kamera'
      } else {
        faceCount.value = 1
        const primaryLandmarks = results.faceLandmarks[0]
        if (!primaryLandmarks) {
          isFaceValid.value = false
          return
        }
        const hasMask = checkMaskOcclusion(video, primaryLandmarks)

        isMaskDetected.value = hasMask

        if (hasMask) {
          isFaceValid.value = false
          statusCode.value = 'MASK_DETECTED'
          statusMessage.value = 'Masker terdeteksi! Harap buka masker'
        } else {
          isFaceValid.value = true
          statusCode.value = 'VALID'
          statusMessage.value = 'Wajah valid & siap absen'
        }
      }
    } catch (err) {
      console.debug('Error detecting frame:', err)
    } finally {
      isProcessing = false
    }
  }

  // Start continuous detection loop
  const startDetection = (video: HTMLVideoElement) => {
    const loop = () => {
      processFrame(video)
      animationFrameId = requestAnimationFrame(loop)
    }
    loop()
  }

  // Stop detection loop
  const stopDetection = () => {
    if (animationFrameId !== null) {
      cancelAnimationFrame(animationFrameId)
      animationFrameId = null
    }
  }

  onUnmounted(() => {
    stopDetection()
    if (landmarker) {
      try {
        landmarker.close()
      } catch (e) {
        // ignore
      }
      landmarker = null
    }
  })

  return {
    isModelLoading,
    isModelReady,
    faceCount,
    isMaskDetected,
    isFaceValid,
    statusCode,
    statusMessage,
    initDetector,
    startDetection,
    stopDetection
  }
}
