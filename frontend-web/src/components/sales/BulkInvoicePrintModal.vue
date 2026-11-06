<script setup lang="ts">
import { computed } from 'vue'
import { Printer, X, FileText, CheckCircle2 } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

const props = defineProps<{
  show: boolean
  sales: any[]
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const formatRupiah = (value: number | string | undefined | null) => {
  if (value === undefined || value === null || isNaN(Number(value))) return 'Rp 0'
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(Number(value))
}

const formatDate = (dateStr: string | undefined | null) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'long',
    year: 'numeric',
  })
}

// Terbilang function
const terbilang = (angka: number): string => {
  const bilangan: string[] = [
    '', 'Satu', 'Dua', 'Tiga', 'Empat', 'Lima',
    'Enam', 'Tujuh', 'Delapan', 'Sembilan', 'Sepuluh', 'Sebelas'
  ]
  angka = Math.floor(Math.abs(angka))
  if (angka < 12) return bilangan[angka] || ''
  if (angka < 20) return terbilang(angka - 10) + ' Belas'
  if (angka < 100) return terbilang(Math.floor(angka / 10)) + ' Puluh ' + terbilang(angka % 10)
  if (angka < 200) return 'Seratus ' + terbilang(angka - 100)
  if (angka < 1000) return terbilang(Math.floor(angka / 100)) + ' Ratus ' + terbilang(angka % 100)
  if (angka < 2000) return 'Seribu ' + terbilang(angka - 1000)
  if (angka < 1000000) return terbilang(Math.floor(angka / 1000)) + ' Ribu ' + terbilang(angka % 1000)
  if (angka < 1000000000) return terbilang(Math.floor(angka / 1000000)) + ' Juta ' + terbilang(angka % 1000000)
  if (angka < 1000000000000) return terbilang(Math.floor(angka / 1000000000)) + ' Milyar ' + terbilang(angka % 1000000000)
  return ''
}

const getTerbilang = (amount: number) => {
  if (!amount) return 'Nol Rupiah'
  const text = terbilang(Number(amount)).trim()
  return text ? `${text} Rupiah` : 'Nol Rupiah'
}

const getBuyerName = (sale: any) => {
  const raw = sale.raw || sale
  if (raw.buyer_type === 'EMPLOYEE' && raw.buyer_employee) {
    return `${raw.buyer_employee.first_name} ${raw.buyer_employee.last_name || ''} (Karyawan)`
  }
  if (raw.buyer_name) return raw.buyer_name
  if (raw.store?.name) return raw.store.name
  if (sale.store) return sale.store
  return 'Pelanggan Umum'
}

const getBuyerAddress = (sale: any) => {
  const raw = sale.raw || sale
  return raw.store?.address || raw.buyer_address || '-'
}

const getBuyerPhone = (sale: any) => {
  const raw = sale.raw || sale
  return raw.store?.phone || raw.buyer_phone || ''
}

const getSalesmanName = (sale: any) => {
  const raw = sale.raw || sale
  if (raw.cashier_employee) {
    return `${raw.cashier_employee.first_name} ${raw.cashier_employee.last_name || ''}`
  }
  if (raw.employee) {
    return `${raw.employee.first_name} ${raw.employee.last_name || ''}`
  }
  if (sale.salesman) return sale.salesman
  return '-'
}

const getSalesmanRole = (sale: any) => {
  const raw = sale.raw || sale
  return raw.employee?.job_position?.title || 'SALESMAN'
}

const getCompanyName = (sale: any) => {
  const raw = sale.raw || sale
  return raw.company?.name || raw.employee?.company?.name || 'Perusahaan Demo'
}

const getReceiptNumber = (sale: any) => {
  const raw = sale.raw || sale
  return raw.receipt_no || sale.receiptNo || (raw.id ? raw.id.slice(0, 8).toUpperCase() : 'FAKTUR')
}

const getItems = (sale: any) => {
  const raw = sale.raw || sale
  return raw.items || sale.rawItems || []
}

const getBonusItems = (sale: any) => {
  const raw = sale.raw || sale
  return raw.bonus_items || []
}

const getReturnItems = (sale: any) => {
  const raw = sale.raw || sale
  if (raw.notes) {
    try {
      const parsed = JSON.parse(raw.notes)
      if (parsed.return_items) return parsed.return_items
    } catch {
      // not json
    }
  }
  return []
}

const getReturnTotal = (sale: any) => {
  const returnItems = getReturnItems(sale)
  return returnItems.reduce((sum: number, item: any) => sum + (item.quantity * (item.price || 0)), 0)
}

const getCustomerSignature = (sale: any) => {
  const raw = sale.raw || sale
  return raw.signature_image_url || sale.signature_image_url || ''
}

const printAllInvoices = () => {
  window.print()
}
</script>

<template>
  <div v-if="show && sales && sales.length > 0" class="fixed inset-0 z-50 overflow-y-auto bg-black/60 backdrop-blur-xs flex items-center justify-center p-0 sm:p-4 print:p-0">
    <div class="relative bg-white rounded-xl shadow-2xl max-w-5xl w-full flex flex-col max-h-[95vh] print:max-h-none print:shadow-none print:rounded-none overflow-hidden my-auto">
      
      <!-- Modal Header Toolbar (Non-print) -->
      <div class="flex flex-col sm:flex-row sm:items-center justify-between p-4 sm:px-6 sm:py-4 border-b border-slate-200 print:hidden bg-slate-50 gap-3">
        <div class="flex items-center gap-3">
          <div class="p-2.5 bg-primary/10 rounded-xl text-primary shrink-0">
            <Printer class="w-6 h-6" />
          </div>
          <div>
            <div class="flex items-center gap-2">
              <h3 class="font-black text-slate-900 text-base sm:text-lg">Cetak Masal Faktur Penjualan</h3>
              <span class="px-2.5 py-0.5 rounded-full text-xs font-black bg-primary text-white">
                {{ sales.length }} Faktur
              </span>
            </div>
            <p class="text-xs text-slate-500 font-medium mt-0.5">
              Setiap faktur akan otomatis dicetak pada halaman terpisah (1 faktur per halaman).
            </p>
          </div>
        </div>
        <div class="flex items-center justify-end gap-2.5 w-full sm:w-auto">
          <Button @click="printAllInvoices" class="gap-2 bg-primary hover:bg-primary/90 text-white font-bold shadow-sm px-5 py-2.5 rounded-xl text-xs sm:text-sm">
            <Printer class="w-4 h-4" /> Cetak {{ sales.length }} Faktur (Print)
          </Button>
          <Button variant="outline" size="icon" @click="emit('close')" class="rounded-xl hover:bg-slate-100 shrink-0">
            <X class="w-4 h-4 text-slate-500" />
          </Button>
        </div>
      </div>

      <!-- Printable Bulk Area -->
      <div id="bulk-invoices-container" class="p-4 sm:p-8 overflow-y-auto print:p-0 print:overflow-visible bg-slate-100/60 print:bg-white space-y-8 print:space-y-0">
        
        <div 
          v-for="(sale, sIdx) in sales" 
          :key="sale.id || sIdx"
          class="bulk-invoice-sheet bg-white rounded-2xl print:rounded-none shadow-sm print:shadow-none border border-slate-200 print:border-0 p-6 sm:p-8 print:p-0 mb-6 print:mb-0"
        >
          <!-- Header Perusahaan & Title -->
          <div class="flex flex-col sm:flex-row justify-between items-start border-b-2 border-slate-900 pb-3 mb-4 print:flex-row print:pb-1 print:mb-2 print:border-black gap-2">
            <div>
              <h2 class="text-base sm:text-xl font-black text-slate-900 tracking-tight uppercase italic print:text-black">
                {{ getCompanyName(sale) }}
              </h2>
              <p class="text-[11px] text-slate-500 font-medium print:text-[9px] print:text-black">
                Unit / Cabang: {{ sale.raw?.branch?.name || 'Pabrik & Distribusi' }}
              </p>
            </div>
            <div class="text-left sm:text-right print:text-right">
              <div class="inline-block bg-slate-900 text-white px-3 py-0.5 text-[11px] font-black tracking-wider uppercase rounded mb-0.5 print:border print:border-black print:text-black print:bg-white">
                FAKTUR PENJUALAN
              </div>
              <p class="text-xs font-black text-slate-900 print:text-black print:text-[11px]">
                #{{ getReceiptNumber(sale) }}
              </p>
              <p class="text-[10px] text-slate-500 font-bold print:text-[9px] print:text-black">
                Tgl: {{ formatDate(sale.rawDate || sale.raw?.transaction_date || sale.date) }}
              </p>
            </div>
          </div>

          <!-- Meta Pelanggan & Salesman Grid -->
          <div class="grid grid-cols-2 gap-4 mb-4 text-xs print:text-[10px] print:gap-2 print:mb-2 border border-slate-200 print:border-black rounded-lg p-3 print:p-2 bg-slate-50/50 print:bg-white">
            <div class="space-y-1">
              <span class="font-black text-slate-400 uppercase tracking-widest text-[9px] print:text-[8px] print:text-black block">
                KEPADA PELANGGAN / TOKO:
              </span>
              <p class="font-black text-slate-900 text-xs sm:text-sm print:text-[11px] uppercase print:text-black">
                {{ getBuyerName(sale) }}
              </p>
              <p class="text-[11px] text-slate-600 print:text-[9px] print:text-black">
                {{ getBuyerAddress(sale) }}
              </p>
              <p v-if="getBuyerPhone(sale)" class="text-[10px] text-slate-500 print:text-[8px] print:text-black">
                Telp: {{ getBuyerPhone(sale) }}
              </p>
            </div>
            <div class="text-right space-y-1">
              <span class="font-black text-slate-400 uppercase tracking-widest text-[9px] print:text-[8px] print:text-black block">
                SALESMAN & PEMBAYARAN:
              </span>
              <p class="font-black text-slate-900 text-xs sm:text-sm print:text-[11px] uppercase print:text-black">
                {{ getSalesmanName(sale) }}
              </p>
              <p class="text-[10px] text-slate-500 font-bold uppercase print:text-[8px] print:text-black">
                {{ getSalesmanRole(sale) }}
              </p>
              <div class="mt-1 flex items-center justify-end gap-1.5">
                <span class="text-[10px] font-bold text-slate-600 print:text-[9px] print:text-black">Metode:</span>
                <span class="font-black text-[10px] px-2 py-0.5 rounded bg-slate-200 text-slate-800 print:bg-transparent print:border print:border-black uppercase print:text-[9px]">
                  {{ sale.raw?.payment_method || 'CASH' }}
                </span>
                <span v-if="sale.raw?.payment_method === 'TEMPO' && sale.raw?.payment_due_date" class="text-[9px] font-bold text-amber-800 print:text-black">
                  (JT: {{ formatDate(sale.raw.payment_due_date) }})
                </span>
              </div>
            </div>
          </div>

          <!-- Items Table -->
          <div class="border border-slate-900 print:border-black rounded-lg overflow-hidden mb-3 print:mb-1.5">
            <table class="w-full text-left text-xs print:text-[9px] border-collapse">
              <thead>
                <tr class="bg-slate-100 print:bg-white border-b border-slate-900 print:border-black">
                  <th class="p-2 print:p-1 font-black text-center w-8 border-r border-slate-300 print:border-black print:text-black">NO</th>
                  <th class="p-2 print:p-1 font-black border-r border-slate-300 print:border-black print:text-black">NAMA PRODUK</th>
                  <th class="p-2 print:p-1 font-black text-center w-16 border-r border-slate-300 print:border-black print:text-black">QTY</th>
                  <th class="p-2 print:p-1 font-black text-right w-28 border-r border-slate-300 print:border-black print:text-black">HARGA (Rp)</th>
                  <th class="p-2 print:p-1 font-black text-right w-28 bg-slate-200/50 print:bg-white print:text-black">SUBTOTAL (Rp)</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-200 print:divide-black">
                <tr 
                  v-for="(item, idx) in getItems(sale)" 
                  :key="item.id || idx"
                  class="h-7 print:h-5"
                >
                  <td class="p-2 print:p-0.5 text-center font-bold border-r border-slate-200 print:border-black print:text-black">{{ Number(idx) + 1 }}</td>
                  <td class="p-2 print:p-0.5 font-bold text-slate-900 uppercase border-r border-slate-200 print:border-black print:text-black">{{ item.product?.name || item.product_name || 'Produk' }}</td>
                  <td class="p-2 print:p-0.5 text-center font-bold border-r border-slate-200 print:border-black print:text-black">{{ item.quantity }} {{ item.unit || '' }}</td>
                  <td class="p-2 print:p-0.5 text-right font-medium border-r border-slate-200 print:border-black print:text-black">{{ (item.price_at_transaction || item.price || 0).toLocaleString('id-ID') }}</td>
                  <td class="p-2 print:p-0.5 text-right font-black bg-slate-50/50 print:bg-white print:text-black">{{ (item.subtotal || ((item.price_at_transaction || item.price || 0) * (item.quantity || 0))).toLocaleString('id-ID') }}</td>
                </tr>

                <!-- Bonus Items -->
                <tr 
                  v-for="(bonus, bIdx) in getBonusItems(sale)" 
                  :key="'b-'+bIdx"
                  class="h-7 print:h-5 bg-emerald-50/40 print:bg-white"
                >
                  <td class="p-2 print:p-0.5 text-center font-bold text-emerald-700 border-r border-slate-200 print:border-black print:text-black">*</td>
                  <td class="p-2 print:p-0.5 font-bold text-emerald-800 border-r border-slate-200 print:border-black print:text-black">
                    {{ bonus.bonus_inventory?.name || 'Bonus Promo' }} (BONUS)
                  </td>
                  <td class="p-2 print:p-0.5 text-center font-bold text-emerald-800 border-r border-slate-200 print:border-black print:text-black">{{ bonus.quantity }}</td>
                  <td class="p-2 print:p-0.5 text-right font-medium border-r border-slate-200 print:border-black print:text-black">0</td>
                  <td class="p-2 print:p-0.5 text-right font-black text-emerald-800 bg-emerald-100/30 print:bg-white print:text-black">0</td>
                </tr>

                <!-- Empty rows padding for clean look if few items -->
                <tr 
                  v-for="n in Math.max(0, 3 - (getItems(sale).length + getBonusItems(sale).length))" 
                  :key="'blank-'+n"
                  class="h-6 print:h-4 print:hidden"
                >
                  <td class="p-1 border-r border-slate-200"></td>
                  <td class="p-1 border-r border-slate-200"></td>
                  <td class="p-1 border-r border-slate-200"></td>
                  <td class="p-1 border-r border-slate-200"></td>
                  <td class="p-1 bg-slate-50/20"></td>
                </tr>
              </tbody>
              <tfoot>
                <tr v-if="Number(sale.raw?.discount_amount || 0) > 0" class="bg-rose-50/80 print:bg-white border-t border-slate-300 print:border-black">
                  <td colspan="4" class="p-1.5 text-right font-bold uppercase text-[10px] print:text-[8px] print:text-black">POTONGAN DISKON NOTA:</td>
                  <td class="p-1.5 text-right font-black text-rose-700 print:text-black">-{{ Number(sale.raw?.discount_amount || 0).toLocaleString('id-ID') }}</td>
                </tr>
                <tr v-if="getReturnTotal(sale) > 0" class="bg-rose-50/80 print:bg-white border-t border-slate-300 print:border-black">
                  <td colspan="4" class="p-1.5 text-right font-bold uppercase text-[10px] print:text-[8px] print:text-black">POTONGAN RETUR:</td>
                  <td class="p-1.5 text-right font-black text-rose-700 print:text-black">-{{ getReturnTotal(sale).toLocaleString('id-ID') }}</td>
                </tr>
                <tr class="bg-slate-900 text-white font-black print:bg-white print:text-black border-t-2 border-slate-900 print:border-black">
                  <td colspan="4" class="p-2 print:p-1 text-right uppercase tracking-wider text-xs print:text-[10px] print:font-bold">
                    {{ (Number(sale.raw?.discount_amount || 0) > 0 || getReturnTotal(sale) > 0) ? 'TOTAL BAYAR AKHIR:' : 'TOTAL TAGIHAN:' }}
                  </td>
                  <td class="p-2 print:p-1 text-right text-sm print:text-[11px] print:font-black">
                    {{ formatRupiah(sale.amount || sale.raw?.total_amount) }}
                  </td>
                </tr>
              </tfoot>
            </table>
          </div>

          <!-- Terbilang Box -->
          <div class="mb-4 print:mb-2 p-2 bg-slate-50 print:bg-white border border-slate-200 print:border-black rounded text-[11px] print:text-[9px]">
            <span class="font-bold text-slate-500 print:text-black">Terbilang: </span>
            <span class="font-bold text-slate-900 italic print:text-black"># {{ getTerbilang(sale.amount || sale.raw?.total_amount) }} #</span>
          </div>

          <!-- Signatures Footer -->
          <div class="grid grid-cols-4 gap-2 pt-2 border-t-2 border-slate-900 print:border-black text-center text-[10px] print:text-[8px]">
            <div>
              <p class="font-black text-slate-500 uppercase print:text-black">PENERIMA / TOKO</p>
              <div class="h-12 print:h-8 flex items-center justify-center border border-dashed border-slate-200 print:border-0 my-1">
                <img v-if="getCustomerSignature(sale)" :src="getCustomerSignature(sale)" class="max-h-10 print:max-h-6 object-contain" />
                <span v-else class="text-[8px] text-slate-300 print:text-transparent">Ttd</span>
              </div>
              <p class="font-black text-slate-800 uppercase print:text-black truncate">
                ( {{ getBuyerName(sale).slice(0, 20) }} )
              </p>
            </div>
            <div>
              <p class="font-black text-slate-500 uppercase print:text-black">SALESMAN / PENGIRIM</p>
              <div class="h-12 print:h-8 flex items-center justify-center my-1"></div>
              <p class="font-black text-slate-800 uppercase print:text-black truncate">
                ( {{ getSalesmanName(sale).slice(0, 20) }} )
              </p>
            </div>
            <div>
              <p class="font-black text-slate-500 uppercase print:text-black">ADMIN NOTA</p>
              <div class="h-12 print:h-8 flex items-center justify-center my-1"></div>
              <p class="font-black text-slate-800 uppercase print:text-black">
                ( .................... )
              </p>
            </div>
            <div>
              <p class="font-black text-slate-500 uppercase print:text-black">KASIR</p>
              <div class="h-12 print:h-8 flex items-center justify-center my-1"></div>
              <p class="font-black text-slate-800 uppercase print:text-black">
                ( .................... )
              </p>
            </div>
          </div>

        </div>

      </div>

    </div>
  </div>
</template>

<style scoped>
@media print {
  @page {
    size: A4 portrait;
    margin: 10mm;
  }
  body, html {
    background: white !important;
    color: black !important;
    overflow: visible !important;
  }
  .print\:hidden {
    display: none !important;
  }
  .bulk-invoice-sheet {
    page-break-after: always !important;
    break-after: page !important;
    width: 100% !important;
    min-height: 13.5cm !important;
    box-sizing: border-box !important;
    background: white !important;
    color: black !important;
    display: block !important;
    padding: 0 !important;
    margin: 0 !important;
  }
  .bulk-invoice-sheet:last-child {
    page-break-after: avoid !important;
    break-after: avoid !important;
  }
}
</style>
