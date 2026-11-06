/**
 * Utility to group individual product packaging variants into high-level Product Families (Induk Produk).
 * Example:
 * - "Kecap Manis" includes variants like KM Botol 600ml, KM Pouch 500ml, KM Jerigen 5L, etc.
 * - "Kecap Asin" includes KA Botol 600ml, KA Pouch 500ml, etc.
 * - "Saus Sambal", "Saus Tomat", "Cuka Karsa HR", etc.
 */

export interface ProductVariantItem {
  id: string
  name: string
  sku?: string
  unit?: string
  imageUrl?: string
  price?: number
  category?: string
  raw?: any
}

export interface ProductFamilyGroup {
  id: string
  name: string
  shortName: string
  category: string
  imageUrl: string
  productIds: string[]
  variants: ProductVariantItem[]
}

export function extractFamilyName(productName: string): { familyName: string; shortName: string; category: string } {
  const upper = productName.toUpperCase().trim()

  // 1. Kecap Manis Family
  if (
    upper.includes('KECAP MANIS') ||
    upper.startsWith('KM ') ||
    upper.startsWith('KM-') ||
    upper.startsWith('KECAP MANIS ')
  ) {
    return {
      familyName: 'Kecap Manis',
      shortName: 'Kecap Manis',
      category: 'Kecap',
    }
  }

  // 2. Kecap Asin Family
  if (
    upper.includes('KECAP ASIN') ||
    upper.startsWith('KA ') ||
    upper.startsWith('KA-') ||
    upper.startsWith('KECAP ASIN ')
  ) {
    return {
      familyName: 'Kecap Asin',
      shortName: 'Kecap Asin',
      category: 'Kecap',
    }
  }

  // 3. Saus Sambal / Saos Cabe Family
  if (
    upper.includes('SAUS SAMBAL') ||
    upper.includes('SAOS SAMBAL') ||
    upper.includes('SAUS CABE') ||
    upper.includes('SAOS CABE') ||
    upper.startsWith('SAMBAL ')
  ) {
    return {
      familyName: 'Saus Sambal',
      shortName: 'Saus Sambal',
      category: 'Saus & Sambal',
    }
  }

  // 4. Saus Tomat Family
  if (upper.includes('SAUS TOMAT') || upper.includes('SAOS TOMAT')) {
    return {
      familyName: 'Saus Tomat',
      shortName: 'Saus Tomat',
      category: 'Saus & Sambal',
    }
  }

  // 5. Cuka Family
  if (upper.includes('CUKA')) {
    return {
      familyName: 'Cuka Karsa HR',
      shortName: 'Cuka',
      category: 'Bumbu Masak',
    }
  }

  // 6. Minyak Goreng Family
  if (upper.includes('MINYAK GORENG') || upper.startsWith('MINYAK ')) {
    return {
      familyName: 'Minyak Goreng',
      shortName: 'Minyak Goreng',
      category: 'Minyak',
    }
  }

  // 7. Kerupuk / Camilan
  if (upper.includes('KERUPUK') || upper.includes('KROPOK')) {
    return {
      familyName: 'Kerupuk Karsa HR',
      shortName: 'Kerupuk',
      category: 'Makanan Ringan',
    }
  }

  // 8. Default fallback: Clean packaging size/measurement suffixes
  // Removes (600ml, 500g, 5L, Botol, Pouch, Jerigen, dsb)
  let clean = productName
    .replace(/\b\d+(\.\d+)?\s*(ML|LITER|L|KG|GRAM|GR|SAK|DUS|PCS|BOX|PACK|BTL|BOTOL|POUCH|JRG|JERIGEN)\b/gi, '')
    .replace(/[(),]/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()

  if (!clean || clean.length < 3) {
    clean = productName
  }

  return {
    familyName: clean,
    shortName: clean,
    category: 'Produk Lainnya',
  }
}

/**
 * Builds array of Product Families from a list of products.
 */
export function buildProductFamilies(products: any[]): ProductFamilyGroup[] {
  if (!Array.isArray(products)) return []

  const familyMap = new Map<string, ProductFamilyGroup>()

  products.forEach((p) => {
    if (!p || !p.id) return
    const name = p.name || 'Produk'
    const { familyName, shortName, category } = extractFamilyName(name)
    const familyKey = familyName.toLowerCase().replace(/[^a-z0-9]/g, '_')

    if (!familyMap.has(familyKey)) {
      familyMap.set(familyKey, {
        id: familyKey,
        name: familyName,
        shortName: shortName,
        category: p.category || category,
        imageUrl: p.image_url || '',
        productIds: [],
        variants: [],
      })
    }

    const group = familyMap.get(familyKey)!
    group.productIds.push(p.id)
    if (!group.imageUrl && p.image_url) {
      group.imageUrl = p.image_url
    }
    group.variants.push({
      id: p.id,
      name: p.name,
      sku: p.sku,
      unit: p.unit,
      imageUrl: p.image_url,
      price: p.selling_price || p.price,
      category: p.category,
      raw: p,
    })
  })


  return Array.from(familyMap.values()).sort((a, b) => a.name.localeCompare(b.name))
}
