import ExcelJS from 'exceljs'
import { saveAs } from 'file-saver'
import type { RawMaterialPR, RawMaterialPO } from '@/stores/factory'

export interface ExportPurchasingOptions {
  prs: RawMaterialPR[]
  pos: RawMaterialPO[]
  factoryName?: string
  scope?: 'FULL' | 'PR_ONLY' | 'PO_ONLY'
  dateRange?: {
    from?: string
    to?: string
  }
  onProgress?: (percent: number, message: string) => void
}

const formatDate = (dateStr?: string | null): string => {
  if (!dateStr) return '-'
  try {
    const d = new Date(dateStr)
    return d.toLocaleDateString('id-ID', {
      day: '2-digit',
      month: 'short',
      year: 'numeric',
    })
  } catch {
    return dateStr
  }
}

const formatPRStatus = (status?: string): string => {
  switch (status) {
    case 'PENDING':
      return 'Menunggu Tindak Lanjut'
    case 'PROCESSED':
      return 'Sudah Dibuat PO'
    case 'REJECTED':
      return 'Ditolak'
    case 'CANCELLED':
      return 'Dibatalkan'
    default:
      return status || '-'
  }
}

const formatPOStatus = (status?: string): string => {
  switch (status) {
    case 'DRAFT':
      return 'Draft'
    case 'PENDING_SPV':
      return 'Menunggu SPV'
    case 'PENDING_PURCHASING':
      return 'Menunggu Purchasing'
    case 'APPROVED':
      return 'Disetujui'
    case 'PARTIAL_RECEIVED':
      return 'Diterima Sebagian'
    case 'RECEIVED':
      return 'Selesai Diterima'
    case 'REJECTED':
      return 'Ditolak'
    case 'CANCELLED':
      return 'Dibatalkan'
    default:
      return status || '-'
  }
}

const formatProgressStatus = (status?: string): string => {
  switch (status) {
    case 'PENDING_APPROVAL':
      return 'Menunggu Persetujuan'
    case 'ORDERED':
      return 'Dipesan ke Supplier'
    case 'PRODUCTION':
      return 'Proses Produksi Vendor'
    case 'SHIPPING':
      return 'Dalam Pengiriman'
    case 'DELIVERED':
      return 'Tiba di Pabrik'
    case 'COMPLETED':
      return 'Selesai'
    default:
      return status || '-'
  }
}

export async function exportPurchasingReportToExcel(options: ExportPurchasingOptions) {
  const { prs, pos, factoryName = 'Semua Unit Pabrik', scope = 'FULL', dateRange, onProgress } = options

  onProgress?.(10, 'Menyiapkan lembar kerja spreadsheet...')

  const workbook = new ExcelJS.Workbook()
  workbook.creator = 'Karsa HR Management ERP'
  workbook.lastModifiedBy = 'Purchasing Department'
  workbook.created = new Date()
  workbook.modified = new Date()

  const periodText =
    dateRange?.from && dateRange?.to
      ? `${formatDate(dateRange.from)} s/d ${formatDate(dateRange.to)}`
      : 'Semua Periode'

  const totalNilaiPengadaan = pos.reduce((sum, p) => sum + (p.total_estimated_amount || 0), 0)
  const totalPR = prs.length
  const totalPRProcessed = prs.filter((p) => p.status === 'PROCESSED').length
  const totalPRPending = prs.filter((p) => p.status === 'PENDING').length
  const totalPRRejected = prs.filter((p) => p.status === 'REJECTED').length
  const conversionRate = totalPR > 0 ? Math.round((totalPRProcessed / totalPR) * 100) : 0

  const totalPO = pos.length
  const totalPOReceived = pos.filter((p) => p.status === 'RECEIVED').length
  const totalPOInProgress = pos.filter((p) => p.status === 'APPROVED' || p.status === 'PARTIAL_RECEIVED').length

  // ===========================================================================
  // SHEET 0: EXECUTIVE SUMMARY / DASHBOARD RINGKASAN (HANYA JIKA FULL)
  // ===========================================================================
  if (scope === 'FULL') {
    onProgress?.(25, 'Membuat Dashboard Ringkasan Eksekutif...')

    const sheet0 = workbook.addWorksheet('Ringkasan Eksekutif', {
      views: [{ showGridLines: true }],
    })

    // Header Banner
    sheet0.mergeCells('A1:H1')
    const title0 = sheet0.getCell('A1')
    title0.value = 'Perusahaan Demo'
    title0.font = { name: 'Arial', size: 14, bold: true, color: { argb: 'FFFFFFFF' } }
    title0.alignment = { horizontal: 'left', vertical: 'middle', indent: 1 }
    title0.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF0F172A' } } // slate-900
    sheet0.getRow(1).height = 30

    sheet0.mergeCells('A2:H2')
    const subtitle0 = sheet0.getCell('A2')
    subtitle0.value = 'EXECUTIVE SUMMARY & PURCHASING PERFORMANCE REPORT'
    subtitle0.font = { name: 'Arial', size: 11, bold: true, color: { argb: 'FF94A3B8' } }
    subtitle0.alignment = { horizontal: 'left', vertical: 'middle', indent: 1 }
    subtitle0.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF1E293B' } } // slate-800
    sheet0.getRow(2).height = 24

    // Info Metadata Box
    sheet0.mergeCells('A3:H3')
    const meta0 = sheet0.getCell('A3')
    meta0.value = `Unit: ${factoryName}   |   Periode: ${periodText}   |   Waktu Cetak: ${new Date().toLocaleString('id-ID')}`
    meta0.font = { name: 'Arial', size: 9, italic: true, color: { argb: 'FF64748B' } }
    meta0.alignment = { horizontal: 'left', vertical: 'middle', indent: 1 }
    meta0.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFF8FAFC' } }
    sheet0.getRow(3).height = 20

    // Blank row 4
    sheet0.getRow(4).height = 10

    // KPI Section Title
    sheet0.getCell('A5').value = 'INDIKATOR KINERJA UTAMA PENGADAAN (KPI SCORECARDS)'
    sheet0.getCell('A5').font = { name: 'Arial', size: 10, bold: true, color: { argb: 'FF0F172A' } }

    // KPI Card 1: Total PR
    sheet0.mergeCells('A6:B6')
    sheet0.getCell('A6').value = 'TOTAL PERMINTAAN (PR)'
    sheet0.getCell('A6').font = { name: 'Arial', size: 9, bold: true, color: { argb: 'FF475569' } }
    sheet0.getCell('A6').alignment = { horizontal: 'center', vertical: 'middle' }
    sheet0.getCell('A6').fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFF1F5F9' } }

    sheet0.mergeCells('A7:B7')
    sheet0.getCell('A7').value = `${totalPR} Request`
    sheet0.getCell('A7').font = { name: 'Arial', size: 14, bold: true, color: { argb: 'FF0F172A' } }
    sheet0.getCell('A7').alignment = { horizontal: 'center', vertical: 'middle' }
    sheet0.getCell('A7').fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFFFFFFF' } }

    // KPI Card 2: Realisasi PR ke PO (%)
    sheet0.mergeCells('C6:D6')
    sheet0.getCell('C6').value = 'RASIO REALISASI PO'
    sheet0.getCell('C6').font = { name: 'Arial', size: 9, bold: true, color: { argb: 'FF065F46' } }
    sheet0.getCell('C6').alignment = { horizontal: 'center', vertical: 'middle' }
    sheet0.getCell('C6').fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFECFDF5' } }

    sheet0.mergeCells('C7:D7')
    sheet0.getCell('C7').value = `${conversionRate}% Sukses PO`
    sheet0.getCell('C7').font = { name: 'Arial', size: 14, bold: true, color: { argb: 'FF059669' } }
    sheet0.getCell('C7').alignment = { horizontal: 'center', vertical: 'middle' }
    sheet0.getCell('C7').fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFFFFFFF' } }

    // KPI Card 3: Total PO Terbit
    sheet0.mergeCells('E6:F6')
    sheet0.getCell('E6').value = 'TOTAL PURCHASE ORDER'
    sheet0.getCell('E6').font = { name: 'Arial', size: 9, bold: true, color: { argb: 'FF1E40AF' } }
    sheet0.getCell('E6').alignment = { horizontal: 'center', vertical: 'middle' }
    sheet0.getCell('E6').fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFEFF6FF' } }

    sheet0.mergeCells('E7:F7')
    sheet0.getCell('E7').value = `${totalPO} PO Terbit`
    sheet0.getCell('E7').font = { name: 'Arial', size: 14, bold: true, color: { argb: 'FF1D4ED8' } }
    sheet0.getCell('E7').alignment = { horizontal: 'center', vertical: 'middle' }
    sheet0.getCell('E7').fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFFFFFFF' } }

    // KPI Card 4: Total Belanja Pengadaan (Rp)
    sheet0.mergeCells('G6:H6')
    sheet0.getCell('G6').value = 'TOTAL NILAI PENGADAAN'
    sheet0.getCell('G6').font = { name: 'Arial', size: 9, bold: true, color: { argb: 'FF065F46' } }
    sheet0.getCell('G6').alignment = { horizontal: 'center', vertical: 'middle' }
    sheet0.getCell('G6').fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFD1FAE5' } }

    sheet0.mergeCells('G7:H7')
    sheet0.getCell('G7').value = totalNilaiPengadaan
    sheet0.getCell('G7').font = { name: 'Arial', size: 13, bold: true, color: { argb: 'FF047857' } }
    sheet0.getCell('G7').alignment = { horizontal: 'center', vertical: 'middle' }
    sheet0.getCell('G7').fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFFFFFFF' } }
    sheet0.getCell('G7').numFmt = '[$Rp-421] #,##0'

    // Add borders to cards
    const cardRanges = [
      ['A6', 'B6', 'A7', 'B7'],
      ['C6', 'D6', 'C7', 'D7'],
      ['E6', 'F6', 'E7', 'F7'],
      ['G6', 'H6', 'G7', 'H7'],
    ]
    cardRanges.forEach((range) => {
      range.forEach((cellName) => {
        const cell = sheet0.getCell(cellName)
        cell.border = {
          top: { style: 'thin', color: { argb: 'FFCBD5E1' } },
          bottom: { style: 'thin', color: { argb: 'FFCBD5E1' } },
          left: { style: 'thin', color: { argb: 'FFCBD5E1' } },
          right: { style: 'thin', color: { argb: 'FFCBD5E1' } },
        }
      })
    })

    // Breakdown Tables
    sheet0.getCell('A10').value = 'Rincian Status Purchase Request (PR)'
    sheet0.getCell('A10').font = { name: 'Arial', size: 10, bold: true, color: { argb: 'FF1E293B' } }

    const prStatusTable = [
      ['Status Permintaan', 'Jumlah Request', 'Persentase'],
      ['Menunggu Tindak Lanjut (Pending)', totalPRPending, totalPR > 0 ? totalPRPending / totalPR : 0],
      ['Sudah Dibuatkan PO (Processed)', totalPRProcessed, totalPR > 0 ? totalPRProcessed / totalPR : 0],
      ['Ditolak / Batal (Rejected)', totalPRRejected, totalPR > 0 ? totalPRRejected / totalPR : 0],
      ['Total Keseluruhan', totalPR, 1],
    ]

    prStatusTable.forEach((rowVals, idx) => {
      const row = sheet0.getRow(11 + idx)
      row.getCell(1).value = rowVals[0]
      row.getCell(2).value = rowVals[1]
      row.getCell(3).value = rowVals[2]
      row.height = 20

      const isHeader = idx === 0
      const isTotal = idx === prStatusTable.length - 1

      for (let c = 1; c <= 3; c++) {
        const cell = row.getCell(c)
        cell.border = {
          top: { style: 'thin', color: { argb: 'FFE2E8F0' } },
          bottom: { style: isTotal ? 'double' : 'thin', color: { argb: 'FFE2E8F0' } },
          left: { style: 'thin', color: { argb: 'FFE2E8F0' } },
          right: { style: 'thin', color: { argb: 'FFE2E8F0' } },
        }
        if (isHeader) {
          cell.font = { name: 'Arial', size: 9, bold: true, color: { argb: 'FFFFFFFF' } }
          cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF334155' } }
          cell.alignment = { horizontal: c === 1 ? 'left' : 'center', vertical: 'middle' }
        } else {
          cell.font = { name: 'Arial', size: 9, bold: isTotal }
          cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: isTotal ? 'FFF8FAFC' : 'FFFFFFFF' } }
          if (c === 2) {
            cell.alignment = { horizontal: 'right', vertical: 'middle' }
            cell.numFmt = '#,##0'
          } else if (c === 3) {
            cell.alignment = { horizontal: 'right', vertical: 'middle' }
            cell.numFmt = '0.0%'
          } else {
            cell.alignment = { horizontal: 'left', vertical: 'middle' }
          }
        }
      }
    })

    // Breakdown PO Status
    sheet0.getCell('E10').value = 'Rincian Realisasi Purchase Order (PO)'
    sheet0.getCell('E10').font = { name: 'Arial', size: 10, bold: true, color: { argb: 'FF1E293B' } }

    const poStatusTable = [
      ['Status Pengadaan PO', 'Jumlah PO', 'Persentase'],
      ['Selesai Diterima Gudang (Received)', totalPOReceived, totalPO > 0 ? totalPOReceived / totalPO : 0],
      ['Sedang Berjalan / Pengiriman', totalPOInProgress, totalPO > 0 ? totalPOInProgress / totalPO : 0],
      ['Lainnya (Draft/Pending/Ditolak)', totalPO - totalPOReceived - totalPOInProgress, totalPO > 0 ? (totalPO - totalPOReceived - totalPOInProgress) / totalPO : 0],
      ['Total Keseluruhan PO', totalPO, 1],
    ]

    poStatusTable.forEach((rowVals, idx) => {
      const row = sheet0.getRow(11 + idx)
      row.getCell(5).value = rowVals[0]
      row.getCell(6).value = rowVals[1]
      row.getCell(7).value = rowVals[2]

      const isHeader = idx === 0
      const isTotal = idx === poStatusTable.length - 1

      for (let c = 5; c <= 7; c++) {
        const cell = row.getCell(c)
        cell.border = {
          top: { style: 'thin', color: { argb: 'FFE2E8F0' } },
          bottom: { style: isTotal ? 'double' : 'thin', color: { argb: 'FFE2E8F0' } },
          left: { style: 'thin', color: { argb: 'FFE2E8F0' } },
          right: { style: 'thin', color: { argb: 'FFE2E8F0' } },
        }
        if (isHeader) {
          cell.font = { name: 'Arial', size: 9, bold: true, color: { argb: 'FFFFFFFF' } }
          cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF065F46' } }
          cell.alignment = { horizontal: c === 5 ? 'left' : 'center', vertical: 'middle' }
        } else {
          cell.font = { name: 'Arial', size: 9, bold: isTotal }
          cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: isTotal ? 'FFF0FDF4' : 'FFFFFFFF' } }
          if (c === 6) {
            cell.alignment = { horizontal: 'right', vertical: 'middle' }
            cell.numFmt = '#,##0'
          } else if (c === 7) {
            cell.alignment = { horizontal: 'right', vertical: 'middle' }
            cell.numFmt = '0.0%'
          } else {
            cell.alignment = { horizontal: 'left', vertical: 'middle' }
          }
        }
      }
    })

    // Sheet 0 Columns width
    sheet0.getColumn(1).width = 30
    sheet0.getColumn(2).width = 16
    sheet0.getColumn(3).width = 16
    sheet0.getColumn(4).width = 6
    sheet0.getColumn(5).width = 32
    sheet0.getColumn(6).width = 16
    sheet0.getColumn(7).width = 16
    sheet0.getColumn(8).width = 24
  }

  // ===========================================================================
  // SHEET 1: REKAP PURCHASE REQUEST & TINDAK LANJUT
  // ===========================================================================
  if (scope === 'FULL' || scope === 'PR_ONLY') {
    onProgress?.(50, 'Memformat tabel Purchase Request & Tindak Lanjut...')

    const sheet1 = workbook.addWorksheet('Rekap PR & Tindak Lanjut', {
      views: [
        {
          state: 'frozen',
          xSplit: 0,
          ySplit: 6, // Freeze header row
          activeCell: 'A7',
          showGridLines: true,
        },
      ],
    })

    // Title block
    sheet1.mergeCells('A1:O1')
    const titleCell1 = sheet1.getCell('A1')
    titleCell1.value = 'LAPORAN PURCHASE REQUEST & STATUS TINDAK LANJUT'
    titleCell1.font = { name: 'Arial', size: 14, bold: true, color: { argb: 'FFFFFFFF' } }
    titleCell1.alignment = { horizontal: 'center', vertical: 'middle' }
    titleCell1.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF0F172A' } }
    sheet1.getRow(1).height = 32

    sheet1.mergeCells('A2:O2')
    const subtitleCell1 = sheet1.getCell('A2')
    subtitleCell1.value = `Unit: ${factoryName}   |   Periode: ${periodText}   |   Total Request: ${totalPR}   |   Dicetak: ${new Date().toLocaleString('id-ID')}`
    subtitleCell1.font = { name: 'Arial', size: 10, italic: true, color: { argb: 'FF475569' } }
    subtitleCell1.alignment = { horizontal: 'center', vertical: 'middle' }
    sheet1.getRow(2).height = 20

    // KPI Summary row 4
    sheet1.getCell('A4').value = 'RINGKASAN STATUS PR'
    sheet1.getCell('A4').font = { bold: true, size: 10, color: { argb: 'FF0F172A' } }
    sheet1.getCell('B4').value = `Total: ${totalPR}`
    sheet1.getCell('C4').value = `Pending: ${totalPRPending}`
    sheet1.getCell('D4').value = `Diproses PO: ${totalPRProcessed}`
    sheet1.getCell('E4').value = `Ditolak: ${totalPRRejected}`
    sheet1.getRow(4).font = { bold: true, size: 9, color: { argb: 'FF334155' } }
    sheet1.getRow(4).height = 20

    // Header Table PR
    const prHeaders = [
      'No',
      'No. Purchase Request',
      'Tgl Pengajuan',
      'Tgl Dibutuhkan',
      'Unit / Pabrik',
      'Pemohon',
      'Nama Bahan Baku / Barang',
      'Qty Diminta',
      'Satuan',
      'Gudang Tujuan',
      'Status PR',
      'Tindak Lanjut (No. PO)',
      'Supplier Terpilih',
      'Status Progress PO',
      'Catatan / Alasan Penolakan',
    ]

    const headerRow1 = sheet1.getRow(6)
    headerRow1.values = prHeaders
    headerRow1.height = 28
    headerRow1.eachCell((cell) => {
      cell.font = { name: 'Arial', size: 10, bold: true, color: { argb: 'FFFFFFFF' } }
      cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF1E293B' } }
      cell.alignment = { horizontal: 'center', vertical: 'middle', wrapText: true }
      cell.border = {
        top: { style: 'thin', color: { argb: 'FF94A3B8' } },
        bottom: { style: 'medium', color: { argb: 'FF0F172A' } },
        left: { style: 'thin', color: { argb: 'FF334155' } },
        right: { style: 'thin', color: { argb: 'FF334155' } },
      }
    })

    // Enable AutoFilter on Header Row
    sheet1.autoFilter = {
      from: 'A6',
      to: 'O6',
    }

    let rowIndex1 = 7
    let prSeq = 1

    for (const pr of prs) {
      const items = pr.items && pr.items.length > 0 ? pr.items : [null]

      for (let i = 0; i < items.length; i++) {
        const item = items[i]
        const row = sheet1.getRow(rowIndex1)

        const isFirstItem = i === 0
        const isEven = prSeq % 2 === 0
        const rowBg = isEven ? 'FFF8FAFC' : 'FFFFFFFF'

        row.values = [
          isFirstItem ? prSeq : '',
          isFirstItem ? pr.pr_number : '',
          isFirstItem ? formatDate(pr.request_date) : '',
          isFirstItem ? formatDate(pr.required_date) : '',
          isFirstItem ? pr.factory?.name || 'Karyawan' : '',
          isFirstItem ? pr.created_by_name || '-' : '',
          (item as any)?.item_name || item?.product?.name || (item as any)?.name || '-',
          item?.requested_qty || 0,
          item?.unit || '-',
          item?.factory_warehouse?.name || '-',
          isFirstItem ? formatPRStatus(pr.status) : '',
          isFirstItem ? pr.purchase_order?.po_number || (pr.status === 'PROCESSED' ? 'Telah Dibuat PO' : '-') : '',
          isFirstItem ? (pr.purchase_order?.supplier?.name || '-') : '',
          isFirstItem ? formatProgressStatus(pr.purchase_order?.progress_status) : '',
          isFirstItem
            ? pr.status === 'REJECTED'
              ? `Ditolak: ${pr.rejection_reason || '-'}`
              : pr.notes || '-'
            : item?.notes || '-',
        ]

        row.height = 22
        row.eachCell({ includeEmpty: true }, (cell, colNumber) => {
          cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: rowBg } }
          cell.border = {
            top: { style: 'thin', color: { argb: 'FFE2E8F0' } },
            bottom: { style: 'thin', color: { argb: 'FFE2E8F0' } },
            left: { style: 'thin', color: { argb: 'FFE2E8F0' } },
            right: { style: 'thin', color: { argb: 'FFE2E8F0' } },
          }
          cell.font = { name: 'Arial', size: 9 }

          if (colNumber === 1 || colNumber === 3 || colNumber === 4 || colNumber === 9) {
            cell.alignment = { horizontal: 'center', vertical: 'middle' }
          } else if (colNumber === 8) {
            cell.alignment = { horizontal: 'right', vertical: 'middle' }
            cell.numFmt = '#,##0'
          } else if (colNumber === 11) {
            // Soft Badge Pill Styling
            cell.alignment = { horizontal: 'center', vertical: 'middle' }
            if (pr.status === 'PROCESSED') {
              cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFECFDF5' } }
              cell.font = { name: 'Arial', size: 9, bold: true, color: { argb: 'FF047857' } }
            } else if (pr.status === 'REJECTED') {
              cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFFFF1F2' } }
              cell.font = { name: 'Arial', size: 9, bold: true, color: { argb: 'FFE11D48' } }
            } else {
              cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFFFFBEB' } }
              cell.font = { name: 'Arial', size: 9, bold: true, color: { argb: 'FFB45309' } }
            }
          } else if (colNumber === 12) {
            cell.alignment = { horizontal: 'center', vertical: 'middle' }
            cell.font = { name: 'Arial', size: 9, bold: true, color: { argb: 'FF1D4ED8' } }
          } else {
            cell.alignment = { horizontal: 'left', vertical: 'middle' }
          }
        })

        rowIndex1++
      }
      prSeq++
    }

    // Auto-fit column widths for Sheet 1
    sheet1.columns.forEach((col, idx) => {
      let maxLen = prHeaders[idx] ? prHeaders[idx].length : 12
      col.eachCell?.({ includeEmpty: false }, (cell) => {
        const val = cell.value ? cell.value.toString() : ''
        if (val.length > maxLen && val.length < 50) {
          maxLen = val.length
        }
      })
      col.width = Math.max(maxLen + 4, 11)
    })
  }

  // ===========================================================================
  // SHEET 2: HISTORI PURCHASING & PO
  // ===========================================================================
  if (scope === 'FULL' || scope === 'PO_ONLY') {
    onProgress?.(75, 'Memformat histori transaksi & formula pembukuan PO...')

    const sheet2 = workbook.addWorksheet('History Purchasing & PO', {
      views: [
        {
          state: 'frozen',
          xSplit: 0,
          ySplit: 4, // Freeze header row
          activeCell: 'A5',
          showGridLines: true,
        },
      ],
    })

    // Title block
    sheet2.mergeCells('A1:R1')
    const titleCell2 = sheet2.getCell('A1')
    titleCell2.value = 'HISTORI PEMBELIAN & PURCHASE ORDER (PO)'
    titleCell2.font = { name: 'Arial', size: 14, bold: true, color: { argb: 'FFFFFFFF' } }
    titleCell2.alignment = { horizontal: 'center', vertical: 'middle' }
    titleCell2.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF064E3B' } } // emerald-900
    sheet2.getRow(1).height = 32

    sheet2.mergeCells('A2:R2')
    const subtitleCell2 = sheet2.getCell('A2')
    subtitleCell2.value = `Unit: ${factoryName}   |   Periode: ${periodText}   |   Total PO: ${pos.length}   |   Nilai Pengadaan: Rp ${totalNilaiPengadaan.toLocaleString('id-ID')}`
    subtitleCell2.font = { name: 'Arial', size: 10, italic: true, color: { argb: 'FF475569' } }
    subtitleCell2.alignment = { horizontal: 'center', vertical: 'middle' }
    sheet2.getRow(2).height = 20

    // Header Table PO
    const poHeaders = [
      'No',
      'No. Purchase Order (PO)',
      'Ref. No. PR',
      'Tgl Order PO',
      'Estimasi Tiba',
      'Unit / Pabrik',
      'Nama Supplier',
      'Item Bahan Baku',
      'Qty Pesan',
      'Qty Diterima',
      'Satuan',
      'Estimasi Harga Satuan (Rp)',
      'Subtotal (Rp)',
      'Total Nilai PO (Rp)',
      'Status Approval PO',
      'Status Progress Vendor',
      'Status Pembayaran',
      'Catatan / Keterangan',
    ]

    const headerRow2 = sheet2.getRow(4)
    headerRow2.values = poHeaders
    headerRow2.height = 28
    headerRow2.eachCell((cell) => {
      cell.font = { name: 'Arial', size: 10, bold: true, color: { argb: 'FFFFFFFF' } }
      cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF047857' } } // emerald-700
      cell.alignment = { horizontal: 'center', vertical: 'middle', wrapText: true }
      cell.border = {
        top: { style: 'thin', color: { argb: 'FF6EE7B7' } },
        bottom: { style: 'medium', color: { argb: 'FF064E3B' } },
        left: { style: 'thin', color: { argb: 'FF065F46' } },
        right: { style: 'thin', color: { argb: 'FF065F46' } },
      }
    })

    // Enable AutoFilter on Sheet 2
    sheet2.autoFilter = {
      from: 'A4',
      to: 'R4',
    }

    let rowIndex2 = 5
    let poSeq = 1
    const startDataRow = 5

    for (const po of pos) {
      const items = po.items && po.items.length > 0 ? po.items : [null]

      for (let i = 0; i < items.length; i++) {
        const item = items[i]
        const row = sheet2.getRow(rowIndex2)

        const isFirstItem = i === 0
        const isEven = poSeq % 2 === 0
        const rowBg = isEven ? 'FFF0FDF4' : 'FFFFFFFF'

        // Real Excel Formula for Subtotal: =I{row}*L{row}
        const calculatedSubtotal = (item?.requested_qty || 0) * (item?.estimated_unit_price || 0)

        row.values = [
          isFirstItem ? poSeq : '',
          isFirstItem ? po.po_number : '',
          isFirstItem ? po.purchase_request?.pr_number || '-' : '',
          isFirstItem ? formatDate(po.order_date) : '',
          isFirstItem ? formatDate(po.estimated_arrival) : '',
          isFirstItem ? po.factory?.name || '-' : '',
          isFirstItem ? po.supplier?.name || '-' : '',
          item?.product?.name || (item as any)?.name || '-',
          item?.requested_qty || 0,
          item?.received_qty || 0,
          item?.unit || '-',
          item?.estimated_unit_price || 0,
          calculatedSubtotal,
          isFirstItem ? po.total_estimated_amount || 0 : '',
          isFirstItem ? formatPOStatus(po.status) : '',
          isFirstItem ? formatProgressStatus(po.progress_status) : '',
          isFirstItem ? po.payment_status || 'UNPAID' : '',
          isFirstItem ? po.notes || '-' : item?.notes || '-',
        ]

        row.height = 22
        row.eachCell({ includeEmpty: true }, (cell, colNumber) => {
          cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: rowBg } }
          cell.border = {
            top: { style: 'thin', color: { argb: 'FFE2E8F0' } },
            bottom: { style: 'thin', color: { argb: 'FFE2E8F0' } },
            left: { style: 'thin', color: { argb: 'FFE2E8F0' } },
            right: { style: 'thin', color: { argb: 'FFE2E8F0' } },
          }
          cell.font = { name: 'Arial', size: 9 }

          if (colNumber === 1 || colNumber === 4 || colNumber === 5 || colNumber === 11) {
            cell.alignment = { horizontal: 'center', vertical: 'middle' }
          } else if (colNumber === 2 || colNumber === 3) {
            cell.alignment = { horizontal: 'center', vertical: 'middle' }
            cell.font = { name: 'Arial', size: 9, bold: true }
          } else if (colNumber === 9 || colNumber === 10) {
            cell.alignment = { horizontal: 'right', vertical: 'middle' }
            cell.numFmt = '#,##0'
          } else if (colNumber === 12 || colNumber === 13 || colNumber === 14) {
            cell.alignment = { horizontal: 'right', vertical: 'middle' }
            cell.numFmt = '[$Rp-421] #,##0'
            if (colNumber === 14) {
              cell.font = { name: 'Arial', size: 9, bold: true }
            }
          } else if (colNumber === 15) {
            // Status Approval Soft Pill
            cell.alignment = { horizontal: 'center', vertical: 'middle' }
            if (po.status === 'RECEIVED' || po.status === 'APPROVED') {
              cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFECFDF5' } }
              cell.font = { name: 'Arial', size: 9, bold: true, color: { argb: 'FF047857' } }
            } else if (po.status === 'REJECTED') {
              cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFFFF1F2' } }
              cell.font = { name: 'Arial', size: 9, bold: true, color: { argb: 'FFE11D48' } }
            } else {
              cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFFFFBEB' } }
              cell.font = { name: 'Arial', size: 9, bold: true, color: { argb: 'FFB45309' } }
            }
          } else if (colNumber === 16) {
            // Status Progress Pill
            cell.alignment = { horizontal: 'center', vertical: 'middle' }
            cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFEFF6FF' } }
            cell.font = { name: 'Arial', size: 9, bold: true, color: { argb: 'FF1D4ED8' } }
          } else if (colNumber === 17) {
            // Payment Status Pill
            cell.alignment = { horizontal: 'center', vertical: 'middle' }
            const isPaid = po.payment_status === 'PAID'
            cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: isPaid ? 'FFECFDF5' : 'FFF8FAFC' } }
            cell.font = { name: 'Arial', size: 9, bold: true, color: { argb: isPaid ? 'FF047857' : 'FF64748B' } }
          } else {
            cell.alignment = { horizontal: 'left', vertical: 'middle' }
          }
        })

        rowIndex2++
      }
      poSeq++
    }

    // Dynamic SUM formula row for Sheet 2
    const totalRow = sheet2.getRow(rowIndex2)
    totalRow.getCell(1).value = ''
    totalRow.getCell(7).value = 'TOTAL KESELURUHAN PENGADAAN (PO):'

    // Use Excel formula if there are rows
    if (rowIndex2 > startDataRow) {
      totalRow.getCell(13).value = { formula: `SUM(M${startDataRow}:M${rowIndex2 - 1})` }
      totalRow.getCell(14).value = { formula: `SUM(N${startDataRow}:N${rowIndex2 - 1})` }
    } else {
      totalRow.getCell(14).value = 0
    }

    totalRow.height = 26
    totalRow.eachCell({ includeEmpty: true }, (cell, colNumber) => {
      cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFE2E8F0' } }
      cell.border = {
        top: { style: 'medium', color: { argb: 'FF0F172A' } },
        bottom: { style: 'double', color: { argb: 'FF0F172A' } }, // Official accounting double-underline
      }
      cell.font = { name: 'Arial', size: 10, bold: true, color: { argb: 'FF0F172A' } }
      if (colNumber === 7) {
        cell.alignment = { horizontal: 'right', vertical: 'middle' }
      } else if (colNumber === 13 || colNumber === 14) {
        cell.alignment = { horizontal: 'right', vertical: 'middle' }
        cell.numFmt = '[$Rp-421] #,##0'
      }
    })

    // Auto-fit column widths for Sheet 2
    sheet2.columns.forEach((col, idx) => {
      let maxLen = poHeaders[idx] ? poHeaders[idx].length : 12
      col.eachCell?.({ includeEmpty: false }, (cell) => {
        const val = cell.value ? cell.value.toString() : ''
        if (val.length > maxLen && val.length < 50) {
          maxLen = val.length
        }
      })
      col.width = Math.max(maxLen + 4, 12)
    })
  }

  onProgress?.(90, 'Mengonversi dan mengompresi workbook...')

  // Trigger File Download
  const buffer = await workbook.xlsx.writeBuffer()
  const blob = new Blob([buffer], {
    type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  })
  const dateStr = new Date().toISOString().slice(0, 10)
  const filename = `Laporan_Purchasing_Karsa HR_${dateStr}.xlsx`

  onProgress?.(100, 'Selesai! Mengunduh berkas...')
  saveAs(blob, filename)
  return filename
}
