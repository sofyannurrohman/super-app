<template>
  <div class="min-h-screen bg-slate-100 py-8 px-4 print:bg-white print:py-0 print:px-0">
    <!-- Action Bar (hidden in print) -->
    <div class="max-w-3xl mx-auto mb-4 flex justify-between items-center print:hidden">
      <div class="flex items-center space-x-2 text-slate-600">
        <svg class="h-5 w-5 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
          />
        </svg>
        <span class="font-semibold text-sm">Faktur Penjualan</span>
      </div>
      <button
        @click="printInvoice"
        class="bg-blue-600 hover:bg-blue-700 text-white px-5 py-2 rounded-lg font-semibold text-sm flex items-center space-x-2 transition"
      >
        <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z"
          />
        </svg>
        <span>Cetak / Simpan PDF</span>
      </button>
    </div>

    <!-- Invoice Document (A4-like) -->
    <div
      class="max-w-3xl mx-auto bg-white shadow-lg print:shadow-none print:max-w-none"
      style="font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif"
    >
      <div class="p-10 print:p-2" v-if="invoice">
        <!-- ===== HEADER ===== -->
        <div class="flex justify-between items-start pb-4 print:pb-2 border-b-2 border-black">
          <div>
            <div class="text-base font-black text-black uppercase tracking-wide">
              {{
                invoice.employee?.company?.name ||
                invoice.company?.name ||
                'Perusahaan Demo'
              }}
            </div>
            <div class="text-xs text-gray-600 mt-1 max-w-xs leading-relaxed">
              {{ invoice.company?.address || 'Kawasan Industri & Distribusi Sentra Prima' }}
            </div>
          </div>
          <div class="text-right">
            <div class="text-lg font-black text-black uppercase tracking-wider">
              FAKTUR PENJUALAN
            </div>
            <div class="text-sm font-bold text-black mt-0.5">{{ invoice.receipt_no }}</div>
            <div class="text-xs text-gray-500 mt-0.5">
              Tanggal: {{ formatDate(invoice.transaction_date) }}
            </div>
          </div>
        </div>

        <!-- ===== BILL TO / PAYMENT INFO ===== -->
        <div class="flex justify-between items-start mt-5 mb-5 print:mt-2 print:mb-2">
          <div>
            <div class="text-xs font-bold text-slate-900 uppercase tracking-wider mb-1">
              DITAGIHKAN KEPADA:
            </div>
            <div class="text-sm font-bold text-black">{{ getCustomerName() }}</div>
            <div class="text-xs text-gray-600 mt-0.5 max-w-xs leading-relaxed" v-if="getCustomerAddress()">
              {{ getCustomerAddress() }}
            </div>
            <div class="text-xs font-bold text-slate-800 mt-2" v-if="getStaffInfo()">
              {{ getStaffInfo() }}
            </div>
          </div>
          <div class="text-right">
            <div class="text-xs font-bold text-slate-900 uppercase tracking-wider mb-1">
              METODE PEMBAYARAN:
            </div>
            <div class="text-sm font-bold text-black uppercase">{{ invoice.payment_method }}</div>
            <div
              class="text-xs font-semibold mt-0.5"
              v-if="invoice.payment_method === 'TEMPO' && invoice.payment_due_date"
            >
              <span class="text-slate-900"
                >Jatuh Tempo: {{ formatDate(invoice.payment_due_date) }}</span
              >
            </div>
            <div class="text-xs font-semibold mt-0.5">
              <span :class="getPaymentStatusBadgeClass(invoice.payment_status)">
                Status: {{ formatPaymentStatus(invoice.payment_status, invoice.payment_method) }}
              </span>
            </div>
          </div>
        </div>

        <!-- ===== PRODUCT TABLE ===== -->
        <table class="w-full text-left border-collapse mt-2">
          <thead>
            <tr class="bg-gray-200">
              <th
                class="py-2 px-2 print:py-1 print:px-1.5 text-xs print:text-[10px] font-bold text-black text-left w-1/2"
              >
                Nama Produk
              </th>
              <th
                class="py-2 px-2 print:py-1 print:px-1.5 text-xs print:text-[10px] font-bold text-black text-center"
              >
                Qty
              </th>
              <th
                class="py-2 px-2 print:py-1 print:px-1.5 text-xs print:text-[10px] font-bold text-black text-right"
              >
                Harga
              </th>
              <th
                class="py-2 px-2 print:py-1 print:px-1.5 text-xs print:text-[10px] font-bold text-black text-right"
              >
                Subtotal
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="item in invoice.items"
              :key="item.id"
              class="border-b border-gray-200 print:text-[10px]"
            >
              <td class="py-3 px-2 print:py-1.5 print:px-1 text-xs print:text-[10px] text-black">
                {{ item.product?.name || item.product_name || '-' }}
              </td>
              <td
                class="py-3 px-2 print:py-1.5 print:px-1 text-xs print:text-[10px] text-black text-center"
              >
                {{ formatItemQty(item) }}
              </td>
              <td
                class="py-3 px-2 print:py-1.5 print:px-1 text-xs print:text-[10px] text-black text-right"
              >
                {{
                  item.price_at_transaction > 0
                    ? 'Rp ' + formatCurrency(item.price_at_transaction)
                    : '-'
                }}
              </td>
              <td
                class="py-3 px-2 print:py-1.5 print:px-1 text-xs print:text-[10px] font-semibold text-black text-right"
              >
                {{ item.subtotal > 0 ? 'Rp ' + formatCurrency(item.subtotal) : '-' }}
              </td>
            </tr>
          </tbody>
        </table>

        <!-- ===== TOTALS ===== -->
        <div class="flex justify-end mt-4 print:mt-2">
          <div class="space-y-1 text-right min-w-48">
            <div class="flex justify-between space-x-8 text-xs text-gray-600">
              <span>Subtotal:</span>
              <span class="font-semibold text-black"
                >Rp {{ formatCurrency(invoice.total_amount) }}</span
              >
            </div>
            <div
              class="flex justify-between space-x-8 text-sm font-bold text-black border-t border-gray-300 pt-1 mt-1"
            >
              <span>Total Bayar:</span>
              <span>Rp {{ formatCurrency(invoice.total_amount) }}</span>
            </div>
          </div>
        </div>

        <!-- ===== SIGNATURES ===== -->
        <div class="flex justify-around mt-10 print:mt-4 pt-6 print:pt-3 border-t border-gray-200">
          <!-- Customer / Penerima -->
          <div class="flex flex-col items-center" style="min-width: 160px">
            <div class="text-xs font-bold text-black mb-2">Penerima / Pelanggan</div>
            <div class="w-36 h-20 print:h-12 flex items-center justify-center">
              <img
                v-if="getCustomerSignature() && !signatureErrors['customer']"
                :src="getFullImageUrl(getCustomerSignature())"
                alt="Tanda Tangan Pelanggan"
                class="max-w-full max-h-full print:max-h-8 object-contain"
                @error="onImgError($event, 'customer')"
              />
              <div
                v-else
                class="w-36 h-16 print:h-8 border-b border-gray-400 flex items-center justify-center"
              >
                <span class="text-gray-400 text-xs italic">Tanda Tangan</span>
              </div>
            </div>
            <div class="text-xs font-semibold text-slate-900 mt-2 text-center max-w-[160px] truncate" :title="getCustomerName()">
              {{ getCustomerName() }}
            </div>
          </div>

          <!-- Courier / Salesman -->
          <div class="flex flex-col items-center" style="min-width: 160px">
            <div class="text-xs font-bold text-black mb-2">{{ getCourierTitle() }}</div>
            <div class="w-36 h-20 print:h-12 flex items-center justify-center">
              <img
                v-if="getCourierSignature() && !signatureErrors['courier']"
                :src="getFullImageUrl(getCourierSignature())"
                alt="Tanda Tangan"
                class="max-w-full max-h-full print:max-h-8 object-contain"
                @error="onImgError($event, 'courier')"
              />
              <div
                v-else
                class="w-36 h-16 print:h-8 border-b border-gray-400 flex items-center justify-center"
              >
                <span class="text-gray-400 text-xs italic">Tanda Tangan</span>
              </div>
            </div>
            <div class="text-xs font-semibold text-slate-900 mt-2 text-center max-w-[160px] truncate" :title="getCourierName()">
              {{ getCourierName() }}
            </div>
          </div>

          <!-- SPV -->
          <div class="flex flex-col items-center" style="min-width: 160px">
            <div class="text-xs font-bold text-black mb-2">SPV</div>
            <div class="w-36 h-20 print:h-12 flex items-center justify-center">
              <div
                class="w-36 h-16 print:h-8 border-b border-gray-400 flex items-center justify-center"
              >
                <span class="text-gray-400 text-xs italic">Tanda Tangan</span>
              </div>
            </div>
            <div class="text-xs font-semibold text-slate-900 mt-2 text-center max-w-[160px] truncate">
              ..................
            </div>
          </div>
        </div>

        <!-- Footer note -->
        <div class="text-center text-xs print:text-[10px] text-gray-400 italic mt-6 print:mt-3">
          Terima kasih atas kerja sama dan kepercayaan Anda.
        </div>
      </div>

      <!-- Loading State -->
      <div class="p-16 text-center text-slate-400 text-sm" v-else-if="loading">
        <div
          class="inline-block animate-spin rounded-full h-8 w-8 border-4 border-slate-300 border-t-slate-800 mb-4"
        ></div>
        <div>Memuat data faktur...</div>
      </div>

      <!-- Error State -->
      <div class="p-16 text-center text-slate-500 text-sm" v-else>
        <svg
          class="mx-auto h-12 w-12 text-slate-300 mb-4"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
          />
        </svg>
        <div class="text-base font-bold text-slate-700">Faktur Tidak Ditemukan</div>
        <p class="text-xs text-slate-400 mt-1">
          Nomor faktur tidak valid atau server mengalami kendala.
        </p>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import axios from 'axios'

export default defineComponent({
  name: 'PublicInvoiceView',
  setup() {
    const route = useRoute()
    const invoice = ref<any>(null)
    const loading = ref(true)
    const signatureErrors = ref<Record<string, boolean>>({})

    const fetchInvoiceDetail = async () => {
      try {
        const rawReceiptNo = route.params.receiptNo
        const receiptNo = Array.isArray(rawReceiptNo) ? rawReceiptNo.join('/') : (rawReceiptNo || '')
        let apiBase =
          import.meta.env.VITE_API_URL ||
          import.meta.env.VITE_API_BASE_URL ||
          window.location.origin + '/api/v1'
        if (!apiBase.includes('/v1')) {
          apiBase = apiBase.endsWith('/') ? `${apiBase}v1` : `${apiBase}/v1`
        }
        const cleanBase = apiBase.endsWith('/') ? apiBase.slice(0, -1) : apiBase
        const response = await axios.get(`${cleanBase}/sales/invoice/detail/${encodeURIComponent(receiptNo)}`, {
          params: { receipt_no: receiptNo },
        })
        if (response.data && response.data.data) {
          invoice.value = response.data.data
          console.log('Invoice signature_image_url:', invoice.value.signature_image_url)
        }
      } catch (err) {
        console.error('Failed to load public invoice:', err)
      } finally {
        loading.value = false
      }
    }

    const isDirectSale = () => {
      const ch = invoice.value?.sales_channel
      return (
        ch === 'DIRECT_FACTORY' ||
        ch === 'DIRECT_BRANCH' ||
        ch === 'EMPLOYEE_SALE' ||
        !!invoice.value?.cashier_employee
      )
    }

    const getCustomerName = () => {
      if (!invoice.value) return '-'
      if (invoice.value.buyer_type === 'EMPLOYEE' && invoice.value.buyer_employee) {
        const fn = invoice.value.buyer_employee.first_name || ''
        const ln = invoice.value.buyer_employee.last_name || ''
        return `${fn} ${ln} (Karyawan)`.trim()
      }
      if (invoice.value.buyer_name) return invoice.value.buyer_name
      if (invoice.value.store?.name) return invoice.value.store.name
      return 'Pelanggan Umum'
    }

    const getCustomerAddress = () => {
      if (!invoice.value) return ''
      if (invoice.value.store?.address) return invoice.value.store.address
      if (invoice.value.buyer_phone) return `Telp: ${invoice.value.buyer_phone}`
      if (invoice.value.branch?.name) return `Cabang: ${invoice.value.branch.name}`
      return ''
    }

    const getStaffInfo = () => {
      if (isDirectSale()) {
        const cashier = invoice.value?.cashier_employee || invoice.value?.employee
        if (cashier) {
          const name = cashier.name || `${cashier.first_name || ''} ${cashier.last_name || ''}`.trim()
          return `Kasir: ${name}`
        }
        return ''
      }
      if (invoice.value?.employee) {
        const emp = invoice.value.employee
        const name = emp.name || `${emp.first_name || ''} ${emp.last_name || ''}`.trim()
        return `Sales: ${name}`
      }
      return ''
    }

    const formatCurrency = (val: number) => {
      if (!val && val !== 0) return '0'
      return new Intl.NumberFormat('id-ID').format(val)
    }

    const formatDate = (dateStr: string) => {
      if (!dateStr) return '-'
      return new Date(dateStr).toLocaleDateString('id-ID', {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
      })
    }

    const getPaymentStatusBadgeClass = (status: string) => {
      return 'text-slate-900 font-bold'
    }

    const formatPaymentStatus = (status: string, paymentMethod?: string) => {
      if (paymentMethod?.toUpperCase() === 'CASH') return 'LUNAS'
      if (!status) return 'BELUM LUNAS'
      const s = status.toUpperCase()
      if (s === 'PAID') return 'LUNAS'
      if (s === 'PARTIAL') return 'BAYAR SEBAGIAN'
      if (s === 'TEMPO') return 'BELUM LUNAS (TEMPO)'
      if (s === 'UNPAID') return 'BELUM LUNAS'
      return status
    }

    const printInvoice = () => window.print()

    const formatItemQty = (item: any) => {
      const qty = item.quantity || 0
      const unit = item.product?.unit || item.unit || 'PCS'
      const pcsPerUnit = item.product?.pcs_per_unit || 1
      if (pcsPerUnit > 1) {
        const bulk = Math.floor(qty / pcsPerUnit)
        const loose = qty % pcsPerUnit
        if (bulk > 0 && loose > 0) return `${bulk} ${unit} + ${loose} PCS`
        if (bulk > 0) return `${bulk} ${unit}`
        return `${loose} PCS`
      }
      return `${qty} ${unit}`
    }

    /**
     * Safely splits the signature_image_url string which may contain
     * two data: URIs joined without a clean separator.
     * Strategy: split on 'data:image/' boundary if present, otherwise split on comma.
     */
    const splitSignatures = (): [string, string] => {
      const raw = invoice.value?.signature_image_url || ''
      if (!raw) return ['', '']

      if (raw.includes('data:image/')) {
        const parts = raw.split('data:image/').filter((p: string) => p.trim())
        const courier = parts[0] ? 'data:image/' + parts[0].replace(/,$/, '').trimEnd() : ''
        const customer = parts[1] ? 'data:image/' + parts[1].replace(/,$/, '').trimEnd() : ''
        return [courier, customer]
      }

      // Plain URL(s) separated by comma
      const parts = raw.split(',').map((s: string) => s.trim())
      if (parts.length === 1 && parts[0]) {
        // If there's only one signature, it is the customer's signature (e.g. driver delivery / POD)
        return ['', parts[0]]
      }
      return [parts[0] || '', parts[1] || '']
    }

    const getCourierSignature = (): string => splitSignatures()[0]
    const getCustomerSignature = (): string => splitSignatures()[1]

    const getFullImageUrl = (path: string): string => {
      if (!path) return ''
      if (path.startsWith('http') || path.startsWith('data:')) return path

      const apiUrl =
        import.meta.env.VITE_API_URL || import.meta.env.VITE_API_BASE_URL || window.location.origin
      try {
        const origin = new URL(apiUrl).origin
        return origin + (path.startsWith('/') ? path : '/' + path)
      } catch {
        const host = apiUrl.replace(/\/api\/v1\/?$/, '').replace(/\/api\/?$/, '')
        return host + (path.startsWith('/') ? path : '/' + path)
      }
    }

    const getCourierTitle = () => {
      if (invoice.value?.delivery_items && invoice.value.delivery_items.length > 0) {
        return 'Driver'
      }
      if (isDirectSale()) {
        return 'Kasir'
      }
      return 'Salesman / Motoris'
    }

    const getCourierName = () => {
      if (invoice.value?.delivery_items && invoice.value.delivery_items.length > 0) {
        const driver = invoice.value.delivery_items[0].delivery_batch?.driver
        if (driver) {
          return `${driver.first_name || ''} ${driver.last_name || ''}`.trim()
        }
        return 'Driver'
      }
      if (isDirectSale()) {
        const cashier = invoice.value?.cashier_employee || invoice.value?.employee
        if (cashier) {
          if (cashier.name) return cashier.name
          return `${cashier.first_name || ''} ${cashier.last_name || ''}`.trim()
        }
        return 'Kasir'
      }
      if (invoice.value?.employee) {
        const emp = invoice.value.employee
        if (emp.name) return emp.name
        return `${emp.first_name || ''} ${emp.last_name || ''}`.trim()
      }
      return 'Sales Motoris'
    }

    const onImgError = (event: Event, type: string) => {
      console.warn(
        `Signature image failed to load (${type}):`,
        (event.target as HTMLImageElement)?.src,
      )
      signatureErrors.value[type] = true
    }

    onMounted(() => fetchInvoiceDetail())

    return {
      invoice,
      loading,
      formatCurrency,
      formatDate,
      getPaymentStatusBadgeClass,
      formatPaymentStatus,
      printInvoice,
      formatItemQty,
      getFullImageUrl,
      getCourierSignature,
      getCustomerSignature,
      getCourierTitle,
      getCourierName,
      getCustomerName,
      getCustomerAddress,
      getStaffInfo,
      onImgError,
      signatureErrors,
    }
  },
})
</script>

<style scoped>
@media print {
  @page {
    margin: 1.5cm;
  }
  body {
    background: white !important;
  }
  .print\:hidden {
    display: none !important;
  }
  .print\:shadow-none {
    box-shadow: none !important;
  }
  .print\:p-8 {
    padding: 2rem !important;
  }
  .print\:max-w-none {
    max-width: none !important;
  }
}
</style>
