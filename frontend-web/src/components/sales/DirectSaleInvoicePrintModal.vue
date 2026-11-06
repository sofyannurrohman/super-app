<script setup lang="ts">
import { computed } from 'vue'
import { Printer, X, CheckCircle2 } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

const props = defineProps<{
  show: boolean
  sale: any
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

const terbilangText = computed(() => {
  if (!props.sale?.total_amount) return 'Nol Rupiah'
  const text = terbilang(Number(props.sale.total_amount)).trim()
  return text ? `${text} Rupiah` : 'Nol Rupiah'
})

const buyerDisplayName = computed(() => {
  if (!props.sale) return '-'
  if (props.sale.buyer_type === 'EMPLOYEE' && props.sale.buyer_employee) {
    return `${props.sale.buyer_employee.first_name} ${props.sale.buyer_employee.last_name} (Karyawan)`
  }
  if (props.sale.buyer_name) return props.sale.buyer_name
  if (props.sale.store?.name) return props.sale.store.name
  return 'Pelanggan Umum'
})

const cashierDisplayName = computed(() => {
  if (!props.sale) return '-'
  if (props.sale.cashier_employee) {
    return `${props.sale.cashier_employee.first_name} ${props.sale.cashier_employee.last_name}`
  }
  if (props.sale.employee) {
    return `${props.sale.employee.first_name} ${props.sale.employee.last_name}`
  }
  return '-'
})

const totalReturnAmount = computed(() => {
  if (!props.sale?.returns) return 0
  return props.sale.returns
    .filter((r: any) => r.return_type === 'POTONG_NOTA')
    .reduce((sum: number, r: any) => sum + (Number(r.total_amount) || 0), 0)
})

const initialSubtotal = computed(() => {
  return Number(props.sale?.total_amount || 0) + Number(props.sale?.discount_amount || 0) + totalReturnAmount.value
})

const printInvoice = () => {
  window.print()
}
</script>

<template>
  <div v-if="show && sale" class="fixed inset-0 z-50 overflow-y-auto bg-black/60 backdrop-blur-xs flex items-center justify-center p-4">
    <div class="relative bg-white rounded-xl shadow-2xl max-w-4xl w-full flex flex-col max-h-[92vh] overflow-hidden">
      <!-- Modal Header (Non-print) -->
      <div class="flex flex-col sm:flex-row sm:items-center justify-between p-4 sm:px-6 sm:py-4 border-b border-slate-200 print:hidden bg-slate-50 gap-3">
        <div class="flex items-center gap-2">
          <div class="p-2 bg-primary/10 rounded-lg text-primary shrink-0">
            <Printer class="w-5 h-5" />
          </div>
          <div>
            <h3 class="font-bold text-slate-900 text-base sm:text-lg">Cetak Faktur Penjualan Langsung</h3>
            <p class="text-xs text-slate-500">No. Faktur: {{ sale.receipt_no }}</p>
          </div>
        </div>
        <div class="flex items-center justify-end gap-2 w-full sm:w-auto">
          <Button @click="printInvoice" class="gap-2 bg-primary hover:bg-primary/90 text-white font-medium shadow-sm flex-1 sm:flex-initial text-xs sm:text-sm py-2">
            <Printer class="w-4 h-4" /> Cetak Faktur (Print)
          </Button>
          <Button variant="outline" size="icon" @click="emit('close')" class="rounded-lg hover:bg-slate-100 shrink-0">
            <X class="w-4 h-4 text-slate-500" />
          </Button>
        </div>
      </div>

      <!-- Invoice Printable Content -->
      <div class="p-4 sm:p-8 overflow-y-auto print:p-0 print:overflow-visible text-slate-800 invoice-paper">
        <!-- Header Info -->
        <div class="border-b-2 border-slate-800 pb-4 mb-4">
          <div class="flex flex-col sm:flex-row justify-between items-start print:flex-row gap-2 sm:gap-0">
            <div>
              <h1 class="text-lg sm:text-xl font-black text-slate-900 tracking-wide uppercase">
                {{ sale.company?.name || 'Perusahaan Demo' }}
              </h1>
              <p class="text-xs text-slate-600 font-medium mt-0.5">
                Unit / Pabrik: {{ sale.branch?.name || 'Pabrik & Gudang Pusat' }}
              </p>
              <p class="text-xs text-slate-500">
                Alamat: {{ sale.branch?.address || 'Jl. Raya Industri Karsa HR No. 1' }}
              </p>
            </div>
            <div class="text-left sm:text-right print:text-right">
              <div class="inline-block bg-slate-900 text-white px-3 py-1 text-xs font-bold tracking-wider uppercase rounded mb-1">
                FAKTUR PENJUALAN
              </div>
              <p class="text-sm font-black text-slate-900">{{ sale.receipt_no }}</p>
              <p class="text-xs text-slate-600 font-medium">Tanggal: {{ formatDate(sale.transaction_date) }}</p>
            </div>
          </div>
        </div>

        <!-- Buyer & Payment Meta Grid -->
        <div class="grid grid-cols-1 sm:grid-cols-2 print:grid-cols-2 gap-3 sm:gap-6 bg-slate-50 border border-slate-200 rounded-lg p-3 text-xs mb-4">
          <div>
            <span class="font-bold text-slate-500 uppercase tracking-wider block text-[10px]">KEPADA PEMBELI:</span>
            <p class="font-black text-slate-900 text-sm mt-0.5">{{ buyerDisplayName }}</p>
            <p v-if="sale.buyer_phone" class="text-slate-600 font-medium mt-0.5">Telp: {{ sale.buyer_phone }}</p>
            <p class="text-[11px] text-slate-500 mt-0.5">
              Tipe: <span class="font-bold text-slate-700">{{ sale.buyer_type || 'Walk-in' }}</span>
            </p>
          </div>
          <div>
            <span class="font-bold text-slate-500 uppercase tracking-wider block text-[10px]">INFO PEMBAYARAN & KASIR:</span>
            <div class="flex items-center gap-2 mt-0.5">
              <span class="text-slate-600">Metode Bayar:</span>
              <span
                :class="[
                  'font-black px-2 py-0.5 rounded border text-xs',
                  sale.payment_method === 'TEMPO'
                    ? 'text-amber-700 bg-amber-50 border-amber-200'
                    : 'text-emerald-700 bg-emerald-50 border-emerald-200'
                ]"
              >
                {{ sale.payment_method === 'TEMPO' ? 'KREDIT / TEMPO' : (sale.payment_method || 'CASH') }}
              </span>
            </div>
            <p v-if="sale.payment_due_date || sale.payment_method === 'TEMPO'" class="text-amber-800 font-bold mt-1 text-[11px]">
              Jatuh Tempo: {{ formatDate(sale.payment_due_date) }}
            </p>
            <p class="text-slate-600 mt-1">
              Kasir: <span class="font-bold text-slate-800">{{ cashierDisplayName }}</span>
            </p>
            <p v-if="sale.kas_pabrik" class="text-slate-600 mt-0.5">
              Kas Penerima: <span class="font-medium text-slate-800">{{ sale.kas_pabrik.kas_name }}</span>
            </p>
          </div>
        </div>

        <!-- Items Table -->
        <div class="border border-slate-300 rounded-lg overflow-x-auto mb-4">
          <table class="w-full text-left text-xs border-collapse min-w-[520px] sm:min-w-full">
            <thead>
              <tr class="bg-slate-100 border-b border-slate-300 font-bold text-slate-700">
                <th class="p-2 border-r border-slate-300 text-center w-8">No</th>
                <th class="p-2 border-r border-slate-300 w-24">Kode</th>
                <th class="p-2 border-r border-slate-300">Deskripsi Barang</th>
                <th class="p-2 border-r border-slate-300 text-center w-16">Qty</th>
                <th class="p-2 border-r border-slate-300 text-center w-16">Satuan</th>
                <th class="p-2 border-r border-slate-300 text-right w-24">Harga</th>
                <th class="p-2 border-r border-slate-300 text-center w-14">Disc %</th>
                <th class="p-2 text-right w-28">Total</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-200">
              <tr v-for="(item, idx) in sale.items" :key="item.id || idx" class="text-slate-800">
                <td class="p-2 border-r border-slate-200 text-center">{{ Number(idx) + 1 }}</td>
                <td class="p-2 border-r border-slate-200 font-mono text-[11px]">{{ item.product?.sku || '-' }}</td>
                <td class="p-2 border-r border-slate-200 font-medium">{{ item.product?.name || 'Produk' }}</td>
                <td class="p-2 border-r border-slate-200 text-center font-bold">{{ item.quantity }}</td>
                <td class="p-2 border-r border-slate-200 text-center text-slate-600">{{ item.product?.unit || 'Pcs' }}</td>
                <td class="p-2 border-r border-slate-200 text-right">{{ formatRupiah(item.price_at_transaction || item.price) }}</td>
                <td class="p-2 border-r border-slate-200 text-center">{{ item.discount_percent || 0 }}%</td>
                <td class="p-2 text-right font-bold">{{ formatRupiah(item.subtotal || (item.quantity * item.price_at_transaction)) }}</td>
              </tr>
              <!-- Bonus Items (if any) -->
              <tr v-for="(bItem, bIdx) in sale.bonus_items" :key="bItem.id || bIdx" class="bg-amber-50/50 text-amber-900 text-[11px]">
                <td class="p-2 border-r border-slate-200 text-center font-bold text-amber-600">B</td>
                <td class="p-2 border-r border-slate-200 font-mono">-</td>
                <td class="p-2 border-r border-slate-200 font-medium">
                  [BONUS] {{ bItem.bonus_inventory?.name || 'Item Bonus' }} ({{ bItem.notes || 'Promo' }})
                </td>
                <td class="p-2 border-r border-slate-200 text-center font-bold">{{ bItem.quantity }}</td>
                <td class="p-2 border-r border-slate-200 text-center">Pcs</td>
                <td class="p-2 border-r border-slate-200 text-right">Rp 0</td>
                <td class="p-2 border-r border-slate-200 text-center">-</td>
                <td class="p-2 text-right font-bold">Rp 0</td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Footer Calculation & Terbilang -->
        <div class="grid grid-cols-1 sm:grid-cols-2 print:grid-cols-2 gap-4 items-start mb-6">
          <div class="space-y-2 text-xs">
            <div class="p-2.5 bg-slate-50 border border-slate-200 rounded">
              <span class="font-bold text-slate-500 uppercase tracking-wider block text-[10px]">TERBILANG:</span>
              <p class="font-bold text-slate-900 italic mt-0.5">"{{ terbilangText }}"</p>
            </div>
            <div v-if="sale.notes" class="text-slate-600">
              <span class="font-bold">Catatan:</span> {{ sale.notes }}
            </div>
          </div>

          <div class="space-y-1.5 text-xs text-right">
            <div class="flex justify-between py-1 border-b border-slate-100">
              <span class="text-slate-600">Subtotal:</span>
              <span class="font-bold text-slate-800">{{ formatRupiah(initialSubtotal) }}</span>
            </div>
            <div v-if="sale.discount_amount > 0" class="flex justify-between py-1 border-b border-slate-100 text-rose-600">
              <span>Potongan Diskon:</span>
              <span class="font-bold">- {{ formatRupiah(sale.discount_amount) }}</span>
            </div>
            <div v-if="totalReturnAmount > 0" class="flex justify-between py-1 border-b border-slate-100 text-rose-600 font-medium">
              <span>Potongan Retur Barang:</span>
              <span class="font-bold">- {{ formatRupiah(totalReturnAmount) }}</span>
            </div>
            <div class="flex justify-between py-1.5 border-b-2 border-slate-800 text-sm">
              <span class="font-black text-slate-900 uppercase">TOTAL AKHIR:</span>
              <span class="font-black text-slate-900 text-base">{{ formatRupiah(sale.total_amount) }}</span>
            </div>
            <div class="flex justify-between py-1 text-emerald-700">
              <span class="font-bold">Jumlah Dibayar:</span>
              <span class="font-black">{{ formatRupiah(sale.paid_amount || sale.total_amount) }}</span>
            </div>
          </div>
        </div>

        <!-- Signatures Column (3 columns: Pembeli, Kasir, Gudang) -->
        <div class="grid grid-cols-1 sm:grid-cols-3 print:grid-cols-3 gap-6 sm:gap-4 text-center text-xs mt-6 sm:mt-8 pt-4 border-t border-slate-200">
          <div class="space-y-6 sm:space-y-12">
            <p class="font-bold text-slate-700">Diterima Oleh (Pembeli),</p>
            <div class="border-b border-slate-400 mx-6"></div>
            <p class="font-medium text-slate-600">({{ buyerDisplayName }})</p>
          </div>
          <div class="space-y-6 sm:space-y-12">
            <p class="font-bold text-slate-700">Petugas Gudang (Penyerah),</p>
            <div class="border-b border-slate-400 mx-6"></div>
            <p class="font-medium text-slate-600">( ........................................ )</p>
          </div>
          <div class="space-y-6 sm:space-y-12">
            <p class="font-bold text-slate-700">Kasir Penerima,</p>
            <div class="border-b border-slate-400 mx-6"></div>
            <p class="font-medium text-slate-600">({{ cashierDisplayName }})</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
@media print {
  body * {
    visibility: hidden;
  }
  .invoice-paper, .invoice-paper * {
    visibility: visible;
  }
  .invoice-paper {
    position: absolute;
    left: 0;
    top: 0;
    width: 100%;
    padding: 10mm;
    box-shadow: none !important;
  }
}
</style>
