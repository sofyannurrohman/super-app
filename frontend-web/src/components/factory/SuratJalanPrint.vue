<template>
  <div
    v-if="shipment"
    id="surat-jalan-card"
    class="surat-jalan-print w-full bg-white font-sans text-slate-900 print:shadow-none print:max-w-full"
  >
    <!-- Header Area -->
    <div class="bg-white p-4 border-b-2 border-slate-900">
      <div class="grid grid-cols-3 gap-6 items-center">
        <div class="space-y-1">
          <p class="text-[9px] font-black text-slate-400 uppercase tracking-widest">
            SOPIR / ARMADA
          </p>
          <p class="text-[11px] font-black text-slate-900">{{ driverName }} / {{ licensePlate }}</p>
          <p class="text-[10px] font-bold text-slate-500 uppercase tracking-tight">
            {{ vehicleName }}
          </p>
        </div>
        <div class="space-y-1 text-center">
          <p class="text-[9px] font-black text-slate-400 uppercase tracking-widest">Tujuan</p>
          <p class="text-[11px] font-black text-slate-900 leading-snug">
            {{ fromFactoryName }} ke {{ toBranchName }}
          </p>
          <p class="text-[10px] font-bold text-slate-500 italic text-xs leading-normal">
            {{ fromFactoryLocation }}
          </p>
        </div>
        <div class="text-right">
          <p
            class="text-[13px] font-mono font-black text-slate-900 pb-1 italic uppercase tracking-tighter"
          >
            SURAT JALAN #{{ shipment.doNo }}
          </p>
          <p v-if="shipment.transferType === 'INTERNAL_TRANSFER' || shipment.paymentMethod === 'INTERNAL' || shipment.items?.[0]?.transfer_type === 'INTERNAL_TRANSFER' || shipment.items?.[0]?.is_billable === false" class="text-[9px] font-black text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded border border-emerald-200 inline-block uppercase">
            Mutasi Internal (Bebas Tagihan)
          </p>
          <p class="text-[9px] font-bold text-slate-500 mt-1 uppercase">
            {{ formatDateLong(targetDate) }}
          </p>
        </div>
      </div>
    </div>

    <div class="p-4 print:p-2 space-y-4 print:space-y-2">
      <!-- Main Table -->
      <div class="border-2 border-slate-900 rounded-lg print:rounded-none overflow-hidden mb-2 mt-2 bg-white">
        <table class="w-full text-left text-[10px] border-collapse whitespace-nowrap">
          <thead>
            <tr class="bg-slate-50 border-b-2 border-slate-900">
              <th
                class="p-2 border-r border-slate-300 font-black text-center w-12"
              >
                NO
              </th>
              <th class="p-2 border-r border-slate-300 font-black w-32">
                NAMA PRODUK
              </th>
              <th
                class="p-2 border-r border-slate-300 font-black text-center w-32"
              >
                BATCH NO
              </th>
              <th
                class="p-2 border-r border-slate-300 font-black text-right w-20"
              >
                JUMLAH
              </th>
              <th
                class="p-2 border-r border-slate-300 font-black text-center w-16"
              >
                SATUAN
              </th>
              <th
                class="p-2 border-r border-slate-300 font-black text-right w-20"
              >
                BERAT (KG)
              </th>
              <th
                class="p-2 border-r border-slate-300 font-black text-right w-24"
              >
                HARGA (Rp)
              </th>
              <th
                class="p-2 border-r border-slate-300 font-black text-right w-28"
              >
                SUBTOTAL (Rp)
              </th>
              <th class="p-2 font-black text-center w-16">KET</th>
            </tr>
          </thead>
          <tbody class="divide-y border-slate-300">
            <tr
              v-for="(item, idx) in shipment.items"
              :key="item.id"
              class="h-8"
            >
              <td
                class="p-2 text-center font-bold border-r border-slate-300 text-slate-500"
              >
                {{ Number(idx) + 1 }}
              </td>
              <td class="p-2 border-r border-slate-300 whitespace-normal">
                <p class="font-bold text-slate-900 uppercase leading-tight">
                  {{ item.product?.name }}
                </p>
                <p class="text-[8px] text-slate-400 font-bold uppercase tracking-widest mt-0.5">
                  {{ item.product?.sku }}
                </p>
              </td>
              <td
                class="p-2 text-center font-mono font-semibold border-r border-slate-300 text-slate-700 bg-white"
              >
                {{ item.batch_no || 'DEFAULT' }}
              </td>
              <td
                class="p-2 text-right font-black border-r border-slate-300 text-slate-950"
              >
                {{ item.quantity }}
              </td>
              <td
                class="p-2 text-center font-bold border-r border-slate-300 text-slate-500 uppercase"
              >
                {{ item.unit || item.product?.unit }}
              </td>
              <td
                class="p-2 text-right font-black border-r border-slate-300 text-slate-950"
              >
                {{ (item.total_weight || 0).toFixed(2) }}
              </td>
              <td
                class="p-2 text-right font-black border-r border-slate-300 text-slate-950"
              >
                {{ formatRupiah(item.product?.factory_price || 0) }}
              </td>
              <td
                class="p-2 text-right font-black border-r border-slate-300 text-slate-950 bg-amber-50/20"
              >
                {{ formatRupiah(item.total_amount || 0) }}
              </td>
              <td class="p-2 text-slate-700 italic text-[9px] leading-tight text-center">
                {{ item.notes || '-' }}
              </td>
            </tr>

            <!-- Filler Rows to fit at least 10 rows -->
            <tr
              v-for="n in Math.max(0, 10 - (shipment.items?.length || 0))"
              :key="'filler-' + n"
              class="h-8"
            >
              <td
                class="p-2 border-r border-slate-300 text-center text-slate-300 font-bold"
              >
                {{ (shipment.items?.length || 0) + Number(n) }}
              </td>
              <td class="p-2 border-r border-slate-300"></td>
              <td class="p-2 border-r border-slate-300 bg-white"></td>
              <td class="p-2 border-r border-slate-300"></td>
              <td class="p-2 border-r border-slate-300"></td>
              <td class="p-2 border-r border-slate-300 bg-white"></td>
              <td class="p-2 border-r border-slate-300"></td>
              <td class="p-2 border-r border-slate-300 bg-amber-50/20"></td>
              <td class="p-2"></td>
            </tr>

            <!-- Total Row (Pure Black and White, no dark background to save ink) -->
            <tr
              class="bg-white text-slate-950 font-black h-10 total-row border-t-2 border-slate-900"
            >
              <td
                colspan="3"
                class="p-2 text-right border-r border-slate-900 uppercase tracking-widest text-[9px] font-black text-slate-700"
              >
                TOTAL MUATAN PABRIK
              </td>
              <td
                class="p-2 text-right border-r border-slate-900 text-sm font-black"
              >
                {{ calculateTotalQty() }}
              </td>
              <td class="border-r border-slate-900"></td>
              <td
                class="p-2 text-right border-r border-slate-900 text-sm font-black"
              >
                {{ calculateTotalWeight().toFixed(2) }}
              </td>
              <td class="border-r border-slate-900"></td>
              <td
                class="p-2 text-right border-r border-slate-900 text-sm font-black"
              >
                {{ formatRupiah(calculateGrandTotal()) }}
              </td>
              <td></td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Footer Section -->
      <div class="grid grid-cols-3 gap-3 pt-4 print:pt-2">
        <div class="text-center space-y-12 print:space-y-8">
          <p class="text-[9px] font-black text-slate-500 uppercase tracking-widest">
            PENGIRIM (PABRIK)
          </p>
          <div class="space-y-1">
            <p class="text-[10px] font-black text-slate-900">
              ( .................................................. )
            </p>
            <p class="text-[8px] text-slate-500 font-bold uppercase tracking-wider">
              Admin Gudang Pabrik
            </p>
          </div>
        </div>
        <div class="text-center space-y-12 print:space-y-8">
          <p class="text-[9px] font-black text-slate-500 uppercase tracking-widest">
            PEMBAWA (DRIVER)
          </p>
          <div class="space-y-1">
            <p class="text-[10px] font-black text-slate-900">( {{ driverName }} )</p>
            <p class="text-[8px] text-slate-500 font-bold uppercase tracking-wider">
              Sopir / Kurir Pengantar
            </p>
          </div>
        </div>
        <div class="flex justify-end items-center pr-4 gap-4 print:pr-0 print:gap-2">
          <div class="text-right">
            <p class="text-[10px] font-black text-slate-900 uppercase tracking-widest leading-none">
              VALIDASI DIGITAL
            </p>
            <p class="text-[7px] text-slate-500 font-bold mt-1 uppercase tracking-tighter">
              Scan untuk konfirmasi
            </p>
          </div>
          <div
            class="relative bg-white p-2 border border-slate-900 rounded-xl shadow-sm flex items-center justify-center"
          >
            <QrcodeVue :value="shipment.doNo" :size="72" level="H" render-as="svg" />
          </div>
        </div>
      </div>
    </div>

    <!-- Footer Notes -->
    <div class="px-6 py-4 border-t border-dashed border-slate-200 text-center">
      <p class="text-[8px] font-black text-slate-400 uppercase tracking-[0.3em]">
        Dokumen ini sah dicetak dari sistem KARSA HR HRIS pada {{ new Date().toLocaleString('id-ID') }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import QrcodeVue from 'qrcode.vue'

const props = defineProps<{
  shipment: any
}>()

const driverName = computed(() => {
  return (
    props.shipment?.driver?.first_name ||
    props.shipment?.items?.[0]?.driver?.first_name ||
    '........................'
  )
})

const licensePlate = computed(() => {
  return (
    props.shipment?.vehicle?.license_plate ||
    props.shipment?.items?.[0]?.vehicle?.license_plate ||
    '..........'
  )
})

const vehicleName = computed(() => {
  return props.shipment?.vehicle?.name || props.shipment?.items?.[0]?.vehicle?.name || ''
})

const fromFactoryName = computed(() => {
  return props.shipment?.from || props.shipment?.items?.[0]?.from_factory?.name || 'Pabrik'
})

const fromFactoryLocation = computed(() => {
  return props.shipment?.items?.[0]?.from_factory?.location || ''
})

const toBranchName = computed(() => {
  return props.shipment?.to || props.shipment?.items?.[0]?.to_branch?.name || 'Cabang Tujuan'
})

const targetDate = computed(() => {
  return props.shipment?.targetDate || props.shipment?.items?.[0]?.target_shipment_date || ''
})

const formatDateLong = (dateStr: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('id-ID', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  })
}

const calculateTotalWeight = () => {
  if (!props.shipment || !props.shipment.items) return 0
  return props.shipment.items.reduce((acc: number, item: any) => acc + (item.total_weight || 0), 0)
}

const calculateTotalQty = () => {
  if (!props.shipment || !props.shipment.items) return 0
  return props.shipment.items.reduce((acc: number, item: any) => acc + (item.quantity || 0), 0)
}

const calculateGrandTotal = () => {
  if (!props.shipment || !props.shipment.items) return 0
  return props.shipment.items.reduce((acc: number, item: any) => acc + (item.total_amount || 0), 0)
}

const formatRupiah = (value: number) => {
  return new Intl.NumberFormat('id-ID').format(value)
}
</script>

<style>
/* Global print styles to reset page layout and hide UI navigation elements */
@media print {
  /* Hide sidebar, headers, and dashboard layout buttons */
  aside,
  header,
  nav,
  button,
  .no-print,
  .print\:hidden {
    display: none !important;
  }

  /* Reset layout constraints for body, html and wrapper containers */
  html,
  body,
  #app,
  .flex,
  main,
  .overflow-y-auto,
  .p-4,
  .p-8 {
    height: auto !important;
    min-height: 0 !important;
    overflow: visible !important;
    position: static !important;
    margin: 0 !important;
    padding: 0 !important;
    background: white !important;
    box-shadow: none !important;
  }

  /* Ensure the printed element is visible and takes full page width */
  .print-only, .surat-jalan-print {
    display: block !important;
    position: static !important;
    width: 100% !important;
    zoom: 0.75; /* Scale down to fit 1 page safely */
  }

  /* Make sure background colors and custom text colors are printed */
  * {
    -webkit-print-color-adjust: exact !important;
    print-color-adjust: exact !important;
  }

  /* Compact Table & Solid Black Colors */
  .surat-jalan-print th,
  .surat-jalan-print td {
    padding-top: 2px !important;
    padding-bottom: 2px !important;
  }

  .surat-jalan-print,
  .surat-jalan-print * {
    color: #000 !important;
    border-color: #000 !important;
  }

  @page {
    size: landscape; /* Match digital delivery order orientation */
    margin: 5mm;
  }
}
</style>

<style scoped>
.surat-jalan-print {
  max-width: 800px;
  margin: 0 auto;
  border: 1px solid #e2e8f0;
  border-radius: 1rem;
  padding: 1.5rem;
  box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1);
}

@media print {
  .surat-jalan-print {
    border: none !important;
    border-radius: 0 !important;
    padding: 0 !important;
    box-shadow: none !important;
    margin: 0 !important;
    max-width: 100% !important;
  }
}
</style>
