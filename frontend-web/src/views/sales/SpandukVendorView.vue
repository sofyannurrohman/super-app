<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import {
  ShoppingCart,
  Printer,
  CheckCircle2,
  Clock,
  Truck,
  X,
  RefreshCw,
  AlertTriangle,
  Building,
  Tag,
  Search,
  ChevronLeft,
  ChevronRight,
  Eye,
  Sparkles,
  Ruler,
  Mail,
  MessageSquare,
  ExternalLink,
  DollarSign,
  Layers,
  ListChecks,
  Check,
  Copy,
  Receipt,
  RotateCcw,
  Calendar,
  Pencil,
  Calculator,
  Send,
  Server,
  SendHorizontal,
  MailCheck,
  MailX,
  Building2,
  Info,
  ArrowDown,
  ArrowUp,
  ArrowUpDown,
} from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import { useMasterDataStore } from '@/stores/masterData'
import { useAuthStore } from '@/stores/auth'
import StoreAddressHoverCard from '@/components/StoreAddressHoverCard.vue'
import {
  calculateBannerDesignSize,
  formatBannerDimension,
} from '@/utils/bannerSize'

const getSizeSuggestion = (task: any) => {
  if (!task) return null
  return calculateBannerDesignSize(task.size_length, task.size_width)
}

const masterStore = useMasterDataStore()
const authStore = useAuthStore()

const isPurchasingAdmin = computed(() => {
  return authStore.user?.role === 'admin_purchasing'
})

const isSuperAdmin = computed(() => {
  return authStore.user && ['superadmin', 'admin', 'hr_admin', 'HR', 'manager'].includes(authStore.user.role)
})

const canManagePurchasing = computed(() => {
  return isPurchasingAdmin.value || isSuperAdmin.value
})

// Summary metrics computed from current list
const totalCostCurrentPage = computed(() => {
  return purchasingTasks.value.reduce((sum, item) => sum + (Number(item.total_price) || 0), 0)
})

const totalAreaCurrentPage = computed(() => {
  return purchasingTasks.value.reduce((sum, item) => {
    const area = getBannerAreaM2(item) || 0
    return sum + area
  }, 0).toFixed(1)
})

const isLoading = ref(false)
const isOnline = ref(true)

const purchasingTasks = ref<any[]>([])
// Tabs: 'purchasing_queue' | 'sent_to_vendor' | 'unpaid_bills' | 'paid_history'
const activeTab = ref('purchasing_queue')

// Master Vendors list for dropdown
const vendorOptions = ref<any[]>([])
const isLoadingVendors = ref(false)

// Filters
const selectedBranchFilter = computed<string>({
  get: () => masterStore.selectedBrandingBranchId || 'ALL_BRANCHES',
  set: (val: string) => {
    masterStore.setSelectedBrandingBranchId(val)
  },
})
const selectedFilterVendorId = ref('all')
const selectedFilterMonth = ref('all')

const selectedBranchName = computed(() => {
  return masterStore.selectedBrandingBranchName
})

// Pagination & Search
const page = ref(1)
const limit = ref(10)
const total = ref(0)
const search = ref('')

// Sorting state (default: DESC / paling baru)
const dateSortDir = ref<'DESC' | 'ASC'>('DESC')

const toggleDateSort = () => {
  dateSortDir.value = dateSortDir.value === 'DESC' ? 'ASC' : 'DESC'
  page.value = 1
  clearSelection()
  fetchTasks()
}

// Selection state for Bulk Operations
const selectedBannerIds = ref<string[]>([])

const isAllSelected = computed(() => {
  if (purchasingTasks.value.length === 0) return false
  return purchasingTasks.value.every((t) => selectedBannerIds.value.includes(t.id))
})

const isSomeSelected = computed(() => {
  return (
    selectedBannerIds.value.length > 0 &&
    !isAllSelected.value &&
    purchasingTasks.value.some((t) => selectedBannerIds.value.includes(t.id))
  )
})

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    const currentPageIds = purchasingTasks.value.map((t) => t.id)
    selectedBannerIds.value = selectedBannerIds.value.filter((id) => !currentPageIds.includes(id))
  } else {
    const newSelected = new Set([...selectedBannerIds.value, ...purchasingTasks.value.map((t) => t.id)])
    selectedBannerIds.value = Array.from(newSelected)
  }
}

const toggleSelectItem = (id: string) => {
  const index = selectedBannerIds.value.indexOf(id)
  if (index > -1) {
    selectedBannerIds.value.splice(index, 1)
  } else {
    selectedBannerIds.value.push(id)
  }
}

const clearSelection = () => {
  selectedBannerIds.value = []
}

// Single// Vendor Assignment Modal (Tugaskan Satuan ke Vendor)
const showVendorModal = ref(false)
const selectedTask = ref<any>(null)
const selectedVendorId = ref('')
const vendorName = ref('')
const vendorEmail = ref('')
const vendorPhone = ref('')
const vendorPrice = ref<number | ''>('')
const autoOpenGmail = ref(true)
const isSubmitting = ref(false)

// Modal Make Sure / Konfirmasi Kirim Email via Server SMTP
const showSmtpConfirmModal = ref(false)
const smtpModalTask = ref<any>(null)
const smtpRecipientEmail = ref('')
const smtpVendorName = ref('')
const smtpSubject = ref('')
const smtpCustomNotes = ref('')
const isSendingSmtp = ref(false)

// Edit Price & Vendor Modal
const showEditPriceModal = ref(false)
const editPriceTask = ref<any>(null)
const editPriceVendorId = ref('')
const editPriceVendorName = ref('')
const editPriceVendorEmail = ref('')
const editPriceVendorPhone = ref('')
const editPriceValue = ref<number | ''>('')
const isEditPriceSubmitting = ref(false)

// Bulk Vendor Modal (Kirim Bulk ke Vendor)
const showBulkModal = ref(false)
const bulkSelectedVendorId = ref('')
const bulkVendorName = ref('')
const bulkVendorEmail = ref('')
const bulkVendorPhone = ref('')
const bulkSendViaSmtp = ref(true)
const bulkCustomNotes = ref('')
const isBulkSubmitting = ref(false)
const bulkItems = ref<
  {
    id: string
    shop_name: string
    shop_address: string
    type: string
    raw_type: string
    qty: number
    size_length: number
    size_width: number
    area: number
    price: number
    design_url: string
    notes: string
    is_vendor_email_sent?: boolean
    vendor_email_sent_at?: string | null
  }[]
>([])

// Bulk Email Modal (Kirim Email Desain Massal ke Vendor via SMTP / Gmail)
const showBulkEmailModal = ref(false)
const bulkEmailSelectedVendorId = ref('')
const bulkEmailVendorName = ref('')
const bulkEmailRecipient = ref('')
const bulkEmailSubject = ref('')
const bulkEmailCustomNotes = ref('')
const isSendingBulkEmail = ref(false)
const bulkEmailTasks = ref<any[]>([])

// Payment Modal (Lunasi Tagihan Satuan / Bulk Akhir Bulan)
const showPaymentModal = ref(false)
const paymentModalTasks = ref<any[]>([])
const paymentReference = ref('')
const paymentNotes = ref('')
const isPaymentSubmitting = ref(false)

const paymentModalTotalCost = computed(() => {
  return paymentModalTasks.value.reduce((sum, item) => sum + (Number(item.total_price) || 0), 0)
})

const fetchVendorsList = async () => {
  isLoadingVendors.value = true
  try {
    const res = await apiClient.get('/branding/vendors?is_active=true')
    if (res.data?.data) {
      vendorOptions.value = res.data.data
    } else {
      vendorOptions.value = []
    }
  } catch (error) {
    console.error('Gagal mengambil master vendor:', error)
  } finally {
    isLoadingVendors.value = false
  }
}

const fetchTasks = async (silent = false) => {
  if (!silent) isLoading.value = true
  try {
    let statusQuery: string | undefined = undefined
    let paymentStatusQuery: string | undefined = undefined
    let sortByQuery: string | undefined = undefined
    let sortDirQuery: string | undefined = dateSortDir.value

    if (activeTab.value === 'purchasing_queue') {
      statusQuery = 'purchasing_queue'
    } else if (activeTab.value === 'sent_to_vendor') {
      statusQuery = 'sent_to_vendor'
      sortByQuery = 'updated_at'
    } else if (activeTab.value === 'unpaid_bills') {
      statusQuery = 'sent_to_vendor,ready_for_operator,diterima_operator,installed,terpasang'
      paymentStatusQuery = 'unpaid'
      sortByQuery = 'updated_at'
    } else if (activeTab.value === 'paid_history') {
      statusQuery = 'sent_to_vendor,ready_for_operator,diterima_operator,installed,terpasang'
      paymentStatusQuery = 'paid'
      sortByQuery = 'paid_at'
    }

    const branchId = selectedBranchFilter.value
    const res = await apiClient.get(`/branding/banners`, {
      params: {
        status: statusQuery,
        payment_status: paymentStatusQuery,
        sort_by: sortByQuery,
        sort_dir: sortDirQuery,
        vendor_id: selectedFilterVendorId.value && selectedFilterVendorId.value !== 'all' ? selectedFilterVendorId.value : undefined,
        month: selectedFilterMonth.value && selectedFilterMonth.value !== 'all' ? selectedFilterMonth.value : undefined,
        page: page.value,
        limit: limit.value,
        search: search.value,
        branch_id: branchId && branchId !== 'ALL_BRANCHES' ? branchId : undefined,
      },
    })

    const responseData = res.data

    if (responseData && responseData.data) {
      purchasingTasks.value = responseData.data.map((b: any) => ({
        id: b.id,
        shop_name: b.shop_name || 'Toko Unknown',
        shop_address: b.shop_address || '',
        shop_phone: b.shop_phone || '',
        province: b.province || '',
        regency: b.regency || '',
        district: b.district || '',
        village: b.village || '',
        latitude: b.latitude || 0,
        longitude: b.longitude || 0,
        sales_name: b.sales_name || '-',
        sales_phone: b.sales_phone || '',
        sales_branch: b.sales_branch || '',
        type: `${b.type || 'Spanduk'} ${formatBannerDimension(b.size_length, b.size_width, 'dual')}`,
        raw_type: b.type || 'Spanduk',
        qty: b.qty || 1,
        size_length: b.size_length || 0,
        size_width: b.size_width || 0,
        status: b.status?.toLowerCase(),
        notes: b.notes || '-',
        design_url: b.design_link || b.design_url || '',
        documentation_image_url: b.documentation_image_url || '',
        vendor_id: b.vendor_id,
        vendor_name: b.vendor_name || 'Belum ditentukan',
        total_price: b.total_price || 0,
        is_paid: Boolean(b.is_paid),
        payment_status: b.payment_status || (b.is_paid ? 'paid' : 'unpaid'),
        paid_at: b.paid_at ? formatDate(b.paid_at) : null,
        payment_reference: b.payment_reference || '',
        payment_notes: b.payment_notes || '',
        validator_name: b.validator_name || '',
        validator_position: b.validator_position || '',
        design_validator_name: b.design_validator_name || b.validator_name || '',
        design_validator_position: b.design_validator_position || b.validator_position || '',
        store_validator_name: b.store_validator_name || b.validator_name || '',
        store_validator_position: b.store_validator_position || b.validator_position || '',
        purchasing_name: b.purchasing_name || '',
        purchasing_position: b.purchasing_position || '',
        is_vendor_email_sent: Boolean(b.is_vendor_email_sent),
        vendor_email_sent_at: b.vendor_email_sent_at ? formatDate(b.vendor_email_sent_at) : null,
        updated_at: formatDate(b.updated_at),
      }))
      total.value = responseData.total || 0
      page.value = responseData.page || 1
      limit.value = responseData.limit || 10
    } else {
      purchasingTasks.value = []
      total.value = 0
    }

    isOnline.value = true
  } catch (e: any) {
    isOnline.value = false
    toast.error('Gagal memuat data dari server.')
  } finally {
    if (!silent) isLoading.value = false
  }
}

let searchTimeout: any = null
watch(search, () => {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    page.value = 1
    clearSelection()
    fetchTasks(true)
  }, 500)
})

watch(activeTab, () => {
  page.value = 1
  search.value = ''
  dateSortDir.value = 'DESC'
  clearSelection()
  fetchTasks()
})

watch(
  [selectedFilterVendorId, selectedFilterMonth],
  () => {
    page.value = 1
    clearSelection()
    fetchTasks()
  },
)

const prevPage = () => {
  if (page.value > 1) {
    page.value--
    clearSelection()
    fetchTasks()
  }
}

const nextPage = () => {
  const maxPage = Math.ceil(total.value / limit.value)
  if (page.value < maxPage) {
    page.value++
    clearSelection()
    fetchTasks()
  }
}

watch(() => masterStore.selectedBrandingBranchId, () => {
  page.value = 1
  clearSelection()
  fetchTasks()
})

onMounted(async () => {
  if (!masterStore.branches || masterStore.branches.length === 0) {
    await masterStore.fetchBranches()
  }
  fetchTasks()
  fetchVendorsList()
})

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  try {
    return new Date(dateStr).toLocaleString('id-ID', {
      day: '2-digit',
      month: 'short',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  } catch {
    return dateStr
  }
}

const formatCurrency = (amount: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(amount)
}

const getBannerAreaM2 = (task: any) => {
  if (!task || !task.size_length || !task.size_width) return null
  const area = (task.size_length / 100) * (task.size_width / 100)
  return Number(area.toFixed(2))
}

const getVendorByTask = (task: any) => {
  if (!task) return null
  if (task.vendor_id) {
    const byId = vendorOptions.value.find((v: any) => String(v.id) === String(task.vendor_id))
    if (byId) return byId
  }
  if (task.vendor_name && task.vendor_name !== 'Belum ditentukan') {
    return vendorOptions.value.find(
      (v: any) => v.name.trim().toLowerCase() === task.vendor_name.trim().toLowerCase(),
    )
  }
  return null
}

const openVendorModal = (task: any) => {
  selectedTask.value = task
  selectedVendorId.value = task.vendor_id || ''
  vendorName.value = task.vendor_name && task.vendor_name !== 'Belum ditentukan' ? task.vendor_name : ''
  vendorEmail.value = ''
  vendorPhone.value = ''
  vendorPrice.value = task.total_price || ''

  // Look up vendor by ID or by Name fallback
  const matchedVendor = getVendorByTask(task)
  if (matchedVendor) {
    selectedVendorId.value = matchedVendor.id
    vendorName.value = matchedVendor.name
    vendorEmail.value = matchedVendor.email || ''
    vendorPhone.value = matchedVendor.phone || ''
    if (!vendorPrice.value && matchedVendor.price_per_meter > 0 && task.size_length && task.size_width) {
      const areaInM2 = (task.size_length / 100) * (task.size_width / 100)
      vendorPrice.value = Math.round(areaInM2 * matchedVendor.price_per_meter)
    }
  }

  showVendorModal.value = true
}

const onVendorSelectChange = () => {
  if (!selectedVendorId.value) {
    vendorName.value = ''
    vendorEmail.value = ''
    vendorPhone.value = ''
    return
  }
  const v = vendorOptions.value.find((item: any) => String(item.id) === String(selectedVendorId.value))
  if (v) {
    vendorName.value = v.name
    vendorEmail.value = v.email || ''
    vendorPhone.value = v.phone || ''
    // Auto-calculate estimate if price_per_meter is set
    if (v.price_per_meter > 0 && selectedTask.value?.size_length && selectedTask.value?.size_width) {
      const areaInM2 = (selectedTask.value.size_length / 100) * (selectedTask.value.size_width / 100)
      vendorPrice.value = Math.round(areaInM2 * v.price_per_meter)
    }
  }
}

const generateVendorEmailSubject = (task: any) => {
  if (!task) return '[ORDER CETAK SPANDUK]'
  return `[ORDER CETAK SPANDUK] - ${task.shop_name} (${task.size_length}x${task.size_width} cm)`
}

const generateVendorEmailBody = (task: any, vName: string, price: number | string) => {
  if (!task) return ''
  const area = getBannerAreaM2(task) ? `${getBannerAreaM2(task)} m²` : '-'
  const formattedPrice = typeof price === 'number' || !isNaN(Number(price))
    ? formatCurrency(Number(price))
    : String(price)
  const design = task.design_url || 'Link desain menyusul'
  const notes = task.notes && task.notes !== '-' ? task.notes : 'Tidak ada'

  return `Halo Tim Percetakan ${vName || 'Vendor'},

Berikut adalah detail Surat Perintah Kerja (SPK) cetak spanduk/banner baru:

--------------------------------------------------
• Nama Toko         : ${task.shop_name}
• Alamat Toko       : ${task.shop_address || '-'}
• Jenis Banner      : ${task.raw_type || 'Spanduk'}
• Ukuran Banner     : ${formatBannerDimension(task.size_length, task.size_width, 'dual')}
• Luas Banner       : ${area}
• Jumlah (Qty)      : ${task.qty || 1} Pcs
• Total Biaya Cetak : ${formattedPrice}
• Catatan Tambahan  : ${notes}
--------------------------------------------------
• LINK FILE DESAIN (SIAP CETAK):
${design}
--------------------------------------------------

Mohon konfirmasi jika file desain sudah diterima dan proses cetak segera dimulai.
Terima kasih.`
}

const getGmailUrl = (task: any, vEmail: string, vName: string, price: number | string) => {
  const subject = generateVendorEmailSubject(task)
  const body = generateVendorEmailBody(task, vName, price)
  const targetEmail = vEmail ? vEmail.trim() : ''
  return `https://mail.google.com/mail/?view=cm&to=${encodeURIComponent(targetEmail)}&su=${encodeURIComponent(subject)}&body=${encodeURIComponent(body)}`
}

const launchGmail = (task: any, vEmail: string, vName: string, price: number | string) => {
  const gmailUrl = getGmailUrl(task, vEmail, vName, price)
  const win = window.open(gmailUrl, '_blank')
  if (!win || win.closed || typeof win.closed === 'undefined') {
    window.location.href = `mailto:${encodeURIComponent(vEmail ? vEmail.trim() : '')}?subject=${encodeURIComponent(generateVendorEmailSubject(task))}&body=${encodeURIComponent(generateVendorEmailBody(task, vName, price))}`
  }
}

const copySpkFormat = async (task: any, vName: string, price: number | string) => {
  const body = generateVendorEmailBody(task, vName, price)
  try {
    await navigator.clipboard.writeText(body)
    toast.success('Format SPK berhasil disalin ke clipboard!')
  } catch {
    toast.error('Gagal menyalin format SPK')
  }
}

const openSmtpConfirmModal = (task: any, overrideEmail?: string, overrideName?: string, overridePrice?: number | string) => {
  if (!task) return
  smtpModalTask.value = task
  const matchedVendor = getVendorByTask(task)
  smtpVendorName.value = overrideName || task.vendor_name || matchedVendor?.name || 'Vendor Percetakan'
  smtpRecipientEmail.value = overrideEmail || matchedVendor?.email || ''
  smtpSubject.value = generateVendorEmailSubject(task)
  smtpCustomNotes.value = ''
  showSmtpConfirmModal.value = true
}

const generateVendorEmailHtml = (task: any, vName: string, price: number | string, customNotes?: string) => {
  if (!task) return ''
  const area = getBannerAreaM2(task) ? `${getBannerAreaM2(task)} m²` : '-'
  const formattedPrice = typeof price === 'number' || !isNaN(Number(price))
    ? formatCurrency(Number(price))
    : String(price)
  const design = task.design_url || 'Link desain menyusul'
  const notes = task.notes && task.notes !== '-' ? task.notes : 'Tidak ada'

  const customNotesHtml = customNotes && customNotes.trim()
    ? `<div style="margin-top: 16px; padding: 12px; background-color: #f8fafc; border-left: 4px solid #3b82f6; border-radius: 4px;">
        <strong style="color: #1e293b; font-size: 13px;">Catatan Tambahan:</strong><br>
        <span style="color: #475569; font-size: 13px; white-space: pre-wrap;">${customNotes.trim()}</span>
       </div>`
    : ''

  return `
    <div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; max-width: 620px; margin: 0 auto; padding: 24px; background-color: #ffffff; border: 1px solid #e2e8f0; border-radius: 16px; color: #1e293b; line-height: 1.5;">
      <div style="border-bottom: 2px solid #2563eb; padding-bottom: 14px; margin-bottom: 20px;">
        <h2 style="margin: 0; color: #1e293b; font-size: 20px; font-weight: 800;">ORDER CETAK SPANDUK (SPK)</h2>
        <p style="margin: 4px 0 0 0; color: #64748b; font-size: 12px;">Sistem Operasional Percetakan KARSA HR</p>
      </div>

      <p style="font-size: 14px; margin-bottom: 16px;">
        Halo Tim Percetakan <strong>${vName || 'Vendor'}</strong>,
      </p>
      <p style="font-size: 14px; margin-bottom: 20px; color: #334155;">
        Berikut rincian Surat Perintah Kerja (SPK) cetak spanduk/banner baru yang siap diproses:
      </p>

      <div style="background-color: #f8fafc; border: 1px solid #e2e8f0; border-radius: 12px; padding: 16px; margin-bottom: 20px;">
        <table style="width: 100%; border-collapse: collapse; font-size: 13px;">
          <tr>
            <td style="padding: 6px 0; color: #64748b; width: 140px;">Nama Toko</td>
            <td style="padding: 6px 0; font-weight: 700; color: #0f172a;">: ${task.shop_name}</td>
          </tr>
          <tr>
            <td style="padding: 6px 0; color: #64748b;">Alamat Toko</td>
            <td style="padding: 6px 0; color: #334155;">: ${task.shop_address || '-'}</td>
          </tr>
          <tr>
            <td style="padding: 6px 0; color: #64748b;">Jenis Spanduk</td>
            <td style="padding: 6px 0; font-weight: 600; color: #0f172a;">: ${task.raw_type || task.type || 'Spanduk'}</td>
          </tr>
          <tr>
            <td style="padding: 6px 0; color: #64748b;">Ukuran Banner</td>
            <td style="padding: 6px 0; font-weight: 700; color: #2563eb;">: ${formatBannerDimension(task.size_length, task.size_width, 'dual')}</td>
          </tr>
          <tr>
            <td style="padding: 6px 0; color: #64748b;">Luas Banner</td>
            <td style="padding: 6px 0; color: #334155;">: ${area}</td>
          </tr>
          <tr>
            <td style="padding: 6px 0; color: #64748b;">Jumlah (Qty)</td>
            <td style="padding: 6px 0; font-weight: 700; color: #0f172a;">: ${task.qty || 1} Pcs</td>
          </tr>
          <tr>
            <td style="padding: 6px 0; color: #64748b;">Total Biaya Cetak</td>
            <td style="padding: 6px 0; font-weight: 800; color: #059669;">: ${formattedPrice}</td>
          </tr>
          <tr>
            <td style="padding: 6px 0; color: #64748b;">Catatan Khusus</td>
            <td style="padding: 6px 0; color: #334155;">: ${notes}</td>
          </tr>
        </table>
      </div>

      <div style="background-color: #eff6ff; border: 1px solid #bfdbfe; border-radius: 12px; padding: 16px; margin-bottom: 20px; text-align: center;">
        <p style="margin: 0 0 10px 0; font-size: 12px; font-weight: 700; color: #1e40af; text-transform: uppercase; letter-spacing: 0.5px;">
          Link File Desain Siap Cetak:
        </p>
        <a href="${design}" target="_blank" style="display: inline-block; background-color: #2563eb; color: #ffffff; text-decoration: none; padding: 10px 20px; border-radius: 8px; font-weight: 700; font-size: 13px;">
          Buka & Unduh Desain GDrive
        </a>
        <p style="margin: 8px 0 0 0; font-size: 11px; color: #60a5fa; word-break: break-all;">
          ${design}
        </p>
      </div>

      ${customNotesHtml}

      <div style="margin-top: 24px; padding-top: 16px; border-top: 1px solid #e2e8f0; font-size: 12px; color: #64748b; text-align: center;">
        <p style="margin: 0 0 4px 0;">Mohon konfirmasi jika file desain sudah diterima dan proses cetak segera dimulai.</p>
        <p style="margin: 0; font-weight: 600;">Terima kasih atas kerjasamanya.</p>
      </div>
    </div>
  `
}

const sendSmtpEmail = async () => {
  if (!smtpModalTask.value) return
  if (!smtpRecipientEmail.value || !smtpRecipientEmail.value.includes('@')) {
    toast.error('Masukkan alamat email vendor penerima yang valid')
    return
  }

  isSendingSmtp.value = true
  try {
    const task = smtpModalTask.value
    const vName = smtpVendorName.value
    const price = task.total_price || vendorPrice.value || 0
    const htmlBody = generateVendorEmailHtml(task, vName, price, smtpCustomNotes.value)

    const res = await apiClient.post('/branding/banners/send-vendor-email', {
      banner_id: task.id,
      vendor_id: task.vendor_id || selectedVendorId.value || undefined,
      vendor_name: vName,
      vendor_email: smtpRecipientEmail.value.trim(),
      subject: smtpSubject.value.trim(),
      custom_notes: smtpCustomNotes.value.trim(),
      html_body: htmlBody,
    })

    toast.success(res.data?.message || `Email SPK berhasil dikirim ke ${smtpRecipientEmail.value}`)
    showSmtpConfirmModal.value = false
    fetchTasks(true)
  } catch (err: any) {
    toast.error('Gagal mengirim email via SMTP: ' + (err.response?.data?.error || err.message))
  } finally {
    isSendingSmtp.value = false
  }
}

// ─────────────────────────────────────────────────────────────────────────────
// BULK EMAIL DESAIN KE VENDOR
// ─────────────────────────────────────────────────────────────────────────────

const bulkEmailTotalArea = computed(() => {
  return bulkEmailTasks.value.reduce((sum, item) => sum + (getBannerAreaM2(item) || 0), 0).toFixed(2)
})

const bulkEmailTotalPrice = computed(() => {
  return bulkEmailTasks.value.reduce((sum, item) => sum + (Number(item.total_price) || 0), 0)
})

const alreadySentEmailCount = computed(() => {
  return bulkEmailTasks.value.filter((t) => t.is_vendor_email_sent).length
})

const openBulkEmailModal = () => {
  if (selectedBannerIds.value.length === 0) {
    toast.error('Pilih setidaknya satu spanduk untuk dikirim email desainnya')
    return
  }

  const selectedTasks = purchasingTasks.value.filter((t) => selectedBannerIds.value.includes(t.id))
  if (selectedTasks.length === 0) {
    toast.error('Tidak ada data banner yang dipilih pada halaman ini')
    return
  }

  bulkEmailTasks.value = selectedTasks
  bulkEmailSelectedVendorId.value = ''
  bulkEmailVendorName.value = ''
  bulkEmailRecipient.value = ''
  bulkEmailCustomNotes.value = ''

  // Jika semua banner yang dipilih memiliki vendor yang sama, auto-select vendor tersebut
  const firstVendorId = selectedTasks[0].vendor_id
  const firstVendorName = selectedTasks[0].vendor_name
  const isSameVendor = selectedTasks.every((t) => (t.vendor_id && t.vendor_id === firstVendorId) || (t.vendor_name && t.vendor_name === firstVendorName))

  if (isSameVendor) {
    const matchedVendor = getVendorByTask(selectedTasks[0])
    if (matchedVendor) {
      bulkEmailSelectedVendorId.value = matchedVendor.id
      bulkEmailVendorName.value = matchedVendor.name
      bulkEmailRecipient.value = matchedVendor.email || ''
    } else if (firstVendorName && firstVendorName !== 'Belum ditentukan') {
      bulkEmailVendorName.value = firstVendorName
    }
  }

  bulkEmailSubject.value = generateBulkVendorEmailSubject(selectedTasks.length, bulkEmailVendorName.value || 'Vendor')
  showBulkEmailModal.value = true
}

const onBulkEmailVendorSelectChange = () => {
  if (!bulkEmailSelectedVendorId.value) {
    bulkEmailVendorName.value = ''
    bulkEmailRecipient.value = ''
    return
  }
  const v = vendorOptions.value.find((item: any) => String(item.id) === String(bulkEmailSelectedVendorId.value))
  if (v) {
    bulkEmailVendorName.value = v.name
    bulkEmailRecipient.value = v.email || ''
    bulkEmailSubject.value = generateBulkVendorEmailSubject(bulkEmailTasks.value.length, v.name)
  }
}

const getTaskArea = (t: any): number => {
  if (typeof t.area === 'number' && t.area > 0) return t.area
  const m2 = getBannerAreaM2(t)
  return typeof m2 === 'number' ? m2 : parseFloat(String(m2 || 0)) || 0
}

const getTaskPrice = (t: any): number => {
  if (t.price !== undefined && t.price !== null && !isNaN(Number(t.price))) return Number(t.price)
  if (t.total_price !== undefined && t.total_price !== null && !isNaN(Number(t.total_price))) return Number(t.total_price)
  return 0
}

const getTaskAddress = (t: any): string => {
  if (t.shop_address && t.shop_address.trim() && t.shop_address !== '-') {
    return t.shop_address.trim()
  }
  if (t.address && t.address.trim() && t.address !== '-') {
    return t.address.trim()
  }
  if (t.location && t.location.trim() && t.location !== '-') {
    return t.location.trim()
  }
  const parts = [t.village, t.district, t.regency, t.province].filter(
    (p: any) => p && typeof p === 'string' && p.trim() && p !== '-'
  )
  return parts.length > 0 ? parts.join(', ') : '-'
}

const generateBulkEmailHtml = (tasks: any[], vName: string, customNotes?: string) => {
  if (!tasks || tasks.length === 0) return ''

  const computedTotalArea = tasks
    .reduce((sum, item) => sum + (getTaskArea(item) || 0), 0)
    .toFixed(2)
  const computedTotalPrice = tasks
    .reduce((sum, item) => sum + (getTaskPrice(item) || 0), 0)

  const rowsHtml = tasks
    .map((t, idx) => {
      const areaVal = getTaskArea(t)
      const areaText = areaVal > 0 ? `${areaVal.toFixed(2)} m²` : '-'
      const priceVal = getTaskPrice(t)
      const formattedPrice = priceVal > 0 ? formatCurrency(priceVal) : 'Rp 0'
      const rawDesign = (t.design_url || '').trim()
      const hasValidDesignLink = rawDesign.startsWith('http://') || rawDesign.startsWith('https://')
      const notes = t.notes && t.notes !== '-' ? t.notes : 'Tidak ada'
      const address = getTaskAddress(t)

      const designCell = hasValidDesignLink
        ? `<a href="${rawDesign}" target="_blank" style="display: inline-block; background-color: #2563eb; color: #ffffff; text-decoration: none; padding: 6px 12px; border-radius: 6px; font-weight: 700; font-size: 11px; white-space: nowrap;">Unduh Desain</a>`
        : `<span style="color: #94a3b8; font-style: italic; font-size: 11px;">${rawDesign || 'Link menyusul'}</span>`

      return `
        <tr style="border-bottom: 1px solid #e2e8f0;">
          <td style="padding: 10px 8px; font-weight: 700; color: #64748b; text-align: center; vertical-align: top;">${idx + 1}</td>
          <td style="padding: 10px 8px; vertical-align: top;">
            <strong style="color: #0f172a; font-size: 13px;">${t.shop_name}</strong><br>
            <span style="color: #64748b; font-size: 11px;">${address || '-'}</span>
            ${notes !== 'Tidak ada' ? `<br><span style="color: #d97706; font-size: 11px;">Catatan: ${notes}</span>` : ''}
          </td>
          <td style="padding: 10px 8px; vertical-align: top; font-size: 12px; color: #334155; white-space: nowrap;">
            ${formatBannerDimension(t.size_length, t.size_width, 'dual')}<br>
            <span style="color: #64748b; font-size: 11px;">(${areaText} • ${t.qty || 1} Pcs)</span>
          </td>
          <td style="padding: 10px 8px; vertical-align: top; font-size: 12px; font-weight: 700; color: #059669; white-space: nowrap;">
            ${formattedPrice}
          </td>
          <td style="padding: 10px 8px; vertical-align: top; text-align: center;">
            ${designCell}
          </td>
        </tr>
      `
    })
    .join('')

  const customNotesHtml = customNotes && customNotes.trim()
    ? `<div style="margin-top: 16px; padding: 12px; background-color: #f8fafc; border-left: 4px solid #3b82f6; border-radius: 4px;">
        <strong style="color: #1e293b; font-size: 13px;">Catatan Tambahan:</strong><br>
        <span style="color: #475569; font-size: 13px; white-space: pre-wrap;">${customNotes.trim()}</span>
       </div>`
    : ''

  return `
    <div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; max-width: 720px; margin: 0 auto; padding: 24px; background-color: #ffffff; border: 1px solid #e2e8f0; border-radius: 16px; color: #1e293b; line-height: 1.5;">
      <div style="border-bottom: 2px solid #2563eb; padding-bottom: 14px; margin-bottom: 20px;">
        <h2 style="margin: 0; color: #1e293b; font-size: 20px; font-weight: 800;">ORDER CETAK SPANDUK GABUNGAN (SPK BULK)</h2>
        <p style="margin: 4px 0 0 0; color: #64748b; font-size: 12px;">Sistem Operasional Percetakan KARSA HR</p>
      </div>

      <p style="font-size: 14px; margin-bottom: 16px;">
        Halo Tim Percetakan <strong>${vName || 'Vendor'}</strong>,
      </p>
      <p style="font-size: 14px; margin-bottom: 20px; color: #334155;">
        Berikut adalah Surat Perintah Kerja (SPK) cetak gabungan untuk <strong>${tasks.length} spanduk/banner</strong> siap cetak:
      </p>

      <!-- Summary Box -->
      <div style="background-color: #f8fafc; border: 1px solid #e2e8f0; border-radius: 12px; padding: 14px 18px; margin-bottom: 20px;">
        <table style="width: 100%; font-size: 13px;">
          <tr>
            <td style="color: #64748b; width: 33%;">Total Pesanan: <strong style="color: #0f172a;">${tasks.length} Banner</strong></td>
            <td style="color: #64748b; width: 33%;">Total Luas: <strong style="color: #2563eb;">${computedTotalArea} m²</strong></td>
            <td style="color: #64748b; width: 34%;">Total Biaya: <strong style="color: #059669;">${formatCurrency(computedTotalPrice)}</strong></td>
          </tr>
        </table>
      </div>

      <!-- Table of items -->
      <div style="border: 1px solid #e2e8f0; border-radius: 12px; overflow: hidden; margin-bottom: 20px;">
        <table style="width: 100%; border-collapse: collapse; font-size: 12px;">
          <thead style="background-color: #f1f5f9; border-bottom: 2px solid #cbd5e1; text-transform: uppercase; font-size: 10px; color: #475569;">
            <tr>
              <th style="padding: 8px; width: 30px;">No</th>
              <th style="padding: 8px; text-align: left;">Nama Toko & Alamat</th>
              <th style="padding: 8px; text-align: left;">Ukuran & Qty</th>
              <th style="padding: 8px; text-align: left;">Biaya Cetak</th>
              <th style="padding: 8px; text-align: center;">File Desain</th>
            </tr>
          </thead>
          <tbody>
            ${rowsHtml}
          </tbody>
        </table>
      </div>

      ${customNotesHtml}

      <div style="margin-top: 24px; padding-top: 16px; border-top: 1px solid #e2e8f0; font-size: 12px; color: #64748b; text-align: center;">
        <p style="margin: 0 0 4px 0;">Mohon konfirmasi jika file-file desain siap cetak di atas sudah dapat diunduh dan proses cetak dapat segera dimulai.</p>
        <p style="margin: 0; font-weight: 600;">Terima kasih atas kerjasamanya.</p>
      </div>
    </div>
  `
}

const sendBulkSmtpEmail = async () => {
  if (bulkEmailTasks.value.length === 0) return
  if (!bulkEmailRecipient.value || !bulkEmailRecipient.value.includes('@')) {
    toast.error('Masukkan alamat email vendor penerima yang valid')
    return
  }

  isSendingBulkEmail.value = true
  try {
    const tasks = bulkEmailTasks.value
    const vName = bulkEmailVendorName.value || 'Vendor'
    const htmlBody = generateBulkEmailHtml(tasks, vName, bulkEmailCustomNotes.value)

    const res = await apiClient.post('/branding/banners/send-vendor-email', {
      banner_ids: tasks.map((t) => t.id),
      vendor_id: bulkEmailSelectedVendorId.value || undefined,
      vendor_name: vName,
      vendor_email: bulkEmailRecipient.value.trim(),
      subject: bulkEmailSubject.value.trim(),
      custom_notes: bulkEmailCustomNotes.value.trim(),
      html_body: htmlBody,
    })

    toast.success(res.data?.message || `Email SPK (${tasks.length} banner) berhasil dikirim ke ${bulkEmailRecipient.value}`)
    showBulkEmailModal.value = false
    clearSelection()
    fetchTasks(true)
  } catch (err: any) {
    toast.error('Gagal mengirim email bulk via SMTP: ' + (err.response?.data?.error || err.message))
  } finally {
    isSendingBulkEmail.value = false
  }
}

const markSelectedAsEmailSent = async () => {
  if (selectedBannerIds.value.length === 0) return
  try {
    toast.loading('Menandai status email sudah dikirim...', { id: 'mark-email-sent' })
    await apiClient.put('/branding/banners/bulk-vendor-process', {
      action: 'mark_email_sent',
      banner_ids: selectedBannerIds.value,
    })
    toast.success(`Berhasil menandai ${selectedBannerIds.value.length} banner sebagai Email Terkirim`, { id: 'mark-email-sent' })
    fetchTasks(true)
  } catch (err: any) {
    toast.error('Gagal memperbarui status email: ' + (err.response?.data?.error || err.message), { id: 'mark-email-sent' })
  }
}

const launchWhatsApp = (task: any, vPhone: string, vName: string, price: number | string) => {
  let clean = (vPhone || '').replace(/[^0-9]/g, '')
  if (clean.startsWith('0')) {
    clean = '62' + clean.slice(1)
  }
  const body = generateVendorEmailBody(task, vName, price)
  const waUrl = clean
    ? `https://wa.me/${clean}?text=${encodeURIComponent(body)}`
    : `https://api.whatsapp.com/send?text=${encodeURIComponent(body)}`
  window.open(waUrl, '_blank')
}

const submitToVendor = async () => {
  if (!selectedTask.value || !vendorName.value || !vendorPrice.value) {
    toast.error('Pilih vendor dan masukkan total harga')
    return
  }

  const currentTask = selectedTask.value
  const vName = vendorName.value
  const vEmail = vendorEmail.value
  const vPrice = vendorPrice.value
  const willOpenGmail = autoOpenGmail.value

  let popupWin: Window | null = null
  if (willOpenGmail) {
    try {
      popupWin = window.open('about:blank', '_blank')
      if (popupWin && popupWin.document) {
        popupWin.document.write(`<!DOCTYPE html><html><head><title>Membuka Gmail Web...</title><meta charset="utf-8"><style>body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;display:flex;align-items:center;justify-content:center;height:100vh;margin:0;background:#f8fafc;color:#1e293b;}.card{background:white;padding:28px 36px;border-radius:20px;box-shadow:0 10px 25px -5px rgba(0,0,0,0.08);text-align:center;border:1px solid #e2e8f0;}.spin{width:36px;height:36px;border:3px solid #e2e8f0;border-top-color:#3b82f6;border-radius:50%;animation:s 0.8s linear infinite;margin:0 auto 14px;}@keyframes s{to{transform:rotate(360deg);}}</style></head><body><div class="card"><div class="spin"></div><h3 style="margin:0 0 6px;font-size:16px;font-weight:700;">Menyiapkan SPK Gmail</h3><p style="margin:0;font-size:13px;color:#64748b;">Mengalihkan ke Gmail Web...</p></div></body></html>`)
        popupWin.document.close()
      }
    } catch (err) {
      console.warn('Unable to pre-open window', err)
    }
  }

  isSubmitting.value = true
  try {
    await apiClient.put(`/branding/banners/${currentTask.id}/vendor-process`, {
      action: 'send_to_vendor',
      vendor_id: selectedVendorId.value || undefined,
      vendor_name: vName,
      total_price: Number(vPrice),
    })

    const gmailUrl = getGmailUrl(currentTask, vEmail, vName, vPrice)
    let opened = false
    if (willOpenGmail && popupWin && !popupWin.closed) {
      popupWin.location.href = gmailUrl
      opened = true
    } else if (willOpenGmail) {
      const win = window.open(gmailUrl, '_blank')
      if (win && !win.closed) opened = true
    }

    toast.success(`Banner ${currentTask.shop_name} berhasil dikirim ke Vendor`, {
      description: !opened && willOpenGmail ? 'Tab diblokir browser. Klik tombol di bawah untuk membuka Gmail.' : undefined,
      action: {
        label: 'Buka Gmail',
        onClick: () => launchGmail(currentTask, vEmail, vName, vPrice),
      },
    })

    showVendorModal.value = false
    fetchTasks(true)
  } catch (e: any) {
    if (popupWin && !popupWin.closed) {
      popupWin.close()
    }
    toast.error('Gagal memproses ke vendor: ' + (e.response?.data?.error || e.message))
  } finally {
    isSubmitting.value = false
  }
}

// ==================== EDIT BIAYA & VENDOR PERCETAKAN ====================

const calculateTaskPrice = (task: any, vendor: any) => {
  if (!task || !vendor || !vendor.price_per_meter || vendor.price_per_meter <= 0) return 0
  const area = (task.size_length / 100) * (task.size_width / 100)
  const qty = task.qty || 1
  return Math.round(area * vendor.price_per_meter * qty)
}

const openEditPriceModal = (task: any) => {
  editPriceTask.value = task
  editPriceVendorId.value = task.vendor_id || ''
  editPriceVendorName.value = task.vendor_name && task.vendor_name !== 'Belum ditentukan' ? task.vendor_name : ''
  editPriceVendorEmail.value = ''
  editPriceVendorPhone.value = ''
  editPriceValue.value = task.total_price || ''

  const matchedVendor = getVendorByTask(task)
  if (matchedVendor) {
    editPriceVendorId.value = matchedVendor.id
    editPriceVendorName.value = matchedVendor.name
    editPriceVendorEmail.value = matchedVendor.email || ''
    editPriceVendorPhone.value = matchedVendor.phone || ''
    if (!editPriceValue.value && matchedVendor.price_per_meter > 0) {
      editPriceValue.value = calculateTaskPrice(task, matchedVendor)
    }
  }

  showEditPriceModal.value = true
}

const onEditPriceVendorSelectChange = () => {
  if (!editPriceVendorId.value) {
    editPriceVendorName.value = ''
    editPriceVendorEmail.value = ''
    editPriceVendorPhone.value = ''
    return
  }
  const v = vendorOptions.value.find((item: any) => String(item.id) === String(editPriceVendorId.value))
  if (v) {
    editPriceVendorName.value = v.name
    editPriceVendorEmail.value = v.email || ''
    editPriceVendorPhone.value = v.phone || ''
    if (v.price_per_meter > 0 && editPriceTask.value) {
      editPriceValue.value = calculateTaskPrice(editPriceTask.value, v)
    }
  }
}

const applyAutoCalculationToEdit = () => {
  if (!editPriceTask.value) return
  let vendor = vendorOptions.value.find((v: any) => String(v.id) === String(editPriceVendorId.value))
  if (!vendor && editPriceVendorName.value) {
    vendor = vendorOptions.value.find(
      (v: any) => v.name.trim().toLowerCase() === editPriceVendorName.value.trim().toLowerCase(),
    )
  }
  if (vendor && vendor.price_per_meter > 0) {
    editPriceValue.value = calculateTaskPrice(editPriceTask.value, vendor)
    toast.success(`Biaya otomatis terhitung: ${formatCurrency(Number(editPriceValue.value))} (${vendor.price_per_meter.toLocaleString('id-ID')}/m²)`)
  } else {
    toast.error('Pilih vendor yang memiliki master tarif harga/m² terlebih dahulu')
  }
}

const submitEditPrice = async () => {
  if (!editPriceTask.value) return
  if (!editPriceVendorName.value && !editPriceVendorId.value) {
    toast.error('Pilih atau masukkan nama vendor')
    return
  }
  if (editPriceValue.value === '' || Number(editPriceValue.value) < 0) {
    toast.error('Masukkan total biaya cetak yang valid')
    return
  }

  isEditPriceSubmitting.value = true
  try {
    await apiClient.put(`/branding/banners/${editPriceTask.value.id}/vendor-process`, {
      action: 'update_pricing',
      vendor_id: editPriceVendorId.value || undefined,
      vendor_name: editPriceVendorName.value,
      total_price: Number(editPriceValue.value),
    })

    toast.success(`Biaya cetak & vendor untuk ${editPriceTask.value.shop_name} berhasil diperbarui`)
    showEditPriceModal.value = false
    fetchTasks(true)
  } catch (e: any) {
    toast.error('Gagal memperbarui biaya vendor: ' + (e.response?.data?.error || e.message))
  } finally {
    isEditPriceSubmitting.value = false
  }
}

// Bulk Auto Calculate Biaya Cetak yang belum ditentukan
const autoCalculateSelectedPrices = async () => {
  if (selectedBannerIds.value.length === 0) {
    toast.error('Pilih setidaknya satu spanduk untuk dihitung biayanya')
    return
  }

  const selectedTasks = purchasingTasks.value.filter((t) => selectedBannerIds.value.includes(t.id))
  const itemsToUpdate: { banner_id: string; total_price: number; vendor_id?: string; vendor_name?: string }[] = []

  for (const t of selectedTasks) {
    const v = getVendorByTask(t)
    if (v && v.price_per_meter > 0) {
      const price = calculateTaskPrice(t, v)
      if (price > 0) {
        itemsToUpdate.push({
          banner_id: t.id,
          total_price: price,
          vendor_id: v.id,
          vendor_name: v.name,
        })
      }
    }
  }

  if (itemsToUpdate.length === 0) {
    toast.error('Tidak ada banner terpilih yang memiliki relasi Master Vendor dengan tarif per m²')
    return
  }

  if (
    !confirm(
      `Hitung Otomatis Biaya Massal:\n\nDitemukan ${itemsToUpdate.length} spanduk yang dapat dihitung otomatis dari tarif master vendor masing-masing.\nLanjutkan pembaruan harga?`,
    )
  ) {
    return
  }

  try {
    toast.loading(`Memperbarui biaya ${itemsToUpdate.length} spanduk...`, { id: 'bulk-calc' })
    await apiClient.put('/branding/banners/bulk-vendor-process', {
      action: 'update_pricing',
      items: itemsToUpdate,
    })
    toast.success(`Berhasil menghitung dan memperbarui biaya untuk ${itemsToUpdate.length} spanduk!`, { id: 'bulk-calc' })
    clearSelection()
    fetchTasks(true)
  } catch (e: any) {
    toast.error('Gagal memperbarui biaya massal: ' + (e.response?.data?.error || e.message), { id: 'bulk-calc' })
  }
}

// ==================== OPERASIONAL: SELESAI CETAK ====================

const markAsFinishedPrinting = async (task: any) => {
  if (
    !confirm(
      `Konfirmasi Selesai Cetak:\n\n• Toko: ${task.shop_name}\n• Vendor: ${task.vendor_name}\n\nApakah spanduk fisik sudah selesai dicetak dan siap diambil/dipasang oleh operator?\n(Catatan: Status pembayaran vendor tetap dicatat di tab Tagihan)`,
    )
  )
    return

  try {
    toast.loading('Menyimpan status selesai cetak...', { id: 'finish-print' })
    await apiClient.put(`/branding/banners/${task.id}/vendor-process`, {
      action: 'ready_for_operator',
    })
    toast.success(`Spanduk ${task.shop_name} berhasil ditandai Selesai Cetak (Siap Pasang)`, { id: 'finish-print' })
    fetchTasks(true)
  } catch (e: any) {
    toast.error('Gagal memperbarui status selesai cetak: ' + (e.response?.data?.error || e.message), { id: 'finish-print' })
  }
}

const submitBulkMarkAsFinishedPrinting = async () => {
  if (selectedBannerIds.value.length === 0) {
    toast.error('Pilih setidaknya satu spanduk untuk ditandai selesai cetak')
    return
  }

  const count = selectedBannerIds.value.length
  if (
    !confirm(
      `Konfirmasi Selesai Cetak Massal:\n\nApakah Anda yakin ingin menandai ${count} spanduk terpilih sebagai Selesai Dicetak (Siap Pasang oleh Operator)?`,
    )
  ) {
    return
  }

  try {
    toast.loading(`Memproses ${count} spanduk selesai cetak...`, { id: 'bulk-finish-print' })
    const payload = {
      action: 'ready_for_operator',
      items: selectedBannerIds.value.map((id) => ({ banner_id: id })),
    }
    await apiClient.put('/branding/banners/bulk-vendor-process', payload)
    toast.success(`Berhasil menandai ${count} spanduk Selesai Cetak`, { id: 'bulk-finish-print' })
    clearSelection()
    fetchTasks(true)
  } catch (e: any) {
    toast.error('Gagal memperbarui status: ' + (e.response?.data?.error || e.message), { id: 'bulk-finish-print' })
  }
}

// ==================== FINANSIAL: PELUNASAN / BAYAR VENDOR ====================

const openSinglePaymentModal = (task: any) => {
  paymentModalTasks.value = [task]
  paymentReference.value = ''
  paymentNotes.value = ''
  showPaymentModal.value = true
}

const openBulkPaymentModal = () => {
  if (selectedBannerIds.value.length === 0) {
    toast.error('Pilih setidaknya satu tagihan spanduk untuk dilunasi')
    return
  }

  const selectedTasks = purchasingTasks.value.filter((t) => selectedBannerIds.value.includes(t.id))
  if (selectedTasks.length === 0) {
    toast.error('Tidak ada data yang dipilih pada halaman ini')
    return
  }

  paymentModalTasks.value = selectedTasks
  paymentReference.value = ''
  paymentNotes.value = ''
  showPaymentModal.value = true
}

const submitPayment = async () => {
  if (paymentModalTasks.value.length === 0) return

  const count = paymentModalTasks.value.length
  const totalCost = paymentModalTotalCost.value

  isPaymentSubmitting.value = true
  try {
    toast.loading(`Menyimpan pelunasan tagihan (${count} item)...`, { id: 'submit-payment' })

    if (count === 1) {
      const task = paymentModalTasks.value[0]
      await apiClient.put(`/branding/banners/${task.id}/vendor-process`, {
        action: 'mark_paid',
        payment_reference: paymentReference.value.trim() || undefined,
        payment_notes: paymentNotes.value.trim() || undefined,
      })
    } else {
      await apiClient.put('/branding/banners/bulk-vendor-process', {
        action: 'mark_paid',
        banner_ids: paymentModalTasks.value.map((t) => t.id),
        payment_reference: paymentReference.value.trim() || undefined,
        payment_notes: paymentNotes.value.trim() || undefined,
      })
    }

    toast.success(`Berhasil melunasi ${count} tagihan vendor (${formatCurrency(totalCost)})`, { id: 'submit-payment' })
    showPaymentModal.value = false
    clearSelection()
    fetchTasks(true)
  } catch (e: any) {
    toast.error('Gagal memproses pelunasan: ' + (e.response?.data?.error || e.message), { id: 'submit-payment' })
  } finally {
    isPaymentSubmitting.value = false
  }
}

const markAsUnpaid = async (task: any) => {
  if (
    !confirm(
      `Batalkan Status Lunas:\n\n• Toko: ${task.shop_name}\n• Vendor: ${task.vendor_name}\n\nApakah Anda yakin ingin membatalkan status lunas dan mengembalikan tagihan ini ke Belum Lunas?`,
    )
  )
    return

  try {
    toast.loading('Membatalkan status lunas...', { id: 'unpaid' })
    await apiClient.put(`/branding/banners/${task.id}/vendor-process`, {
      action: 'mark_unpaid',
    })
    toast.success(`Status pembayaran spanduk ${task.shop_name} dikembalikan ke Belum Lunas`, { id: 'unpaid' })
    fetchTasks(true)
  } catch (e: any) {
    toast.error('Gagal mengubah status pembayaran: ' + (e.response?.data?.error || e.message), { id: 'unpaid' })
  }
}

// ==================== BULK ACTIONS: KIRIM KE VENDOR ====================

const bulkTotalArea = computed(() => {
  return bulkItems.value.reduce((sum, item) => sum + (item.area || 0), 0).toFixed(2)
})

const bulkTotalPrice = computed(() => {
  return bulkItems.value.reduce((sum, item) => sum + (Number(item.price) || 0), 0)
})

const openBulkModal = () => {
  if (selectedBannerIds.value.length === 0) {
    toast.error('Pilih setidaknya satu spanduk untuk diproses bulk')
    return
  }

  const selectedTasks = purchasingTasks.value.filter((t) => selectedBannerIds.value.includes(t.id))
  if (selectedTasks.length === 0) {
    toast.error('Tidak ada data banner yang dipilih pada halaman ini')
    return
  }

  // Populate bulk items
  bulkItems.value = selectedTasks.map((t) => {
    const area = getBannerAreaM2(t) || 0
    return {
      id: t.id,
      shop_name: t.shop_name,
      shop_address: t.shop_address,
      type: t.type,
      raw_type: t.raw_type,
      qty: t.qty || 1,
      size_length: t.size_length || 0,
      size_width: t.size_width || 0,
      area: area,
      price: t.total_price || 0,
      design_url: t.design_url || '',
      notes: t.notes || '-',
    }
  })

  // Reset vendor selections
  bulkSelectedVendorId.value = ''
  bulkVendorName.value = ''
  bulkVendorEmail.value = ''
  bulkVendorPhone.value = ''
  bulkCustomNotes.value = ''
  bulkSendViaSmtp.value = true

  // Auto-select vendor jika semua banner yang dipilih memiliki vendor yang sama
  const firstVendorId = selectedTasks[0].vendor_id
  const firstVendorName = selectedTasks[0].vendor_name
  const isSameVendor = selectedTasks.every(
    (t) => (t.vendor_id && t.vendor_id === firstVendorId) || (t.vendor_name && t.vendor_name === firstVendorName)
  )

  if (isSameVendor) {
    const matchedVendor = getVendorByTask(selectedTasks[0])
    if (matchedVendor) {
      bulkSelectedVendorId.value = matchedVendor.id
      bulkVendorName.value = matchedVendor.name
      bulkVendorEmail.value = matchedVendor.email || ''
      bulkVendorPhone.value = matchedVendor.phone || ''
      bulkSendViaSmtp.value = Boolean(matchedVendor.email)

      if (matchedVendor.price_per_meter > 0) {
        bulkItems.value.forEach((item) => {
          if (item.area > 0) {
            item.price = Math.round(item.area * matchedVendor.price_per_meter)
          }
        })
      }
    } else if (firstVendorName && firstVendorName !== 'Belum ditentukan') {
      bulkVendorName.value = firstVendorName
    }
  }

  showBulkModal.value = true
}

const onBulkVendorSelectChange = () => {
  if (!bulkSelectedVendorId.value) {
    bulkVendorName.value = ''
    bulkVendorEmail.value = ''
    bulkVendorPhone.value = ''
    return
  }
  const v = vendorOptions.value.find((item: any) => String(item.id) === String(bulkSelectedVendorId.value))
  if (v) {
    bulkVendorName.value = v.name
    bulkVendorEmail.value = v.email || ''
    bulkVendorPhone.value = v.phone || ''
    bulkSendViaSmtp.value = Boolean(v.email)

    if (v.price_per_meter > 0) {
      bulkItems.value.forEach((item) => {
        if (item.area > 0) {
          item.price = Math.round(item.area * v.price_per_meter)
        }
      })
    }
  }
}

const generateBulkVendorEmailSubject = (count: number, vName: string) => {
  return `[ORDER CETAK SPANDUK GABUNGAN] - ${count} Pesanan Desain Siap Cetak (${vName || 'Vendor'})`
}

const generateBulkVendorEmailBody = (items: any[], vName: string) => {
  const totalArea = items.reduce((sum, item) => sum + getTaskArea(item), 0).toFixed(2)
  const totalPrice = items.reduce((sum, item) => sum + getTaskPrice(item), 0)

  const itemsText = items
    .map((t, idx) => {
      const areaVal = getTaskArea(t)
      const areaText = areaVal > 0 ? `${areaVal.toFixed(2)} m²` : '-'
      const priceVal = getTaskPrice(t)
      const formattedPrice = priceVal > 0 ? formatCurrency(priceVal) : 'Belum dihitung'
      const dimension = formatBannerDimension(t.size_length, t.size_width, 'dual')
      const notes = t.notes && t.notes !== '-' ? t.notes : ''
      const design = t.design_url || 'Link desain menyusul'
      const address = getTaskAddress(t)

      let details = `${idx + 1}. TOKO: ${t.shop_name}
   - Alamat Toko    : ${address}
   - Jenis & Ukuran : ${t.raw_type || 'Spanduk'} (${dimension})
   - Luas & Qty     : ${areaText} (${t.qty || 1} Pcs)
   - Biaya Cetak    : ${formattedPrice}`

      if (notes) {
        details += `\n   - Catatan        : ${notes}`
      }
      details += `\n   - Link Desain    : ${design}`

      return details
    })
    .join('\n\n')

  return `Halo Tim Percetakan ${vName || 'Vendor'},

Berikut adalah detail Surat Perintah Kerja (SPK) CETAK GABUNGAN untuk ${items.length} Spanduk/Banner:

==================================================
RINGKASAN ORDER GABUNGAN:
- Total Pesanan   : ${items.length} Banner
- Total Luas Area : ${totalArea} m²
- Total Estimasi  : ${formatCurrency(totalPrice)}
==================================================

DETAIL RINCIAN PER PESANAN:

${itemsText}

==================================================
Mohon konfirmasi jika link file desain siap cetak di atas sudah dapat diakses dan proses cetak segera dimulai.
Terima kasih.`
}

const generateCompactBulkVendorEmailBody = (items: any[], vName: string) => {
  const totalArea = items.reduce((sum, item) => sum + getTaskArea(item), 0).toFixed(2)
  const totalPrice = items.reduce((sum, item) => sum + getTaskPrice(item), 0)

  // Buat daftar dengan 5 poin lengkap (Alamat Toko, Jenis & Ukuran, Luas & Qty, Biaya Cetak, Link Desain)
  const compactBlocks: string[] = []
  let charBudget = 4200 // Alokasi karakter aman untuk rincian item dalam URL

  let includedCount = 0
  for (let i = 0; i < items.length; i++) {
    const t = items[i]
    const areaVal = getTaskArea(t)
    const areaText = areaVal > 0 ? `${areaVal.toFixed(2)} m²` : '-'
    const priceVal = getTaskPrice(t)
    const formattedPrice = priceVal > 0 ? formatCurrency(priceVal) : 'Belum dihitung'
    const dim = formatBannerDimension(t.size_length, t.size_width, 'dual')
    const design = t.design_url || 'Link desain menyusul'
    const address = getTaskAddress(t)

    const block = `${i + 1}. TOKO: ${t.shop_name}
   - Alamat Toko    : ${address}
   - Jenis & Ukuran : ${t.raw_type || 'Spanduk'} (${dim})
   - Luas & Qty     : ${areaText} (${t.qty || 1} Pcs)
   - Biaya Cetak    : ${formattedPrice}
   - Link Desain    : ${design}`

    if (charBudget - block.length < 0) {
      break
    }
    compactBlocks.push(block)
    charBudget -= block.length + 2
    includedCount++
  }

  const remaining = items.length - includedCount
  const remainingText = remaining > 0
    ? `\n\n... dan ${remaining} pesanan lainnya. Format SPK lengkap (100% rincian) telah disalin otomatis ke clipboard. Silakan Paste (Ctrl+V) jika ingin melampirkan seluruh rincian toko.`
    : ''

  return `Halo Tim Percetakan ${vName || 'Vendor'},

Berikut adalah ringkasan Surat Perintah Kerja (SPK) CETAK GABUNGAN untuk ${items.length} Spanduk/Banner:

==================================================
RINGKASAN ORDER GABUNGAN:
- Total Pesanan   : ${items.length} Banner
- Total Luas Area : ${totalArea} m²
- Total Estimasi  : ${formatCurrency(totalPrice)}
==================================================

DAFTAR RINCIAN PESANAN SIAP CETAK:

${compactBlocks.join('\n\n')}${remainingText}

==================================================
Mohon konfirmasi jika file desain sudah diterima dan proses cetak dapat segera dimulai.
Terima kasih.`
}

// Batas aman panjang URL browser/Gmail Web (Google Web Server limit ~8192 bytes)
const GMAIL_URL_MAX_LENGTH = 7000

const getBulkGmailUrl = (items: any[], vEmail: string, vName: string) => {
  const subject = generateBulkVendorEmailSubject(items.length, vName)
  const fullBody = generateBulkVendorEmailBody(items, vName)
  const targetEmail = vEmail ? vEmail.trim() : ''

  const fullUrl = `https://mail.google.com/mail/?view=cm&to=${encodeURIComponent(targetEmail)}&su=${encodeURIComponent(subject)}&body=${encodeURIComponent(fullBody)}`

  // Jika URL terlalu panjang, gunakan compact body agar tidak memicu HTTP 400 Bad Request
  if (fullUrl.length > GMAIL_URL_MAX_LENGTH) {
    const compactBody = generateCompactBulkVendorEmailBody(items, vName)
    return {
      url: `https://mail.google.com/mail/?view=cm&to=${encodeURIComponent(targetEmail)}&su=${encodeURIComponent(subject)}&body=${encodeURIComponent(compactBody)}`,
      isCompact: true,
      fullBody: fullBody,
    }
  }

  return {
    url: fullUrl,
    isCompact: false,
    fullBody: fullBody,
  }
}

const launchBulkGmail = async (items: any[], vEmail: string, vName: string) => {
  const { url, isCompact, fullBody } = getBulkGmailUrl(items, vEmail, vName)

  if (isCompact) {
    try {
      await navigator.clipboard.writeText(fullBody)
      toast.info(
        `Karena pesanan berjumlah besar (${items.length} banner), rincian lengkap SPK otomatis disalin ke clipboard. Gunakan Paste (Ctrl+V) jika diperlukan!`,
        { duration: 8000 }
      )
    } catch {
      console.warn('Gagal menyalin teks ke clipboard secara otomatis')
    }
  }

  const win = window.open(url, '_blank')
  if (!win || win.closed || typeof win.closed === 'undefined') {
    const subject = generateBulkVendorEmailSubject(items.length, vName)
    const bodyToUse = isCompact ? generateCompactBulkVendorEmailBody(items, vName) : fullBody
    window.location.href = `mailto:${encodeURIComponent(vEmail ? vEmail.trim() : '')}?subject=${encodeURIComponent(subject)}&body=${encodeURIComponent(bodyToUse)}`
  }
}

const copyBulkSpkFormat = async (items: any[], vName: string) => {
  const body = generateBulkVendorEmailBody(items, vName)
  try {
    await navigator.clipboard.writeText(body)
    toast.success(`Format SPK (${items.length} pesanan) berhasil disalin ke clipboard!`)
  } catch {
    toast.error('Gagal menyalin format SPK')
  }
}

const launchBulkWa = (items: any[], vPhone: string, vName: string) => {
  let clean = (vPhone || '').replace(/[^0-9]/g, '')
  if (clean.startsWith('0')) {
    clean = '62' + clean.slice(1)
  }
  const body = generateBulkVendorEmailBody(items, vName)
  const waUrl = clean
    ? `https://wa.me/${clean}?text=${encodeURIComponent(body)}`
    : `https://api.whatsapp.com/send?text=${encodeURIComponent(body)}`
  window.open(waUrl, '_blank')
}

const submitBulkToVendor = async () => {
  if (bulkItems.value.length === 0) {
    toast.error('Tidak ada banner yang diproses')
    return
  }

  const vName = bulkVendorName.value.trim()
  if (!bulkSelectedVendorId.value && !vName) {
    toast.error('Pilih vendor atau masukkan nama vendor')
    return
  }

  const currentItems = [...bulkItems.value]
  const vEmail = bulkVendorEmail.value ? bulkVendorEmail.value.trim() : ''
  const shouldSendSmtp = bulkSendViaSmtp.value

  if (shouldSendSmtp && (!vEmail || !vEmail.includes('@'))) {
    toast.error('Masukkan alamat email vendor yang valid untuk pengiriman SPK via SMTP')
    return
  }

  isBulkSubmitting.value = true
  try {
    const payload = {
      action: 'send_to_vendor',
      vendor_id: bulkSelectedVendorId.value || undefined,
      vendor_name: vName,
      items: currentItems.map((item) => ({
        banner_id: item.id,
        total_price: Number(item.price) || 0,
      })),
    }

    await apiClient.put('/branding/banners/bulk-vendor-process', payload)

    let smtpSuccess = false
    let smtpErrMsg = ''
    if (shouldSendSmtp && vEmail) {
      try {
        const htmlBody = generateBulkEmailHtml(currentItems, vName, bulkCustomNotes.value)
        await apiClient.post('/branding/banners/send-vendor-email', {
          banner_ids: currentItems.map((item) => item.id),
          vendor_id: bulkSelectedVendorId.value || undefined,
          vendor_name: vName,
          vendor_email: vEmail,
          subject: generateBulkVendorEmailSubject(currentItems.length, vName),
          custom_notes: bulkCustomNotes.value?.trim() || undefined,
          html_body: htmlBody,
        })
        smtpSuccess = true
      } catch (err: any) {
        smtpErrMsg = err.response?.data?.error || err.message
        console.error('Failed to send bulk SMTP email:', err)
      }
    }

    if (shouldSendSmtp && smtpSuccess) {
      toast.success(
        `Berhasil menugaskan ${currentItems.length} banner dan email SPK terkirim via SMTP ke ${vEmail}`
      )
    } else if (shouldSendSmtp && !smtpSuccess) {
      toast.warning(
        `Status ${currentItems.length} banner diperbarui ke vendor, namun email SMTP gagal dikirim: ${smtpErrMsg}`
      )
    } else {
      toast.success(`Berhasil menugaskan ${currentItems.length} banner ke vendor ${vName}`)
    }

    showBulkModal.value = false
    clearSelection()
    fetchTasks(true)
  } catch (e: any) {
    toast.error('Gagal memproses bulk ke vendor: ' + (e.response?.data?.error || e.message))
  } finally {
    isBulkSubmitting.value = false
  }
}
</script>

<template>
  <div class="space-y-8 pb-10">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-6">
      <div>
        <div class="flex items-center flex-wrap gap-2 sm:gap-3">
          <h1 class="text-3xl font-extrabold text-slate-900 tracking-tight flex items-center flex-wrap gap-2">
            <span>Purchasing & Percetakan</span>
            <span class="text-primary text-2xl font-bold">
              — {{ selectedBranchName }}
            </span>
          </h1>
          <span
            v-if="!canManagePurchasing"
            class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-black bg-slate-100 text-slate-600 border border-slate-200 shadow-sm"
          >
            <Eye class="w-3.5 h-3.5 text-slate-400" /> Mode Pantau
          </span>
        </div>
        <p class="text-[15px] font-medium text-slate-500 mt-1.5 flex items-center gap-2">
          <ShoppingCart class="w-4 h-4 text-primary" />
          Kelola pesanan ke vendor percetakan, status selesai cetak, dan rekap pelunasan tagihan tempo.
        </p>
      </div>
      <button
        @click="fetchTasks()"
        class="flex items-center gap-2 bg-white px-4 py-2 rounded-xl border border-slate-200 shadow-sm font-bold text-slate-600 hover:bg-slate-50 transition"
      >
        <RefreshCw :class="['w-4 h-4', { 'animate-spin': isLoading }]" /> Refresh
      </button>
    </div>

    <div
      v-if="!isOnline"
      class="bg-red-50 text-red-600 p-4 rounded-xl border border-red-200 flex items-center gap-3"
    >
      <AlertTriangle class="w-5 h-5" />
      <span class="font-medium"
        >Sistem sedang offline atau tidak dapat menghubungi server Branding.</span
      >
    </div>

    <!-- Summary Stats (Total Biaya, Total Luas, Total Banner) -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
      <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-sm flex items-center gap-4">
        <div
          :class="[
            'w-12 h-12 rounded-xl border flex items-center justify-center shrink-0',
            activeTab === 'unpaid_bills'
              ? 'bg-amber-50 border-amber-200 text-amber-600'
              : activeTab === 'paid_history'
              ? 'bg-emerald-50 border-emerald-200 text-emerald-600'
              : 'bg-emerald-50 border-emerald-100 text-emerald-600',
          ]"
        >
          <DollarSign class="w-6 h-6" />
        </div>
        <div>
          <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">
            {{ activeTab === 'unpaid_bills' ? 'Total Tagihan Belum Lunas' : activeTab === 'paid_history' ? 'Total Sudah Dibayar' : 'Total Biaya Cetak' }}
          </p>
          <p class="text-lg font-black text-slate-900 mt-0.5">{{ formatCurrency(totalCostCurrentPage) }}</p>
          <p class="text-[11px] font-medium text-slate-400">Total data pada halaman ini</p>
        </div>
      </div>

      <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-sm flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-blue-50 border border-blue-100 flex items-center justify-center text-blue-600 shrink-0">
          <Ruler class="w-6 h-6" />
        </div>
        <div>
          <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">Total Luas Banner</p>
          <p class="text-lg font-black text-slate-900 mt-0.5">{{ totalAreaCurrentPage }} m²</p>
          <p class="text-[11px] font-medium text-slate-400">Estimasi luas pada halaman ini</p>
        </div>
      </div>

      <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-sm flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-indigo-50 border border-indigo-100 flex items-center justify-center text-indigo-600 shrink-0">
          <Layers class="w-6 h-6" />
        </div>
        <div>
          <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">Jumlah Pesanan</p>
          <p class="text-lg font-black text-slate-900 mt-0.5">{{ total }} Banner</p>
          <p class="text-[11px] font-medium text-slate-400">Total item pada tab ini</p>
        </div>
      </div>
    </div>

    <!-- Navigation Tabs -->
    <div class="flex gap-4 border-b border-slate-200 overflow-x-auto pb-1">
      <button
        @click="activeTab = 'purchasing_queue'"
        :class="[
          'whitespace-nowrap pb-3 text-sm font-bold transition-all border-b-2 flex items-center gap-2',
          activeTab === 'purchasing_queue'
            ? 'border-primary text-primary'
            : 'border-transparent text-slate-400 hover:text-slate-600',
        ]"
      >
        <Clock class="w-4 h-4" />
        Antrian Kirim Vendor
      </button>

      <button
        @click="activeTab = 'sent_to_vendor'"
        :class="[
          'whitespace-nowrap pb-3 text-sm font-bold transition-all border-b-2 flex items-center gap-2',
          activeTab === 'sent_to_vendor'
            ? 'border-orange-500 text-orange-500'
            : 'border-transparent text-slate-400 hover:text-slate-600',
        ]"
      >
        <Printer class="w-4 h-4" />
        Di Percetakan (Proses Cetak)
      </button>

      <button
        @click="activeTab = 'unpaid_bills'"
        :class="[
          'whitespace-nowrap pb-3 text-sm font-bold transition-all border-b-2 flex items-center gap-2',
          activeTab === 'unpaid_bills'
            ? 'border-amber-500 text-amber-600 font-extrabold'
            : 'border-transparent text-slate-400 hover:text-slate-600',
        ]"
      >
        <Receipt class="w-4 h-4" />
        Tagihan Belum Lunas (Tempo)
      </button>

      <button
        @click="activeTab = 'paid_history'"
        :class="[
          'whitespace-nowrap pb-3 text-sm font-bold transition-all border-b-2 flex items-center gap-2',
          activeTab === 'paid_history'
            ? 'border-emerald-500 text-emerald-600 font-extrabold'
            : 'border-transparent text-slate-400 hover:text-slate-600',
        ]"
      >
        <CheckCircle2 class="w-4 h-4" />
        Riwayat Lunas
      </button>
    </div>

    <!-- Controls: Search, Branch Filter, Vendor Filter, Month Filter & Pagination -->
    <div
      class="flex flex-col lg:flex-row items-center justify-between gap-4 bg-white p-4 rounded-2xl border border-slate-200 shadow-sm"
    >
      <div class="flex flex-col sm:flex-row items-center gap-3 w-full lg:w-auto">
        <!-- Search -->
        <div class="relative w-full sm:w-64">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input
            type="text"
            v-model="search"
            placeholder="Cari toko / vendor..."
            class="w-full pl-9 pr-4 py-2 bg-slate-50 border border-slate-200 rounded-xl text-xs font-medium focus:outline-none focus:border-primary focus:bg-white transition"
          />
        </div>

        <!-- Filter Cabang -->
        <div class="relative w-full sm:w-48">
          <Building2 class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 pointer-events-none" />
          <select
            v-model="selectedBranchFilter"
            @change="() => { page = 1; clearSelection(); fetchTasks() }"
            class="w-full pl-9 pr-8 py-2 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold text-slate-700 outline-none focus:border-primary focus:bg-white transition cursor-pointer"
          >
            <option value="ALL_BRANCHES">Semua Cabang</option>
            <option disabled>──────────</option>
            <option v-for="branch in masterStore.branches" :key="branch.id" :value="String(branch.id)">
              {{ branch.name }}
            </option>
          </select>
        </div>

        <!-- Filter Vendor Percetakan (Especially useful for Unpaid & History) -->
        <div class="relative w-full sm:w-56">
          <Building class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 pointer-events-none" />
          <select
            v-model="selectedFilterVendorId"
            class="w-full pl-9 pr-8 py-2 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold text-slate-700 outline-none focus:border-primary focus:bg-white transition cursor-pointer"
          >
            <option value="all">Semua Vendor Percetakan</option>
            <option v-for="v in vendorOptions" :key="v.id" :value="v.id">
              {{ v.name }}
            </option>
          </select>
        </div>

        <!-- Filter Bulan/Tahun -->
        <div class="relative w-full sm:w-44">
          <Calendar class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 pointer-events-none" />
          <input
            type="month"
            v-model="selectedFilterMonth"
            class="w-full pl-9 pr-3 py-1.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold text-slate-700 outline-none focus:border-primary focus:bg-white transition cursor-pointer"
            title="Filter Bulan Tagihan"
          />
        </div>
      </div>

      <!-- Pagination -->
      <div class="flex items-center gap-4 w-full lg:w-auto justify-between lg:justify-end">
        <span class="text-xs font-medium text-slate-500">
          Halaman {{ page }} dari {{ Math.ceil(total / limit) || 1 }} ({{ total }} Data)
        </span>
        <div class="flex gap-2">
          <button
            @click="prevPage"
            :disabled="page === 1"
            class="p-2 rounded-lg border border-slate-200 text-slate-600 hover:bg-slate-50 disabled:opacity-50 disabled:cursor-not-allowed transition"
          >
            <ChevronLeft class="w-4 h-4" />
          </button>
          <button
            @click="nextPage"
            :disabled="page >= Math.ceil(total / limit) || total === 0"
            class="p-2 rounded-lg border border-slate-200 text-slate-600 hover:bg-slate-50 disabled:opacity-50 disabled:cursor-not-allowed transition"
          >
            <ChevronRight class="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>

    <!-- Data Table -->
    <div class="bg-white border border-slate-200 rounded-2xl shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr
              class="bg-slate-50 border-b border-slate-200 text-xs uppercase tracking-wider text-slate-500 font-bold"
            >
              <th v-if="canManagePurchasing && activeTab !== 'paid_history'" class="p-4 w-10 text-center">
                <input
                  type="checkbox"
                  :checked="isAllSelected"
                  :indeterminate.prop="isSomeSelected"
                  @change="toggleSelectAll"
                  class="w-4 h-4 rounded text-primary focus:ring-primary border-slate-300 cursor-pointer"
                  title="Pilih Semua di Halaman Ini"
                />
              </th>
              <th
                class="p-4 whitespace-nowrap cursor-pointer select-none hover:text-slate-800 transition"
                @click="toggleDateSort"
                title="Klik untuk mengubah urutan tanggal terbaru / terlama"
              >
                <div class="flex items-center gap-1.5">
                  <span>Tanggal</span>
                  <ArrowDown v-if="dateSortDir === 'DESC'" class="w-3.5 h-3.5 text-primary" />
                  <ArrowUp v-else-if="dateSortDir === 'ASC'" class="w-3.5 h-3.5 text-primary" />
                  <ArrowUpDown v-else class="w-3.5 h-3.5 text-slate-400" />
                </div>
              </th>
              <th class="p-4 whitespace-nowrap">Nama Toko</th>
              <th class="p-4 whitespace-nowrap">Spesifikasi & Status Fisik</th>
              <th class="p-4 whitespace-nowrap">Vendor Percetakan</th>
              <th class="p-4 whitespace-nowrap">Status Email Desain</th>
              <th class="p-4 whitespace-nowrap">Biaya & Status Bayar</th>
              <th class="p-4 text-center whitespace-nowrap">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="isLoading && purchasingTasks.length === 0">
              <td :colspan="canManagePurchasing && activeTab !== 'paid_history' ? 8 : 7" class="p-8 text-center">
                <RefreshCw class="w-8 h-8 text-primary animate-spin mx-auto mb-2" />
                <p class="text-slate-500 font-medium">Memuat data...</p>
              </td>
            </tr>
            <tr v-else-if="purchasingTasks.length === 0">
              <td :colspan="canManagePurchasing && activeTab !== 'paid_history' ? 8 : 7" class="p-12 text-center">
                <ShoppingCart class="w-12 h-12 text-slate-200 mx-auto mb-3" />
                <p class="text-slate-500 font-medium">Tidak ada data ditemukan.</p>
              </td>
            </tr>
            <tr
              v-for="task in purchasingTasks"
              :key="task.id"
              :class="[
                'transition',
                selectedBannerIds.includes(task.id) ? 'bg-primary/5 hover:bg-primary/10' : 'hover:bg-slate-50',
              ]"
            >
              <!-- Checkbox -->
              <td v-if="canManagePurchasing && activeTab !== 'paid_history'" class="p-4 text-center">
                <input
                  type="checkbox"
                  :checked="selectedBannerIds.includes(task.id)"
                  @change="toggleSelectItem(task.id)"
                  class="w-4 h-4 rounded text-primary focus:ring-primary border-slate-300 cursor-pointer"
                />
              </td>

              <!-- Tanggal -->
              <td class="p-4 text-xs text-slate-600 whitespace-nowrap">{{ task.updated_at }}</td>

              <!-- Toko Info -->
              <td class="p-4">
                <div class="font-bold text-slate-900 truncate max-w-[200px]">
                  {{ task.shop_name }}
                </div>
                <!-- Hover Card Alamat Toko -->
                <StoreAddressHoverCard
                  :address="task.shop_address"
                  :shop-name="task.shop_name"
                  :province="task.province"
                  :regency="task.regency"
                  :district="task.district"
                  :village="task.village"
                  :phone="task.shop_phone"
                  :latitude="task.latitude"
                  :longitude="task.longitude"
                  :sales-name="task.sales_name"
                  :sales-branch="task.sales_branch"
                  class="mt-1 text-[11px] max-w-[200px]"
                />
                <div
                  v-if="task.notes && task.notes !== '-'"
                  class="text-xs text-slate-500 mt-1.5 line-clamp-1 max-w-[200px]"
                  :title="task.notes"
                >
                  Catatan: {{ task.notes }}
                </div>
              </td>

              <!-- Spesifikasi & Status Fisik -->
              <td class="p-4 text-xs font-medium text-slate-700 whitespace-nowrap">
                <div class="font-bold text-slate-900">{{ task.type }}</div>
                <div v-if="getSizeSuggestion(task)" class="mt-1">
                  <span class="inline-flex items-center gap-1 text-[10px] font-extrabold text-blue-700 bg-blue-50 px-2 py-0.5 rounded border border-blue-200">
                    <Sparkles class="w-3 h-3 text-blue-500 shrink-0" />
                    Desain: {{ getSizeSuggestion(task)?.designLength }}×{{ getSizeSuggestion(task)?.designWidth }} cm
                  </span>
                </div>

                <!-- Badge Status Produksi/Fisik di Tab Tagihan & Riwayat -->
                <div v-if="activeTab === 'unpaid_bills' || activeTab === 'paid_history'" class="mt-2 flex items-center gap-1.5">
                  <span
                    v-if="task.status === 'sent_to_vendor'"
                    class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-orange-50 text-orange-700 border border-orange-200"
                  >
                    Sedang Dicetak
                  </span>
                  <span
                    v-else-if="task.status === 'ready_for_operator' || task.status === 'diterima_operator'"
                    class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-blue-50 text-blue-700 border border-blue-200"
                  >
                    Siap Pasang / Di Operator
                  </span>
                  <span
                    v-else-if="task.status === 'installed' || task.status === 'terpasang'"
                    class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-50 text-emerald-700 border border-emerald-200"
                  >
                    Terpasang
                  </span>
                </div>
              </td>

              <!-- Vendor Percetakan -->
              <td class="p-4 text-xs text-slate-600 whitespace-nowrap">
                <template v-if="task.vendor_name && task.vendor_name !== 'Belum ditentukan'">
                  <div class="font-bold text-slate-800">{{ task.vendor_name }}</div>
                  <div v-if="task.purchasing_name" class="text-[10px] text-slate-500 font-medium mt-0.5">
                    Diproses: <span class="font-bold text-slate-700">{{ task.purchasing_name }}</span>
                  </div>
                </template>
                <template v-else>
                  <span class="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg text-xs font-semibold bg-slate-100 text-slate-500 border border-slate-200">
                    <Clock class="w-3 h-3 text-slate-400" /> Belum dipilih
                  </span>
                </template>
              </td>

              <!-- Status Email Desain ke Vendor -->
              <td class="p-4 text-xs whitespace-nowrap">
                <div v-if="task.is_vendor_email_sent" class="space-y-1">
                  <span
                    class="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg text-[11px] font-extrabold bg-emerald-50 text-emerald-700 border border-emerald-200 shadow-xs"
                    title="Email desain SPK telah berhasil dikirimkan ke vendor"
                  >
                    <MailCheck class="w-3.5 h-3.5 text-emerald-600" /> Email Terkirim
                  </span>
                  <div v-if="task.vendor_email_sent_at" class="text-[10px] text-slate-400 font-medium">
                    {{ task.vendor_email_sent_at }}
                  </div>
                </div>
                <div v-else class="space-y-1">
                  <span
                    class="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg text-[11px] font-bold bg-amber-50/80 text-amber-700 border border-amber-200/80"
                    title="Email desain SPK belum dikirim ke vendor"
                  >
                    <MailX class="w-3.5 h-3.5 text-amber-500" /> Belum Dikirim
                  </span>
                  <div class="text-[10px] text-slate-400">
                    Belum ada riwayat
                  </div>
                </div>
              </td>

              <!-- Biaya Cetak & Status Bayar -->
              <td class="p-4 text-xs whitespace-nowrap">
                <div v-if="task.total_price && task.total_price > 0" class="space-y-1">
                  <div class="flex items-center gap-1.5">
                    <span class="font-extrabold text-slate-900 text-sm">
                      {{ formatCurrency(task.total_price) }}
                    </span>
                    <button
                      v-if="canManagePurchasing && (activeTab === 'sent_to_vendor' || activeTab === 'unpaid_bills')"
                      type="button"
                      @click="openEditPriceModal(task)"
                      title="Edit Biaya / Vendor"
                      class="p-1 rounded-md text-slate-400 hover:text-primary hover:bg-primary/10 transition"
                    >
                      <Pencil class="w-3.5 h-3.5" />
                    </button>
                  </div>
                  <div v-if="getBannerAreaM2(task)" class="text-[10px] font-bold text-slate-500 flex items-center gap-1">
                    <Ruler class="w-3 h-3 text-slate-400" />
                    Luas: <span class="text-slate-700">{{ getBannerAreaM2(task) }} m²</span>
                  </div>

                  <!-- Badge Status Bayar -->
                  <div class="mt-1">
                    <span
                      v-if="task.is_paid || task.payment_status === 'paid'"
                      class="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[10px] font-extrabold bg-emerald-50 text-emerald-700 border border-emerald-200"
                    >
                      <Check class="w-3 h-3 text-emerald-600" /> LUNAS
                    </span>
                    <span
                      v-else
                      class="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[10px] font-bold bg-amber-50 text-amber-700 border border-amber-200"
                    >
                      Belum Lunas
                    </span>
                  </div>

                  <div v-if="task.paid_at" class="text-[10px] text-slate-400">
                    Dibayar: {{ task.paid_at }}
                  </div>
                  <div v-if="task.payment_reference" class="text-[10px] font-mono text-slate-500 truncate max-w-[140px]" :title="task.payment_reference">
                    Ref: {{ task.payment_reference }}
                  </div>
                </div>
                <div v-else class="space-y-1">
                  <div class="flex items-center gap-1.5">
                    <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[10px] font-bold bg-amber-50 text-amber-700 border border-amber-200">
                      Belum Ditentukan
                    </span>
                    <button
                      v-if="canManagePurchasing"
                      type="button"
                      @click="openEditPriceModal(task)"
                      title="Tentukan Biaya / Vendor"
                      class="px-2 py-0.5 rounded-md text-[10px] font-bold bg-primary/10 text-primary hover:bg-primary/20 transition flex items-center gap-1"
                    >
                      <Pencil class="w-3 h-3" /> Tentukan
                    </button>
                  </div>
                  <div v-if="getBannerAreaM2(task)" class="text-[10px] font-medium text-slate-400 flex items-center gap-1">
                    <Ruler class="w-3 h-3 text-slate-400" />
                    Luas: {{ getBannerAreaM2(task) }} m²
                  </div>
                </div>
              </td>

              <!-- Aksi per Tab -->
              <td class="p-4 text-center">
                <!-- Tab 1: Antrian Cetak -->
                <template v-if="activeTab === 'purchasing_queue'">
                  <button
                    v-if="canManagePurchasing"
                    @click="openVendorModal(task)"
                    class="inline-flex items-center gap-2 bg-primary hover:bg-primary/90 text-white px-3.5 py-1.5 rounded-xl text-xs font-bold transition-all shadow-md shadow-primary/20"
                  >
                    <Truck class="w-3.5 h-3.5" /> Kirim ke Vendor
                  </button>
                  <span
                    v-else
                    class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-slate-100 text-slate-600 text-xs font-bold border border-slate-200"
                  >
                    <Clock class="w-3.5 h-3.5 text-slate-400" /> Antrian
                  </span>
                </template>

                <!-- Tab 2: Di Percetakan (Tombol: SELESAI CETAK & EDIT BIAYA) -->
                <template v-else-if="activeTab === 'sent_to_vendor'">
                  <div class="flex items-center justify-center gap-2">
                    <button
                      v-if="canManagePurchasing"
                      @click="markAsFinishedPrinting(task)"
                      title="Konfirmasi fisik spanduk selesai dicetak (Siap dipasang operator)"
                      class="inline-flex items-center gap-1.5 bg-blue-600 hover:bg-blue-700 text-white px-3.5 py-2 rounded-xl text-xs font-extrabold transition-all shadow-md shadow-blue-600/20"
                    >
                      <Printer class="w-3.5 h-3.5" /> Selesai Cetak
                    </button>
                    <button
                      v-if="canManagePurchasing"
                      type="button"
                      title="Edit Biaya Cetak / Ganti Vendor"
                      @click="openEditPriceModal(task)"
                      class="p-2 rounded-xl bg-amber-50 text-amber-700 hover:bg-amber-100 transition border border-amber-200/60"
                    >
                      <Pencil class="w-4 h-4" />
                    </button>
                    <button
                      type="button"
                      title="Kirim SPK langsung via Server SMTP"
                      @click="openSmtpConfirmModal(task)"
                      class="p-2 rounded-xl bg-indigo-50 text-indigo-600 hover:bg-indigo-100 transition border border-indigo-200/60"
                    >
                      <SendHorizontal class="w-4 h-4" />
                    </button>
                    <button
                      type="button"
                      title="Buka Draft Pesan di Gmail Web"
                      @click="launchGmail(task, getVendorByTask(task)?.email || '', task.vendor_name, task.total_price)"
                      class="p-2 rounded-xl bg-blue-50 text-blue-600 hover:bg-blue-100 transition border border-blue-200/60"
                    >
                      <Mail class="w-4 h-4" />
                    </button>
                    <button
                      v-if="getVendorByTask(task)?.phone"
                      type="button"
                      title="Kirim Pesan ke WhatsApp Vendor"
                      @click="launchWhatsApp(task, getVendorByTask(task)?.phone || '', task.vendor_name, task.total_price)"
                      class="p-2 rounded-xl bg-emerald-50 text-emerald-600 hover:bg-emerald-100 transition border border-emerald-200/60"
                    >
                      <MessageSquare class="w-4 h-4" />
                    </button>
                    <a
                      v-if="task.design_url"
                      :href="task.design_url"
                      target="_blank"
                      title="Buka File Desain Siap Cetak"
                      class="p-2 rounded-xl bg-slate-50 text-slate-600 hover:bg-slate-100 transition border border-slate-200"
                    >
                      <ExternalLink class="w-4 h-4" />
                    </a>
                  </div>
                </template>

                <!-- Tab 3: Tagihan Belum Lunas (Tombol: TANDAI LUNAS) -->
                <template v-else-if="activeTab === 'unpaid_bills'">
                  <div class="flex items-center justify-center gap-2">
                    <button
                      v-if="canManagePurchasing"
                      @click="openSinglePaymentModal(task)"
                      title="Proses pelunasan pembayaran tagihan spanduk ini"
                      class="inline-flex items-center gap-1.5 bg-emerald-600 hover:bg-emerald-700 text-white px-3.5 py-2 rounded-xl text-xs font-extrabold transition-all shadow-md shadow-emerald-600/20"
                    >
                      <DollarSign class="w-3.5 h-3.5" /> Tandai Lunas
                    </button>
                    <a
                      v-if="task.design_url"
                      :href="task.design_url"
                      target="_blank"
                      title="Buka File Desain"
                      class="p-2 rounded-xl bg-slate-50 text-slate-600 hover:bg-slate-100 transition border border-slate-200"
                    >
                      <ExternalLink class="w-4 h-4" />
                    </a>
                  </div>
                </template>

                <!-- Tab 4: Riwayat Lunas -->
                <template v-else-if="activeTab === 'paid_history'">
                  <div class="flex items-center justify-center gap-2">
                    <span class="inline-flex items-center gap-1 px-3 py-1 rounded-full text-xs font-black bg-emerald-50 text-emerald-700 border border-emerald-200">
                      <CheckCircle2 class="w-3.5 h-3.5 text-emerald-600" /> Sudah Lunas
                    </span>
                    <button
                      v-if="isSuperAdmin || isPurchasingAdmin"
                      @click="markAsUnpaid(task)"
                      title="Batalkan Lunas (Kembalikan ke Tagihan Belum Lunas)"
                      class="p-1.5 rounded-lg bg-slate-100 text-slate-400 hover:text-rose-600 hover:bg-rose-50 transition border border-slate-200"
                    >
                      <RotateCcw class="w-3.5 h-3.5" />
                    </button>
                  </div>
                </template>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Floating Bulk Action Bar -->
    <div
      v-if="selectedBannerIds.length > 0 && canManagePurchasing && activeTab !== 'paid_history'"
      class="fixed bottom-6 left-1/2 -translate-x-1/2 z-40 bg-slate-900 text-white px-6 py-3.5 rounded-2xl shadow-2xl flex items-center gap-4 border border-slate-700 animate-in fade-in slide-in-from-bottom-4 duration-300"
    >
      <div class="flex items-center gap-2.5">
        <span class="w-2.5 h-2.5 rounded-full bg-emerald-400 animate-pulse"></span>
        <span class="text-sm font-black tracking-wide">{{ selectedBannerIds.length }} Banner Dipilih</span>
      </div>

      <div class="h-5 w-[1px] bg-slate-700"></div>

      <!-- Action Button based on Tab -->
      <template v-if="activeTab === 'purchasing_queue'">
        <button
          @click="openBulkModal"
          class="bg-primary hover:bg-primary/90 text-white px-4 py-2 rounded-xl text-xs font-black transition flex items-center gap-2 shadow-lg shadow-primary/30"
        >
          <Truck class="w-4 h-4" /> Kirim ke Vendor (Bulk)
        </button>
      </template>
      <template v-else-if="activeTab === 'sent_to_vendor'">
        <button
          @click="openBulkEmailModal"
          class="bg-violet-600 hover:bg-violet-700 text-white px-4 py-2 rounded-xl text-xs font-black transition flex items-center gap-2 shadow-lg shadow-violet-600/30"
          title="Kirim email desain massal ke vendor via SMTP / Gmail"
        >
          <Mail class="w-4 h-4" /> Kirim Desain ke Vendor (Bulk Email)
        </button>
        <button
          @click="autoCalculateSelectedPrices"
          class="bg-indigo-600 hover:bg-indigo-700 text-white px-4 py-2 rounded-xl text-xs font-black transition flex items-center gap-2 shadow-lg shadow-indigo-600/30"
          title="Hitung otomatis biaya cetak dari tarif master vendor untuk banner terpilih"
        >
          <Calculator class="w-4 h-4" /> Hitung Biaya Otomatis
        </button>
        <button
          @click="submitBulkMarkAsFinishedPrinting"
          class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-xl text-xs font-black transition flex items-center gap-2 shadow-lg shadow-blue-600/30"
        >
          <Printer class="w-4 h-4" /> Tandai Selesai Cetak (Bulk)
        </button>
      </template>
      <template v-else-if="activeTab === 'unpaid_bills'">
        <button
          @click="openBulkPaymentModal"
          class="bg-emerald-600 hover:bg-emerald-700 text-white px-4 py-2 rounded-xl text-xs font-black transition flex items-center gap-2 shadow-lg shadow-emerald-600/30"
        >
          <DollarSign class="w-4 h-4" /> Bayar & Lunasi Tagihan Terpilih (Bulk Akhir Bulan)
        </button>
      </template>

      <button
        @click="clearSelection"
        class="p-1.5 hover:bg-slate-800 rounded-lg text-slate-400 hover:text-white transition"
        title="Batalkan Pilihan"
      >
        <X class="w-4 h-4" />
      </button>
    </div>

    <!-- Single Vendor Modal (Kirim ke Percetakan) -->
    <div v-if="showVendorModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div
        class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
        @click="!isSubmitting && (showVendorModal = false)"
      ></div>
      <div
        class="bg-white rounded-[32px] w-full max-w-lg relative z-10 shadow-2xl p-8 animate-in fade-in zoom-in duration-300 max-h-[90vh] overflow-y-auto"
      >
        <div class="flex justify-between items-center mb-6">
          <h3 class="text-xl font-black text-slate-900 flex items-center gap-2">
            <Truck class="w-6 h-6 text-primary" /> Kirim ke Percetakan
          </h3>
          <button
            v-if="!isSubmitting"
            @click="showVendorModal = false"
            class="p-2 hover:bg-slate-100 rounded-xl text-slate-400"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <p class="text-sm font-bold text-slate-500 mb-6">
          Toko: <span class="text-primary">{{ selectedTask?.shop_name }}</span>
          <span v-if="selectedTask?.type" class="ml-2 px-2 py-0.5 text-xs rounded-md bg-slate-100 font-semibold text-slate-600">
            {{ selectedTask.type }}
          </span>
        </p>

        <div class="space-y-5 mb-8">
          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-2">
              Pilih Vendor Percetakan <span class="text-rose-500">*</span>
            </label>
            <div class="relative">
              <Building class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400 pointer-events-none" />
              <select
                v-model="selectedVendorId"
                @change="onVendorSelectChange"
                class="w-full pl-12 pr-8 py-3 bg-slate-50 border border-slate-200 rounded-xl font-bold text-slate-800 outline-none focus:border-primary focus:bg-white transition text-sm cursor-pointer"
              >
                <option value="">-- Pilih dari Master Vendor --</option>
                <option
                  v-for="vendor in vendorOptions"
                  :key="vendor.id"
                  :value="vendor.id"
                >
                  {{ vendor.name }} {{ vendor.email ? `• ${vendor.email}` : '' }} {{ vendor.price_per_meter > 0 ? `(Rp ${vendor.price_per_meter.toLocaleString('id-ID')}/m²)` : '' }}
                </option>
              </select>
            </div>
            <p v-if="vendorOptions.length === 0" class="text-xs text-amber-600 font-medium mt-1.5 flex items-center gap-1">
              <AlertTriangle class="w-3.5 h-3.5" /> Belum ada vendor di master. Anda dapat menginput nama vendor manual di bawah.
            </p>
          </div>

          <!-- Fallback Nama Vendor Manual -->
          <div v-if="!selectedVendorId">
            <label class="block text-xs font-bold text-slate-500 uppercase tracking-wider mb-2">
              Atau Input Nama Vendor Manual
            </label>
            <input
              type="text"
              v-model="vendorName"
              placeholder="Contoh: CV Sentosa Print"
              class="w-full px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-xl font-medium text-slate-700 outline-none focus:border-primary focus:bg-white transition text-sm"
            />
          </div>

          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-2">
              Email Vendor Percetakan
            </label>
            <div class="relative">
              <Mail class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
              <input
                type="email"
                v-model="vendorEmail"
                placeholder="Misal: order@percetakan.com"
                class="w-full pl-12 pr-4 py-3 bg-slate-50 border border-slate-200 rounded-xl font-medium text-slate-800 outline-none focus:border-primary focus:bg-white transition text-sm"
              />
            </div>
            <p class="text-[11px] text-slate-400 mt-1">Alamat email vendor tujuan SPK & link desain.</p>
          </div>

          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-2">
              No. WhatsApp Vendor (Opsional)
            </label>
            <div class="relative">
              <MessageSquare class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
              <input
                type="text"
                v-model="vendorPhone"
                placeholder="Misal: 081234567890"
                class="w-full pl-12 pr-4 py-3 bg-slate-50 border border-slate-200 rounded-xl font-medium text-slate-800 outline-none focus:border-primary focus:bg-white transition text-sm"
              />
            </div>
          </div>

          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-2">
              Total Biaya Cetak (Rp) <span class="text-rose-500">*</span>
            </label>
            <div class="relative">
              <Tag class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
              <input
                type="number"
                v-model="vendorPrice"
                placeholder="Misal: 50000"
                class="w-full pl-12 pr-4 py-3 bg-slate-50 border border-slate-200 rounded-xl font-bold text-slate-800 outline-none focus:border-primary focus:bg-white transition text-sm"
              />
            </div>
            <p class="text-[11px] text-slate-400 mt-1">Biaya cetak yang harus dibayar ke vendor percetakan.</p>
          </div>

          <!-- Link Desain Preview Info -->
          <div class="p-3.5 bg-blue-50/70 border border-blue-200/80 rounded-2xl flex items-start gap-3">
            <Sparkles class="w-5 h-5 text-blue-600 shrink-0 mt-0.5" />
            <div class="text-xs text-blue-900 space-y-1">
              <p class="font-bold">Link Desain Otomatis Terlampir:</p>
              <p class="text-blue-700 font-mono text-[11px] break-all">
                {{ selectedTask?.design_url || 'Link desain akan disertakan otomatis' }}
              </p>
            </div>
          </div>

          <!-- Opsi Otomatis Buka Gmail -->
          <div class="flex items-center gap-3 p-3 bg-slate-50 border border-slate-200 rounded-xl">
            <input
              type="checkbox"
              id="autoOpenGmailCheck"
              v-model="autoOpenGmail"
              class="w-4 h-4 rounded text-primary focus:ring-primary border-slate-300 cursor-pointer"
            />
            <label for="autoOpenGmailCheck" class="text-xs font-bold text-slate-700 cursor-pointer select-none">
              Buka draft pesan di Gmail Web secara otomatis setelah simpan
            </label>
          </div>
        </div>

        <div class="space-y-2">
          <button
            @click="submitToVendor"
            :disabled="(!selectedVendorId && !vendorName) || !vendorPrice || isSubmitting"
            class="w-full bg-primary hover:bg-primary/90 disabled:bg-slate-200 disabled:text-slate-400 text-white py-4 rounded-2xl font-black transition-all shadow-lg shadow-primary/20 flex items-center justify-center gap-2"
          >
            <RefreshCw v-if="isSubmitting" class="w-5 h-5 animate-spin" />
            <Mail v-else-if="autoOpenGmail" class="w-4 h-4" />
            {{ isSubmitting ? 'Menyimpan...' : (autoOpenGmail ? 'Simpan & Buka Gmail' : 'Konfirmasi & Kirim ke Vendor') }}
          </button>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
            <button
              type="button"
              @click="openSmtpConfirmModal(selectedTask, vendorEmail, vendorName, vendorPrice)"
              class="w-full bg-indigo-600 hover:bg-indigo-700 text-white py-3 rounded-2xl font-bold transition-all shadow-md shadow-indigo-600/20 flex items-center justify-center gap-2 text-sm"
              title="Kirim SPK langsung via Server SMTP dengan preview & konfirmasi"
            >
              <SendHorizontal class="w-4 h-4" />
              Kirim via SMTP Server
            </button>
            <button
              type="button"
              @click="launchGmail(selectedTask, vendorEmail, vendorName, vendorPrice)"
              class="w-full bg-slate-100 hover:bg-slate-200 text-slate-700 py-3 rounded-2xl font-bold transition-all flex items-center justify-center gap-2 text-sm"
              title="Buka Draft Pesan di Gmail Web"
            >
              <Mail class="w-4 h-4 text-rose-500" />
              Buka di Gmail Web
            </button>
          </div>

          <div class="flex items-center gap-2">
            <button
              type="button"
              @click="copySpkFormat(selectedTask, vendorName, vendorPrice)"
              class="flex-1 bg-slate-100 hover:bg-slate-200 text-slate-700 py-3 rounded-2xl font-bold transition-all flex items-center justify-center gap-2 text-sm"
              title="Salin draft format SPK ke clipboard"
            >
              <Copy class="w-4 h-4" />
              Salin Format SPK
            </button>
            <button
              v-if="vendorPhone"
              type="button"
              @click="launchWhatsApp(selectedTask, vendorPhone, vendorName, vendorPrice)"
              class="flex-1 bg-emerald-600 hover:bg-emerald-700 text-white py-3 rounded-2xl font-bold transition-all shadow-md shadow-emerald-600/20 flex items-center justify-center gap-2 text-sm"
            >
              <MessageSquare class="w-4 h-4" />
              Kirim via WhatsApp
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Bulk Vendor Modal (Kirim Bulk ke Percetakan) -->
    <div v-if="showBulkModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div
        class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
        @click="!isBulkSubmitting && (showBulkModal = false)"
      ></div>
      <div
        class="bg-white rounded-[32px] w-full max-w-2xl relative z-10 shadow-2xl p-6 sm:p-8 animate-in fade-in zoom-in duration-300 max-h-[90vh] flex flex-col"
      >
        <div class="flex justify-between items-center pb-4 border-b border-slate-100 shrink-0">
          <div>
            <h3 class="text-xl font-black text-slate-900 flex items-center gap-2">
              <Truck class="w-6 h-6 text-primary" /> Kirim Bulk ke Percetakan
            </h3>
            <p class="text-xs font-semibold text-slate-500 mt-1">
              Menugaskan <span class="text-primary font-bold">{{ bulkItems.length }} spanduk</span> sekaligus ke satu vendor yang sama.
            </p>
          </div>
          <button
            v-if="!isBulkSubmitting"
            @click="showBulkModal = false"
            class="p-2 hover:bg-slate-100 rounded-xl text-slate-400 transition"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <div class="space-y-6 py-6 overflow-y-auto pr-1">
          <!-- Vendor Selection -->
          <div class="bg-slate-50 p-5 rounded-2xl border border-slate-200 space-y-4">
            <div>
              <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-2">
                Pilih Vendor Percetakan Tujuan <span class="text-rose-500">*</span>
              </label>
              <div class="relative">
                <Building class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400 pointer-events-none" />
                <select
                  v-model="bulkSelectedVendorId"
                  @change="onBulkVendorSelectChange"
                  class="w-full pl-12 pr-8 py-3 bg-white border border-slate-200 rounded-xl font-bold text-slate-800 outline-none focus:border-primary transition text-sm cursor-pointer"
                >
                  <option value="">-- Pilih dari Master Vendor --</option>
                  <option
                    v-for="vendor in vendorOptions"
                    :key="vendor.id"
                    :value="vendor.id"
                  >
                    {{ vendor.name }} {{ vendor.email ? `• ${vendor.email}` : '' }} {{ vendor.price_per_meter > 0 ? `(Rp ${vendor.price_per_meter.toLocaleString('id-ID')}/m²)` : '' }}
                  </option>
                </select>
              </div>
            </div>

            <!-- Fallback Nama Vendor Manual -->
            <div v-if="!bulkSelectedVendorId">
              <label class="block text-xs font-bold text-slate-500 uppercase tracking-wider mb-1.5">
                Atau Input Nama Vendor Manual
              </label>
              <input
                type="text"
                v-model="bulkVendorName"
                placeholder="Contoh: CV Sentosa Print"
                class="w-full px-4 py-2.5 bg-white border border-slate-200 rounded-xl font-medium text-slate-700 outline-none focus:border-primary transition text-sm"
              />
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div>
                <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
                  Email Vendor (Tujuan SPK)
                </label>
                <div class="relative">
                  <Mail class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
                  <input
                    type="email"
                    v-model="bulkVendorEmail"
                    placeholder="order@percetakan.com"
                    class="w-full pl-10 pr-3 py-2 bg-white border border-slate-200 rounded-xl font-medium text-slate-800 outline-none focus:border-primary transition text-xs"
                  />
                </div>
              </div>

              <div>
                <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
                  WhatsApp Vendor (Opsional)
                </label>
                <div class="relative">
                  <MessageSquare class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
                  <input
                    type="text"
                    v-model="bulkVendorPhone"
                    placeholder="081234567890"
                    class="w-full pl-10 pr-3 py-2 bg-white border border-slate-200 rounded-xl font-medium text-slate-800 outline-none focus:border-primary transition text-xs"
                  />
                </div>
              </div>
            </div>
          </div>

          <!-- Items Breakdown Table -->
          <div>
            <div class="flex items-center justify-between mb-2">
              <label class="text-xs font-bold text-slate-700 uppercase tracking-wider flex items-center gap-1.5">
                <ListChecks class="w-4 h-4 text-primary" />
                Rincian Biaya per Banner ({{ bulkItems.length }} Item)
              </label>
              <span class="text-[11px] font-semibold text-slate-400">
                Tarif otomatis terhitung jika vendor memiliki harga/m²
              </span>
            </div>

            <div class="border border-slate-200 rounded-2xl overflow-hidden bg-white">
              <div class="max-h-60 overflow-y-auto">
                <table class="w-full text-left text-xs border-collapse">
                  <thead class="bg-slate-50 sticky top-0 z-10 border-b border-slate-200 text-slate-500 font-bold uppercase tracking-wider text-[10px]">
                    <tr>
                      <th class="p-3">No</th>
                      <th class="p-3">Nama Toko & Spesifikasi</th>
                      <th class="p-3 text-center">Luas</th>
                      <th class="p-3 w-40">Biaya Cetak (Rp)</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-slate-100">
                    <tr v-for="(item, idx) in bulkItems" :key="item.id" class="hover:bg-slate-50">
                      <td class="p-3 font-bold text-slate-400 text-center w-8">{{ idx + 1 }}</td>
                      <td class="p-3">
                        <p class="font-bold text-slate-800">{{ item.shop_name }}</p>
                        <p v-if="item.shop_address" class="text-[11px] text-slate-500 line-clamp-1 mt-0.5" :title="item.shop_address">
                          {{ item.shop_address }}
                        </p>
                        <p class="text-[11px] text-slate-400 mt-0.5">{{ formatBannerDimension(item.size_length, item.size_width, 'dual') }} • {{ item.raw_type || 'Spanduk' }}</p>
                      </td>
                      <td class="p-3 text-center whitespace-nowrap font-semibold text-slate-600">
                        {{ item.area }} m²
                      </td>
                      <td class="p-3">
                        <input
                          type="number"
                          v-model.number="item.price"
                          min="0"
                          placeholder="Rp 0"
                          class="w-full px-3 py-1.5 bg-slate-50 border border-slate-200 rounded-lg font-bold text-slate-800 outline-none focus:bg-white focus:border-primary text-xs"
                        />
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>

          <!-- Total Summary Card -->
          <div class="grid grid-cols-3 gap-3 bg-slate-900 text-white p-4 rounded-2xl">
            <div>
              <p class="text-[10px] font-bold uppercase tracking-wider text-slate-400">Total Banner</p>
              <p class="text-base font-black mt-0.5">{{ bulkItems.length }} Pcs</p>
            </div>
            <div>
              <p class="text-[10px] font-bold uppercase tracking-wider text-slate-400">Total Luas</p>
              <p class="text-base font-black mt-0.5 text-blue-400">{{ bulkTotalArea }} m²</p>
            </div>
            <div>
              <p class="text-[10px] font-bold uppercase tracking-wider text-slate-400">Total Biaya Cetak</p>
              <p class="text-base font-black mt-0.5 text-emerald-400">{{ formatCurrency(bulkTotalPrice) }}</p>
            </div>
          </div>

          <!-- Catatan Tambahan untuk Vendor -->
          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5 flex items-center justify-between">
              <span>Catatan Tambahan untuk Vendor (Opsional)</span>
              <span class="text-[10px] text-slate-400 font-normal">Akan dicantumkan pada email SPK</span>
            </label>
            <textarea
              v-model="bulkCustomNotes"
              rows="2"
              placeholder="Contoh: Mohon diprioritaskan selesai lusa, bahan flexi 340gr..."
              class="w-full px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-xl font-medium text-slate-800 outline-none focus:bg-white focus:border-primary transition text-xs resize-none"
            ></textarea>
          </div>

          <!-- Opsi Pengiriman Email via SMTP Server -->
          <div class="p-4 bg-blue-50/70 border border-blue-200/80 rounded-2xl space-y-2">
            <div class="flex items-center gap-3">
              <input
                type="checkbox"
                id="bulkSendViaSmtpCheck"
                v-model="bulkSendViaSmtp"
                :disabled="!bulkVendorEmail"
                class="w-4 h-4 rounded text-primary focus:ring-primary border-slate-300 cursor-pointer disabled:opacity-50"
              />
              <label for="bulkSendViaSmtpCheck" class="text-xs font-bold text-slate-800 cursor-pointer select-none flex items-center gap-2">
                <SendHorizontal class="w-4 h-4 text-blue-600" />
                Kirim SPK gabungan langsung via Server SMTP ke Email Vendor
              </label>
            </div>
            <p class="text-[11px] text-slate-600 pl-7 leading-relaxed">
              Email SPK berisi tabel rincian lengkap seluruh <span class="font-bold text-slate-800">{{ bulkItems.length }} banner</span> dan tautan unduh desain akan otomatis terkirim langsung dari server tanpa membuka aplikasi/tab Gmail di browser.
            </p>
            <p v-if="!bulkVendorEmail" class="text-[11px] text-amber-700 font-semibold pl-7 flex items-center gap-1.5">
              <AlertTriangle class="w-3.5 h-3.5 shrink-0" />
              Isi Email Vendor di atas jika ingin mengirimkan SPK langsung via SMTP.
            </p>
          </div>
        </div>

        <!-- Modal Footer Actions -->
        <div class="pt-4 border-t border-slate-100 flex flex-col sm:flex-row gap-3 shrink-0">
          <button
            @click="submitBulkToVendor"
            :disabled="(!bulkSelectedVendorId && !bulkVendorName) || isBulkSubmitting || (bulkSendViaSmtp && !bulkVendorEmail)"
            class="flex-1 bg-primary hover:bg-primary/90 disabled:bg-slate-200 disabled:text-slate-400 text-white py-3.5 rounded-2xl font-black transition-all shadow-lg shadow-primary/20 flex items-center justify-center gap-2 text-sm"
          >
            <RefreshCw v-if="isBulkSubmitting" class="w-4 h-4 animate-spin" />
            <SendHorizontal v-else-if="bulkSendViaSmtp" class="w-4 h-4" />
            <Truck v-else class="w-4 h-4" />
            {{ isBulkSubmitting ? 'Memproses Bulk...' : (bulkSendViaSmtp ? 'Simpan & Kirim SPK via SMTP' : 'Simpan & Kirim ke Vendor') }}
          </button>

          <button
            type="button"
            @click="copyBulkSpkFormat(bulkItems, bulkVendorName)"
            class="bg-slate-100 hover:bg-slate-200 text-slate-700 px-4 py-3.5 rounded-2xl font-bold transition-all flex items-center justify-center gap-2 text-sm shrink-0"
            title="Salin draft SPK gabungan ke clipboard"
          >
            <Copy class="w-4 h-4" />
            Salin SPK
          </button>

          <button
            v-if="bulkVendorEmail"
            type="button"
            @click="launchBulkGmail(bulkItems, bulkVendorEmail, bulkVendorName)"
            class="bg-slate-100 hover:bg-slate-200 text-slate-700 px-4 py-3.5 rounded-2xl font-bold transition-all flex items-center justify-center gap-2 text-sm shrink-0"
            title="Buka Draft Pesan Manual di Gmail Web (Opsional)"
          >
            <Mail class="w-4 h-4 text-rose-500" />
            Buka Gmail Web
          </button>

          <button
            v-if="bulkVendorPhone"
            type="button"
            @click="launchBulkWa(bulkItems, bulkVendorPhone, bulkVendorName)"
            class="bg-emerald-600 hover:bg-emerald-700 text-white px-5 py-3.5 rounded-2xl font-bold transition-all shadow-md shadow-emerald-600/20 flex items-center justify-center gap-2 text-sm shrink-0"
          >
            <MessageSquare class="w-4 h-4" />
            WhatsApp
          </button>
        </div>
      </div>
    </div>

    <!-- Payment Modal (Lunasi Tagihan Satuan / Bulk Akhir Bulan) -->
    <div v-if="showPaymentModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div
        class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
        @click="!isPaymentSubmitting && (showPaymentModal = false)"
      ></div>
      <div
        class="bg-white rounded-[32px] w-full max-w-lg relative z-10 shadow-2xl p-6 sm:p-8 animate-in fade-in zoom-in duration-300 max-h-[90vh] flex flex-col"
      >
        <div class="flex justify-between items-center pb-4 border-b border-slate-100 shrink-0">
          <div>
            <h3 class="text-xl font-black text-slate-900 flex items-center gap-2">
              <DollarSign class="w-6 h-6 text-emerald-600" />
              Konfirmasi Pelunasan Tagihan
            </h3>
            <p class="text-xs font-semibold text-slate-500 mt-1">
              Pelunasan pembayaran untuk <span class="font-bold text-slate-800">{{ paymentModalTasks.length }} pesanan spanduk</span>.
            </p>
          </div>
          <button
            v-if="!isPaymentSubmitting"
            @click="showPaymentModal = false"
            class="p-2 hover:bg-slate-100 rounded-xl text-slate-400 transition"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <div class="space-y-5 py-6 overflow-y-auto pr-1">
          <!-- Total Card -->
          <div class="bg-emerald-50/80 border border-emerald-200/80 rounded-2xl p-5">
            <p class="text-xs font-bold uppercase tracking-wider text-emerald-700">Total Nominal Pelunasan</p>
            <p class="text-2xl font-black text-emerald-900 mt-1">{{ formatCurrency(paymentModalTotalCost) }}</p>
            <p class="text-[11px] font-medium text-emerald-600 mt-1">
              {{ paymentModalTasks.length }} item tagihan percetakan
            </p>
          </div>

          <!-- Rincian Singkat Toko & Vendor -->
          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-2">
              Rincian Item Tagihan
            </label>
            <div class="border border-slate-200 rounded-xl overflow-hidden max-h-40 overflow-y-auto bg-slate-50/50">
              <div
                v-for="(item, idx) in paymentModalTasks"
                :key="item.id"
                class="p-3 text-xs border-b border-slate-100 last:border-b-0 flex items-center justify-between gap-3"
              >
                <div>
                  <p class="font-bold text-slate-800">{{ idx + 1 }}. {{ item.shop_name }}</p>
                  <p class="text-[10px] text-slate-500">Vendor: {{ item.vendor_name }} • {{ item.type }}</p>
                </div>
                <div class="font-black text-slate-900 whitespace-nowrap">
                  {{ formatCurrency(item.total_price) }}
                </div>
              </div>
            </div>
          </div>

          <!-- No. Referensi Transfer / Pembayaran -->
          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-2">
              No. Referensi / Bukti Transfer (Opsional)
            </label>
            <div class="relative">
              <Receipt class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
              <input
                type="text"
                v-model="paymentReference"
                placeholder="Misal: TRF-BCA-20260908 / Nota #0881"
                class="w-full pl-11 pr-4 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold text-slate-800 outline-none focus:border-emerald-500 focus:bg-white transition"
              />
            </div>
            <p class="text-[10px] text-slate-400 mt-1">No. ref bank, bukti transaksi, atau nomor nota rekapitulasi vendor.</p>
          </div>

          <!-- Catatan Tambahan -->
          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-2">
              Catatan Pembayaran (Opsional)
            </label>
            <textarea
              v-model="paymentNotes"
              rows="2"
              placeholder="Misal: Pembayaran rekap tagihan bulan September 2026..."
              class="w-full px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-medium text-slate-800 outline-none focus:border-emerald-500 focus:bg-white transition"
            ></textarea>
          </div>
        </div>

        <!-- Footer -->
        <div class="pt-4 border-t border-slate-100 flex gap-3 shrink-0">
          <button
            @click="submitPayment"
            :disabled="isPaymentSubmitting"
            class="flex-1 bg-emerald-600 hover:bg-emerald-700 disabled:bg-slate-200 disabled:text-slate-400 text-white py-3.5 rounded-2xl font-black transition-all shadow-lg shadow-emerald-600/20 flex items-center justify-center gap-2 text-sm"
          >
            <RefreshCw v-if="isPaymentSubmitting" class="w-4 h-4 animate-spin" />
            <CheckCircle2 v-else class="w-4 h-4" />
            {{ isPaymentSubmitting ? 'Memproses...' : 'Konfirmasi Lunas' }}
          </button>
          <button
            type="button"
            v-if="!isPaymentSubmitting"
            @click="showPaymentModal = false"
            class="px-5 py-3.5 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-2xl font-bold transition-all text-sm"
          >
            Batal
          </button>
        </div>
      </div>
    </div>

    <!-- Modal Edit Biaya & Vendor (Untuk Tab Di Percetakan / Tagihan) -->
    <div v-if="showEditPriceModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div
        class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
        @click="!isEditPriceSubmitting && (showEditPriceModal = false)"
      ></div>
      <div
        class="bg-white rounded-[32px] w-full max-w-lg relative z-10 shadow-2xl p-8 animate-in fade-in zoom-in duration-300 max-h-[90vh] overflow-y-auto"
      >
        <div class="flex justify-between items-center mb-6">
          <h3 class="text-xl font-black text-slate-900 flex items-center gap-2">
            <Pencil class="w-6 h-6 text-primary" /> Edit Biaya & Vendor Percetakan
          </h3>
          <button
            v-if="!isEditPriceSubmitting"
            @click="showEditPriceModal = false"
            class="p-2 hover:bg-slate-100 rounded-xl text-slate-400"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <div class="p-4 bg-slate-50 border border-slate-200 rounded-2xl mb-6 space-y-1">
          <p class="text-xs font-bold uppercase tracking-wider text-slate-400">Informasi Banner</p>
          <p class="text-sm font-black text-slate-800">{{ editPriceTask?.shop_name }}</p>
          <p class="text-xs text-slate-600 font-medium">
            {{ editPriceTask?.type }} • Luas: <span class="font-bold text-slate-800">{{ getBannerAreaM2(editPriceTask) }} m²</span> (Qty: {{ editPriceTask?.qty || 1 }})
          </p>
        </div>

        <div class="space-y-5 mb-8">
          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-2">
              Pilih Vendor Percetakan <span class="text-rose-500">*</span>
            </label>
            <div class="relative">
              <Building class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400 pointer-events-none" />
              <select
                v-model="editPriceVendorId"
                @change="onEditPriceVendorSelectChange"
                class="w-full pl-12 pr-8 py-3 bg-slate-50 border border-slate-200 rounded-xl font-bold text-slate-800 outline-none focus:border-primary focus:bg-white transition text-sm cursor-pointer"
              >
                <option value="">-- Pilih dari Master Vendor --</option>
                <option
                  v-for="vendor in vendorOptions"
                  :key="vendor.id"
                  :value="vendor.id"
                >
                  {{ vendor.name }} {{ vendor.price_per_meter > 0 ? `(Rp ${vendor.price_per_meter.toLocaleString('id-ID')}/m²)` : '' }}
                </option>
              </select>
            </div>
          </div>

          <!-- Fallback Nama Vendor Manual -->
          <div v-if="!editPriceVendorId">
            <label class="block text-xs font-bold text-slate-500 uppercase tracking-wider mb-2">
              Atau Input Nama Vendor Manual
            </label>
            <input
              type="text"
              v-model="editPriceVendorName"
              placeholder="Contoh: RAJAKU PRINTING"
              class="w-full px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-xl font-medium text-slate-700 outline-none focus:border-primary focus:bg-white transition text-sm"
            />
          </div>

          <!-- Total Biaya Cetak dengan Quick Auto-Calculate -->
          <div>
            <div class="flex items-center justify-between mb-2">
              <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider">
                Total Biaya Cetak (Rp) <span class="text-rose-500">*</span>
              </label>
              <button
                type="button"
                @click="applyAutoCalculationToEdit"
                class="inline-flex items-center gap-1 text-xs font-bold text-indigo-600 hover:text-indigo-800 bg-indigo-50 hover:bg-indigo-100 px-2.5 py-1 rounded-lg transition"
                title="Hitung Luas × Tarif Vendor/m²"
              >
                <Calculator class="w-3.5 h-3.5" /> Hitung Otomatis by Sistem
              </button>
            </div>
            <div class="relative">
              <Tag class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
              <input
                type="number"
                v-model.number="editPriceValue"
                placeholder="Misal: 50000"
                class="w-full pl-12 pr-4 py-3 bg-slate-50 border border-slate-200 rounded-xl font-bold text-slate-800 outline-none focus:border-primary focus:bg-white transition text-sm"
              />
            </div>
            <p v-if="editPriceValue && Number(editPriceValue) > 0" class="text-xs font-bold text-emerald-600 mt-1.5">
              Nominal Terbaca: {{ formatCurrency(Number(editPriceValue)) }}
            </p>
          </div>
        </div>

        <div class="flex gap-3">
          <button
            @click="submitEditPrice"
            :disabled="(!editPriceVendorId && !editPriceVendorName) || editPriceValue === '' || isEditPriceSubmitting"
            class="flex-1 bg-primary hover:bg-primary/90 disabled:bg-slate-200 disabled:text-slate-400 text-white py-3.5 rounded-2xl font-black transition-all shadow-lg shadow-primary/20 flex items-center justify-center gap-2 text-sm"
          >
            <RefreshCw v-if="isEditPriceSubmitting" class="w-4 h-4 animate-spin" />
            <CheckCircle2 v-else class="w-4 h-4" />
            {{ isEditPriceSubmitting ? 'Menyimpan...' : 'Simpan Perubahan Biaya' }}
          </button>
          <button
            type="button"
            v-if="!isEditPriceSubmitting"
            @click="showEditPriceModal = false"
            class="px-5 py-3.5 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-2xl font-bold transition-all text-sm"
          >
            Batal
          </button>
        </div>
      </div>
    </div>

    <!-- Modal Konfirmasi Make Sure Kirim Email via Server SMTP -->
    <div v-if="showSmtpConfirmModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div
        class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
        @click="!isSendingSmtp && (showSmtpConfirmModal = false)"
      ></div>
      <div
        class="bg-white rounded-[32px] w-full max-w-xl relative z-10 shadow-2xl p-6 sm:p-8 animate-in fade-in zoom-in duration-300 max-h-[90vh] flex flex-col"
      >
        <div class="flex justify-between items-center pb-4 border-b border-slate-100 shrink-0">
          <div>
            <h3 class="text-xl font-black text-slate-900 flex items-center gap-2">
              <Server class="w-6 h-6 text-indigo-600" /> Konfirmasi Kirim Email (SMTP)
            </h3>
            <p class="text-xs font-semibold text-slate-500 mt-1">
              Email SPK cetak akan dikirimkan langsung dari server ke vendor percetakan.
            </p>
          </div>
          <button
            @click="!isSendingSmtp && (showSmtpConfirmModal = false)"
            class="p-2 hover:bg-slate-100 rounded-full text-slate-400 hover:text-slate-600 transition"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <div class="space-y-4 py-4 overflow-y-auto flex-1 pr-1">
          <!-- Rincian Pesanan / Task -->
          <div class="p-4 bg-slate-50 border border-slate-200 rounded-2xl space-y-2">
            <div class="flex items-center justify-between">
              <span class="text-xs font-bold text-slate-500">Nama Toko</span>
              <span class="text-xs font-black text-slate-900">{{ smtpModalTask?.shop_name }}</span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-xs font-bold text-slate-500">Ukuran & Jenis</span>
              <span class="text-xs font-bold text-blue-600">
                {{ smtpModalTask?.raw_type || smtpModalTask?.type }} ({{ formatBannerDimension(smtpModalTask?.size_length, smtpModalTask?.size_width, 'dual') }})
              </span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-xs font-bold text-slate-500">Biaya Cetak</span>
              <span class="text-xs font-extrabold text-emerald-600">
                {{ formatCurrency(smtpModalTask?.total_price || vendorPrice || 0) }}
              </span>
            </div>
            <div v-if="smtpModalTask?.design_url" class="pt-2 border-t border-slate-200/60">
              <span class="text-[11px] font-bold text-slate-500 block mb-1">File Desain Terlampir:</span>
              <a
                :href="smtpModalTask.design_url"
                target="_blank"
                class="text-xs font-bold text-primary hover:underline flex items-center gap-1 break-all"
              >
                <ExternalLink class="w-3 h-3 shrink-0" /> {{ smtpModalTask.design_url }}
              </a>
            </div>
          </div>

          <!-- Input Email Penerima Vendor -->
          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-2">
              Email Vendor Penerima <span class="text-rose-500">*</span>
            </label>
            <div class="relative">
              <Mail class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
              <input
                type="email"
                v-model="smtpRecipientEmail"
                placeholder="Contoh: percetakan@gmail.com"
                class="w-full pl-12 pr-4 py-3 bg-slate-50 border border-slate-200 rounded-xl font-bold text-slate-800 outline-none focus:border-indigo-600 focus:bg-white transition text-sm"
              />
            </div>
            <p class="text-[11px] text-slate-400 mt-1">Pastikan alamat email vendor aktif dan dapat menerima pesan.</p>
          </div>

          <!-- Input Subject Email -->
          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-2">
              Subjek Email
            </label>
            <input
              type="text"
              v-model="smtpSubject"
              class="w-full px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-xl font-medium text-slate-800 outline-none focus:border-indigo-600 focus:bg-white transition text-xs"
            />
          </div>

          <!-- Catatan Tambahan (Opsional) -->
          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-2">
              Catatan Tambahan untuk Vendor (Opsional)
            </label>
            <textarea
              v-model="smtpCustomNotes"
              rows="3"
              placeholder="Contoh: Mohon pakai mata ayam di 4 sudut, dikirim sebelum hari Jumat..."
              class="w-full px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-xl font-medium text-slate-800 outline-none focus:border-indigo-600 focus:bg-white transition text-xs"
            ></textarea>
          </div>
        </div>

        <div class="flex gap-3 pt-3 border-t border-slate-100 shrink-0">
          <button
            @click="sendSmtpEmail"
            :disabled="!smtpRecipientEmail || isSendingSmtp"
            class="flex-1 bg-indigo-600 hover:bg-indigo-700 disabled:bg-slate-200 disabled:text-slate-400 text-white py-3.5 rounded-2xl font-black transition-all shadow-lg shadow-indigo-600/20 flex items-center justify-center gap-2 text-sm"
          >
            <RefreshCw v-if="isSendingSmtp" class="w-4 h-4 animate-spin" />
            <Send v-else class="w-4 h-4" />
            {{ isSendingSmtp ? 'Mengirim Email...' : 'Ya, Kirim Email Sekarang' }}
          </button>
          <button
            type="button"
            v-if="!isSendingSmtp"
            @click="showSmtpConfirmModal = false"
            class="px-5 py-3.5 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-2xl font-bold transition-all text-sm"
          >
            Batal
          </button>
        </div>
      </div>
    </div>

    <!-- Modal Bulk Kirim Email Desain ke Vendor (SMTP & Gmail) -->
    <div v-if="showBulkEmailModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div
        class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
        @click="!isSendingBulkEmail && (showBulkEmailModal = false)"
      ></div>
      <div
        class="bg-white rounded-[32px] w-full max-w-2xl relative z-10 shadow-2xl p-6 sm:p-8 animate-in fade-in zoom-in duration-300 max-h-[92vh] flex flex-col"
      >
        <div class="flex justify-between items-center pb-4 border-b border-slate-100 shrink-0">
          <div>
            <h3 class="text-xl font-black text-slate-900 flex items-center gap-2">
              <Mail class="w-6 h-6 text-violet-600" /> Bulk Kirim Email Desain ke Vendor
            </h3>
            <p class="text-xs font-semibold text-slate-500 mt-1">
              Kirim Surat Perintah Kerja (SPK) & link desain cetak untuk <span class="text-violet-600 font-bold">{{ bulkEmailTasks.length }} pesanan spanduk</span>.
            </p>
          </div>
          <button
            v-if="!isSendingBulkEmail"
            @click="showBulkEmailModal = false"
            class="p-2 hover:bg-slate-100 rounded-full text-slate-400 hover:text-slate-600 transition"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <div class="space-y-4 py-4 overflow-y-auto flex-1 pr-1">
          <!-- Warning jika ada banner yang sudah pernah dikirim email -->
          <div
            v-if="alreadySentEmailCount > 0"
            class="p-3.5 bg-amber-50 border border-amber-200 rounded-2xl flex items-start gap-3"
          >
            <AlertTriangle class="w-5 h-5 text-amber-600 shrink-0 mt-0.5" />
            <div class="text-xs text-amber-800">
              <p class="font-bold">Perhatian: Terdeteksi {{ alreadySentEmailCount }} banner sudah pernah dikirimkan email sebelumnya.</p>
              <p class="mt-0.5 text-amber-700">Pastikan tidak terjadi pengiriman ganda kecuali memang ingin mengirimkan revisi/pengingat ke vendor.</p>
            </div>
          </div>

          <!-- Info tips jika pesanan banyak -->
          <div
            v-if="bulkEmailTasks.length >= 20"
            class="p-3.5 bg-indigo-50 border border-indigo-200 rounded-2xl flex items-start gap-3"
          >
            <Info class="w-5 h-5 text-indigo-600 shrink-0 mt-0.5" />
            <div class="text-xs text-indigo-800">
              <p class="font-bold">Tips Pesanan Banyak ({{ bulkEmailTasks.length }} Banner):</p>
              <p class="mt-0.5 text-indigo-700">Untuk pesanan banyak, disarankan menggunakan tombol <strong>Kirim via SMTP Server</strong> agar seluruh rincian dan tabel desain terkirim langsung ke vendor tanpa batasan karakter URL browser.</p>
            </div>
          </div>

          <!-- Vendor Selection & Email Target -->
          <div class="bg-slate-50 p-4 rounded-2xl border border-slate-200 space-y-3">
            <div>
              <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
                Pilih Vendor Percetakan Tujuan <span class="text-rose-500">*</span>
              </label>
              <div class="relative">
                <Building class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 pointer-events-none" />
                <select
                  v-model="bulkEmailSelectedVendorId"
                  @change="onBulkEmailVendorSelectChange"
                  class="w-full pl-10 pr-8 py-2.5 bg-white border border-slate-200 rounded-xl font-bold text-slate-800 outline-none focus:border-violet-600 transition text-xs cursor-pointer"
                >
                  <option value="">-- Pilih dari Master Vendor --</option>
                  <option
                    v-for="vendor in vendorOptions"
                    :key="vendor.id"
                    :value="vendor.id"
                  >
                    {{ vendor.name }} {{ vendor.email ? `• ${vendor.email}` : '' }}
                  </option>
                </select>
              </div>
            </div>

            <!-- Email Vendor Input -->
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div>
                <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
                  Email Vendor Penerima <span class="text-rose-500">*</span>
                </label>
                <div class="relative">
                  <Mail class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
                  <input
                    type="email"
                    v-model="bulkEmailRecipient"
                    placeholder="order@percetakan.com"
                    class="w-full pl-10 pr-3 py-2 bg-white border border-slate-200 rounded-xl font-medium text-slate-800 outline-none focus:border-violet-600 transition text-xs"
                  />
                </div>
              </div>
              <div>
                <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
                  Nama Vendor
                </label>
                <input
                  type="text"
                  v-model="bulkEmailVendorName"
                  placeholder="Contoh: Percetakan Sentosa"
                  class="w-full px-3 py-2 bg-white border border-slate-200 rounded-xl font-medium text-slate-800 outline-none focus:border-violet-600 transition text-xs"
                />
              </div>
            </div>

            <div>
              <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
                Subjek Email
              </label>
              <input
                type="text"
                v-model="bulkEmailSubject"
                class="w-full px-3 py-2 bg-white border border-slate-200 rounded-xl font-medium text-slate-800 outline-none focus:border-violet-600 transition text-xs"
              />
            </div>
          </div>

          <!-- Rincian Daftar Banner yang Dipilih -->
          <div>
            <div class="flex items-center justify-between mb-2">
              <label class="text-xs font-bold text-slate-700 uppercase tracking-wider flex items-center gap-1.5">
                <ListChecks class="w-4 h-4 text-violet-600" />
                Daftar File Desain Siap Cetak ({{ bulkEmailTasks.length }} Banner)
              </label>
              <span class="text-[11px] font-semibold text-slate-500">
                Total Luas: {{ bulkEmailTotalArea }} m² • Biaya: {{ formatCurrency(bulkEmailTotalPrice) }}
              </span>
            </div>

            <div class="border border-slate-200 rounded-2xl overflow-hidden bg-white">
              <div class="max-h-52 overflow-y-auto">
                <table class="w-full text-left text-xs border-collapse">
                  <thead class="bg-slate-50 sticky top-0 z-10 border-b border-slate-200 text-slate-500 font-bold uppercase tracking-wider text-[10px]">
                    <tr>
                      <th class="p-2.5 text-center w-8">No</th>
                      <th class="p-2.5">Nama Toko & Ukuran</th>
                      <th class="p-2.5 text-center">Status Email</th>
                      <th class="p-2.5 text-center">Link Desain</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-slate-100">
                    <tr v-for="(task, idx) in bulkEmailTasks" :key="task.id" class="hover:bg-slate-50">
                      <td class="p-2.5 font-bold text-slate-400 text-center">{{ idx + 1 }}</td>
                      <td class="p-2.5">
                        <p class="font-bold text-slate-800">{{ task.shop_name }}</p>
                        <p v-if="task.shop_address" class="text-[11px] text-slate-500 line-clamp-1 mt-0.5" :title="task.shop_address">
                          {{ task.shop_address }}
                        </p>
                        <p class="text-[11px] text-slate-400 mt-0.5">
                          {{ formatBannerDimension(task.size_length, task.size_width, 'dual') }} (Qty: {{ task.qty || 1 }})
                        </p>
                      </td>
                      <td class="p-2.5 text-center">
                        <span
                          v-if="task.is_vendor_email_sent"
                          class="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-50 text-emerald-700 border border-emerald-200"
                        >
                          <Check class="w-3 h-3 text-emerald-600" /> Pernah Dikirim
                        </span>
                        <span
                          v-else
                          class="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[10px] font-bold bg-slate-100 text-slate-600 border border-slate-200"
                        >
                          Belum Dikirim
                        </span>
                      </td>
                      <td class="p-2.5 text-center">
                        <a
                          v-if="task.design_url"
                          :href="task.design_url"
                          target="_blank"
                          class="inline-flex items-center gap-1 text-[11px] font-bold text-primary hover:underline"
                        >
                          <ExternalLink class="w-3 h-3" /> Desain
                        </a>
                        <span v-else class="text-[11px] text-slate-400">-</span>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>

          <!-- Catatan Tambahan -->
          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
              Catatan Tambahan untuk Vendor (Opsional)
            </label>
            <textarea
              v-model="bulkEmailCustomNotes"
              rows="2"
              placeholder="Contoh: Mohon diproses cetak mulai besok pagi, sertakan label nama toko pada masing-masing spanduk..."
              class="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-medium text-slate-800 outline-none focus:border-violet-600 focus:bg-white transition text-xs"
            ></textarea>
          </div>
        </div>

        <!-- Footer Actions -->
        <div class="pt-4 border-t border-slate-100 flex flex-col gap-2 shrink-0">
          <button
            @click="sendBulkSmtpEmail"
            :disabled="!bulkEmailRecipient || isSendingBulkEmail"
            class="w-full bg-violet-600 hover:bg-violet-700 disabled:bg-slate-200 disabled:text-slate-400 text-white py-3.5 rounded-2xl font-black transition-all shadow-lg shadow-violet-600/20 flex items-center justify-center gap-2 text-sm"
          >
            <RefreshCw v-if="isSendingBulkEmail" class="w-4 h-4 animate-spin" />
            <Send v-else class="w-4 h-4" />
            {{ isSendingBulkEmail ? 'Mengirim Email Bulk...' : 'Kirim Email Sekarang (via SMTP Server)' }}
          </button>

          <div class="grid grid-cols-1 sm:grid-cols-3 gap-2">
            <button
              type="button"
              @click="launchBulkGmail(bulkEmailTasks, bulkEmailRecipient, bulkEmailVendorName)"
              class="bg-slate-100 hover:bg-slate-200 text-slate-700 py-2.5 rounded-xl font-bold transition flex items-center justify-center gap-1.5 text-xs"
              title="Buka Draft Gabungan di Gmail Web"
            >
              <Mail class="w-3.5 h-3.5 text-rose-500" />
              Buka Gmail Web
            </button>
            <button
              type="button"
              @click="copyBulkSpkFormat(bulkEmailTasks, bulkEmailVendorName)"
              class="bg-slate-100 hover:bg-slate-200 text-slate-700 py-2.5 rounded-xl font-bold transition flex items-center justify-center gap-1.5 text-xs"
              title="Salin Rincian SPK ke Clipboard"
            >
              <Copy class="w-3.5 h-3.5" />
              Salin SPK
            </button>
            <button
              type="button"
              @click="markSelectedAsEmailSent"
              class="bg-emerald-50 hover:bg-emerald-100 text-emerald-700 border border-emerald-200/80 py-2.5 rounded-xl font-bold transition flex items-center justify-center gap-1.5 text-xs"
              title="Tandai data terpilih sudah dikirim email secara manual"
            >
              <MailCheck class="w-3.5 h-3.5 text-emerald-600" />
              Tandai Terkirim
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
