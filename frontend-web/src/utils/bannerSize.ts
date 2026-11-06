export interface BannerSizeSuggestion {
  reqLength: number
  reqWidth: number
  designLength: number
  designWidth: number
  reduction: number
  ruleText: string
  formattedReqDual?: string
  formattedDesignDual?: string
}

export interface BannerDimensionDetails {
  lengthCm: number
  widthCm: number
  lengthM: number
  widthM: number
  areaM2: number
  formattedCm: string
  formattedM: string
  formattedDual: string
  formattedShort: string
}

/**
 * Normalisasi pasangan ukuran banner (panjang & lebar) ke satuan Centimeter (cm).
 * Membedakan input satuan Meter vs Centimeter secara cerdas:
 * - Jika salah satu dimensi >= 20, atau keduanya >= 10, dianggap sudah dalam Centimeter (cm).
 * - Jika kedua dimensi bernilai kecil (misal 3 x 1, 1.47 x 0.2, 0.8 x 0.4), otomatis dikonversi dari Meter ke Centimeter (x 100).
 */
export function normalizeDimensionPair(
  rawLength: number | string | null | undefined,
  rawWidth: number | string | null | undefined
): { lengthCm: number; widthCm: number } {
  const l = Number(rawLength) || 0
  const w = Number(rawWidth) || 0
  if (l <= 0 && w <= 0) return { lengthCm: 0, widthCm: 0 }

  // Jika salah satu dimensi >= 20, atau keduanya >= 10, kemungkinan besar sudah dalam cm
  const isAlreadyCm = l >= 20 || w >= 20 || (l >= 10 && w >= 10)

  const lengthCm = isAlreadyCm ? Math.round(l) : Math.round(l * 100)
  const widthCm = isAlreadyCm ? Math.round(w) : Math.round(w * 100)

  return { lengthCm, widthCm }
}

/**
 * Normalisasi ukuran tunggal input/database ke satuan Centimeter (cm).
 * Jika nilai <= 10 (misal input meter seperti 3, 1.5, 0.8), dikonversi ke cm (x 100).
 * Untuk dimensi berpasangan (panjang & lebar), disarankan menggunakan normalizeDimensionPair.
 */
export function normalizeDimensionToCm(val: number | string | null | undefined): number {
  const num = Number(val) || 0
  if (num <= 0) return 0
  return num <= 10 ? Math.round(num * 100) : Math.round(num)
}

/**
 * Format ukuran dimensi banner:
 * - 'dual' (default): '300 × 100 cm (3 × 1 m)'
 * - 'cm': '300 × 100 cm'
 * - 'm': '3 × 1 m'
 * - 'short': '300×100 cm (3×1 m)'
 */
export function formatBannerDimension(
  rawLength: number | string | null | undefined,
  rawWidth: number | string | null | undefined,
  format: 'dual' | 'cm' | 'm' | 'short' = 'dual'
): string {
  const { lengthCm: lCm, widthCm: wCm } = normalizeDimensionPair(rawLength, rawWidth)
  if (lCm <= 0 || wCm <= 0) return '-'

  const lM = +(lCm / 100).toFixed(2)
  const wM = +(wCm / 100).toFixed(2)

  if (format === 'cm') return `${lCm} × ${wCm} cm`
  if (format === 'm') return `${lM} × ${wM} m`
  if (format === 'short') return `${lCm}×${wCm} cm (${lM}×${wM} m)`

  return `${lCm} × ${wCm} cm (${lM} × ${wM} m)`
}

/**
 * Mendapatkan objek detail dimensi spanduk lengkap (cm, meter, luas m²).
 */
export function getBannerDimensionDetails(
  rawLength: number | string | null | undefined,
  rawWidth: number | string | null | undefined
): BannerDimensionDetails | null {
  const { lengthCm: lCm, widthCm: wCm } = normalizeDimensionPair(rawLength, rawWidth)
  if (lCm <= 0 || wCm <= 0) return null

  const lM = +(lCm / 100).toFixed(2)
  const wM = +(wCm / 100).toFixed(2)
  const areaM2 = +((lCm / 100) * (wCm / 100)).toFixed(2)

  return {
    lengthCm: lCm,
    widthCm: wCm,
    lengthM: lM,
    widthM: wM,
    areaM2,
    formattedCm: `${lCm} × ${wCm} cm`,
    formattedM: `${lM} × ${wM} m`,
    formattedDual: `${lCm} × ${wCm} cm (${lM} × ${wM} m)`,
    formattedShort: `${lCm}×${wCm} cm (${lM}×${wM} m)`,
  }
}

/**
 * Menghitung sugesti ukuran banner untuk tim desain:
 * - Jika panjang >= 100 cm (>= 1m): panjang dikurangi 30 cm, lebar mengikuti aspek rasio proporsional.
 * - Jika panjang < 100 cm (< 1m): panjang dikurangi 10 cm, lebar mengikuti aspek rasio proporsional.
 */
export function calculateBannerDesignSize(
  rawLength: number | string | undefined | null,
  rawWidth: number | string | undefined | null,
): BannerSizeSuggestion | null {
  const l = Number(rawLength) || 0
  const w = Number(rawWidth) || 0

  if (l <= 0 || w <= 0) return null

  const { lengthCm: normL, widthCm: normW } = normalizeDimensionPair(l, w)
  if (normL <= 0 || normW <= 0) return null

  const reduction = normL >= 100 ? 30 : 10
  const designLength = Math.max(normL - reduction, 5)
  const designWidth = Math.max(Math.round((normW * designLength) / normL), 1)

  return {
    reqLength: normL,
    reqWidth: normW,
    designLength,
    designWidth,
    reduction,
    ruleText: normL >= 100 ? '-30cm (Proporsional)' : '-10cm (<1m Proporsional)',
    formattedReqDual: formatBannerDimension(normL, normW, 'dual'),
    formattedDesignDual: formatBannerDimension(designLength, designWidth, 'dual'),
  }
}
