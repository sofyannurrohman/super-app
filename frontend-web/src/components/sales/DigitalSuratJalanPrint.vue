<template>
  <div class="digital-surat-jalan-print w-full bg-white font-sans text-slate-900 print:shadow-none print:max-w-full print:block hidden">
    <!-- Header -->
    <div class="bg-white p-4 print:p-2 border-b border-slate-100">
      <div class="grid grid-cols-3 gap-8 items-center">
        <div class="space-y-1">
          <p class="text-[9px] font-black text-slate-400 uppercase tracking-widest">
            SOPIR / ARMADA
          </p>
          <p class="text-[11px] font-black text-slate-900">
            {{
              selectedDO.driver
                ? `${selectedDO.driver.first_name} ${selectedDO.driver.last_name}`
                : (selectedDO.daily_driver?.name || '........................')
            }}
            / {{ selectedDO.vehicle?.license_plate || '..........' }}
          </p>
          <p v-if="selectedDO.helper || selectedDO.daily_helper" class="text-[9px] font-bold text-slate-600 mt-0.5">
            Helper: {{
              selectedDO.helper
                ? `${selectedDO.helper.first_name} ${selectedDO.helper.last_name}`
                : (selectedDO.daily_helper?.name || '')
            }}
            <span v-if="selectedDO.daily_helper" class="text-amber-500 ml-1">(Harian)</span>
          </p>
        </div>
        <div class="space-y-1 text-center">
          <div class="inline-block bg-primary/10 px-4 py-1 rounded-full border border-primary/20 print:border-slate-900">
            <p class="text-[10px] font-black text-primary uppercase tracking-tighter print:text-slate-900">
              SURAT JALAN DIGITAL
            </p>
          </div>
          <p class="text-[12px] font-black text-slate-900 mt-1 uppercase tracking-tight">
            {{ selectedDO.delivery_order_no }}
          </p>
        </div>
        <div class="text-right">
          <p class="text-[9px] font-black text-slate-400 uppercase tracking-widest">
            TANGGAL TERBIT
          </p>
          <p class="text-[10px] font-black text-slate-900 uppercase leading-tight mt-1">
            {{ formatDate(selectedDO.created_at) }}
          </p>
        </div>
      </div>
    </div>

    <div class="p-4 sm:p-6 space-y-4 print:p-2">
      <!-- Main Table -->
      <div class="border-2 border-slate-900 rounded-lg print:rounded-none overflow-x-auto bg-white">
        <table class="w-full text-left text-[10px] border-collapse whitespace-nowrap">
          <thead>
            <tr class="bg-slate-50 border-b-2 border-slate-900">
              <th class="p-2 border-r border-slate-300 font-black text-center w-8">NO</th>
              <th class="p-1 border-r border-slate-300 font-black text-center w-12 text-[8px]">QR SCAN</th>
              <th class="p-2 border-r border-slate-300 font-black">NAMA TOKO / ALAMAT</th>
              <th class="p-2 border-r border-slate-300 font-black text-right w-20">TAGIHAN</th>
              <th class="p-2 border-r border-slate-300 font-black text-center w-24">BAYAR</th>
              <th class="p-2 border-r border-slate-300 font-black text-center w-16">KET</th>

              <th
                v-for="prod in batchProductColumns"
                :key="prod.id"
                class="p-1 border-r border-slate-300 font-black text-center text-[7px] w-10 uppercase"
              >
                {{ prod.name }}
              </th>

              <th class="p-2 border-r border-slate-300 font-black text-center w-10 text-[8px]">HADIAH</th>
              <th class="p-2 font-black text-center w-10 text-[8px]">RETUR</th>
            </tr>
          </thead>
          <tbody class="divide-y border-slate-300">
            <tr v-for="(item, idx) in selectedDO.items" :key="item.id" class="h-8">
              <td class="p-2 border-r border-slate-300 text-center font-bold">{{ Number(idx) + 1 }}</td>
              <td class="p-1 border-r border-slate-300 text-center align-middle">
                <img
                  :src="`https://api.qrserver.com/v1/create-qr-code/?size=60x60&margin=0&data=${(item.sales_order || item.sales_transaction)?.so_number || (item.sales_order || item.sales_transaction)?.invoice_no || item.id}`"
                  class="w-8 h-8 mx-auto"
                />
              </td>
              <td class="p-2 border-r border-slate-300 whitespace-normal">
                <div class="flex items-center gap-1">
                  <p class="font-black text-slate-900 leading-none">
                    {{ (item.sales_order || item.sales_transaction)?.store?.name }}
                  </p>
                  <span v-if="item.is_subsequent" class="inline-block px-1 py-0.5 text-[6.5px] font-black uppercase bg-amber-100 text-amber-900 border border-amber-400 rounded">SUSULAN</span>
                </div>
                <p class="text-[7px] text-slate-500 font-bold mt-0.5">
                  {{ (item.sales_order || item.sales_transaction)?.store?.address }}
                </p>
              </td>
              <td class="p-2 border-r border-slate-300 text-right font-black">
                {{ formatCurrency((item.sales_order || item.sales_transaction)?.total_amount || 0).replace('Rp', '').trim() }}
              </td>
              <td class="p-2 border-r border-slate-300"></td>
              <td class="p-2 border-r border-slate-300 text-center font-bold text-[8px]">
                {{
                  (item.sales_order || item.sales_transaction)?.payment_method === 'VA' &&
                  (item.sales_order || item.sales_transaction)?.midtrans_bank
                    ? `VA ${(item.sales_order || item.sales_transaction).midtrans_bank.toUpperCase()}`
                    : (item.sales_order || item.sales_transaction)?.payment_method || 'COD'
                }}
              </td>
              <td
                v-for="prod in batchProductColumns"
                :key="prod.id"
                class="p-1 border-r border-slate-300 text-center font-black"
              >
                {{ getProductQty(item, prod.id) || '' }}
              </td>
              <td class="p-1 border-r border-slate-300 text-[6px] font-bold whitespace-pre-wrap leading-tight text-center align-middle">
                {{ (item.sales_order || item.sales_transaction)?.bonus_items?.map((b: any) => `${b.quantity} ${b.bonus_inventory?.name}`).join('\n') || '' }}
              </td>
              <td class="p-2"></td>
            </tr>

            <!-- ROW UNTUK SPARE / DO -->
            <tr v-if="selectedDO.spares && selectedDO.spares.length > 0" class="h-8">
              <td class="p-2 border-r border-slate-300 text-center font-bold">{{ (selectedDO.items?.length || 0) + 1 }}</td>
              <td class="p-1 border-r border-slate-300 text-center align-middle">-</td>
              <td class="p-2 border-r border-slate-300 whitespace-normal">
                <p class="font-black text-slate-900 leading-none">[SPARE / DO]</p>
                <p class="text-[7px] text-slate-500 font-bold mt-0.5">Barang Tambahan Gudang</p>
              </td>
              <td class="p-2 border-r border-slate-300 text-right font-black">-</td>
              <td class="p-2 border-r border-slate-300">-</td>
              <td class="p-2 border-r border-slate-300 text-center font-bold text-[8px]">-</td>
              <td
                v-for="prod in batchProductColumns"
                :key="'spare-'+prod.id"
                class="p-1 border-r border-slate-300 text-center font-black"
              >
                {{ getSpareQty(prod.id) || '' }}
              </td>
              <td class="p-2 border-r border-slate-300">-</td>
              <td class="p-2">-</td>
            </tr>

            <!-- Placeholder Rows for Manual Entry -->
            <tr
              v-for="n in Math.max(0, 15 - (selectedDO.items?.length || 0))"
              :key="'blank-' + n"
              class="h-6"
            >
              <td class="p-2 border-r border-slate-300 text-center text-slate-300 font-bold">
                {{ (selectedDO.items?.length || 0) + n }}
              </td>
              <td class="p-2 border-r border-slate-300"></td>
              <td class="p-2 border-r border-slate-300"></td>
              <td class="p-2 border-r border-slate-300"></td>
              <td class="p-2 border-r border-slate-300"></td>
              <td class="p-2 border-r border-slate-300"></td>
              <td v-for="prod in batchProductColumns" :key="prod.id" class="p-1 border-r border-slate-300"></td>
              <td class="p-2 border-r border-slate-300"></td>
              <td class="p-2"></td>
            </tr>

            <!-- Total Row -->
            <tr class="font-black border-t-2 border-slate-900 h-8">
              <td colspan="3" class="p-2 text-right border-r border-slate-300 uppercase tracking-widest text-[8px]">
                TOTAL BATCH
              </td>
              <td class="p-2 text-right border-r border-slate-300">
                {{ formatCurrency(calculateBatchTotal).replace('Rp', '').trim() }}
              </td>
              <td colspan="2" class="border-r border-slate-300"></td>
              <td
                v-for="prod in batchProductColumns"
                :key="prod.id"
                class="p-1 border-r border-slate-300 text-center"
              >
                {{ calculateProductTotal(prod.id) }}
              </td>
              <td colspan="2"></td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Signature Section with Approver Name & Job Position -->
      <div class="grid grid-cols-4 gap-4 mt-6 text-center text-[9px]">
        <div class="space-y-6">
          <p class="font-black uppercase tracking-widest text-slate-500">SOPIR / HELPER</p>
          <p class="font-bold text-slate-900 mt-8">
            ( {{ selectedDO.driver?.first_name ? `${selectedDO.driver.first_name} ${selectedDO.driver.last_name || ''}`.trim() : '..................................' }} )
          </p>
        </div>
        <div class="space-y-6">
          <p class="font-black uppercase tracking-widest text-slate-500">ADMIN NOTA / GUDANG</p>
          <div class="mt-8">
            <p class="font-bold text-slate-900">
              ( {{ selectedDO.admin_nota?.first_name ? `${selectedDO.admin_nota.first_name} ${selectedDO.admin_nota.last_name || ''}`.trim() : '..................................' }} )
            </p>
            <p v-if="selectedDO.admin_nota?.job_position?.title" class="text-[7px] text-slate-400">
              {{ selectedDO.admin_nota.job_position.title }}
            </p>
          </div>
        </div>
        <div class="space-y-6">
          <p class="font-black uppercase tracking-widest text-slate-500">SUPERVISOR (SPV)</p>
          <div class="mt-8">
            <p class="font-bold text-slate-900">
              ( {{ selectedDO.supervisor?.first_name ? `${selectedDO.supervisor.first_name} ${selectedDO.supervisor.last_name || ''}`.trim() : '..................................' }} )
            </p>
            <p v-if="selectedDO.supervisor?.job_position?.title" class="text-[7px] text-slate-400">
              {{ selectedDO.supervisor.job_position.title }}
            </p>
          </div>
        </div>
        <div class="flex justify-center items-center">
          <img
            :src="`https://api.qrserver.com/v1/create-qr-code/?size=150x150&data=${selectedDO.delivery_order_no}`"
            class="w-16 h-16 border border-slate-900 p-0.5 rounded bg-white"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  selectedDO: any
}>()

const formatDate = (date: string) => {
  if (!date) return '—'
  try {
    return new Intl.DateTimeFormat('id-ID', {
      day: '2-digit',
      month: 'short',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }).format(new Date(date))
  } catch (e) {
    return date
  }
}

const batchProductColumns = computed(() => {
  if (!props.selectedDO) return []

  const productMap = new Map()
  if (props.selectedDO.items) {
    props.selectedDO.items.forEach((item: any) => {
      const trx = item.sales_order || item.sales_transaction
      if (trx?.items) {
        trx.items.forEach((si: any) => {
          if (si.product) {
            productMap.set(si.product_id, {
              id: si.product_id,
              name: si.product.name
                .split(' ')
                .map((w: string) => w[0])
                .join(''), // Abbreviate
            })
          }
        })
      }
    })
  }

  if (props.selectedDO.spares) {
    props.selectedDO.spares.forEach((spare: any) => {
      if (spare.product) {
        productMap.set(spare.product_id, {
          id: spare.product_id,
          name: spare.product.name
            .split(' ')
            .map((w: string) => w[0])
            .join(''),
        })
      }
    })
  }

  return Array.from(productMap.values())
})

const getProductQty = (item: any, productId: string) => {
  const trx = item.sales_order || item.sales_transaction
  if (!trx || !trx.items) return 0
  const si = trx.items.find((i: any) => i.product_id === productId)
  return si ? si.quantity : 0
}

const getSpareQty = (productId: string) => {
  if (!props.selectedDO || !props.selectedDO.spares) return 0
  const spare = props.selectedDO.spares.find((s: any) => s.product_id === productId)
  return spare ? spare.quantity : 0
}

const calculateBatchTotal = computed(() => {
  if (!props.selectedDO || !props.selectedDO.items) return 0
  return props.selectedDO.items.reduce((sum: number, item: any) => {
    const trx = item.sales_order || item.sales_transaction
    return sum + (trx?.total_amount || 0)
  }, 0)
})

const calculateProductTotal = (productId: string) => {
  if (!props.selectedDO) return 0
  let total = 0
  if (props.selectedDO.items) {
    total += props.selectedDO.items.reduce((sum: number, item: any) => {
      return sum + getProductQty(item, productId)
    }, 0)
  }
  if (props.selectedDO.spares) {
    const spare = props.selectedDO.spares.find((s: any) => s.product_id === productId)
    if (spare) {
      total += spare.quantity || 0
    }
  }
  return total
}

const formatCurrency = (val: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(val)
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
  .p-8,
  .min-h-screen {
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
  .digital-surat-jalan-print {
    display: block !important;
    position: static !important;
    width: 100% !important;
    zoom: 0.75; /* Scale down to ensure it fits */
  }

  /* Hide modals if they are open */
  .fixed.inset-0 {
    display: none !important;
  }
  
  .print-only {
    display: block !important;
  }

  /* Make sure background colors and custom text colors are printed */
  * {
    -webkit-print-color-adjust: exact !important;
    print-color-adjust: exact !important;
  }

  /* Compact Table & Solid Black Colors */
  .digital-surat-jalan-print th,
  .digital-surat-jalan-print td {
    padding-top: 2px !important;
    padding-bottom: 2px !important;
  }

  .digital-surat-jalan-print,
  .digital-surat-jalan-print * {
    color: #000 !important;
    border-color: #000 !important;
  }

  @page {
    size: landscape; /* Landscape digital delivery order */
    margin: 5mm;
  }
}
</style>
