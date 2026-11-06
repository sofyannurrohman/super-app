<script setup lang="ts">
import { computed } from 'vue'
import { Printer, X, Building2, Truck, CheckCircle2, ShieldCheck, User, Calendar, CreditCard, Package } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

const props = defineProps<{
  show: boolean
  transfer: any
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const formatRupiah = (val: number | string | undefined | null) => {
  if (val === undefined || val === null || isNaN(Number(val))) return 'Rp 0'
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(Number(val))
}

const formatDate = (dateStr?: string | null) => {
  if (!dateStr) return '-'
  try {
    const d = new Date(dateStr)
    return d.toLocaleDateString('id-ID', {
      day: 'numeric',
      month: 'long',
      year: 'numeric',
    })
  } catch {
    return dateStr
  }
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
  const total = Number(props.transfer?.total_amount || 0)
  if (!total) return 'Nol Rupiah'
  const text = terbilang(total).trim()
  return text ? `${text} Rupiah` : 'Nol Rupiah'
})

const deliveryOrderNo = computed(() => {
  if (!props.transfer) return '-'
  return props.transfer.delivery_order_no || `SJ-${props.transfer.id?.substring(0, 8).toUpperCase()}`
})

const qrCodeUrl = computed(() => {
  const code = deliveryOrderNo.value
  return `https://api.qrserver.com/v1/create-qr-code/?size=120x120&margin=0&data=${encodeURIComponent(code)}`
})

const companyName = computed(() => {
  return props.transfer?.from_factory?.company?.name || 'Perusahaan Demo'
})

const driverName = computed(() => {
  if (!props.transfer) return '-'
  if (props.transfer.driver) {
    return `${props.transfer.driver.first_name} ${props.transfer.driver.last_name || ''}`.trim()
  }
  if (props.transfer.driver_name) return props.transfer.driver_name
  return '-'
})

const vehicleInfo = computed(() => {
  if (!props.transfer) return '-'
  if (props.transfer.vehicle) {
    return `${props.transfer.vehicle.name || 'Armada'} (${props.transfer.vehicle.plate_number || props.transfer.vehicle.license_plate || '-'})`
  }
  if (props.transfer.vehicle_plate) return props.transfer.vehicle_plate
  return '-'
})

interface InvoiceItemLine {
  sku: string
  name: string
  batch_no: string
  quantity: number
  unit: string
  price: number
  subtotal: number
}

const itemsList = computed<InvoiceItemLine[]>(() => {
  if (!props.transfer) return []
  if (props.transfer.items && Array.isArray(props.transfer.items) && props.transfer.items.length > 0) {
    return props.transfer.items.map((it: any) => {
      const qty = Number(it.quantity) || 1
      const total = Number(it.total_amount) || 0
      const price = Number(it.price) || (total > 0 ? total / qty : (it.product?.factory_price || it.product?.base_price || 0))
      const subtotal = total > 0 ? total : price * qty
      return {
        sku: it.product?.sku || it.sku || '-',
        name: it.product?.name || it.product_name || 'Produk',
        batch_no: it.batch_no || 'DEFAULT',
        quantity: qty,
        unit: it.unit || it.product?.unit || 'PCS',
        price: price,
        subtotal: subtotal,
      }
    })
  }

  // Single item fallback
  const qty = Number(props.transfer.quantity) || 1
  const total = Number(props.transfer.total_amount) || 0
  const price = total > 0 ? total / qty : (props.transfer.product?.factory_price || props.transfer.product?.base_price || 0)
  return [{
    sku: props.transfer.product?.sku || props.transfer.sku || '-',
    name: props.transfer.product?.name || props.transfer.product_name || 'Produk',
    batch_no: props.transfer.batch_no || 'DEFAULT',
    quantity: qty,
    unit: props.transfer.unit || props.transfer.product?.unit || 'PCS',
    price: price,
    subtotal: total > 0 ? total : price * qty,
  }]
})

const totalQuantity = computed(() => {
  if (!props.transfer) return 0
  if (props.transfer.total_quantity) return Number(props.transfer.total_quantity)
  return itemsList.value.reduce((sum: number, it: InvoiceItemLine) => sum + (Number(it.quantity) || 0), 0)
})

const calculatedTotalAmount = computed(() => {
  if (!props.transfer) return 0
  if (props.transfer.total_amount) return Number(props.transfer.total_amount)
  return itemsList.value.reduce((sum: number, it: InvoiceItemLine) => sum + (Number(it.subtotal) || 0), 0)
})

const handlePrint = () => {
  window.print()
}
</script>

<template>
  <div v-if="show && transfer" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/70 backdrop-blur-xs p-2 sm:p-4 overflow-y-auto">
    <div class="relative w-full max-w-4xl bg-white rounded-2xl shadow-2xl border border-slate-200 overflow-hidden my-auto flex flex-col max-h-[95vh] print:max-h-none print:shadow-none print:border-none print:m-0 print:rounded-none">
      
      <!-- Toolbar (Non-print) -->
      <div class="px-5 py-3.5 bg-slate-900 text-white flex items-center justify-between border-b border-slate-800 print:hidden shrink-0">
        <div class="flex items-center gap-3">
          <div class="p-2 bg-indigo-500/20 text-indigo-400 rounded-lg border border-indigo-400/20">
            <Printer class="w-5 h-5" />
          </div>
          <div>
            <div class="flex items-center gap-2">
              <h3 class="font-bold text-sm sm:text-base text-white">Preview Faktur Pengiriman Cabang</h3>
              <span
                :class="[
                  'px-2 py-0.5 rounded-full text-[10px] font-black uppercase tracking-wider',
                  transfer.status === 'RECEIVED' ? 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/30' :
                  transfer.status === 'SHIPPED' ? 'bg-indigo-500/20 text-indigo-300 border border-indigo-500/30' :
                  transfer.status === 'APPROVED' ? 'bg-blue-500/20 text-blue-300 border border-blue-500/30' :
                  'bg-amber-500/20 text-amber-300 border border-amber-500/30'
                ]"
              >
                {{ transfer.status || 'DRAFT' }}
              </span>
            </div>
            <p class="text-xs text-slate-400 mt-0.5">No. DO: <span class="font-mono text-slate-200 font-bold">{{ deliveryOrderNo }}</span></p>
          </div>
        </div>

        <div class="flex items-center gap-2.5">
          <Button
            @click="handlePrint"
            class="bg-indigo-600 hover:bg-indigo-700 text-white font-bold text-xs rounded-xl px-4 py-2 flex items-center gap-1.5 shadow-sm shadow-indigo-600/30 transition-all cursor-pointer"
          >
            <Printer class="w-4 h-4" /> Cetak Faktur (Print)
          </Button>
          <button
            @click="emit('close')"
            class="p-2 text-slate-400 hover:text-white rounded-xl hover:bg-white/10 transition-colors"
          >
            <X class="w-5 h-5" />
          </button>
        </div>
      </div>

      <!-- Printable Invoice Sheet Wrapper -->
      <div class="overflow-y-auto p-4 sm:p-8 bg-slate-100/70 print:bg-white print:p-0 print:overflow-visible">
        <div class="invoice-paper bg-white rounded-xl print:rounded-none shadow-sm print:shadow-none border border-slate-200 print:border-none p-6 sm:p-8 print:p-0 text-slate-900 font-sans mx-auto max-w-[210mm]">
          
          <!-- Header Perusahaan & Judul Dokumen -->
          <div class="flex flex-row justify-between items-start border-b-2 border-slate-900 pb-4 mb-4">
            <div class="flex items-start gap-3.5">
              <img
                src="/brand.svg"
                alt="Logo perusahaan"
                class="w-14 h-14 object-contain shrink-0 mt-0.5"
                onerror="this.style.display='none'"
              />
              <div>
                <h1 class="text-lg sm:text-xl font-black tracking-tight text-slate-900 uppercase">
                  {{ companyName }}
                </h1>
                <p class="text-xs font-bold text-indigo-900 mt-0.5">
                  Unit Pabrik / Pengirim: {{ transfer.from_factory?.name || 'Pabrik & Gudang Pusat Karsa HR' }}
                </p>
                <p class="text-[11px] text-slate-600 max-w-md mt-0.5 leading-tight">
                  {{ transfer.from_factory?.address || transfer.from_factory?.location || 'Pusat Operasional Pabrik & Logistik Distribusi Antar-Cabang' }}
                </p>
              </div>
            </div>

            <div class="text-right shrink-0">
              <div class="inline-block px-3 py-1 bg-slate-900 text-white text-[11px] font-black rounded uppercase tracking-wider mb-1 print:bg-black">
                FAKTUR PENGIRIMAN CABANG
              </div>
              <div class="text-sm font-black font-mono text-slate-900 tracking-tight">
                {{ deliveryOrderNo }}
              </div>
              <div class="text-[11px] text-slate-600 font-medium mt-0.5">
                Tgl Terbit: <span class="font-bold text-slate-800">{{ formatDate(transfer.shipped_at || transfer.created_at) }}</span>
              </div>
            </div>
          </div>

          <!-- Meta Information (3 Columns Grid) -->
          <div class="grid grid-cols-3 gap-3 bg-slate-50 print:bg-white p-3.5 rounded-xl border border-slate-200 print:border-slate-400 text-xs mb-4">
            <!-- Kolom 1: Pabrik Pengirim -->
            <div class="space-y-1">
              <span class="font-black text-slate-400 uppercase tracking-wider text-[9px] block">
                PENGIRIM (PABRIK):
              </span>
              <div class="font-black text-slate-900 text-xs uppercase">
                {{ transfer.from_factory?.name || 'Pabrik Utama Pusat' }}
              </div>
              <div class="text-[11px] text-slate-600 leading-tight">
                {{ transfer.from_factory?.address || 'Pusat Logistik Pabrik' }}
              </div>
              <div v-if="transfer.approved_by_name" class="text-[10px] text-slate-500 pt-0.5">
                Admin: <span class="font-semibold text-slate-700">{{ transfer.approved_by_name }}</span>
              </div>
            </div>

            <!-- Kolom 2: Cabang Penerima -->
            <div class="space-y-1 border-l border-slate-200 print:border-slate-300 pl-3">
              <span class="font-black text-slate-400 uppercase tracking-wider text-[9px] block">
                PENERIMA (CABANG TUJUAN):
              </span>
              <div class="font-black text-slate-900 text-xs uppercase">
                {{ transfer.to_branch?.name || 'Cabang Tujuan' }}
              </div>
              <div class="text-[11px] text-slate-600 leading-tight">
                {{ transfer.to_branch?.address || 'Gudang Cabang Distribusi' }}
              </div>
              <div class="text-[10px] text-slate-500 pt-0.5">
                Tujuan: <span class="font-semibold text-indigo-900">Stok Gudang Cabang</span>
              </div>
            </div>

            <!-- Kolom 3: Logistik & Syarat Tagihan -->
            <div class="space-y-1 border-l border-slate-200 print:border-slate-300 pl-3">
              <span class="font-black text-slate-400 uppercase tracking-wider text-[9px] block">
                ARMADA & PEMBAYARAN:
              </span>
              <div class="text-[11px] text-slate-700">
                Sopir: <span class="font-bold text-slate-900">{{ driverName }}</span>
              </div>
              <div class="text-[11px] text-slate-700">
                Armada: <span class="font-bold text-slate-900">{{ vehicleInfo }}</span>
              </div>
              <div class="flex items-center gap-1.5 pt-0.5">
                <span class="text-[10px] text-slate-500">Metode:</span>
                <span class="font-black text-[10px] px-2 py-0.5 bg-indigo-50 text-indigo-800 border border-indigo-200 rounded print:bg-transparent print:border-black uppercase">
                  {{ transfer.payment_method || 'NET' }}
                </span>
                <span v-if="transfer.payment_due_at" class="text-[10px] font-bold text-amber-700">
                  (JT: {{ formatDate(transfer.payment_due_at) }})
                </span>
              </div>
            </div>
          </div>

          <!-- Line Items Table -->
          <div class="border border-slate-900 print:border-black rounded-lg overflow-hidden mb-4">
            <table class="w-full text-left text-xs border-collapse">
              <thead>
                <tr class="bg-slate-100 print:bg-white text-slate-800 font-black text-[11px] border-b border-slate-900 print:border-black">
                  <th class="p-2 border-r border-slate-300 print:border-black w-8 text-center">NO</th>
                  <th class="p-2 border-r border-slate-300 print:border-black w-24">SKU</th>
                  <th class="p-2 border-r border-slate-300 print:border-black">NAMA PRODUK / VARIAN</th>
                  <th class="p-2 border-r border-slate-300 print:border-black w-28 text-center">BATCH NO</th>
                  <th class="p-2 border-r border-slate-300 print:border-black w-20 text-center">QTY</th>
                  <th class="p-2 border-r border-slate-300 print:border-black w-16 text-center">SATUAN</th>
                  <th class="p-2 border-r border-slate-300 print:border-black w-28 text-right">HARGA PABRIK</th>
                  <th class="p-2 text-right w-32 bg-slate-200/50 print:bg-white font-black">SUBTOTAL (RP)</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-200 print:divide-slate-400">
                <tr v-for="(it, idx) in itemsList" :key="idx" class="hover:bg-slate-50/50">
                  <td class="p-2 border-r border-slate-200 print:border-black text-center font-bold">{{ Number(idx) + 1 }}</td>
                  <td class="p-2 border-r border-slate-200 print:border-black font-mono text-[11px] text-slate-600">{{ it.sku }}</td>
                  <td class="p-2 border-r border-slate-200 print:border-black font-bold text-slate-900 uppercase">{{ it.name }}</td>
                  <td class="p-2 border-r border-slate-200 print:border-black text-center font-mono text-[11px]">{{ it.batch_no }}</td>
                  <td class="p-2 border-r border-slate-200 print:border-black text-center font-black text-slate-900">{{ it.quantity }}</td>
                  <td class="p-2 border-r border-slate-200 print:border-black text-center font-medium text-slate-600 uppercase">{{ it.unit }}</td>
                  <td class="p-2 border-r border-slate-200 print:border-black text-right font-medium">{{ formatRupiah(it.price) }}</td>
                  <td class="p-2 text-right font-black text-slate-900 bg-slate-50/50 print:bg-white">{{ formatRupiah(it.subtotal) }}</td>
                </tr>

                <!-- Empty rows padding if few items for clean receipt layout -->
                <tr
                  v-for="n in Math.max(0, 3 - itemsList.length)"
                  :key="'blank-' + n"
                  class="h-6 print:hidden"
                >
                  <td class="p-1 border-r border-slate-200"></td>
                  <td class="p-1 border-r border-slate-200"></td>
                  <td class="p-1 border-r border-slate-200"></td>
                  <td class="p-1 border-r border-slate-200"></td>
                  <td class="p-1 border-r border-slate-200"></td>
                  <td class="p-1 border-r border-slate-200"></td>
                  <td class="p-1 border-r border-slate-200"></td>
                  <td class="p-1 bg-slate-50/20"></td>
                </tr>
              </tbody>
              <tfoot>
                <tr class="bg-slate-100 print:bg-white font-bold text-slate-900 border-t-2 border-slate-900 print:border-black">
                  <td colspan="4" class="p-2 text-right uppercase text-[10px] tracking-wider border-r border-slate-300 print:border-black font-black">
                    TOTAL KUANTITAS:
                  </td>
                  <td class="p-2 text-center font-black text-indigo-900 border-r border-slate-300 print:border-black">
                    {{ totalQuantity }}
                  </td>
                  <td class="p-2 text-right uppercase text-[10px] tracking-wider border-r border-slate-300 print:border-black font-black" colspan="2">
                    TOTAL TAGIHAN PABRIK (PIUTANG CABANG):
                  </td>
                  <td class="p-2 text-right text-sm font-black text-indigo-950 print:text-black">
                    {{ formatRupiah(calculatedTotalAmount) }}
                  </td>
                </tr>
              </tfoot>
            </table>
          </div>

          <!-- Terbilang & Catatan Grid -->
          <div class="grid grid-cols-3 gap-3 mb-5 items-start">
            <div class="col-span-2 space-y-2">
              <div class="p-2.5 bg-slate-50 print:bg-white border border-slate-200 print:border-slate-400 rounded-lg text-xs">
                <span class="font-black text-slate-400 uppercase tracking-wider text-[9px] block">
                  TERBILANG:
                </span>
                <p class="font-black text-slate-900 italic text-[11px] mt-0.5">
                  # {{ terbilangText }} #
                </p>
              </div>

              <div class="text-[10px] text-slate-600 bg-slate-50/60 print:bg-white p-2.5 rounded-lg border border-slate-200 print:border-slate-400">
                <p class="font-bold text-slate-800 mb-0.5">Keterangan / Catatan:</p>
                <p class="leading-relaxed">{{ transfer.notes || 'Pengiriman stok barang antar unit dari Pabrik Utama ke Gudang Cabang.' }}</p>
                <p class="text-[9px] text-slate-500 italic mt-1 pt-1 border-t border-slate-200 print:border-slate-300">
                  * Barang telah diperiksa dalam kondisi baik dan tersegel. Klaim selisih atau kerusakan wajib dilaporkan maksimal 1x24 jam sejak barang diterima.
                </p>
              </div>
            </div>

            <!-- QR Code Box -->
            <div class="col-span-1 border border-slate-200 print:border-slate-400 rounded-lg p-2 flex flex-col items-center justify-center text-center bg-slate-50/30 print:bg-white">
              <img
                :src="qrCodeUrl"
                alt="QR Code DO"
                class="w-20 h-20 border border-slate-900 p-0.5 rounded bg-white object-contain"
              />
              <span class="text-[9px] font-mono font-bold text-slate-700 mt-1.5">{{ deliveryOrderNo }}</span>
              <span class="text-[8px] text-slate-400 font-bold uppercase">Verifikasi Digital</span>
            </div>
          </div>

          <!-- Tanda Tangan 4 Kolom Operasional Distribusi -->
          <div class="grid grid-cols-4 gap-3 pt-3 border-t-2 border-slate-900 print:border-black text-center text-[10px]">
            <div>
              <p class="font-black text-slate-600 uppercase text-[9px]">Dibuat Oleh (Admin Pabrik)</p>
              <div class="h-14 flex items-end justify-center">
                <span class="font-bold border-b border-slate-900 pb-0.5 px-2 truncate max-w-full">
                  ( {{ transfer.approved_by_name || 'Admin Nota Pabrik' }} )
                </span>
              </div>
            </div>
            <div>
              <p class="font-black text-slate-600 uppercase text-[9px]">Diserahkan (Driver/Ekspedisi)</p>
              <div class="h-14 flex items-end justify-center">
                <span class="font-bold border-b border-slate-900 pb-0.5 px-2 truncate max-w-full">
                  ( {{ driverName !== '-' ? driverName : 'Pengemudi Pabrik' }} )
                </span>
              </div>
            </div>
            <div>
              <p class="font-black text-slate-600 uppercase text-[9px]">Diperiksa (Gudang Pabrik)</p>
              <div class="h-14 flex items-end justify-center">
                <span class="font-bold border-b border-slate-900 pb-0.5 px-2">
                  ( Petugas Gudang )
                </span>
              </div>
            </div>
            <div>
              <p class="font-black text-slate-600 uppercase text-[9px]">Diterima (Gudang Cabang)</p>
              <div class="h-14 flex items-end justify-center">
                <span class="font-bold border-b border-slate-900 pb-0.5 px-2">
                  ( Kepala Gudang Cabang )
                </span>
              </div>
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
    margin: 8mm;
  }

  body, html {
    background: white !important;
    color: black !important;
    overflow: visible !important;
  }

  /* Hide entire application background and UI chrome */
  body * {
    visibility: hidden;
  }

  /* Only reveal the invoice paper and its contents */
  .invoice-paper,
  .invoice-paper * {
    visibility: visible;
  }

  .invoice-paper {
    position: absolute;
    left: 0;
    top: 0;
    width: 100% !important;
    margin: 0 !important;
    padding: 0 !important;
    box-shadow: none !important;
    border: none !important;
    background: white !important;
    color: black !important;
  }

  .print\:hidden {
    display: none !important;
  }

  * {
    -webkit-print-color-adjust: exact !important;
    print-color-adjust: exact !important;
  }
}
</style>
