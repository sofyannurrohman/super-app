export interface UOMGroup {
  name: string
  units: string[]
}

export const COMMON_UOM_GROUPS: UOMGroup[] = [
  {
    name: 'Berat / Massa',
    units: ['KG', 'GRAM', 'TON', 'SAK', 'KARUNG', 'KWINTAL'],
  },
  {
    name: 'Volume / Cairan',
    units: ['LITER', 'ML', 'JERIGEN', 'DRUM', 'GALON'],
  },
  {
    name: 'Kemasan / Wadah / Hitungan',
    units: ['PCS', 'BOTOL', 'LEMBAR', 'KARTON', 'DUS', 'ROLL', 'PACK', 'BOX', 'TUTUP', 'BALL'],
  },
  {
    name: 'Grosir & Hadiah Promosi',
    units: ['LUSIN', 'SET', 'KODI', 'GROSS'],
  },
]

export const ALL_COMMON_UNITS = [
  'KG',
  'GRAM',
  'TON',
  'SAK',
  'KARUNG',
  'LITER',
  'ML',
  'JERIGEN',
  'DRUM',
  'PCS',
  'BOTOL',
  'LEMBAR',
  'KARTON',
  'DUS',
  'ROLL',
  'PACK',
  'BOX',
  'TUTUP',
  'LUSIN',
  'SET',
]

export interface GroupedProducts<T = any> {
  rawMaterials: T[]
  packagingMaterials: T[]
  tradingProducts: T[]
  giftProducts: T[]
  finishedGoods: T[]
  productionUsableProducts: T[]
}

/**
 * Checks if a product is an explicit Finished Good (Barang Jadi) based on category or name prefix.
 */
export function isFinishedGood<T extends { name?: string; category?: string }>(p: T): boolean {
  if (!p) return false
  const cat = (p.category || '').toUpperCase().trim()
  const name = (p.name || '').toUpperCase().trim()

  // 1. Explicit finished goods categories
  if (
    cat.includes('KECAP') ||
    cat.includes('SAUS') ||
    cat.includes('SAOS') ||
    cat.includes('SAMBAL') ||
    cat.includes('CUKA') ||
    cat.includes('MINYAK') ||
    cat.includes('KERUPUK') ||
    cat.includes('MIE') ||
    cat.includes('PETIS') ||
    cat.includes('TERASI') ||
    cat.includes('BARANG JADI') ||
    cat.includes('BARANG_JADI') ||
    cat.includes('FINISHED') ||
    cat.includes('FINISHED_GOODS') ||
    cat.includes('PRODUK JADI') ||
    cat.includes('MAKANAN') ||
    cat.includes('MINUMAN')
  ) {
    return true
  }

  // 2. Name prefix indicating finished goods (KM = Kecap Manis, KA = Kecap Asin, etc.)
  if (
    name.startsWith('KM ') ||
    name.startsWith('KM-') ||
    name.startsWith('KA ') ||
    name.startsWith('KA-') ||
    name.startsWith('KECAP ') ||
    name.startsWith('SAOS ') ||
    name.startsWith('SAUS ') ||
    name.startsWith('SAMBAL ') ||
    name.startsWith('CUKA ') ||
    name.startsWith('MINYAK ')
  ) {
    return true
  }

  return false
}

/**
 * Checks if a product is an Ancillary Good (Barang Sampingan) or Gift/Merchandise (Hadiah & Promosi).
 */
export function isAncillaryOrGiftProduct<T extends { name?: string; category?: string }>(p: T): boolean {
  if (!p) return false
  if (isFinishedGood(p)) return false

  const cat = (p.category || '').toUpperCase().trim()
  const name = (p.name || '').toUpperCase().trim()

  // Category match - Highest priority for ancillary/returnable/gift goods in products table
  if (
    cat.includes('SAMPINGAN') ||
    cat.includes('HADIAH') ||
    cat.includes('MERCHANDISE') ||
    cat.includes('PROMOSI') ||
    cat.includes('RETUR')
  ) {
    return true
  }

  // If explicitly categorized as raw material or packaging, it is not a gift item
  if (
    cat.includes('BAHAN') ||
    cat.includes('RAW') ||
    cat.includes('KEMASAN') ||
    cat.includes('PACKAGING') ||
    cat.includes('WADAH')
  ) {
    return false
  }

  // Exact / standalone merchandise & ancillary return item names
  const standaloneGiftNames = [
    'MANGKOK',
    'PIRING',
    'GELAS',
    'SENDOK',
    'GARPU',
    'MANGKOK AYAM',
    'GELAS KERAMIK',
    'PIRING KERAMIK',
    'KAOS PROMOSI',
    'PAYUNG PROMOSI',
    'TOPI PROMOSI',
    'TAS SPUNBOND',
    'APRON',
    'KRAT KAYU ISI 24',
    'KRAT PLASTIK',
    'KRAT',
    'GALON',
  ]

  return (
    standaloneGiftNames.includes(name) ||
    name.startsWith('HADIAH ') ||
    name.startsWith('MERCHANDISE ') ||
    name.startsWith('PROMO ') ||
    name.startsWith('KRAT ')
  )
}

/**
 * Checks if a product is a Raw Material (Bahan Baku Utama).
 */
export function isRawMaterial<T extends { name?: string; category?: string }>(p: T): boolean {
  if (!p) return false
  if (isFinishedGood(p)) return false

  const cat = (p.category || '').toUpperCase().trim()
  const name = (p.name || '').toUpperCase().trim()

  // Explicit raw material category - Highest priority
  if (
    cat.includes('BAHAN') ||
    cat.includes('RAW') ||
    cat === 'BUMBU' ||
    cat === 'KEDELAI' ||
    cat === 'GULA' ||
    cat === 'GARAM'
  ) {
    return true
  }

  if (isAncillaryOrGiftProduct(p)) return false

  // Exact / standalone raw ingredients
  return (
    name.startsWith('BAHAN ') ||
    name.startsWith('RAW ') ||
    name === 'KEDELAI' ||
    name === 'GULA' ||
    name === 'GARAM' ||
    name.startsWith('KEDELAI ') ||
    name.startsWith('GULA ') ||
    name.startsWith('GARAM ') ||
    name.startsWith('TEPUNG ') ||
    name.startsWith('ASAM SITRAT') ||
    name.startsWith('NATRIUM BENZOAT') ||
    name.startsWith('CARAMEL')
  )
}

/**
 * Checks if a product is a Packaging Material (Kemasan & Wadah).
 */
export function isPackagingMaterial<T extends { name?: string; category?: string }>(p: T): boolean {
  if (!p) return false
  if (isFinishedGood(p)) return false
  if (isAncillaryOrGiftProduct(p)) return false

  const cat = (p.category || '').toUpperCase().trim()
  const name = (p.name || '').toUpperCase().trim()

  // Explicit packaging category - Highest priority
  if (
    cat.includes('KEMASAN') ||
    cat.includes('PACKAGING') ||
    cat.includes('WADAH') ||
    cat === 'BOTOL' ||
    cat === 'KARTON' ||
    cat === 'DUS' ||
    cat === 'LABEL' ||
    cat === 'TUTUP' ||
    cat === 'JERIGEN'
  ) {
    return true
  }

  // Explicit empty packaging / supply materials
  return (
    name.includes('BOTOL KOSONG') ||
    name.includes('JERIGEN KOSONG') ||
    name.includes('TUTUP BOTOL') ||
    name.includes('TUTUP JERIGEN') ||
    name.includes('KARTON BOX') ||
    name.includes('PLASTIK ROLL') ||
    name.includes('POUCH KOSONG') ||
    name.includes('LABEL STIKER') ||
    name.startsWith('KARTON ') ||
    name.startsWith('DUS ') ||
    name.startsWith('KEMASAN ') ||
    name.startsWith('LABEL ')
  )
}

/**
 * Checks if a product is a Trading Product / Produk Pelengkap.
 */
export function isTradingProduct<T extends { name?: string; category?: string }>(p: T): boolean {
  if (!p) return false
  if (isFinishedGood(p)) return false

  const cat = (p.category || '').toUpperCase().trim()
  return (
    cat.includes('TRADING') ||
    cat.includes('TAMBAHAN') ||
    cat.includes('PRODUK_PELENGKAP') ||
    cat.includes('PELENGKAP')
  )
}

export function groupProductsByCategory<T extends { name?: string; category?: string; sku?: string; unit?: string }>(
  products: T[]
): GroupedProducts<T> {
  const rawMaterials: T[] = []
  const packagingMaterials: T[] = []
  const tradingProducts: T[] = []
  const giftProducts: T[] = []
  const finishedGoods: T[] = []

  for (const p of products) {
    // 1. Finished Goods (Barang Jadi) - High priority check
    if (isFinishedGood(p)) {
      finishedGoods.push(p)
      continue
    }

    // 2. Raw Materials (Bahan Baku Utama)
    if (isRawMaterial(p)) {
      rawMaterials.push(p)
      continue
    }

    // 3. Packaging / Kemasan / Wadah
    if (isPackagingMaterial(p)) {
      packagingMaterials.push(p)
      continue
    }

    // 4. Trading / Produk Pelengkap
    if (isTradingProduct(p)) {
      tradingProducts.push(p)
      continue
    }

    // 5. Hadiah / Merchandise Promosi / Barang Sampingan
    if (isAncillaryOrGiftProduct(p)) {
      giftProducts.push(p)
      continue
    }

    // 6. Default to Finished Goods (Barang Jadi)
    finishedGoods.push(p)
  }

  return {
    rawMaterials,
    packagingMaterials,
    tradingProducts,
    giftProducts,
    finishedGoods,
    // For logging production, all non-finished goods (raw, packaging, trading, and gift/sampingan items)
    productionUsableProducts: [...rawMaterials, ...packagingMaterials, ...tradingProducts, ...giftProducts],
  }
}

/**
 * Checks if a product is either a raw material or packaging material.
 */
export function isRawOrPackagingMaterial<T extends { name?: string; category?: string }>(p: T): boolean {
  return isRawMaterial(p) || isPackagingMaterial(p)
}

/**
 * Filters out raw materials and packaging materials, leaving only sellable/finished products,
 * trading items, and gifts/merchandise for sales & retail dropdowns.
 */
export function filterSellableProducts<T extends { name?: string; category?: string }>(products: T[]): T[] {
  if (!Array.isArray(products)) return []
  return products.filter((p) => !isRawOrPackagingMaterial(p))
}

/**
 * Filters products from the products table suitable for invoice returns (Retur Barang Sampingan / Produk).
 * Excludes only raw ingredients (Kedelai, Gula, Garam, etc.) and prioritizes Barang Sampingan / Wadah Retur at the top.
 */
export function filterReturnableProducts<T extends { name?: string; category?: string }>(products: T[]): T[] {
  if (!Array.isArray(products)) return []
  return products
    .filter((p) => !isRawMaterial(p))
    .slice()
    .sort((a, b) => {
      const aIsSampingan = isAncillaryOrGiftProduct(a) ? 0 : 1
      const bIsSampingan = isAncillaryOrGiftProduct(b) ? 0 : 1
      if (aIsSampingan !== bIsSampingan) return aIsSampingan - bIsSampingan
      return (a.name || '').localeCompare(b.name || '')
    })
}

/**
 * Returns only raw materials and packaging materials (useful for factory/recipe views).
 */
export function filterRawMaterials<T extends { name?: string; category?: string }>(products: T[]): T[] {
  if (!Array.isArray(products)) return []
  return products.filter((p) => isRawOrPackagingMaterial(p))
}

// ==========================================
// MULTI-UNIT CONVERSION & FORMATTING UTILITIES
// ==========================================

export type UnitCategory = 'weight' | 'volume' | 'count' | 'other'

/** Standard conversion factors relative to the base unit in its category */
const WEIGHT_TO_GRAM: Record<string, number> = {
  MG: 0.001,
  G: 1,
  GR: 1,
  GRAM: 1,
  ONS: 100,
  HG: 100,
  KG: 1000,
  KILOGRAM: 1000,
  KWINTAL: 100000,
  KUINTAL: 100000,
  Q: 100000,
  TON: 1000000,
}

const VOLUME_TO_ML: Record<string, number> = {
  ML: 1,
  MILLILITER: 1,
  CL: 10,
  DL: 100,
  L: 1000,
  LTR: 1000,
  LITER: 1000,
  GALON: 19000,
  JERIGEN: 20000,
  DRUM: 200000,
}

const COUNT_TO_PCS: Record<string, number> = {
  PCS: 1,
  PIECE: 1,
  PIECES: 1,
  BUAH: 1,
  BIJI: 1,
  BOTOL: 1,
  LEMBAR: 1,
  TUTUP: 1,
  ROLL: 1,
  PACK: 1,
  BOX: 1,
  SET: 1,
  LUSIN: 12,
  DOZEN: 12,
  KODI: 20,
  GROSS: 144,
}

/**
 * Normalizes unit string to uppercase trimmed format.
 */
export function normalizeUnit(unit?: string): string {
  return (unit || '').toUpperCase().trim()
}

/**
 * Determines the measurement category of a given unit.
 */
export function getUnitCategory(unit?: string): UnitCategory {
  const u = normalizeUnit(unit)
  if (WEIGHT_TO_GRAM[u] !== undefined) return 'weight'
  if (VOLUME_TO_ML[u] !== undefined) return 'volume'
  if (COUNT_TO_PCS[u] !== undefined || u === 'DUS' || u === 'KARTON' || u === 'BALL' || u === 'SAK') return 'count'
  return 'other'
}

/**
 * Gets a list of compatible units within the same category.
 */
export function getCompatibleUnits(unit?: string): string[] {
  const cat = getUnitCategory(unit)
  if (cat === 'weight') {
    return ['KG', 'GRAM', 'TON', 'KWINTAL']
  }
  if (cat === 'volume') {
    return ['LITER', 'ML', 'JERIGEN', 'DRUM']
  }
  if (cat === 'count') {
    return ['PCS', 'DUS', 'KARTON', 'PACK', 'LUSIN', 'BOX']
  }
  return [normalizeUnit(unit) || 'PCS']
}

/**
 * Returns the standard highest unit for a given unit.
 * E.g., for GRAM/GR/KG -> 'KG', for ML/LITER -> 'LITER', for PCS -> 'DUS' if pcs_per_unit > 1.
 */
export function getHighestUnit(unit?: string, pcsPerUnit: number = 1): string {
  const cat = getUnitCategory(unit)
  if (cat === 'weight') return 'KG'
  if (cat === 'volume') return 'LITER'
  if (cat === 'count' && pcsPerUnit > 1) return 'DUS'
  return normalizeUnit(unit) || 'PCS'
}

/**
 * Converts a quantity from one unit to another within the same category.
 * If units are incompatible or unknown, returns the original quantity.
 */
export function convertUnitQuantity(
  qty: number,
  fromUnit: string,
  toUnit: string,
  pcsPerUnit: number = 1
): number {
  if (!qty || isNaN(qty)) return 0
  const from = normalizeUnit(fromUnit)
  const to = normalizeUnit(toUnit)
  if (from === to) return qty

  const fromCat = getUnitCategory(from)
  const toCat = getUnitCategory(to)

  // Weight conversion
  if (fromCat === 'weight' && toCat === 'weight') {
    const fromFactor = WEIGHT_TO_GRAM[from] || 1000
    const toFactor = WEIGHT_TO_GRAM[to] || 1000
    const inGram = qty * fromFactor
    return inGram / toFactor
  }

  // Volume conversion
  if (fromCat === 'volume' && toCat === 'volume') {
    const fromFactor = VOLUME_TO_ML[from] || 1000
    const toFactor = VOLUME_TO_ML[to] || 1000
    const inMl = qty * fromFactor
    return inMl / toFactor
  }

  // Count / Packaging conversion
  if (fromCat === 'count' && toCat === 'count') {
    const getPcsFactor = (u: string) => {
      if (u === 'DUS' || u === 'KARTON' || u === 'BALL' || u === 'SAK') {
        return pcsPerUnit > 0 ? pcsPerUnit : 24
      }
      return COUNT_TO_PCS[u] || 1
    }
    const fromFactor = getPcsFactor(from)
    const toFactor = getPcsFactor(to)
    const inPcs = qty * fromFactor
    return inPcs / toFactor
  }

  return qty
}

export interface MultiUnitFormat {
  highestValue: number
  highestUnit: string
  formattedHighest: string
  breakdown: string
  baseFormatted: string
  summary: string
  category: UnitCategory
}

/**
 * Formats a stock quantity into comprehensive multi-unit representations:
 * - Highest Unit (e.g. "12,5 KG")
 * - Breakdown format (e.g. "12 KG 500 GR")
 * - Base Unit format (e.g. "12.500 GRAM")
 * - Summary (e.g. "12,5 KG (12.500 GRAM)")
 */
export function formatMultiUnit(
  qty: number = 0,
  unit?: string,
  pcsPerUnit: number = 1
): MultiUnitFormat {
  const normUnit = normalizeUnit(unit) || 'PCS'
  const cat = getUnitCategory(normUnit)
  const numQty = typeof qty === 'number' ? qty : Number(qty) || 0

  if (cat === 'weight') {
    // Determine base in Gram vs KG
    const factorToGram = WEIGHT_TO_GRAM[normUnit] || 1000
    const totalGrams = numQty * factorToGram
    const totalKg = totalGrams / 1000

    const formattedHighest = `${totalKg.toLocaleString('id-ID', { maximumFractionDigits: 3 })} KG`
    const baseFormatted = `${totalGrams.toLocaleString('id-ID', { maximumFractionDigits: 2 })} GRAM`

    // Detailed breakdown (e.g. 12 KG 500 GR)
    const wholeKg = Math.floor(Math.abs(totalKg)) * (totalKg < 0 ? -1 : 1)
    const remGrams = Math.round(Math.abs(totalGrams) % 1000)
    let breakdown = `${wholeKg.toLocaleString('id-ID')} KG`
    if (remGrams > 0) {
      breakdown += ` ${remGrams.toLocaleString('id-ID')} GR`
    }

    let summary = formattedHighest
    if (normUnit === 'GRAM' || normUnit === 'GR') {
      summary = `${formattedHighest} (${baseFormatted})`
    } else if (totalKg !== totalGrams) {
      summary = `${formattedHighest} (${baseFormatted})`
    }

    return {
      highestValue: totalKg,
      highestUnit: 'KG',
      formattedHighest,
      breakdown,
      baseFormatted,
      summary,
      category: 'weight',
    }
  }

  if (cat === 'volume') {
    const factorToMl = VOLUME_TO_ML[normUnit] || 1000
    const totalMl = numQty * factorToMl
    const totalLiter = totalMl / 1000

    const formattedHighest = `${totalLiter.toLocaleString('id-ID', { maximumFractionDigits: 3 })} LITER`
    const baseFormatted = `${totalMl.toLocaleString('id-ID', { maximumFractionDigits: 2 })} ML`

    const wholeL = Math.floor(Math.abs(totalLiter)) * (totalLiter < 0 ? -1 : 1)
    const remMl = Math.round(Math.abs(totalMl) % 1000)
    let breakdown = `${wholeL.toLocaleString('id-ID')} L`
    if (remMl > 0) {
      breakdown += ` ${remMl.toLocaleString('id-ID')} ML`
    }

    const summary = `${formattedHighest} (${baseFormatted})`

    return {
      highestValue: totalLiter,
      highestUnit: 'LITER',
      formattedHighest,
      breakdown,
      baseFormatted,
      summary,
      category: 'volume',
    }
  }

  if (cat === 'count') {
    const ratio = pcsPerUnit > 1 ? pcsPerUnit : 1
    let totalPcs = numQty
    if (normUnit === 'DUS' || normUnit === 'KARTON' || normUnit === 'BALL' || normUnit === 'SAK') {
      totalPcs = numQty * ratio
    } else if (COUNT_TO_PCS[normUnit]) {
      totalPcs = numQty * COUNT_TO_PCS[normUnit]
    }

    if (ratio > 1) {
      const totalDus = totalPcs / ratio
      const wholeDus = Math.floor(Math.abs(totalPcs) / ratio) * (totalPcs < 0 ? -1 : 1)
      const remPcs = Math.round(Math.abs(totalPcs) % ratio)

      let breakdown = `${wholeDus.toLocaleString('id-ID')} DUS`
      if (remPcs > 0) {
        breakdown += ` ${remPcs} PCS`
      }
      const formattedHighest = `${totalDus.toLocaleString('id-ID', { maximumFractionDigits: 2 })} DUS`
      const baseFormatted = `${totalPcs.toLocaleString('id-ID')} PCS`
      const summary = `${breakdown} (${baseFormatted})`

      return {
        highestValue: totalDus,
        highestUnit: 'DUS',
        formattedHighest: breakdown,
        breakdown,
        baseFormatted,
        summary,
        category: 'count',
      }
    }

    const baseFormatted = `${totalPcs.toLocaleString('id-ID')} ${normUnit}`
    return {
      highestValue: totalPcs,
      highestUnit: normUnit,
      formattedHighest: baseFormatted,
      breakdown: baseFormatted,
      baseFormatted,
      summary: baseFormatted,
      category: 'count',
    }
  }

  const baseFormatted = `${numQty.toLocaleString('id-ID')} ${normUnit}`
  return {
    highestValue: numQty,
    highestUnit: normUnit,
    formattedHighest: baseFormatted,
    breakdown: baseFormatted,
    baseFormatted,
    summary: baseFormatted,
    category: 'other',
  }
}

