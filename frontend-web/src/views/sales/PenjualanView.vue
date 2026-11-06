<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import apiClient from '@/api/axios'
import DeliveryManagement from './DeliveryManagement.vue'
import {
  Search,
  Filter,
  CheckCircle2,
  XCircle,
  Eye,
  Banknote,
  Calendar,
  Image as ImageIcon,
  Plus,
  Pencil,
  Trash2,
  Save,
  X,
  Store,
  User,
  Clock,
  Maximize2,
  ZoomIn,
  ZoomOut,
  RotateCcw,
  Truck,
  Printer,
  ClipboardList,
  MapPin,
  RefreshCw,
  PackageOpen,
  AlertCircle,
  AlertTriangle,
  MapPinOff,
  Navigation,
  ShoppingBag,
  FileText,
  Upload,
  CreditCard,
  Wallet,
  CalendarClock,
  ExternalLink,
  Percent,
  ArrowRightLeft,
  Ban,
  Loader2,
} from 'lucide-vue-next'
import { getDistance } from '@/utils/geolocation'
import { filterReturnableProducts } from '@/utils/productUnits'
import { useMasterDataStore } from '@/stores/masterData'
import { useFactoryStore } from '@/stores/factory'
import { useAuthStore } from '@/stores/auth'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardFooter } from '@/components/ui/card/index'
import { Button } from '@/components/ui/button'
import SearchSelect, { type SelectOption } from '@/components/SearchSelect.vue'
import ProductSearchSelect from '@/components/sales/ProductSearchSelect.vue'
import ConfirmationDialog from '@/components/ui/ConfirmationDialog.vue'
import OrderLocationMapModal from '@/components/sales/OrderLocationMapModal.vue'
import DirectSaleModal from '@/components/sales/DirectSaleModal.vue'
import DirectSaleInvoicePrintModal from '@/components/sales/DirectSaleInvoicePrintModal.vue'
import DirectSaleReturnModal from '@/components/sales/DirectSaleReturnModal.vue'
import BulkInvoicePrintModal from '@/components/sales/BulkInvoicePrintModal.vue'
import BranchTransferInvoiceModal from '@/components/sales/BranchTransferInvoiceModal.vue'
import BranchInvoicePrintModal from '@/components/sales/BranchInvoicePrintModal.vue'
import BranchFactoryReturnModal from '@/components/sales/BranchFactoryReturnModal.vue'
import RefreshDataButton from '@/components/ui/RefreshDataButton.vue'
import { useDataRefresh } from '@/composables/useDataRefresh'
import { toast } from 'vue-sonner'

const searchQuery = ref('')
const selectedStatus = ref('ALL')
const isLoading = ref(true)
const offlineImportInput = ref<HTMLInputElement | null>(null)
const isImportingOffline = ref(false)
const showOfflineImportModal = ref(false)
const offlineJsonText = ref('')

const salesData = ref<any[]>([])

const processImportOfflinePayload = async (jsonString: string) => {
  if (!jsonString || !jsonString.trim()) {
    toast.error('Teks data cadangan tidak boleh kosong.')
    return
  }

  try {
    isImportingOffline.value = true
    const payload = JSON.parse(jsonString.trim())
    if (!Array.isArray(payload?.transactions) || payload.transactions.length === 0) {
      throw new Error('Data tidak memiliki daftar transaksi offline yang valid (field "transactions" kosong atau tidak ditemukan).')
    }
    const response = await apiClient.post('/admin/sales/import-offline-backup', {
      transactions: payload.transactions,
    })
    const result = response.data?.data || {}
    toast.success(`Import selesai: ${result.success || 0} berhasil baru, ${result.duplicates || 0} duplikat, ${result.failed || 0} gagal.`)
    if (result.error_items?.length) {
      toast.error(result.error_items.slice(0, 2).join(' | '))
    }
    showOfflineImportModal.value = false
    offlineJsonText.value = ''
    await fetchData()
  } catch (error: any) {
    const message = error?.response?.data?.error || error?.message || 'Gagal mengimpor data offline.'
    toast.error(message)
  } finally {
    isImportingOffline.value = false
  }
}

const importOfflineBackup = async (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return

  try {
    const text = await file.text()
    await processImportOfflinePayload(text)
  } catch (error: any) {
    toast.error('Gagal membaca file: ' + error.message)
  }
}

const handlePasteClipboard = async () => {
  try {
    if (navigator.clipboard && navigator.clipboard.readText) {
      const text = await navigator.clipboard.readText()
      if (text) {
        offlineJsonText.value = text
        toast.info('Teks berhasil ditempel dari clipboard.')
      }
    }
  } catch (_) {
    toast.info('Silakan tekan Ctrl+V / Cmd+V pada kotak teks untuk menempel.')
  }
}
const salesmen = ref<any[]>([])
const allStores = ref<any[]>([])
const salesmanStores = ref<any[]>([])
const isLoadingSalesmanStores = ref(false)
const showModal = ref(false)
const isEditing = ref(false)
const mobileCrudTab = ref<'form' | 'receipt'>('form')
const showImageModal = ref(false)
const previewImageUrl = ref('')

const showDirectSaleModal = ref(false)
const showDirectSalePrintModal = ref(false)
const selectedDirectSale = ref<any>(null)

const showDirectSaleReturnModal = ref(false)
const selectedReturnSale = ref<any>(null)

const openDirectSaleReturn = (sale: any) => {
  selectedReturnSale.value = sale.raw || sale
  showDirectSaleReturnModal.value = true
}

const isPaidInvoiceEdit = computed(() => isEditing.value && Number(currentSale.value.paidAmount || 0) > 0)

const handleDirectSaleReturnCreated = () => {
  showDirectSaleReturnModal.value = false
  fetchData()
}

const handleDirectSaleCreated = (sale: any, printImmediately: boolean) => {
  showDirectSaleModal.value = false
  fetchData()
  if (printImmediately && sale) {
    selectedDirectSale.value = sale
    showDirectSalePrintModal.value = true
  }
}

const openDirectSalePrint = (sale: any) => {
  selectedDirectSale.value = sale.raw || sale
  showDirectSalePrintModal.value = true
}
const showDeleteConfirmModal = ref(false)
const deleteTargetId = ref('')
const deleteTargetSale = ref<any>(null)
const deleteReason = ref('')
const deleteConfirmText = ref('')
const isDeletingTrx = ref(false)

const showRestoreConfirmModal = ref(false)
const restoreTargetSale = ref<any>(null)
const isRestoringTrx = ref(false)

const zoomLevel = ref(1)
const rotationAngle = ref(0)
const crudZoomLevel = ref(1)
const crudRotationAngle = ref(0)
const handleCrudZoom = (val: number) => {
  crudZoomLevel.value = Math.max(0.5, Math.min(3, crudZoomLevel.value + val))
}
const masterDataStore = useMasterDataStore()
const authStore = useAuthStore()

const canEdit = (sale: any) => {
  if (sale.deletedAt || sale.raw?.deleted_at || sale.status === 'CANCELLED') return false
  return (
    sale.status === 'PENDING' ||
    sale.status === 'APPROVED' ||
    sale.status === 'VERIFIED' ||
    authStore.user?.role === 'admin_nota' ||
    authStore.user?.role === 'superadmin'
  )
}
const canDelete = (sale: any) => {
  if (sale.deletedAt || sale.raw?.deleted_at || sale.status === 'CANCELLED') return false
  return (
    sale.status === 'PENDING' ||
    sale.status === 'APPROVED' ||
    authStore.user?.role === 'admin_nota' ||
    authStore.user?.role === 'superadmin'
  )
}

const selectedDO = ref<any>(null)
const selectedNota = ref<any>(null)
const selectedQR = ref<string | null>(null)
const showTempoPaymentModal = ref(false)
const tempoPaymentAmount = ref(0)
const tempoPaymentDate = ref(new Date().toISOString().split('T')[0])
const tempoPaymentMethod = ref('CASH')
const tempoNextDueDate = ref('')
const tempoTransferProofURL = ref('')
const tempoBankAccountId = ref('')
const tempoBankAccounts = ref<any[]>([])
const fetchTempoBankAccounts = async () => {
  try {
    const res = await apiClient.get('/finance/bank-accounts')
    tempoBankAccounts.value = (res.data?.data || []).filter((bank: any) => bank.is_active)
  } catch {
    tempoBankAccounts.value = []
    toast.error('Gagal mengambil daftar rekening bank')
  }
}
const tempoNotes = ref('')
const isUploadingTempoProof = ref(false)
const isSubmittingPayment = ref(false)

const getTempoProgressInfo = (saleOrRaw: any) => {
  const raw = saleOrRaw?.raw || saleOrRaw || {}
  const payStatus = String(raw.payment_status || saleOrRaw?.paymentStatus || '').toUpperCase()
  const totalAmount = Number(raw.total_amount || saleOrRaw?.amount || 0)
  const paidAmount = Number(raw.paid_amount || 0)
  const remainingAmount = Math.max(0, totalAmount - paidAmount)
  const isPaid = payStatus === 'PAID' || (totalAmount > 0 && remainingAmount <= 0 && payStatus !== 'PARTIAL')
  const percent = totalAmount > 0 ? Math.min(100, Math.round((paidAmount / totalAmount) * 100)) : (isPaid ? 100 : 0)

  // Payments list
  const payments = Array.isArray(raw.payments) ? raw.payments : []
  const successfulPayments = payments.filter((p: any) => p.payment_status === 'SUCCESS')
  const pendingPayments = payments.filter((p: any) => p.payment_status === 'PENDING')
  const totalPaidLogsCount = payments.filter((p: any) => p.payment_status !== 'FAILED').length

  // Installments list
  const installments = Array.isArray(raw.installments) ? raw.installments : []
  const paidInstallments = installments.filter((i: any) => i.status === 'PAID')

  // Total tenor target (jumlah termin cicilan yang direncanakan toko)
  // Normalisasi cerdas: jika raw.tenor >= 7 dan installments kosong, itu data lama jumlah hari sehingga dianggap 1x
  let targetTenor = 1
  if (installments.length > 0) {
    targetTenor = installments.length
  } else if (typeof raw.tenor === 'number' && raw.tenor > 0 && raw.tenor < 7) {
    targetTenor = raw.tenor
  } else if (raw.tenor && !isNaN(Number(raw.tenor)) && Number(raw.tenor) > 0 && Number(raw.tenor) < 7) {
    targetTenor = Number(raw.tenor)
  }

  const paidTermsCount = paidInstallments.length > 0 ? paidInstallments.length : successfulPayments.length
  const remainingTermsCount = Math.max(0, targetTenor - paidTermsCount)
  const isExceedingPlannedTenor = paidTermsCount >= targetTenor && !isPaid && remainingAmount > 0

  // Check Overdue
  const dueDate = raw.payment_due_date ? new Date(raw.payment_due_date) : null
  const isOverdue = dueDate ? dueDate < new Date() && !isPaid : false

  // Detect payment method used for tempo (Cash or Transfer)
  const hasCash = raw.dp_payment_method === 'CASH' || payments.some((p: any) => p.payment_method === 'CASH')
  const hasTransfer = raw.dp_payment_method === 'TRANSFER' || raw.dp_payment_method === 'VA' || payments.some((p: any) => p.payment_method === 'TRANSFER' || p.payment_method === 'VA')
  
  let tempoMethodLabel = 'TEMPO'
  if (hasCash && hasTransfer) {
    tempoMethodLabel = 'TEMPO (CASH & TF)'
  } else if (hasTransfer) {
    tempoMethodLabel = 'TEMPO (TRANSFER)'
  } else if (hasCash || raw.payment_method === 'TEMPO') {
    tempoMethodLabel = 'TEMPO (CASH)'
  } else if (payStatus === 'PARTIAL') {
    tempoMethodLabel = raw.payment_method ? `${raw.payment_method} (PARTIAL)` : 'PARTIAL'
  }

  const nextPaymentNumber = totalPaidLogsCount + 1
  const nextTermNumber = paidTermsCount + 1

  // Determine current step / term description dengan Fleksibilitas Penuh
  let stepText = ''
  let detailedStepText = ''
  const maxTenorText = `Maks. ${targetTenor}x Cicilan`

  if (isPaid) {
    stepText = paidTermsCount > targetTenor 
      ? `✓ LUNAS (${paidTermsCount}x Fleksibel)`
      : `✓ LUNAS (Maks. ${targetTenor}x)`
    detailedStepText = `Lunas (${paidTermsCount}x dari Maks. ${targetTenor}x Cicilan)`
  } else if (isExceedingPlannedTenor) {
    stepText = `Cicilan ${nextTermNumber} (Fleksibel • Target ${targetTenor}x)`
    detailedStepText = `Cicilan ke-${nextTermNumber} (Fleksibel, target awal: Maks. ${targetTenor}x Cicilan)`
  } else if (targetTenor > 1) {
    if (paidTermsCount === 0) {
      stepText = `Maks. ${targetTenor}x (0%)`
      detailedStepText = `Belum bayar (Target: ${targetTenor}x cicilan)`
    } else {
      stepText = `Cicilan ${nextTermNumber} / Maks. ${targetTenor}x`
      detailedStepText = `Cicilan ke-${nextTermNumber} dari Maks. ${targetTenor}x (Sisa: ${remainingTermsCount}x lagi)`
    }
  } else {
    if (paidAmount > 0 || payStatus === 'PARTIAL') {
      stepText = `Partial (${percent}%) • Cicilan ${nextPaymentNumber}`
      detailedStepText = `Pembayaran sebagian (${percent}%) • Cicilan ke-${nextPaymentNumber} (Fleksibel)`
    } else {
      stepText = 'Maks. 1x Bayar'
      detailedStepText = 'Pelunasan 1x (Jatuh Tempo)'
    }
  }

  const estimatedTermAmount = targetTenor > 1 && remainingTermsCount > 0
    ? Math.round(remainingAmount / remainingTermsCount)
    : remainingAmount

  return {
    totalAmount,
    paidAmount,
    remainingAmount,
    percent,
    isOverdue,
    dueDate,
    payments,
    successfulPayments,
    pendingPayments,
    totalPaidLogsCount,
    nextPaymentNumber,
    nextTermNumber,
    targetTenor,
    maxTenorText,
    paidTermsCount,
    remainingTermsCount,
    isExceedingPlannedTenor,
    estimatedTermAmount,
    installments,
    tempoMethodLabel,
    stepText,
    detailedStepText,
    isPaid,
  }
}

const openTempoPaymentDialog = () => {
  if (selectedNota.value) {
    const info = getTempoProgressInfo(selectedNota.value)
    // Default to 1 term amount if multi-term tempo, or full remaining
    tempoPaymentAmount.value = info.targetTenor > 1 && info.estimatedTermAmount > 0
      ? info.estimatedTermAmount
      : (info.remainingAmount > 0 ? info.remainingAmount : info.totalAmount)
    tempoPaymentDate.value = new Date().toISOString().split('T')[0]
    tempoPaymentMethod.value = 'CASH'
    tempoBankAccountId.value = ''
    fetchTempoBankAccounts()
    tempoTransferProofURL.value = ''
    tempoNotes.value = ''
    const raw = selectedNota.value.raw || {}
    tempoNextDueDate.value = raw.payment_due_date ? String(raw.payment_due_date).split('T')[0] || '' : ''
    showTempoPaymentModal.value = true
  }
}

const isTempoEligibleForPayment = (sale: any) => {
  if (!sale || sale.status === 'DELETED' || sale.status === 'CANCELLED' || sale.deletedAt) return false
  const raw = sale.raw || sale
  const payStatus = String(raw.payment_status || sale.paymentStatus || '').toUpperCase()
  const method = String(raw.payment_method || sale.paymentMethod || '').toUpperCase()

  if (method !== 'TEMPO') return false
  // Jika sudah PAID maka tidak perlu pembayaran lagi
  if (payStatus === 'PAID') return false

  const isPartialOrUnpaid = payStatus === 'PARTIAL' || payStatus === 'UNPAID' || payStatus === 'OVERDUE'
  const info = getTempoProgressInfo(sale)

  return isPartialOrUnpaid && info.remainingAmount > 0 && !info.isPaid
}

const openTempoPaymentDialogForSale = (sale: any) => {
  selectedNota.value = sale
  openTempoPaymentDialog()
}

const handleTempoProofUpload = async (event: any) => {
  const file = event.target.files?.[0]
  if (!file) return

  const formData = new FormData()
  formData.append('file', file)
  isUploadingTempoProof.value = true

  try {
    const res = await apiClient.post('/admin/sales/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    if (res.data?.url) {
      tempoTransferProofURL.value = res.data.url
      toast.success('Bukti transfer berhasil diunggah')
    }
  } catch (error) {
    console.error('Upload tempo transfer proof failed:', error)
    toast.error('Gagal mengunggah bukti transfer')
  } finally {
    isUploadingTempoProof.value = false
  }
}

const submitTempoPayment = async () => {
  if (!selectedNota.value || !selectedNota.value.id) return
  if (tempoPaymentAmount.value <= 0) {
    toast.error('Jumlah pembayaran harus lebih dari 0')
    return
  }
  if (tempoPaymentMethod.value === 'TRANSFER' && !tempoBankAccountId.value) {
    toast.error('Pilih rekening bank penerima transfer')
    return
  }

  isSubmittingPayment.value = true
  try {
    const info = getTempoProgressInfo(selectedNota.value)
    const termInfoText = info.targetTenor > 1 
      ? `Cicilan ke-${info.nextTermNumber} dari ${info.targetTenor}x` 
      : `Pembayaran ke-${info.nextPaymentNumber}`
    const payload: any = {
      amount: Number(tempoPaymentAmount.value),
      payment_method: tempoPaymentMethod.value,
      ...(tempoPaymentMethod.value === 'TRANSFER' ? { company_bank_account_id: tempoBankAccountId.value } : {}),
      notes: tempoNotes.value || `${termInfoText} via Web Admin (${tempoPaymentMethod.value})`,
    }
    if (tempoPaymentDate.value) {
      payload.payment_date = new Date(`${tempoPaymentDate.value}T12:00:00Z`).toISOString()
    }
    if (tempoTransferProofURL.value) {
      payload.transfer_proof_url = tempoTransferProofURL.value
    }
    if (tempoNextDueDate.value && Number(tempoPaymentAmount.value) < info.remainingAmount) {
      payload.next_due_date = new Date(tempoNextDueDate.value).toISOString()
    }

    // Call admin payment endpoint
    await apiClient.post(`/admin/sales/transactions/${selectedNota.value.id}/payments`, payload)

    toast.success(`Pembayaran ${termInfoText} berhasil dicatat. Menunggu verifikasi kasir.`)
    showTempoPaymentModal.value = false

    // Refresh main list
    await fetchData()
    // Refresh selectedNota if still open
    const updated = salesData.value.find((s: any) => s.id === selectedNota.value.id)
    if (updated) {
      selectedNota.value = updated
    }
  } catch (err: any) {
    console.error('Failed to submit tempo payment:', err)
    toast.error(err.response?.data?.error || 'Gagal mencatat pembayaran')
  } finally {
    isSubmittingPayment.value = false
  }
}

const route = useRoute()
const router = useRouter()

const openSalesmanLPS = (sale?: any) => {
  const empId = sale?.raw?.employee_id || sale?.employeeID || ''
  const dateVal = sale?.raw?.transaction_date ? String(sale.raw.transaction_date).split('T')[0] : (selectedDate.value || new Date().toISOString().split('T')[0])
  const query: any = { date: dateVal }
  if (empId) query.salesman_id = empId
  router.push({ path: '/sales/lps', query })
}

const parsedNotaNotes = computed(() => {
  if (!selectedNota.value?.raw?.notes) return null
  try {
    return JSON.parse(selectedNota.value.raw.notes)
  } catch (e) {
    return null
  }
})

const notaTotalReturn = computed(() => {
  if (!parsedNotaNotes.value?.return_items) return 0
  return parsedNotaNotes.value.return_items.reduce((sum: number, item: any) => sum + (item.quantity * (item.price || 0)), 0)
})

const notaDiscountAmount = computed(() => {
  return Number(selectedNota.value?.raw?.discount_amount || 0)
})

const notaBaseTotal = computed(() => {
  if (!selectedNota.value) return 0
  return selectedNota.value.amount + notaTotalReturn.value + notaDiscountAmount.value
})

const currentSale = ref({
  id: '',
  companyID: '',
  employeeID: '',
  storeID: '',
  storeCategory: 'TOKO_LAMA',
  totalAmount: 0,
  transactionDate: new Date().toISOString().split('T')[0],
  receiptNo: '',
  receiptImageURL: '',
  transferProofURL: '',
  status: 'VERIFIED',
  paymentStatus: 'UNPAID',
  paymentMethod: 'CASH',
  paidAmount: 0,
  paymentDueDate: '',
  tempoInstallmentCount: 1 as number | 'custom',
  tempoCustomInstallmentCount: 3,
  tempoTenorDays: 14 as number | 'custom',
  discountAmount: 0,
  globalDiscountPercent: 0,
  globalDiscountType: 'NOMINAL' as 'NOMINAL' | 'PERCENT',
  notes: '',
  items: [] as any[],
  return_items: [] as any[],
})

// Subtotal kotor sebelum diskon item
const crudItemsGrossTotal = computed(() => {
  if (!currentSale.value.items || currentSale.value.items.length === 0) return 0
  return currentSale.value.items.reduce((acc: number, item: any) => {
    return acc + (Number(item.quantity || 0) * Number(item.price || 0))
  }, 0)
})

// Total potongan dari diskon per produk
const crudTotalItemDiscounts = computed(() => {
  if (!currentSale.value.items || currentSale.value.items.length === 0) return 0
  return currentSale.value.items.reduce((acc: number, item: any) => {
    const gross = Number(item.quantity || 0) * Number(item.price || 0)
    const discPercent = Number(item.discount_percent || 0)
    return acc + (gross * discPercent / 100)
  }, 0)
})

// Subtotal setelah diskon per produk
const crudItemsNetTotal = computed(() => {
  return Math.max(0, crudItemsGrossTotal.value - crudTotalItemDiscounts.value)
})

// Total nominal retur sampingan pada form
const crudTotalReturnAmount = computed(() => {
  if (!currentSale.value.return_items || currentSale.value.return_items.length === 0) return 0
  return currentSale.value.return_items.reduce((acc: number, item: any) => {
    return acc + (Number(item.quantity || 0) * Number(item.price || 0))
  }, 0)
})

// Helper untuk sinkronisasi diskon per-item antara persen dan nominal
const updateItemDiscountFromNominal = (index: number, nominalVal: number) => {
  const item = currentSale.value.items[index]
  if (!item) return
  const gross = Number(item.quantity || 0) * Number(item.price || 0)
  if (gross <= 0) {
    item.discount_percent = 0
    return
  }
  const pct = (Number(nominalVal || 0) / gross) * 100
  item.discount_percent = Math.min(100, Math.max(0, Math.round(pct * 100) / 100))
}

const getItemDiscountNominal = (item: any) => {
  const gross = Number(item.quantity || 0) * Number(item.price || 0)
  const pct = Number(item.discount_percent || 0)
  return Math.round((gross * pct) / 100)
}

const isUploadingTransferProof = ref(false)

const estimatedTempoPerInstallment = computed(() => {
  const tot = Number(currentSale.value.totalAmount || 0)
  const dp = Number(currentSale.value.paidAmount || 0)
  const remaining = Math.max(0, tot - dp)
  const count = currentSale.value.tempoInstallmentCount === 'custom'
    ? Number(currentSale.value.tempoCustomInstallmentCount || 1)
    : Number(currentSale.value.tempoInstallmentCount || 1)
  if (count <= 1) return remaining
  return Math.round(remaining / count)
})

const selectedNotaTransferProof = computed(() => {
  if (!selectedNota.value?.raw) return ''
  const payments = selectedNota.value.raw.payments
  if (payments && Array.isArray(payments)) {
    const pWithProof = payments.find((p: any) => p.transfer_proof_url)
    if (pWithProof) return pWithProof.transfer_proof_url
  }
  return selectedNota.value.raw.transfer_proof_url || ''
})

const calculateDueDateForSale = (days: number): string => {
  const dateStr: string = (currentSale.value.transactionDate && currentSale.value.transactionDate.length > 0)
    ? currentSale.value.transactionDate
    : new Date().toISOString().split('T')[0]!
  const base = new Date(`${dateStr}T00:00:00Z`)
  const count = currentSale.value.tempoInstallmentCount === 'custom'
    ? Number(currentSale.value.tempoCustomInstallmentCount || 1)
    : Number(currentSale.value.tempoInstallmentCount || 1)
  base.setUTCDate(base.getUTCDate() + days * Math.max(1, count))
  return base.toISOString().split('T')[0]!
}

const setCurrentSaleTenorDays = (days: number | 'custom') => {
  currentSale.value.tempoTenorDays = days
  if (days !== 'custom') {
    currentSale.value.paymentDueDate = calculateDueDateForSale(days)
  }
}

const setCurrentSaleInstallmentCount = (count: number | 'custom') => {
  currentSale.value.tempoInstallmentCount = count
  if (typeof currentSale.value.tempoTenorDays === 'number') {
    currentSale.value.paymentDueDate = calculateDueDateForSale(currentSale.value.tempoTenorDays)
  }
}

const inferTempoTenorDays = (transactionDate: string, dueDate: string, count: number): number | 'custom' => {
  if (!transactionDate || !dueDate) return 'custom'
  const totalDays = Math.round((Date.parse(`${dueDate}T00:00:00Z`) - Date.parse(`${transactionDate}T00:00:00Z`)) / 86400000)
  const daysPerInstallment = totalDays / Math.max(1, count)
  return [7, 14, 30, 45].includes(daysPerInstallment) ? daysPerInstallment : 'custom'
}

const syncCurrentSalePaymentStatus = () => {
  if (currentSale.value.paymentMethod === 'TEMPO' || isPaidInvoiceEdit.value) {
    const dp = Number(currentSale.value.paidAmount || 0)
    const tot = Number(currentSale.value.totalAmount || 0)
    if (dp > 0 && dp < tot) {
      currentSale.value.paymentStatus = 'PARTIAL'
    } else if (dp >= tot && tot > 0) {
      currentSale.value.paymentStatus = 'PAID'
    } else {
      currentSale.value.paymentStatus = 'UNPAID'
    }
  }
}

watch(() => currentSale.value.transactionDate, () => {
  if (currentSale.value.paymentMethod === 'TEMPO' && typeof currentSale.value.tempoTenorDays === 'number') {
    currentSale.value.paymentDueDate = calculateDueDateForSale(currentSale.value.tempoTenorDays)
  }
})

watch(() => currentSale.value.paymentMethod, (newMethod) => {
  if (isPaidInvoiceEdit.value) {
    syncCurrentSalePaymentStatus()
    return
  }
  if (newMethod === 'TEMPO') {
    if (!currentSale.value.paymentDueDate) {
      setCurrentSaleTenorDays(14)
    }
    syncCurrentSalePaymentStatus()
  } else {
    currentSale.value.paymentStatus = 'PAID'
    currentSale.value.paidAmount = currentSale.value.totalAmount
  }
})

watch(() => currentSale.value.paidAmount, () => {
  syncCurrentSalePaymentStatus()
})

watch(() => currentSale.value.totalAmount, () => {
  syncCurrentSalePaymentStatus()
})

watch(
  () => currentSale.value?.employeeID,
  async (newEmpId) => {
    if (newEmpId) {
      await fetchStoresForSalesman(newEmpId)
    } else {
      salesmanStores.value = []
    }
  },
)

const parsedCurrentNotes = computed(() => {
  if (!currentSale.value.notes) return null
  try {
    return JSON.parse(currentSale.value.notes)
  } catch (e) {
    return null
  }
})

const factoryStore = useFactoryStore()

const returnableProducts = computed(() => {
  return filterReturnableProducts(factoryStore.products)
})

const handlePrint = () => {
  window.print()
}

/**
 * Resolve a relative signature path to a full absolute URL.
 * data: URIs and http(s) URLs are returned as-is.
 */
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

/**
 * signature_image_url format from mobile: "courier_sig_url,customer_sig_url"
 * If only one URL is present (no comma), it is the customer (pelanggan/toko) signature.
 */
const splitNotaSignatures = (nota: any): [string, string] => {
  const raw = nota?.raw?.signature_image_url || ''
  if (!raw) return ['', '']

  if (raw.includes('data:image/')) {
    // Two base64 data URIs joined without a separator — split on the second data:image/
    const parts = raw.split('data:image/').filter((p: string) => p.trim())
    const courier = parts[0] ? 'data:image/' + parts[0].replace(/,$/, '').trimEnd() : ''
    const customer = parts[1] ? 'data:image/' + parts[1].replace(/,$/, '').trimEnd() : ''
    return [courier, customer]
  }

  const parts = raw.split(',').map((s: string) => s.trim())
  if (parts.length === 1 && parts[0]) {
    // Single URL: treat as customer-only signature (e.g. POD / delivery scenario)
    return ['', parts[0]]
  }
  return [parts[0] || '', parts[1] || '']
}

const getCourierSignature = (nota: any): string => {
  return getFullImageUrl(splitNotaSignatures(nota)[0])
}

const getCourierTitle = (nota: any): string => {
  if (nota?.raw?.delivery_items && nota.raw.delivery_items.length > 0) {
    return 'DRIVER'
  }
  return 'SALESMAN / MOTORIS'
}

const getCourierName = (nota: any): string => {
  if (nota?.raw?.delivery_items && nota.raw.delivery_items.length > 0) {
    const driver = nota.raw.delivery_items[0].delivery_batch?.driver
    if (driver) {
      return `${driver.first_name || ''} ${driver.last_name || ''}`.trim()
    }
    return 'Driver'
  }
  if (nota?.raw?.employee) {
    const emp = nota.raw.employee
    if (emp.name) return emp.name
    return `${emp.first_name || ''} ${emp.last_name || ''}`.trim()
  }
  return nota?.salesman || 'Sales Motoris'
}

const rejectedPayments = computed(() => {
  if (!selectedNota.value?.raw?.payments) return []
  return selectedNota.value.raw.payments.filter((p: any) => p.payment_status === 'FAILED')
})

const getCustomerSignature = (nota: any): string => {
  return getFullImageUrl(splitNotaSignatures(nota)[1])
}

const openImagePreview = (url: string) => {
  previewImageUrl.value = url
  zoomLevel.value = 1
  rotationAngle.value = 0
  showImageModal.value = true
}

const handleZoom = (val: number) => {
  zoomLevel.value = Math.max(0.5, Math.min(3, zoomLevel.value + val))
}

const fetchStores = async () => {
  try {
    const params: Record<string, any> = {
      status: 'ALL',
      minimal: true,
      limit: 5000,
    }
    if (masterDataStore.selectedBranchId && masterDataStore.selectedBranchId !== 'ALL_BRANCHES') {
      params.branch_id = masterDataStore.selectedBranchId
    }
    const companyId = selectedPT?.value || masterDataStore.selectedBranchCompanyId
    if (companyId) {
      params.company_id = companyId
    }
    const res = await apiClient.get('/stores', { params })
    if (res.data?.data) {
      allStores.value = res.data.data
    }
  } catch (error) {
    console.error('Failed to fetch stores:', error)
  }
}

const fetchStoresForSalesman = async (employeeId?: string) => {
  if (!employeeId) {
    salesmanStores.value = []
    return
  }
  isLoadingSalesmanStores.value = true
  try {
    // Toko salesman dimuat menyeluruh tanpa dibatasi filter cabang atau company di header
    // branch_id: 'ALL_BRANCHES' untuk melewati auto-scoping cabang di backend
    const params: Record<string, any> = {
      assigned_employee_id: employeeId,
      branch_id: 'ALL_BRANCHES',
      status: 'ALL',
      minimal: true,
      limit: 5000,
    }
    const res = await apiClient.get('/stores', { params })
    salesmanStores.value = res.data?.data || []
  } catch (err) {
    console.error('Failed to fetch stores for salesman:', err)
  } finally {
    isLoadingSalesmanStores.value = false
  }
}

let remoteSearchTimeout: any = null
const searchStoresRemotely = (query: string) => {
  const trimmed = query.trim()
  if (!trimmed || trimmed.length < 2) return
  if (remoteSearchTimeout) clearTimeout(remoteSearchTimeout)
  remoteSearchTimeout = setTimeout(async () => {
    try {
      const res = await apiClient.get('/stores', {
        params: {
          search: trimmed,
          status: 'ALL',
          branch_id: 'ALL_BRANCHES',
          minimal: true,
          limit: 50,
        },
      })
      const found = res.data?.data || []
      found.forEach((st: any) => {
        if (st?.id && !allStores.value.some((s: any) => s.id === st.id)) {
          allStores.value.push(st)
        }
        if (
          currentSale.value?.employeeID &&
          (st.assigned_employee_id === currentSale.value.employeeID ||
            st.created_by_employee_id === currentSale.value.employeeID)
        ) {
          if (!salesmanStores.value.some((s: any) => s.id === st.id)) {
            salesmanStores.value.push(st)
          }
        }
      })
    } catch (_) {}
  }, 300)
}

const fetchSalesmen = async (force = false) => {
  if (!force && salesmen.value.length > 0) return
  try {
    const res = await apiClient.get('/employees?limit=all') // Assuming all employees can be salesmen or filter by role
    if (res.data?.data) {
      salesmen.value = res.data.data
    }
  } catch (error) {
    console.error('Failed to fetch salesmen:', error)
  }
}

const currentPage = ref(1)
const itemsPerPage = ref<number | 'all'>(10)
const totalItems = ref(0)
const serverTotalAmount = ref(0)
const totalPages = computed(() => {
  if (itemsPerPage.value === 'all') return 1
  return Math.ceil(totalItems.value / (itemsPerPage.value as number))
})

// Instant invoices carry item prices and discounts. Some older records have a
// total_amount that predates the item discount, so use the item breakdown for
// their displayed omzet while keeping the stored value available in `raw`.
const invoiceAmount = (trx: any): number => {
  const stored = Number(trx.total_amount || 0)
  if (!String(trx.receipt_no || '').startsWith('INV-INSTANT-')) return stored
  if (!Array.isArray(trx.items) || trx.items.length === 0) return stored
  if (Number(trx.discount_amount || 0) !== 0 || trx.returns?.length) return stored

  const calculated = trx.items.reduce((sum: number, item: any) =>
    sum + Number(item.quantity || 0) * Number(item.price_at_transaction || 0) *
      (1 - Number(item.discount_percent || 0) / 100), 0)
  return Number.isFinite(calculated) ? Math.round(calculated) : stored
}

const fetchData = async () => {
  if (!masterDataStore.selectedBranchId) return
  isLoading.value = true
  const companyId = selectedPT.value || masterDataStore.selectedBranchCompanyId || ''

  const params = new URLSearchParams()
  params.append('page', currentPage.value.toString())
  params.append('limit', itemsPerPage.value.toString())
  
  const hasBranch = masterDataStore.selectedBranchId && masterDataStore.selectedBranchId !== 'ALL_BRANCHES'
  if (!hasBranch && companyId) {
    params.append('company_id', companyId)
  }
  if (hasBranch) {
    params.append('branch_id', masterDataStore.selectedBranchId)
  }
  if (selectedInvoiceType.value && selectedInvoiceType.value !== 'ALL') {
    if (selectedInvoiceType.value === 'DIRECT_SALE') {
      params.append('sales_type', 'direct_sale')
    } else if (selectedInvoiceType.value === 'REGULAR_SALES') {
      params.append('sales_type', 'regular_sales')
    }
  } else if (masterDataStore.selectedSalesType) {
    params.append('sales_type', masterDataStore.selectedSalesType)
  }
  if (selectedTargetEntity.value) {
    params.append('target_entity', selectedTargetEntity.value)
  }
  if (selectedStatus.value && selectedStatus.value !== 'ALL') {
    if (selectedStatus.value === 'TEMPO_UNPAID') {
      params.append('status', 'TEMPO_UNPAID')
    } else {
      const apiStatus = selectedStatus.value === 'APPROVED' ? 'VERIFIED' : (selectedStatus.value === 'DELETED' ? 'DELETED' : selectedStatus.value)
      params.append('status', apiStatus)
    }
  }
  if (searchQuery.value) params.append('search', searchQuery.value.trim())
  if (selectedPT.value) params.append('company_id', selectedPT.value)
  if (selectedSalesman.value) params.append('salesman', selectedSalesman.value)
  if (selectedMonth.value && selectedMonth.value !== 'ALL') params.append('month', selectedMonth.value)
  if (selectedYear.value && selectedYear.value !== 'ALL') params.append('year', selectedYear.value)
  if (selectedDate.value) params.append('date', selectedDate.value)
  if (selectedPaymentMethod.value && selectedPaymentMethod.value !== 'ALL') params.append('payment_method', selectedPaymentMethod.value)
  if (selectedPaymentStatus.value && selectedPaymentStatus.value !== 'ALL') params.append('payment_status', selectedPaymentStatus.value)

  try {
    const res = await apiClient.get(`/admin/sales/transactions/all?${params.toString()}`)
    if (res.data?.data) {
      salesData.value = res.data.data.map((trx: any) => ({
        id: trx.id,
        employeeID: trx.employee_id,
        storeID: trx.store_id,
        storeCategory: trx.store_category,
        salesman: trx.sales_channel === 'DIRECT_FACTORY' || trx.sales_channel === 'DIRECT_BRANCH' || trx.sales_channel === 'EMPLOYEE_SALE'
          ? (trx.cashier_employee ? `Kasir: ${trx.cashier_employee.first_name} ${trx.cashier_employee.last_name}` : 'Kasir Pabrik')
          : (trx.employee ? `${trx.employee.first_name} ${trx.employee.last_name}` : 'Unknown'),
        store: trx.buyer_type === 'EMPLOYEE' && trx.buyer_employee
          ? `${trx.buyer_employee.first_name} ${trx.buyer_employee.last_name} (Karyawan)`
          : (trx.buyer_name || trx.store?.name || (trx.sales_channel === 'DIRECT_FACTORY' ? 'Pelanggan Pabrik' : 'Toko Manual')),
        amount: invoiceAmount(trx),
        date: formatDate(trx.transaction_date),
        status: trx.deleted_at ? 'DELETED' : (trx.status === 'VERIFIED' ? 'APPROVED' : trx.status),
        deletedAt: trx.deleted_at,
        deletedBy: trx.deleted_by,
        deletionReason: trx.deletion_reason,
        receiptNo: trx.receipt_no,
        items:
          trx.items?.map((i: any) => `${i.product?.name} (${i.quantity})`) ||
          (trx.notes ? [trx.notes] : []),
        rawItems: trx.items || [],
        rawDate: trx.transaction_date,
        notaImage:
          trx.receipt_image_url || 'https://placehold.co/600x400/f1f5f9/64748b?text=Foto+Nota',
        deliveryOrders:
          trx.delivery_items?.map((di: any) => ({
            ...di.delivery_batch,
            id: di.delivery_batch?.id || di.delivery_batch_id,
            no: di.delivery_batch?.delivery_order_no,
            status: di.delivery_batch?.status,
            date: di.delivery_batch?.assigned_at,
          })) || [],
        returnsTotal:
          trx.returns
            ?.filter((r: any) => r.return_type === 'POTONG_NOTA')
            .reduce((sum: number, r: any) => sum + (Number(r.total_amount) || 0), 0) || 0,
        returns: trx.returns || [],
        raw: trx,
      }))
      totalItems.value = res.data.total || 0
      serverTotalAmount.value = res.data.total_amount || 0
    }
  } catch (error) {
    console.error('Failed to fetch sales data:', error)
  } finally {
    isLoading.value = false
  }
}

const handleExportXlsx = () => {
  const params = new URLSearchParams()
  if (masterDataStore.selectedBranchId && masterDataStore.selectedBranchId !== 'ALL_BRANCHES') {
    params.append('branch_id', masterDataStore.selectedBranchId)
  }
  if (selectedStatus.value !== 'ALL') {
    if (selectedStatus.value === 'TEMPO_UNPAID') {
      params.append('status', 'TEMPO_UNPAID')
    } else {
      const apiStatus = selectedStatus.value === 'APPROVED' ? 'VERIFIED' : selectedStatus.value
      params.append('status', apiStatus)
    }
  }
  if (searchQuery.value) params.append('search', searchQuery.value)
  if (selectedPT.value) params.append('company_id', selectedPT.value)
  if (selectedSalesman.value) params.append('salesman', selectedSalesman.value)
  if (selectedMonth.value && selectedMonth.value !== 'ALL') params.append('month', selectedMonth.value)
  if (selectedYear.value && selectedYear.value !== 'ALL') params.append('year', selectedYear.value)
  if (selectedDate.value) params.append('date', selectedDate.value)
  if (selectedInvoiceType.value && selectedInvoiceType.value !== 'ALL') {
    if (selectedInvoiceType.value === 'DIRECT_SALE') {
      params.append('sales_type', 'direct_sale')
    } else if (selectedInvoiceType.value === 'REGULAR_SALES') {
      params.append('sales_type', 'regular_sales')
    }
  } else if (masterDataStore.selectedSalesType) {
    params.append('sales_type', masterDataStore.selectedSalesType)
  }
  if (selectedTargetEntity.value) {
    params.append('target_entity', selectedTargetEntity.value)
  }
  if (selectedPaymentMethod.value && selectedPaymentMethod.value !== 'ALL') params.append('payment_method', selectedPaymentMethod.value)
  if (selectedPaymentStatus.value && selectedPaymentStatus.value !== 'ALL') params.append('payment_status', selectedPaymentStatus.value)

  apiClient.get('/admin/sales/transactions/export', {
    params: Object.fromEntries(params),
    responseType: 'blob'
  })
    .then((res) => {
      const blob = new Blob([res.data])
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.style.display = 'none'
      a.href = url
      a.download = 'faktur_penjualan.xlsx'
      document.body.appendChild(a)
      a.click()
      window.URL.revokeObjectURL(url)
    })
    .catch((err) => {
      toast.error('Gagal export data excel')
    })
}

const openAddModal = () => {
  isEditing.value = false
  mobileCrudTab.value = 'form'
  crudZoomLevel.value = 1
  crudRotationAngle.value = 0
  salesmanStores.value = []
  if (allStores.value.length === 0) {
    fetchStores()
  }
  currentSale.value = {
    id: '',
    companyID: '',
    employeeID: '',
    storeID: '',
    storeCategory: 'TOKO_LAMA',
    totalAmount: 0,
    transactionDate: new Date().toISOString().split('T')[0],
    receiptNo: '',
    receiptImageURL: '',
    transferProofURL: '',
    status: 'PENDING',
    paymentStatus: 'UNPAID',
    paymentMethod: 'CASH',
    paidAmount: 0,
    paymentDueDate: '',
    tempoInstallmentCount: 1,
    tempoCustomInstallmentCount: 3,
    tempoTenorDays: 14,
    discountAmount: 0,
    globalDiscountPercent: 0,
    globalDiscountType: 'NOMINAL',
    notes: '',
    items: [],
    return_items: [],
  }
  showModal.value = true
}

const handleEdit = (sale: any) => {
  isEditing.value = true
  mobileCrudTab.value = 'form'
  crudZoomLevel.value = 1
  crudRotationAngle.value = 0
  if (allStores.value.length === 0) {
    fetchStores()
  }
  // Find the raw data from salesData if needed or just use what we have
  const rawSale = salesData.value.find((s) => s.id === sale.id)
  const isDelivery = (rawSale?.deliveryOrders && rawSale.deliveryOrders.length > 0) || (rawSale?.raw?.delivery_items && rawSale.raw.delivery_items.length > 0)
  
  if (rawSale?.employeeID || rawSale?.raw?.employee_id) {
    fetchStoresForSalesman(rawSale.employeeID || rawSale.raw?.employee_id)
  }
  
  // Ensure the transaction's store/customer is in the dropdown options
  if (rawSale?.raw?.store_id) {
    const storeExists = allStores.value.some((st) => st.id === rawSale.raw.store_id)
    if (!storeExists) {
      if (rawSale.raw.store) {
        allStores.value.push(rawSale.raw.store)
      } else {
        allStores.value.push({
          id: rawSale.raw.store_id,
          name: rawSale.store || 'Toko Manual',
          address: '',
        })
      }
    }
  }

  // Ensure products (e.g. injected bonus products) are in the dropdown options
  if (rawSale?.raw?.items) {
    rawSale.raw.items.forEach((item: any) => {
      if (item.product && item.product.id) {
        const productExists = factoryStore.products.some((p: any) => p.id === item.product.id)
        if (!productExists) {
          factoryStore.products.push(item.product)
        }
      }
    })
  }

  const payMethod = rawSale?.raw?.payment_method || 'CASH'
  const payDueDate = rawSale?.raw?.payment_due_date ? String(rawSale.raw.payment_due_date).split('T')[0] || '' : ''
  const saleDate = sale.rawDate ? new Date(sale.rawDate).toISOString().split('T')[0]! : new Date().toISOString().split('T')[0]!
  const paidAmt = Number(rawSale?.raw?.paid_amount || 0)
  const installmentCountVal = Number(rawSale?.raw?.tenor || rawSale?.raw?.installments?.length || 1)
  const existingTransferProof = rawSale?.raw?.payments?.find((p: any) => p.transfer_proof_url)?.transfer_proof_url
    || (rawSale?.raw?.payments && rawSale.raw.payments.length > 0 ? rawSale.raw.payments[rawSale.raw.payments.length - 1]?.transfer_proof_url : '')
    || rawSale?.raw?.transfer_proof_url
    || ''

  const mappedItems =
    sale.rawItems?.map((ri: any) => ({
      product_id: ri.product_id,
      quantity: ri.quantity,
      price: ri.price_at_transaction,
      discount_percent: ri.discount_percent || 0,
    })) || []

  const returnItems = (() => {
    try {
      const rawReturns = JSON.parse(rawSale?.raw?.notes || '{}').return_items || []
      if (!Array.isArray(rawReturns)) return []
      return rawReturns.map((ret: any) => {
        let resolvedProductId = ret.product_id || ''
        let resolvedName = ret.product_name || ''
        let resolvedUnit = (ret.unit || 'PCS').toUpperCase()

        if (resolvedProductId) {
          const foundById = factoryStore.products.find((p: any) => p.id === resolvedProductId)
          if (foundById) {
            resolvedName = foundById.name
            if (!ret.unit && foundById.unit) {
              resolvedUnit = foundById.unit.toUpperCase()
            }
          }
        } else if (resolvedName) {
          const normalizedTarget = resolvedName.trim().toLowerCase()
          const noSpaceTarget = normalizedTarget.replace(/\s+/g, '')
          const foundByName = factoryStore.products.find((p: any) => {
            const pName = (p.name || '').trim().toLowerCase()
            return pName === normalizedTarget || pName.replace(/\s+/g, '') === noSpaceTarget
          })
          if (foundByName) {
            resolvedProductId = foundByName.id
            resolvedName = foundByName.name
            if (!ret.unit && foundByName.unit) {
              resolvedUnit = foundByName.unit.toUpperCase()
            }
          }
        }

        return {
          product_id: resolvedProductId,
          product_name: resolvedName,
          quantity: Number(ret.quantity || 1),
          unit: resolvedUnit,
          price: Number(ret.price || 0),
        }
      })
    } catch (e) {
      return []
    }
  })()

  const rawSubtotalItems = mappedItems.reduce((acc: number, item: any) => {
    const qty = Number(item.quantity || 0)
    const prc = Number(item.price || 0)
    const disc = Number(item.discount_percent || 0)
    return acc + (qty * prc * (1 - disc / 100))
  }, 0)

  const rawReturnTotal = returnItems.reduce((acc: number, item: any) => {
    return acc + (Number(item.quantity || 0) * Number(item.price || 0))
  }, 0)

  // Inferred discount if discount_amount was not recorded in DB but items subtotal exceeds sale.amount
  let existingDiscount = Number(rawSale?.raw?.discount_amount || 0)
  if (existingDiscount <= 0 && rawSubtotalItems > (sale.amount + rawReturnTotal)) {
    existingDiscount = Math.round(rawSubtotalItems - rawReturnTotal - sale.amount)
  }

  currentSale.value = {
    id: sale.id,
    companyID:
      rawSale?.raw?.company_id ||
      rawSale?.raw?.store?.company_id ||
      rawSale?.raw?.employee?.company_id ||
      masterDataStore.selectedBranchCompanyId ||
      allStores.value.find((s: any) => s.id === (rawSale?.storeID || rawSale?.raw?.store_id))?.company_id ||
      '',
    employeeID: rawSale?.employeeID || '',
    storeID: rawSale.storeID || rawSale.raw?.store_id || '',
    storeCategory: rawSale.storeCategory || 'TOKO_LAMA',
    totalAmount: sale.amount,
    transactionDate: saleDate,
    receiptNo: sale.receiptNo,
    receiptImageURL:
      isDelivery || sale.notaImage === 'https://placehold.co/600x400/f1f5f9/64748b?text=Foto+Nota'
        ? ''
        : sale.notaImage,
    transferProofURL: existingTransferProof,
    status: sale.status === 'APPROVED' ? 'VERIFIED' : sale.status,
    paymentStatus: rawSale?.raw?.payment_status || 'UNPAID',
    paymentMethod: payMethod,
    paidAmount: paidAmt,
    paymentDueDate: payDueDate || '',
    tempoInstallmentCount: [1, 2, 3, 4].includes(installmentCountVal) ? installmentCountVal : 'custom',
    tempoCustomInstallmentCount: installmentCountVal > 0 ? installmentCountVal : 3,
    tempoTenorDays: inferTempoTenorDays(saleDate, payDueDate, installmentCountVal),
    discountAmount: existingDiscount,
    globalDiscountPercent: 0,
    globalDiscountType: 'NOMINAL',
    notes: rawSale?.raw?.notes || '',
    return_items: returnItems,
    items: mappedItems,
  }
  showModal.value = true
}

const hasFailedDelivery = (sale: any) => {
  if (!sale.raw?.delivery_items || sale.raw.delivery_items.length === 0) return false
  return sale.raw.delivery_items.some((item: any) => item.status === 'FAILED')
}

const failedDeliveriesCount = computed(() => {
  return salesData.value.filter(trx => hasFailedDelivery(trx) && trx.status === 'PENDING').length
})

// --- Failed delivery: return the same SO and invoice to the delivery queue ---
const showCancelDeliveryModal = ref(false)
const cancelDeliveryTransaction = ref<any>(null)
const cancelDeliveryReason = ref('')
const cancelDeliveryLoading = ref(false)
const canCancelFailedDelivery = computed(() => ['admin_nota', 'superadmin', 'manager'].includes(authStore.user?.role || ''))

const openCancelDeliveryModal = (sale: any) => {
  cancelDeliveryTransaction.value = sale
  cancelDeliveryReason.value = ''
  showCancelDeliveryModal.value = true
}

const confirmCancelDelivery = async () => {
  if (!cancelDeliveryTransaction.value) return
  if (!cancelDeliveryReason.value.trim()) {
    toast.error('Mohon masukkan alasan gagal kirim!')
    return
  }
  cancelDeliveryLoading.value = true
  try {
    await apiClient.post(`/admin/sales/transactions/${cancelDeliveryTransaction.value.id}/redeliver`, {
      reason: cancelDeliveryReason.value.trim(),
    })
    toast.success('Nota kembali ke Manajemen Pengiriman. Faktur akan tampil lagi setelah surat jalan baru dibuat.')
    showCancelDeliveryModal.value = false
    cancelDeliveryTransaction.value = null
    fetchData()
  } catch (error: any) {
    console.error('Gagal mengembalikan nota ke antrean pengiriman:', error)
    toast.error(error.response?.data?.error || 'Gagal mengembalikan nota ke antrean pengiriman')
  } finally {
    cancelDeliveryLoading.value = false
  }
}

const confirmFinalCancelDelivery = async () => {
  if (!cancelDeliveryTransaction.value || !cancelDeliveryReason.value.trim()) return
  cancelDeliveryLoading.value = true
  try {
    await apiClient.post(`/admin/sales/transactions/${cancelDeliveryTransaction.value.id}/cancel-failed-delivery`, {
      reason: cancelDeliveryReason.value.trim(),
    })
    toast.success('Faktur gagal kirim dibatalkan. Pengembalian barang tetap diproses oleh gudang.')
    showCancelDeliveryModal.value = false
    cancelDeliveryTransaction.value = null
    fetchData()
  } catch (error: any) {
    toast.error(error.response?.data?.error || 'Gagal membatalkan faktur')
  } finally {
    cancelDeliveryLoading.value = false
  }
}
// -----------------------------------

const handleDelete = (sale: any) => {
  deleteTargetSale.value = sale
  deleteReason.value = ''
  deleteConfirmText.value = ''
  showDeleteConfirmModal.value = true
}

const confirmDelete = async () => {
  if (!deleteTargetSale.value) return
  if (!deleteReason.value.trim()) {
    toast.error('Alasan penghapusan wajib diisi')
    return
  }
  if (deleteConfirmText.value.trim() !== deleteTargetSale.value.receiptNo) {
    toast.error('Nomor nota konfirmasi tidak cocok')
    return
  }
  isDeletingTrx.value = true
  try {
    await apiClient.delete(`/admin/sales/transactions/${deleteTargetSale.value.id}`, {
      data: { reason: deleteReason.value }
    })
    toast.success('Faktur penjualan berhasil dihapus (Soft Delete)')
    showDeleteConfirmModal.value = false
    deleteTargetSale.value = null
    fetchData()
  } catch (error: any) {
    toast.error(error.response?.data?.error || 'Gagal menghapus transaksi')
  } finally {
    isDeletingTrx.value = false
  }
}

const handleRestore = (sale: any) => {
  restoreTargetSale.value = sale
  showRestoreConfirmModal.value = true
}

const confirmRestore = async () => {
  if (!restoreTargetSale.value) return
  isRestoringTrx.value = true
  try {
    await apiClient.post(`/admin/sales/transactions/${restoreTargetSale.value.id}/restore`)
    toast.success('Faktur penjualan berhasil dipulihkan')
    showRestoreConfirmModal.value = false
    restoreTargetSale.value = null
    fetchData()
  } catch (error: any) {
    toast.error(error.response?.data?.error || 'Gagal memulihkan faktur')
  } finally {
    isRestoringTrx.value = false
  }
}

const handleFileUpload = async (event: any) => {
  const file = event.target.files[0]
  if (!file) return

  const formData = new FormData()
  formData.append('file', file)

  try {
    const res = await apiClient.post('/admin/sales/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    if (res.data?.url) {
      currentSale.value.receiptImageURL = res.data.url
      toast.success('Foto nota berhasil diunggah')
    }
  } catch (error) {
    console.error('Upload failed:', error)
    toast.error('Gagal mengunggah foto nota')
  }
}

const handleTransferProofUpload = async (event: any) => {
  const file = event.target.files?.[0]
  if (!file) return

  const formData = new FormData()
  formData.append('file', file)
  isUploadingTransferProof.value = true

  try {
    const res = await apiClient.post('/admin/sales/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    if (res.data?.url) {
      currentSale.value.transferProofURL = res.data.url
      toast.success('Bukti pembayaran berhasil diunggah')
    }
  } catch (error) {
    console.error('Upload transfer proof failed:', error)
    toast.error('Gagal mengunggah bukti pembayaran')
  } finally {
    isUploadingTransferProof.value = false
  }
}

const handleRemoveTransferProof = () => {
  currentSale.value.transferProofURL = ''
}

const handleSubmit = async () => {
  if (!currentSale.value.employeeID) {
    toast.error('Silakan pilih salesman terlebih dahulu!')
    return
  }

  if (currentSale.value.paymentMethod === 'TEMPO') {
    if (!currentSale.value.paymentDueDate) {
      toast.error('Silakan tentukan tanggal jatuh tempo untuk faktur cicilan/tempo!')
      return
    }
    const tDate = currentSale.value.transactionDate ? new Date(currentSale.value.transactionDate) : new Date()
    if (new Date(currentSale.value.paymentDueDate) < tDate) {
      toast.error('Tanggal jatuh tempo tidak boleh sebelum tanggal transaksi!')
      return
    }
  }

  if (Array.isArray(currentSale.value.return_items) && currentSale.value.return_items.length > 0) {
    for (let i = 0; i < currentSale.value.return_items.length; i++) {
      const r = currentSale.value.return_items[i]
      if (!r.product_id && !String(r.product_name || '').trim()) {
        toast.error(`Silakan pilih barang retur pada baris ke-${i + 1}!`)
        return
      }
      if (Number(r.quantity || 0) <= 0) {
        toast.error(`Jumlah barang retur pada baris ke-${i + 1} harus lebih dari 0!`)
        return
      }
    }
  }

  try {
    const normalizedReturnItems = (currentSale.value.return_items || [])
      .filter((r: any) => r.product_id || String(r.product_name || '').trim())
      .map((r: any) => {
        let matchedProd = r.product_id
          ? factoryStore.products.find((p: any) => p.id === r.product_id)
          : null
        if (!matchedProd && r.product_name) {
          const cleanName = String(r.product_name).trim().toLowerCase().replace(/\s+/g, '')
          matchedProd = factoryStore.products.find((p: any) => (p.name || '').trim().toLowerCase().replace(/\s+/g, '') === cleanName)
        }
        return {
          product_id: matchedProd ? matchedProd.id : (r.product_id || null),
          product_name: matchedProd ? matchedProd.name : String(r.product_name || '').trim(),
          quantity: Number(r.quantity || 1),
          unit: String(r.unit || matchedProd?.unit || 'PCS').toUpperCase(),
          price: Number(r.price || 0),
        }
      })

    let baseNotesObj: Record<string, any> = {}
    if (currentSale.value.notes) {
      try {
        const parsed = JSON.parse(currentSale.value.notes)
        if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
          baseNotesObj = { ...parsed }
        }
      } catch {
        // If previous notes was plain text, preserve it under 'notes'
        baseNotesObj = { notes: currentSale.value.notes }
      }
    }
    baseNotesObj.return_items = normalizedReturnItems

    const payload = {
      employee_id: currentSale.value.employeeID || null,
      store_id: currentSale.value.storeID || null,
      store_category: currentSale.value.storeCategory,
      total_amount: currentSale.value.totalAmount,
      discount_amount: Number(currentSale.value.discountAmount || 0),
      transaction_date: currentSale.value.transactionDate,
      receipt_no: currentSale.value.receiptNo,
      receipt_image_url: currentSale.value.receiptImageURL,
      transfer_proof_url: currentSale.value.transferProofURL || undefined,
      status: currentSale.value.status === 'APPROVED' ? 'VERIFIED' : currentSale.value.status,
      payment_method: currentSale.value.paymentMethod,
      payment_status: authStore.user?.role === 'admin_nota' && !isEditing.value ? 'UNPAID' : currentSale.value.paymentStatus,
      paid_amount: authStore.user?.role === 'admin_nota' && !isEditing.value ? 0 : Number(currentSale.value.paidAmount || 0),
      payment_due_date: currentSale.value.paymentMethod === 'TEMPO' && currentSale.value.paymentDueDate ? new Date(currentSale.value.paymentDueDate).toISOString() : undefined,
      tenor: currentSale.value.paymentMethod === 'TEMPO' 
        ? (currentSale.value.tempoInstallmentCount === 'custom' ? Number(currentSale.value.tempoCustomInstallmentCount || 1) : Number(currentSale.value.tempoInstallmentCount || 1))
        : 1,
      company_id:
        currentSale.value.companyID ||
        masterDataStore.selectedBranchCompanyId ||
        allStores.value.find((s: any) => s.id === currentSale.value.storeID)?.company_id ||
        null,
      items: currentSale.value.items.filter((i: any) => i.product_id).map((i: any) => ({
        product_id: i.product_id || null,
        quantity: Number(i.quantity),
        price: Number(i.price),
        discount_percent: Number(i.discount_percent || 0),
      })),
      notes: JSON.stringify(baseNotesObj),
    }

    if (isEditing.value) {
      await apiClient.put(`/admin/sales/transactions/${currentSale.value.id}`, payload)
      toast.success('Transaksi diperbarui')
    } else {
      await apiClient.post('/admin/sales/manual-entry', payload)
      toast.success('Transaksi baru ditambahkan')
    }
    showModal.value = false
    fetchData()
  } catch (error: any) {
    const errorMsg = error.response?.data?.error || error.message || 'Gagal menyimpan transaksi'
    toast.error(errorMsg)
  }
}

const handleApprove = async (id: string) => {
  const item = salesData.value.find((s) => s.id === id)
  if (!item) return

  try {
    // API requires receipt_no and total_amount for verification
    await apiClient.patch(`/sales/transactions/${id}/verify`, {
      receipt_no: item.receiptNo || 'VERIFIED-' + id.slice(0, 4),
      total_amount: item.amount,
      notes: 'Verified by Admin',
    })
    toast.success(
      `Nota ${item.receiptNo || id.slice(0, 8)} disetujui! Omzet salesman akan diperbarui otomatis.`,
    )
    fetchData() // Refresh
  } catch (error) {
    toast.error('Gagal menyetujui nota')
  }
}

const addItem = () => {
  currentSale.value.items.push({
    product_id: '',
    quantity: 1,
    price: 0,
    discount_percent: 0,
  })
}

const removeItem = (index: number) => {
  currentSale.value.items.splice(index, 1)
}

const addReturnItem = () => {
  if (!currentSale.value.return_items) {
    currentSale.value.return_items = []
  }
  currentSale.value.return_items.push({
    product_id: '',
    product_name: '',
    quantity: 1,
    unit: 'PCS',
    price: 0,
  })
}

const removeReturnItem = (index: number) => {
  currentSale.value.return_items.splice(index, 1)
}

const updateReturnItemProduct = (index: number) => {
  const ret = currentSale.value.return_items?.[index]
  if (!ret) return
  const product = factoryStore.products.find((p) => p.id === ret.product_id)
  if (product) {
    ret.product_name = product.name
    ret.unit = (product.unit || 'PCS').toUpperCase()
    ret.price = Number(product.selling_price || product.cost_price || 0)
  }
}

const updateItemPrice = (index: number) => {
  const item = currentSale.value.items[index]
  const product = factoryStore.products.find((p) => p.id === item.product_id)
  if (product) {
    item.price = product.selling_price
  }
}

watch(
  () => [
    currentSale.value.items,
    currentSale.value.return_items,
    currentSale.value.discountAmount,
    currentSale.value.globalDiscountPercent,
    currentSale.value.globalDiscountType,
  ],
  ([newItems, newReturnItems]) => {
    // Only recalculate total if items or return_items are provided
    const hasItems = Array.isArray(newItems) && newItems.length > 0
    const hasReturnItems = Array.isArray(newReturnItems) && newReturnItems.length > 0
    if (hasItems || hasReturnItems) {
      let subtotalItems = 0
      if (hasItems) {
        subtotalItems = newItems.reduce((acc: number, item: any) => {
          const qty = Number(item.quantity || 0)
          const prc = Number(item.price || 0)
          const disc = Number(item.discount_percent || 0)
          return acc + (qty * prc * (1 - disc / 100))
        }, 0)
      }

      // If global discount type is PERCENT, recalculate discountAmount
      if (currentSale.value.globalDiscountType === 'PERCENT') {
        const pct = Number(currentSale.value.globalDiscountPercent || 0)
        currentSale.value.discountAmount = Math.round((subtotalItems * pct) / 100)
      }

      const globalDisc = Number(currentSale.value.discountAmount || 0)
      let returnTotal = 0
      if (hasReturnItems) {
        returnTotal = newReturnItems.reduce((acc: number, item: any) => {
          return acc + (Number(item.quantity || 0) * Number(item.price || 0))
        }, 0)
      }

      const calculated = Math.max(0, subtotalItems - globalDisc - returnTotal)
      currentSale.value.totalAmount = Math.round(calculated)
    }
  },
  { deep: true },
)

const handleReject = async (id: string) => {
  if (!confirm('Apakah Anda yakin ingin menolak nota ini?')) return
  try {
    await apiClient.patch(`/sales/transactions/${id}/reject`, { notes: 'Ditolak oleh Admin' })
    toast.success('Nota berhasil ditolak')
    fetchData()
  } catch (err) {
    toast.error('Gagal menolak nota')
  }
}

const handleConvertToSO = async (sale: any) => {
  const receiptLabel = sale.receiptNo || sale.id.slice(0, 8)
  const storeLabel = sale.store || 'Toko Pelanggan'
  if (
    !confirm(
      `Ubah faktur ini (${receiptLabel} - ${storeLabel}) menjadi Pesanan SO (Sales Order)?\n\nPesanan akan dialihkan ke menu Pesanan SO agar bisa dijadwalkan pengirimannya via Batch Pengiriman / Surat Jalan.`,
    )
  ) {
    return
  }

  try {
    const res = await apiClient.post(`/admin/sales/transactions/${sale.id}/convert-to-so`)
    const soNumber = res.data?.data?.so_number || ''
    toast.success(`Faktur berhasil diubah menjadi Pesanan SO ${soNumber}!`)
    fetchData()
  } catch (err: any) {
    const msg = err.response?.data?.error || 'Gagal mengubah faktur menjadi SO'
    toast.error(msg)
  }
}

const generateReceiptNo = async () => {
  try {
    const companyId = masterDataStore.selectedBranchCompanyId
    const date = currentSale.value.transactionDate
    const res = await apiClient.get(
      `/admin/sales/transactions/next-receipt-no?company_id=${companyId}&date=${date}`,
    )
    if (res.data?.receipt_no) {
      currentSale.value.receiptNo = res.data.receipt_no
    }
  } catch (error) {
    toast.error('Gagal membuat nomor nota')
  }
}

const {
  isRefreshing,
  lastRefreshedTimeStr,
  markRefreshed,
  refresh: triggerRefresh,
} = useDataRefresh(
  async ({ isManual }) => {
    const tasks: Promise<unknown>[] = [
      fetchData(),
      fetchSalesmen(true),
    ]
    if (activeMainInvoiceTab.value === 'cabang') {
      tasks.push(fetchBranchTransfers())
    }
    if (showModal.value || allStores.value.length > 0) {
      tasks.push(fetchStores())
    }
    if (isManual) {
      tasks.push(
        factoryStore.fetchProducts(true),
        masterDataStore.fetchCompanies(true),
        masterDataStore.fetchEmployees(true),
        masterDataStore.fetchBranches(true),
      )
    }
    await Promise.allSettled(tasks)

    // Sinkronkan detail nota yang sedang dibuka di modal pratinjau
    if (selectedNota.value?.id) {
      const updatedSale = salesData.value.find((s) => s.id === selectedNota.value.id)
      if (updatedSale) {
        selectedNota.value = {
          ...updatedSale,
          raw: updatedSale.raw || updatedSale,
        }
      }
    }
    if (selectedBranchTransfer.value?.id && showBranchPrintModal.value) {
      const updatedTransfer = branchTransfers.value.find(
        (t: any) =>
          t.id === selectedBranchTransfer.value.id ||
          t.delivery_order_no === selectedBranchTransfer.value.delivery_order_no,
      )
      if (updatedTransfer) {
        selectedBranchTransfer.value = updatedTransfer
      }
    }
  },
  {
    successMessage: 'Data Faktur Penjualan berhasil diperbarui',
    canAutoRefresh: () =>
      !showModal.value &&
      !showDirectSaleModal.value &&
      !showDirectSaleReturnModal.value &&
      !showBranchInvoiceModal.value &&
      !showBranchReturnModal.value &&
      !showTempoPaymentModal.value &&
      !showDeleteConfirmModal.value &&
      !showRestoreConfirmModal.value &&
      !showOfflineImportModal.value &&
      !isLoading.value &&
      !isLoadingBranchTransfers.value,
  },
)

onMounted(() => {
  if (route.path.includes('/pesanan')) {
    selectedStatus.value = 'PENDING'
  } else if (route.path.includes('/faktur')) {
    selectedStatus.value = 'APPROVED'
  }

  Promise.allSettled([
    fetchData(),
    fetchSalesmen(),
    masterDataStore.fetchCompanies(),
    factoryStore.fetchProducts(),
    masterDataStore.fetchEmployees(),
    masterDataStore.fetchBranches(),
  ]).then(() => markRefreshed())
})

// Watch for branch changes
watch(
  () => masterDataStore.selectedBranchId,
  (newBranchId, oldBranchId) => {
    if (newBranchId === oldBranchId) return
    currentPage.value = 1
    selectedSalesman.value = ''
    if (currentSale.value) {
      currentSale.value.employeeID = ''
    }
    fetchData().then(() => markRefreshed())
  },
)

watch(
  () => masterDataStore.selectedSalesType,
  (newType, oldType) => {
    if (newType === oldType) return
    currentPage.value = 1
    selectedSalesman.value = ''
    fetchData().then(() => markRefreshed())
  },
)

const selectedInvoiceType = ref<'ALL' | 'DIRECT_SALE' | 'REGULAR_SALES'>('ALL')
const selectedTargetEntity = ref('')
const selectedSalesman = ref('')
const selectedMonth = ref('ALL')
const selectedYear = ref('ALL')
const selectedDate = ref('')
const selectedPT = ref('')
const selectedPaymentMethod = ref('ALL')
const selectedPaymentStatus = ref('ALL')

const getBranchBadge = (storeName: string) => {
  if (!storeName) return null
  const upper = storeName.toUpperCase()
  if (upper.includes('SANKE BERSINAR TERANG') || upper.includes('SBT')) return 'PT SBT'
  if (upper.includes('SHOWA LANCAR SEJAHTERA') || upper.includes('SLS')) return 'PT SLS'
  if (upper.includes('PREDATOR ALAM SEMESTA') || upper.includes('PAS')) return 'PT PAS'
  if (upper.includes('MUKASI OGON') || upper.includes('MOB')) return 'PT MOB'
  if (upper.includes('MURAMASA UNIVERSAL') || upper.includes('MUI')) return 'PT MUI'
  if (upper.includes('CUAN BESAR INDONESIA KUDUS')) return 'PT CBI Kudus'
  if (upper.includes('CUAN BESAR INDONESIA CABANG SERANG') || upper.includes('CUAN BESAR INDONESIA SERANG')) return 'PT CBI Serang'
  if (upper.includes('CUAN BESAR INDONESIA BANDUNG')) return 'PT CBI Bandung'
  if (upper.includes('CUAN BESAR INDONESIA CABANG BOGOR') || upper.includes('CUAN BESAR INDONESIA BOGOR')) return 'PT CBI Bogor'
  if (upper.includes('BOOSTER LATUBO')) return 'Booster Latubo'
  if (upper.includes('BOOSTER SURABAYA')) return 'Booster Surabaya'
  if (upper.includes('KOHAKU SUMBER ABADI') || upper.includes('KSA')) return 'PT KSA'
  return null
}

const years = computed(() => {
  const currentYear = new Date().getFullYear()
  return Array.from({ length: 5 }, (_, i) => (currentYear - i).toString())
})

const resetFilters = () => {
  selectedInvoiceType.value = 'ALL'
  selectedTargetEntity.value = ''
  selectedPT.value = ''
  selectedSalesman.value = ''
  selectedMonth.value = 'ALL'
  selectedYear.value = 'ALL'
  selectedDate.value = ''
  selectedPaymentMethod.value = 'ALL'
  selectedPaymentStatus.value = 'ALL'
}

let debounceTimer: any
watch(
  [
    selectedStatus,
    searchQuery,
    selectedInvoiceType,
    selectedTargetEntity,
    selectedPT,
    selectedSalesman,
    selectedMonth,
    selectedYear,
    selectedDate,
    selectedPaymentMethod,
    selectedPaymentStatus,
    itemsPerPage,
  ],
  () => {
    clearTimeout(debounceTimer)
    debounceTimer = setTimeout(() => {
      currentPage.value = 1
      fetchData().then(() => markRefreshed())
    }, 500) // debounce API calls
  },
)

watch(currentPage, () => {
  fetchData().then(() => markRefreshed())
})

const isMapModalOpen = ref(false)
const selectedOrderForMap = ref<any>(null)

const getResolvedOrder = (order: any) => {
  if (!order) return null
  const base = order.raw ? { ...order.raw, ...order } : { ...order }
  
  // If store coordinates are missing on base.store, try to find in allStores or masterDataStore
  const storeId = base.store_id || base.storeID || base.store?.id
  if (storeId) {
    const foundStore = allStores.value.find((s: any) => s.id === storeId) || 
                       masterDataStore.stores?.find((s: any) => s.id === storeId)
    if (foundStore && (foundStore.latitude || foundStore.longitude)) {
      base.store = { ...foundStore, ...base.store }
    }
  }

  // If latitude / longitude is in visit, fallback to visit
  if ((!base.latitude || !base.longitude) && base.visit) {
    base.latitude = base.visit.latitude
    base.longitude = base.visit.longitude
  }

  return base
}

const openMapModal = (order: any) => {
  const target = getResolvedOrder(order)
  if (!target) return
  selectedOrderForMap.value = target
  isMapModalOpen.value = true
}

const getOrderLocationStatus = (order: any) => {
  const item = getResolvedOrder(order)
  if (!item) {
    return { status: 'UNKNOWN', text: 'Tidak ada data', color: 'text-slate-400', bg: 'bg-slate-100' }
  }

  // Khusus transaksi penjualan langsung pabrik / cabang / walk-in kasir
  if (
    item.sales_channel === 'DIRECT_FACTORY' ||
    item.sales_channel === 'DIRECT_BRANCH' ||
    item.sales_channel === 'EMPLOYEE_SALE' ||
    item.buyer_type === 'EMPLOYEE' ||
    item.receipt_no?.startsWith('FP-CAB/') ||
    item.receiptNo?.startsWith('FP-CAB/')
  ) {
    return { status: 'DIRECT', text: 'Penjualan Langsung', color: 'text-emerald-700', bg: 'bg-emerald-50', icon: Store }
  }

  const resolveNum = (...vals: any[]): number | null => {
    for (const v of vals) {
      if (v !== undefined && v !== null && v !== '' && !isNaN(Number(v))) {
        const num = Number(v)
        if (num !== 0) return num
      }
    }
    return null
  }

  const lat = resolveNum(item.latitude, item.visit?.latitude, item.raw?.latitude)
  const lng = resolveNum(item.longitude, item.visit?.longitude, item.raw?.longitude)
  const stLat = resolveNum(item.store?.latitude, item.visit?.store?.latitude, item.raw?.store?.latitude)
  const stLng = resolveNum(item.store?.longitude, item.visit?.store?.longitude, item.raw?.store?.longitude)

  if (lat === null || lng === null || stLat === null || stLng === null) {
    return { status: 'UNKNOWN', text: 'Tidak ada data', color: 'text-slate-400', bg: 'bg-slate-100' }
  }
  const dist = getDistance(lat, lng, stLat, stLng)
  if (dist > 500) {
    return { status: 'OUTSIDE', text: `Luar Radius (${Math.round(dist)}m)`, color: 'text-red-600', bg: 'bg-red-50', icon: MapPinOff }
  }
  return { status: 'INSIDE', text: 'Sesuai Radius', color: 'text-emerald-600', bg: 'bg-emerald-50', icon: MapPin }
}

const formatCurrency = (val: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(val)
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  })
}

const salesEmployees = computed(() => {
  return salesmen.value.filter((emp: any) => {
    // 0. Filter by active status
    const isActive = emp.user?.is_active !== false && emp.employment_status !== 'RESIGNED' && emp.employment_status !== 'TERMINATED'

    // 1. Filter by branch (respect global branch filter in header)
    const branchId = masterDataStore.selectedBranchId
    const empBranchId = emp.operating_branch_id || emp.branch_id
    const belongsToBranch = !branchId || branchId === 'ALL_BRANCHES' || empBranchId === branchId

    // 2. Filter by job title (sales / motoris / booster / marketing / task order / to)
    const jobTitle = (emp.job_position?.title || '').toLowerCase()
    const isSales = /\b(sales|motoris|booster|marketing|task order|take order|to|mt|kam|modern trade)\b/i.test(jobTitle)

    // 3. Filter by Sales Type
    let matchSalesType = true
    const type = masterDataStore.selectedSalesType
    if (type === 'sales_mt') {
      matchSalesType = /\b(mt|modern trade|kam|key account)\b/i.test(jobTitle)
    } else if (type === 'sales_to') {
      matchSalesType = /\b(to|take order)\b/i.test(jobTitle)
    } else if (type === 'sales_motoris_booster') {
      matchSalesType = /\b(motoris|booster)\b/i.test(jobTitle)
    }

    return isActive && belongsToBranch && isSales && matchSalesType
  })
})

const uniqueSalesmen = computed(() => {
  return salesEmployees.value.map((s: any) => ({
    name: `${s.first_name} ${s.last_name}`,
    displayName: `${s.first_name} ${s.last_name} (${s.job_position?.title || 'Sales'})`
  }))
})

const salesmanOptions = computed<SelectOption[]>(() => {
  return salesEmployees.value.map((s: any) => ({
    value: s.id,
    label: `${s.first_name} ${s.last_name}`,
    sublabel: [s.job_position?.title || 'Sales', s.operating_branch?.name || s.branch?.name].filter(Boolean).join(' • '),
    extra: s,
  }))
})

const handleStoreSelect = (_storeId: string, option?: SelectOption) => {
  const store = option?.extra || salesmanStores.value.find((s: any) => s.id === _storeId) || allStores.value.find((s: any) => s.id === _storeId)
  if (store) {
    if (store.category === 'TOKO_BARU') {
      currentSale.value.storeCategory = 'TOKO_BARU'
    } else {
      currentSale.value.storeCategory = 'TOKO_LAMA'
    }
  }
}

const storeOptions = computed<SelectOption[]>(() => {
  const currentEmpId = currentSale.value?.employeeID

  // Combine fallback pools: allStores and masterDataStore.stores
  const basePool = [...allStores.value]
  if (masterDataStore.stores && masterDataStore.stores.length > 0) {
    masterDataStore.stores.forEach((st: any) => {
      if (st?.id && !basePool.some((b: any) => b.id === st.id)) {
        basePool.push(st)
      }
    })
  }

  const mapStoreToOption = (st: any, isAssignedSalesman = false): SelectOption => {
    const addressParts = [st.address, st.district, st.regency, st.province].filter(Boolean)
    const fullAddress = addressParts.length > 0 ? addressParts.join(', ') : (st.address || '')
    const sublabelParts: string[] = []
    if (st.id) {
      sublabelParts.push(`ID: ${st.id.slice(0, 8).toUpperCase()}`)
    }
    if (fullAddress) sublabelParts.push(fullAddress)
    if (st.owner_name) sublabelParts.push(`Pemilik: ${st.owner_name}`)
    if (st.phone_number) sublabelParts.push(`Telp: ${st.phone_number}`)

    let badge: string | undefined
    let badgeColor: string | undefined

    if (isAssignedSalesman) {
      badge = 'Toko Binaan'
      badgeColor = 'bg-blue-50 text-blue-700 border border-blue-200'
    } else if (st.category === 'TOKO_BARU') {
      badge = 'Toko Baru'
      badgeColor = 'bg-emerald-50 text-emerald-700'
    }

    return {
      value: st.id,
      label: st.name,
      sublabel: sublabelParts.join(' • '),
      badge,
      badgeColor,
      extra: st,
    }
  }

  if (currentEmpId) {
    const storeMap = new Map<string, SelectOption>()

    // 1. First add stores returned specifically for this salesman
    salesmanStores.value.forEach((st: any) => {
      if (st?.id) {
        storeMap.set(st.id, mapStoreToOption(st, true))
      }
    })

    // 2. Add stores transacted by this salesman on currently loaded salesData
    salesData.value.forEach((sale: any) => {
      const empId = sale.employeeID || sale.raw?.employee_id
      if (empId === currentEmpId && sale.raw?.store?.id) {
        if (!storeMap.has(sale.raw.store.id)) {
          storeMap.set(sale.raw.store.id, mapStoreToOption(sale.raw.store, true))
        }
      }
    })

    // 3. Also check basePool for stores with matching assigned_employee_id or created_by_employee_id
    basePool.forEach((st: any) => {
      const assignedId = st.assigned_employee_id || st.assigned_employee?.id || st.created_by_employee_id
      if (st?.id && assignedId === currentEmpId) {
        if (!storeMap.has(st.id)) {
          storeMap.set(st.id, mapStoreToOption(st, true))
        }
      }
    })

    // 4. Fallback: also append other branch stores so user can select any store
    basePool.forEach((st: any) => {
      if (st?.id && !storeMap.has(st.id)) {
        storeMap.set(st.id, mapStoreToOption(st, false))
      }
    })

    // If current selected storeID is not in the list, ensure it's included
    if (currentSale.value?.storeID && !storeMap.has(currentSale.value.storeID)) {
      const existing = basePool.find((s: any) => s.id === currentSale.value.storeID)
      if (existing) {
        storeMap.set(existing.id, mapStoreToOption(existing, false))
      }
    }

    return Array.from(storeMap.values())
  }

  // If no salesman selected yet, return all available stores from basePool
  return basePool.map((st: any) => mapStoreToOption(st, false))
})

const uniquePTs = computed(() => {
  return masterDataStore.companies.map((c: any) => ({ id: c.id, name: c.name }))
})

const filteredSales = computed(() => {
  return salesData.value
})

const totalFilteredAmount = computed(() => {
  return serverTotalAmount.value
})

// ─── Multi-Selection & Bulk Print ───────────────────────────────────────────
const selectedSaleIds = ref<Set<string>>(new Set())
const showBulkInvoicePrintModal = ref(false)

const isAllSelected = computed(() => {
  if (filteredSales.value.length === 0) return false
  return filteredSales.value.every(s => selectedSaleIds.value.has(s.id))
})

const isSomeSelected = computed(() => {
  return selectedSaleIds.value.size > 0 && !isAllSelected.value
})

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedSaleIds.value.clear()
  } else {
    filteredSales.value.forEach(s => selectedSaleIds.value.add(s.id))
  }
}

const toggleSelectSale = (id: string) => {
  if (selectedSaleIds.value.has(id)) {
    selectedSaleIds.value.delete(id)
  } else {
    selectedSaleIds.value.add(id)
  }
}

const clearSelection = () => {
  selectedSaleIds.value.clear()
}

const selectedSalesList = computed(() => {
  return salesData.value.filter(s => selectedSaleIds.value.has(s.id))
})

const selectedSalesTotalAmount = computed(() => {
  return selectedSalesList.value.reduce((sum, s) => sum + (s.amount || 0), 0)
})

const openBulkPrint = () => {
  if (selectedSaleIds.value.size === 0) {
    toast.error('Pilih setidaknya 1 faktur untuk dicetak')
    return
  }
  showBulkInvoicePrintModal.value = true
}

const batchProductColumns = computed(() => {
  if (!selectedDO.value || !selectedDO.value.items) return []
  const productMap = new Map()
  selectedDO.value.items.forEach((item: any) => {
    if (item.sales_transaction?.items) {
      item.sales_transaction.items.forEach((si: any) => {
        if (si.product) {
          productMap.set(si.product_id, {
            id: si.product_id,
            name: si.product.name
              .split(' ')
              .map((w: string) => w[0])
              .join(''),
          })
        }
      })
    }
  })
  return Array.from(productMap.values())
})

const getProductQty = (trx: any, productId: string) => {
  if (!trx || !trx.items) return 0
  const item = trx.items.find((i: any) => i.product_id === productId)
  return item ? item.quantity : 0
}

const getDiscrepancyStatus = (sale: any) => {
  if (!sale.raw?.payments || sale.raw.payments.length === 0) return null
  
  for (const p of sale.raw.payments) {
    if (p.notes && p.notes.includes('[SELISIH')) {
      if (p.notes.includes('[ADJUSTED]')) {
        return 'RESOLVED'
      } else if (p.payment_status === 'PENDING') {
        return 'PENDING'
      }
    }
  }
  return null
}

const calculateBatchTotal = computed(() => {
  if (!selectedDO.value || !selectedDO.value.items) return 0
  return selectedDO.value.items.reduce(
    (sum: number, item: any) => sum + (item.sales_transaction?.total_amount || 0),
    0,
  )
})

const calculateProductTotal = (productId: string) => {
  if (!selectedDO.value || !selectedDO.value.items) return 0
  return selectedDO.value.items.reduce(
    (sum: number, item: any) => sum + getProductQty(item.sales_transaction, productId),
    0,
  )
}

const openSJPriority = async (batch: any) => {
  if (!batch.id) return
  isLoading.value = true
  try {
    const res = await apiClient.get(`/delivery/batch/${batch.id}`)
    if (res.data) {
      selectedDO.value = res.data
    }
  } catch (err) {
    toast.error('Gagal mengambil detail Surat Jalan')
  } finally {
    isLoading.value = false
  }
}
const openNotaPreview = (sale: any) => {
  // Find raw data to get all details including store and items
  const rawSale = salesData.value.find((s) => s.id === sale.id)
  selectedNota.value = {
    ...sale,
    raw: rawSale?.raw || rawSale,
  }
}

const currentBranchName = computed(() => {
  if (masterDataStore.selectedBranchId === 'ALL_BRANCHES') return 'Semua Cabang'
  const branch = masterDataStore.branches.find((b) => b.id === masterDataStore.selectedBranchId)
  return branch ? branch.name : 'Semua Cabang'
})

// === LOGIKA FAKTUR PENGIRIMAN CABANG (ADMIN NOTA PABRIK / PUSAT) ===
const activeMainInvoiceTab = ref<'toko' | 'tempo' | 'cabang'>('toko')
const branchTransfers = ref<any[]>([])
const isLoadingBranchTransfers = ref(false)
const showBranchInvoiceModal = ref(false)
const showBranchPrintModal = ref(false)
const selectedBranchTransfer = ref<any>(null)
const branchSearchQuery = ref('')
const branchStatusFilter = ref('ALL')
const branchTargetFilter = ref('ALL')

const branchCurrentPage = ref(1)
const branchItemsPerPage = ref<number | 'all'>(10)
const branchTotalItems = ref(0)
const branchServerTotalAmount = ref(0)
const branchServerStats = ref<{
  total_invoices: number
  total_amount: number
  pending_count: number
  shipped_count: number
  received_count: number
} | null>(null)

const branchTotalPages = computed(() => {
  if (branchItemsPerPage.value === 'all') return 1
  return Math.ceil(branchTotalItems.value / (branchItemsPerPage.value as number)) || 1
})

const branchVisiblePages = computed(() => {
  const current = branchCurrentPage.value
  const total = branchTotalPages.value
  const delta = 2
  const pages: (number | string)[] = []

  for (let i = 1; i <= total; i++) {
    if (i === 1 || i === total || (i >= current - delta && i <= current + delta)) {
      pages.push(i)
    } else if (pages[pages.length - 1] !== '...') {
      pages.push('...')
    }
  }
  return pages
})

const fetchBranchTransfers = async () => {
  isLoadingBranchTransfers.value = true
  try {
    const params = new URLSearchParams()
    if (branchItemsPerPage.value !== 'all') {
      params.append('page', branchCurrentPage.value.toString())
      params.append('limit', branchItemsPerPage.value.toString())
    }
    if (branchTargetFilter.value && branchTargetFilter.value !== 'ALL') {
      params.append('branch_id', branchTargetFilter.value)
    }
    if (branchStatusFilter.value && branchStatusFilter.value !== 'ALL') {
      params.append('status', branchStatusFilter.value)
    }
    if (branchSearchQuery.value) {
      params.append('search', branchSearchQuery.value)
    }
    if (selectedPT.value) {
      params.append('company_id', selectedPT.value)
    }

    const res = await apiClient.get(`/factory/stock/transfers?${params.toString()}`)
    const rawData = res.data
    const list: any[] = Array.isArray(rawData) ? rawData : (rawData?.data || [])

    if (rawData && !Array.isArray(rawData) && rawData.total !== undefined) {
      branchTotalItems.value = rawData.total || 0
      branchServerTotalAmount.value = rawData.total_amount || 0
      if (rawData.stats) {
        branchServerStats.value = rawData.stats
      }
    } else {
      branchTotalItems.value = list.length
      branchServerTotalAmount.value = list.reduce((s: number, it: any) => s + (it.total_amount || 0), 0)
    }

    // Group by delivery_order_no or id
    const grouped: Record<string, any> = {}
    for (const item of list) {
      const key = item.delivery_order_no || item.id
      if (!grouped[key]) {
        grouped[key] = {
          id: item.id,
          delivery_order_no: item.delivery_order_no || `SJ-${item.id?.substring(0, 8)}`,
          status: item.status,
          to_branch_id: item.to_branch_id,
          to_branch: item.to_branch,
          from_factory_id: item.from_factory_id,
          from_factory: item.from_factory,
          vehicle: item.vehicle,
          driver: item.driver,
          payment_method: item.payment_method,
          payment_due_at: item.payment_due_at,
          payment_status: item.payment_status,
          notes: item.notes,
          initiated_by: item.initiated_by,
          shipped_at: item.shipped_at,
          created_at: item.created_at,
          approved_by_name: item.approved_by_name,
          total_amount: 0,
          total_quantity: 0,
          items: [],
        }
      }
      grouped[key].total_amount += (item.total_amount || 0)
      grouped[key].total_quantity += (item.quantity || 0)
      grouped[key].items.push(item)
    }

    branchTransfers.value = Object.values(grouped).sort(
      (a: any, b: any) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
    )
  } catch (error) {
    console.error('Failed to load branch transfers:', error)
  } finally {
    isLoadingBranchTransfers.value = false
  }
}

const filteredBranchTransfers = computed(() => {
  return branchTransfers.value
})

const branchStats = computed(() => {
  if (branchServerStats.value) {
    return {
      totalInvoices: branchServerStats.value.total_invoices || 0,
      totalAmount: branchServerStats.value.total_amount || 0,
      pendingCount: branchServerStats.value.pending_count || 0,
      shippedCount: branchServerStats.value.shipped_count || 0,
      receivedCount: branchServerStats.value.received_count || 0,
    }
  }
  const totalInvoices = branchTransfers.value.length
  const totalAmount = branchTransfers.value.reduce((s: number, t: any) => s + (t.total_amount || 0), 0)
  const pendingCount = branchTransfers.value.filter((t: any) => t.status === 'REQUESTED' || t.status === 'APPROVED').length
  const shippedCount = branchTransfers.value.filter((t: any) => t.status === 'SHIPPED' || t.status === 'ARRIVED').length
  const receivedCount = branchTransfers.value.filter((t: any) => t.status === 'RECEIVED').length
  return {
    totalInvoices,
    totalAmount,
    pendingCount,
    shippedCount,
    receivedCount,
  }
})

watch(branchCurrentPage, () => {
  fetchBranchTransfers()
})

watch([branchItemsPerPage, branchTargetFilter, branchStatusFilter], () => {
  branchCurrentPage.value = 1
  fetchBranchTransfers()
})

let branchSearchTimeout: any = null
watch(branchSearchQuery, () => {
  if (branchSearchTimeout) clearTimeout(branchSearchTimeout)
  branchSearchTimeout = setTimeout(() => {
    branchCurrentPage.value = 1
    fetchBranchTransfers()
  }, 400)
})

watch(activeMainInvoiceTab, (tab) => {
  if (tab === 'cabang') {
    fetchBranchTransfers()
  } else if (tab === 'tempo') {
    selectedStatus.value = 'TEMPO_UNPAID'
    currentPage.value = 1
    fetchData()
  } else if (tab === 'toko') {
    if (selectedStatus.value === 'TEMPO_UNPAID') {
      selectedStatus.value = 'ALL'
    }
    currentPage.value = 1
    fetchData()
  }
})

const openBranchInvoicePreview = (transfer: any) => {
  selectedBranchTransfer.value = transfer
  showBranchPrintModal.value = true
}

const showBranchReturnModal = ref(false)
const selectedReturnTransfer = ref<any>(null)

const openBranchReturnModal = (transfer: any) => {
  selectedReturnTransfer.value = transfer
  showBranchReturnModal.value = true
}

const handleBranchReturnCreated = () => {
  showBranchReturnModal.value = false
  fetchBranchTransfers()
}

const handleBranchTransferCreated = () => {
  fetchBranchTransfers()
}

const selectedEditBranchTransfer = ref<any>(null)

const openBranchCreateModal = () => {
  selectedEditBranchTransfer.value = null
  showBranchInvoiceModal.value = true
}

const openBranchEditModal = (transfer: any) => {
  selectedEditBranchTransfer.value = transfer
  showBranchInvoiceModal.value = true
}

const handleBranchTransferUpdated = () => {
  selectedEditBranchTransfer.value = null
  showBranchInvoiceModal.value = false
  fetchBranchTransfers()
}

const canEditBranchTransfer = (transfer: any) => {
  if (transfer.status === 'REQUESTED' || transfer.status === 'APPROVED' || transfer.status === 'LOADING') {
    return true
  }
  const role = authStore.user?.role
  return role === 'superadmin' || role === 'admin_nota'
}

const handleShipTransfer = async (transfer: any) => {
  if (!confirm(`Kirim faktur pengiriman ${transfer.delivery_order_no}? Stok pabrik akan dipotong permanen.`)) return
  try {
    await apiClient.post(`/factory/transfer/${transfer.id}/ship`)
    toast.success('Pengiriman berhasil dieksekusi! Stok pabrik terpotong dan piutang cabang aktif.')
    fetchBranchTransfers()
  } catch (error: any) {
    toast.error(error.response?.data?.error || error.message || 'Gagal mengeksekusi pengiriman')
  }
}
</script>

<template>
  <div class="print:bg-white">
    <div class="space-y-8 pb-20 print:hidden">
      <!-- TABS NAV: Penjualan Toko vs Faktur Tempo vs Faktur Pengiriman Cabang -->
      <div class="flex flex-wrap items-center gap-2 bg-slate-100 p-1.5 rounded-2xl border border-slate-200/80 w-fit">
        <button
          type="button"
          @click="activeMainInvoiceTab = 'toko'"
          :class="[
            'flex items-center gap-2 px-5 py-2.5 rounded-xl font-black text-xs sm:text-sm tracking-tight transition-all',
            activeMainInvoiceTab === 'toko'
              ? 'bg-white text-slate-900 shadow-sm border border-slate-200/60'
              : 'text-slate-500 hover:text-slate-900',
          ]"
        >
          <Store class="w-4 h-4 text-emerald-600" />
          Faktur Penjualan Toko (Pelanggan)
        </button>

        <button
          type="button"
          @click="activeMainInvoiceTab = 'tempo'"
          :class="[
            'flex items-center gap-2 px-5 py-2.5 rounded-xl font-black text-xs sm:text-sm tracking-tight transition-all',
            activeMainInvoiceTab === 'tempo'
              ? 'bg-indigo-600 text-white shadow-md shadow-indigo-600/30 ring-2 ring-indigo-400'
              : 'text-slate-600 hover:text-indigo-600 hover:bg-slate-200/60',
          ]"
        >
          <CalendarClock class="w-4 h-4 text-amber-400" />
          Faktur Tempo Belum Lunas
        </button>

        <button
          type="button"
          @click="activeMainInvoiceTab = 'cabang'"
          :class="[
            'flex items-center gap-2 px-5 py-2.5 rounded-xl font-black text-xs sm:text-sm tracking-tight transition-all',
            activeMainInvoiceTab === 'cabang'
              ? 'bg-indigo-600 text-white shadow-md shadow-indigo-600/20'
              : 'text-slate-500 hover:text-slate-900',
          ]"
        >
          <Truck class="w-4 h-4" />
          Faktur Pengiriman Cabang (Pusat ke Cabang)
          <span
            v-if="branchStats.pendingCount > 0"
            class="ml-1 px-2 py-0.5 text-[10px] font-black rounded-full bg-amber-400 text-slate-900"
          >
            {{ branchStats.pendingCount }} Baru
          </span>
        </button>
      </div>

      <!-- TAB 1 & 2: Penjualan Toko / Pelanggan (Reguler & Tempo Belum Lunas) -->
      <div v-if="activeMainInvoiceTab === 'toko' || activeMainInvoiceTab === 'tempo'" class="space-y-8">
        <!-- SECTION 1: Admin Nota (Log Batch & Verifikasi Penjualan) -->
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-6">
        <div>
          <h1 class="text-xl sm:text-2xl font-extrabold text-slate-900 tracking-tight flex items-center flex-wrap gap-2">
            <span>Verifikasi Faktur Penjualan Admin Nota</span>
            <span v-if="currentBranchName" class="text-primary text-lg sm:text-xl font-bold">
              — {{ currentBranchName }}
            </span>
          </h1>
          <p class="text-[15px] font-medium text-slate-500 mt-1.5 flex items-center gap-2">
            Cetak Faktur Penjualan dari batch pengiriman yang ditentukan oleh Supervisor dan verifikasi
            Faktur Penjualan baru.
          </p>
        </div>
        <div class="flex flex-wrap items-center gap-2 sm:gap-4 w-full sm:w-auto">
          <RefreshDataButton
            :loading="isRefreshing || isLoading || isLoadingBranchTransfers"
            :last-updated="lastRefreshedTimeStr"
            @click="triggerRefresh()"
          />
          <input
            ref="offlineImportInput"
            type="file"
            accept="application/json,.json"
            class="hidden"
            @change="importOfflineBackup"
          />
          <Button
            @click="showOfflineImportModal = true"
            :disabled="isImportingOffline"
            class="bg-amber-500 hover:bg-amber-600 text-white rounded-xl sm:rounded-2xl px-3.5 py-2.5 sm:px-6 sm:py-6 text-xs sm:text-base font-bold shadow-md sm:shadow-lg shadow-amber-500/20 flex items-center justify-center gap-1.5 sm:gap-2 flex-1 sm:flex-initial min-w-[170px] sm:min-w-0"
          >
            <Upload class="w-4 h-4 sm:w-5 sm:h-5" />
            {{ isImportingOffline ? 'Mengimpor...' : 'Import Offline' }}
          </Button>
          <Button
            @click="openSalesmanLPS()"
            class="bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl sm:rounded-2xl px-3.5 py-2.5 sm:px-6 sm:py-6 text-xs sm:text-base font-bold shadow-md sm:shadow-lg shadow-indigo-600/20 flex items-center justify-center gap-1.5 sm:gap-2 flex-1 sm:flex-initial min-w-[130px] sm:min-w-0"
            title="Buka Lembar Penjualan Sales (LPS)"
          >
            <FileText class="w-4 h-4 sm:w-5 sm:h-5" /> Buka LPS
          </Button>
          <Button
            @click="handleExportXlsx"
            class="bg-green-600 hover:bg-green-700 text-white rounded-xl sm:rounded-2xl px-3.5 py-2.5 sm:px-6 sm:py-6 text-xs sm:text-base font-bold shadow-md sm:shadow-lg shadow-green-600/20 flex items-center justify-center gap-1.5 sm:gap-2 flex-1 sm:flex-initial min-w-[130px] sm:min-w-0"
          >
            <ClipboardList class="w-4 h-4 sm:w-5 sm:h-5" /> Export Excel
          </Button>
          <Button
            @click="showDirectSaleModal = true"
            class="bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl sm:rounded-2xl px-3.5 py-2.5 sm:px-6 sm:py-6 text-xs sm:text-base font-bold shadow-md sm:shadow-lg shadow-emerald-600/20 flex items-center justify-center gap-1.5 sm:gap-2 flex-1 sm:flex-initial min-w-[170px] sm:min-w-0"
          >
            <ShoppingBag class="w-4 h-4 sm:w-5 sm:h-5" /> + Faktur Langsung
          </Button>
          <Button
            @click="openAddModal"
            class="bg-primary hover:bg-primary/90 text-white rounded-xl sm:rounded-2xl px-3.5 py-2.5 sm:px-6 sm:py-6 text-xs sm:text-base font-bold shadow-md sm:shadow-lg shadow-primary/20 flex items-center justify-center gap-1.5 sm:gap-2 flex-1 sm:flex-initial min-w-[150px] sm:min-w-0"
          >
            <Plus class="w-4 h-4 sm:w-5 sm:h-5" /> Buat Faktur
          </Button>
          <div
            class="hidden md:block bg-emerald-50 px-4 py-2 rounded-2xl border border-emerald-100"
          >
            <p class="text-[10px] font-bold text-emerald-600 uppercase tracking-wider">
              Total Data Tampil
            </p>
            <p class="text-lg font-black text-emerald-700">
              {{ formatCurrency(totalFilteredAmount) }}
            </p>
          </div>
        </div>
      </div>

    <!-- Filters -->
    <div class="space-y-4">
      <div
        class="bg-white p-4 rounded-3xl border border-slate-200 shadow-sm flex flex-col lg:flex-row gap-4 items-center"
      >
        <div class="relative flex-1 w-full">
          <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Cari No. Nota, Toko, Cabang/PT (misal: SBT, SLS, PAS), atau Salesman..."
            class="w-full pl-12 pr-4 py-3.5 bg-slate-50 border-none rounded-2xl focus:ring-2 focus:ring-primary/20 transition-all font-medium text-slate-900"
          />
        </div>
        <div class="flex gap-3 w-full lg:w-auto overflow-x-auto no-scrollbar pb-1 lg:pb-0">
          <button
            v-for="status in ['ALL', 'PENDING', 'APPROVED', 'REJECTED', 'CANCELLED', 'DELETED']"
            :key="status"
            @click="selectedStatus = status"
            :class="[
              'px-5 py-2.5 rounded-xl font-bold text-xs tracking-wider transition-all duration-300 flex items-center gap-1.5 shrink-0',
              selectedStatus === status
                ? (status === 'DELETED' ? 'bg-rose-600 text-white shadow-lg shadow-rose-600/20' : (status === 'PENDING' ? 'bg-amber-500 text-white shadow-lg shadow-amber-500/20' : (status === 'APPROVED' ? 'bg-emerald-600 text-white shadow-lg shadow-emerald-600/20' : 'bg-slate-900 text-white shadow-lg shadow-slate-900/20')))
                : (status === 'DELETED' ? 'bg-rose-50 text-rose-600 hover:bg-rose-100' : (status === 'PENDING' ? 'bg-amber-50 text-amber-700 hover:bg-amber-100 font-extrabold' : 'bg-slate-50 text-slate-500 hover:bg-slate-100')),
            ]"
          >
            <span>{{ status === 'DELETED' ? 'SAMPAH (TERHAPUS)' : (status === 'PENDING' ? 'PENDING (BUTUH APPROVAL)' : (status === 'APPROVED' ? 'APPROVED (DISETUJUI)' : (status === 'CANCELLED' ? 'DIBATALKAN' : status))) }}</span>
          </button>
        </div>
      </div>

      <!-- Additional Filters -->
      <div
        class="bg-white p-4 rounded-3xl border border-slate-200 shadow-sm flex flex-wrap gap-4 items-center"
      >
        <!-- Filter Tipe Faktur -->
        <div class="flex-1 min-w-[150px]">
          <label class="block text-[10px] font-bold text-slate-500 uppercase tracking-wider mb-1"
            >Tipe Faktur</label
          >
          <select
            v-model="selectedInvoiceType"
            class="w-full bg-slate-50 border-none rounded-xl py-2.5 px-3 text-sm font-medium text-slate-700 focus:ring-2 focus:ring-primary/20"
          >
            <option value="ALL">Semua Tipe Faktur</option>
            <option value="DIRECT_SALE">⚡ Penjualan Langsung (Direct)</option>
            <option value="REGULAR_SALES">🏪 Reguler Salesman (TO/Canvas)</option>
          </select>
        </div>

        <!-- Filter Cabang / PT Tujuan -->
        <div class="flex-1 min-w-[170px]">
          <label class="block text-[10px] font-bold text-slate-500 uppercase tracking-wider mb-1"
            >Cabang / PT Tujuan</label
          >
          <select
            v-model="selectedTargetEntity"
            class="w-full bg-slate-50 border-none rounded-xl py-2.5 px-3 text-sm font-medium text-slate-700 focus:ring-2 focus:ring-primary/20"
          >
            <option value="">Semua Cabang / PT Tujuan</option>
            <optgroup label="Entitas PT Cabang (Penjualan Pabrik)">
              <option value="SBT">PT SBT (PT Sanke Bersinar Terang)</option>
              <option value="SLS">PT SLS (PT Showa Lancar Sejahtera)</option>
              <option value="PAS">PT PAS (PT Predator Alam Semesta)</option>
              <option value="MOB">PT MOB (PT Mukasi Ogon Berkilau)</option>
              <option value="MUI">PT MUI (PT Muramasa Universal Indonesia)</option>
              <option value="CBI_KUDUS">PT CBI Kudus</option>
              <option value="CBI_SERANG">PT CBI Serang</option>
              <option value="CBI_BANDUNG">PT CBI Bandung</option>
              <option value="CBI_BOGOR">PT CBI Bogor</option>
              <option value="BOOSTER_LATUBO">Booster Latubo</option>
              <option value="BOOSTER_SBY">Booster Surabaya</option>
              <option value="KSA">PT KSA (PT Kohaku Sumber Abadi)</option>
            </optgroup>
            <optgroup label="Cabang Fisik Lainnya">
              <option v-for="b in masterDataStore.branches" :key="b.id" :value="b.id">
                {{ b.name }}
              </option>
            </optgroup>
          </select>
        </div>

        <!-- Filter PT -->
        <div class="flex-1 min-w-[140px]">
          <label class="block text-[10px] font-bold text-slate-500 uppercase tracking-wider mb-1"
            >Filter PT</label
          >
          <select
            v-model="selectedPT"
            class="w-full bg-slate-50 border-none rounded-xl py-2.5 px-3 text-sm font-medium text-slate-700 focus:ring-2 focus:ring-primary/20"
          >
            <option value="">Semua PT</option>
            <option v-for="pt in uniquePTs" :key="pt.id" :value="pt.id">{{ pt.name }}</option>
          </select>
        </div>

        <!-- Filter Sales -->
        <div class="flex-1 min-w-[140px]">
          <label class="block text-[10px] font-bold text-slate-500 uppercase tracking-wider mb-1"
            >Filter Sales</label
          >
          <select
            v-model="selectedSalesman"
            class="w-full bg-slate-50 border-none rounded-xl py-2.5 px-3 text-sm font-medium text-slate-700 focus:ring-2 focus:ring-primary/20"
          >
            <option value="">Semua Sales</option>
            <option v-for="sales in uniqueSalesmen" :key="sales.name" :value="sales.name">
              {{ sales.displayName }}
            </option>
          </select>
        </div>

        <!-- Filter Metode Pembayaran -->
        <div class="flex-1 min-w-[160px]">
          <label class="block text-[10px] font-bold text-slate-500 uppercase tracking-wider mb-1"
            >Metode Pembayaran</label
          >
          <select
            v-model="selectedPaymentMethod"
            class="w-full bg-slate-50 border-none rounded-xl py-2.5 px-3 text-sm font-medium text-slate-700 focus:ring-2 focus:ring-primary/20"
          >
            <option value="ALL">Semua Metode</option>
            <option value="CASH">CASH (Tunai)</option>
            <option value="TRANSFER">TRANSFER BANK</option>
            <option value="TEMPO">TEMPO (Semua Tempo)</option>
            <option value="TEMPO_CASH">TEMPO - Bayar CASH</option>
            <option value="TEMPO_TRANSFER">TEMPO - Bayar TRANSFER</option>
            <option value="QRIS">QRIS</option>
            <option value="POTONG_GAJI">POTONG GAJI</option>
          </select>
        </div>

        <!-- Filter Status Pelunasan -->
        <div class="flex-1 min-w-[140px]">
          <label class="block text-[10px] font-bold text-slate-500 uppercase tracking-wider mb-1"
            >Status Pelunasan</label
          >
          <select
            v-model="selectedPaymentStatus"
            class="w-full bg-slate-50 border-none rounded-xl py-2.5 px-3 text-sm font-medium text-slate-700 focus:ring-2 focus:ring-primary/20"
          >
            <option value="ALL">Semua Status</option>
            <option value="UNPAID">Belum Bayar (UNPAID)</option>
            <option value="PARTIAL">Cicilan Berjalan (PARTIAL)</option>
            <option value="OVERDUE">⚠ Jatuh Tempo (OVERDUE)</option>
            <option value="PAID">Lunas (PAID)</option>
          </select>
        </div>

        <!-- Filter Bulanan -->
        <div class="flex-1 min-w-[120px]">
          <label class="block text-[10px] font-bold text-slate-500 uppercase tracking-wider mb-1"
            >Bulan</label
          >
          <select
            v-model="selectedMonth"
            @change="selectedDate = ''"
            class="w-full bg-slate-50 border-none rounded-xl py-2.5 px-3 text-sm font-medium text-slate-700 focus:ring-2 focus:ring-primary/20"
          >
            <option value="ALL">Semua Bulan</option>
            <option v-for="m in 12" :key="m" :value="m.toString()">
              {{ new Date(2024, m - 1, 1).toLocaleString('id-ID', { month: 'long' }) }}
            </option>
          </select>
        </div>

        <!-- Filter Tahun -->
        <div class="flex-1 min-w-[110px]">
          <label class="block text-[10px] font-bold text-slate-500 uppercase tracking-wider mb-1"
            >Tahun</label
          >
          <select
            v-model="selectedYear"
            @change="selectedDate = ''"
            class="w-full bg-slate-50 border-none rounded-xl py-2.5 px-3 text-sm font-medium text-slate-700 focus:ring-2 focus:ring-primary/20"
          >
            <option value="ALL">Semua Tahun</option>
            <option v-for="year in years" :key="year" :value="year">
              {{ year }}
            </option>
          </select>
        </div>

        <!-- Filter Harian -->
        <div class="flex-1 min-w-[130px]">
          <label class="block text-[10px] font-bold text-slate-500 uppercase tracking-wider mb-1"
            >Tanggal</label
          >
          <input
            type="date"
            v-model="selectedDate"
            @change="selectedMonth = 'ALL'; selectedYear = 'ALL'"
            class="w-full bg-slate-50 border-none rounded-xl py-2.5 px-3 text-sm font-medium text-slate-700 focus:ring-2 focus:ring-primary/20"
          />
        </div>

        <!-- Reset Filter -->
        <div class="flex items-end">
          <button
            @click="resetFilters"
            class="bg-red-50 text-red-600 hover:bg-red-100 px-4 py-2.5 rounded-xl font-bold text-xs tracking-wider transition-all h-[40px]"
          >
            RESET
          </button>
        </div>
      </div>
    </div>

    <!-- Alert Gagal Pengiriman -->
    <div v-if="failedDeliveriesCount > 0" class="mb-4 p-4 bg-red-50 border border-red-200 rounded-lg dark:bg-red-950/20 dark:border-red-900 flex items-start gap-3">
      <AlertCircle class="w-5 h-5 text-red-600 dark:text-red-400 shrink-0 mt-0.5" />
      <div>
        <h4 class="font-semibold text-red-800 dark:text-red-300 text-sm">
          Perhatian: Terdapat {{ failedDeliveriesCount }} Faktur dengan Pengiriman Gagal (Toko Tutup)
        </h4>
        <p class="text-xs text-red-700 dark:text-red-400 mt-1">
          Pesanan ini gagal dikirim oleh driver karena toko tutup. Mohon klik tombol <strong>Kirim Ulang</strong> pada daftar faktur di bawah untuk menjadwalkan ulang pengiriman dan mengembalikan pesanan ke menu SO.
        </p>
      </div>
    </div>

    <!-- Sales Table -->
    <div
      class="bg-white rounded-[32px] border border-slate-200 shadow-sm overflow-hidden flex flex-col"
    >
      <div class="overflow-x-auto no-scrollbar">
        <table class="w-full border-collapse text-left">
          <thead class="bg-slate-50/50 border-b border-slate-100">
            <tr>
              <th class="p-6 w-12 text-center">
                <input
                  type="checkbox"
                  :checked="isAllSelected"
                  :indeterminate="isSomeSelected"
                  @change="toggleSelectAll"
                  class="w-4 h-4 text-primary rounded border-slate-300 focus:ring-primary/20 cursor-pointer accent-primary"
                  title="Pilih Semua Faktur"
                />
              </th>
              <th class="p-6 text-[10px] font-black text-slate-400 uppercase tracking-widest">
                No. Nota
              </th>
              <th class="p-6 text-[10px] font-black text-slate-400 uppercase tracking-widest">
                Toko / Customer
              </th>
              <th class="p-6 text-[10px] font-black text-slate-400 uppercase tracking-widest">
                Salesman
              </th>
              <th class="p-6 text-[10px] font-black text-slate-400 uppercase tracking-widest">
                Tanggal
              </th>
              <th
                class="p-6 text-[10px] font-black text-slate-400 uppercase tracking-widest text-right"
              >
                Pembayaran & Omzet
              </th>
              <th
                class="p-6 text-[10px] font-black text-slate-400 uppercase tracking-widest text-center"
              >
                Lokasi Penerbitan
              </th>
              <th
                class="p-6 text-[10px] font-black text-slate-400 uppercase tracking-widest text-center"
              >
                Status
              </th>
              <th
                class="p-6 text-[10px] font-black text-slate-400 uppercase tracking-widest text-right"
              >
                Aksi
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-50">
            <tr v-if="isLoading" v-for="i in 5" :key="i" class="animate-pulse">
              <td v-for="j in 9" :key="j" class="p-6">
                <div class="h-4 bg-slate-100 rounded-lg w-full"></div>
              </td>
            </tr>

            <tr v-else-if="filteredSales.length === 0">
              <td colspan="9" class="py-24 text-center">
                <div class="flex flex-col items-center gap-3">
                  <ClipboardList class="w-12 h-12 text-slate-200" />
                  <p class="font-bold text-slate-400">Tidak ada transaksi ditemukan.</p>
                </div>
              </td>
            </tr>

            <tr
              v-for="sale in filteredSales"
              :key="sale.id"
              :class="[
                'group transition-all',
                selectedSaleIds.has(sale.id) ? 'bg-indigo-50/40 hover:bg-indigo-50/60' : (
                  sale.status === 'PENDING' && sale.notaImage && !sale.notaImage.includes('placehold.co') && (!sale.rawItems || sale.rawItems.length === 0)
                    ? 'bg-amber-50/45 hover:bg-amber-50/70 border-l-4 border-l-amber-500'
                    : 'hover:bg-slate-50/50'
                )
              ]"
            >
              <td class="p-6 w-12 text-center" @click.stop>
                <input
                  type="checkbox"
                  :checked="selectedSaleIds.has(sale.id)"
                  @change="toggleSelectSale(sale.id)"
                  class="w-4 h-4 text-primary rounded border-slate-300 focus:ring-primary/20 cursor-pointer accent-primary"
                />
              </td>
              <td class="p-6">
                <div class="flex items-center gap-3">
                  <div>
                    <div class="flex items-center gap-2">
                      <p class="text-[12px] font-black text-slate-900 leading-none">
                        {{ sale.receiptNo || 'MANUAL' }}
                      </p>
                      <span
                        v-if="sale.raw?.sales_channel === 'DIRECT_FACTORY' || sale.raw?.sales_channel === 'DIRECT_BRANCH'"
                        class="inline-flex items-center text-[8px] font-black px-1.5 py-0.5 bg-emerald-100 text-emerald-800 rounded-md uppercase tracking-wider shrink-0"
                      >
                        DIRECT PABRIK
                      </span>
                      <span
                        v-else-if="sale.raw?.sales_channel === 'EMPLOYEE_SALE' || sale.raw?.buyer_type === 'EMPLOYEE'"
                        class="inline-flex items-center text-[8px] font-black px-1.5 py-0.5 bg-indigo-100 text-indigo-800 rounded-md uppercase tracking-wider shrink-0"
                      >
                        KARYAWAN
                      </span>
                      <span
                        v-if="sale.returns && sale.returns.length > 0"
                        class="inline-flex items-center gap-1 text-[8px] font-black px-1.5 py-0.5 bg-rose-100 text-rose-800 rounded-md uppercase tracking-wider shrink-0"
                        :title="`Terdapat ${sale.returns.length} Pengajuan Retur Barang`"
                      >
                        <RotateCcw class="w-2.5 h-2.5" /> {{ sale.returns.length }} RETUR
                      </span>
                      <span
                        v-if="sale.status === 'PENDING' && sale.notaImage && !sale.notaImage.includes('placehold.co') && (!sale.rawItems || sale.rawItems.length === 0)"
                        class="inline-flex items-center gap-1 text-[8px] font-black px-1.5 py-0.5 bg-amber-500 text-white rounded-md uppercase tracking-wider animate-pulse shrink-0"
                        title="Nota fisik belum diinput rincian barangnya"
                      >
                        <AlertCircle class="w-2.5 h-2.5" /> BUTUH INPUT
                      </span>
                      <span
                        v-if="getDiscrepancyStatus(sale) === 'PENDING'"
                        class="inline-flex items-center gap-1 text-[8px] font-black px-1.5 py-0.5 bg-rose-500 text-white rounded-md uppercase tracking-wider shrink-0"
                        title="Ada selisih setoran kasir, silakan periksa & sesuaikan laporan surat jalan"
                      >
                        <AlertTriangle class="w-2.5 h-2.5" /> SELISIH (KASIR)
                      </span>
                      <span
                        v-else-if="getDiscrepancyStatus(sale) === 'RESOLVED'"
                        class="inline-flex items-center gap-1 text-[8px] font-black px-1.5 py-0.5 bg-emerald-500 text-white rounded-md uppercase tracking-wider shrink-0"
                        title="Selisih sudah disesuaikan, menunggu kasir menyelesaikan kasus"
                      >
                        <CheckCircle2 class="w-2.5 h-2.5" /> SELISIH DISELESAIKAN
                      </span>
                    </div>
                    <p class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-tighter">
                      {{ sale.id.split('-')[0] }}
                    </p>
                  </div>
                </div>
              </td>
              <td class="p-6">
                <div class="flex items-center gap-2">
                  <Store class="w-4 h-4 text-slate-300" />
                  <div>
                    <div class="flex items-center gap-1.5 flex-wrap">
                      <p class="text-[12px] font-black text-slate-900 leading-none">{{ sale.store }}</p>
                      <span
                        v-if="getBranchBadge(sale.store)"
                        class="inline-flex items-center text-[9px] font-black px-1.5 py-0.5 bg-indigo-50 text-indigo-700 rounded-md tracking-wider border border-indigo-200 shrink-0"
                        :title="`Faktur Penjualan Pabrik ke Cabang ${getBranchBadge(sale.store)}`"
                      >
                        ➔ {{ getBranchBadge(sale.store) }}
                      </span>
                    </div>
                    <p class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-tighter">
                      {{ sale.storeCategory }} • {{ sale.rawItems?.length || 0 }} SKU
                      <span v-if="sale.raw?.bonus_items?.length" class="text-emerald-500 font-black ml-1">
                        + {{ sale.raw?.bonus_items?.length }} BONUS
                      </span>
                    </p>
                  </div>
                </div>
              </td>
              <td class="p-6">
                <div class="flex items-center gap-2">
                  <User class="w-4 h-4 text-slate-300" />
                  <p class="text-sm font-bold text-slate-700">{{ sale.salesman }}</p>
                </div>
              </td>
              <td class="p-6">
                <div class="flex items-center gap-2">
                  <Calendar class="w-4 h-4 text-slate-300" />
                  <p class="text-[12px] font-bold text-slate-700">{{ sale.date }}</p>
                </div>
              </td>
              <td class="p-6">
                <div class="flex flex-col items-end gap-1.5 min-w-[170px]">
                  <p class="text-sm font-black text-slate-900 leading-none">
                    {{ formatCurrency(sale.amount) }}
                  </p>
                  <span v-if="Number(sale.raw?.paid_amount || 0) > Number(sale.raw?.total_amount || 0)" class="text-[9px] font-bold text-rose-700 bg-rose-50 px-1.5 py-0.5 rounded border border-rose-200">
                    Kelebihan bayar: {{ formatCurrency(Number(sale.raw.paid_amount) - Number(sale.raw.total_amount)) }} · Perlu penyelesaian kasir
                  </span>
                  <span
                    v-if="sale.returnsTotal > 0"
                    class="text-[9px] font-bold text-rose-600 bg-rose-50 px-1.5 py-0.5 rounded border border-rose-100 flex items-center gap-1"
                    :title="`Total nilai retur potong nota: Rp ${sale.returnsTotal}`"
                  >
                    <RotateCcw class="w-2.5 h-2.5" /> -{{ formatCurrency(sale.returnsTotal) }} (Retur)
                  </span>
                  <div class="flex items-center gap-1.5 mt-0.5 flex-wrap justify-end">
                    <span
                      class="text-[9px] font-black px-1.5 py-0.5 rounded uppercase border"
                      :class="sale.raw?.payment_method === 'TEMPO' ? 'bg-amber-50 text-amber-800 border-amber-200' : (sale.raw?.payment_method === 'TRANSFER' ? 'bg-indigo-50 text-indigo-700 border-indigo-200' : 'bg-slate-50 text-slate-600 border-slate-200')"
                    >
                      {{ sale.raw?.payment_method === 'TEMPO' ? getTempoProgressInfo(sale).tempoMethodLabel : (sale.raw?.payment_method || 'CASH') }}
                    </span>
                    <span
                      :class="[
                        'text-[9px] font-black px-1.5 py-0.5 rounded uppercase border',
                        sale.raw?.payment_status === 'PAID'
                          ? 'bg-emerald-50 text-emerald-600 border-emerald-100'
                          : (sale.raw?.payment_status === 'PARTIAL' ? 'bg-amber-50 text-amber-700 border-amber-200' : 'bg-slate-50 text-slate-400 border-slate-100'),
                      ]"
                    >
                      {{ sale.raw?.payment_status || 'UNPAID' }}
                    </span>
                    <button
                      v-if="sale.raw?.payments?.some((p: any) => p.transfer_proof_url) || sale.raw?.transfer_proof_url"
                      @click.stop="openImagePreview(sale.raw?.payments?.find((p: any) => p.transfer_proof_url)?.transfer_proof_url || sale.raw?.transfer_proof_url)"
                      type="button"
                      class="text-[9px] font-black px-1.5 py-0.5 rounded uppercase border bg-indigo-50 text-indigo-600 border-indigo-200 hover:bg-indigo-100 flex items-center gap-1 cursor-pointer transition-all"
                      title="Lihat Bukti Transfer"
                    >
                      <ExternalLink class="w-2.5 h-2.5" /> Bukti
                    </button>
                  </div>
                  <!-- TEMPO Progress & Step Tracking -->
                  <div v-if="isTempoEligibleForPayment(sale) || String(sale.raw?.payment_method || '').toUpperCase().includes('TEMPO')" class="w-full mt-1.5 bg-slate-50 p-2 rounded-xl border border-slate-100 space-y-1">
                    <div class="flex items-center justify-between gap-2">
                      <div class="flex items-center gap-1.5">
                        <span
                          class="text-[8px] font-black px-1.5 py-0.5 rounded uppercase border"
                          :class="getTempoProgressInfo(sale).isPaid ? 'bg-emerald-100 text-emerald-800 border-emerald-300' : (getTempoProgressInfo(sale).isOverdue ? 'bg-rose-100 text-rose-700 border-rose-300 animate-pulse' : 'bg-amber-100 text-amber-800 border-amber-200')"
                        >
                          {{ getTempoProgressInfo(sale).isPaid ? '✓ LUNAS' : (getTempoProgressInfo(sale).isOverdue ? '⚠ JATUH TEMPO' : getTempoProgressInfo(sale).stepText) }}
                        </span>
                        <button
                          v-if="isTempoEligibleForPayment(sale)"
                          type="button"
                          @click.stop="openTempoPaymentDialogForSale(sale)"
                          class="text-[8px] font-black px-1.5 py-0.5 rounded bg-indigo-50 text-indigo-700 border border-indigo-200 hover:bg-indigo-600 hover:text-white transition-all flex items-center gap-1 cursor-pointer"
                          title="Klik untuk langsung mencatat pembayaran cicilan tempo"
                        >
                          <Wallet class="w-2.5 h-2.5" /> + Bayar
                        </button>
                      </div>
                      <span class="text-[9px] font-black text-slate-700">
                        {{ formatCurrency(getTempoProgressInfo(sale).paidAmount) }} <span class="text-slate-400 font-normal">/ {{ formatCurrency(getTempoProgressInfo(sale).totalAmount) }}</span>
                      </span>
                    </div>
                    <div class="w-full bg-slate-200 rounded-full h-1.5 overflow-hidden">
                      <div
                        class="h-1.5 rounded-full transition-all duration-500"
                        :class="getTempoProgressInfo(sale).isPaid ? 'bg-emerald-500' : (getTempoProgressInfo(sale).isOverdue ? 'bg-rose-500' : 'bg-amber-500')"
                        :style="{ width: getTempoProgressInfo(sale).percent + '%' }"
                      ></div>
                    </div>
                    <div class="flex items-center justify-between text-[8px] font-bold text-slate-400">
                      <span v-if="!getTempoProgressInfo(sale).isPaid && getTempoProgressInfo(sale).targetTenor > 1" class="text-amber-700 font-extrabold">
                        Sisa: {{ formatCurrency(getTempoProgressInfo(sale).remainingAmount) }} ({{ getTempoProgressInfo(sale).remainingTermsCount }}x cicilan lagi)
                      </span>
                      <span v-else-if="!getTempoProgressInfo(sale).isPaid" class="text-slate-500">
                        Sisa: {{ formatCurrency(getTempoProgressInfo(sale).remainingAmount) }}
                      </span>
                      <span v-else class="text-emerald-600 font-extrabold">
                        {{ getTempoProgressInfo(sale).detailedStepText }}
                      </span>
                      <span v-if="!getTempoProgressInfo(sale).isPaid && sale.raw?.payment_due_date" :class="getTempoProgressInfo(sale).isOverdue ? 'text-rose-600 font-extrabold' : 'text-slate-500'">
                        {{ getTempoProgressInfo(sale).isOverdue ? 'Lewat: ' : 'Due: ' }}{{ formatDate(sale.raw.payment_due_date) }}
                      </span>
                    </div>
                  </div>
                </div>
              </td>
              <td class="p-6">
                <div v-if="getOrderLocationStatus(sale.raw).status !== 'UNKNOWN'" class="flex justify-center items-center">
                  <Badge 
                    :class="`${getOrderLocationStatus(sale.raw).bg} ${getOrderLocationStatus(sale.raw).color} border-0 hover:bg-opacity-80 cursor-pointer`"
                    @click="openMapModal(sale.raw)"
                    title="Klik untuk memverifikasi lokasi di Peta"
                  >
                    <component :is="getOrderLocationStatus(sale.raw).icon" class="w-3 h-3 mr-1" />
                    {{ getOrderLocationStatus(sale.raw).text }}
                  </Badge>
                </div>
                <div v-else class="flex justify-center text-xs text-slate-400 font-medium">Tidak ada data</div>
              </td>
              <td class="p-6">
                <div class="flex flex-col items-center">
                  <span
                    :class="[
                      'text-[10px] font-black px-3 py-1 rounded-full uppercase tracking-wider border',
                      sale.status === 'DELETED'
                        ? 'bg-rose-50 text-rose-700 border-rose-200'
                        : sale.status === 'PENDING'
                          ? 'bg-amber-50 text-amber-600 border-amber-100'
                          : sale.status === 'APPROVED'
                            ? 'bg-emerald-50 text-emerald-600 border-emerald-100'
                            : 'bg-red-50 text-red-600 border-red-100',
                    ]"
                  >
                    {{ sale.status === 'DELETED' ? 'TERHAPUS' : sale.status === 'CANCELLED' ? 'DIBATALKAN' : sale.status }}
                  </span>
                  <div v-if="sale.status === 'APPROVED' && sale.raw?.verified_by" class="mt-1 flex flex-col items-center text-[10px] text-slate-500 leading-tight">
                    <span class="font-bold text-slate-700">✓ {{ sale.raw.verified_by.first_name }} {{ sale.raw.verified_by.last_name || '' }}</span>
                    <span v-if="sale.raw.verified_by.job_position?.title" class="text-[9px] text-slate-400 font-medium">({{ sale.raw.verified_by.job_position.title }})</span>
                  </div>
                  <div v-if="sale.deletedAt" class="mt-1 text-[9px] text-rose-600 text-center max-w-[140px] truncate" :title="`Dihapus: ${sale.deletedBy ? sale.deletedBy.first_name : 'Admin'}\nAlasan: ${sale.deletionReason || '-'}`">
                    {{ sale.deletionReason || 'Dihapus' }}
                  </div>
                </div>
              </td>
              <td class="p-6 text-right">
                <div class="flex justify-end gap-1 items-center">
                  <!-- Restore button if deleted -->
                  <button
                    v-if="sale.deletedAt || sale.status === 'DELETED'"
                    @click="handleRestore(sale)"
                    class="px-3 py-1.5 bg-emerald-100 hover:bg-emerald-200 text-emerald-700 rounded-xl transition-all text-xs font-bold flex items-center gap-1 shadow-sm"
                    title="Pulihkan Faktur Ini"
                  >
                    <RotateCcw class="w-3.5 h-3.5" />
                    <span>Pulihkan</span>
                  </button>

                  <!-- Approval Actions for Pending (only if not deleted) -->
                  <template v-else-if="sale.status === 'PENDING'">
                    <button
                      v-if="hasFailedDelivery(sale)"
                      @click="openCancelDeliveryModal(sale)"
                      class="px-2.5 py-1.5 mr-1 bg-amber-50 hover:bg-amber-100 border border-amber-200 rounded-xl transition-all text-amber-700 text-xs font-bold flex items-center gap-1.5"
                      title="Kembalikan nota ke Manajemen Pengiriman"
                    >
                      <RotateCcw class="w-3.5 h-3.5" />
                      <span>Gagal Kirim</span>
                    </button>
                    
                    <button
                      v-if="!hasFailedDelivery(sale)"
                      @click="handleConvertToSO(sale)"
                      class="px-2.5 py-1.5 bg-blue-50 hover:bg-blue-100 text-blue-700 rounded-xl transition-all text-xs font-bold flex items-center gap-1 border border-blue-200 shrink-0"
                      title="Ubah Faktur Pending Ini Menjadi Pesanan SO (Sales Order) Agar Bisa Dijadwalkan Pengirimannya"
                    >
                      <ArrowRightLeft class="w-3.5 h-3.5" />
                      <span class="hidden sm:inline">Ubah ke SO</span>
                    </button>
                    <button
                      v-if="!hasFailedDelivery(sale)"
                      @click="handleApprove(sale.id)"
                      class="p-2 hover:bg-emerald-50 rounded-xl transition-all text-emerald-500"
                      title="Setujui"
                    >
                      <CheckCircle2 class="w-5 h-5" />
                    </button>
                    <button
                      v-if="!hasFailedDelivery(sale)"
                      @click="handleReject(sale.id)"
                      class="p-2 hover:bg-red-50 rounded-xl transition-all text-red-500"
                      title="Tolak"
                    >
                      <XCircle class="w-5 h-5" />
                    </button>
                  </template>

                  <template v-if="!sale.deletedAt && sale.status !== 'DELETED' && sale.status !== 'CANCELLED'">
                    <!-- Shortcut Bayar Cicilan (Ikon Dompet) untuk Faktur Tempo Belum Lunas -->
                    <button
                      v-if="isTempoEligibleForPayment(sale)"
                      @click="openTempoPaymentDialogForSale(sale)"
                      class="px-2.5 py-1.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl transition-all text-xs font-black flex items-center gap-1.5 shadow-md shadow-indigo-200 cursor-pointer shrink-0"
                      :title="`Catat Pembayaran Cicilan Tempo (${getTempoProgressInfo(sale).stepText})`"
                    >
                      <Wallet class="w-4 h-4" />
                      <span class="hidden sm:inline">Bayar</span>
                    </button>

                    <button
                      @click="openSalesmanLPS(sale)"
                      class="p-2 hover:bg-indigo-50 rounded-xl transition-all text-slate-400 hover:text-indigo-600"
                      title="Lihat LPS Salesman untuk Faktur Ini"
                    >
                      <FileText class="w-5 h-5" />
                    </button>
                    <button
                      v-if="sale.raw?.sales_channel === 'DIRECT_FACTORY' || sale.raw?.sales_channel === 'DIRECT_BRANCH' || sale.raw?.sales_channel === 'EMPLOYEE_SALE'"
                      @click="openDirectSalePrint(sale)"
                      class="p-2 hover:bg-emerald-50 rounded-xl transition-all text-emerald-600 hover:text-emerald-700"
                      title="Cetak Faktur Penjualan Langsung"
                    >
                      <Printer class="w-5 h-5" />
                    </button>
                    <button
                      v-else
                      @click="openNotaPreview(sale)"
                      class="p-2 hover:bg-slate-100 rounded-xl transition-all text-slate-400 hover:text-primary"
                      title="Cetak Nota"
                    >
                      <Printer class="w-5 h-5" />
                    </button>
                    <button
                      v-if="sale.raw?.sales_channel === 'DIRECT_FACTORY' || sale.raw?.sales_channel === 'DIRECT_BRANCH' || sale.raw?.sales_channel === 'EMPLOYEE_SALE' || authStore.user?.role === 'admin_nota' || authStore.user?.role === 'superadmin' || authStore.user?.role === 'sales_supervisor' || authStore.user?.role === 'manager'"
                      @click="openDirectSaleReturn(sale)"
                      class="p-2 hover:bg-red-50 rounded-xl transition-all text-red-500 hover:text-red-700"
                      title="Ajukan Retur Barang"
                    >
                      <RotateCcw class="w-5 h-5" />
                    </button>
                    <button
                      v-if="canEdit(sale)"
                      @click="handleEdit(sale)"
                      class="p-2 hover:bg-slate-100 rounded-xl transition-all text-slate-400 hover:text-amber-500"
                      title="Edit"
                    >
                      <Pencil class="w-5 h-5" />
                    </button>
                    <button
                      v-if="canDelete(sale)"
                      @click="handleDelete(sale)"
                      class="p-2 hover:bg-red-50 rounded-xl transition-all text-red-400 hover:text-red-600"
                      title="Hapus (Soft Delete)"
                    >
                      <Trash2 class="w-5 h-5" />
                    </button>
                  </template>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination UI -->
      <div
        class="p-4 border-t border-slate-100 flex flex-col sm:flex-row items-center justify-between gap-4"
      >
        <div class="flex items-center gap-3">
          <div class="text-xs font-medium text-slate-500">
            Menampilkan
            <span v-if="itemsPerPage === 'all'">Semua</span>
            <span v-else
              >{{ (currentPage - 1) * (itemsPerPage as number) + 1 }} -
              {{ Math.min(currentPage * (itemsPerPage as number), totalItems) }}</span
            >
            dari {{ totalItems }} transaksi
          </div>
          <select
            v-model="itemsPerPage"
            class="text-xs bg-slate-50 border border-slate-200 rounded-lg px-2 py-1 outline-none focus:ring-2 focus:ring-primary/20 text-slate-600 font-medium"
          >
            <option :value="10">10 / hal</option>
            <option :value="25">25 / hal</option>
            <option :value="50">50 / hal</option>
            <option :value="100">100 / hal</option>
            <option value="all">Semua Data</option>
          </select>
        </div>

        <div class="flex items-center gap-2" v-if="totalPages > 1">
          <button
            @click="currentPage > 1 ? currentPage-- : null"
            :disabled="currentPage === 1"
            class="w-8 h-8 flex items-center justify-center rounded-lg border border-slate-200 text-slate-600 hover:bg-slate-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            &laquo;
          </button>
          <div class="flex items-center gap-1">
            <button
              v-for="page in totalPages"
              :key="page"
              @click="currentPage = page"
              :class="[
                'w-8 h-8 flex items-center justify-center rounded-lg text-xs font-bold transition-colors',
                currentPage === page
                  ? 'bg-primary text-white border border-primary'
                  : 'border border-slate-200 text-slate-600 hover:bg-slate-50',
              ]"
            >
              {{ page }}
            </button>
          </div>
          <button
            @click="currentPage < totalPages ? currentPage++ : null"
            :disabled="currentPage === totalPages"
            class="w-8 h-8 flex items-center justify-center rounded-lg border border-slate-200 text-slate-600 hover:bg-slate-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            &raquo;
          </button>
        </div>
      </div>
    </div>

    <!-- Floating Bulk Action Toolbar -->
    <transition
      enter-active-class="transition-all duration-300 ease-out"
      enter-from-class="opacity-0 translate-y-8 scale-95"
      enter-to-class="opacity-100 translate-y-0 scale-100"
      leave-active-class="transition-all duration-200 ease-in"
      leave-from-class="opacity-100 translate-y-0 scale-100"
      leave-to-class="opacity-0 translate-y-8 scale-95"
    >
      <div
        v-if="selectedSaleIds.size > 0"
        class="fixed bottom-6 left-1/2 -translate-x-1/2 z-40 bg-slate-900/95 text-white px-6 py-4 rounded-2xl shadow-2xl backdrop-blur-md border border-slate-700/60 flex flex-wrap items-center gap-5 max-w-4xl w-[92vw] sm:w-auto justify-between"
      >
        <div class="flex items-center gap-3">
          <div class="w-8 h-8 rounded-full bg-primary flex items-center justify-center font-black text-xs text-white shadow-sm">
            {{ selectedSaleIds.size }}
          </div>
          <div>
            <p class="text-[11px] font-bold text-slate-300 uppercase tracking-wider">Faktur Terpilih</p>
            <p class="text-sm font-black text-white">
              Total: {{ formatCurrency(selectedSalesTotalAmount) }}
            </p>
          </div>
        </div>

        <div class="flex items-center gap-2.5">
          <button
            @click="clearSelection"
            class="px-3.5 py-2 text-xs font-bold text-slate-300 hover:text-white hover:bg-slate-800 rounded-xl transition-all"
          >
            Batal
          </button>
          
          <button
            @click="openBulkPrint"
            class="flex items-center gap-2 px-5 py-2.5 bg-primary hover:bg-primary/90 text-white text-xs sm:text-sm font-black rounded-xl transition-all shadow-lg shadow-primary/30"
          >
            <Printer class="w-4 h-4" />
            <span>Cetak Masal ({{ selectedSaleIds.size }} Faktur)</span>
          </button>
        </div>
      </div>
    </transition>

    </div>

    <!-- TAB 2: Faktur Pengiriman Cabang (Pusat / Pabrik ke Cabang) -->
    <div v-if="activeMainInvoiceTab === 'cabang'" class="space-y-6">
      <!-- Header & Actions -->
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 class="text-xl sm:text-2xl font-extrabold text-slate-900 tracking-tight flex items-center gap-2">
            <span>Faktur Pengiriman Cabang (Tagihan Pabrik)</span>
            <span class="text-xs bg-indigo-100 text-indigo-800 font-bold px-2.5 py-0.5 rounded-full">
              Inter-Branch Billing
            </span>
          </h2>
          <p class="text-sm font-medium text-slate-500 mt-1">
            Kelola faktur pengiriman stok barang dari Pabrik Pusat ke Gudang Cabang. Nilai faktur tercatat otomatis ke Piutang Cabang.
          </p>
        </div>

        <div class="flex flex-wrap items-center gap-2 sm:gap-3">
          <RefreshDataButton
            size="sm"
            :loading="isRefreshing || isLoadingBranchTransfers"
            :last-updated="lastRefreshedTimeStr"
            @click="triggerRefresh()"
          />
          <Button
            @click="openBranchCreateModal"
            class="bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl px-4 py-2 text-xs sm:text-sm font-bold shadow-md shadow-indigo-600/20 flex items-center gap-2"
          >
            <Plus class="w-4 h-4" /> + Buat Faktur Cabang
          </Button>
        </div>
      </div>

      <!-- Summary Metric Cards -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div class="p-4 bg-white rounded-2xl border border-slate-200/80 shadow-sm space-y-1">
          <span class="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Total Faktur Cabang</span>
          <div class="text-xl font-black text-slate-900">{{ branchStats.totalInvoices }} Pengiriman</div>
          <p class="text-[11px] text-slate-500">Akumulasi pengiriman cabang</p>
        </div>

        <div class="p-4 bg-white rounded-2xl border border-slate-200/80 shadow-sm space-y-1">
          <span class="text-[11px] font-bold text-amber-600 uppercase tracking-wider">Menunggu Persetujuan / Kirim</span>
          <div class="text-xl font-black text-amber-600">{{ branchStats.pendingCount }} Faktur</div>
          <p class="text-[11px] text-slate-500">Status REQUESTED / APPROVED</p>
        </div>

        <div class="p-4 bg-white rounded-2xl border border-slate-200/80 shadow-sm space-y-1">
          <span class="text-[11px] font-bold text-indigo-600 uppercase tracking-wider">Dalam Perjalanan (Kirim)</span>
          <div class="text-xl font-black text-indigo-600">{{ branchStats.shippedCount }} Armada</div>
          <p class="text-[11px] text-slate-500">Status SHIPPED / ARRIVED</p>
        </div>

        <div class="p-4 bg-white rounded-2xl border border-slate-200/80 shadow-sm space-y-1">
          <span class="text-[11px] font-bold text-emerald-600 uppercase tracking-wider">Total Nilai Tagihan Cabang</span>
          <div class="text-xl font-black text-emerald-700">{{ formatCurrency(branchStats.totalAmount) }}</div>
          <p class="text-[11px] text-slate-500">Nilai total piutang barang cabang</p>
        </div>
      </div>

      <!-- Filters -->
      <div class="flex flex-col sm:flex-row gap-3">
        <div class="relative flex-1">
          <Search class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input
            v-model="branchSearchQuery"
            type="text"
            placeholder="Cari nomor faktur / DO, cabang, produk..."
            class="w-full pl-10 pr-4 py-2.5 bg-white border border-slate-200 rounded-xl text-xs font-medium text-slate-800 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500"
          />
        </div>

        <div class="flex gap-2">
          <select
            v-model="branchTargetFilter"
            class="bg-white border border-slate-200 rounded-xl px-3 py-2 text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-indigo-500"
          >
            <option value="ALL">Semua Cabang Tujuan</option>
            <option v-for="b in masterDataStore.branches" :key="b.id" :value="b.id">{{ b.name }}</option>
          </select>

          <select
            v-model="branchStatusFilter"
            class="bg-white border border-slate-200 rounded-xl px-3 py-2 text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-indigo-500"
          >
            <option value="ALL">Semua Status</option>
            <option value="REQUESTED">REQUESTED (Permintaan)</option>
            <option value="APPROVED">APPROVED (Disetujui)</option>
            <option value="SHIPPED">SHIPPED (Dikirim)</option>
            <option value="RECEIVED">RECEIVED (Diterima Cabang)</option>
          </select>
        </div>
      </div>

      <!-- Table of Branch Invoices -->
      <div class="bg-white border border-slate-200/80 rounded-2xl shadow-sm overflow-hidden">
        <div v-if="isLoadingBranchTransfers" class="p-12 text-center text-slate-400 font-medium text-xs">
          <RefreshCw class="w-6 h-6 animate-spin mx-auto mb-2 text-indigo-500" />
          Memuat data faktur pengiriman cabang...
        </div>

        <div v-else-if="filteredBranchTransfers.length === 0" class="p-12 text-center text-slate-400 text-xs">
          Tidak ada data faktur pengiriman cabang yang sesuai filter.
        </div>

        <div v-else class="overflow-x-auto">
          <table class="w-full text-left text-xs border-collapse">
            <thead class="bg-slate-50 text-slate-600 uppercase font-black text-[11px] border-b border-slate-200">
              <tr>
                <th class="p-3.5 pl-5">No. Faktur / DO</th>
                <th class="p-3.5">Pabrik & Cabang</th>
                <th class="p-3.5">Rincian Muatan Produk</th>
                <th class="p-3.5 text-right">Nilai Tagihan (Rp)</th>
                <th class="p-3.5">Metode Bayar</th>
                <th class="p-3.5 text-center">Status</th>
                <th class="p-3.5 pr-5 text-right">Aksi</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 font-medium">
              <tr v-for="t in filteredBranchTransfers" :key="t.id" class="hover:bg-slate-50/80 transition-colors">
                <td class="p-3.5 pl-5">
                  <div class="font-extrabold text-slate-900 text-sm flex items-center gap-1.5">
                    <Truck class="w-3.5 h-3.5 text-indigo-600" />
                    {{ t.delivery_order_no }}
                  </div>
                  <div class="text-[11px] text-slate-400 mt-0.5">
                    Tgl: {{ t.shipped_at?.split('T')[0] || t.created_at?.split('T')[0] }}
                  </div>
                </td>

                <td class="p-3.5">
                  <div class="font-bold text-slate-900">{{ t.to_branch?.name || 'Cabang Tujuan' }}</div>
                  <div class="text-[11px] text-slate-500">Dari: {{ t.from_factory?.name || 'Pabrik Pusat' }}</div>
                </td>

                <td class="p-3.5">
                  <div class="font-bold text-slate-800">
                    {{ t.items?.length || 1 }} Produk ({{ t.total_quantity }} Unit)
                  </div>
                  <div class="text-[11px] text-slate-500 truncate max-w-xs">
                    {{ t.items?.map((it: any) => `${it.product?.name || 'Produk'} (${it.quantity})`).join(', ') }}
                  </div>
                </td>

                <td class="p-3.5 text-right font-extrabold text-emerald-700 text-sm">
                  {{ formatCurrency(t.total_amount) }}
                </td>

                <td class="p-3.5">
                  <span class="font-bold text-slate-700">{{ t.payment_method || 'NET' }}</span>
                  <div v-if="t.payment_due_at" class="text-[10px] text-slate-400">
                    Tempo: {{ t.payment_due_at?.split('T')[0] }}
                  </div>
                </td>

                <td class="p-3.5 text-center">
                  <span
                    v-if="t.status === 'REQUESTED'"
                    class="px-2.5 py-1 bg-amber-50 text-amber-700 border border-amber-200 rounded-full font-bold text-[10px]"
                  >
                    REQUESTED
                  </span>
                  <span
                    v-else-if="t.status === 'APPROVED'"
                    class="px-2.5 py-1 bg-blue-50 text-blue-700 border border-blue-200 rounded-full font-bold text-[10px]"
                  >
                    APPROVED
                  </span>
                  <span
                    v-else-if="t.status === 'SHIPPED'"
                    class="px-2.5 py-1 bg-indigo-50 text-indigo-700 border border-indigo-200 rounded-full font-bold text-[10px]"
                  >
                    SHIPPED
                  </span>
                  <span
                    v-else-if="t.status === 'RECEIVED'"
                    class="px-2.5 py-1 bg-emerald-50 text-emerald-700 border border-emerald-200 rounded-full font-bold text-[10px]"
                  >
                    RECEIVED
                  </span>
                  <span v-else class="px-2.5 py-1 bg-slate-100 text-slate-700 rounded-full font-bold text-[10px]">
                    {{ t.status }}
                  </span>
                </td>

                <td class="p-3.5 pr-5 text-right space-x-1.5">
                  <Button
                    v-if="canEditBranchTransfer(t)"
                    variant="outline"
                    size="sm"
                    @click="openBranchEditModal(t)"
                    class="rounded-xl border-amber-200 text-amber-700 hover:bg-amber-50 text-xs font-bold px-2.5 py-1"
                    title="Edit Faktur Pengiriman Cabang"
                  >
                    <Pencil class="w-3.5 h-3.5 mr-1 text-amber-500" /> Edit
                  </Button>

                  <Button
                    variant="outline"
                    size="sm"
                    @click="openBranchReturnModal(t)"
                    class="rounded-xl border-rose-200 text-rose-700 hover:bg-rose-50 text-xs font-bold px-2.5 py-1"
                    title="Input Retur / Barang Rusak dari Faktur Ini"
                  >
                    <RotateCcw class="w-3.5 h-3.5 mr-1 text-rose-600" /> Retur
                  </Button>

                  <Button
                    variant="outline"
                    size="sm"
                    @click="openBranchInvoicePreview(t)"
                    class="rounded-xl border-slate-200 text-slate-700 hover:bg-slate-100 text-xs font-bold px-2.5 py-1"
                    title="Cetak Faktur Pengiriman Cabang"
                  >
                    <Printer class="w-3.5 h-3.5 mr-1 text-slate-500" /> Cetak
                  </Button>

                  <Button
                    v-if="t.status === 'REQUESTED' || t.status === 'APPROVED'"
                    size="sm"
                    @click="handleShipTransfer(t)"
                    class="bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl text-xs font-bold px-3 py-1"
                    title="Eksekusi Pengiriman & Potong Stok Pabrik"
                  >
                    <Truck class="w-3.5 h-3.5 mr-1" /> Kirim
                  </Button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination UI for Cabang -->
        <div
          class="p-4 border-t border-slate-100 flex flex-col sm:flex-row items-center justify-between gap-4"
        >
          <div class="flex items-center gap-3">
            <div class="text-xs font-medium text-slate-500">
              Menampilkan
              <span v-if="branchItemsPerPage === 'all'">Semua</span>
              <span v-else
                >{{ (branchCurrentPage - 1) * (branchItemsPerPage as number) + 1 }} -
                {{ Math.min(branchCurrentPage * (branchItemsPerPage as number), branchTotalItems) }}</span
              >
              dari {{ branchTotalItems }} faktur pengiriman
            </div>
            <select
              v-model="branchItemsPerPage"
              class="text-xs bg-slate-50 border border-slate-200 rounded-lg px-2 py-1 outline-none focus:ring-2 focus:ring-indigo-500/20 text-slate-600 font-medium"
            >
              <option :value="10">10 / hal</option>
              <option :value="25">25 / hal</option>
              <option :value="50">50 / hal</option>
              <option value="all">Semua</option>
            </select>
          </div>

          <div v-if="branchTotalPages > 1" class="flex items-center gap-1">
            <button
              @click="branchCurrentPage--"
              :disabled="branchCurrentPage === 1"
              class="px-2.5 py-1 text-xs font-bold text-slate-500 hover:text-slate-800 disabled:opacity-30 disabled:hover:text-slate-500 transition-colors"
            >
              &laquo;
            </button>
            <template v-for="(p, idx) in branchVisiblePages" :key="idx">
              <button
                v-if="p !== '...'"
                @click="branchCurrentPage = Number(p)"
                :class="[
                  'w-7 h-7 flex items-center justify-center rounded-lg text-xs font-bold transition-all',
                  branchCurrentPage === p
                    ? 'bg-indigo-600 text-white shadow-sm'
                    : 'text-slate-600 hover:bg-slate-100',
                ]"
              >
                {{ p }}
              </button>
              <span v-else class="px-1 text-xs text-slate-400">...</span>
            </template>
            <button
              @click="branchCurrentPage++"
              :disabled="branchCurrentPage === branchTotalPages"
              class="px-2.5 py-1 text-xs font-bold text-slate-500 hover:text-slate-800 disabled:opacity-30 disabled:hover:text-slate-500 transition-colors"
            >
              &raquo;
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- CRUD Modal -->
    <div
      v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-0 sm:p-4">
      <div
        class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
        @click="showModal = false"
      ></div>
      <div
        :class="[
          'bg-white rounded-none sm:rounded-[32px] w-full relative z-10 shadow-2xl flex flex-col h-[100dvh] sm:h-auto sm:max-h-[90vh] overflow-hidden animate-in fade-in zoom-in duration-300 transition-all duration-300',
          currentSale.receiptImageURL ? 'max-w-6xl' : 'max-w-2xl sm:max-w-3xl'
        ]"
      >
        <!-- Header -->
        <div
          class="px-4 sm:px-6 py-3.5 sm:py-5 border-b border-slate-100 flex items-center justify-between bg-white shrink-0"
        >
          <div>
            <h3 class="text-lg sm:text-xl font-black text-slate-900 tracking-tight leading-tight">
              {{ isEditing ? 'Edit Transaksi' : 'Tambah Transaksi' }}
            </h3>
            <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest mt-0.5">
              Laporan Penjualan
            </p>
          </div>
          <button
            @click="showModal = false"
            class="p-2 hover:bg-slate-100 rounded-xl transition-all text-slate-400 hover:text-slate-900"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <!-- Mobile Tab Switcher (if Receipt Image Exists) -->
        <div v-if="currentSale.receiptImageURL" class="md:hidden flex bg-slate-100 p-1 border-b border-slate-200 shrink-0">
          <button
            type="button"
            @click="mobileCrudTab = 'form'"
            :class="[
              'flex-1 py-2 text-xs font-bold rounded-lg flex items-center justify-center gap-1.5 transition-all',
              mobileCrudTab === 'form' ? 'bg-white text-primary shadow-xs' : 'text-slate-600'
            ]"
          >
            <FileText class="w-3.5 h-3.5" /> Form Input Data
          </button>
          <button
            type="button"
            @click="mobileCrudTab = 'receipt'"
            :class="[
              'flex-1 py-2 text-xs font-bold rounded-lg flex items-center justify-center gap-1.5 transition-all',
              mobileCrudTab === 'receipt' ? 'bg-white text-primary shadow-xs' : 'text-slate-600'
            ]"
          >
            <ImageIcon class="w-3.5 h-3.5" /> Foto Nota Transaksi
          </button>
        </div>

        <!-- Body (Scrollable/Split-Screen if Receipt Image Exists) -->
        <div class="flex-1 min-h-0 overflow-hidden flex flex-col md:flex-row bg-white">
          <!-- Left Panel: Zoomable receipt image viewer -->
          <div
            v-if="currentSale.receiptImageURL"
            :class="[
              'w-full md:w-1/2 bg-slate-950 border-r border-slate-100 flex-col relative min-h-[300px] md:min-h-0 select-none',
              mobileCrudTab === 'receipt' ? 'flex flex-1 min-h-0' : 'hidden md:flex'
            ]"
          >
            <!-- Zoom controls overlay -->
            <div class="absolute top-4 left-4 flex items-center gap-2 z-20">
              <button
                @click="handleCrudZoom(0.2)"
                type="button"
                class="p-2 bg-slate-900/80 hover:bg-slate-800 text-white rounded-xl backdrop-blur-sm border border-slate-700/50 transition-all shadow-md"
                title="Zoom In"
              >
                <ZoomIn class="w-4 h-4" />
              </button>
              <button
                @click="handleCrudZoom(-0.2)"
                type="button"
                class="p-2 bg-slate-900/80 hover:bg-slate-800 text-white rounded-xl backdrop-blur-sm border border-slate-700/50 transition-all shadow-md"
                title="Zoom Out"
              >
                <ZoomOut class="w-4 h-4" />
              </button>
              <button
                @click="crudZoomLevel = 1; crudRotationAngle = 0"
                type="button"
                class="p-2 bg-slate-900/80 hover:bg-slate-800 text-white rounded-xl backdrop-blur-sm border border-slate-700/50 transition-all shadow-md"
                title="Reset View"
              >
                <RefreshCw class="w-4 h-4" />
              </button>
              <button
                @click="crudRotationAngle -= 90"
                type="button"
                class="p-2 bg-slate-900/80 hover:bg-slate-800 text-white rounded-xl backdrop-blur-sm border border-slate-700/50 transition-all shadow-md"
                title="Rotate Image"
              >
                <RotateCcw class="w-4 h-4" />
              </button>
            </div>
            <!-- Image viewport -->
            <div class="flex-1 overflow-auto flex items-center justify-center p-6 custom-scrollbar cursor-grab active:cursor-grabbing min-h-0">
              <img
                :src="currentSale.receiptImageURL"
                :style="`transform: scale(${crudZoomLevel}) rotate(${crudRotationAngle}deg); transform-origin: center;`"
                class="max-w-full max-h-[60vh] object-contain transition-transform duration-200"
              />
            </div>
            <!-- Indicator footer -->
            <div class="p-3.5 bg-slate-900 border-t border-slate-800 text-center text-[10px] font-black text-slate-400 tracking-widest uppercase">
              Foto Nota Transaksi Fisik (Kunjungan Sales)
            </div>
          </div>

          <!-- Right Panel / Main Form Container -->
          <div
            :class="[
              'flex-1 overflow-y-auto p-4 sm:p-6 space-y-4 sm:space-y-5 custom-scrollbar min-h-0',
              mobileCrudTab === 'form' || !currentSale.receiptImageURL ? 'block' : 'hidden md:block'
            ]"
          >
            <div class="space-y-1.5">
              <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">Pilih Salesman *</label>
              <SearchSelect
                v-model="currentSale.employeeID"
                :options="salesmanOptions"
                placeholder="Pilih Salesman..."
                search-placeholder="Cari nama salesman atau jabatan..."
                :clearable="true"
                size="md"
              />
            </div>

            <div class="space-y-1.5">
              <div class="flex items-center justify-between">
                <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">Pilih Toko / Pelanggan</label>
                <span
                  v-if="currentSale.employeeID && salesmanStores.length > 0"
                  class="text-[10px] font-black text-blue-600 bg-blue-50 px-2 py-0.5 rounded-full border border-blue-200"
                >
                  {{ salesmanStores.length }} Toko Binaan
                </span>
              </div>
              <SearchSelect
                v-model="currentSale.storeID"
                :options="storeOptions"
                :placeholder="currentSale.employeeID ? (salesmanStores.length > 0 ? 'Pilih Toko Binaan atau Toko Cabang...' : 'Pilih Toko Cabang...') : 'Pilih Toko (Opsional)...'"
                search-placeholder="Cari nama toko, ID toko (cth: 6A61806B), pemilik, alamat, telp..."
                :clearable="true"
                size="md"
                @change="handleStoreSelect"
                @search="searchStoresRemotely"
              />
              <div v-if="isLoadingSalesmanStores" class="flex items-center gap-1.5 text-xs text-blue-600 font-semibold px-1 py-0.5">
                <Loader2 class="w-3.5 h-3.5 animate-spin" />
                <span>Memuat daftar toko binaan salesman...</span>
              </div>
              <p
                v-else-if="currentSale.employeeID && salesmanStores.length > 0"
                class="text-[11px] text-slate-500 font-medium px-1 flex items-center gap-1"
              >
                <span>💡 Menampilkan {{ salesmanStores.length }} toko binaan sales di urutan teratas.</span>
              </p>
              <p
                v-else-if="currentSale.employeeID && salesmanStores.length === 0"
                class="text-[11px] text-amber-600 font-medium px-1 flex items-center gap-1"
              >
                <span>Salesman ini belum memiliki toko binaan khusus. Menampilkan seluruh toko cabang.</span>
              </p>
            </div>

            <!-- Items Section -->
            <div class="space-y-3">
              <div class="flex items-center justify-between">
                <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">Rincian Produk</label>
                <button
                  @click="addItem"
                  type="button"
                  class="text-[10px] font-black text-primary flex items-center gap-1 hover:underline cursor-pointer"
                >
                  <Plus class="w-3 h-3" /> TAMBAH PRODUK
                </button>
              </div>

              <!-- Unified Spacious Card Layout for Items -->
              <div v-if="currentSale.items?.length > 0" class="space-y-3">
                <div
                  v-for="(item, index) in currentSale.items"
                  :key="'crud-item-' + index"
                  class="bg-slate-50/80 hover:bg-slate-50 p-3.5 sm:p-4 rounded-2xl border border-slate-200/80 hover:border-slate-300 transition-all space-y-3 shadow-2xs"
                >
                  <!-- Top Row: Product Selector & Action -->
                  <div class="flex items-start gap-2 sm:gap-3">
                    <span class="w-6 h-6 rounded-lg bg-primary/10 text-primary flex items-center justify-center text-[11px] font-black shrink-0 mt-1">
                      {{ index + 1 }}
                    </span>
                    <div class="flex-1 min-w-0">
                      <label class="block text-[9px] font-black text-slate-400 uppercase tracking-wider mb-1 ml-0.5">Pilih Produk *</label>
                      <ProductSearchSelect
                        v-model="item.product_id"
                        :products="factoryStore.sellableProducts"
                        @change="() => updateItemPrice(index)"
                      />
                    </div>
                    <button
                      @click="removeItem(index)"
                      type="button"
                      class="p-2 text-rose-400 hover:text-rose-600 hover:bg-rose-50 rounded-xl transition-colors shrink-0 mt-4 cursor-pointer"
                      title="Hapus Produk"
                    >
                      <Trash2 class="w-4 h-4" />
                    </button>
                  </div>

                  <!-- Bottom Row: Input Fields & Subtotal -->
                  <div class="grid grid-cols-1 sm:grid-cols-12 gap-2.5 sm:gap-3 items-end pt-2 border-t border-slate-200/60">
                    <div class="sm:col-span-3">
                      <label class="block text-[9px] font-black text-slate-400 uppercase tracking-wider mb-1 ml-0.5">Jumlah (Qty)</label>
                      <input
                        v-model.number="item.quantity"
                        type="number"
                        min="1"
                        placeholder="1"
                        class="w-full bg-white border border-slate-200 rounded-xl px-3 py-2 text-xs font-bold text-center text-slate-900 focus:ring-2 focus:ring-primary/20 focus:border-primary/50 outline-none transition-all shadow-2xs"
                      />
                    </div>
                    <div class="sm:col-span-4">
                      <label class="block text-[9px] font-black text-slate-400 uppercase tracking-wider mb-1 ml-0.5">Harga Satuan (Rp)</label>
                      <div class="relative">
                        <span class="absolute left-3 top-1/2 -translate-y-1/2 text-[10px] font-bold text-slate-400">Rp</span>
                        <input
                          v-model.number="item.price"
                          type="number"
                          min="0"
                          placeholder="0"
                          class="w-full bg-white border border-slate-200 rounded-xl pl-8 pr-3 py-2 text-xs font-bold text-right text-slate-900 focus:ring-2 focus:ring-primary/20 focus:border-primary/50 outline-none transition-all shadow-2xs"
                        />
                      </div>
                    </div>
                    <div class="sm:col-span-3">
                      <div class="flex items-center justify-between mb-1 ml-0.5">
                        <label class="block text-[9px] font-black text-slate-400 uppercase tracking-wider">Diskon Produk</label>
                        <span v-if="item.discount_percent > 0" class="text-[8px] font-bold text-rose-600 bg-rose-50 px-1 rounded">
                          -{{ formatCurrency(getItemDiscountNominal(item)) }}
                        </span>
                      </div>
                      <div class="grid grid-cols-2 gap-1.5">
                        <div class="relative">
                          <input
                            v-model.number="item.discount_percent"
                            type="number"
                            min="0"
                            max="100"
                            placeholder="0"
                            class="w-full bg-white border border-slate-200 rounded-xl pl-2 pr-5 py-2 text-xs font-bold text-center text-slate-900 focus:ring-2 focus:ring-primary/20 focus:border-primary/50 outline-none transition-all shadow-2xs"
                            title="Diskon dalam Persen (%)"
                          />
                          <span class="absolute right-1.5 top-1/2 -translate-y-1/2 text-[9px] font-bold text-slate-400">%</span>
                        </div>
                        <div class="relative">
                          <span class="absolute left-1.5 top-1/2 -translate-y-1/2 text-[8px] font-bold text-slate-400">Rp</span>
                          <input
                            :value="getItemDiscountNominal(item)"
                            @input="(e: any) => updateItemDiscountFromNominal(index, Number(e.target.value || 0))"
                            type="number"
                            min="0"
                            placeholder="0"
                            class="w-full bg-white border border-slate-200 rounded-xl pl-5 pr-1.5 py-2 text-xs font-bold text-right text-slate-900 focus:ring-2 focus:ring-primary/20 focus:border-primary/50 outline-none transition-all shadow-2xs"
                            title="Diskon dalam Nominal Rupiah (Rp)"
                          />
                        </div>
                      </div>
                    </div>
                    <div class="sm:col-span-2">
                      <label class="block text-[9px] font-black text-slate-400 uppercase tracking-wider mb-1 ml-0.5 text-right sm:text-right">Subtotal</label>
                      <div class="bg-white border border-slate-200 rounded-xl px-2.5 py-2 flex items-center justify-end h-[38px] shadow-2xs">
                        <span class="text-xs font-black text-primary truncate">
                          {{ formatCurrency((item.quantity * item.price) * (1 - (item.discount_percent || 0)/100)) }}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <div
                v-else
                class="text-center py-6 border-2 border-dashed border-slate-100 rounded-2xl"
              >
                <p class="text-[10px] font-bold text-slate-300 uppercase tracking-widest italic">
                  Belum ada item ditambahkan
                </p>
              </div>
            </div>

            <!-- Return Items (Editable) -->
            <div class="space-y-3">
              <div class="flex items-center justify-between">
                <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">Retur Barang Sampingan</label>
                <button
                  @click="addReturnItem"
                  type="button"
                  class="text-[10px] font-black text-primary flex items-center gap-1 hover:underline cursor-pointer"
                >
                  <Plus class="w-3 h-3" /> TAMBAH RETUR
                </button>
              </div>

              <!-- Unified Card Layout for Returns -->
              <div v-if="currentSale.return_items && currentSale.return_items.length > 0" class="space-y-3">
                <div
                  v-for="(ret, index) in currentSale.return_items"
                  :key="'ret-edit-' + index"
                  class="bg-rose-50/60 hover:bg-rose-50 p-3.5 sm:p-4 rounded-2xl border border-rose-100 hover:border-rose-200 transition-all space-y-3 shadow-2xs"
                >
                  <!-- Top Row: Return Product Selector & Action -->
                  <div class="flex items-start gap-2 sm:gap-3">
                    <span class="w-6 h-6 rounded-lg bg-rose-100 text-rose-700 flex items-center justify-center text-[11px] font-black shrink-0 mt-1">
                      {{ index + 1 }}
                    </span>
                    <div class="flex-1 min-w-0">
                      <label class="block text-[9px] font-black text-rose-700 uppercase tracking-wider mb-1 ml-0.5">Pilih Barang Retur *</label>
                      <ProductSearchSelect
                        v-model="ret.product_id"
                        :products="returnableProducts"
                        :exclude-raw-materials="false"
                        :show-catalog-button="false"
                        :fallback-label="ret.product_name"
                        placeholder="Pilih barang retur dari Master Produk..."
                        @change="() => updateReturnItemProduct(index)"
                      />
                    </div>
                    <button
                      @click="removeReturnItem(index)"
                      type="button"
                      class="p-2 text-rose-400 hover:text-rose-600 hover:bg-rose-100/80 rounded-xl transition-colors shrink-0 mt-4 cursor-pointer"
                      title="Hapus Retur"
                    >
                      <Trash2 class="w-4 h-4" />
                    </button>
                  </div>

                  <!-- Bottom Row: Qty, Unit, Price, Subtotal -->
                  <div class="grid grid-cols-1 sm:grid-cols-12 gap-2.5 sm:gap-3 items-end pt-2 border-t border-rose-100">
                    <div class="sm:col-span-3">
                      <label class="block text-[9px] font-black text-rose-700 uppercase tracking-wider mb-1 ml-0.5">Jumlah (Qty)</label>
                      <input
                        v-model.number="ret.quantity"
                        type="number"
                        min="1"
                        placeholder="1"
                        class="w-full bg-white border border-rose-200 rounded-xl px-3 py-2 text-xs font-bold text-center text-slate-900 focus:ring-2 focus:ring-rose-300 outline-none shadow-2xs"
                      />
                    </div>
                    <div class="sm:col-span-3">
                      <label class="block text-[9px] font-black text-rose-700 uppercase tracking-wider mb-1 ml-0.5">Satuan</label>
                      <input
                        v-model="ret.unit"
                        type="text"
                        placeholder="Btl / Krat / Pcs"
                        class="w-full bg-white border border-rose-200 rounded-xl px-3 py-2 text-xs font-bold uppercase text-slate-900 focus:ring-2 focus:ring-rose-300 outline-none shadow-2xs"
                      />
                    </div>
                    <div class="sm:col-span-3">
                      <label class="block text-[9px] font-black text-rose-700 uppercase tracking-wider mb-1 ml-0.5">Harga Satuan (Rp)</label>
                      <div class="relative">
                        <span class="absolute left-3 top-1/2 -translate-y-1/2 text-[10px] font-bold text-slate-400">Rp</span>
                        <input
                          v-model.number="ret.price"
                          type="number"
                          min="0"
                          placeholder="0"
                          class="w-full bg-white border border-rose-200 rounded-xl pl-8 pr-3 py-2 text-xs font-bold text-right text-slate-900 focus:ring-2 focus:ring-rose-300 outline-none shadow-2xs"
                        />
                      </div>
                    </div>
                    <div class="sm:col-span-3">
                      <label class="block text-[9px] font-black text-rose-700 uppercase tracking-wider mb-1 ml-0.5 text-right sm:text-right">Subtotal</label>
                      <div class="bg-white border border-rose-200 rounded-xl px-3 py-2 flex items-center justify-end h-[38px] shadow-2xs">
                        <span class="text-xs font-black text-rose-700 truncate">
                          {{ formatCurrency((ret.quantity || 0) * (ret.price || 0)) }}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <div
                v-else
                class="text-center py-6 border-2 border-dashed border-slate-100 rounded-2xl"
              >
                <p class="text-[10px] font-bold text-slate-300 uppercase tracking-widest italic">
                  Tidak ada retur barang sampingan
                </p>
              </div>
            </div>

            <!-- Discount & Calculation Breakdown Card -->
            <div class="bg-slate-50/80 border border-slate-200 rounded-2xl p-4 space-y-3 shadow-2xs">
              <div class="flex items-center justify-between">
                <span class="text-[11px] font-black text-slate-700 uppercase tracking-wider flex items-center gap-1.5">
                  <Percent class="w-3.5 h-3.5 text-primary" /> Diskon Nota & Ringkasan Perhitungan
                </span>
                <span v-if="crudItemsNetTotal > 0" class="text-[10px] font-bold text-slate-500">
                  Subtotal Bersih Produk: <strong class="text-slate-800">{{ formatCurrency(crudItemsNetTotal) }}</strong>
                </span>
              </div>

              <div class="grid grid-cols-1 md:grid-cols-12 gap-3 items-end pt-1">
                <!-- Mode Diskon Nota -->
                <div class="md:col-span-3">
                  <label class="block text-[9px] font-black text-slate-500 uppercase tracking-wider mb-1 ml-0.5">Jenis Diskon Nota</label>
                  <select
                    v-model="currentSale.globalDiscountType"
                    @change="() => {
                      if (currentSale.globalDiscountType === 'PERCENT') {
                        currentSale.globalDiscountPercent = crudItemsNetTotal > 0 ? Math.round((currentSale.discountAmount / crudItemsNetTotal) * 100) : 0
                      }
                    }"
                    class="w-full bg-white border border-slate-200 rounded-xl px-3 py-2 text-xs font-bold text-slate-800 focus:ring-2 focus:ring-primary/20 outline-none shadow-2xs"
                  >
                    <option value="NOMINAL">Nominal Rupiah (Rp)</option>
                    <option value="PERCENT">Persentase (%)</option>
                  </select>
                </div>

                <!-- Input Nilai Diskon Nota -->
                <div class="md:col-span-5">
                  <label class="block text-[9px] font-black text-slate-500 uppercase tracking-wider mb-1 ml-0.5">
                    Nilai Diskon Total Nota {{ currentSale.globalDiscountType === 'PERCENT' ? '(%)' : '(Rp)' }}
                  </label>
                  <div v-if="currentSale.globalDiscountType === 'NOMINAL'" class="relative">
                    <span class="absolute left-3 top-1/2 -translate-y-1/2 text-[10px] font-bold text-slate-400">Rp</span>
                    <input
                      v-model.number="currentSale.discountAmount"
                      type="number"
                      min="0"
                      placeholder="0"
                      class="w-full bg-white border border-slate-200 rounded-xl pl-8 pr-3 py-2 text-xs font-bold text-right text-slate-900 focus:ring-2 focus:ring-primary/20 outline-none shadow-2xs"
                    />
                  </div>
                  <div v-else class="relative">
                    <input
                      v-model.number="currentSale.globalDiscountPercent"
                      type="number"
                      min="0"
                      max="100"
                      placeholder="0"
                      class="w-full bg-white border border-slate-200 rounded-xl pl-3 pr-8 py-2 text-xs font-bold text-center text-slate-900 focus:ring-2 focus:ring-primary/20 outline-none shadow-2xs"
                    />
                    <span class="absolute right-3 top-1/2 -translate-y-1/2 text-[10px] font-bold text-slate-400">%</span>
                  </div>
                </div>

                <!-- Hasil Potongan Diskon Nota -->
                <div class="md:col-span-4">
                  <label class="block text-[9px] font-black text-slate-500 uppercase tracking-wider mb-1 ml-0.5 text-right">Potongan Diskon Nota</label>
                  <div class="bg-rose-50 border border-rose-200 rounded-xl px-3 py-2 flex items-center justify-end h-[38px] shadow-2xs">
                    <span class="text-xs font-black text-rose-700">
                      -{{ formatCurrency(currentSale.discountAmount || 0) }}
                    </span>
                  </div>
                </div>
              </div>

              <!-- Rangkuman Ringkas -->
              <div class="pt-2 border-t border-slate-200/60 grid grid-cols-2 sm:grid-cols-4 gap-2 text-center text-[10px]">
                <div class="p-2 bg-white rounded-xl border border-slate-100">
                  <p class="text-slate-400 font-bold uppercase text-[9px]">Kotor Produk</p>
                  <p class="font-black text-slate-700 mt-0.5">{{ formatCurrency(crudItemsGrossTotal) }}</p>
                </div>
                <div class="p-2 bg-white rounded-xl border border-slate-100">
                  <p class="text-rose-500 font-bold uppercase text-[9px]">Diskon Produk</p>
                  <p class="font-black text-rose-600 mt-0.5">-{{ formatCurrency(crudTotalItemDiscounts) }}</p>
                </div>
                <div class="p-2 bg-white rounded-xl border border-slate-100">
                  <p class="text-rose-500 font-bold uppercase text-[9px]">Diskon Nota</p>
                  <p class="font-black text-rose-600 mt-0.5">-{{ formatCurrency(currentSale.discountAmount || 0) }}</p>
                </div>
                <div class="p-2 bg-white rounded-xl border border-slate-100">
                  <p class="text-amber-500 font-bold uppercase text-[9px]">Retur Sampingan</p>
                  <p class="font-black text-amber-600 mt-0.5">-{{ formatCurrency(crudTotalReturnAmount) }}</p>
                </div>
              </div>
            </div>

            <!-- Form Summary & Metadata -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-3.5 sm:gap-5">
              <div class="space-y-1.5">
                <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">Kategori Toko</label>
                <select
                  v-model="currentSale.storeCategory"
                  class="w-full bg-slate-50 border-2 border-transparent rounded-2xl px-4 py-3 font-bold text-xs text-slate-900 focus:bg-white focus:border-primary/20 focus:ring-0 transition-all outline-none"
                >
                  <option value="TOKO_LAMA">Toko Lama</option>
                  <option value="TOKO_BARU">Toko Baru</option>
                </select>
              </div>
              <div class="space-y-1.5">
                <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">Total Omzet</label>
                <div class="relative">
                  <span class="absolute left-4 top-1/2 -translate-y-1/2 font-black text-slate-400 text-xs">Rp</span>
                  <input
                    v-model.number="currentSale.totalAmount"
                    type="number"
                    min="0"
                    placeholder="0"
                    class="w-full bg-slate-50 border-2 border-transparent rounded-2xl pl-10 pr-4 py-3 font-black text-xs text-slate-900 focus:bg-white focus:border-primary/20 focus:ring-0 transition-all outline-none"
                  />
                </div>
                <p v-if="authStore.user?.role === 'admin_nota'" class="text-[9px] text-slate-400 font-medium italic mt-1 ml-1">
                  Admin Nota: Anda dapat merubah total ini secara manual.
                </p>
                <p v-if="authStore.user?.role === 'admin_nota' && isPaidInvoiceEdit" class="text-[9px] text-amber-700 font-medium mt-1 ml-1">
                  Pembayaran tercatat tetap {{ formatCurrency(currentSale.paidAmount) }}; status dan sisa tagihan akan mengikuti total baru.
                </p>
              </div>
            </div>

            <!-- Payment Method Selection -->
            <div class="space-y-2 bg-slate-50/80 p-3.5 sm:p-4 rounded-2xl border border-slate-100">
              <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1 flex items-center gap-1.5">
                <CreditCard class="w-3.5 h-3.5 text-primary" /> Metode Pembayaran Faktur
              </label>
              <div class="grid grid-cols-2 sm:grid-cols-4 gap-2">
                <button
                  type="button"
                  @click="currentSale.paymentMethod = 'CASH'"
                  :class="[
                    'px-3 py-2.5 text-xs font-bold rounded-xl border flex items-center justify-center gap-1.5 transition-all cursor-pointer',
                    currentSale.paymentMethod === 'CASH'
                      ? 'bg-emerald-600 text-white border-emerald-600 shadow-sm shadow-emerald-200'
                      : 'bg-white text-slate-700 border-slate-200 hover:bg-slate-50'
                  ]"
                >
                  <CreditCard class="w-3.5 h-3.5" /> Tunai (Cash)
                </button>
                <button
                  type="button"
                  @click="currentSale.paymentMethod = 'TRANSFER'"
                  :class="[
                    'px-3 py-2.5 text-xs font-bold rounded-xl border flex items-center justify-center gap-1.5 transition-all cursor-pointer',
                    currentSale.paymentMethod === 'TRANSFER'
                      ? 'bg-emerald-600 text-white border-emerald-600 shadow-sm shadow-emerald-200'
                      : 'bg-white text-slate-700 border-slate-200 hover:bg-slate-50'
                  ]"
                >
                  <Wallet class="w-3.5 h-3.5" /> Transfer
                </button>
                <button
                  type="button"
                  @click="currentSale.paymentMethod = 'QRIS'"
                  :class="[
                    'px-3 py-2.5 text-xs font-bold rounded-xl border flex items-center justify-center gap-1.5 transition-all cursor-pointer',
                    currentSale.paymentMethod === 'QRIS'
                      ? 'bg-emerald-600 text-white border-emerald-600 shadow-sm shadow-emerald-200'
                      : 'bg-white text-slate-700 border-slate-200 hover:bg-slate-50'
                  ]"
                >
                  <Wallet class="w-3.5 h-3.5" /> QRIS
                </button>
                <button
                  type="button"
                  @click="currentSale.paymentMethod = 'TEMPO'"
                  :class="[
                    'px-3 py-2.5 text-xs font-bold rounded-xl border flex items-center justify-center gap-1.5 transition-all cursor-pointer',
                    currentSale.paymentMethod === 'TEMPO'
                      ? 'bg-amber-600 text-white border-amber-600 shadow-sm shadow-amber-200'
                      : 'bg-white text-slate-700 border-slate-200 hover:bg-slate-50'
                  ]"
                >
                  <CalendarClock class="w-3.5 h-3.5" /> Kredit / Tempo
                </button>
              </div>

              <!-- Panel Upload Bukti Pembayaran Non-Tunai (Transfer / QRIS) -->
              <div
                v-if="currentSale.paymentMethod === 'TRANSFER' || currentSale.paymentMethod === 'QRIS'"
                class="mt-3 p-3.5 sm:p-4 bg-indigo-50/90 border border-indigo-200 rounded-xl space-y-3 animate-in fade-in duration-150"
              >
                <div class="flex items-center justify-between">
                  <span class="text-[11px] font-black text-indigo-900 uppercase tracking-wider flex items-center gap-1.5">
                    <Wallet class="w-3.5 h-3.5 text-indigo-600" /> Bukti Pembayaran ({{ currentSale.paymentMethod }})
                  </span>
                  <span class="text-[10px] text-indigo-700 font-bold bg-indigo-100 px-2 py-0.5 rounded-md border border-indigo-200">
                    Non-Tunai
                  </span>
                </div>

                <div class="flex items-center gap-3 sm:gap-4">
                  <input
                    type="file"
                    @change="handleTransferProofUpload"
                    accept="image/*"
                    class="hidden"
                    id="transfer-proof-upload"
                  />
                  <label
                    for="transfer-proof-upload"
                    class="flex-1 flex items-center justify-center gap-2 bg-white border-2 border-dashed border-indigo-200 rounded-xl py-3 cursor-pointer hover:border-indigo-400 hover:bg-indigo-50/50 transition-all text-indigo-700"
                  >
                    <Upload v-if="isUploadingTransferProof" class="w-4 h-4 animate-spin text-indigo-600" />
                    <ImageIcon v-else class="w-4 h-4 text-indigo-500" />
                    <span class="text-xs font-bold">
                      {{ isUploadingTransferProof ? 'Mengunggah Bukti...' : (currentSale.transferProofURL ? 'Ganti Bukti Transfer' : 'Pilih / Unggah Bukti Transfer') }}
                    </span>
                  </label>
                  <div
                    v-if="currentSale.transferProofURL"
                    class="flex items-center gap-2"
                  >
                    <div
                      @click="openImagePreview(currentSale.transferProofURL)"
                      class="w-11 h-11 sm:w-12 sm:h-12 rounded-xl overflow-hidden border border-indigo-200 shadow-sm cursor-zoom-in hover:opacity-80 transition-all shrink-0 bg-white"
                      title="Klik untuk memperbesar bukti transfer"
                    >
                      <img :src="currentSale.transferProofURL" class="w-full h-full object-cover" />
                    </div>
                    <button
                      type="button"
                      @click="handleRemoveTransferProof"
                      class="p-2 text-red-500 hover:bg-red-100 rounded-xl transition-all cursor-pointer"
                      title="Hapus Bukti Transfer"
                    >
                      <Trash2 class="w-4 h-4" />
                    </button>
                  </div>
                </div>
                <p class="text-[10px] text-indigo-600 font-medium italic">
                  Bukti transfer ini akan otomatis tersinkronisasi ke antrean persetujuan Kasir & Finance.
                </p>
              </div>

              <!-- Panel Pengaturan Tempo & Uang Muka (DP) -->
              <div
                v-if="currentSale.paymentMethod === 'TEMPO'"
                class="mt-3 p-3.5 sm:p-4 bg-amber-50/90 border border-amber-200 rounded-2xl space-y-3.5 animate-in fade-in duration-150"
              >
                <div class="flex items-center justify-between">
                  <span class="text-[11px] font-black text-amber-900 uppercase tracking-wider flex items-center gap-1.5">
                    <CalendarClock class="w-3.5 h-3.5 text-amber-600" /> Skema Cicilan Tempo & Pembayaran Awal (DP)
                  </span>
                  <span class="text-[10px] text-amber-700 font-bold bg-amber-100 px-2.5 py-0.5 rounded-md border border-amber-200">
                    Faktur Cicilan (Piutang)
                  </span>
                </div>

                <!-- Pilihan Target Jumlah Cicilan (Tenor) -->
                <div>
                  <div class="flex items-center justify-between mb-1.5">
                    <label class="block text-[10px] font-black text-amber-900 uppercase tracking-wider">
                      Target Berapa Kali Cicilan?
                    </label>
                    <span v-if="estimatedTempoPerInstallment > 0 && currentSale.tempoInstallmentCount !== 1" class="text-[10px] font-extrabold text-amber-700">
                      Estimasi: ~{{ formatCurrency(estimatedTempoPerInstallment) }} / cicilan
                    </span>
                  </div>
                  <div class="grid grid-cols-5 gap-1.5">
                    <button
                      type="button"
                      @click="setCurrentSaleInstallmentCount(1)"
                      :class="[
                        'py-2 px-2 text-xs font-black rounded-xl border text-center transition-all cursor-pointer',
                        currentSale.tempoInstallmentCount === 1
                          ? 'bg-amber-600 text-white border-amber-600 shadow-sm'
                          : 'bg-white text-amber-900 border-amber-200 hover:bg-amber-100/50'
                      ]"
                    >
                      1x (Lunas Sekaligus)
                    </button>
                    <button
                      type="button"
                      @click="setCurrentSaleInstallmentCount(2)"
                      :class="[
                        'py-2 px-2 text-xs font-black rounded-xl border text-center transition-all cursor-pointer',
                        currentSale.tempoInstallmentCount === 2
                          ? 'bg-amber-600 text-white border-amber-600 shadow-sm'
                          : 'bg-white text-amber-900 border-amber-200 hover:bg-amber-100/50'
                      ]"
                    >
                      2x Cicilan
                    </button>
                    <button
                      type="button"
                      @click="setCurrentSaleInstallmentCount(3)"
                      :class="[
                        'py-2 px-2 text-xs font-black rounded-xl border text-center transition-all cursor-pointer',
                        currentSale.tempoInstallmentCount === 3
                          ? 'bg-amber-600 text-white border-amber-600 shadow-sm'
                          : 'bg-white text-amber-900 border-amber-200 hover:bg-amber-100/50'
                      ]"
                    >
                      3x Cicilan
                    </button>
                    <button
                      type="button"
                      @click="setCurrentSaleInstallmentCount(4)"
                      :class="[
                        'py-2 px-2 text-xs font-black rounded-xl border text-center transition-all cursor-pointer',
                        currentSale.tempoInstallmentCount === 4
                          ? 'bg-amber-600 text-white border-amber-600 shadow-sm'
                          : 'bg-white text-amber-900 border-amber-200 hover:bg-amber-100/50'
                      ]"
                    >
                      4x Cicilan
                    </button>
                    <button
                      type="button"
                      @click="setCurrentSaleInstallmentCount('custom')"
                      :class="[
                        'py-2 px-2 text-xs font-black rounded-xl border text-center transition-all cursor-pointer',
                        currentSale.tempoInstallmentCount === 'custom'
                          ? 'bg-amber-600 text-white border-amber-600 shadow-sm'
                          : 'bg-white text-amber-900 border-amber-200 hover:bg-amber-100/50'
                      ]"
                    >
                      Custom (X)
                    </button>
                  </div>
                  <div v-if="currentSale.tempoInstallmentCount === 'custom'" class="mt-2 flex items-center gap-2">
                    <label class="text-[11px] font-bold text-amber-900 shrink-0">Jumlah Termin Cicilan:</label>
                    <input
                      type="number"
                      min="1"
                      max="24"
                      v-model.number="currentSale.tempoCustomInstallmentCount"
                      @input="typeof currentSale.tempoTenorDays === 'number' && (currentSale.paymentDueDate = calculateDueDateForSale(currentSale.tempoTenorDays))"
                      class="w-24 bg-white border border-amber-300 rounded-xl px-3 py-1.5 text-xs font-black text-slate-900 outline-none focus:ring-2 focus:ring-amber-400"
                    />
                    <span class="text-xs font-bold text-amber-800">Kali</span>
                  </div>
                </div>

                <!-- Pilihan Jangka Waktu / Jatuh Tempo -->
                <div>
                  <label class="block text-[10px] font-black text-amber-900 uppercase tracking-wider mb-1.5">
                    Periode Jatuh Tempo Cicilan
                  </label>
                  <div class="grid grid-cols-5 gap-1.5">
                    <button
                      type="button"
                      @click="setCurrentSaleTenorDays(7)"
                      :class="[
                        'py-1.5 px-2 text-xs font-bold rounded-lg border text-center transition-all cursor-pointer',
                        currentSale.tempoTenorDays === 7
                          ? 'bg-amber-700 text-white border-amber-700'
                          : 'bg-white text-amber-900 border-amber-200 hover:bg-amber-100/50'
                      ]"
                    >
                      7 Hari
                    </button>
                    <button
                      type="button"
                      @click="setCurrentSaleTenorDays(14)"
                      :class="[
                        'py-1.5 px-2 text-xs font-bold rounded-lg border text-center transition-all cursor-pointer',
                        currentSale.tempoTenorDays === 14
                          ? 'bg-amber-700 text-white border-amber-700'
                          : 'bg-white text-amber-900 border-amber-200 hover:bg-amber-100/50'
                      ]"
                    >
                      14 Hari
                    </button>
                    <button
                      type="button"
                      @click="setCurrentSaleTenorDays(30)"
                      :class="[
                        'py-1.5 px-2 text-xs font-bold rounded-lg border text-center transition-all cursor-pointer',
                        currentSale.tempoTenorDays === 30
                          ? 'bg-amber-700 text-white border-amber-700'
                          : 'bg-white text-amber-900 border-amber-200 hover:bg-amber-100/50'
                      ]"
                    >
                      30 Hari
                    </button>
                    <button
                      type="button"
                      @click="setCurrentSaleTenorDays(45)"
                      :class="[
                        'py-1.5 px-2 text-xs font-bold rounded-lg border text-center transition-all cursor-pointer',
                        currentSale.tempoTenorDays === 45
                          ? 'bg-amber-700 text-white border-amber-700'
                          : 'bg-white text-amber-900 border-amber-200 hover:bg-amber-100/50'
                      ]"
                    >
                      45 Hari
                    </button>
                    <button
                      type="button"
                      @click="setCurrentSaleTenorDays('custom')"
                      :class="[
                        'py-1.5 px-2 text-xs font-bold rounded-lg border text-center transition-all cursor-pointer',
                        currentSale.tempoTenorDays === 'custom'
                          ? 'bg-amber-700 text-white border-amber-700'
                          : 'bg-white text-amber-900 border-amber-200 hover:bg-amber-100/50'
                      ]"
                    >
                      Custom Tgl
                    </button>
                  </div>
                </div>

                <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <!-- Jatuh Tempo Date -->
                  <div>
                    <label class="block text-[10px] font-bold text-amber-900 uppercase tracking-wider mb-1">
                      Tanggal Jatuh Tempo *
                    </label>
                    <input
                      type="date"
                      v-model="currentSale.paymentDueDate"
                      @input="currentSale.tempoTenorDays = 'custom'"
                      class="w-full bg-white border border-amber-300 rounded-xl px-3 py-2 text-xs font-bold text-slate-900 focus:ring-2 focus:ring-amber-400 outline-none"
                    />
                  </div>

                  <!-- Down Payment (DP) Input -->
                  <div>
                    <label class="block text-[10px] font-bold text-amber-900 uppercase tracking-wider mb-1">
                      Uang Muka / Pembayaran Awal (DP) - Opsional
                    </label>
                    <div class="relative">
                      <span class="absolute left-3 top-1/2 -translate-y-1/2 text-xs font-bold text-slate-400">Rp</span>
                      <input
                        type="number"
                        min="0"
                        :max="isPaidInvoiceEdit ? undefined : currentSale.totalAmount"
                        :disabled="authStore.user?.role === 'admin_nota' && isPaidInvoiceEdit"
                        v-model.number="currentSale.paidAmount"
                        placeholder="0"
                        class="w-full bg-white border border-amber-300 rounded-xl pl-9 pr-3 py-2 text-xs font-bold text-right text-slate-900 focus:ring-2 focus:ring-amber-400 outline-none"
                      />
                    </div>
                  </div>
                </div>

                <!-- Info Ringkasan Piutang -->
                <div class="pt-2.5 border-t border-amber-200/80 flex items-center justify-between text-xs font-bold">
                  <span class="text-amber-800">Sisa Tagihan Piutang:</span>
                  <div class="text-right">
                    <span class="text-amber-950 font-black text-sm">
                      {{ formatCurrency(Math.max(0, currentSale.totalAmount - (currentSale.paidAmount || 0))) }}
                    </span>
                    <span v-if="currentSale.tempoInstallmentCount !== 1 && estimatedTempoPerInstallment > 0" class="block text-[10px] text-amber-700 font-medium">
                      (~{{ formatCurrency(estimatedTempoPerInstallment) }} / cicilan)
                    </span>
                  </div>
                </div>
              </div>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-3 gap-3.5 sm:gap-5">
              <div class="space-y-1.5">
                <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">Status Transaksi</label>
                <select
                  v-model="currentSale.status"
                  class="w-full bg-slate-50 border-2 border-transparent rounded-2xl px-4 py-3 font-bold text-xs text-slate-900 focus:bg-white focus:border-primary/20 focus:ring-0 transition-all outline-none"
                >
                  <option value="VERIFIED">Approved (Disetujui)</option>
                  <option value="PENDING">Pending (Menunggu)</option>
                  <option value="REJECTED">Rejected (Ditolak)</option>
                </select>
              </div>
              <div class="space-y-1.5">
                <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">Status Pembayaran</label>
                <select
                  v-model="currentSale.paymentStatus"
                  :disabled="authStore.user?.role === 'admin_nota'"
                  class="w-full bg-slate-50 border-2 border-transparent rounded-2xl px-4 py-3 font-bold text-xs text-slate-900 focus:bg-white focus:border-primary/20 focus:ring-0 transition-all outline-none"
                >
                  <option value="UNPAID">Unpaid (Belum Lunas)</option>
                  <option value="PARTIAL">Partial (Cicilan)</option>
                  <option value="PAID">Paid (Lunas)</option>
                </select>
              </div>
              <div class="space-y-1.5">
                <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">Tanggal Transaksi</label>
                <input
                  v-model="currentSale.transactionDate"
                  type="date"
                  class="w-full bg-slate-50 border-2 border-transparent rounded-2xl px-4 py-3 font-bold text-xs text-slate-900 focus:bg-white focus:border-primary/20 focus:ring-0 transition-all outline-none"
                />
              </div>
            </div>

            <div class="space-y-1.5">
              <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">No. Nota (Opsional)</label>
              <div class="flex gap-2">
                <input
                  v-model="currentSale.receiptNo"
                  type="text"
                  placeholder="INV/YYYYMMDD/SEQ"
                  class="flex-1 bg-slate-50 border-2 border-transparent rounded-2xl px-4 py-3 font-bold text-xs text-slate-900 focus:bg-white focus:border-primary/20 focus:ring-0 transition-all outline-none"
                />
                <button
                  @click="generateReceiptNo"
                  type="button"
                  class="bg-slate-900 text-white px-4 rounded-2xl text-[10px] font-black uppercase tracking-widest hover:bg-slate-800 transition-all shrink-0 cursor-pointer"
                >
                  GENERATE
                </button>
              </div>
            </div>

            <div class="space-y-1.5">
              <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">Foto Nota (Opsional - Digital First)</label>
              <div class="flex items-center gap-3 sm:gap-4">
                <input
                  type="file"
                  @change="handleFileUpload"
                  accept="image/*"
                  class="hidden"
                  id="nota-upload"
                />
                <label
                  for="nota-upload"
                  class="flex-1 flex items-center justify-center gap-2 bg-slate-50 border-2 border-dashed border-slate-200 rounded-2xl py-3 sm:py-4 cursor-pointer hover:border-primary/40 hover:bg-slate-100/50 transition-all"
                >
                  <ImageIcon class="w-4 h-4 sm:w-5 sm:h-5 text-slate-400" />
                  <span class="text-xs font-bold text-slate-500">{{
                    currentSale.receiptImageURL ? 'Ganti Foto' : 'Pilih File Foto'
                  }}</span>
                </label>
                <div
                  v-if="currentSale.receiptImageURL"
                  @click="openImagePreview(currentSale.receiptImageURL)"
                  class="w-11 h-11 sm:w-12 sm:h-12 rounded-xl overflow-hidden border border-slate-100 shadow-sm cursor-zoom-in hover:opacity-80 transition-all shrink-0"
                  title="Klik untuk memperbesar"
                >
                  <img :src="currentSale.receiptImageURL" class="w-full h-full object-cover" />
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="p-3.5 sm:p-6 border-t border-slate-100 flex gap-3 bg-white shrink-0">
          <button
            @click="showModal = false"
            type="button"
            class="flex-1 py-3 sm:py-3.5 text-xs font-black text-slate-500 hover:text-slate-900 border border-slate-200 rounded-xl sm:rounded-2xl transition-all uppercase tracking-widest cursor-pointer"
          >
            BATAL
          </button>
          <button
            @click="handleSubmit"
            type="button"
            class="flex-[2] bg-primary text-white rounded-xl sm:rounded-2xl py-3 sm:py-3.5 text-xs font-black shadow-lg shadow-primary/20 hover:scale-[1.02] active:scale-[0.98] transition-all flex items-center justify-center gap-2 uppercase tracking-widest cursor-pointer"
          >
            <Save class="w-4 h-4" /> SIMPAN DATA
          </button>
        </div>
      </div>
    </div>

    <!-- Image Preview Modal -->
    <div v-if="showImageModal" class="fixed inset-0 z-[100] flex items-center justify-center p-4">
      <div
        class="absolute inset-0 bg-slate-900/90 backdrop-blur-md"
        @click="showImageModal = false"
      ></div>

      <!-- Controls -->
      <div class="absolute top-6 right-6 flex items-center gap-3 z-[110]">
        <button
          @click="handleZoom(0.2)"
          class="p-3 bg-white/10 hover:bg-white/20 text-white rounded-2xl backdrop-blur-xl border border-white/10 transition-all"
        >
          <ZoomIn class="w-6 h-6" />
        </button>
        <button
          @click="handleZoom(-0.2)"
          class="p-3 bg-white/10 hover:bg-white/20 text-white rounded-2xl backdrop-blur-xl border border-white/10 transition-all"
        >
          <ZoomOut class="w-6 h-6" />
        </button>
        <button
          @click="zoomLevel = 1; rotationAngle = 0"
          class="p-3 bg-white/10 hover:bg-white/20 text-white rounded-2xl backdrop-blur-xl border border-white/10 transition-all"
          title="Reset View"
        >
          <RefreshCw class="w-6 h-6" />
        </button>
        <button
          @click="rotationAngle -= 90"
          class="p-3 bg-white/10 hover:bg-white/20 text-white rounded-2xl backdrop-blur-xl border border-white/10 transition-all"
          title="Rotate Image"
        >
          <RotateCcw class="w-6 h-6" />
        </button>
        <div class="w-px h-8 bg-white/10 mx-2"></div>
        <button
          @click="showImageModal = false"
          class="p-3 bg-white hover:bg-slate-100 text-slate-900 rounded-2xl transition-all"
        >
          <X class="w-6 h-6" />
        </button>
      </div>

      <div
        class="relative z-[105] max-w-5xl max-h-[85vh] transition-transform duration-300 ease-out flex items-center justify-center overflow-hidden rounded-3xl shadow-2xl"
      >
        <img
          :src="previewImageUrl"
          :style="`transform: scale(${zoomLevel}) rotate(${rotationAngle}deg)`"
          class="max-w-full max-h-full object-contain transition-transform duration-200 cursor-grab active:cursor-grabbing"
        />
      </div>
    </div>

    <!-- DO Printing Modal (SJ Preview) - LANDSCAPE VERSION -->
    <div
      v-if="selectedDO"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-sm z-[200] flex items-start justify-center p-0 sm:p-8 overflow-y-auto print:p-0"
    >
      <Card
        id="surat-jalan-card"
        class="w-full max-w-[297mm] bg-white shadow-2xl border-0 sm:rounded-[24px] print:rounded-none my-auto"
      >
        <!-- Compact Landscape Header -->
        <div class="bg-white p-4 print:p-2">
          <div class="grid grid-cols-1 sm:grid-cols-3 print:grid-cols-3 gap-3 sm:gap-8 items-start sm:items-center">
            <div class="space-y-1">
              <p class="text-[9px] font-black text-slate-900 uppercase tracking-widest">
                SOPIR / ARMADA
              </p>
              <p class="text-[11px] font-black text-slate-900">
                {{ selectedDO.driver?.first_name || '........................' }} /
                {{ selectedDO.vehicle?.license_plate || '..........' }}
              </p>
            </div>
            <div class="space-y-1">
              <p class="text-[9px] font-black text-slate-900 uppercase tracking-widest">
                SALESMAN / HELPER
              </p>
              <p class="text-[11px] font-black text-slate-900 leading-tight">
                {{
                  selectedDO.items?.[0]?.sales_transaction?.employee?.first_name ||
                  '........................'
                }}
              </p>
              <p class="text-[8px] font-bold text-slate-500 uppercase tracking-tighter">
                {{
                  selectedDO.items?.[0]?.sales_transaction?.employee?.job_position?.title || '.....'
                }}
                •
                {{ selectedDO.items?.[0]?.sales_transaction?.employee?.company?.name || '.....' }}
              </p>
            </div>
            <div class="text-left sm:text-right">
              <p
                class="text-[10px] font-mono font-bold text-slate-900 pb-1 italic uppercase tracking-tighter"
              >
                SURAT JALAN #{{ selectedDO.delivery_order_no }}
              </p>
              <p class="text-[9px] font-bold text-slate-900 mt-1 uppercase">
                {{
                  new Date(selectedDO.created_at).toLocaleDateString('id-ID', { dateStyle: 'long' })
                }}
              </p>
            </div>
          </div>
        </div>

        <div class="p-4 sm:p-6 space-y-4 print:p-2">
          <!-- Main Table with Placeholders -->
          <div class="border-2 border-slate-900 rounded-lg overflow-x-auto bg-white">
            <table class="w-full text-left text-[10px] border-collapse whitespace-nowrap min-w-[700px] sm:min-w-full">
              <thead>
                <tr class="bg-slate-50 border-b-2 border-slate-900">
                  <th class="p-2 border-r border-slate-300 font-black text-center w-8">NO</th>
                  <th class="p-2 border-r border-slate-300 font-black">
                    NAMA TOKO / TELP / ALAMAT
                  </th>
                  <th class="p-2 border-r border-slate-300 font-black text-right w-20 bg-amber-50">
                    TAGIHAN
                  </th>

                  <th
                    class="p-1 border-r border-slate-300 font-black text-center w-8 bg-slate-100/50"
                  >
                    GLN
                  </th>
                  <th
                    class="p-1 border-r border-slate-300 font-black text-center w-8 bg-slate-100/50"
                  >
                    BTL
                  </th>
                  <th
                    class="p-1 border-r border-slate-300 font-black text-center w-8 bg-slate-100/50"
                  >
                    KRT
                  </th>
                  <th
                    class="p-2 border-r border-slate-300 font-black text-center w-24 bg-emerald-50"
                  >
                    BAYAR
                  </th>
                  <th class="p-2 border-r border-slate-300 font-black text-center w-16">KET</th>

                  <th
                    v-for="prod in batchProductColumns"
                    :key="prod.id"
                    class="p-1 border-r border-slate-300 font-black text-center text-[7px] bg-slate-100/30 w-10 uppercase"
                  >
                    {{ prod.name }}
                  </th>

                  <th
                    class="p-2 border-r border-slate-300 font-black text-center w-10 bg-pink-50 text-[8px]"
                  >
                    HADIAH
                  </th>
                  <th class="p-2 font-black text-center w-10 text-[8px]">RETUR</th>
                </tr>
              </thead>
              <tbody class="divide-y border-slate-300">
                <tr v-for="(item, idx) in selectedDO.items" :key="item.id" class="h-8">
                  <td class="p-2 border-r border-slate-300 text-center font-bold">
                    {{ (idx as number) + 1 }}
                  </td>
                  <td class="p-2 border-r border-slate-300 whitespace-normal">
                    <p class="font-black text-slate-900 leading-none">
                      {{ item.sales_transaction?.store?.name }}
                      <span class="text-[8px] text-slate-900 font-normal"
                        >/ {{ item.sales_transaction?.store?.phone || '-' }}</span
                      >
                    </p>
                    <p class="text-[7px] text-slate-900 font-bold mt-0.5">
                      {{ item.sales_transaction?.store?.address }}
                    </p>
                  </td>
                  <td class="p-2 border-r border-slate-300 text-right font-black bg-amber-50/20">
                    {{
                      formatCurrency(item.sales_transaction?.total_amount).replace('Rp', '').trim()
                    }}
                  </td>
                  <td class="p-2 border-r border-slate-300"></td>
                  <td class="p-2 border-r border-slate-300"></td>
                  <td class="p-2 border-r border-slate-300"></td>
                  <td class="p-2 border-r border-slate-300 bg-emerald-50/10"></td>
                  <td class="p-2 border-r border-slate-300 text-center font-bold text-[8px]">
                    {{
                      item.sales_transaction?.payment_method === 'VA' &&
                      item.sales_transaction?.midtrans_bank
                        ? `VA ${item.sales_transaction.midtrans_bank.toUpperCase()}`
                        : item.sales_transaction?.payment_method || 'COD'
                    }}
                  </td>
                  <td
                    v-for="prod in batchProductColumns"
                    :key="prod.id"
                    class="p-1 border-r border-slate-300 text-center font-black"
                  >
                    {{ getProductQty(item.sales_transaction, prod.id) || '' }}
                  </td>
                  <td class="p-2 border-r border-slate-300"></td>
                  <td class="p-2"></td>
                </tr>

                <!-- Placeholder Rows -->
                <tr
                  v-for="n in Math.max(0, 20 - (selectedDO.items?.length || 0))"
                  :key="'blank-' + n"
                  class="h-6 bg-slate-50/10"
                >
                  <td class="p-2 border-r border-slate-300 text-center text-slate-900">
                    {{ (selectedDO.items?.length || 0) + n }}
                  </td>
                  <td class="p-2 border-r border-slate-300"></td>
                  <td class="p-2 border-r border-slate-300 bg-amber-50/10"></td>
                  <td class="p-2 border-r border-slate-300"></td>
                  <td class="p-2 border-r border-slate-300"></td>
                  <td class="p-2 border-r border-slate-300"></td>
                  <td class="p-2 border-r border-slate-300 bg-emerald-50/5"></td>
                  <td class="p-2 border-r border-slate-300"></td>
                  <td
                    v-for="prod in batchProductColumns"
                    :key="prod.id"
                    class="p-1 border-r border-slate-300"
                  ></td>
                  <td class="p-2 border-r border-slate-300"></td>
                  <td class="p-2"></td>
                </tr>

                <!-- Total Row -->
                <tr class="bg-slate-50 font-black border-t-2 border-slate-900 h-8">
                  <td
                    colspan="2"
                    class="p-2 text-right border-r border-slate-300 uppercase tracking-widest text-[8px]"
                  >
                    TOTAL BATCH
                  </td>
                  <td class="p-2 text-right border-r border-slate-300 bg-amber-50">
                    {{ formatCurrency(calculateBatchTotal).replace('Rp', '').trim() }}
                  </td>
                  <td colspan="4" class="border-r border-slate-300"></td>
                  <td class="border-r border-slate-300"></td>
                  <td
                    v-for="prod in batchProductColumns"
                    :key="prod.id"
                    class="p-1 border-r border-slate-300 text-center text-primary bg-slate-100/30"
                  >
                    {{ calculateProductTotal(prod.id) }}
                  </td>
                  <td colspan="2"></td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Signatures -->
          <div class="grid grid-cols-1 sm:grid-cols-3 print:grid-cols-3 gap-6 sm:gap-8 mt-6">
            <div class="text-center space-y-4 sm:space-y-8">
              <p class="text-[9px] font-black uppercase tracking-widest">SOPIR / HELPER</p>
              <p class="text-[8px] font-bold text-slate-900">
                ( .................................. )
              </p>
            </div>
            <div class="text-center space-y-4 sm:space-y-8">
              <p class="text-[9px] font-black uppercase tracking-widest">ADMIN NOTA</p>
              <p class="text-[8px] font-bold text-slate-900">
                ( {{ selectedDO.admin_nota?.first_name || '........................' }} )
              </p>
            </div>
            <div class="flex justify-center sm:justify-end items-center sm:pr-10">
              <div class="text-right mr-4">
                <p class="text-[8px] font-black text-slate-400 uppercase leading-none">
                  SCAN VALIDASI
                </p>
                <p class="text-[7px] text-slate-300 italic mt-1 uppercase text-slate-400">
                  SJ DIGITAL
                </p>
              </div>
              <img
                :src="`https://api.qrserver.com/v1/create-qr-code/?size=150x150&data=${selectedDO.delivery_order_no}`"
                class="w-16 h-16 sm:w-20 sm:h-20 border-2 border-slate-900 p-1 rounded bg-white shadow-sm"
              />
            </div>
          </div>
        </div>

        <CardFooter class="p-4 sm:p-6 bg-slate-50 border-t flex flex-col-reverse sm:flex-row justify-end gap-2 sm:gap-3 print:hidden">
          <Button variant="ghost" class="font-bold text-slate-500 w-full sm:w-auto" @click="selectedDO = null"
            >TUTUP PREVIEW</Button
          >
          <Button
            class="bg-slate-900 hover:bg-slate-800 text-white font-black px-6 sm:px-12 py-3.5 sm:py-6 rounded-xl sm:rounded-2xl shadow-xl shadow-slate-200 w-full sm:w-auto flex items-center justify-center text-xs sm:text-base"
            @click="handlePrint"
          >
            <Printer class="w-4 h-4 sm:w-5 sm:h-5 mr-2 sm:mr-3" />
            CETAK SURAT JALAN
          </Button>
        </CardFooter>
      </Card>
    </div>

    <!-- Nota Printing Modal (Individual Receipt) -->
    <div
      v-if="selectedNota"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-sm z-[200] flex items-start justify-center p-0 sm:p-8 overflow-y-auto print:p-0"
    >
      <Card
        id="nota-penjualan-card"
        class="w-full max-w-[210mm] bg-white shadow-2xl border-0 sm:rounded-[24px] print:rounded-none my-auto"
      >
        <div class="p-4 sm:p-8 print:p-0">
          <!-- Location Radius Verification Banner (Non-print) -->
          <div 
            v-if="getOrderLocationStatus(selectedNota).status !== 'UNKNOWN'" 
            class="mb-6 p-4 rounded-2xl flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 print:hidden shadow-sm border"
            :class="getOrderLocationStatus(selectedNota).status === 'OUTSIDE' ? 'bg-red-50/90 border-red-200' : 'bg-emerald-50/90 border-emerald-200'"
          >
            <div class="flex items-center gap-3">
              <div 
                :class="[
                  'p-2.5 rounded-xl text-white shrink-0 shadow-sm',
                  getOrderLocationStatus(selectedNota).status === 'OUTSIDE' ? 'bg-red-600' : 'bg-emerald-600'
                ]"
              >
                <component :is="getOrderLocationStatus(selectedNota).icon" class="w-5 h-5" />
              </div>
              <div>
                <div class="flex items-center gap-2">
                  <span class="text-xs font-black text-slate-900">Verifikasi Radius Penerbitan</span>
                  <Badge 
                    :class="`${getOrderLocationStatus(selectedNota).bg} ${getOrderLocationStatus(selectedNota).color} border-0 text-[10px] font-black uppercase tracking-wider`"
                  >
                    {{ getOrderLocationStatus(selectedNota).text }}
                  </Badge>
                </div>
                <p class="text-xs text-slate-600 mt-0.5">
                  {{ getOrderLocationStatus(selectedNota).status === 'OUTSIDE' ? 'Perhatian: Transaksi ini dibuat di luar batas toleransi radius 500m dari toko.' : 'Transaksi diterbitkan di dalam toleransi radius toko.' }}
                </p>
              </div>
            </div>
            <Button 
              @click="openMapModal(selectedNota)"
              size="sm"
              variant="outline"
              class="h-9 px-3.5 rounded-xl font-bold text-xs bg-white hover:bg-slate-100 border-slate-300 text-slate-700 shadow-sm shrink-0 flex items-center gap-1.5"
            >
              <Navigation class="w-3.5 h-3.5 text-primary" />
              Lihat di Peta
            </Button>
          </div>

          <!-- Rejected Payments Banner -->
          <div v-if="rejectedPayments.length > 0" class="mb-6 bg-rose-50 border border-rose-200 rounded-xl p-4 flex gap-3 print:hidden shadow-sm">
            <div class="bg-rose-100 rounded-full p-2 h-fit text-rose-600">
              <AlertCircle class="w-5 h-5" />
            </div>
            <div>
              <h3 class="text-sm font-bold text-rose-800 mb-1">Setoran Ditolak Kasir</h3>
              <ul class="text-xs text-rose-600 font-medium space-y-1.5 list-disc list-inside">
                <li v-for="rp in rejectedPayments" :key="rp.id">
                  Setoran sebesar <span class="font-bold">{{ formatCurrency(rp.amount) }}</span> pada {{ formatDate(rp.created_at) }} ditolak.
                  <span v-if="rp.rejection_reason" class="block mt-0.5 ml-4 text-rose-500 italic">Alasan: "{{ rp.rejection_reason }}"</span>
                </li>
              </ul>
              <p class="text-xs text-rose-700 mt-2 font-semibold">Silakan lakukan penyetoran ulang untuk faktur ini.</p>
            </div>
          </div>

          <!-- Header -->
          <div
            class="flex flex-col sm:flex-row justify-between items-start border-b border-slate-900 pb-4 sm:pb-6 mb-4 sm:mb-6 print:flex-row print:pb-2 print:mb-2 print:border-black gap-2"
          >
            <!-- logo bks gja ksa -->
            <div class="flex items-center gap-4">
              <!-- <div class="bg-primary p-3 rounded-2xl">
                <Store class="w-8 h-8 text-white" />
              </div> -->
              <div>
                <h2 class="text-lg sm:text-2xl font-black text-slate-900 tracking-tighter uppercase italic">
                  {{ selectedNota.raw?.employee?.company?.name || 'Perusahaan Demo' }}
                </h2>
              </div>
            </div>
            <div class="text-left sm:text-right">
              <h3 class="text-base sm:text-xl font-black text-slate-900 italic tracking-tighter">
                FAKTUR PENJUALAN
              </h3>
              <p class="text-xs font-black text-slate-400 mt-0.5 sm:mt-1">
                #{{ selectedNota.receiptNo || selectedNota.id.slice(0, 8).toUpperCase() }}
              </p>
            </div>
          </div>

          <!-- Transaction Info -->
          <div class="grid grid-cols-1 sm:grid-cols-2 print:grid-cols-2 gap-4 sm:gap-8 mb-4 sm:mb-2 print:gap-2 print:mb-2">
            <div class="space-y-3 sm:space-y-4 print:space-y-1">
              <div>
                <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest mb-1 print:text-black print:text-[10px]">
                  PELANGGAN
                </p>
                <p class="text-[13px] sm:text-[14px] font-black text-slate-900 uppercase print:text-black print:text-[12px]">
                  {{ selectedNota.store }}
                </p>
                <p class="text-[11px] font-bold text-slate-500 mt-0.5 print:text-black print:text-[10px] print:font-normal">
                  {{ selectedNota.raw?.store?.address || '-' }}
                </p>
              </div>
              <div>
                <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest mb-1 print:text-black print:text-[10px]">
                  SALESMAN
                </p>
                <p class="text-[13px] sm:text-[14px] font-black text-slate-900 uppercase print:text-black print:text-[12px]">
                  {{ selectedNota.salesman }}
                </p>
                <p class="text-[10px] font-bold text-slate-500 mt-0.5 print:text-black print:text-[10px] print:font-normal">
                  {{ selectedNota.raw?.employee?.job_position?.title || 'SALES' }} •
                  {{ selectedNota.raw?.employee?.company?.name || '-' }}
                </p>
              </div>
            </div>
            <div class="space-y-3 sm:space-y-4 text-left sm:text-right print:space-y-1 print:text-right">
              <div>
                <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest mb-1 print:text-black print:text-[10px]">
                  TANGGAL TRANSAKSI
                </p>
                <p class="text-[10px] font-black text-slate-900 print:text-black print:text-[10px]">
                  {{ formatDate(selectedNota.rawDate) }}
                </p>
              </div>
              <div>
                <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest mb-1 print:text-black print:text-[10px]">
                  METODE PEMBAYARAN
                </p>
                <p class="text-[12px] font-black text-slate-900 uppercase print:text-black print:text-[10px]">
                  {{
                    selectedNota.raw?.payment_method === 'VA' && selectedNota.raw?.midtrans_bank
                      ? `VA ${selectedNota.raw.midtrans_bank.toUpperCase()}`
                      : selectedNota.raw?.payment_method || 'CASH'
                  }}
                </p>
              </div>
              <div v-if="selectedNota.raw?.payment_method === 'TEMPO'">
                <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest mb-1 print:text-black print:text-[10px]">
                  JATUH TEMPO
                </p>
                <p class="text-[12px] font-black text-slate-900 print:text-black print:text-[10px]">
                  {{
                    selectedNota.raw?.payment_due_date
                      ? formatDate(selectedNota.raw.payment_due_date)
                      : 'BELUM DITENTUKAN'
                  }}
                </p>
              </div>
              <!-- Midtrans Info -->
              <div v-if="selectedNota.raw?.midtrans_qris_url" class="flex flex-col items-start sm:items-end print:items-end">
                <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest mb-1">
                  SCAN QRIS
                </p>
                <img
                  :src="`https://api.qrserver.com/v1/create-qr-code/?size=100x100&data=${selectedNota.raw.midtrans_qris_url}`"
                  class="w-24 h-24 border border-slate-200 p-1 rounded"
                />
              </div>
              <div
                v-if="selectedNota.raw?.midtrans_va_number || selectedNota.raw?.midtrans_bill_key"
                class="text-left sm:text-right print:text-right"
              >
                <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest mb-1 print:text-black print:text-[10px]">
                  {{ selectedNota.raw?.midtrans_bank?.toUpperCase() }} VA
                </p>
                <p class="text-[14px] font-black text-slate-900 tracking-wider print:text-black print:text-[12px]">
                  {{ selectedNota.raw?.midtrans_va_number || selectedNota.raw?.midtrans_bill_key }}
                </p>
                <p
                  v-if="selectedNota.raw?.midtrans_biller_code"
                  class="text-[10px] font-bold text-slate-500 print:text-black print:font-normal"
                >
                  Biller: {{ selectedNota.raw?.midtrans_biller_code }}
                </p>
              </div>

              <!-- Bukti Transfer Lampiran -->
              <div v-if="selectedNotaTransferProof" class="flex flex-col items-start sm:items-end print:hidden">
                <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest mb-1">
                  BUKTI TRANSFER
                </p>
                <button
                  @click="openImagePreview(selectedNotaTransferProof)"
                  type="button"
                  class="flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-black text-indigo-600 bg-indigo-50 hover:bg-indigo-100 border border-indigo-200 transition-all shadow-sm cursor-pointer"
                >
                  <ExternalLink class="w-3.5 h-3.5" />
                  <span>Lihat Bukti Transfer</span>
                </button>
              </div>
            </div>
          </div>

          <!-- Items Table -->
          <div class="border border-slate-900 rounded-xl overflow-x-auto mb-8 print:mb-2 print:border-black print:rounded-none">
            <table class="w-full text-left text-xs border-collapse min-w-[480px] sm:min-w-full">
              <thead>
                <tr class="bg-slate-50 border-b border-slate-900 print:bg-white print:border-black print:border-b-2">
                  <th class="p-3 print:p-1 font-black text-center w-12 border-r border-slate-300 print:border-black print:text-black print:font-bold">
                    NO
                  </th>
                  <th class="p-3 print:p-1 font-black border-r border-slate-300 print:border-black print:text-black print:font-bold">NAMA PRODUK</th>
                  <th class="p-3 print:p-1 font-black text-center w-20 border-r border-slate-300 print:border-black print:text-black print:font-bold">
                    QTY
                  </th>
                  <th class="p-3 print:p-1 font-black text-right w-32 border-r border-slate-300 print:border-black print:text-black print:font-bold">
                    HARGA
                  </th>
                  <th class="p-3 print:p-1 font-black text-right w-32 bg-slate-100/50 print:bg-white print:text-black print:font-bold">SUBTOTAL</th>
                </tr>
              </thead>
              <tbody class="divide-y border-slate-300">
                <tr
                  v-for="(item, idx) in selectedNota.raw?.items"
                  :key="item.id"
                  class="h-10 print:h-6 print:text-[10px]"
                >
                  <td class="p-3 print:p-1 text-center font-bold border-r border-slate-300 print:border-black print:text-black print:font-normal">
                    {{ (idx as number) + 1 }}
                  </td>
                  <td
                    class="p-3 print:p-1 font-bold text-slate-900 border-r border-slate-300 uppercase print:border-black print:text-black print:font-normal"
                  >
                    {{ item.product?.name }}
                  </td>
                  <td class="p-3 print:p-1 text-center font-bold border-r border-slate-300 print:border-black print:text-black print:font-normal">
                    {{ item.quantity }}
                  </td>
                  <td class="p-3 print:p-1 text-right font-bold border-r border-slate-300 print:border-black print:text-black print:font-normal">
                    {{ formatCurrency(item.price_at_transaction).replace('Rp', '').trim() }}
                  </td>
                  <td class="p-3 print:p-1 text-right font-black bg-slate-50/50 print:bg-white print:text-black print:font-normal">
                    {{ formatCurrency(item.subtotal).replace('Rp', '').trim() }}
                  </td>
                </tr>
                <tr
                  v-for="(bonus, idx) in selectedNota.raw?.bonus_items"
                  :key="'bonus-'+bonus.id"
                  class="h-10 print:h-6 print:text-[10px] bg-emerald-50/50"
                >
                  <td class="p-3 print:p-1 text-center font-bold text-emerald-600 border-r border-emerald-200 print:border-black print:text-black print:font-normal">
                    *
                  </td>
                  <td
                    class="p-3 print:p-1 font-bold text-emerald-700 border-r border-emerald-200 uppercase flex flex-col print:border-black print:text-black print:font-normal"
                  >
                    <span>{{ bonus.bonus_inventory?.name }} (BONUS)</span>
                    <span v-if="bonus.notes" class="text-[9px] text-emerald-500 font-medium italic lowercase print:text-black">{{ bonus.notes }}</span>
                  </td>
                  <td class="p-3 print:p-1 text-center font-bold text-emerald-700 border-r border-emerald-200 print:border-black print:text-black print:font-normal">
                    {{ bonus.quantity }}
                  </td>
                  <td class="p-3 print:p-1 text-right font-bold text-emerald-700 border-r border-emerald-200 print:border-black print:text-black print:font-normal">
                    0
                  </td>
                  <td class="p-3 print:p-1 text-right font-black text-emerald-700 bg-emerald-100/50 print:bg-white print:text-black print:font-normal">
                    0
                  </td>
                </tr>
                 <!-- Padding Rows if few items -->
                <tr
                  v-for="n in Math.max(0, 4 - (selectedNota.raw?.items?.length || 0))"
                  :key="'blank-' + n"
                  class="h-10 print:h-5 print:hidden"
                >
                  <td class="p-3 print:p-1 border-r border-slate-300 print:border-black"></td>
                  <td class="p-3 print:p-1 border-r border-slate-300 print:border-black"></td>
                  <td class="p-3 print:p-1 border-r border-slate-300 print:border-black"></td>
                  <td class="p-3 print:p-1 border-r border-slate-300 print:border-black"></td>
                  <td class="p-3 print:p-1 bg-slate-50/20 print:bg-white"></td>
                </tr>
              </tbody>
              <tfoot>
                <!-- Breakdown if discounts or returns exist -->
                <template v-if="notaTotalReturn > 0 || notaDiscountAmount > 0">
                  <tr class="bg-slate-50 text-slate-600 font-bold print:text-[10px] print:bg-white print:text-black">
                    <td colspan="4" class="p-2 print:p-1 text-right uppercase tracking-widest text-[10px] print:border-t print:border-black print:font-bold">
                      TOTAL BELANJA
                    </td>
                    <td class="p-2 print:p-1 text-right text-sm print:text-[10px] print:border-t print:border-black print:font-bold">
                      {{ formatCurrency(notaBaseTotal) }}
                    </td>
                  </tr>
                  <tr v-if="notaDiscountAmount > 0" class="bg-rose-50 text-rose-700 font-bold print:text-[10px] print:bg-white print:text-black">
                    <td colspan="4" class="p-2 print:p-1 text-right uppercase tracking-widest text-[10px] print:font-bold">
                      POTONGAN DISKON NOTA
                    </td>
                    <td class="p-2 print:p-1 text-right text-sm print:text-[10px] print:font-bold">
                      -{{ formatCurrency(notaDiscountAmount) }}
                    </td>
                  </tr>
                  <tr v-if="notaTotalReturn > 0" class="bg-rose-50 text-rose-700 font-bold print:text-[10px] print:bg-white print:text-black">
                    <td colspan="4" class="p-2 print:p-1 text-right uppercase tracking-widest text-[10px] print:font-bold">
                      POTONGAN RETUR SAMPINGAN
                    </td>
                    <td class="p-2 print:p-1 text-right text-sm print:text-[10px] print:font-bold">
                      -{{ formatCurrency(notaTotalReturn) }}
                    </td>
                  </tr>
                </template>
                <tr class="bg-slate-900 text-white font-black print:text-[10px] print:bg-white print:text-black print:border-t-2 print:border-black">
                  <td
                    colspan="4"
                    class="p-4 print:p-1.5 text-right uppercase tracking-widest text-[10px] print:font-bold"
                  >
                    {{ (notaTotalReturn > 0 || notaDiscountAmount > 0) ? 'TOTAL BAYAR AKHIR' : 'TOTAL TAGIHAN' }}
                  </td>
                  <td class="p-4 print:p-1.5 text-right text-base print:text-[11px] print:font-bold">
                    {{ formatCurrency(selectedNota.amount) }}
                  </td>
                </tr>
              </tfoot>
            </table>
          </div>

          <!-- Return Items from Mobile App JSON -->
          <div
            v-if="parsedNotaNotes?.return_items && parsedNotaNotes.return_items.length > 0"
            class="mb-8 print:mb-2 border-2 border-rose-900 rounded-xl overflow-x-auto"
          >
            <div class="bg-rose-50 p-2 font-black text-rose-900 text-xs uppercase text-center border-b-2 border-rose-900 tracking-widest">
              RETUR / PENGEMBALIAN BARANG SAMPINGAN
            </div>
            <table class="w-full text-left text-xs border-collapse min-w-[480px] sm:min-w-full">
              <thead>
                <tr class="bg-rose-50 border-b-2 border-rose-900">
                  <th class="p-3 print:p-1 font-black text-rose-900 border-r border-rose-300">NAMA PRODUK</th>
                  <th class="p-3 print:p-1 font-black text-center w-20 text-rose-900 border-r border-rose-300">QTY</th>
                  <th class="p-3 print:p-1 font-black text-right w-32 text-rose-900 border-r border-rose-300">HARGA</th>
                  <th class="p-3 print:p-1 font-black text-right w-32 text-rose-900">POTONGAN</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="(ret, idx) in parsedNotaNotes.return_items"
                  :key="'ret-'+idx"
                  class="h-10 print:h-6 print:text-[10px]"
                >
                  <td class="p-3 print:p-1 font-bold text-rose-900 border-r border-rose-300 uppercase">
                    {{ ret.product_name }}
                  </td>
                  <td class="p-3 print:p-1 text-center font-bold text-rose-900 border-r border-rose-300">
                    {{ ret.quantity }} {{ ret.unit || 'PCS' }}
                  </td>
                  <td class="p-3 print:p-1 text-right font-medium text-rose-900 border-r border-rose-300">
                    {{ formatCurrency(ret.price).replace('Rp', '').trim() }}
                  </td>
                  <td class="p-3 print:p-1 text-right font-black text-rose-700 bg-rose-50">
                    -{{ formatCurrency(ret.quantity * ret.price).replace('Rp', '').trim() }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Logs Retur Pelanggan (Pengurangan Faktur) -->
          <div
            v-if="selectedNota.raw?.returns && selectedNota.raw.returns.length > 0"
            class="mt-6 border border-amber-200 dark:border-amber-900/60 rounded-xl overflow-hidden bg-amber-50/5 dark:bg-amber-950/10 print:hidden"
          >
            <div class="bg-amber-50 dark:bg-amber-950/40 p-3 border-b border-amber-200 dark:border-amber-900/60 flex items-center gap-2 text-amber-800 dark:text-amber-300 font-bold text-xs uppercase tracking-wider">
              <RotateCcw class="w-4 h-4" />
              <span>Histori Retur/Pengembalian Barang Pelanggan</span>
            </div>
            <div class="p-4 space-y-4">
              <div
                v-for="ret in selectedNota.raw.returns"
                :key="ret.id"
                class="border border-slate-200 dark:border-slate-800 rounded-lg p-3 bg-card"
              >
                <div class="flex items-center justify-between gap-2 border-b pb-2 mb-2">
                  <div>
                    <span class="font-mono font-bold text-xs text-primary">{{ ret.return_no }}</span>
                    <span class="text-[10px] text-muted-foreground ml-2">({{ formatDate(ret.created_at) }})</span>
                  </div>
                  <Badge
                    :variant="ret.status === 'APPROVED' ? 'default' : 'secondary'"
                    class="text-[10px] uppercase font-bold"
                    :class="ret.status === 'APPROVED' ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-200' : 'bg-yellow-100 text-yellow-800 dark:bg-yellow-950 dark:text-yellow-200'"
                  >
                    {{ ret.status === 'APPROVED' ? 'Disetujui' : 'Tertunda (Pending)' }}
                  </Badge>
                </div>
                
                <!-- Return items -->
                <div class="space-y-1.5">
                  <div class="grid grid-cols-3 text-[10px] font-bold text-muted-foreground border-b pb-1">
                    <span>Produk</span>
                    <span class="text-center">Kuantitas</span>
                    <span class="text-right">Alasan</span>
                  </div>
                  <div
                    v-for="item in ret.items"
                    :key="item.id"
                    class="grid grid-cols-3 text-xs py-1 border-b last:border-0 border-slate-100 dark:border-slate-900"
                  >
                    <span class="font-medium text-left">{{ item.product?.name || 'Produk' }}</span>
                    <span class="text-center font-bold text-red-500">{{ item.quantity }} Pcs</span>
                    <span class="text-right text-muted-foreground">{{ item.reason || '-' }}</span>
                  </div>
                </div>

                <div class="flex justify-between items-center mt-3 pt-2 border-t border-dashed">
                  <span class="text-xs text-muted-foreground">Catatan: {{ ret.notes || '-' }}</span>
                  <span class="text-xs font-bold text-red-500">
                    Nilai Pengurangan: -{{ formatCurrency(ret.total_amount) }}
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- Log Penyesuaian Kuantitas (Admin Nota) -->
          <div
            v-if="selectedNota.raw?.adjustment_logs && selectedNota.raw.adjustment_logs.length > 0"
            class="mt-6 border border-blue-200 dark:border-blue-900/60 rounded-xl overflow-hidden bg-blue-50/5 dark:bg-blue-950/10 print:hidden"
          >
            <div class="bg-blue-50 dark:bg-blue-950/40 p-3 border-b border-blue-200 dark:border-blue-900/60 flex items-center gap-2 text-blue-800 dark:text-blue-300 font-bold text-xs uppercase tracking-wider">
              <AlertCircle class="w-4 h-4" />
              <span>Log Penyesuaian Kuantitas Faktur (Admin Nota)</span>
            </div>
            <div class="p-3 overflow-x-auto">
              <table class="w-full text-xs text-left min-w-[480px] sm:min-w-full">
                <thead>
                  <tr class="text-muted-foreground border-b font-semibold">
                    <th class="p-2">Produk</th>
                    <th class="p-2 text-center">Qty Awal</th>
                    <th class="p-2 text-center">Selisih</th>
                    <th class="p-2 text-center">Qty Akhir</th>
                    <th class="p-2">Catatan</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="log in selectedNota.raw.adjustment_logs"
                    :key="log.id"
                    class="border-b last:border-0"
                  >
                    <td class="p-2 font-medium">{{ log.product?.name || 'Produk' }}</td>
                    <td class="p-2 text-center text-muted-foreground">{{ log.original_quantity }}</td>
                    <td class="p-2 text-center text-red-500 font-bold">-{{ log.difference }}</td>
                    <td class="p-2 text-center text-emerald-500 font-bold">{{ log.adjusted_quantity }}</td>
                    <td class="p-2 text-muted-foreground">{{ log.notes || '-' }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Riwayat & Jadwal Angsuran Tempo (Multi-Tier Payment Tracking) -->
          <div
            v-if="isTempoEligibleForPayment(selectedNota) || String(selectedNota.raw?.payment_method || '').toUpperCase().includes('TEMPO') || (selectedNota.raw?.payments && selectedNota.raw.payments.length > 0)"
            class="mt-6 border border-indigo-200 dark:border-indigo-900/60 rounded-2xl overflow-hidden bg-indigo-50/10 dark:bg-indigo-950/10 print:hidden"
          >
            <div class="bg-indigo-50/80 dark:bg-indigo-950/40 p-3.5 border-b border-indigo-200 dark:border-indigo-900/60 flex items-center justify-between flex-wrap gap-2 text-indigo-900 dark:text-indigo-200">
              <div class="flex items-center gap-2 font-black text-xs uppercase tracking-wider">
                <Wallet class="w-4 h-4 text-indigo-600" />
                <span>Progress & Riwayat Pembayaran Tempo</span>
              </div>
              <div class="flex items-center gap-2">
                <span
                  class="text-[10px] font-black px-2.5 py-1 rounded-full uppercase border"
                  :class="getTempoProgressInfo(selectedNota).isPaid ? 'bg-emerald-100 text-emerald-800 border-emerald-300' : (getTempoProgressInfo(selectedNota).isOverdue ? 'bg-rose-100 text-rose-700 border-rose-300' : 'bg-amber-100 text-amber-800 border-amber-300')"
                >
                  {{ getTempoProgressInfo(selectedNota).isPaid ? '✓ LUNAS (100%)' : (getTempoProgressInfo(selectedNota).isOverdue ? '⚠ JATUH TEMPO' : `CICILAN BERJALAN (${getTempoProgressInfo(selectedNota).percent}%)`) }}
                </span>
                <Button
                  v-if="isTempoEligibleForPayment(selectedNota)"
                  size="sm"
                  class="bg-indigo-600 hover:bg-indigo-700 text-white font-bold text-xs h-7 rounded-lg shadow-sm flex items-center gap-1 cursor-pointer"
                  @click="openTempoPaymentDialog"
                >
                  <Wallet class="w-3.5 h-3.5 mr-0.5" /> + Bayar Cicilan
                </Button>
              </div>
            </div>

            <!-- Ringkasan Finansial Tagihan Tempo -->
            <div class="p-4 grid grid-cols-2 sm:grid-cols-4 gap-3 bg-white border-b border-indigo-100">
              <div class="bg-slate-50 p-3 rounded-xl border border-slate-100">
                <p class="text-[9px] font-bold text-slate-400 uppercase tracking-wider">Total Nilai Faktur</p>
                <p class="text-sm font-black text-slate-900 mt-0.5">{{ formatCurrency(getTempoProgressInfo(selectedNota).totalAmount) }}</p>
              </div>
              <div class="bg-emerald-50/60 p-3 rounded-xl border border-emerald-100">
                <p class="text-[9px] font-bold text-emerald-600 uppercase tracking-wider">Total Sudah Dibayar</p>
                <p class="text-sm font-black text-emerald-700 mt-0.5">{{ formatCurrency(getTempoProgressInfo(selectedNota).paidAmount) }}</p>
              </div>
              <div class="bg-amber-50/60 p-3 rounded-xl border border-amber-100">
                <p class="text-[9px] font-bold text-amber-600 uppercase tracking-wider">Sisa Tagihan (Piutang)</p>
                <p class="text-sm font-black text-amber-700 mt-0.5">{{ formatCurrency(getTempoProgressInfo(selectedNota).remainingAmount) }}</p>
              </div>
              <div class="bg-indigo-50/60 p-3 rounded-xl border border-indigo-100">
                <p class="text-[9px] font-bold text-indigo-600 uppercase tracking-wider">Target / Batas Cicilan</p>
                <p class="text-xs font-black text-indigo-800 mt-0.5">{{ getTempoProgressInfo(selectedNota).maxTenorText }}</p>
                <p class="text-[9px] font-medium text-indigo-600 mt-0.5">
                  <span v-if="getTempoProgressInfo(selectedNota).isExceedingPlannedTenor">
                    Sudah {{ getTempoProgressInfo(selectedNota).paidTermsCount }}x cicilan (Fleksibel)
                  </span>
                  <span v-else>
                    Sudah {{ getTempoProgressInfo(selectedNota).paidTermsCount }}x • Sisa {{ getTempoProgressInfo(selectedNota).remainingTermsCount }}x lagi
                  </span>
                </p>
              </div>
            </div>

            <!-- Tabel Riwayat Pembayaran (SalesPayment Logs) -->
            <div class="p-4 space-y-2">
              <p class="text-[11px] font-black text-slate-600 uppercase tracking-wider">Riwayat Log Penagihan / Pembayaran</p>
              <div v-if="!selectedNota.raw?.payments || selectedNota.raw.payments.length === 0" class="text-center py-6 bg-slate-50 rounded-xl border border-dashed text-slate-400 text-xs font-medium">
                Belum ada log pembayaran cicilan tercatat. Klik tombol "+ Bayar Cicilan" untuk mencatat pembayaran baru.
              </div>
              <div v-else class="overflow-x-auto rounded-xl border border-slate-200">
                <table class="w-full text-xs text-left">
                  <thead class="bg-slate-50 text-slate-500 font-bold uppercase text-[9px] border-b border-slate-200">
                    <tr>
                      <th class="p-2.5 text-center w-12">No.</th>
                      <th class="p-2.5">Tanggal Bayar</th>
                      <th class="p-2.5">Metode</th>
                      <th class="p-2.5 text-right">Nominal</th>
                      <th class="p-2.5">Diterima / Dicatat</th>
                      <th class="p-2.5 text-center">Bukti Transfer</th>
                      <th class="p-2.5 text-center">Status Kasir</th>
                      <th class="p-2.5">Catatan</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-slate-100 bg-white">
                    <tr v-for="(pay, pIdx) in selectedNota.raw.payments" :key="pay.id || pIdx" class="hover:bg-slate-50/60">
                      <td class="p-2.5 text-center font-black text-indigo-700">#{{ Number(pIdx) + 1 }}</td>
                      <td class="p-2.5 font-medium text-slate-700">{{ formatDate(pay.payment_date || pay.created_at) }}</td>
                      <td class="p-2.5">
                        <span class="px-2 py-0.5 rounded text-[9px] font-black uppercase border" :class="pay.payment_method === 'TRANSFER' || pay.payment_method === 'VA' ? 'bg-indigo-50 text-indigo-700 border-indigo-200' : 'bg-slate-50 text-slate-700 border-slate-200'">
                          {{ pay.payment_method }}
                        </span>
                        <div v-if="pay.company_bank_account" class="text-[10px] text-indigo-700 mt-1">
                          {{ pay.company_bank_account.bank_name }} • {{ pay.company_bank_account.account_number }}
                        </div>
                      </td>
                      <td class="p-2.5 text-right font-black text-emerald-600">{{ formatCurrency(pay.amount) }}</td>
                      <td class="p-2.5 text-slate-600">
                        {{ pay.collected_by || (pay.employee ? `${pay.employee.first_name} ${pay.employee.last_name || ''}` : 'Admin') }}
                      </td>
                      <td class="p-2.5 text-center">
                        <button
                          v-if="pay.transfer_proof_url"
                          @click="openImagePreview(pay.transfer_proof_url)"
                          type="button"
                          class="px-2 py-1 bg-indigo-50 hover:bg-indigo-100 text-indigo-700 rounded text-[10px] font-bold inline-flex items-center gap-1 border border-indigo-200"
                        >
                          <ExternalLink class="w-2.5 h-2.5" /> Lihat Bukti
                        </button>
                        <span v-else class="text-slate-300 text-[10px]">-</span>
                      </td>
                      <td class="p-2.5 text-center">
                        <span
                          class="px-2 py-0.5 rounded-full text-[9px] font-black uppercase border"
                          :class="pay.payment_status === 'SUCCESS' ? 'bg-emerald-50 text-emerald-700 border-emerald-200' : (pay.payment_status === 'FAILED' ? 'bg-rose-50 text-rose-700 border-rose-200' : 'bg-amber-50 text-amber-700 border-amber-200')"
                        >
                          {{ pay.payment_status === 'SUCCESS' ? 'Disetujui' : (pay.payment_status === 'FAILED' ? 'Ditolak' : 'Menunggu') }}
                        </span>
                      </td>
                      <td class="p-2.5 text-slate-500 italic max-w-[200px] truncate" :title="pay.notes || '-'">{{ pay.notes || '-' }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <!-- Tabel Jadwal Termin (SalesInstallment jika ada) -->
            <div v-if="selectedNota.raw?.installments && selectedNota.raw.installments.length > 0" class="p-4 pt-0 space-y-2">
              <p class="text-[11px] font-black text-slate-600 uppercase tracking-wider">Jadwal Termin / Angsuran Berjangka</p>
              <div class="overflow-x-auto rounded-xl border border-slate-200">
                <table class="w-full text-xs text-left">
                  <thead class="bg-slate-50 text-slate-500 font-bold uppercase text-[9px] border-b border-slate-200">
                    <tr>
                      <th class="p-2.5 text-center w-16">Termin</th>
                      <th class="p-2.5">Jatuh Tempo</th>
                      <th class="p-2.5 text-right">Target Tagihan</th>
                      <th class="p-2.5 text-right">Sudah Dibayar</th>
                      <th class="p-2.5 text-center">Status Termin</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-slate-100 bg-white">
                    <tr v-for="inst in selectedNota.raw.installments" :key="inst.id || inst.term_number" class="hover:bg-slate-50/60">
                      <td class="p-2.5 text-center font-black text-indigo-700">Ke-{{ inst.term_number }}</td>
                      <td class="p-2.5 font-medium text-slate-700">{{ formatDate(inst.due_date) }}</td>
                      <td class="p-2.5 text-right font-bold text-slate-900">{{ formatCurrency(inst.amount) }}</td>
                      <td class="p-2.5 text-right font-black text-emerald-600">{{ formatCurrency(inst.paid_amount || 0) }}</td>
                      <td class="p-2.5 text-center">
                        <span
                          class="px-2 py-0.5 rounded-full text-[9px] font-black uppercase border"
                          :class="inst.status === 'PAID' ? 'bg-emerald-50 text-emerald-700 border-emerald-200' : 'bg-amber-50 text-amber-700 border-amber-200'"
                        >
                          {{ inst.status === 'PAID' ? 'LUNAS' : 'BELUM LUNAS' }}
                        </span>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>

          <!-- Footer/Signatures -->
          <div class="grid grid-cols-2 sm:grid-cols-5 print:grid-cols-5 gap-3 mt-6 print:mt-1 print:gap-1 items-end border-t-2 border-slate-900 pt-4 print:pt-1">
            <!-- Col 1: Pelanggan -->
            <div class="text-center space-y-1 print:space-y-0.5">
              <p
                class="text-[9px] font-black uppercase tracking-widest text-slate-400 print:text-[7px]"
              >
                PELANGGAN
              </p>
              <div
                class="h-16 print:h-8 flex items-center justify-center border border-dashed border-slate-200 rounded-xl bg-slate-50/50 p-2 print:p-0.5"
              >
                <img
                  v-if="getCustomerSignature(selectedNota)"
                  :src="getCustomerSignature(selectedNota)"
                  alt="Customer Signature"
                  class="max-h-12 print:max-h-6 object-contain"
                />
                <span
                  v-else
                  class="text-[8px] print:text-[6px] font-black text-slate-300 uppercase tracking-widest"
                  >Belum ttd</span
                >
              </div>
              <p
                class="text-[9px] font-black uppercase text-slate-800 leading-none mt-1 print:text-[7px] print:mt-0.5"
              >
                ( {{ selectedNota.store }} )
              </p>
            </div>

            <!-- Col 2: Courier / Salesman -->
            <div class="text-center space-y-1 print:space-y-0.5">
              <p
                class="text-[9px] font-black uppercase tracking-widest text-slate-400 print:text-[7px]"
              >
                {{ getCourierTitle(selectedNota) }}
              </p>
              <div
                class="h-16 print:h-8 flex items-center justify-center border border-dashed border-slate-200 rounded-xl bg-slate-50/50 p-2 print:p-0.5"
              >
                <img
                  v-if="getCourierSignature(selectedNota)"
                  :src="getCourierSignature(selectedNota)"
                  alt="Courier Signature"
                  class="max-h-12 print:max-h-6 object-contain"
                />
                <span
                  v-else
                  class="text-[8px] print:text-[6px] font-black text-slate-300 uppercase tracking-widest"
                  >Belum ttd</span
                >
              </div>
              <p
                class="text-[9px] font-black uppercase text-slate-800 leading-none mt-1 print:text-[7px] print:mt-0.5"
              >
                ( {{ getCourierName(selectedNota) }} )
              </p>
            </div>

            <!-- Col 3: SPV -->
            <div class="text-center space-y-1 print:space-y-0.5">
              <p
                class="text-[9px] font-black uppercase tracking-widest text-slate-400 print:text-[7px]"
              >
                SPV / KEPALA
              </p>
              <div
                class="h-16 print:h-8 flex items-center justify-center border border-dashed border-slate-200 rounded-xl bg-slate-50/50 p-2 print:p-0.5"
              >
                <span
                  class="text-[8px] print:text-[6px] font-black text-slate-300 uppercase tracking-widest"
                  >Belum ttd</span
                >
              </div>
              <p
                class="text-[9px] font-black uppercase text-slate-800 leading-none mt-1 print:text-[7px] print:mt-0.5"
              >
                ( .................... )
              </p>
            </div>

            <!-- Col 4: Disiapkan & Disetujui -->
            <div class="text-center space-y-1 print:space-y-0.5 text-[9px] print:text-[7px] text-slate-800 font-bold border-l border-slate-200 pl-2">
              <div class="text-left">
                <span class="text-slate-400 font-black uppercase tracking-tight text-[8px] print:text-[6px]">TGL DISIAPKAN:</span>
                <p class="font-black text-slate-900 leading-tight">
                  {{ formatDate(selectedNota.rawDate) }}
                </p>
              </div>
              <div class="text-left border-t border-slate-100 pt-1 mt-1">
                <span class="text-slate-400 font-black uppercase tracking-tight text-[8px] print:text-[6px]">DISETUJUI OLEH:</span>
                <p class="font-black text-slate-900 leading-tight">
                  {{ selectedNota.raw?.verified_by ? `${selectedNota.raw.verified_by.first_name} ${selectedNota.raw.verified_by.last_name || ''}` : (selectedNota.raw?.verified_by_user ? `${selectedNota.raw.verified_by_user.first_name} ${selectedNota.raw.verified_by_user.last_name || ''}` : 'ADMIN NOTA') }}
                </p>
                <p v-if="selectedNota.raw?.verified_by?.job_position?.title" class="text-[8px] print:text-[6px] text-slate-500 font-normal">
                  ({{ selectedNota.raw.verified_by.job_position.title }})
                </p>
              </div>
            </div>

            <!-- Col 5: Validasi QR -->
            <div class="col-span-2 sm:col-span-1 print:col-span-1 flex justify-center sm:justify-end print:justify-end items-center pr-0 sm:pr-2 print:pr-0 pt-2 sm:pt-0">
              <div class="text-right mr-3 print:mr-1">
                <p
                  class="text-[8px] print:text-[6px] font-black text-slate-400 uppercase leading-none"
                >
                  VALIDASI NOTA
                </p>
                <p class="text-[6px] print:text-[5px] text-slate-300 italic mt-0.5 uppercase">
                  DIGITAL RECORD
                </p>
              </div>
              <img
                :src="`https://api.qrserver.com/v1/create-qr-code/?size=100x100&data=${selectedNota.receiptNo || selectedNota.id}`"
                class="w-12 h-12 sm:w-14 sm:h-14 print:w-9 print:h-9 border-2 border-slate-900 p-1 rounded bg-white shadow-sm"
              />
            </div>
          </div>
        </div>

        <CardFooter class="p-4 sm:p-6 bg-slate-50 border-t flex flex-col-reverse sm:flex-row justify-end gap-2 sm:gap-3 print:hidden">
          <Button variant="ghost" class="font-bold text-slate-500 w-full sm:w-auto" @click="selectedNota = null"
            >BATAL</Button
          >
          <Button
            v-if="isTempoEligibleForPayment(selectedNota)"
            class="bg-indigo-600 hover:bg-indigo-700 text-white font-black px-4 sm:px-8 py-3.5 sm:py-6 rounded-xl sm:rounded-2xl shadow-xl shadow-indigo-200 w-full sm:w-auto flex items-center justify-center text-xs sm:text-base cursor-pointer"
            @click="openTempoPaymentDialog"
          >
            <Wallet class="w-4 h-4 sm:w-5 sm:h-5 mr-2 sm:mr-3" />
            TERIMA PEMBAYARAN TEMPO
          </Button>

          <Button
            class="bg-slate-900 hover:bg-slate-800 text-white font-black px-6 sm:px-12 py-3.5 sm:py-6 rounded-xl sm:rounded-2xl shadow-xl shadow-slate-200 w-full sm:w-auto flex items-center justify-center text-xs sm:text-base"
            @click="handlePrint"
          >
            <Printer class="w-4 h-4 sm:w-5 sm:h-5 mr-2 sm:mr-3" />
            CETAK NOTA
          </Button>
        </CardFooter>
      </Card>
    </div>

    <!-- Receipt QR Modal -->
    <div
      v-if="selectedQR"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-sm z-[250] flex items-center justify-center p-6"
      @click="selectedQR = null"
    >
      <Card class="w-full max-w-sm bg-white shadow-2xl rounded-[32px] overflow-hidden" @click.stop>
        <div class="p-8 text-center space-y-6">
          <div>
            <h3 class="text-lg font-black text-slate-900 uppercase tracking-tighter italic">
              Scan QR Nota
            </h3>
            <p class="text-xs font-bold text-slate-400 mt-1 uppercase tracking-widest">
              {{ selectedQR }}
            </p>
          </div>

          <div
            class="bg-slate-50 p-6 rounded-3xl border-2 border-slate-100 flex items-center justify-center"
          >
            <img
              :src="`https://api.qrserver.com/v1/create-qr-code/?size=300x300&data=${selectedQR}`"
              class="w-64 h-64 shadow-sm"
            />
          </div>

          <Button
            class="w-full bg-slate-900 hover:bg-slate-800 text-white rounded-2xl py-6 font-black"
            @click="selectedQR = null"
          >
            TUTUP
          </Button>
        </div>
      </Card>
    </div>

    <!-- Tempo Payment Modal -->
    <div
      v-if="showTempoPaymentModal"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-sm z-[300] flex items-center justify-center p-4 sm:p-6 overflow-y-auto"
      @click="showTempoPaymentModal = false"
    >
      <Card class="w-full max-w-lg bg-white shadow-2xl rounded-[32px] overflow-hidden my-auto" @click.stop>
        <div class="p-6 sm:p-8 space-y-5">
          <div class="text-center">
            <div
              class="w-14 h-14 bg-indigo-50 text-indigo-600 rounded-full flex items-center justify-center mx-auto mb-3 border-4 border-indigo-100"
            >
              <Wallet class="w-7 h-7" />
            </div>
            <div class="inline-flex items-center gap-1.5 px-3.5 py-1 bg-indigo-50 text-indigo-700 rounded-full text-xs font-black uppercase mb-1 border border-indigo-100">
              <span v-if="getTempoProgressInfo(selectedNota).isExceedingPlannedTenor">
                Mencatat Cicilan ke-{{ getTempoProgressInfo(selectedNota).nextTermNumber }} (Fleksibel • Target: {{ getTempoProgressInfo(selectedNota).maxTenorText }})
              </span>
              <span v-else-if="getTempoProgressInfo(selectedNota).targetTenor > 1">
                Mencatat Cicilan ke-{{ getTempoProgressInfo(selectedNota).nextTermNumber }} dari {{ getTempoProgressInfo(selectedNota).maxTenorText }}
              </span>
              <span v-else>
                Mencatat Pembayaran ke-{{ getTempoProgressInfo(selectedNota).nextPaymentNumber }} (Fleksibel)
              </span>
            </div>
            <h3 class="text-xl font-black text-slate-900 uppercase tracking-tight mt-1">
              {{ selectedNota?.receiptNo || 'Faktur Tempo' }}
            </h3>
            <p class="text-xs font-bold text-slate-500 mt-0.5">
              {{ selectedNota?.store || 'Pelanggan' }} • {{ selectedNota?.salesman || 'Salesman' }}
            </p>
            <p class="text-[11px] font-bold text-indigo-600 mt-1">
              <span v-if="getTempoProgressInfo(selectedNota).isExceedingPlannedTenor">
                Toko telah mencicil {{ getTempoProgressInfo(selectedNota).paidTermsCount }}x (Target awal: {{ getTempoProgressInfo(selectedNota).maxTenorText }}). Pembayaran cicilan tetap dapat dicatat fleksibel hingga lunas.
              </span>
              <span v-else-if="getTempoProgressInfo(selectedNota).targetTenor > 1">
                Target Rencana: <span class="font-black">{{ getTempoProgressInfo(selectedNota).maxTenorText }}</span>
                <span v-if="getTempoProgressInfo(selectedNota).remainingTermsCount > 0"> (Tersisa {{ getTempoProgressInfo(selectedNota).remainingTermsCount }}x lagi, bisa lebih jika fleksibel)</span>
              </span>
              <span v-else>
                Pembayaran cicilan faktur fleksibel dapat dicatat bertahap hingga lunas.
              </span>
            </p>
          </div>

          <!-- Mini Stats -->
          <div class="grid grid-cols-3 gap-2 bg-slate-50 p-3.5 rounded-2xl border border-slate-100 text-center">
            <div>
              <p class="text-[9px] font-bold text-slate-400 uppercase">Total Nilai Faktur</p>
              <p class="text-xs font-black text-slate-800 mt-0.5">{{ formatCurrency(getTempoProgressInfo(selectedNota).totalAmount) }}</p>
            </div>
            <div>
              <p class="text-[9px] font-bold text-emerald-600 uppercase">Sudah Dibayar</p>
              <p class="text-xs font-black text-emerald-700 mt-0.5">{{ formatCurrency(getTempoProgressInfo(selectedNota).paidAmount) }}</p>
            </div>
            <div>
              <p class="text-[9px] font-bold text-amber-600 uppercase">Sisa Tagihan</p>
              <p class="text-xs font-black text-amber-700 mt-0.5">{{ formatCurrency(getTempoProgressInfo(selectedNota).remainingAmount) }}</p>
            </div>
          </div>

          <div class="space-y-4 pt-1">
            <!-- Metode Bayar -->
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div>
                <label class="text-xs font-black text-slate-700 uppercase tracking-widest"
                  >Metode Pembayaran</label
                >
                <select
                  v-model="tempoPaymentMethod"
                  class="w-full mt-1.5 rounded-xl border-slate-200 bg-slate-50 px-4 py-3 font-semibold text-sm focus:border-indigo-500 focus:ring-indigo-500 transition-colors"
                >
                  <option value="CASH">CASH (Tunai / Titip Sales / Setor Driver)</option>
                  <option value="TRANSFER">TRANSFER BANK (Rekening Resmi Perusahaan)</option>
                  <option value="QRIS">QRIS / Virtual Account</option>
                </select>
              </div>

              <!-- Tanggal Uang Masuk / Pembayaran -->
              <div>
                <label class="text-xs font-black text-slate-700 uppercase tracking-widest flex items-center justify-between"
                  ><span>Tgl Uang Masuk *</span>
                  <span class="text-[10px] text-indigo-600 font-bold lowercase">sesuai kas masuk</span>
                </label>
                <input
                  type="date"
                  v-model="tempoPaymentDate"
                  class="w-full mt-1.5 rounded-xl border-slate-200 bg-slate-50 px-4 py-2.5 font-bold text-xs focus:border-indigo-500 focus:ring-indigo-500 transition-colors"
                />
              </div>
            </div>

            <div v-if="tempoPaymentMethod === 'TRANSFER'">
              <label class="text-xs font-black text-slate-700 uppercase tracking-widest">Rekening Bank Penerima *</label>
              <select v-model="tempoBankAccountId" class="w-full mt-1.5 rounded-xl border-slate-200 bg-slate-50 px-4 py-3 font-semibold text-sm">
                <option value="">Pilih rekening tujuan transfer</option>
                <option v-for="bank in tempoBankAccounts" :key="bank.id" :value="bank.id">
                  {{ bank.bank_name }} • {{ bank.account_number }} • {{ bank.account_holder }}
                </option>
              </select>
              <p v-if="tempoBankAccounts.length === 0" class="text-xs text-amber-700 mt-1">Belum ada rekening aktif di menu Rekening Bank.</p>
            </div>

            <!-- Upload Bukti Transfer jika non-tunai -->
            <div v-if="tempoPaymentMethod === 'TRANSFER' || tempoPaymentMethod === 'QRIS'" class="space-y-2">
              <label class="text-xs font-black text-slate-700 uppercase tracking-widest flex items-center justify-between">
                <span>Foto Bukti Transfer</span>
                <span v-if="tempoTransferProofURL" class="text-emerald-600 text-[10px] font-bold">✓ Terlampir</span>
              </label>
              <div class="flex items-center gap-3">
                <input
                  type="file"
                  @change="handleTempoProofUpload"
                  accept="image/*"
                  class="hidden"
                  id="tempo-proof-upload"
                />
                <label
                  for="tempo-proof-upload"
                  class="flex-1 flex items-center justify-center gap-2 bg-indigo-50/50 border-2 border-dashed border-indigo-200 rounded-xl py-3 cursor-pointer hover:border-indigo-400 hover:bg-indigo-50 transition-all text-indigo-700"
                >
                  <Upload v-if="isUploadingTempoProof" class="w-4 h-4 animate-spin text-indigo-600" />
                  <ImageIcon v-else class="w-4 h-4 text-indigo-500" />
                  <span class="text-xs font-bold">
                    {{ isUploadingTempoProof ? 'Mengunggah...' : (tempoTransferProofURL ? 'Ganti Foto Bukti' : 'Pilih / Unggah Bukti Transfer') }}
                  </span>
                </label>
                <button
                  v-if="tempoTransferProofURL"
                  @click="openImagePreview(tempoTransferProofURL)"
                  type="button"
                  class="px-3 py-3 bg-indigo-100 text-indigo-800 rounded-xl text-xs font-bold hover:bg-indigo-200 transition-all"
                  title="Pratinjau Bukti"
                >
                  <Eye class="w-4 h-4" />
                </button>
              </div>
            </div>

            <!-- Jumlah Dibayar & Quick Fills -->
            <div>
              <div class="flex flex-wrap justify-between items-center gap-1.5 mb-1.5">
                <label class="text-xs font-black text-slate-700 uppercase tracking-widest"
                  >Jumlah Pembayaran (Rp) *</label
                >
                <div class="flex items-center gap-2">
                  <button
                    v-if="getTempoProgressInfo(selectedNota).targetTenor > 1 && getTempoProgressInfo(selectedNota).estimatedTermAmount > 0"
                    type="button"
                    @click="tempoPaymentAmount = getTempoProgressInfo(selectedNota).estimatedTermAmount"
                    class="text-[10px] font-bold px-2 py-0.5 bg-indigo-50 text-indigo-700 rounded-md border border-indigo-200 hover:bg-indigo-100"
                  >
                    1 Termin ({{ formatCurrency(getTempoProgressInfo(selectedNota).estimatedTermAmount) }})
                  </button>
                  <button
                    type="button"
                    @click="tempoPaymentAmount = getTempoProgressInfo(selectedNota).remainingAmount"
                    class="text-[10px] font-bold px-2 py-0.5 bg-emerald-50 text-emerald-700 rounded-md border border-emerald-200 hover:bg-emerald-100"
                  >
                    Lunas Semua ({{ formatCurrency(getTempoProgressInfo(selectedNota).remainingAmount) }})
                  </button>
                </div>
              </div>
              <input
                type="number"
                v-model="tempoPaymentAmount"
                class="w-full rounded-xl border-slate-200 bg-slate-50 px-4 py-3 font-black text-base focus:border-indigo-500 focus:ring-indigo-500 transition-colors"
                placeholder="Masukkan nominal..."
              />
            </div>

            <!-- Jatuh Tempo Berikutnya jika belum lunas -->
            <div v-if="Number(tempoPaymentAmount) < getTempoProgressInfo(selectedNota).remainingAmount" class="space-y-1.5 bg-amber-50/60 p-3.5 rounded-xl border border-amber-200/80">
              <div class="flex justify-between items-center">
                <label class="text-[11px] font-black text-amber-900 uppercase tracking-wider">Jatuh Tempo Cicilan Berikutnya</label>
                <span class="text-[10px] font-black text-amber-800">
                  Sisa Tagihan: {{ formatCurrency(Math.max(0, getTempoProgressInfo(selectedNota).remainingAmount - Number(tempoPaymentAmount))) }}
                </span>
              </div>
              <input
                type="date"
                v-model="tempoNextDueDate"
                class="w-full rounded-xl border-amber-300 bg-white px-3 py-2 text-xs font-bold text-slate-800 focus:border-amber-500 focus:ring-amber-500 outline-none"
              />
            </div>

            <!-- Catatan Pembayaran -->
            <div>
              <label class="text-xs font-black text-slate-700 uppercase tracking-widest"
                >Catatan Pembayaran (Opsional)</label
              >
              <input
                type="text"
                v-model="tempoNotes"
                placeholder="Contoh: Titip ke driver / Transfer via BCA"
                class="w-full mt-1.5 rounded-xl border-slate-200 bg-slate-50 px-4 py-2.5 font-medium text-xs focus:border-indigo-500 focus:ring-indigo-500 transition-colors"
              />
            </div>
          </div>

          <div class="flex gap-3 pt-2">
            <Button
              variant="ghost"
              class="flex-1 font-bold rounded-2xl py-6"
              @click="showTempoPaymentModal = false"
            >
              BATAL
            </Button>
            <Button
              class="flex-1 bg-indigo-600 hover:bg-indigo-700 text-white font-black rounded-2xl py-6 shadow-xl shadow-indigo-200"
              :disabled="isSubmittingPayment"
              @click="submitTempoPayment"
            >
              <span v-if="isSubmittingPayment">Memproses...</span>
              <span v-else>SIMPAN PEMBAYARAN</span>
            </Button>
          </div>
        </div>
      </Card>
    </div>

    <!-- Safety Soft Delete Modal -->
    <div v-if="showDeleteConfirmModal && deleteTargetSale" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm" @click="showDeleteConfirmModal = false" />
      <Card class="relative w-full max-w-lg bg-white rounded-3xl shadow-2xl border-none overflow-hidden z-10 animate-in fade-in zoom-in-95 duration-200">
        <div class="p-6 space-y-5">
          <div class="flex items-center gap-3 text-rose-600">
            <div class="p-3 bg-rose-50 rounded-2xl">
              <AlertTriangle class="w-6 h-6 text-rose-600" />
            </div>
            <div>
              <h3 class="font-extrabold text-lg text-slate-900 leading-tight">Hapus Faktur Penjualan</h3>
              <p class="text-xs text-rose-600 font-semibold mt-0.5">Penghapusan Aman (Soft Delete - Data dapat dipulihkan)</p>
            </div>
          </div>

          <div class="bg-slate-50 p-4 rounded-2xl border border-slate-200 space-y-2 text-xs">
            <div class="flex justify-between">
              <span class="text-slate-500 font-medium">No. Faktur / Nota:</span>
              <span class="font-mono font-bold text-slate-900">{{ deleteTargetSale.receiptNo }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-slate-500 font-medium">Toko:</span>
              <span class="font-bold text-slate-900">{{ deleteTargetSale.store }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-slate-500 font-medium">Salesman:</span>
              <span class="font-bold text-slate-900">{{ deleteTargetSale.salesman }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-slate-500 font-medium">Total Nominal:</span>
              <span class="font-black text-slate-900">{{ formatCurrency(deleteTargetSale.amount) }}</span>
            </div>
          </div>

          <div class="space-y-3">
            <div>
              <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
                Alasan Penghapusan <span class="text-rose-500">*</span>
              </label>
              <textarea
                v-model="deleteReason"
                placeholder="Contoh: Salah input toko, nota dibatalkan customer, double input..."
                rows="2"
                class="w-full bg-slate-50 border border-slate-200 rounded-xl p-3 text-xs font-medium text-slate-900 placeholder:text-slate-400 focus:bg-white focus:ring-2 focus:ring-rose-500/20 focus:border-rose-500 outline-none transition-all"
              ></textarea>
            </div>

            <div>
              <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
                Ketik <span class="font-mono text-rose-600 bg-rose-50 px-1.5 py-0.5 rounded">{{ deleteTargetSale.receiptNo }}</span> untuk konfirmasi <span class="text-rose-500">*</span>
              </label>
              <input
                v-model="deleteConfirmText"
                type="text"
                :placeholder="deleteTargetSale.receiptNo"
                class="w-full bg-slate-50 border border-slate-200 rounded-xl p-3 text-xs font-mono font-bold text-slate-900 placeholder:text-slate-300 focus:bg-white focus:ring-2 focus:ring-rose-500/20 focus:border-rose-500 outline-none transition-all"
              />
            </div>
          </div>

          <div class="flex gap-3 pt-2">
            <Button
              variant="outline"
              class="flex-1 rounded-xl py-5 font-bold text-xs"
              @click="showDeleteConfirmModal = false"
            >
              Batal
            </Button>
            <Button
              class="flex-1 bg-rose-600 hover:bg-rose-700 text-white font-bold rounded-xl py-5 text-xs shadow-lg shadow-rose-600/20 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
              :disabled="isDeletingTrx || !deleteReason.trim() || deleteConfirmText.trim() !== deleteTargetSale.receiptNo"
              @click="confirmDelete"
            >
              <RefreshCw v-if="isDeletingTrx" class="w-4 h-4 animate-spin mr-1.5" />
              <span>{{ isDeletingTrx ? 'Menghapus...' : 'Ya, Hapus Faktur' }}</span>
            </Button>
          </div>
        </div>
      </Card>
    </div>

    <!-- Restore Confirm Modal -->
    <ConfirmationDialog
      v-model:open="showRestoreConfirmModal"
      title="Pulihkan Faktur Penjualan"
      :description="`Apakah Anda yakin ingin memulihkan faktur ${restoreTargetSale?.receiptNo || ''} (${restoreTargetSale?.store || ''}) kembali ke daftar aktif? Perhitungan KPI/omzet akan disesuaikan kembali.`"
      variant="info"
      confirmText="Ya, Pulihkan Faktur"
      cancelText="Batal"
      :isLoading="isRestoringTrx"
      @confirm="confirmRestore"
      @cancel="showRestoreConfirmModal = false"
    />
    <!-- Dynamic Print Styles for specific Modals -->
    <component :is="'style'" v-if="selectedDO">
      @media print {
        @page { size: landscape; margin: 0; }
      }
    </component>
    <component :is="'style'" v-if="selectedNota">
      @media print {
        @page { size: 21cm 14cm; margin: 0; }
      }
    </component>
    </div>

  <div v-if="showCancelDeliveryModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <div
      class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
      @click="showCancelDeliveryModal = false"
    ></div>
    <div
      class="relative bg-white dark:bg-slate-900 rounded-3xl w-full max-w-md shadow-2xl p-6 sm:p-8 transform transition-all"
    >
      <div class="flex flex-col items-center text-center space-y-4">
        <div class="w-16 h-16 bg-rose-100 dark:bg-rose-900/30 rounded-2xl flex items-center justify-center mb-2">
          <Ban class="w-8 h-8 text-rose-600 dark:text-rose-400" />
        </div>
        
        <h3 class="text-xl font-black text-slate-900 dark:text-white">
          Gagal Kirim — Jadwalkan Ulang
        </h3>
        
        <div class="bg-slate-50 dark:bg-slate-800/50 rounded-xl p-4 w-full border border-slate-100 dark:border-slate-800 text-left">
          <p class="text-xs text-slate-500 font-medium mb-1">Transaksi / Nomor Nota</p>
          <p class="text-sm font-bold text-slate-900 dark:text-white">{{ cancelDeliveryTransaction?.receiptNo }}</p>
          
          <p class="text-xs text-slate-500 font-medium mt-3 mb-1">Customer / Toko</p>
          <p class="text-sm font-bold text-slate-900 dark:text-white">{{ cancelDeliveryTransaction?.store }}</p>
        </div>
        
        <div class="w-full text-left">
          <label class="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1.5">
            Alasan Gagal Kirim <span class="text-rose-500">*</span>
          </label>
          <textarea
            v-model="cancelDeliveryReason"
            rows="3"
            class="w-full bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl p-3 text-xs focus:ring-2 focus:ring-rose-500/20 focus:border-rose-500 outline-none transition-all placeholder:text-slate-400 text-slate-900 dark:text-white"
            placeholder="Contoh: Toko tutup / penerima tidak ada..."
          ></textarea>
        </div>

        <p class="text-xs text-slate-500 dark:text-slate-400 leading-relaxed font-medium text-left bg-rose-50/60 dark:bg-rose-950/20 p-3 rounded-xl border border-rose-100 dark:border-rose-900/50">
          Pilih "Kembalikan ke Pengiriman" untuk membuat surat jalan baru. Jika pesanan tidak jadi dipenuhi, pilih "Batalkan Faktur"; proses pengembalian barang ke gudang tetap berjalan.
        </p>
      </div>

      <div class="flex flex-wrap items-center gap-3 mt-6 w-full">
        <Button
          @click="showCancelDeliveryModal = false"
          variant="outline"
          class="flex-1 rounded-xl h-12 font-bold"
        >
          Tutup
        </Button>
        <Button
          v-if="canCancelFailedDelivery"
          @click="confirmFinalCancelDelivery"
          :disabled="cancelDeliveryLoading || !cancelDeliveryReason.trim()"
          variant="outline"
          class="flex-1 rounded-xl h-12 font-bold text-rose-700 border-rose-300"
        >
          Batalkan Faktur
        </Button>
        <Button
          @click="confirmCancelDelivery"
          :disabled="cancelDeliveryLoading || !cancelDeliveryReason.trim()"
          variant="destructive"
          class="flex-1 rounded-xl h-12 font-bold bg-rose-600 hover:bg-rose-700 text-white"
        >
          {{ cancelDeliveryLoading ? 'Memproses...' : 'Kembalikan ke Pengiriman' }}
        </Button>
      </div>
    </div>
  </div>

  <OrderLocationMapModal 
    :open="isMapModalOpen" 
    @update:open="isMapModalOpen = $event" 
    :order-data="selectedOrderForMap" 
  />
  <DirectSaleModal 
    :show="showDirectSaleModal" 
    @close="showDirectSaleModal = false" 
    @created="handleDirectSaleCreated" 
  />
  <DirectSaleInvoicePrintModal 
    :show="showDirectSalePrintModal" 
    :sale="selectedDirectSale" 
    @close="showDirectSalePrintModal = false" 
  />
  <DirectSaleReturnModal
    :show="showDirectSaleReturnModal"
    :sale="selectedReturnSale"
    @close="showDirectSaleReturnModal = false"
    @created="handleDirectSaleReturnCreated"
  />
  <BulkInvoicePrintModal
    :show="showBulkInvoicePrintModal"
    :sales="selectedSalesList"
    @close="showBulkInvoicePrintModal = false"
  />
  <BranchTransferInvoiceModal
    :show="showBranchInvoiceModal"
    :edit-transfer="selectedEditBranchTransfer"
    @close="showBranchInvoiceModal = false; selectedEditBranchTransfer = null"
    @created="handleBranchTransferCreated"
    @updated="handleBranchTransferUpdated"
  />
  <BranchInvoicePrintModal
    :show="showBranchPrintModal"
    :transfer="selectedBranchTransfer"
    @close="showBranchPrintModal = false"
  />
  <BranchFactoryReturnModal
    :show="showBranchReturnModal"
    :initialTransfer="selectedReturnTransfer"
    @close="showBranchReturnModal = false"
    @created="handleBranchReturnCreated"
  />

  <!-- Modal Impor Data Offline -->
  <div
    v-if="showOfflineImportModal"
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-sm animate-in fade-in duration-200"
  >
    <div
      class="bg-white rounded-3xl max-w-xl w-full p-6 sm:p-8 shadow-2xl border border-slate-100 flex flex-col gap-6"
    >
      <div class="flex items-center justify-between border-b border-slate-100 pb-4">
        <div class="flex items-center gap-3">
          <div class="p-3 bg-amber-50 rounded-2xl text-amber-600">
            <Upload class="w-6 h-6" />
          </div>
          <div>
            <h3 class="text-lg font-black text-slate-800">Impor Transaksi Offline</h3>
            <p class="text-xs font-medium text-slate-500">
              Tempel teks salinan cadangan dari WhatsApp atau unggah file .json
            </p>
          </div>
        </div>
        <button
          @click="showOfflineImportModal = false"
          class="p-2 rounded-xl hover:bg-slate-100 text-slate-400 hover:text-slate-600 transition"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <div class="space-y-4">
        <div>
          <div class="flex items-center justify-between mb-2">
            <label class="text-xs font-bold text-slate-700 uppercase tracking-wider">
              Teks Data Cadangan (JSON)
            </label>
            <button
              type="button"
              @click="handlePasteClipboard"
              class="text-xs font-bold text-primary hover:underline flex items-center gap-1"
            >
              <ClipboardList class="w-3.5 h-3.5" /> Tempel dari Clipboard
            </button>
          </div>
          <textarea
            v-model="offlineJsonText"
            rows="6"
            placeholder='Tempel teks JSON di sini, contoh: {"exported_at":"...", "transactions":[...]}'
            class="w-full p-3.5 text-xs font-mono bg-slate-50 border border-slate-200 rounded-2xl focus:bg-white focus:border-amber-500 focus:ring-2 focus:ring-amber-500/20 outline-none transition resize-y"
          ></textarea>
        </div>

        <div class="flex items-center gap-3 py-1">
          <div class="h-[1px] flex-1 bg-slate-100"></div>
          <span class="text-[11px] font-bold text-slate-400 uppercase">ATAU</span>
          <div class="h-[1px] flex-1 bg-slate-100"></div>
        </div>

        <div>
          <Button
            type="button"
            variant="outline"
            @click="offlineImportInput?.click()"
            :disabled="isImportingOffline"
            class="w-full py-5 border-dashed border-2 border-slate-200 hover:border-amber-500 hover:bg-amber-50/50 rounded-2xl flex items-center justify-center gap-2 text-xs font-bold text-slate-600"
          >
            <Upload class="w-4 h-4 text-amber-500" />
            Pilih File Cadangan (.json) dari Komputer
          </Button>
        </div>
      </div>

      <div class="flex items-center justify-end gap-3 pt-4 border-t border-slate-100">
        <Button
          type="button"
          variant="outline"
          @click="showOfflineImportModal = false"
          :disabled="isImportingOffline"
          class="rounded-xl px-5 py-2 text-xs font-bold text-slate-600"
        >
          Batal
        </Button>
        <Button
          type="button"
          @click="processImportOfflinePayload(offlineJsonText)"
          :disabled="isImportingOffline || !offlineJsonText.trim()"
          class="bg-amber-500 hover:bg-amber-600 text-white rounded-xl px-6 py-2 text-xs font-bold shadow-lg shadow-amber-500/20 flex items-center gap-2"
        >
          <Upload class="w-4 h-4" />
          {{ isImportingOffline ? 'Memproses Impor...' : 'Proses Impor' }}
        </Button>
      </div>
    </div>
  </div>
  </div>
</template>


<style scoped>
@media print {
  @page {
    margin: 0;
  }
  html,
  body {
    height: auto !important;
    margin: 0 !important;
    padding: 0 !important;
    overflow: visible !important;
    background: white !important;
  }
  .fixed.inset-0,
  .fixed.inset-0 *,
  #surat-jalan-card,
  #surat-jalan-card *,
  #nota-penjualan-card,
  #nota-penjualan-card *,
  #bulk-invoices-container,
  #bulk-invoices-container *,
  .bulk-invoice-sheet,
  .bulk-invoice-sheet * {
    visibility: visible;
  }
  .fixed.inset-0 {
    position: absolute !important;
    top: 0 !important;
    left: 0 !important;
    width: 100% !important;
    height: auto !important;
    background: white !important;
    padding: 0 !important;
    overflow: visible !important;
    display: block !important;
    z-index: 9999 !important;
  }
  #surat-jalan-card,
  #nota-penjualan-card {
    position: static !important;
    margin: 0 auto !important;
    width: 100% !important;
    border: none !important;
    box-shadow: none !important;
    transform: none !important;
    page-break-after: avoid !important;
    break-after: avoid !important;
  }
  #surat-jalan-card {
    zoom: 0.8;
    -moz-transform: scale(0.8);
    -moz-transform-origin: top left;
    transform: scale(0.8);
    transform-origin: top left;
  }
  #nota-penjualan-card {
    width: 21cm !important;
    height: 14cm !important;
    padding: 8px 15px !important;
    box-sizing: border-box !important;
    overflow: hidden !important;
    display: flex !important;
    flex-direction: column !important;
    justify-content: space-between !important;
  }
  #surat-jalan-card .p-4,
  #nota-penjualan-card .p-8 {
    padding: 0 !important;
  }
  #surat-jalan-card table,
  #nota-penjualan-card table {
    font-size: 7.5pt !important;
    border-collapse: collapse !important;
    width: 100% !important;
    line-height: 1 !important;
  }
  #surat-jalan-card th,
  #surat-jalan-card td,
  #nota-penjualan-card th,
  #nota-penjualan-card td {
    padding: 1.5px !important;
    border: 1px solid #000 !important;
    visibility: visible !important;
  }
  .print\:hidden {
    display: none !important;
  }
  #surat-jalan-card div,
  #nota-penjualan-card div {
    max-height: none !important;
    overflow: visible !important;
  }
}
</style>
