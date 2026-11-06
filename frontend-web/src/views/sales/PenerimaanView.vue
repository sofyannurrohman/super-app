<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import apiClient from '@/api/axios'
import { useMasterDataStore } from '@/stores/masterData'

const masterDataStore = useMasterDataStore()
import {
  Search,
  Filter,
  CheckCircle2,
  AlertCircle,
  Eye,
  Calendar,
  Plus,
  Pencil,
  Trash2,
  Save,
  X,
  Store,
  User,
  Clock,
  Truck,
  Printer,
  ClipboardList,
  RefreshCw,
  FileText,
  BadgeAlert,
  ArrowRight,
  Info,
  Image as ImageIcon,
  Wallet,
  Banknote,
  CreditCard,
  CheckCheck,
} from 'lucide-vue-next'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card/index'
import { Button } from '@/components/ui/button'
import ConfirmationDialog from '@/components/ui/ConfirmationDialog.vue'
import RefreshDataButton from '@/components/ui/RefreshDataButton.vue'
import { useDataRefresh } from '@/composables/useDataRefresh'
import { toast } from 'vue-sonner'

// Scope Selector: Pengiriman (Delivery) vs Sales (Progress Tempo)
const activeScope = ref<'DELIVERY' | 'SALES_TEMPO'>('DELIVERY')

// Core states (Delivery)
const transactions = ref<any[]>([])
const salesmen = ref<any[]>([])
const isLoading = ref(true)
const activeTab = ref<'APPROVED' | 'VERIFIED'>('APPROVED')

// Filter states (Delivery)
const searchQuery = ref('')
const selectedSalesman = ref('')
const selectedDate = ref('')
const currentPage = ref(1)
const itemsPerPage = ref(10)
const totalItems = ref(0)
const totalAmountFiltered = ref(0)

// Core states (Sales Tempo)
const tempoPayments = ref<any[]>([])
const isTempoLoading = ref(false)
const tempoActiveTab = ref<'PENDING' | 'VERIFIED'>('PENDING')
const tempoSearchQuery = ref('')
const tempoSelectedSalesman = ref('')
const tempoSelectedDate = ref('')
const tempoCurrentPage = ref(1)
const tempoItemsPerPage = ref(10)
const tempoTotalItems = ref(0)
const tempoTotalAmount = ref(0)

// Modal states (Sales Tempo Verify)
const showVerifyTempoModal = ref(false)
const selectedTempoPayment = ref<any>(null)
const tempoVerificationNotes = ref('')
const isSubmittingTempoVerify = ref(false)

// Modal states (Delivery)
const showVerifyModal = ref(false)
const showDetailsModal = ref(false)
const selectedTransaction = ref<any>(null)
const verificationGeneralNotes = ref('')
const verifyItems = ref<any[]>([])
const isSubmitting = ref(false)

const parsedTransactionNotes = computed(() => {
  if (!selectedTransaction.value?.notes) return null
  try {
    return JSON.parse(selectedTransaction.value.notes)
  } catch (e) {
    return { notes: selectedTransaction.value.notes }
  }
})

// Redeliver modal state
const showRedeliverModal = ref(false)
const redeliverTransaction = ref<any>(null)

// Image preview states
const showImageModal = ref(false)
const previewImageUrl = ref('')

const openImagePreview = (url: string) => {
  previewImageUrl.value = url
  showImageModal.value = true
}

const getFullImageUrl = (path: string): string => {
  if (!path) return ''
  if (path.startsWith('http') || path.startsWith('data:')) return path
  const apiUrl =
    import.meta.env.VITE_API_URL ||
    import.meta.env.VITE_API_BASE_URL ||
    window.location.origin
  try {
    const origin = new URL(apiUrl).origin
    return origin + (path.startsWith('/') ? path : '/' + path)
  } catch {
    const host = apiUrl.replace(/\/api\/v1\/?$/, '').replace(/\/api\/?$/, '')
    return host + (path.startsWith('/') ? path : '/' + path)
  }
}

const deliveryProofImages = computed(() => {
  if (!selectedTransaction.value) return []
  const urls: string[] = []
  
  if (selectedTransaction.value.receipt_image_url) {
    urls.push(getFullImageUrl(selectedTransaction.value.receipt_image_url))
  }
  
  if (selectedTransaction.value.delivery_items) {
    for (const item of selectedTransaction.value.delivery_items) {
      if (item.pod_image_url) {
        const fullUrl = getFullImageUrl(item.pod_image_url)
        if (!urls.includes(fullUrl)) {
          urls.push(fullUrl)
        }
      }
    }
  }
  
  return urls
})

// Fetch all transactions with status and paging
const fetchTransactions = async () => {
  isLoading.value = true
  try {
    const params: any = {
      page: currentPage.value,
      limit: itemsPerPage.value,
      status: activeTab.value,
    }

    if (searchQuery.value) {
      params.search = searchQuery.value
    }
    if (selectedSalesman.value) {
      params.salesman = selectedSalesman.value
    }
    if (selectedDate.value) {
      params.date = selectedDate.value
    }
    if (masterDataStore.selectedBranchId && masterDataStore.selectedBranchId !== 'ALL_BRANCHES') {
      params.branch_id = masterDataStore.selectedBranchId
    }

    const response = await apiClient.get('/admin/sales/transactions/all', { params })
    if (response.data) {
      transactions.value = response.data.data || []
      totalItems.value = response.data.total || 0
      totalAmountFiltered.value = response.data.total_amount || 0
    }
  } catch (error: any) {
    console.error('Gagal mengambil data transaksi:', error)
    toast.error('Gagal mengambil data penerimaan')
  } finally {
    isLoading.value = false
  }
}

// Fetch list of salesmen for filtering
const fetchSalesmen = async () => {
  try {
    const res = await apiClient.get('/employees?limit=all')
    if (res.data?.data) {
      salesmen.value = res.data.data.filter((e: any) => 
        e.job_position?.title?.toLowerCase().includes('sales') || 
        e.job_position?.title?.toLowerCase().includes('motoris') ||
        e.job_position?.title?.toLowerCase().includes('booster')
      )
    }
  } catch (error) {
    console.error('Gagal mengambil data salesman:', error)
  }
}

// Fetch Tempo Payments for Tab Sales
const fetchTempoPayments = async () => {
  isTempoLoading.value = true
  try {
    const params: any = {
      page: tempoCurrentPage.value,
      limit: tempoItemsPerPage.value,
      status: tempoActiveTab.value,
    }
    if (tempoSearchQuery.value) params.search = tempoSearchQuery.value
    if (tempoSelectedSalesman.value) params.salesman = tempoSelectedSalesman.value
    if (tempoSelectedDate.value) params.date = tempoSelectedDate.value
    if (masterDataStore.selectedBranchId && masterDataStore.selectedBranchId !== 'ALL_BRANCHES') {
      params.branch_id = masterDataStore.selectedBranchId
    }

    const res = await apiClient.get('/admin/sales/penerimaan/tempo-payments', { params })
    if (res.data) {
      tempoPayments.value = res.data.data || []
      tempoTotalItems.value = res.data.total || 0
      tempoTotalAmount.value = res.data.total_amount || 0
    }
  } catch (error: any) {
    console.error('Gagal mengambil data pembayaran tempo:', error)
    const status = error.response?.status
    toast.error(status ? `Gagal mengambil data pembayaran tempo (HTTP ${status})` : 'Gagal mengambil data pembayaran tempo')
  } finally {
    isTempoLoading.value = false
  }
}

const openVerifyTempoModal = (payment: any) => {
  selectedTempoPayment.value = payment
  tempoVerificationNotes.value = ''
  showVerifyTempoModal.value = true
}

const confirmVerifyTempoPayment = async () => {
  if (!selectedTempoPayment.value) return
  isSubmittingTempoVerify.value = true
  try {
    await apiClient.post(`/admin/sales/penerimaan/tempo-payments/${selectedTempoPayment.value.id}/verify`, {
      notes: tempoVerificationNotes.value,
    })
    toast.success('Penerimaan pembayaran tempo berhasil diverifikasi! Omzet sales telah disahkan & diperbarui.')
    showVerifyTempoModal.value = false
    selectedTempoPayment.value = null
    await fetchTempoPayments()
  } catch (err: any) {
    console.error('Failed to verify tempo payment:', err)
    toast.error(err.response?.data?.error || 'Gagal memverifikasi pembayaran tempo')
  } finally {
    isSubmittingTempoVerify.value = false
  }
}

const applyTempoFilters = () => {
  tempoCurrentPage.value = 1
  fetchTempoPayments()
}

const resetTempoFilters = () => {
  tempoSearchQuery.value = ''
  tempoSelectedSalesman.value = ''
  tempoSelectedDate.value = ''
  tempoCurrentPage.value = 1
  fetchTempoPayments()
}

const {
  isRefreshing,
  lastRefreshedTimeStr,
  markRefreshed,
  refresh: triggerRefresh,
} = useDataRefresh(
  async ({ isManual }) => {
    const tasks: Promise<unknown>[] = [
      fetchTransactions(),
      fetchTempoPayments(),
      fetchSalesmen(),
    ]
    if (isManual) {
      tasks.push(masterDataStore.fetchBranches(true))
    }
    await Promise.allSettled(tasks)

    // Sinkronkan detail transaksi yang sedang dilihat
    if (selectedTransaction.value?.id && showDetailsModal.value) {
      const updated = transactions.value.find((t: any) => t.id === selectedTransaction.value.id)
      if (updated) {
        selectedTransaction.value = updated
      }
    }
  },
  {
    successMessage: 'Data Penerimaan berhasil diperbarui',
    canAutoRefresh: () =>
      !showVerifyModal.value &&
      !showVerifyTempoModal.value &&
      !showRedeliverModal.value &&
      !isSubmitting.value &&
      !isSubmittingTempoVerify.value &&
      !isLoading.value &&
      !isTempoLoading.value,
  },
)

// Trigger initial data load
onMounted(() => {
  Promise.allSettled([
    fetchTransactions(),
    fetchTempoPayments(),
    fetchSalesmen(),
  ]).then(() => markRefreshed())
})

// Approval dilakukan dari layar lain, jadi refresh daftar penerimaan berkala
// agar faktur baru tidak menunggu user menekan tombol Refresh.
const realtimeRefreshTimer = window.setInterval(() => {
  if (
    !showVerifyModal.value &&
    !showVerifyTempoModal.value &&
    !showRedeliverModal.value &&
    !isSubmitting.value &&
    !isSubmittingTempoVerify.value
  ) {
    if (activeScope.value === 'DELIVERY' && !isLoading.value) {
      fetchTransactions().then(() => markRefreshed())
    } else if (activeScope.value === 'SALES_TEMPO' && !isTempoLoading.value) {
      fetchTempoPayments().then(() => markRefreshed())
    }
  }
}, 30000)

onUnmounted(() => window.clearInterval(realtimeRefreshTimer))

// Watchers for automatic filtering
watch([activeTab, currentPage, itemsPerPage], () => {
  fetchTransactions().then(() => markRefreshed())
})

watch([tempoActiveTab, tempoCurrentPage, tempoItemsPerPage], () => {
  if (activeScope.value === 'SALES_TEMPO') {
    fetchTempoPayments().then(() => markRefreshed())
  }
})

watch(activeScope, (newScope) => {
  if (newScope === 'SALES_TEMPO' && tempoPayments.value.length === 0) {
    fetchTempoPayments().then(() => markRefreshed())
  } else if (newScope === 'DELIVERY' && transactions.value.length === 0) {
    fetchTransactions().then(() => markRefreshed())
  }
})

watch(
  () => masterDataStore.selectedBranchId,
  () => {
    currentPage.value = 1
    tempoCurrentPage.value = 1
    selectedSalesman.value = ''
    tempoSelectedSalesman.value = ''
    Promise.allSettled([
      fetchTransactions(),
      fetchTempoPayments(),
    ]).then(() => markRefreshed())
  },
)

watch(
  () => masterDataStore.selectedSalesType,
  () => {
    currentPage.value = 1
    tempoCurrentPage.value = 1
    Promise.allSettled([
      fetchTransactions(),
      fetchTempoPayments(),
    ]).then(() => markRefreshed())
  },
)

const filteredSalesmen = computed(() => {
  return salesmen.value.filter((emp: any) => {
    const branchId = masterDataStore.selectedBranchId
    const empBranchId = emp.operating_branch_id || emp.branch_id
    return !branchId || branchId === 'ALL_BRANCHES' || empBranchId === branchId
  })
})

const applyFilters = () => {
  currentPage.value = 1
  fetchTransactions()
}

const resetFilters = () => {
  searchQuery.value = ''
  selectedSalesman.value = ''
  selectedDate.value = ''
  currentPage.value = 1
  fetchTransactions()
}

// Format numbers as Rupiah
const formatPrice = (val: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0
  }).format(val)
}

// Format date to local Indonesian format
const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'long',
    year: 'numeric'
  })
}

// Helper to check if a transaction went through delivery
const getDeliveryOrderNo = (trx: any) => {
  if (trx?.delivery_items && trx.delivery_items.length > 0) {
    const batch = trx.delivery_items[0].delivery_batch
    return batch ? batch.delivery_order_no || 'Ada Pengiriman (No SJ Kosong)' : '-'
  }
  return '-'
}

const getDriverName = (trx: any) => {
  if (trx?.delivery_items && trx.delivery_items.length > 0) {
    const batch = trx.delivery_items[0].delivery_batch
    if (batch && batch.driver) {
      return `${batch.driver.first_name} ${batch.driver.last_name || ''}`
    }
  }
  return '-'
}

// Helper to check if a transaction has a failed delivery (Toko Tutup)
const hasFailedDelivery = (trx: any) => {
  if (!trx?.delivery_items || trx.delivery_items.length === 0) return false
  return trx.delivery_items.some((item: any) => item.status === 'FAILED')
}

// Computed to count the total number of transactions with failed deliveries in the current list
const failedDeliveriesCount = computed(() => {
  return transactions.value.filter(trx => hasFailedDelivery(trx) && trx.status === 'PENDING').length
})

// Open Verification Modal
const openVerifyModal = (trx: any) => {
  selectedTransaction.value = trx
  verificationGeneralNotes.value = ''
  
  // Initialize items for editing/adjusting
  verifyItems.value = (trx.items || []).map((it: any) => ({
    product_id: it.product_id,
    name: it.product ? it.product.name : 'Produk Tidak Dikenal',
    sku: it.product ? it.product.sku : '',
    price: it.price_at_transaction,
    original_qty: it.quantity,
    quantity: it.quantity, // starts with current quantity
    notes: '' // notes for this specific item adjustment
  }))
  
  showVerifyModal.value = true
}

// Open Details Modal
const openDetailsModal = (trx: any) => {
  selectedTransaction.value = trx
  showDetailsModal.value = true
}

// Submit Verification
const submitVerification = async () => {
  if (!selectedTransaction.value) return

  // Validate quantities
  for (const item of verifyItems.value) {
    if (item.quantity < 0) {
      toast.error(`Jumlah untuk ${item.name} tidak boleh negatif`)
      return
    }
  }

  isSubmitting.value = true
  try {
    const payload = {
      items: verifyItems.value.map(it => ({
        product_id: it.product_id,
        quantity: Number(it.quantity),
        notes: it.notes
      })),
      notes: verificationGeneralNotes.value
    }

    await apiClient.post(`/admin/sales/transactions/${selectedTransaction.value.id}/verify-penerimaan`, payload)
    
    toast.success('Penerimaan berhasil diverifikasi dan disetujui!')
    showVerifyModal.value = false
    fetchTransactions()
  } catch (error: any) {
    console.error('Gagal verifikasi:', error)
    toast.error(error.response?.data?.error || 'Gagal melakukan verifikasi')
  } finally {
    isSubmitting.value = false
  }
}

const openRedeliverModal = (trx: any) => {
  redeliverTransaction.value = trx
  showRedeliverModal.value = true
}

const confirmRedeliver = async () => {
  if (!redeliverTransaction.value) return
  try {
    await apiClient.post(`/admin/sales/transactions/${redeliverTransaction.value.id}/redeliver`)
    toast.success('Pesanan berhasil dikembalikan ke tahap Surat Jalan!')
    showRedeliverModal.value = false
    fetchTransactions()
  } catch (error: any) {
    console.error('Gagal memproses pengiriman ulang:', error)
    toast.error(error.response?.data?.error || 'Gagal mengembalikan pesanan ke surat jalan')
    // keep modal open for retry or cancel
  }
}

// Calculated totals for verification modal
const adjustedTotalAmount = computed(() => {
  return verifyItems.value.reduce((sum, item) => sum + (item.quantity * item.price), 0)
})

const totalOriginalQty = computed(() => {
  return verifyItems.value.reduce((sum, item) => sum + item.original_qty, 0)
})

const totalAdjustedQty = computed(() => {
  return verifyItems.value.reduce((sum, item) => sum + item.quantity, 0)
})
</script>

<template>
  <div class="p-6 space-y-6 max-w-7xl mx-auto">
    <!-- Header Page -->
    <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold tracking-tight flex items-center flex-wrap gap-2">
          <span>Penerimaan &amp; Verifikasi Nota</span>
          <span v-if="masterDataStore.selectedBranchName" class="text-primary text-xl font-bold">
            — {{ masterDataStore.selectedBranchName }}
          </span>
        </h1>
        <p class="text-muted-foreground text-sm">
          Menu terakhir untuk memverifikasi penerimaan barang, penyesuaian pesanan pengiriman, dan pembaruan omzet salesman.
        </p>
      </div>
      <div class="flex items-center gap-2">
        <RefreshDataButton
          :loading="isRefreshing || isLoading || isTempoLoading"
          :last-updated="lastRefreshedTimeStr"
          @click="triggerRefresh()"
        />
      </div>
    </div>

    <!-- Scope Switcher: Pengiriman Driver vs Sales Tempo -->
    <div class="flex items-center gap-2 border-b border-border pb-3">
      <button
        type="button"
        @click="activeScope = 'DELIVERY'"
        class="flex items-center gap-2 px-5 py-2.5 rounded-xl text-sm font-bold transition-all shadow-sm"
        :class="activeScope === 'DELIVERY' ? 'bg-primary text-primary-foreground' : 'bg-muted/50 text-muted-foreground hover:bg-muted'"
      >
        <Truck class="w-4 h-4" />
        Faktur Pengiriman (Surat Jalan)
        <Badge variant="secondary" v-if="activeTab === 'APPROVED' && totalItems > 0" class="ml-1.5 bg-yellow-400 text-yellow-950 font-black text-[10px]">
          {{ totalItems }}
        </Badge>
      </button>
      <button
        type="button"
        @click="() => { activeScope = 'SALES_TEMPO'; if (tempoPayments.length === 0) fetchTempoPayments() }"
        class="flex items-center gap-2 px-5 py-2.5 rounded-xl text-sm font-bold transition-all shadow-sm"
        :class="activeScope === 'SALES_TEMPO' ? 'bg-indigo-600 text-white' : 'bg-muted/50 text-muted-foreground hover:bg-muted'"
      >
        <Wallet class="w-4 h-4" />
        Sales &amp; Progress Pembayaran Tempo
        <Badge variant="secondary" v-if="tempoTotalItems > 0 && tempoActiveTab === 'PENDING'" class="ml-1.5 bg-amber-400 text-amber-950 font-black text-[10px]">
          {{ tempoTotalItems }}
        </Badge>
      </button>
    </div>

    <!-- ============================================================== -->
    <!-- DELIVERY SCOPE -->
    <!-- ============================================================== -->
    <div v-if="activeScope === 'DELIVERY'" class="space-y-6">
      <!-- Stats summary cards -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <Card class="bg-card border shadow-sm">
        <CardContent class="p-4 flex items-center justify-between">
          <div>
            <p class="text-xs font-semibold text-muted-foreground uppercase">Menunggu Verifikasi</p>
            <h3 class="text-2xl font-bold mt-1 text-yellow-600">
              {{ activeTab === 'APPROVED' ? totalItems : '-' }}
            </h3>
            <p class="text-xs text-muted-foreground mt-1">Faktur Penjualan baru</p>
          </div>
          <div class="p-3 bg-yellow-100 dark:bg-yellow-950/40 text-yellow-600 rounded-full">
            <Clock class="w-6 h-6" />
          </div>
        </CardContent>
      </Card>

      <Card class="bg-card border shadow-sm">
        <CardContent class="p-4 flex items-center justify-between">
          <div>
            <p class="text-xs font-semibold text-muted-foreground uppercase">Selesai Diverifikasi</p>
            <h3 class="text-2xl font-bold mt-1 text-emerald-600">
              {{ activeTab === 'VERIFIED' ? totalItems : '-' }}
            </h3>
            <p class="text-xs text-muted-foreground mt-1">Telah disetujui &amp; update omzet</p>
          </div>
          <div class="p-3 bg-emerald-100 dark:bg-emerald-950/40 text-emerald-600 rounded-full">
            <CheckCircle2 class="w-6 h-6" />
          </div>
        </CardContent>
      </Card>

      <Card class="bg-card border shadow-sm">
        <CardContent class="p-4 flex items-center justify-between">
          <div>
            <p class="text-xs font-semibold text-muted-foreground uppercase">Total Nilai Faktur</p>
            <h3 class="text-2xl font-bold mt-1">
              {{ formatPrice(totalAmountFiltered) }}
            </h3>
            <p class="text-xs text-muted-foreground mt-1">Pada tab &amp; filter aktif</p>
          </div>
          <div class="p-3 bg-primary/10 text-primary rounded-full">
            <FileText class="w-6 h-6" />
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- Tabs Selection -->
    <div class="border-b border-border flex items-center gap-4">
      <button
        @click="activeTab = 'APPROVED'"
        class="py-2.5 px-4 font-semibold text-sm transition-all border-b-2 -mb-[2px] flex items-center gap-2"
        :class="activeTab === 'APPROVED' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'"
      >
        <Clock class="w-4 h-4" />
        Menunggu Verifikasi
        <Badge variant="secondary" v-if="activeTab === 'APPROVED'" class="ml-1 bg-yellow-100 text-yellow-800 dark:bg-yellow-950 dark:text-yellow-200">
          {{ totalItems }}
        </Badge>
      </button>
      <button
        @click="activeTab = 'VERIFIED'"
        class="py-2.5 px-4 font-semibold text-sm transition-all border-b-2 -mb-[2px] flex items-center gap-2"
        :class="activeTab === 'VERIFIED' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'"
      >
        <CheckCircle2 class="w-4 h-4" />
        Riwayat Verifikasi
      </button>
    </div>

    <!-- Filters Section -->
    <Card class="bg-card border shadow-sm">
      <CardContent class="p-4 space-y-4">
        <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
          <!-- Search -->
          <div class="space-y-1.5">
            <label class="text-xs font-semibold text-muted-foreground">Cari Faktur / Toko</label>
            <div class="relative">
              <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
              <input
                v-model="searchQuery"
                type="text"
                placeholder="No Faktur, Nama Toko..."
                class="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 pl-9 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                @keyup.enter="applyFilters"
              />
            </div>
          </div>

          <!-- Salesman filter -->
          <div class="space-y-1.5">
            <label class="text-xs font-semibold text-muted-foreground">Salesman</label>
            <select
              v-model="selectedSalesman"
              class="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
              @change="applyFilters"
            >
              <option value="">Semua Salesman</option>
              <option v-for="sales in filteredSalesmen" :key="sales.id" :value="`${sales.first_name} ${sales.last_name || ''}`">
                {{ sales.first_name }} {{ sales.last_name || '' }} ({{ sales.job_position?.title || 'Sales' }})
              </option>
            </select>
          </div>

          <!-- Date Filter -->
          <div class="space-y-1.5">
            <label class="text-xs font-semibold text-muted-foreground">Tanggal Transaksi</label>
            <div class="relative">
              <input
                v-model="selectedDate"
                type="date"
                class="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                @change="applyFilters"
              />
            </div>
          </div>

          <!-- Filter actions -->
          <div class="flex items-end gap-2">
            <Button @click="applyFilters" class="flex-1">
              Filter
            </Button>
            <Button variant="outline" @click="resetFilters">
              Reset
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- Alert Gagal Pengiriman -->
    <div v-if="failedDeliveriesCount > 0" class="mb-4 p-4 bg-red-50 border border-red-200 rounded-lg dark:bg-red-950/20 dark:border-red-900 flex items-start gap-3">
      <AlertCircle class="w-5 h-5 text-red-600 dark:text-red-400 shrink-0 mt-0.5" />
      <div>
        <h4 class="font-semibold text-red-800 dark:text-red-300 text-sm">
          Perhatian: Terdapat {{ failedDeliveriesCount }} Pengiriman yang Gagal (Toko Tutup)
        </h4>
        <p class="text-xs text-red-700 dark:text-red-400 mt-1">
          Pesanan ini gagal dikirim oleh driver karena toko tutup. Mohon klik tombol <strong>Reset & Kirim Ulang</strong> pada daftar faktur di bawah untuk menjadwalkan ulang pengiriman.
        </p>
      </div>
    </div>

    <!-- Data Table -->
    <Card class="bg-card border shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-sm text-left text-foreground border-collapse">
          <thead class="bg-muted text-muted-foreground text-xs uppercase font-semibold">
            <tr>
              <th class="p-4 border-b">Faktur Penjualan</th>
              <th class="p-4 border-b">Surat Jalan (SJ)</th>
              <th class="p-4 border-b">Toko / Customer</th>
              <th class="p-4 border-b">Salesman</th>
              <th class="p-4 border-b">Tgl Transaksi</th>
              <th class="p-4 border-b">Nilai Faktur</th>
              <th class="p-4 border-b text-center">Status</th>
              <th class="p-4 border-b text-center">Cashier</th>
              <th class="p-4 border-b text-right">Aksi</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="isLoading" class="border-b">
              <td colspan="9" class="p-8 text-center text-muted-foreground">
                <RefreshCw class="w-6 h-6 animate-spin mx-auto mb-2 text-primary" />
                Memuat data penerimaan...
              </td>
            </tr>
            <tr v-else-if="transactions.length === 0" class="border-b">
              <td colspan="9" class="p-8 text-center text-muted-foreground">
                Tidak ada data transaksi yang ditemukan.
              </td>
            </tr>
            <tr
              v-else
              v-for="trx in transactions"
              :key="trx.id"
              class="border-b hover:bg-muted/30 transition-colors"
            >
              <td class="p-4 font-mono font-semibold text-primary">
                {{ trx.receipt_no }}
                <div v-if="hasFailedDelivery(trx)" class="mt-1">
                  <Badge class="bg-red-500 hover:bg-red-600 text-white font-bold text-[9px] uppercase px-1.5 py-0.5 rounded flex items-center gap-1 w-fit animate-pulse">
                    ⚠️ Toko Tutup
                  </Badge>
                </div>
              </td>
              <td class="p-4 font-mono">
                <div v-if="getDeliveryOrderNo(trx) !== '-'" class="flex flex-col">
                  <span class="font-semibold">{{ getDeliveryOrderNo(trx) }}</span>
                  <span class="text-xs text-muted-foreground">Driver: {{ getDriverName(trx) }}</span>
                </div>
                <span v-else class="text-muted-foreground">-</span>
              </td>
              <td class="p-4">
                <div class="font-semibold">{{ trx.store?.name || 'Toko Noname' }}</div>
                <div class="text-xs text-muted-foreground">{{ trx.store?.address || '-' }}</div>
              </td>
              <td class="p-4">
                <div class="font-semibold">{{ trx.employee?.first_name }} {{ trx.employee?.last_name }}</div>
                <div class="text-xs text-muted-foreground bg-primary/5 text-primary w-fit px-1.5 py-0.5 rounded uppercase font-bold text-[10px]">
                  {{ trx.employee?.job_position?.title || 'Sales' }}
                </div>
              </td>
              <td class="p-4">
                {{ formatDate(trx.transaction_date) }}
              </td>
              <td class="p-4 font-semibold text-right">
                <div>{{ formatPrice(trx.total_amount) }}</div>
                <div class="text-[10px] mt-1 font-medium font-sans">
                  <span v-if="trx.status === 'APPROVED' || trx.status === 'VERIFIED'" class="text-emerald-600 dark:text-emerald-400">
                    Disetujui Admin Nota
                  </span>
                  <span v-else class="text-amber-600 dark:text-amber-400">
                    Belum Disetujui
                  </span>
                </div>
              </td>
              <td class="p-4 text-center">
                <Badge
                  :variant="trx.status === 'VERIFIED' ? 'default' : 'secondary'"
                  class="font-semibold uppercase"
                  :class="trx.status === 'VERIFIED' ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-200' : 'bg-yellow-100 text-yellow-800 dark:bg-yellow-950 dark:text-yellow-200'"
                >
                  {{ trx.status === 'VERIFIED' ? 'Diverifikasi' : 'Tertunda' }}
                </Badge>
              </td>
              <td class="p-4 text-center">
                <Badge
                  :variant="trx.payment_status === 'PAID' ? 'default' : 'secondary'"
                  class="font-semibold text-[10px]"
                  :class="trx.payment_status === 'PAID' ? 'bg-emerald-100 text-emerald-800' : 'bg-gray-100 text-gray-600'"
                >
                  {{ trx.payment_status === 'PAID' ? 'Tervalidasi' : 'Belum' }}
                </Badge>
              </td>
              <td class="p-4 text-right">
                <div class="flex items-center justify-end gap-2">
                  <Button
                    v-if="trx.status === 'APPROVED' && trx.payment_status === 'PAID'"
                    size="sm"
                    @click="openVerifyModal(trx)"
                    class="bg-yellow-600 hover:bg-yellow-500 text-white flex items-center gap-1.5"
                  >
                    <CheckCircle2 class="w-3.5 h-3.5" />
                    Verifikasi
                  </Button>
                  <Button
                    v-if="trx.status === 'APPROVED' && hasFailedDelivery(trx)"
                    size="sm"
                    variant="destructive"
                    @click="openRedeliverModal(trx)"
                    class="bg-red-600 hover:bg-red-500 text-white flex items-center gap-1.5"
                  >
                    <RotateCcw class="w-3.5 h-3.5" />
                    Reset &amp; Kirim Ulang
                  </Button>
                  <Button
                    size="sm"
                    variant="outline"
                    @click="openDetailsModal(trx)"
                    class="flex items-center gap-1.5"
                  >
                    <Eye class="w-3.5 h-3.5" />
                    Detail
                  </Button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination Footer -->
      <div class="p-4 flex items-center justify-between border-t bg-muted/20">
        <div class="text-xs text-muted-foreground">
          Menampilkan {{ transactions.length }} dari {{ totalItems }} item
        </div>
        <div class="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            :disabled="currentPage === 1"
            @click="currentPage--"
          >
            Sebelumnya
          </Button>
          <span class="text-xs font-semibold px-2">Halaman {{ currentPage }}</span>
          <Button
            variant="outline"
            size="sm"
            :disabled="currentPage * itemsPerPage >= totalItems"
            @click="currentPage++"
          >
            Berikutnya
          </Button>
        </div>
      </div>
    </Card>
    </div>

    <!-- ============================================================== -->
    <!-- SALES TEMPO SCOPE -->
    <!-- ============================================================== -->
    <div v-else-if="activeScope === 'SALES_TEMPO'" class="space-y-6">
      <!-- Stats summary cards -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card class="bg-card border shadow-sm">
          <CardContent class="p-4 flex items-center justify-between">
            <div>
              <p class="text-xs font-semibold text-muted-foreground uppercase">Cicilan Menunggu Verifikasi</p>
              <h3 class="text-2xl font-bold mt-1 text-amber-600">
                {{ tempoActiveTab === 'PENDING' ? tempoTotalItems : '-' }}
              </h3>
              <p class="text-xs text-muted-foreground mt-1">ACC Kasir siap disahkan ke omzet</p>
            </div>
            <div class="p-3 bg-amber-100 dark:bg-amber-950/40 text-amber-600 rounded-full">
              <Clock class="w-6 h-6" />
            </div>
          </CardContent>
        </Card>

        <Card class="bg-card border shadow-sm">
          <CardContent class="p-4 flex items-center justify-between">
            <div>
              <p class="text-xs font-semibold text-muted-foreground uppercase">Selesai Diverifikasi</p>
              <h3 class="text-2xl font-bold mt-1 text-emerald-600">
                {{ tempoActiveTab === 'VERIFIED' ? tempoTotalItems : '-' }}
              </h3>
              <p class="text-xs text-muted-foreground mt-1">Sah masuk pencapaian omzet sales</p>
            </div>
            <div class="p-3 bg-emerald-100 dark:bg-emerald-950/40 text-emerald-600 rounded-full">
              <CheckCircle2 class="w-6 h-6" />
            </div>
          </CardContent>
        </Card>

        <Card class="bg-card border shadow-sm">
          <CardContent class="p-4 flex items-center justify-between">
            <div>
              <p class="text-xs font-semibold text-muted-foreground uppercase">Total Nilai Pembayaran</p>
              <h3 class="text-2xl font-bold mt-1 text-indigo-600">
                {{ formatPrice(tempoTotalAmount) }}
              </h3>
              <p class="text-xs text-muted-foreground mt-1">Pada tab &amp; filter aktif</p>
            </div>
            <div class="p-3 bg-indigo-100 dark:bg-indigo-950/40 text-indigo-600 rounded-full">
              <Wallet class="w-6 h-6" />
            </div>
          </CardContent>
        </Card>
      </div>

      <!-- Tabs Selection -->
      <div class="border-b border-border flex items-center gap-4">
        <button
          @click="tempoActiveTab = 'PENDING'"
          class="py-2.5 px-4 font-semibold text-sm transition-all border-b-2 -mb-[2px] flex items-center gap-2"
          :class="tempoActiveTab === 'PENDING' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'"
        >
          <Clock class="w-4 h-4" />
          Menunggu Verifikasi (ACC Kasir)
          <Badge variant="secondary" v-if="tempoActiveTab === 'PENDING'" class="ml-1 bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-200">
            {{ tempoTotalItems }}
          </Badge>
        </button>
        <button
          @click="tempoActiveTab = 'VERIFIED'"
          class="py-2.5 px-4 font-semibold text-sm transition-all border-b-2 -mb-[2px] flex items-center gap-2"
          :class="tempoActiveTab === 'VERIFIED' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'"
        >
          <CheckCircle2 class="w-4 h-4" />
          Riwayat Verifikasi (Sah di Omzet)
          <Badge variant="secondary" v-if="tempoActiveTab === 'VERIFIED'" class="ml-1 bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-200">
            {{ tempoTotalItems }}
          </Badge>
        </button>
      </div>

      <!-- Filters Section -->
      <Card class="bg-card border shadow-sm">
        <CardContent class="p-4 space-y-4">
          <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
            <!-- Search -->
            <div class="space-y-1.5">
              <label class="text-xs font-semibold text-muted-foreground">Cari Faktur / Toko / Sales</label>
              <div class="relative">
                <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                <input
                  v-model="tempoSearchQuery"
                  type="text"
                  placeholder="No Faktur, Nama Toko, Sales..."
                  class="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 pl-9 text-sm shadow-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                  @keyup.enter="applyTempoFilters"
                />
              </div>
            </div>

            <!-- Salesman filter -->
            <div class="space-y-1.5">
              <label class="text-xs font-semibold text-muted-foreground">Salesman</label>
              <select
                v-model="tempoSelectedSalesman"
                class="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                @change="applyTempoFilters"
              >
                <option value="">Semua Salesman</option>
                <option v-for="sales in filteredSalesmen" :key="sales.id" :value="sales.id">
                  {{ sales.first_name }} {{ sales.last_name || '' }} ({{ sales.job_position?.title || 'Sales' }})
                </option>
              </select>
            </div>

            <!-- Date Filter -->
            <div class="space-y-1.5">
              <label class="text-xs font-semibold text-muted-foreground">Tgl Pembayaran</label>
              <div class="relative">
                <input
                  v-model="tempoSelectedDate"
                  type="date"
                  class="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                  @change="applyTempoFilters"
                />
              </div>
            </div>

            <!-- Filter actions -->
            <div class="flex items-end gap-2">
              <Button @click="applyTempoFilters" class="flex-1 bg-indigo-600 hover:bg-indigo-500 text-white">
                Filter
              </Button>
              <Button variant="outline" @click="resetTempoFilters">
                Reset
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      <!-- Data Table -->
      <Card class="bg-card border shadow-sm overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full text-sm text-left text-foreground border-collapse">
            <thead class="bg-muted text-muted-foreground text-xs uppercase font-semibold">
              <tr>
                <th class="p-4 border-b">Faktur Penjualan</th>
                <th class="p-4 border-b">Toko / Customer</th>
                <th class="p-4 border-b">Salesman</th>
                <th class="p-4 border-b">Rincian Pembayaran</th>
                <th class="p-4 border-b text-right">Nominal ACC Kasir</th>
                <th class="p-4 border-b">Status Piutang Nota</th>
                <th class="p-4 border-b text-center">Validasi Kasir</th>
                <th class="p-4 border-b text-center">Status Penerimaan</th>
                <th class="p-4 border-b text-right">Aksi</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="isTempoLoading" class="border-b">
                <td colspan="9" class="p-8 text-center text-muted-foreground">
                  <RefreshCw class="w-6 h-6 animate-spin mx-auto mb-2 text-indigo-600" />
                  Memuat data pembayaran tempo...
                </td>
              </tr>
              <tr v-else-if="tempoPayments.length === 0" class="border-b">
                <td colspan="9" class="p-8 text-center text-muted-foreground">
                  Tidak ada data pembayaran tempo yang ditemukan.
                </td>
              </tr>
              <tr
                v-else
                v-for="pmt in tempoPayments"
                :key="pmt.id"
                class="border-b hover:bg-muted/30 transition-colors"
              >
                <!-- Faktur Penjualan -->
                <td class="p-4 font-mono">
                  <div class="font-bold text-primary">{{ pmt.sales_transaction?.receipt_no || '-' }}</div>
                  <div class="text-xs text-muted-foreground">Total: {{ formatPrice(pmt.sales_transaction?.total_amount || 0) }}</div>
                </td>

                <!-- Toko / Customer -->
                <td class="p-4">
                  <div class="font-semibold">{{ pmt.sales_transaction?.store?.name || 'Toko Noname' }}</div>
                  <div class="text-xs text-muted-foreground line-clamp-1">{{ pmt.sales_transaction?.store?.address || '-' }}</div>
                </td>

                <!-- Salesman -->
                <td class="p-4">
                  <div class="font-semibold">
                    {{ (pmt.sales_transaction?.employee || pmt.employee)?.first_name }} {{ (pmt.sales_transaction?.employee || pmt.employee)?.last_name }}
                  </div>
                  <div class="text-xs text-muted-foreground bg-primary/5 text-primary w-fit px-1.5 py-0.5 rounded uppercase font-bold text-[10px]">
                    {{ (pmt.sales_transaction?.employee || pmt.employee)?.job_position?.title || 'Sales' }}
                  </div>
                </td>

                <!-- Rincian Pembayaran -->
                <td class="p-4">
                  <div class="flex items-center gap-2">
                    <Badge variant="outline" class="font-semibold text-xs">
                      {{ pmt.payment_method }}
                    </Badge>
                    <span class="text-xs text-muted-foreground">{{ formatDate(pmt.payment_date) }}</span>
                  </div>
                  <div class="text-xs text-muted-foreground mt-1 line-clamp-1">
                    {{ pmt.notes || 'Pembayaran tempo/cicilan' }}
                  </div>
                  <div v-if="pmt.transfer_proof_url" class="mt-1.5">
                    <button
                      type="button"
                      @click="openImagePreview(getFullImageUrl(pmt.transfer_proof_url))"
                      class="inline-flex items-center gap-1 text-[11px] text-indigo-600 hover:text-indigo-700 font-semibold underline decoration-dotted"
                    >
                      <ImageIcon class="w-3.5 h-3.5" />
                      Lihat Bukti Transfer
                    </button>
                  </div>
                </td>

                <!-- Nominal ACC Kasir -->
                <td class="p-4 text-right">
                  <div class="text-sm font-bold text-emerald-600 dark:text-emerald-400">
                    {{ formatPrice(pmt.amount) }}
                  </div>
                  <span class="text-[10px] text-muted-foreground">Disetujui Kasir</span>
                </td>

                <!-- Status Piutang Nota -->
                <td class="p-4">
                  <div class="text-xs space-y-0.5">
                    <div>Terbayar: <span class="font-semibold text-emerald-600">{{ formatPrice(pmt.sales_transaction?.paid_amount || 0) }}</span></div>
                    <div>Sisa: <span class="font-semibold text-amber-600">{{ formatPrice(pmt.sales_transaction?.remaining_amount || 0) }}</span></div>
                  </div>
                </td>

                <!-- Validasi Kasir -->
                <td class="p-4 text-center">
                  <Badge class="bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-200 text-xs">
                    ACC Kasir
                  </Badge>
                  <div class="text-[11px] text-muted-foreground mt-1">
                    {{ pmt.verified_by?.first_name || 'Kasir' }}
                  </div>
                  <div class="text-[10px] text-muted-foreground">
                    {{ formatDate(pmt.approved_at) }}
                  </div>
                </td>

                <!-- Status Penerimaan -->
                <td class="p-4 text-center">
                  <div v-if="pmt.penerimaan_status === 'VERIFIED'">
                    <Badge class="bg-emerald-600 text-white font-bold text-xs flex items-center justify-center gap-1 mx-auto w-fit">
                      <CheckCircle2 class="w-3 h-3" /> Sah di Omzet
                    </Badge>
                    <div class="text-[11px] text-muted-foreground mt-1">
                      {{ pmt.penerimaan_verified_by?.first_name || 'Admin Nota' }}
                    </div>
                    <div class="text-[10px] text-muted-foreground">
                      {{ formatDate(pmt.penerimaan_verified_at) }}
                    </div>
                  </div>
                  <div v-else>
                    <Badge class="bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-200 text-xs flex items-center justify-center gap-1 mx-auto w-fit">
                      <Clock class="w-3 h-3" /> Menunggu Admin
                    </Badge>
                  </div>
                </td>

                <!-- Aksi -->
                <td class="p-4 text-right">
                  <div class="flex items-center justify-end gap-2">
                    <Button
                      v-if="pmt.penerimaan_status !== 'VERIFIED'"
                      size="sm"
                      @click="openVerifyTempoModal(pmt)"
                      class="bg-indigo-600 hover:bg-indigo-500 text-white flex items-center gap-1.5 shadow-sm font-semibold"
                    >
                      <CheckCheck class="w-3.5 h-3.5" />
                      Sahkan Omzet
                    </Button>
                    <Badge
                      v-else
                      variant="outline"
                      class="text-xs bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-300 border-emerald-300 font-medium"
                    >
                      <CheckCircle2 class="w-3.5 h-3.5 mr-1" />
                      Selesai
                    </Badge>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination Footer -->
        <div class="p-4 flex items-center justify-between border-t bg-muted/20">
          <div class="text-xs text-muted-foreground">
            Menampilkan {{ tempoPayments.length }} dari {{ tempoTotalItems }} item
          </div>
          <div class="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              :disabled="tempoCurrentPage === 1"
              @click="tempoCurrentPage--"
            >
              Sebelumnya
            </Button>
            <span class="text-xs font-semibold px-2">Halaman {{ tempoCurrentPage }}</span>
            <Button
              variant="outline"
              size="sm"
              :disabled="tempoCurrentPage * tempoItemsPerPage >= tempoTotalItems"
              @click="tempoCurrentPage++"
            >
              Berikutnya
            </Button>
          </div>
        </div>
      </Card>
    </div>
    <div v-if="showVerifyModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="showVerifyModal = false"></div>
      <div class="relative bg-background border rounded-lg shadow-lg w-full max-w-4xl max-h-[90vh] overflow-hidden flex flex-col z-10 animate-in fade-in zoom-in duration-200">
        <!-- Modal Header -->
        <div class="p-4 border-b flex items-center justify-between bg-yellow-600 text-white">
          <div>
            <h2 class="text-lg font-bold">Verifikasi Penerimaan Faktur &amp; Logistics</h2>
            <p class="text-xs opacity-90">Faktur: {{ selectedTransaction?.receipt_no }} | Toko: {{ selectedTransaction?.store?.name }}</p>
          </div>
          <button @click="showVerifyModal = false" class="text-white/80 hover:text-white transition-colors">
            <X class="w-6 h-6" />
          </button>
        </div>

        <!-- Modal Body -->
        <div class="p-6 overflow-y-auto space-y-6 flex-1">
          <!-- Transaction Info Summary -->
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4 bg-muted/40 p-4 rounded-lg border">
            <div>
              <span class="text-xs text-muted-foreground block font-semibold uppercase">Salesman / Employee</span>
              <span class="text-sm font-semibold">{{ selectedTransaction?.employee?.first_name }} {{ selectedTransaction?.employee?.last_name }}</span>
            </div>
            <div>
              <span class="text-xs text-muted-foreground block font-semibold uppercase">Surat Jalan (SJ)</span>
              <span class="text-sm font-semibold">{{ getDeliveryOrderNo(selectedTransaction) }}</span>
            </div>
            <div>
              <span class="text-xs text-muted-foreground block font-semibold uppercase">Driver Pengirim</span>
              <span class="text-sm font-semibold">{{ getDriverName(selectedTransaction) }}</span>
            </div>
          </div>

          <!-- Alert Note about logistics adjustment -->
          <div class="bg-blue-50 dark:bg-blue-950/30 text-blue-800 dark:text-blue-200 p-3 rounded-lg border border-blue-200 dark:border-blue-900 flex items-start gap-2 text-xs">
            <Info class="w-4 h-4 mt-0.5 flex-shrink-0" />
            <div>
              <p class="font-semibold">Informasi Penyesuaian Kuantitas (Quantity Adjustment)</p>
              <p class="mt-1">
                Jika kuantitas barang yang diterima pelanggan tidak sesuai (misal: pesan 2 dus, diterima 1 dus), ubah kolom <strong>"Kuantitas Aktual"</strong> di bawah. Selisih barang akan dikembalikan otomatis ke persediaan salesman/driver dan log penyesuaian akan disimpan.
              </p>
            </div>
          </div>

          <!-- Items list to adjust -->
          <div class="space-y-3">
            <h3 class="font-bold text-sm text-foreground border-b pb-1">Daftar Item &amp; Penyesuaian Barang</h3>
            <div class="overflow-x-auto border rounded-lg">
              <table class="w-full text-xs text-left text-foreground">
                <thead class="bg-muted text-muted-foreground font-semibold">
                  <tr>
                    <th class="p-3">Nama Produk</th>
                    <th class="p-3 text-right">Harga</th>
                    <th class="p-3 text-center">Qty Pesanan (Awal)</th>
                    <th class="p-3 text-center w-32">Qty Aktual (Diterima)</th>
                    <th class="p-3 text-center">Selisih</th>
                    <th class="p-3">Catatan / Alasan Penyesuaian</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="item in verifyItems" :key="item.product_id" class="border-b last:border-0 hover:bg-muted/10">
                    <td class="p-3">
                      <div class="font-semibold">{{ item.name }}</div>
                      <div class="text-[10px] text-muted-foreground font-mono">{{ item.sku }}</div>
                    </td>
                    <td class="p-3 text-right font-semibold">
                      {{ formatPrice(item.price) }}
                    </td>
                    <td class="p-3 text-center font-semibold text-muted-foreground">
                      {{ item.original_qty }}
                    </td>
                    <td class="p-3 text-center">
                      <input
                        v-model.number="item.quantity"
                        type="number"
                        min="0"
                        :max="item.original_qty"
                        class="flex h-8 w-full rounded-md border border-input bg-transparent px-2 py-1 text-center font-bold text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                      />
                    </td>
                    <td class="p-3 text-center font-bold" :class="item.original_qty - item.quantity > 0 ? 'text-red-500' : 'text-muted-foreground'">
                      {{ item.original_qty - item.quantity }}
                    </td>
                    <td class="p-3">
                      <input
                        v-model="item.notes"
                        type="text"
                        placeholder="Misal: Customer menolak 1 dus rusak"
                        :disabled="item.original_qty - item.quantity === 0"
                        class="flex h-8 w-full rounded-md border border-input bg-transparent px-2 py-1 text-xs shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:opacity-40 disabled:cursor-not-allowed"
                      />
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- General note for verification -->
          <div class="space-y-1.5">
            <label class="text-xs font-bold text-foreground">Catatan Verifikasi Admin Nota (General Notes)</label>
            <textarea
              v-model="verificationGeneralNotes"
              rows="3"
              placeholder="Masukkan catatan verifikasi menyeluruh di sini..."
              class="flex w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            ></textarea>
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="p-4 border-t flex items-center justify-between bg-muted/30">
          <div class="flex flex-col text-left">
            <span class="text-xs text-muted-foreground">Estimasi Nilai Faktur Baru:</span>
            <span class="text-lg font-extrabold text-primary">{{ formatPrice(adjustedTotalAmount) }}</span>
          </div>
          <div class="flex items-center gap-2">
            <Button variant="outline" @click="showVerifyModal = false">Batal</Button>
            <Button
              @click="submitVerification"
              class="bg-emerald-600 hover:bg-emerald-500 text-white flex items-center gap-2"
              :disabled="isSubmitting"
            >
              <CheckCircle2 class="w-4 h-4" />
              {{ isSubmitting ? 'Memproses...' : 'Setujui & Update Omzet' }}
            </Button>
          </div>
        </div>
      </div>
    </div>

    <!-- Modal: Detail View -->
    <div v-if="showDetailsModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="showDetailsModal = false"></div>
      <div class="relative bg-background border rounded-lg shadow-lg w-full max-w-3xl max-h-[90vh] overflow-hidden flex flex-col z-10 animate-in fade-in zoom-in duration-200">
        <!-- Modal Header -->
        <div class="p-4 border-b flex items-center justify-between">
          <div>
            <h2 class="text-lg font-bold">Rincian Faktur Penjualan</h2>
            <p class="text-xs text-muted-foreground">Faktur: {{ selectedTransaction?.receipt_no }}</p>
          </div>
          <button @click="showDetailsModal = false" class="text-muted-foreground hover:text-foreground transition-colors">
            <X class="w-6 h-6" />
          </button>
        </div>

        <!-- Modal Body -->
        <div class="p-6 overflow-y-auto space-y-6 flex-1">
          <!-- Grid Info Utama -->
          <div class="grid grid-cols-2 gap-4">
            <div>
              <span class="text-xs text-muted-foreground uppercase font-semibold">Toko / Customer</span>
              <p class="text-sm font-bold mt-0.5">{{ selectedTransaction?.store?.name || 'Toko Noname' }}</p>
              <p class="text-xs text-muted-foreground">{{ selectedTransaction?.store?.address || '-' }}</p>
            </div>
            <div>
              <span class="text-xs text-muted-foreground uppercase font-semibold">Salesman</span>
              <p class="text-sm font-bold mt-0.5">{{ selectedTransaction?.employee?.first_name }} {{ selectedTransaction?.employee?.last_name }}</p>
              <p class="text-xs text-muted-foreground">Status: {{ selectedTransaction?.employee?.job_position?.title }}</p>
            </div>
            <div>
              <span class="text-xs text-muted-foreground uppercase font-semibold">Tanggal Transaksi</span>
              <p class="text-sm font-bold mt-0.5">{{ formatDate(selectedTransaction?.transaction_date) }}</p>
            </div>
            <div>
              <span class="text-xs text-muted-foreground uppercase font-semibold">Nilai Faktur Penjualan</span>
              <p class="text-sm font-extrabold text-primary mt-0.5">{{ formatPrice(selectedTransaction?.total_amount) }}</p>
            </div>
            <div>
              <span class="text-xs text-muted-foreground uppercase font-semibold">Surat Jalan (Logistics)</span>
              <p class="text-sm font-bold mt-0.5">{{ getDeliveryOrderNo(selectedTransaction) }}</p>
            </div>
            <div>
              <span class="text-xs text-muted-foreground uppercase font-semibold">Driver</span>
              <p class="text-sm font-bold mt-0.5">{{ getDriverName(selectedTransaction) }}</p>
            </div>
          </div>

          <!-- Status Verifikasi Audit -->
          <div class="p-4 bg-muted/40 rounded-lg border space-y-2">
            <h4 class="font-bold text-xs uppercase tracking-wider text-muted-foreground">Status Verifikasi Penerimaan</h4>
            <div class="flex items-center gap-2">
              <Badge
                :variant="selectedTransaction?.status === 'VERIFIED' ? 'default' : 'secondary'"
                class="font-semibold uppercase"
                :class="selectedTransaction?.status === 'VERIFIED' ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-200' : 'bg-yellow-100 text-yellow-800 dark:bg-yellow-950 dark:text-yellow-200'"
              >
                {{ selectedTransaction?.status === 'VERIFIED' ? 'Diverifikasi' : 'Tertunda' }}
              </Badge>
              <span v-if="selectedTransaction?.status === 'VERIFIED'" class="text-xs text-muted-foreground">
                Oleh: {{ selectedTransaction?.verified_by ? `${selectedTransaction.verified_by.first_name} ${selectedTransaction.verified_by.last_name || ''}` : 'System' }} pada {{ formatDate(selectedTransaction?.verified_at) }}
              </span>
            </div>
          </div>

          <!-- Bukti Pengiriman (POD Image) -->
          <div v-if="deliveryProofImages.length > 0" class="space-y-2">
            <h4 class="font-bold text-sm text-foreground">Bukti Pengiriman (POD)</h4>
            <div class="flex flex-wrap gap-3">
              <div
                v-for="(imgUrl, index) in deliveryProofImages"
                :key="index"
                @click="openImagePreview(imgUrl)"
                class="relative w-32 h-32 rounded-lg overflow-hidden border border-border shadow-sm cursor-zoom-in hover:opacity-90 transition-all shrink-0 bg-muted flex items-center justify-center group"
                title="Klik untuk memperbesar"
              >
                <img :src="imgUrl" class="w-full h-full object-cover" />
                <div class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
                  <span class="text-white text-xs font-semibold">Perbesar</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Items Table -->
          <div class="space-y-2">
            <h4 class="font-bold text-sm text-foreground">Detail Barang Faktur</h4>
            <div class="border rounded-lg overflow-hidden">
              <table class="w-full text-xs text-left text-foreground">
                <thead class="bg-muted text-muted-foreground font-semibold">
                  <tr>
                    <th class="p-3">Nama Produk</th>
                    <th class="p-3 text-right">Harga Satuan</th>
                    <th class="p-3 text-center">Kuantitas</th>
                    <th class="p-3 text-right">Subtotal</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="item in selectedTransaction?.items" :key="item.id" class="border-b last:border-0">
                    <td class="p-3 font-semibold">
                      {{ item.product?.name || 'Produk Tidak Dikenal' }}
                    </td>
                    <td class="p-3 text-right">
                      {{ formatPrice(item.price_at_transaction) }}
                    </td>
                    <td class="p-3 text-center font-bold">
                      {{ item.quantity }}
                    </td>
                    <td class="p-3 text-right font-bold">
                      {{ formatPrice(item.subtotal) }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Adjustment Logs (penyesuaian barang) jika ada -->
          <div v-if="selectedTransaction?.adjustment_logs && selectedTransaction.adjustment_logs.length > 0" class="space-y-3">
            <h4 class="font-bold text-sm text-red-500 flex items-center gap-1.5">
              <BadgeAlert class="w-4 h-4" />
              Log Penyesuaian Barang (Kuantitas Tidak Terkirim Full)
            </h4>
            <div class="border border-red-200 dark:border-red-950/60 rounded-lg overflow-hidden">
              <table class="w-full text-xs text-left text-foreground">
                <thead class="bg-red-50 dark:bg-red-950/30 text-red-700 dark:text-red-300 font-semibold">
                  <tr>
                    <th class="p-3">Nama Produk</th>
                    <th class="p-3 text-center">Qty Awal</th>
                    <th class="p-3 text-center">Selisih</th>
                    <th class="p-3 text-center">Qty Akhir</th>
                    <th class="p-3">Alasan / Catatan Penyesuaian</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="log in selectedTransaction.adjustment_logs" :key="log.id" class="border-b last:border-0 hover:bg-red-50/10">
                    <td class="p-3 font-semibold">{{ log.product?.name || 'Produk Tidak Dikenal' }}</td>
                    <td class="p-3 text-center font-medium text-muted-foreground">{{ log.original_quantity }}</td>
                    <td class="p-3 text-center font-extrabold text-red-600">-{{ log.difference }}</td>
                    <td class="p-3 text-center font-bold text-emerald-600">{{ log.adjusted_quantity }}</td>
                    <td class="p-3 text-muted-foreground">{{ log.notes || '-' }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Return Items (Barang Sampingan) -->
          <div v-if="parsedTransactionNotes?.return_items && parsedTransactionNotes.return_items.length > 0" class="space-y-2 mt-4">
            <h4 class="font-bold text-sm text-rose-600 flex items-center gap-1.5">
              <BadgeAlert class="w-4 h-4" />
              Retur / Pengembalian Barang Sampingan
            </h4>
            <div class="border border-rose-200 dark:border-rose-950/60 rounded-lg overflow-hidden">
              <table class="w-full text-xs text-left text-foreground">
                <thead class="bg-rose-50 dark:bg-rose-950/30 text-rose-700 dark:text-rose-300 font-semibold">
                  <tr>
                    <th class="p-3">Nama Produk</th>
                    <th class="p-3 text-center">Kuantitas</th>
                    <th class="p-3 text-right">Harga Satuan</th>
                    <th class="p-3 text-right">Potongan</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(ret, idx) in parsedTransactionNotes.return_items" :key="'ret-'+idx" class="border-b last:border-0 hover:bg-rose-50/10">
                    <td class="p-3 font-semibold uppercase">{{ ret.product_name }}</td>
                    <td class="p-3 text-center font-bold text-rose-600">{{ ret.quantity }} {{ ret.unit || 'PCS' }}</td>
                    <td class="p-3 text-right">{{ formatPrice(ret.price) }}</td>
                    <td class="p-3 text-right font-bold text-rose-700">-{{ formatPrice(ret.quantity * ret.price) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Notes from Salesman -->
          <div v-if="parsedTransactionNotes?.notes" class="space-y-2 mt-4">
            <h4 class="font-bold text-sm text-foreground flex items-center gap-1.5">
              <MessageSquare class="w-4 h-4" />
              Catatan dari Salesman
            </h4>
            <div class="bg-muted/30 p-3 rounded-lg border border-border">
              <p class="text-foreground whitespace-pre-line">{{ parsedTransactionNotes.notes }}</p>
            </div>
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="p-4 border-t flex justify-end bg-muted/20">
          <Button @click="showDetailsModal = false">Tutup</Button>
        </div>
      </div>
    </div>
  
    <ConfirmationDialog
      :open="showRedeliverModal"
      title="Konfirmasi Redeliver"
      description="Apakah Anda yakin ingin membatalkan pengiriman gagal ini dan mengembalikan pesanan ke tahap pembuatan Surat Jalan? Semua stok gudang utama akan dikembalikan, faktur pending ini akan dihapus, dan Sales Order (SO) dapat dibatch ulang tanpa harus menginput pesanan dari awal."
      variant="danger"
      @confirm="confirmRedeliver"
      @cancel="showRedeliverModal = false"
    />

    <!-- Image Preview Modal -->
    <div v-if="showImageModal" class="fixed inset-0 z-[100] flex items-center justify-center p-4">
      <div
        class="absolute inset-0 bg-slate-900/90 backdrop-blur-md"
        @click="showImageModal = false"
      ></div>

      <!-- Close button -->
      <div class="absolute top-6 right-6 flex items-center gap-3 z-[110]">
        <button
          @click="showImageModal = false"
          class="p-3 bg-white hover:bg-slate-100 text-slate-900 rounded-2xl transition-all shadow-lg"
        >
          <X class="w-6 h-6" />
        </button>
      </div>

      <div
        class="relative z-[105] max-w-5xl max-h-[85vh] flex items-center justify-center overflow-hidden rounded-3xl shadow-2xl bg-black"
      >
        <img
          :src="previewImageUrl"
          class="max-w-full max-h-full object-contain"
        />
      </div>
    </div>

    <!-- Modal: Verification / Sahkan Omzet Pembayaran Tempo -->
    <div v-if="showVerifyTempoModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="showVerifyTempoModal = false"></div>
      <div class="relative bg-background border rounded-xl shadow-2xl w-full max-w-xl max-h-[90vh] overflow-hidden flex flex-col z-10 animate-in fade-in zoom-in duration-200">
        <!-- Modal Header -->
        <div class="p-5 border-b flex items-center justify-between bg-indigo-600 text-white">
          <div class="flex items-center gap-3">
            <div class="p-2 bg-white/20 rounded-lg">
              <Wallet class="w-5 h-5 text-white" />
            </div>
            <div>
              <h2 class="text-lg font-bold leading-tight">Pengesahan Penerimaan Pembayaran Tempo</h2>
              <p class="text-xs text-white/80 mt-0.5">
                Faktur: {{ selectedTempoPayment?.sales_transaction?.receipt_no }} | Toko: {{ selectedTempoPayment?.sales_transaction?.store?.name }}
              </p>
            </div>
          </div>
          <button @click="showVerifyTempoModal = false" class="text-white/80 hover:text-white transition-colors">
            <X class="w-6 h-6" />
          </button>
        </div>

        <!-- Modal Body -->
        <div class="p-6 overflow-y-auto space-y-5">
          <!-- Highlight Nominal -->
          <div class="p-4 bg-indigo-50 dark:bg-indigo-950/40 border border-indigo-200 dark:border-indigo-900 rounded-xl flex items-center justify-between">
            <div>
              <p class="text-xs font-semibold text-indigo-700 dark:text-indigo-300 uppercase tracking-wider">Nominal Pembayaran (ACC Kasir)</p>
              <h3 class="text-2xl font-extrabold text-indigo-900 dark:text-indigo-100 mt-1">
                {{ formatPrice(selectedTempoPayment?.amount || 0) }}
              </h3>
            </div>
            <div class="text-right">
              <Badge class="bg-emerald-600 text-white text-xs font-bold">
                {{ selectedTempoPayment?.payment_method }}
              </Badge>
              <p class="text-xs text-muted-foreground mt-1">{{ formatDate(selectedTempoPayment?.payment_date) }}</p>
            </div>
          </div>

          <!-- Transaction & Salesman Info -->
          <div class="grid grid-cols-2 gap-4 text-xs bg-muted/40 p-4 rounded-xl border">
            <div>
              <span class="text-muted-foreground block">Salesman Penerima Omzet:</span>
              <span class="font-bold text-foreground text-sm">
                {{ (selectedTempoPayment?.sales_transaction?.employee || selectedTempoPayment?.employee)?.first_name }}
                {{ (selectedTempoPayment?.sales_transaction?.employee || selectedTempoPayment?.employee)?.last_name }}
              </span>
              <p class="text-[11px] text-muted-foreground">
                {{ (selectedTempoPayment?.sales_transaction?.employee || selectedTempoPayment?.employee)?.job_position?.title || 'Sales' }}
              </p>
            </div>
            <div>
              <span class="text-muted-foreground block">Disetujui Kasir:</span>
              <span class="font-bold text-foreground text-sm">
                {{ selectedTempoPayment?.verified_by?.first_name || 'Kasir' }} {{ selectedTempoPayment?.verified_by?.last_name || '' }}
              </span>
              <p class="text-[11px] text-muted-foreground">
                {{ formatDate(selectedTempoPayment?.approved_at) }}
              </p>
            </div>
            <div class="col-span-2 pt-2 border-t flex items-center justify-between">
              <div>
                <span class="text-muted-foreground">Total Nilai Faktur:</span>
                <span class="font-semibold text-foreground ml-1">{{ formatPrice(selectedTempoPayment?.sales_transaction?.total_amount || 0) }}</span>
              </div>
              <div>
                <span class="text-muted-foreground">Sisa Piutang:</span>
                <span class="font-bold text-amber-600 ml-1">{{ formatPrice(selectedTempoPayment?.sales_transaction?.remaining_amount || 0) }}</span>
              </div>
            </div>
            <div v-if="selectedTempoPayment?.notes" class="col-span-2 pt-2 border-t">
              <span class="text-muted-foreground">Catatan Pembayaran:</span>
              <p class="text-foreground italic mt-0.5">{{ selectedTempoPayment.notes }}</p>
            </div>
            <div v-if="selectedTempoPayment?.transfer_proof_url" class="col-span-2 pt-2 border-t flex items-center justify-between">
              <span class="text-muted-foreground">Bukti Transfer:</span>
              <button
                type="button"
                @click="openImagePreview(getFullImageUrl(selectedTempoPayment.transfer_proof_url))"
                class="inline-flex items-center gap-1 text-xs text-indigo-600 hover:text-indigo-700 font-semibold underline"
              >
                <ImageIcon class="w-3.5 h-3.5" />
                Buka Gambar Bukti Transfer
              </button>
            </div>
          </div>

          <!-- Explanation Callout -->
          <div class="flex items-start gap-3 p-3.5 bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-900 rounded-xl text-xs text-amber-900 dark:text-amber-200">
            <Info class="w-4 h-4 text-amber-600 shrink-0 mt-0.5" />
            <p>
              Dengan mengesahkan pembayaran tempo ini, nominal <strong>{{ formatPrice(selectedTempoPayment?.amount || 0) }}</strong> akan diverifikasi dan langsung dihitung ke dalam <strong>Pencapaian Omzet Salesman</strong> untuk bulan transaksi.
            </p>
          </div>

          <!-- Verification Notes -->
          <div class="space-y-1.5">
            <label class="text-xs font-semibold text-foreground">Catatan Verifikasi Admin Penerimaan (Opsional)</label>
            <textarea
              v-model="tempoVerificationNotes"
              rows="2"
              placeholder="Contoh: Dana cicilan telah masuk rekening operasional &amp; tervalidasi kasir..."
              class="w-full text-sm p-3 border rounded-lg bg-background focus:ring-2 focus:ring-indigo-500 focus:outline-none"
            ></textarea>
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="p-4 border-t flex items-center justify-end gap-3 bg-muted/20">
          <Button variant="outline" @click="showVerifyTempoModal = false" :disabled="isSubmittingTempoVerify">
            Batal
          </Button>
          <Button
            @click="confirmVerifyTempoPayment"
            :disabled="isSubmittingTempoVerify"
            class="bg-indigo-600 hover:bg-indigo-500 text-white flex items-center gap-2 font-bold px-5"
          >
            <RefreshCw v-if="isSubmittingTempoVerify" class="w-4 h-4 animate-spin" />
            <CheckCheck v-else class="w-4 h-4" />
            {{ isSubmittingTempoVerify ? 'Memproses...' : 'Sahkan ke Omzet Sales' }}
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>
