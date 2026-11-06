<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import apiClient from '@/api/axios'
import { toast } from 'vue-sonner'
import { useMasterDataStore } from '@/stores/masterData'
import { useAuthStore } from '@/stores/auth'
import { resolvePhotoUrl } from '@/utils/imageUrl'
import SearchSelect, { type SelectOption } from '@/components/SearchSelect.vue'
import StoreAddressHoverCard from '@/components/StoreAddressHoverCard.vue'

const masterStore = useMasterDataStore()
const authStore = useAuthStore()

const canManageBanners = computed(() => {
  const role = authStore.user?.role
  return [
    'superadmin',
    'admin',
    'hr_admin',
    'HR',
    'manager',
    'validator_desain',
    'validator_toko',
    'admin_designer',
    'editor',
    'admin_purchasing',
  ].includes(role)
})
import {
  Megaphone,
  CheckCircle2,
  Clock,
  AlertCircle,
  Image,
  Store,
  Edit2,
  Trash2,
  Users,
  MapPin,
  ExternalLink,
  RefreshCw,
  Search,
  Wifi,
  WifiOff,
  Ruler,
  Package,
  TrendingUp,
  Eye,
  X,
  Download,
  Plus,
  Loader2,
  Sparkles,
  AlertTriangle,
  UploadCloud,
  FileText,
  FileDown,
  Paperclip,
  Calendar,
  UserCheck,
  Building2,
  SlidersHorizontal,
  RotateCcw,
  Filter,
  User,
} from 'lucide-vue-next'
import ExcelJS from 'exceljs'
import { saveAs } from 'file-saver'
import {
  calculateBannerDesignSize,
  formatBannerDimension,
  normalizeDimensionToCm,
} from '@/utils/bannerSize'

// ── Tab ────────────────────────────────────────────────────────────────────
const activeTab = ref<'branding' | 'internal'>('branding')

// ── Branding App (PostgreSQL Super App) ──────────────────────────────────
const isBrandingLoading = ref(false)
const isSearching = ref(false)
const isExporting = ref(false)
const brandingBanners = ref<any[]>([])
const totalBanners = ref(0)
const brandingStats = ref<any>(null)
const brandingSource = ref('')
const brandingOnline = ref(true)
const selectedBranchFilter = computed<string>({
  get: () => masterStore.selectedBrandingBranchId || 'ALL_BRANCHES',
  set: (val: string) => {
    masterStore.setSelectedBrandingBranchId(val)
  },
})
const statusFilter = ref('all')
const searchQuery = ref('')
const salesFilter = ref('all')
const bihunFilter = ref<'all' | 'bihun' | 'non_bihun'>('all')
let searchTimeout: any = null
const monthFilter = ref('all')
const startDate = ref('')
const endDate = ref('')
const datePreset = ref<'all' | 'today' | '7d' | '30d' | 'custom'>('all')

const selectedBranchName = computed(() => {
  return masterStore.selectedBrandingBranchName
})

const getTodayString = (): string => new Date().toISOString().split('T')[0] || ''

const setDatePreset = (preset: 'all' | 'today' | '7d' | '30d') => {
  datePreset.value = preset
  if (preset === 'today') {
    startDate.value = getTodayString()
    endDate.value = getTodayString()
  } else if (preset === '7d') {
    const d = new Date()
    d.setDate(d.getDate() - 7)
    startDate.value = d.toISOString().split('T')[0] || ''
    endDate.value = getTodayString()
  } else if (preset === '30d') {
    const d = new Date()
    d.setDate(d.getDate() - 30)
    startDate.value = d.toISOString().split('T')[0] || ''
    endDate.value = getTodayString()
  } else {
    startDate.value = ''
    endDate.value = ''
  }
}

const previewPhoto = ref<string | null>(null)
const previewTitle = ref('Pratinjau Foto')

const photoUrl = (url?: string | null) => {
  return resolvePhotoUrl(url)
}

const openPhotoPreview = (url?: string | null, title = 'Pratinjau Foto') => {
  const resolved = resolvePhotoUrl(url)
  if (resolved) {
    previewTitle.value = title
    previewPhoto.value = resolved
  }
}

const internalStatusColor = (status: string) => {
  switch (status?.toLowerCase()) {
    case 'pending':
    case 'request':
      return 'bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-400 border-amber-200'
    case 'in_design':
    case 'editing':
      return 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400 border-blue-200'
    case 'approved':
    case 'validation':
      return 'bg-indigo-100 text-indigo-800 dark:bg-indigo-900/30 dark:text-indigo-400 border-indigo-200'
    case 'installed':
    case 'done':
      return 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-400 border-emerald-200'
    case 'rejected':
      return 'bg-rose-100 text-rose-800 dark:bg-rose-900/30 dark:text-rose-400 border-rose-200'
    default:
      return 'bg-slate-100 text-slate-800 dark:bg-slate-800 dark:text-slate-300 border-slate-200'
  }
}

// ── Pagination & View ──────────────────────────────────────────────────────
const currentPage = ref(1)
const itemsPerPage = ref(10)
const viewMode = ref<'table' | 'grid'>('table')

const isOffline = ref(false)

const fetchBrandingStats = async () => {
  try {
    const branchId = selectedBranchFilter.value
    const query = branchId && branchId !== 'ALL_BRANCHES' ? `?branch_id=${branchId}` : ''
    const res = await apiClient.get(`/branding/stats${query}`)
    if (res.data?.data !== undefined) {
      brandingStats.value = res.data.data
    }
  } catch (e) {
    console.error('Failed to fetch branding stats', e)
  }
}

// ── CRUD Banner ────────────────────────────────────────────────────────────
const isEditModalOpen = ref(false)
const isAddModalOpen = ref(false)
const isSaving = ref(false)
const currentBanner = ref({
  id: '',
  shop_id: '',
  shop_name: '',
  sales_name: '',
  type: 'Custom',
  is_bihun: false,
  qty: 1,
  size_length: 0,
  size_width: 0,
  notes: '',
  status: 'request',
  editor_id: '',
  vendor_id: '',
  vendor_name: '',
  total_price: 0,
  spk_date: '',
  operator_id: '',
  operator_name: '',
  operator_reward: 0,
  installed_at: '',
  photo_installed: '',
  design_link: '',
  documentation_image_url: '',
  is_design_approved: false,
  is_store_approved: false,
})

const newBanner = ref({
  shop_id: '',
  shop_name: '',
  sales_name: '',
  type: 'Spanduk',
  is_bihun: false,
  qty: 1,
  size_length: 0,
  size_width: 0,
  notes: '',
  documentation_image_url: '',
})

// ── Attachment State for Edit Modal ──
const editAttachmentFile = ref<File | null>(null)
const editAttachmentPreview = ref<string | null>(null)
const isEditPdfAttachment = ref(false)

const handleEditFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  if (file.size > 15 * 1024 * 1024) {
    toast.error('Ukuran file maksimal 15MB')
    target.value = ''
    return
  }

  editAttachmentFile.value = file
  if (file.type === 'application/pdf' || file.name.toLowerCase().endsWith('.pdf')) {
    isEditPdfAttachment.value = true
    editAttachmentPreview.value = null
  } else {
    isEditPdfAttachment.value = false
    const reader = new FileReader()
    reader.onload = (e) => {
      editAttachmentPreview.value = e.target?.result as string
    }
    reader.readAsDataURL(file)
  }
}

const clearEditAttachmentFile = () => {
  editAttachmentFile.value = null
  editAttachmentPreview.value = null
  isEditPdfAttachment.value = false
}

const removeExistingEditAttachment = () => {
  currentBanner.value.documentation_image_url = ''
  clearEditAttachmentFile()
}

// ── Attachment State for Add Modal ──
const addAttachmentFile = ref<File | null>(null)
const addAttachmentPreview = ref<string | null>(null)
const isAddPdfAttachment = ref(false)

const handleAddFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  if (file.size > 15 * 1024 * 1024) {
    toast.error('Ukuran file maksimal 15MB')
    target.value = ''
    return
  }

  addAttachmentFile.value = file
  if (file.type === 'application/pdf' || file.name.toLowerCase().endsWith('.pdf')) {
    isAddPdfAttachment.value = true
    addAttachmentPreview.value = null
  } else {
    isAddPdfAttachment.value = false
    const reader = new FileReader()
    reader.onload = (e) => {
      addAttachmentPreview.value = e.target?.result as string
    }
    reader.readAsDataURL(file)
  }
}

const clearAddAttachmentFile = () => {
  addAttachmentFile.value = null
  addAttachmentPreview.value = null
  isAddPdfAttachment.value = false
}

const allShops = ref<any[]>([])
const isShopsLoading = ref(false)
const shopsBySales = ref<Record<string, any[]>>({})  // cache per salesman name
const isShopsBySalesLoading = ref(false)
const editors = ref<any[]>([])
const vendors = ref<any[]>([])
const operators = ref<any[]>([])
const brandingSalesAgents = ref<any[]>([])

const fetchBrandingSalesAgents = async () => {
  try {
    const branchId = selectedBranchFilter.value
    const query = branchId && branchId !== 'ALL_BRANCHES' ? `?branch_id=${branchId}` : ''
    const res = await apiClient.get(`/branding/agents${query}`)
    brandingSalesAgents.value = res.data?.data || []
  } catch (e) {
    console.error('Failed to fetch branding agents', e)
  }
}

const fetchEditors = async () => {
  try {
    const res = await apiClient.get('/branding/editors')
    editors.value = res.data?.data || []
  } catch (e) {
    console.error('Failed to fetch editors', e)
  }
}

const fetchVendors = async () => {
  try {
    const res = await apiClient.get('/branding/vendors')
    vendors.value = res.data?.data || []
  } catch (e) {
    console.error('Failed to fetch vendors', e)
  }
}

const fetchOperators = async () => {
  try {
    const res = await apiClient.get('/branding/operators')
    operators.value = res.data?.data || []
  } catch (e) {
    console.error('Failed to fetch operators', e)
  }
}

const fetchAllShops = async () => {
  if (isShopsLoading.value) return
  isShopsLoading.value = true
  try {
    const branchId = selectedBranchFilter.value
    const query = branchId && branchId !== 'ALL_BRANCHES' ? `?branch_id=${branchId}&limit=2000` : '?limit=2000'
    const res = await apiClient.get(`/branding/shops${query}`)
    allShops.value = res.data?.data || []
  } catch (e) {
    console.error('Failed to fetch shops', e)
  } finally {
    isShopsLoading.value = false
  }
}

// Fetch toko secara spesifik berdasarkan nama salesman, dengan cache lokal
const fetchShopsForSales = async (salesName: string) => {
  if (!salesName) return
  // Gunakan cache jika sudah ada
  if (shopsBySales.value[salesName] !== undefined) return
  isShopsBySalesLoading.value = true
  try {
    const encodedName = encodeURIComponent(salesName)
    const res = await apiClient.get(`/branding/shops?sales_name=${encodedName}&limit=500`)
    shopsBySales.value = {
      ...shopsBySales.value,
      [salesName]: res.data?.data || [],
    }
  } catch (e) {
    console.error('Failed to fetch shops for sales', e)
    shopsBySales.value = { ...shopsBySales.value, [salesName]: [] }
  } finally {
    isShopsBySalesLoading.value = false
  }
}

// Convert uniqueSales to SearchSelect options
const salesSelectOptions = computed<SelectOption[]>(() => {
  return uniqueSales.value.map((s) => ({
    value: s.name,
    label: s.name,
    sublabel: s.displayName,
    badge: 'Sales',
    badgeColor: 'bg-blue-100 text-blue-700',
  }))
})

// Build smart shop options for Add modal
const shopSelectOptionsForAdd = computed<SelectOption[]>(() => {
  const salesRaw = newBanner.value.sales_name || ''
  const selectedSales = salesRaw.trim().toLowerCase()
  if (!selectedSales) {
    return []
  }

  const salesSpecific = shopsBySales.value[salesRaw] || []
  const allMatching = allShops.value.filter((s: any) => {
    const shopSalesName = (s.sales_name || '').trim().toLowerCase()
    return shopSalesName === selectedSales
  })

  // Combine unique by shop ID or Name
  const combinedMap = new Map<string, any>()
  for (const s of [...salesSpecific, ...allMatching]) {
    const key = (s.id || s.name || '').toString().toLowerCase()
    if (key && !combinedMap.has(key)) {
      combinedMap.set(key, s)
    }
  }

  const mapped: SelectOption[] = Array.from(combinedMap.values()).map((s: any) => ({
    value: s.name,
    label: s.name,
    sublabel: [s.address_detail, s.sales_area].filter(Boolean).join(' • ') || 'Toko Binaan',
    badge: 'Toko Binaan',
    badgeColor: 'bg-emerald-100 text-emerald-700',
    extra: s,
  }))

  return mapped.sort((a, b) => a.label.localeCompare(b.label))
})

// Build smart shop options for Edit modal
const shopSelectOptionsForEdit = computed<SelectOption[]>(() => {
  const salesRaw = currentBanner.value.sales_name || ''
  const selectedSales = salesRaw.trim().toLowerCase()
  if (!selectedSales) {
    if (currentBanner.value.shop_name) {
      return [{
        value: currentBanner.value.shop_name,
        label: currentBanner.value.shop_name,
        sublabel: 'Toko saat ini pada banner',
        badge: 'Toko Terpasang',
        badgeColor: 'bg-primary/10 text-primary',
      }]
    }
    return []
  }

  const salesSpecific = shopsBySales.value[salesRaw] || []
  const allMatching = allShops.value.filter((s: any) => {
    const shopSalesName = (s.sales_name || '').trim().toLowerCase()
    return shopSalesName === selectedSales
  })

  const combinedMap = new Map<string, any>()
  for (const s of [...salesSpecific, ...allMatching]) {
    const key = (s.id || s.name || '').toString().toLowerCase()
    if (key && !combinedMap.has(key)) {
      combinedMap.set(key, s)
    }
  }

  const mapped: SelectOption[] = Array.from(combinedMap.values()).map((s: any) => ({
    value: s.name,
    label: s.name,
    sublabel: [s.address_detail, s.sales_area].filter(Boolean).join(' • ') || 'Toko Binaan',
    badge: 'Toko Binaan',
    badgeColor: 'bg-emerald-100 text-emerald-700',
    extra: s,
  }))

  // If current banner has a shop not yet in the list, prepend it so historical selection is preserved
  if (currentBanner.value.shop_name && !mapped.some((m) => m.value.toLowerCase() === currentBanner.value.shop_name.toLowerCase())) {
    mapped.unshift({
      value: currentBanner.value.shop_name,
      label: currentBanner.value.shop_name,
      sublabel: 'Toko saat ini pada banner',
      badge: 'Toko Terpasang',
      badgeColor: 'bg-primary/10 text-primary',
    })
  }

  return mapped.sort((a, b) => a.label.localeCompare(b.label))
})

const onSalesChangeAdd = async (salesName?: string) => {
  const name = salesName || ''
  newBanner.value.sales_name = name
  if (name) {
    await fetchShopsForSales(name)
  }
  // Check if current shop_name belongs to new sales, if not, reset it
  if (newBanner.value.shop_name && name) {
    const shops = shopsBySales.value[name] || allShops.value.filter((s: any) => (s.sales_name || '').trim().toLowerCase() === name.trim().toLowerCase())
    const belongsToSales = shops.some((s: any) => 
      s.name.toLowerCase() === newBanner.value.shop_name.toLowerCase()
    )
    if (!belongsToSales) {
      newBanner.value.shop_name = ''
      newBanner.value.shop_id = ''
    }
  } else if (!name) {
    newBanner.value.shop_name = ''
    newBanner.value.shop_id = ''
  }
}

const onSalesChangeEdit = async (salesName?: string) => {
  const name = salesName || ''
  currentBanner.value.sales_name = name
  if (name) {
    await fetchShopsForSales(name)
  }
  // Check if current shop_name belongs to new sales
  if (currentBanner.value.shop_name && name) {
    const shops = shopsBySales.value[name] || allShops.value.filter((s: any) => (s.sales_name || '').trim().toLowerCase() === name.trim().toLowerCase())
    const belongsToSales = shops.some((s: any) => 
      s.name.toLowerCase() === currentBanner.value.shop_name.toLowerCase()
    )
    if (!belongsToSales) {
      currentBanner.value.shop_name = ''
      currentBanner.value.shop_id = ''
    }
  } else if (!name) {
    currentBanner.value.shop_name = ''
    currentBanner.value.shop_id = ''
  }
}

const onSelectShopAdd = (opt?: SelectOption) => {
  newBanner.value.shop_id = opt?.extra?.id || ''
  if (opt?.extra?.sales_name && opt.extra.sales_name !== 'Unknown Sales' && !newBanner.value.sales_name) {
    newBanner.value.sales_name = opt.extra.sales_name
    fetchShopsForSales(opt.extra.sales_name)
  }
}

const onSelectShopEdit = (opt?: SelectOption) => {
  if (opt?.value) {
    currentBanner.value.shop_name = opt.value.trim()
  }
  currentBanner.value.shop_id = opt?.extra?.id || ''
  if (opt?.extra?.sales_name && opt.extra.sales_name !== 'Unknown Sales' && !currentBanner.value.sales_name) {
    currentBanner.value.sales_name = opt.extra.sales_name
    fetchShopsForSales(opt.extra.sales_name)
  }
}

// Banner Unit & Dimension helpers
const newBannerUnit = ref<'cm' | 'm'>('cm')
const currentBannerUnit = ref<'cm' | 'm'>('cm')

const setAddBannerUnit = (unit: 'cm' | 'm') => {
  if (newBannerUnit.value === unit) return
  if (unit === 'm') {
    if (newBanner.value.size_length) newBanner.value.size_length = +(newBanner.value.size_length / 100).toFixed(2)
    if (newBanner.value.size_width) newBanner.value.size_width = +(newBanner.value.size_width / 100).toFixed(2)
  } else {
    if (newBanner.value.size_length) newBanner.value.size_length = Math.round(newBanner.value.size_length * 100)
    if (newBanner.value.size_width) newBanner.value.size_width = Math.round(newBanner.value.size_width * 100)
  }
  newBannerUnit.value = unit
}

const setEditBannerUnit = (unit: 'cm' | 'm') => {
  if (currentBannerUnit.value === unit) return
  if (unit === 'm') {
    if (currentBanner.value.size_length) currentBanner.value.size_length = +(currentBanner.value.size_length / 100).toFixed(2)
    if (currentBanner.value.size_width) currentBanner.value.size_width = +(currentBanner.value.size_width / 100).toFixed(2)
  } else {
    if (currentBanner.value.size_length) currentBanner.value.size_length = Math.round(currentBanner.value.size_length * 100)
    if (currentBanner.value.size_width) currentBanner.value.size_width = Math.round(currentBanner.value.size_width * 100)
  }
  currentBannerUnit.value = unit
}

const openAddModal = () => {
  newBannerUnit.value = 'cm'
  clearAddAttachmentFile()
  newBanner.value = {
    shop_id: '',
    shop_name: '',
    sales_name: '',
    type: 'Custom',
    is_bihun: false,
    qty: 1,
    size_length: 0,
    size_width: 0,
    notes: '',
    documentation_image_url: '',
  }
  if (allShops.value.length === 0) {
    fetchAllShops()
  }
  isAddModalOpen.value = true
}

const handleCreateBanner = async () => {
  if (!newBanner.value.shop_name) {
    toast.error('Nama toko wajib diisi!')
    return
  }
  isSaving.value = true
  try {
    const lengthCm = newBannerUnit.value === 'm'
      ? Math.round((newBanner.value.size_length || 0) * 100)
      : Math.round(newBanner.value.size_length || 0)
    const widthCm = newBannerUnit.value === 'm'
      ? Math.round((newBanner.value.size_width || 0) * 100)
      : Math.round(newBanner.value.size_width || 0)

    if (addAttachmentFile.value) {
      const formData = new FormData()
      formData.append('shop_name', newBanner.value.shop_name)
      formData.append('shop_id', newBanner.value.shop_id)
      formData.append('sales_name', newBanner.value.sales_name || '')
      formData.append('type', newBanner.value.type || 'Custom')
      formData.append('is_bihun', String(newBanner.value.is_bihun))
      formData.append('qty', String(newBanner.value.qty || 1))
      formData.append('size_length', String(lengthCm))
      formData.append('size_width', String(widthCm))
      formData.append('notes', newBanner.value.notes || '')
      formData.append('attachment', addAttachmentFile.value)

      await apiClient.post(`/branding/banners`, formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
    } else {
      const payload = {
        ...newBanner.value,
        size_length: lengthCm,
        size_width: widthCm,
      }
      await apiClient.post(`/branding/banners`, payload)
    }

    toast.success('Banner berhasil ditambahkan!')
    isAddModalOpen.value = false
    clearAddAttachmentFile()
    fetchBrandingBanners()
    fetchBrandingStats()
  } catch (error: any) {
    toast.error(error.response?.data?.error || 'Gagal menambahkan banner')
  } finally {
    isSaving.value = false
  }
}

const openEditModal = (b: any) => {
  currentBannerUnit.value = 'cm'
  clearEditAttachmentFile()
  currentBanner.value = {
    id: b.id,
    shop_id: b.shop_id || '',
    shop_name: b.shop_name || `Shop #${b.shop_id}`,
    sales_name: b.sales_name || '',
    type: b.type || 'Custom',
    is_bihun: b.is_bihun || false,
    qty: b.qty || 1,
    size_length: normalizeDimensionToCm(b.size_length),
    size_width: normalizeDimensionToCm(b.size_width),
    notes: b.notes || '',
    status: (b.status || 'request').toLowerCase(),
    editor_id: b.editor_id || '',
    vendor_id: b.vendor_id || '',
    vendor_name: b.vendor_name || '',
    total_price: b.total_price || 0,
    spk_date: b.spk_date || '',
    operator_id: b.operator_id || '',
    operator_name: b.operator_name || '',
    operator_reward: b.operator_reward || 0,
    installed_at: b.installed_at ? b.installed_at.substring(0, 10) : '',
    photo_installed: b.photo_installed || '',
    design_link: b.design_link || b.design_url || '',
    documentation_image_url: b.documentation_image_url || '',
    is_design_approved: Boolean(b.is_design_approved),
    is_store_approved: Boolean(b.is_store_approved),
  }
  if (currentBanner.value.sales_name) {
    fetchShopsForSales(currentBanner.value.sales_name)
  }
  if (allShops.value.length === 0) {
    fetchAllShops()
  }
  fetchEditors()
  fetchVendors()
  fetchOperators()
  isEditModalOpen.value = true
}

const handleSaveBanner = async () => {
  if (!currentBanner.value.shop_name) {
    toast.error('Nama toko wajib diisi!')
    return
  }
  isSaving.value = true
  try {
    const lengthCm = currentBannerUnit.value === 'm'
      ? Math.round((currentBanner.value.size_length || 0) * 100)
      : Math.round(currentBanner.value.size_length || 0)
    const widthCm = currentBannerUnit.value === 'm'
      ? Math.round((currentBanner.value.size_width || 0) * 100)
      : Math.round(currentBanner.value.size_width || 0)

    const isReturningToDraftOrValidation = ['validation', 'val_design', 'val_store', 'editing', 'request'].includes(currentBanner.value.status)
    const isDesignApproved = isReturningToDraftOrValidation ? false : Boolean(currentBanner.value.is_design_approved)
    const isStoreApproved = isReturningToDraftOrValidation ? false : Boolean(currentBanner.value.is_store_approved)

    const selectedOp = operators.value.find((o) => o.id === currentBanner.value.operator_id)
    const opName = selectedOp ? selectedOp.name : ''

    if (editAttachmentFile.value) {
      const formData = new FormData()
      formData.append('shop_id', currentBanner.value.shop_id || '')
      formData.append('shop_name', currentBanner.value.shop_name)
      formData.append('sales_name', currentBanner.value.sales_name || '')
      formData.append('type', currentBanner.value.type || 'Custom')
      formData.append('is_bihun', String(currentBanner.value.is_bihun))
      formData.append('qty', String(currentBanner.value.qty || 1))
      formData.append('size_length', String(lengthCm))
      formData.append('size_width', String(widthCm))
      formData.append('notes', currentBanner.value.notes || '')
      formData.append('status', currentBanner.value.status || 'request')
      if (currentBanner.value.editor_id !== undefined) formData.append('editor_id', currentBanner.value.editor_id || '')
      if (currentBanner.value.vendor_id !== undefined) formData.append('vendor_id', currentBanner.value.vendor_id || '')
      if (currentBanner.value.vendor_name !== undefined) formData.append('vendor_name', currentBanner.value.vendor_name || '')
      if (currentBanner.value.total_price !== undefined) formData.append('total_price', String(currentBanner.value.total_price || 0))
      if (currentBanner.value.spk_date !== undefined) formData.append('spk_date', currentBanner.value.spk_date || '')
      if (currentBanner.value.operator_id !== undefined) formData.append('operator_id', currentBanner.value.operator_id || '')
      formData.append('operator_name', opName)
      if (currentBanner.value.operator_reward !== undefined) formData.append('operator_reward', String(currentBanner.value.operator_reward || 0))
      if (currentBanner.value.installed_at !== undefined) formData.append('installed_at', currentBanner.value.installed_at || '')
      if (currentBanner.value.photo_installed !== undefined) formData.append('photo_installed', currentBanner.value.photo_installed || '')
      if (currentBanner.value.design_link !== undefined) formData.append('design_link', currentBanner.value.design_link || '')
      formData.append('is_design_approved', String(isDesignApproved))
      formData.append('is_store_approved', String(isStoreApproved))
      formData.append('documentation_image_url', currentBanner.value.documentation_image_url || '')
      formData.append('attachment', editAttachmentFile.value)

      await apiClient.put(`/branding/banners/${currentBanner.value.id}`, formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
    } else {
      const payload = {
        ...currentBanner.value,
        operator_name: opName,
        size_length: lengthCm,
        size_width: widthCm,
        is_design_approved: isDesignApproved,
        is_store_approved: isStoreApproved,
      }
      await apiClient.put(`/branding/banners/${currentBanner.value.id}`, payload)
    }

    toast.success('Data banner berhasil diperbarui!')
    isEditModalOpen.value = false
    clearEditAttachmentFile()
    fetchBrandingBanners()
    fetchBrandingStats()
  } catch (error: any) {
    toast.error(error.response?.data?.error || 'Gagal memperbarui banner')
  } finally {
    isSaving.value = false
  }
}

// ── Delete Confirmation Modal ──────────────────────────────────────────────
const isDeleteModalOpen = ref(false)
const bannerToDelete = ref<any>(null)
const isDeleting = ref(false)

const openDeleteModal = (b: any) => {
  bannerToDelete.value = b
  isDeleteModalOpen.value = true
}

const closeDeleteModal = () => {
  bannerToDelete.value = null
  isDeleteModalOpen.value = false
}

const confirmDeleteBanner = async () => {
  if (!bannerToDelete.value?.id) return
  isDeleting.value = true
  try {
    await apiClient.delete(`/branding/banners/${bannerToDelete.value.id}`)
    toast.success(`Banner untuk toko "${bannerToDelete.value.shop_name || bannerToDelete.value.shop_id}" berhasil dihapus`)
    closeDeleteModal()
    fetchBrandingBanners()
    fetchBrandingStats()
  } catch (error: any) {
    toast.error(error.response?.data?.error || 'Gagal menghapus banner')
  } finally {
    isDeleting.value = false
  }
}

let latestSearchSeq = 0

const fetchBrandingBanners = async (isKeywordSearch: boolean | any = false) => {
  const reqSeq = ++latestSearchSeq
  if (isKeywordSearch === true) {
    isSearching.value = true
  } else {
    isBrandingLoading.value = true
  }

  try {
    const branchId = selectedBranchFilter.value
    const params = new URLSearchParams()
    params.append('page', String(currentPage.value))
    params.append('limit', String(itemsPerPage.value))

    if (statusFilter.value && statusFilter.value !== 'all') {
      params.append('status', statusFilter.value)
    }
    if (branchId && branchId !== 'ALL_BRANCHES') {
      params.append('branch_id', branchId)
    }
    if (searchQuery.value.trim()) {
      params.append('search', searchQuery.value.trim())
    }
    if (salesFilter.value && salesFilter.value !== 'all') {
      params.append('sales_name', salesFilter.value)
    }
    if (monthFilter.value && monthFilter.value !== 'all') {
      params.append('month', monthFilter.value)
    }
    if (startDate.value) {
      params.append('start_date', startDate.value)
    }
    if (endDate.value) {
      params.append('end_date', endDate.value)
    }
    if (bihunFilter.value === 'bihun') {
      params.append('is_bihun', 'true')
    } else if (bihunFilter.value === 'non_bihun') {
      params.append('is_bihun', 'false')
    }

    const res = await apiClient.get(`/branding/banners?${params.toString()}`)
    
    // Ignore stale responses from previous keystrokes
    if (reqSeq !== latestSearchSeq) return

    brandingBanners.value = res.data?.data || []
    totalBanners.value = res.data?.total || 0
    brandingSource.value = res.data?.source || 'live'
    isOffline.value = res.data?.source === 'offline'
    brandingOnline.value = !isOffline.value && res.data?.source !== 'stale_cache'
  } catch (e: any) {
    if (reqSeq !== latestSearchSeq) return
    brandingOnline.value = false
    isOffline.value = true
    toast.error('Gagal memuat data banner Branding: ' + (e.response?.data?.error || e.message))
  } finally {
    if (reqSeq === latestSearchSeq) {
      isBrandingLoading.value = false
      isSearching.value = false
    }
  }
}

const handleSearchInput = () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    currentPage.value = 1
    fetchBrandingBanners(true)
  }, 350)
}

// ── Internal HRIS ──────────────────────────────────────────────────────────
const isInternalLoading = ref(false)
const internalBanners = ref<any[]>([])

const fetchInternalBanners = async () => {
  isInternalLoading.value = true
  try {
    const res = await apiClient.get('/banner-orders')
    internalBanners.value = (res.data?.data || []).map((b: any) => ({
      id: b.id,
      title: b.campaign_name || 'Kampanye Spanduk',
      location: b.location || '-',
      type: b.banner_type || 'Spanduk',
      status: b.status,
      date: formatDate(b.created_at),
      designer: b.designer ? `${b.designer.first_name} ${b.designer.last_name || ''}` : '-',
      installer: b.installer ? `${b.installer.first_name} ${b.installer.last_name || ''}` : '-',
      image: b.documentation_image_url,
      designURL: b.design_url,
      notes: b.notes,
    }))
  } catch {
  } finally {
    isInternalLoading.value = false
  }
}

const isSalesPosition = (title?: string, deptName?: string) => {
  const t = (title || '').toLowerCase()
  const d = (deptName || '').toLowerCase()

  // Blacklist tegas posisi non-sales, harian, logistik, operasional, teknis, kasir
  const isExcluded = /(operator|driver|sopir|kuli|helper|gudang|warehouse|security|cleaning|harian|validator|editor|designer|desain|purchasing|auditor|kasir|cashier|teknisi|technician)/i.test(t)
  if (isExcluded) return false

  // Whitelist sales & marketing
  const isSales = /(sales|motoris|booster|marketing|take order|task order|\bto\b|\bmt\b|modern trade|key account|\bkam\b)/i.test(t) ||
                  /(sales|marketing|commercial)/i.test(d)

  return isSales
}

// ── Computed ───────────────────────────────────────────────────────────────
const uniqueSales = computed(() => {
  const branchId = selectedBranchFilter.value
  const salesEmps = masterStore.employees.filter((emp: any) => {
    const isActive = emp.user?.is_active !== false && 
      emp.employment_status !== 'RESIGNED' && 
      emp.employment_status !== 'TERMINATED' &&
      !String(emp.employment_status || '').toLowerCase().includes('harian')
    const empBranchId = emp.operating_branch_id || emp.branch_id
    const belongsToBranch = !branchId || branchId === 'ALL_BRANCHES' || empBranchId === branchId
    const jobTitle = emp.job_position?.title || ''
    const deptName = emp.department?.name || ''
    return isActive && belongsToBranch && isSalesPosition(jobTitle, deptName)
  })
  
  const list: { name: string; displayName: string }[] = []
  const seen = new Set<string>()

  // 1. Dari HRIS Employees
  for (const e of salesEmps) {
    const name = `${e.first_name} ${e.last_name || ''}`.trim()
    if (name && !seen.has(name.toLowerCase())) {
      seen.add(name.toLowerCase())
      list.push({
        name,
        displayName: `${name} (${e.job_position?.title || 'Sales'})`
      })
    }
  }

  // 2. Dari Branding Agents API (/branding/agents)
  for (const agent of brandingSalesAgents.value) {
    const name = (agent.name || '').trim()
    const role = (agent.role || '').trim()
    if (name && !seen.has(name.toLowerCase()) && isSalesPosition(role)) {
      seen.add(name.toLowerCase())
      list.push({
        name,
        displayName: `${name} (${role || 'Sales Motoris'})`
      })
    }
  }

  return list.sort((a, b) => a.name.localeCompare(b.name))
})

const availableMonths = computed(() => {
  const months: string[] = []
  const now = new Date()
  for (let i = 0; i < 18; i++) {
    const d = new Date(now.getFullYear(), now.getMonth() - i, 1)
    const monthKey = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`
    months.push(monthKey)
  }
  return months
})

const formatMonthKey = (key: string) => {
  const [y, m] = key.split('-')
  const date = new Date(parseInt(y || '2026'), parseInt(m || '1') - 1, 1)
  return date.toLocaleDateString('id-ID', { month: 'long', year: 'numeric' })
}

const totalPages = computed(() => Math.max(1, Math.ceil(totalBanners.value / itemsPerPage.value)))

const nextPage = () => {
  if (currentPage.value < totalPages.value) {
    currentPage.value++
    fetchBrandingBanners()
  }
}
const prevPage = () => {
  if (currentPage.value > 1) {
    currentPage.value--
    fetchBrandingBanners()
  }
}

const activeStats = computed(() => {
  if (activeTab.value === 'branding') {
    if (brandingStats.value) return brandingStats.value
    return {
      total_banners: totalBanners.value,
      installed_banners: 0,
      ready_banners: 0,
      pending_banners: 0,
      total_expense_banner: 0,
      total_shops: '-',
    }
  } else {
    const banners = internalBanners.value
    const installed = banners.filter((b: any) => b.status === 'TERPASANG').length
    const ready = banners.filter(
      (b: any) => b.status === 'READY_FOR_OPERATOR' || b.status === 'SIAP_PASANG' || b.status === 'DITERIMA_OPERATOR',
    ).length
    return {
      total_banners: banners.length,
      total_shops: brandingStats.value?.total_shops || '-',
      installed_banners: installed,
      ready_banners: ready,
      pending_banners: ready,
      total_expense_banner: 0,
    }
  }
})

// ── Helpers ────────────────────────────────────────────────────────────────
const formatCurrency = (v: number) =>
  new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(v || 0)

const formatDate = (d: string) => {
  if (!d) return ''
  const date = new Date(d)
  return new Intl.DateTimeFormat('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  }).format(date)
}

const formatSizeCm = (val: any) => {
  return normalizeDimensionToCm(val)
}

const getSizeSuggestion = (banner: any) => {
  if (!banner) return null
  return calculateBannerDesignSize(banner.size_length, banner.size_width)
}

const brandingStatusColor = (s: string) => {
  const st = (s || '').toLowerCase()
  if (st === 'installed' || st === 'terpasang') return 'bg-emerald-500/10 text-emerald-700 border-emerald-200'
  if (st === 'request') return 'bg-amber-500/10 text-amber-700 border-amber-200'
  if (st === 'editing' || st === 'in_design') return 'bg-purple-500/10 text-purple-700 border-purple-200'
  if (st === 'validation' || st === 'admin_review') return 'bg-indigo-500/10 text-indigo-700 border-indigo-200'
  if (st === 'sent_to_vendor' || st === 'purchasing_queue' || st === 'percetakan') return 'bg-cyan-500/10 text-cyan-700 border-cyan-200'
  if (st === 'ready_for_operator') return 'bg-blue-500/10 text-blue-700 border-blue-200'
  if (st === 'diterima_operator') return 'bg-teal-500/10 text-teal-700 border-teal-200'
  return 'bg-slate-100 text-slate-500 border-slate-200'
}

const brandingStatusLabel = (s: string) => {
  const st = (s || '').toLowerCase()
  if (st === 'installed' || st === 'terpasang') return 'Terpasang'
  if (st === 'request') return 'Request'
  if (st === 'editing' || st === 'in_design') return 'Desain / Edit'
  if (st === 'validation' || st === 'admin_review') return 'Validasi'
  if (st === 'sent_to_vendor' || st === 'purchasing_queue' || st === 'percetakan') return 'Di Vendor'
  if (st === 'ready_for_operator') return 'Siap Pasang'
  if (st === 'diterima_operator') return 'Dibawa Operator'
  return s
}

// ── Export to Excel ────────────────────────────────────────────────────────
const exportToExcel = async () => {
  isExporting.value = true
  toast.loading('Menyiapkan file Excel profesional...', { id: 'export-excel' })
  try {
    const params = new URLSearchParams()
    params.append('limit', '5000')
    const branchId = selectedBranchFilter.value
    if (branchId && branchId !== 'ALL_BRANCHES') {
      params.append('branch_id', branchId)
    }
    if (searchQuery.value) {
      params.append('search', searchQuery.value)
    }
    if (statusFilter.value && statusFilter.value !== 'all') {
      params.append('status', statusFilter.value)
    }
    if (salesFilter.value && salesFilter.value !== 'all') {
      params.append('sales_name', salesFilter.value)
    }
    if (monthFilter.value && monthFilter.value !== 'all') {
      params.append('month', monthFilter.value)
    }
    if (startDate.value) {
      params.append('start_date', startDate.value)
    }
    if (endDate.value) {
      params.append('end_date', endDate.value)
    }
    if (bihunFilter.value === 'bihun') {
      params.append('is_bihun', 'true')
    } else if (bihunFilter.value === 'non_bihun') {
      params.append('is_bihun', 'false')
    }

    const res = await apiClient.get(`/branding/banners?${params.toString()}`)
    const exportDataRaw = res.data?.data || []

    if (exportDataRaw.length === 0) {
      toast.error('Tidak ada data untuk diexport.')
      return
    }

    const branchName = selectedBranchName.value || 'Semua Cabang'
    const workbook = new ExcelJS.Workbook()
    workbook.creator = 'HRIS Karsa HR SuperApp'
    workbook.created = new Date()

    const worksheet = workbook.addWorksheet('Data Spanduk & Banner', {
      views: [{ state: 'frozen', ySplit: 5 }],
      pageSetup: { orientation: 'landscape', fitToPage: true, fitToWidth: 1, fitToHeight: 0 },
    })

    // 1. Header Banner
    worksheet.mergeCells('A1:T1')
    const titleCell = worksheet.getCell('A1')
    titleCell.value = 'LAPORAN REKAPITULASI PEMASANGAN SPANDUK & BANNER'
    titleCell.font = { bold: true, size: 14, color: { argb: 'FFFFFFFF' } }
    titleCell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF0F172A' } }
    titleCell.alignment = { horizontal: 'center', vertical: 'middle' }
    worksheet.getRow(1).height = 34

    // 2. Subtitle / Metadata Info
    let periodeLabel = 'Semua Periode'
    if (startDate.value && endDate.value) {
      periodeLabel = `${formatDate(startDate.value)} s/d ${formatDate(endDate.value)}`
    } else if (startDate.value) {
      periodeLabel = `Sejak ${formatDate(startDate.value)}`
    } else if (endDate.value) {
      periodeLabel = `Sampai ${formatDate(endDate.value)}`
    } else if (monthFilter.value && monthFilter.value !== 'all') {
      periodeLabel = formatMonthKey(monthFilter.value)
    }

    const salesFilterLabel = salesFilter.value === 'all' ? 'Semua Sales' : salesFilter.value
    const statusFilterLabel = statusFilter.value === 'all' ? 'Semua Status' : brandingStatusLabel(statusFilter.value)
    const bihunFilterLabel = bihunFilter.value === 'all' ? 'Semua Kategori' : bihunFilter.value === 'bihun' ? '🍜 Desain Bihun' : 'Bukan Bihun'

    worksheet.mergeCells('A2:T2')
    const subCell = worksheet.getCell('A2')
    subCell.value = `Cabang: ${branchName}   |   Periode: ${periodeLabel}   |   Filter Sales: ${salesFilterLabel}   |   Kategori: ${bihunFilterLabel}   |   Status: ${statusFilterLabel}   |   Total Data: ${exportDataRaw.length} Banner   |   Diexport: ${new Date().toLocaleString('id-ID')}`
    subCell.font = { italic: true, size: 9, color: { argb: 'FF334155' } }
    subCell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFF1F5F9' } }
    subCell.alignment = { horizontal: 'center', vertical: 'middle' }
    worksheet.getRow(2).height = 22

    worksheet.addRow({}) // Row 3 spacer
    worksheet.getRow(3).height = 8

    // 3. Kolom Definisi (Column Setup)
    worksheet.columns = [
      { key: 'no', width: 6 },
      { key: 'created_at', width: 15 },
      { key: 'shop_name', width: 26 },
      { key: 'shop_address', width: 34 },
      { key: 'region', width: 28 },
      { key: 'shop_phone', width: 16 },
      { key: 'gmaps_link', width: 22 },
      { key: 'branch_name', width: 20 },
      { key: 'sales_name', width: 22 },
      { key: 'is_bihun', width: 22 },
      { key: 'editor_name', width: 18 },
      { key: 'operator_name', width: 22 },
      { key: 'operator_phone', width: 16 },
      { key: 'dimension', width: 20 },
      { key: 'area_m2', width: 13 },
      { key: 'qty', width: 10 },
      { key: 'total_price', width: 18 },
      { key: 'status', width: 18 },
      { key: 'installed_at', width: 16 },
      { key: 'notes', width: 28 },
    ]

    // 4. Header Baris Tabel (Row 4)
    const headerRow = worksheet.getRow(4)
    const columnTitles = [
      'No',
      'Tgl Order',
      'Nama Toko',
      'Alamat Toko',
      'Wilayah / Kota',
      'No. Telp Toko',
      'Link Google Maps',
      'Cabang',
      'Sales Agent',
      'Kategori Desain',
      'Desainer / Editor',
      'Operator Pasang',
      'No. Telp Operator',
      'Ukuran (Dimensi)',
      'Luas (m²)',
      'Qty',
      'Total Biaya (Rp)',
      'Status',
      'Tgl Terpasang',
      'Catatan',
    ]

    columnTitles.forEach((title, index) => {
      const cell = headerRow.getCell(index + 1)
      cell.value = title
      cell.font = { bold: true, size: 10, color: { argb: 'FFFFFFFF' } }
      cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF1E293B' } }
      cell.alignment = { horizontal: 'center', vertical: 'middle', wrapText: true }
      cell.border = {
        top: { style: 'thin', color: { argb: 'FF475569' } },
        bottom: { style: 'medium', color: { argb: 'FF0F172A' } },
        left: { style: 'thin', color: { argb: 'FF475569' } },
        right: { style: 'thin', color: { argb: 'FF475569' } },
      }
    })
    headerRow.height = 28

    // 5. Data Rows
    let grandTotalArea = 0
    let grandTotalQty = 0
    let grandTotalPrice = 0

    exportDataRaw.forEach((b: any, index: number) => {
      const lCm = normalizeDimensionToCm(b.size_length)
      const wCm = normalizeDimensionToCm(b.size_width)
      const lM = +(lCm / 100).toFixed(2)
      const wM = +(wCm / 100).toFixed(2)
      const areaM2 = +(lM * wM).toFixed(2)
      const qty = Number(b.qty) || 1
      const price = Number(b.total_price) || 0

      grandTotalArea += areaM2
      grandTotalQty += qty
      grandTotalPrice += price

      // Region string
      const regionParts = [b.village, b.district, b.regency, b.province].filter(Boolean)
      const regionStr = regionParts.length > 0 ? regionParts.join(', ') : '-'

      // Google Maps Coordinates & Link
      const lat = Number(b.latitude)
      const lng = Number(b.longitude)
      const hasGmaps = !isNaN(lat) && !isNaN(lng) && (lat !== 0 || lng !== 0)
      const gmapsUrl = hasGmaps ? `https://www.google.com/maps/search/?api=1&query=${lat},${lng}` : ''

      const isEven = index % 2 === 0
      const rowBg = isEven ? 'FFFFFFFF' : 'FFF8FAFC'

      const row = worksheet.addRow({
        no: index + 1,
        created_at: formatDate(b.created_at) || '-',
        shop_name: b.shop_name || `Shop #${b.shop_id || '-'}`,
        shop_address: b.shop_address || '-',
        region: regionStr,
        shop_phone: b.shop_phone || '-',
        gmaps_link: hasGmaps ? 'Buka Google Maps' : '-',
        branch_name: b.branch_name || b.sales_branch || '-',
        sales_name: b.sales_name || '-',
        is_bihun: b.is_bihun ? '🍜 Desain Bihun' : 'Reguler (Bukan Bihun)',
        editor_name: b.editor_name || '-',
        operator_name: b.operator_name || '-',
        operator_phone: b.operator_phone || '-',
        dimension: formatBannerDimension(lCm, wCm, 'dual'),
        area_m2: areaM2,
        qty: qty,
        total_price: price,
        status: brandingStatusLabel(b.status),
        installed_at: b.installed_at ? formatDate(b.installed_at) : '-',
        notes: b.notes || '-',
      })

      row.eachCell({ includeEmpty: true }, (cell, colNumber) => {
        cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: rowBg } }
        cell.border = {
          top: { style: 'thin', color: { argb: 'FFE2E8F0' } },
          bottom: { style: 'thin', color: { argb: 'FFE2E8F0' } },
          left: { style: 'thin', color: { argb: 'FFE2E8F0' } },
          right: { style: 'thin', color: { argb: 'FFE2E8F0' } },
        }

        // Alignments & Formats
        if (colNumber === 1) { // No
          cell.alignment = { horizontal: 'center', vertical: 'middle' }
          cell.font = { size: 9, color: { argb: 'FF64748B' } }
        } else if (colNumber === 2 || colNumber === 19) { // Tgl Order & Tgl Pasang
          cell.alignment = { horizontal: 'center', vertical: 'middle' }
          cell.font = { size: 9 }
        } else if (colNumber === 3) { // Nama Toko
          cell.alignment = { horizontal: 'left', vertical: 'middle' }
          cell.font = { bold: true, size: 9.5, color: { argb: 'FF0F172A' } }
        } else if (colNumber === 4 || colNumber === 5) { // Alamat & Wilayah
          cell.alignment = { horizontal: 'left', vertical: 'middle', wrapText: true }
          cell.font = { size: 9, color: { argb: 'FF334155' } }
        } else if (colNumber === 6) { // No Telp
          cell.alignment = { horizontal: 'center', vertical: 'middle' }
          cell.font = { size: 9 }
        } else if (colNumber === 7) { // Link Google Maps
          cell.alignment = { horizontal: 'center', vertical: 'middle' }
          if (hasGmaps) {
            cell.value = {
              text: '📍 Buka Google Maps',
              hyperlink: gmapsUrl,
              tooltip: `Koordinat: ${lat}, ${lng}`
            }
            cell.font = { size: 9, bold: true, color: { argb: 'FF2563EB' }, underline: true }
          } else {
            cell.value = '-'
            cell.font = { size: 9, color: { argb: 'FF94A3B8' } }
          }
        } else if (colNumber === 8) { // Cabang
          cell.alignment = { horizontal: 'left', vertical: 'middle' }
          cell.font = { size: 9, color: { argb: 'FF1E293B' } }
        } else if (colNumber === 9) { // Sales
          cell.alignment = { horizontal: 'left', vertical: 'middle' }
          cell.font = { size: 9, color: { argb: 'FF1E293B' } }
        } else if (colNumber === 10) { // Kategori Desain Bihun
          cell.alignment = { horizontal: 'center', vertical: 'middle' }
          if (b.is_bihun) {
            cell.font = { bold: true, size: 9, color: { argb: 'FFB45309' } }
            cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFFFFBEB' } }
          } else {
            cell.font = { size: 9, color: { argb: 'FF64748B' } }
          }
        } else if (colNumber === 11 || colNumber === 12) { // Editor & Operator
          cell.alignment = { horizontal: 'left', vertical: 'middle' }
          cell.font = { size: 9, color: { argb: 'FF1E293B' } }
        } else if (colNumber === 13) { // No Telp Operator
          cell.alignment = { horizontal: 'center', vertical: 'middle' }
          cell.font = { size: 9 }
        } else if (colNumber === 14) { // Dimensi
          cell.alignment = { horizontal: 'center', vertical: 'middle' }
          cell.font = { size: 9 }
        } else if (colNumber === 15) { // Luas m2
          cell.alignment = { horizontal: 'right', vertical: 'middle' }
          cell.numFmt = '#,##0.00'
          cell.font = { size: 9, bold: true, color: { argb: 'FF0F766E' } }
        } else if (colNumber === 16) { // Qty
          cell.alignment = { horizontal: 'center', vertical: 'middle' }
          cell.numFmt = '#,##0'
          cell.font = { size: 9, bold: true }
        } else if (colNumber === 17) { // Total Price
          cell.alignment = { horizontal: 'right', vertical: 'middle' }
          cell.numFmt = '[$Rp-421] #,##0'
          cell.font = { size: 9, bold: true, color: { argb: 'FF1E293B' } }
        } else if (colNumber === 18) { // Status
          cell.alignment = { horizontal: 'center', vertical: 'middle' }
          cell.font = { bold: true, size: 9 }
          const st = (b.status || '').toLowerCase()
          if (st === 'installed' || st === 'terpasang') {
            cell.font = { bold: true, size: 9, color: { argb: 'FF047857' } }
            cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFECFDF5' } }
          } else if (st === 'request') {
            cell.font = { bold: true, size: 9, color: { argb: 'FFB45309' } }
            cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFFFFBEB' } }
          } else if (st === 'editing' || st === 'in_design') {
            cell.font = { bold: true, size: 9, color: { argb: 'FF6D28D9' } }
            cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFF5F3FF' } }
          } else if (st === 'validation' || st === 'admin_review') {
            cell.font = { bold: true, size: 9, color: { argb: 'FF4338CA' } }
            cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFEEF2FF' } }
          } else if (st === 'sent_to_vendor' || st === 'purchasing_queue' || st === 'percetakan') {
            cell.font = { bold: true, size: 9, color: { argb: 'FF0E7490' } }
            cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFECFEFF' } }
          } else if (st === 'ready_for_operator' || st === 'diterima_operator') {
            cell.font = { bold: true, size: 9, color: { argb: 'FF1D4ED8' } }
            cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFEFF6FF' } }
          }
        } else if (colNumber === 20) { // Catatan
          cell.alignment = { horizontal: 'left', vertical: 'middle', wrapText: true }
          cell.font = { size: 8.5, color: { argb: 'FF64748B' } }
        }
      })

      row.height = 26
    })

    // 6. Grand Total Row
    const totalRow = worksheet.addRow({
      no: '',
      created_at: '',
      shop_name: 'GRAND TOTAL',
      shop_address: `${exportDataRaw.length} Banner`,
      region: '',
      shop_phone: '',
      gmaps_link: '',
      branch_name: '',
      sales_name: '',
      is_bihun: '',
      editor_name: '',
      operator_name: '',
      operator_phone: '',
      dimension: '',
      area_m2: grandTotalArea,
      qty: grandTotalQty,
      total_price: grandTotalPrice,
      status: '',
      installed_at: '',
      notes: '',
    })

    totalRow.eachCell({ includeEmpty: true }, (cell, colNumber) => {
      cell.font = { bold: true, size: 10, color: { argb: 'FF0F172A' } }
      cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFE2E8F0' } }
      cell.border = {
        top: { style: 'medium', color: { argb: 'FF0F172A' } },
        bottom: { style: 'double', color: { argb: 'FF0F172A' } },
        left: { style: 'thin', color: { argb: 'FFCBD5E1' } },
        right: { style: 'thin', color: { argb: 'FFCBD5E1' } },
      }

      if (colNumber === 3) {
        cell.alignment = { horizontal: 'center', vertical: 'middle' }
      } else if (colNumber === 4) {
        cell.alignment = { horizontal: 'left', vertical: 'middle' }
        cell.font = { italic: true, size: 9, color: { argb: 'FF475569' } }
      } else if (colNumber === 15) {
        cell.numFmt = '#,##0.00'
        cell.alignment = { horizontal: 'right', vertical: 'middle' }
        cell.font = { bold: true, color: { argb: 'FF0F766E' } }
      } else if (colNumber === 16) {
        cell.numFmt = '#,##0'
        cell.alignment = { horizontal: 'center', vertical: 'middle' }
      } else if (colNumber === 17) {
        cell.numFmt = '[$Rp-421] #,##0'
        cell.alignment = { horizontal: 'right', vertical: 'middle' }
      }
    })
    totalRow.height = 28

    // 7. Write and Trigger File Download
    const buffer = await workbook.xlsx.writeBuffer()
    const blob = new Blob([buffer], {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    })

    let dateSuffix = new Date().toISOString().slice(0, 10)
    if (startDate.value && endDate.value) {
      dateSuffix = `${startDate.value}_sd_${endDate.value}`
    } else if (startDate.value) {
      dateSuffix = `sejak_${startDate.value}`
    } else if (endDate.value) {
      dateSuffix = `sd_${endDate.value}`
    }

    const fileName = `Data_Spanduk_${dateSuffix}.xlsx`
    saveAs(blob, fileName)
    toast.success(`Data ${exportDataRaw.length} spanduk berhasil diexport ke Excel.`)
  } catch (error: any) {
    toast.error('Gagal mengeksport data: ' + (error.response?.data?.error || error.message))
  } finally {
    isExporting.value = false
  }
}

const isSyncing = ref(false)
const isRecovering = ref(false)

const handleSyncFromBrandingApp = async () => {
  isSyncing.value = true
  const tId = toast.loading('Menghubungkan ke MySQL Branding & sinkronisasi data...')
  try {
    const res = await apiClient.post('/branding/sync')
    const { total_banners, updated_count, created_count, completed_count, recovered_count } = res.data
    const recoveredInfo = recovered_count ? ` (${recovered_count} status otomatis diproteksi/dipulihkan)` : ''
    toast.success(
      `Sinkronisasi berhasil! ${total_banners} banner diproses (${updated_count} diupdate, ${created_count} baru, ${completed_count} selesai/terpasang)${recoveredInfo}.`,
      { id: tId, duration: 6000 }
    )
    await Promise.all([fetchBrandingBanners(), fetchBrandingStats()])
  } catch (err: any) {
    const errMsg = err?.response?.data?.error || err?.message || 'Gagal sinkronisasi data dari Branding App'
    toast.error(`Sinkronisasi gagal: ${errMsg}`, { id: tId, duration: 6000 })
  } finally {
    isSyncing.value = false
  }
}

const handleRecoverWorkflow = async () => {
  isRecovering.value = true
  const tId = toast.loading('Merekonstruksi dan memulihkan status workflow Super App...')
  try {
    const res = await apiClient.post('/branding/recover')
    const recoveredCount = res.data?.recovered_count || 0
    toast.success(
      `Pemulihan selesai! ${recoveredCount} banner berhasil direkonstruksi kembali ke status kerja aslinya.`,
      { id: tId, duration: 6000 }
    )
    await Promise.all([fetchBrandingBanners(), fetchBrandingStats()])
  } catch (err: any) {
    const errMsg = err?.response?.data?.error || err?.message || 'Gagal merekonstruksi status workflow'
    toast.error(`Pemulihan gagal: ${errMsg}`, { id: tId, duration: 6000 })
  } finally {
    isRecovering.value = false
  }
}

const handleRefreshBranding = () => {
  currentPage.value = 1
  fetchBrandingBanners()
  fetchBrandingStats()
  fetchBrandingSalesAgents()
  masterStore.fetchEmployees(true)
}

const handleBranchFilterChange = () => {
  salesFilter.value = 'all'
  currentPage.value = 1
  fetchBrandingBanners()
  fetchBrandingStats()
  fetchBrandingSalesAgents()
  fetchAllShops()
}

const handleFilterBrandingBanners = (s: string) => {
  statusFilter.value = s
  currentPage.value = 1
  fetchBrandingBanners()
}

const hasActiveFilters = computed(() => {
  return (
    Boolean(searchQuery.value.trim()) ||
    statusFilter.value !== 'all' ||
    (selectedBranchFilter.value && selectedBranchFilter.value !== 'ALL_BRANCHES') ||
    salesFilter.value !== 'all' ||
    bihunFilter.value !== 'all' ||
    monthFilter.value !== 'all' ||
    Boolean(startDate.value) ||
    Boolean(endDate.value)
  )
})

const resetAllFilters = () => {
  searchQuery.value = ''
  statusFilter.value = 'all'
  masterStore.setSelectedBrandingBranchId('ALL_BRANCHES')
  salesFilter.value = 'all'
  bihunFilter.value = 'all'
  monthFilter.value = 'all'
  datePreset.value = 'all'
  startDate.value = ''
  endDate.value = ''
  currentPage.value = 1
  fetchBrandingBanners()
}

watch(() => masterStore.selectedBrandingBranchId, () => {
  handleBranchFilterChange()
})

watch([salesFilter, monthFilter, startDate, endDate, bihunFilter], () => {
  currentPage.value = 1
  fetchBrandingBanners()
})

onMounted(async () => {
  if (!masterStore.branches || masterStore.branches.length === 0) {
    await masterStore.fetchBranches()
  }
  fetchBrandingStats()
  fetchBrandingBanners()
  fetchInternalBanners()
  fetchBrandingSalesAgents()
  masterStore.fetchEmployees(true)
})
</script>

<template>
  <div class="space-y-8 pb-12 animate-in fade-in duration-500">
    <!-- Header -->
    <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-4 sm:gap-5">
      <div>
        <div class="flex flex-wrap items-center gap-2 sm:gap-3 mb-1">
          <h1 class="text-2xl sm:text-4xl font-black text-slate-900 tracking-tight flex items-center flex-wrap gap-2">
            <span>Spanduk & Banner</span>
            <span class="text-primary text-xl sm:text-3xl font-bold">
              — {{ selectedBranchName }}
            </span>
          </h1>
          <span
            class="flex items-center gap-1.5 px-2.5 sm:px-3 py-1 rounded-full text-[9px] sm:text-[10px] font-black uppercase tracking-widest border bg-emerald-500/10 text-emerald-700 border-emerald-200 shadow-sm"
          >
            <CheckCircle2 class="w-3 h-3 text-emerald-600 shrink-0" />
            Database Super App
          </span>
        </div>
        <p class="text-xs sm:text-sm font-medium text-slate-500">
          Manajemen & Monitoring Spanduk <strong class="text-slate-700">Perusahaan Demo</strong>
        </p>
      </div>
      <div class="grid grid-cols-2 sm:flex sm:flex-wrap items-center gap-2 sm:gap-3 w-full lg:w-auto">
        <button
          v-if="activeTab === 'branding' && canManageBanners"
          @click="openAddModal"
          class="flex items-center justify-center gap-1.5 sm:gap-2 px-3 sm:px-5 py-2.5 sm:py-3 bg-blue-600 text-white rounded-xl sm:rounded-2xl font-bold text-xs sm:text-sm shadow-md sm:shadow-lg shadow-blue-600/20 hover:bg-blue-700 active:scale-95 transition-all cursor-pointer"
        >
          <Plus class="w-4 h-4 shrink-0" />
          <span class="truncate">Tambah Banner</span>
        </button>
        <button
          v-if="activeTab === 'branding' && canManageBanners"
          @click="handleSyncFromBrandingApp"
          :disabled="isSyncing || isBrandingLoading || isRecovering"
          class="flex items-center justify-center gap-1.5 sm:gap-2 px-3 sm:px-5 py-2.5 sm:py-3 bg-indigo-600 text-white rounded-xl sm:rounded-2xl font-bold text-xs sm:text-sm shadow-md sm:shadow-lg shadow-indigo-600/20 hover:bg-indigo-700 active:scale-95 transition-all disabled:opacity-50 cursor-pointer"
          title="Sinkronisasi & tarik status terbaru (selesai/bukti pasang) langsung dari MySQL Branding"
        >
          <RefreshCw class="w-4 h-4 shrink-0" :class="isSyncing ? 'animate-spin' : ''" />
          <span class="truncate">{{ isSyncing ? 'Menyinkronkan...' : 'Tarik Data Branding' }}</span>
        </button>
        <button
          v-if="activeTab === 'branding' && canManageBanners"
          @click="handleRecoverWorkflow"
          :disabled="isRecovering || isBrandingLoading || isSyncing"
          class="flex items-center justify-center gap-1.5 sm:gap-2 px-3 sm:px-5 py-2.5 sm:py-3 bg-amber-500 hover:bg-amber-600 text-white rounded-xl sm:rounded-2xl font-bold text-xs sm:text-sm shadow-md sm:shadow-lg shadow-amber-500/20 active:scale-95 transition-all disabled:opacity-50 cursor-pointer"
          title="Rekonstruksi status semua spanduk berdasarkan bukti pengerjaan (desain, validasi, percetakan, operator)"
        >
          <Sparkles class="w-4 h-4 shrink-0" :class="isRecovering ? 'animate-spin' : ''" />
          <span class="truncate">{{ isRecovering ? 'Memulihkan...' : 'Pulihkan Status' }}</span>
        </button>
        <button
          v-if="activeTab === 'branding'"
          @click="exportToExcel"
          :disabled="isExporting"
          class="flex items-center justify-center gap-1.5 sm:gap-2 px-3 sm:px-5 py-2.5 sm:py-3 bg-emerald-600 text-white rounded-xl sm:rounded-2xl font-bold text-xs sm:text-sm shadow-md sm:shadow-lg shadow-emerald-600/20 hover:bg-emerald-700 active:scale-95 transition-all disabled:opacity-50 cursor-pointer"
        >
          <Download class="w-4 h-4 shrink-0" :class="isExporting ? 'animate-bounce' : ''" />
          <span class="truncate">{{ isExporting ? 'Mengexport...' : 'Tarik Excel' }}</span>
        </button>
        <button
          @click="handleRefreshBranding"
          :disabled="isBrandingLoading"
          class="col-span-2 sm:col-span-1 flex items-center justify-center gap-1.5 sm:gap-2 px-3 sm:px-5 py-2.5 sm:py-3 bg-primary text-white rounded-xl sm:rounded-2xl font-bold text-xs sm:text-sm shadow-md sm:shadow-lg shadow-primary/20 hover:bg-primary/90 active:scale-95 transition-all disabled:opacity-50 cursor-pointer"
        >
          <RefreshCw class="w-4 h-4 shrink-0" :class="isBrandingLoading ? 'animate-spin' : ''" />
          <span>Refresh</span>
        </button>
      </div>
    </div>

    <!-- Dynamic Stats -->
    <div class="grid grid-cols-2 md:grid-cols-5 gap-3 sm:gap-4" v-if="activeStats">
      <div class="bg-white p-3.5 sm:p-5 rounded-2xl sm:rounded-[1.75rem] border border-slate-200 shadow-sm">
        <Package class="w-5 h-5 sm:w-6 sm:h-6 text-slate-400 mb-1 sm:mb-2" />
        <p class="text-[8px] sm:text-[9px] font-black text-slate-400 uppercase tracking-widest">Total Banner</p>
        <h3 class="text-2xl sm:text-3xl font-black text-slate-900 mt-0.5">{{ activeStats.total_banners }}</h3>
      </div>
      <div class="bg-white p-3.5 sm:p-5 rounded-2xl sm:rounded-[1.75rem] border border-slate-200 shadow-sm">
        <Store class="w-5 h-5 sm:w-6 sm:h-6 text-primary mb-1 sm:mb-2 opacity-70" />
        <p class="text-[8px] sm:text-[9px] font-black text-slate-400 uppercase tracking-widest">Total Toko</p>
        <h3 class="text-2xl sm:text-3xl font-black text-slate-900 mt-0.5">
          {{ activeStats.total_shops || '-' }}
        </h3>
      </div>
      <div class="bg-white p-3.5 sm:p-5 rounded-2xl sm:rounded-[1.75rem] border border-slate-200 shadow-sm">
        <CheckCircle2 class="w-5 h-5 sm:w-6 sm:h-6 text-emerald-500 mb-1 sm:mb-2" />
        <p class="text-[8px] sm:text-[9px] font-black text-slate-400 uppercase tracking-widest">Terpasang</p>
        <h3 class="text-2xl sm:text-3xl font-black text-emerald-600 mt-0.5">
          {{ activeStats.installed_banners }}
        </h3>
      </div>
      <div class="bg-white p-3.5 sm:p-5 rounded-2xl sm:rounded-[1.75rem] border border-slate-200 shadow-sm">
        <Clock class="w-5 h-5 sm:w-6 sm:h-6 text-blue-500 mb-1 sm:mb-2" />
        <p class="text-[8px] sm:text-[9px] font-black text-slate-400 uppercase tracking-widest">Siap Pasang</p>
        <h3 class="text-2xl sm:text-3xl font-black text-blue-600 mt-0.5">
          {{ activeStats.ready_banners ?? activeStats.pending_banners }}
        </h3>
      </div>
      <div
        class="bg-gradient-to-br from-primary to-primary/80 p-3.5 sm:p-5 rounded-2xl sm:rounded-[1.75rem] shadow-lg shadow-primary/25 relative overflow-hidden col-span-2 md:col-span-1"
      >
        <TrendingDown class="absolute -right-2 -bottom-2 w-14 h-14 sm:w-16 sm:h-16 opacity-10" />
        <p class="text-[8px] sm:text-[9px] font-black text-white/70 uppercase tracking-widest">Pengeluaran</p>
        <h3 class="text-lg sm:text-xl font-black text-white mt-0.5 leading-tight">
          {{ formatCurrency(activeStats.total_expense_banner || activeStats.total_revenue_banner || 0) }}
        </h3>
      </div>
    </div>
    <div v-else-if="isBrandingLoading" class="grid grid-cols-2 md:grid-cols-5 gap-3 sm:gap-4">
      <div v-for="i in 5" :key="i" class="h-20 sm:h-24 bg-slate-100 rounded-2xl sm:rounded-[1.75rem] animate-pulse"></div>
    </div>

    <!-- Layer 1: Search, Workflow Status Pills & View Mode Toggle -->
    <div
      v-if="activeTab === 'branding'"
      class="flex flex-col xl:flex-row xl:items-center justify-between gap-3.5"
    >
      <!-- Search Input & Workflow Status Tabs -->
      <div class="flex flex-col lg:flex-row items-stretch lg:items-center gap-3 flex-1">
        <!-- Search Input -->
        <div class="relative w-full lg:w-80 shrink-0">
          <Loader2 v-if="isSearching" class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-primary animate-spin" />
          <Search v-else class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input
            v-model="searchQuery"
            @input="handleSearchInput"
            type="text"
            placeholder="Cari nama toko atau sales..."
            class="w-full bg-white border border-slate-200 rounded-2xl py-2.5 pl-11 pr-9 text-xs sm:text-sm font-medium focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all placeholder:text-slate-400 shadow-xs"
          />
          <button
            v-if="searchQuery"
            type="button"
            @click="searchQuery = ''; handleSearchInput()"
            class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 p-0.5 cursor-pointer"
            title="Hapus pencarian"
          >
            <X class="w-4 h-4" />
          </button>
        </div>

        <!-- Workflow Status Pills -->
        <div class="flex flex-wrap items-center gap-1.5 overflow-x-auto no-scrollbar py-0.5">
          <button
            v-for="s in ['all', 'request', 'editing', 'validation', 'sent_to_vendor', 'ready_for_operator', 'installed']"
            :key="s"
            @click="handleFilterBrandingBanners(s)"
            :class="[
              'px-3 sm:px-3.5 py-2 rounded-xl text-[10px] sm:text-xs font-black uppercase tracking-wider border transition-all cursor-pointer whitespace-nowrap',
              statusFilter === s
                ? 'bg-primary text-white border-primary shadow-sm shadow-primary/25'
                : 'bg-white text-slate-600 border-slate-200 hover:border-primary/40 hover:bg-slate-50',
            ]"
          >
            {{ s === 'all' ? `Semua Status (${totalBanners})` : brandingStatusLabel(s) }}
          </button>
        </div>
      </div>

      <!-- View Toggle (Tabel / Kartu) -->
      <div class="bg-white border border-slate-200 p-1 rounded-xl shadow-xs shrink-0 self-start xl:self-auto flex items-center">
        <button
          @click="viewMode = 'table'"
          :class="[
            'px-3.5 py-1.5 rounded-lg text-xs font-black uppercase tracking-wider transition-all cursor-pointer',
            viewMode === 'table'
              ? 'bg-primary text-white shadow-xs'
              : 'text-slate-500 hover:text-slate-800',
          ]"
        >
          Tabel
        </button>
        <button
          @click="viewMode = 'grid'"
          :class="[
            'px-3.5 py-1.5 rounded-lg text-xs font-black uppercase tracking-wider transition-all cursor-pointer',
            viewMode === 'grid'
              ? 'bg-primary text-white shadow-xs'
              : 'text-slate-500 hover:text-slate-800',
          ]"
        >
          Kartu
        </button>
      </div>
    </div>

    <!-- Layer 2: Filter Panel Grid Card -->
    <div
      v-if="activeTab === 'branding'"
      class="bg-white border border-slate-200/90 rounded-2xl p-4 sm:p-5 shadow-xs space-y-3.5"
    >
      <div class="flex items-center justify-between gap-2 border-b border-slate-100 pb-3">
        <div class="flex items-center gap-2">
          <div class="p-1.5 bg-primary/10 rounded-lg text-primary">
            <SlidersHorizontal class="w-4 h-4" />
          </div>
          <span class="text-xs sm:text-sm font-black text-slate-800 tracking-tight">Filter Data Spanduk</span>
        </div>
        <button
          v-if="hasActiveFilters"
          type="button"
          @click="resetAllFilters"
          class="text-[11px] sm:text-xs font-bold text-red-600 hover:text-red-700 bg-red-50 hover:bg-red-100/80 px-2.5 py-1 rounded-lg transition-all flex items-center gap-1.5 cursor-pointer shadow-2xs"
        >
          <RotateCcw class="w-3 h-3" />
          Reset Filter
        </button>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-12 gap-3.5">
        <!-- Filter Cabang -->
        <div class="lg:col-span-3 space-y-1.5">
          <label class="flex items-center gap-1.5 text-[11px] font-extrabold text-slate-600">
            <Building2 class="w-3.5 h-3.5 text-slate-400" />
            Cabang Perusahaan
          </label>
          <select
            v-model="selectedBranchFilter"
            @change="handleBranchFilterChange"
            class="w-full bg-slate-50 hover:bg-white focus:bg-white border border-slate-200 rounded-xl px-3.5 py-2 text-xs sm:text-sm font-bold text-slate-700 focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all cursor-pointer shadow-2xs"
          >
            <option value="ALL_BRANCHES">Semua Cabang</option>
            <option disabled>──────────</option>
            <option v-for="b in masterStore.branches" :key="b.id" :value="String(b.id)">
              {{ b.name }}
            </option>
          </select>
        </div>

        <!-- Filter Sales -->
        <div class="lg:col-span-3 space-y-1.5">
          <label class="flex items-center gap-1.5 text-[11px] font-extrabold text-slate-600">
            <Users class="w-3.5 h-3.5 text-slate-400" />
            Sales Agent / Binaan
          </label>
          <select
            v-model="salesFilter"
            class="w-full bg-slate-50 hover:bg-white focus:bg-white border border-slate-200 rounded-xl px-3.5 py-2 text-xs sm:text-sm font-bold text-slate-700 focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all cursor-pointer shadow-2xs"
          >
            <option value="all">Semua Sales</option>
            <option v-for="s in uniqueSales" :key="s.name" :value="s.name">{{ s.displayName }}</option>
          </select>
        </div>

        <!-- Filter Kategori Bihun -->
        <div class="lg:col-span-2 space-y-1.5">
          <label class="flex items-center gap-1.5 text-[11px] font-extrabold text-slate-600">
            <Sparkles class="w-3.5 h-3.5 text-amber-500" />
            Kategori Desain
          </label>
          <select
            v-model="bihunFilter"
            class="w-full bg-slate-50 hover:bg-white focus:bg-white border border-slate-200 rounded-xl px-3.5 py-2 text-xs sm:text-sm font-bold text-slate-700 focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all cursor-pointer shadow-2xs"
          >
            <option value="all">Semua Kategori</option>
            <option value="bihun">🍜 Desain Bihun</option>
            <option value="non_bihun">Bukan Bihun</option>
          </select>
        </div>

        <!-- Filter Periode & Tanggal -->
        <div class="lg:col-span-4 space-y-1.5">
          <label class="flex items-center justify-between gap-1 text-[11px] font-extrabold text-slate-600">
            <span class="flex items-center gap-1.5">
              <Calendar class="w-3.5 h-3.5 text-slate-400" />
              Periode & Rentang Waktu
            </span>
            <span v-if="startDate || endDate" class="text-[10px] text-primary font-bold">Custom Range</span>
          </label>
          <div class="flex flex-col sm:flex-row items-stretch gap-2">
            <!-- Dropdown Bulan -->
            <select
              v-model="monthFilter"
              @change="if (monthFilter !== 'all') { startDate = ''; endDate = ''; datePreset = 'all' }"
              class="sm:w-36 shrink-0 bg-slate-50 hover:bg-white focus:bg-white border border-slate-200 rounded-xl px-3 py-2 text-xs sm:text-sm font-bold text-slate-700 focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all cursor-pointer shadow-2xs"
            >
              <option value="all">Semua Bulan</option>
              <option v-for="m in availableMonths" :key="m" :value="m">{{ formatMonthKey(m) }}</option>
            </select>

            <!-- Quick Preset Buttons & Date Input Container -->
            <div class="flex-1 flex flex-wrap items-center gap-1.5 bg-slate-50 border border-slate-200 p-1 rounded-xl shadow-2xs">
              <div class="flex items-center gap-0.5 bg-slate-200/60 p-0.5 rounded-lg">
                <button
                  type="button"
                  @click="setDatePreset('all')"
                  :class="[
                    'px-2 py-1 text-[10px] font-extrabold rounded-md transition-all cursor-pointer',
                    datePreset === 'all' && !startDate && !endDate ? 'bg-white shadow-xs text-slate-900' : 'text-slate-500 hover:text-slate-800'
                  ]"
                >
                  Semua
                </button>
                <button
                  type="button"
                  @click="monthFilter = 'all'; setDatePreset('today')"
                  :class="[
                    'px-2 py-1 text-[10px] font-extrabold rounded-md transition-all cursor-pointer',
                    datePreset === 'today' ? 'bg-white shadow-xs text-slate-900' : 'text-slate-500 hover:text-slate-800'
                  ]"
                >
                  Hari Ini
                </button>
                <button
                  type="button"
                  @click="monthFilter = 'all'; setDatePreset('7d')"
                  :class="[
                    'px-2 py-1 text-[10px] font-extrabold rounded-md transition-all cursor-pointer',
                    datePreset === '7d' ? 'bg-white shadow-xs text-slate-900' : 'text-slate-500 hover:text-slate-800'
                  ]"
                >
                  7H
                </button>
                <button
                  type="button"
                  @click="monthFilter = 'all'; setDatePreset('30d')"
                  :class="[
                    'px-2 py-1 text-[10px] font-extrabold rounded-md transition-all cursor-pointer',
                    datePreset === '30d' ? 'bg-white shadow-xs text-slate-900' : 'text-slate-500 hover:text-slate-800'
                  ]"
                >
                  30H
                </button>
              </div>
              <div class="flex items-center gap-1 text-[11px] text-slate-500 font-semibold flex-1 min-w-[170px]">
                <input
                  v-model="startDate"
                  type="date"
                  @change="monthFilter = 'all'; datePreset = 'custom'"
                  class="w-full bg-white border border-slate-200 rounded-lg px-2 py-1 text-[11px] font-medium text-slate-700 focus:outline-none focus:ring-1 focus:ring-primary/30"
                  title="Tanggal Awal"
                />
                <span class="text-slate-400 font-bold">—</span>
                <input
                  v-model="endDate"
                  type="date"
                  @change="monthFilter = 'all'; datePreset = 'custom'"
                  class="w-full bg-white border border-slate-200 rounded-lg px-2 py-1 text-[11px] font-medium text-slate-700 focus:outline-none focus:ring-1 focus:ring-primary/30"
                  title="Tanggal Akhir"
                />
                <button
                  v-if="startDate || endDate"
                  type="button"
                  @click="setDatePreset('all')"
                  class="text-slate-400 hover:text-red-500 p-0.5 shrink-0 cursor-pointer"
                  title="Reset rentang tanggal"
                >
                  <X class="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Layer 3: Active Filter Chips Bar -->
      <div
        v-if="hasActiveFilters"
        class="pt-3 border-t border-slate-100 flex flex-wrap items-center gap-2"
      >
        <span class="text-[11px] font-bold text-slate-400 flex items-center gap-1">
          <Filter class="w-3 h-3 text-slate-400" /> Filter aktif:
        </span>
        
        <span
          v-if="searchQuery"
          class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-slate-100 text-slate-700 border border-slate-200"
        >
          Pencarian: "{{ searchQuery }}"
          <button @click="searchQuery = ''; handleSearchInput()" class="hover:text-red-500 cursor-pointer"><X class="w-3 h-3" /></button>
        </span>

        <span
          v-if="statusFilter !== 'all'"
          class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-primary/10 text-primary border border-primary/20"
        >
          Status: {{ brandingStatusLabel(statusFilter) }}
          <button @click="handleFilterBrandingBanners('all')" class="hover:text-red-500 cursor-pointer"><X class="w-3 h-3" /></button>
        </span>

        <span
          v-if="selectedBranchFilter && selectedBranchFilter !== 'ALL_BRANCHES'"
          class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-blue-50 text-blue-700 border border-blue-200"
        >
          Cabang: {{ selectedBranchName }}
          <button @click="selectedBranchFilter = 'ALL_BRANCHES'; handleBranchFilterChange()" class="hover:text-red-500 cursor-pointer"><X class="w-3 h-3" /></button>
        </span>

        <span
          v-if="salesFilter !== 'all'"
          class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-indigo-50 text-indigo-700 border border-indigo-200"
        >
          Sales: {{ salesFilter }}
          <button @click="salesFilter = 'all'" class="hover:text-red-500 cursor-pointer"><X class="w-3 h-3" /></button>
        </span>

        <span
          v-if="bihunFilter !== 'all'"
          class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-amber-50 text-amber-800 border border-amber-200"
        >
          {{ bihunFilter === 'bihun' ? '🍜 Desain Bihun Saja' : 'Bukan Bihun' }}
          <button @click="bihunFilter = 'all'" class="hover:text-red-500 cursor-pointer"><X class="w-3 h-3" /></button>
        </span>

        <span
          v-if="monthFilter !== 'all'"
          class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-emerald-50 text-emerald-700 border border-emerald-200"
        >
          Bulan: {{ formatMonthKey(monthFilter) }}
          <button @click="monthFilter = 'all'" class="hover:text-red-500 cursor-pointer"><X class="w-3 h-3" /></button>
        </span>

        <span
          v-if="startDate || endDate"
          class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-teal-50 text-teal-700 border border-teal-200"
        >
          Rentang: {{ formatDate(startDate) || 'Awal' }} s/d {{ formatDate(endDate) || 'Akhir' }}
          <button @click="setDatePreset('all')" class="hover:text-red-500 cursor-pointer"><X class="w-3 h-3" /></button>
        </span>
      </div>
    </div>

    <!-- Offline Notice -->
    <div
      v-if="isOffline"
      class="bg-amber-50 border border-amber-200 rounded-[1.5rem] p-5 flex items-start justify-between gap-4"
    >
      <div class="flex items-start gap-4">
        <WifiOff class="w-5 h-5 text-amber-500 shrink-0 mt-0.5" />
        <div>
          <p class="font-black text-amber-800 text-sm">
            Gagal Terhubung ke Database Branding
          </p>
          <p class="text-xs text-amber-700 mt-1 leading-relaxed">
            Tidak dapat memuat data spanduk dari server. Silakan periksa koneksi jaringan Anda atau muat ulang halaman.
          </p>
        </div>
      </div>
      <button
        @click="fetchBrandingBanners"
        class="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-bold text-amber-800 bg-amber-100 hover:bg-amber-200 rounded-lg transition shrink-0"
      >
        <RefreshCw class="w-3.5 h-3.5" />
        Muat Ulang
      </button>
    </div>

    <!-- ── TAB: BRANDING APP ──────────────────────────────────────────── -->
    <div v-if="activeTab === 'branding'">
      <div v-if="isBrandingLoading" class="py-20 flex justify-center">
        <span
          class="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin"
        ></span>
      </div>

      <div
        v-else-if="brandingBanners.length === 0"
        class="bg-white rounded-[2.5rem] border border-slate-200 p-20 text-center shadow-sm"
      >
        <Megaphone class="w-16 h-16 text-slate-200 mx-auto mb-4" />
        <h3 class="text-xl font-black text-slate-900">Belum Ada Data Spanduk</h3>
        <p class="text-slate-400 text-sm mt-2">Daftar spanduk dan banner masih kosong.</p>
      </div>

      <div
        v-else-if="viewMode === 'table'"
        :class="[
          'bg-white rounded-[2.5rem] border border-slate-200 shadow-sm overflow-hidden transition-opacity duration-200',
          isSearching ? 'opacity-60 pointer-events-none' : ''
        ]"
      >
        <div class="overflow-x-auto">
          <table class="w-full text-left border-collapse">
            <thead>
              <tr
                class="bg-slate-50 border-b border-slate-100 text-[9px] font-black uppercase tracking-[0.12em] text-slate-400"
              >
                <th class="px-5 py-4">Toko</th>
                <th class="px-5 py-4">Tipe</th>
                <th class="px-5 py-4">Sales Agent</th>
                <th class="px-5 py-4">Editor</th>
                <th class="px-5 py-4">Operator</th>
                <th class="px-5 py-4">Ukuran</th>
                <th class="px-5 py-4">Qty</th>
                <th class="px-5 py-4">Harga</th>
                <th class="px-5 py-4">Tanggal Order</th>
                <th class="px-5 py-4">Status</th>
                <th class="px-5 py-4">Foto Pasang</th>
                <th class="px-5 py-4">Tanggal Pasang</th>
                <th class="px-5 py-4">Desain</th>
                <th class="px-5 py-4">Selesai Desain</th>
                <th class="px-5 py-4 text-center">Aksi</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr
                v-for="b in brandingBanners"
                :key="b.id"
                class="hover:bg-slate-50/60 transition-colors group"
              >
                <td class="px-5 py-4">
                  <div class="flex items-center gap-2.5">
                    <!-- Foto Toko / Fallback Avatar -->
                    <div
                      v-if="b.shop_photo"
                      class="relative w-9 h-9 rounded-xl overflow-hidden bg-slate-100 border border-slate-200/80 shrink-0 cursor-pointer group/shop hover:ring-2 hover:ring-primary/40 transition-all shadow-sm"
                      @click="openPhotoPreview(b.shop_photo, `Foto Toko - ${b.shop_name || 'Outlet'}`)"
                      title="Klik untuk pratinjau foto toko"
                    >
                      <img
                        :src="photoUrl(b.shop_photo)!"
                        :alt="b.shop_name"
                        class="w-full h-full object-cover group-hover/shop:scale-110 transition-transform duration-300"
                      />
                      <div class="absolute inset-0 bg-black/20 opacity-0 group-hover/shop:opacity-100 transition-opacity flex items-center justify-center">
                        <Eye class="w-3.5 h-3.5 text-white drop-shadow" />
                      </div>
                    </div>
                    <div
                      v-else
                      class="w-9 h-9 rounded-xl bg-slate-100 border border-slate-200/60 flex items-center justify-center shrink-0 text-slate-400"
                      title="Foto toko belum tersedia"
                    >
                      <Store class="w-4 h-4 text-slate-400" />
                    </div>

                    <div class="min-w-0 max-w-[200px]">
                      <p class="font-black text-slate-800 text-sm leading-tight truncate">
                        {{ b.shop_name || `Shop #${b.shop_id}` }}
                      </p>
                      <!-- Hover Card Alamat Toko -->
                      <StoreAddressHoverCard
                        :address="b.shop_address"
                        :shop-name="b.shop_name || `Shop #${b.shop_id}`"
                        :province="b.province"
                        :regency="b.regency"
                        :district="b.district"
                        :village="b.village"
                        :phone="b.shop_phone"
                        :latitude="b.latitude"
                        :longitude="b.longitude"
                        :sales-name="b.sales_name"
                        :sales-branch="b.sales_branch"
                        class="mt-0.5 text-[11px]"
                      />
                      <p
                        v-if="b.notes"
                        class="text-[10px] text-slate-400 font-medium line-clamp-1 mt-0.5"
                      >
                        {{ b.notes }}
                      </p>
                      <!-- Lampiran Dokumen Pendukung -->
                      <div v-if="b.documentation_image_url" class="mt-1 flex items-center gap-1">
                        <template v-if="b.documentation_image_url.toLowerCase().endsWith('.pdf')">
                          <a
                            :href="photoUrl(b.documentation_image_url)!"
                            target="_blank"
                            class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded bg-rose-50 text-rose-700 border border-rose-200 text-[9px] font-extrabold hover:bg-rose-100 transition-colors"
                            title="Buka Dokumen PDF Pendukung"
                          >
                            <FileText class="w-2.5 h-2.5" /> PDF
                          </a>
                        </template>
                        <template v-else>
                          <button
                            type="button"
                            @click.stop="openPhotoPreview(b.documentation_image_url, `Dokumen/Foto Pendukung - ${b.shop_name}`)"
                            class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded bg-indigo-50 text-indigo-700 border border-indigo-200 text-[9px] font-extrabold hover:bg-indigo-100 transition-colors cursor-pointer"
                            title="Lihat Foto/Logo Pendukung"
                          >
                            <Paperclip class="w-2.5 h-2.5" /> Lampiran
                          </button>
                        </template>
                      </div>
                    </div>
                  </div>
                </td>
                <td class="px-5 py-4">
                  <div class="flex flex-wrap items-center gap-1.5">
                    <span
                      :class="[
                        'px-2.5 py-1 rounded-full text-[10px] font-black uppercase tracking-wider border inline-flex items-center gap-1',
                        (b.type || '').toLowerCase() === 'universal'
                          ? 'bg-sky-50 text-sky-700 border-sky-200'
                          : 'bg-indigo-50 text-indigo-700 border-indigo-200'
                      ]"
                    >
                      {{ b.type || 'Custom' }}
                    </span>
                    <span
                      v-if="b.is_bihun"
                      class="px-2 py-0.5 rounded-full text-[10px] font-black uppercase tracking-wider bg-amber-50 text-amber-700 border border-amber-200 shadow-xs inline-flex items-center gap-1"
                    >
                      🍜 Bihun
                    </span>
                  </div>
                </td>
                <td class="px-5 py-4">
                  <div class="flex items-center gap-1.5 text-xs font-bold text-slate-600">
                    <Users class="w-3.5 h-3.5 text-slate-400 shrink-0" />
                    {{ b.sales_name || '-' }}
                  </div>
                </td>
                <td class="px-5 py-4 text-xs font-bold text-slate-600">
                  {{ b.editor_name || '-' }}
                </td>
                <td class="px-5 py-4">
                  <div v-if="b.operator_name" class="space-y-0.5">
                    <div class="flex items-center gap-1.5 text-xs font-bold text-slate-700 whitespace-nowrap">
                      <UserCheck class="w-3.5 h-3.5 text-emerald-600 shrink-0" />
                      <span>{{ b.operator_name }}</span>
                    </div>
                    <div v-if="b.operator_phone" class="text-[10px] text-slate-400 font-medium pl-5">
                      {{ b.operator_phone }}
                    </div>
                  </div>
                  <span v-else class="text-xs font-medium text-slate-300 italic">Belum ditugaskan</span>
                </td>
                <td class="px-5 py-4">
                  <div class="space-y-1">
                    <div class="flex items-center gap-1.5 text-xs font-bold text-slate-700">
                      <Ruler class="w-3.5 h-3.5 text-slate-400 shrink-0" />
                      {{ formatBannerDimension(b.size_length, b.size_width, 'dual') }}
                    </div>
                    <div v-if="getSizeSuggestion(b)" class="flex items-center gap-1 text-[10px] font-extrabold text-blue-700 bg-blue-50 px-1.5 py-0.5 rounded border border-blue-200/80 w-fit" title="Sugesti ukuran tim desain by sistem">
                      <Sparkles class="w-3 h-3 text-blue-500 shrink-0" />
                      Desain: {{ getSizeSuggestion(b)?.formattedDesignDual || `${getSizeSuggestion(b)?.designLength}×${getSizeSuggestion(b)?.designWidth} cm` }}
                    </div>
                  </div>
                </td>
                <td class="px-5 py-4 text-sm font-black text-slate-700 text-center">{{ b.qty }}</td>
                <td class="px-5 py-4 text-sm font-black text-slate-800">
                  {{ formatCurrency(b.total_price) }}
                </td>
                <td class="px-5 py-4 text-xs font-bold text-slate-600">
                  {{ formatDate(b.created_at) || '-' }}
                </td>
                <td class="px-5 py-4">
                  <span
                    :class="[
                      'px-2.5 py-1 rounded-full text-[9px] font-black uppercase tracking-widest border',
                      brandingStatusColor(b.status),
                    ]"
                  >
                    {{ brandingStatusLabel(b.status) }}
                  </span>
                </td>
                <td class="px-5 py-4">
                  <button
                    v-if="photoUrl(b.photo_installed)"
                    @click="previewPhoto = photoUrl(b.photo_installed)"
                    class="flex items-center gap-1 text-[11px] font-bold text-primary hover:underline"
                  >
                    <Eye class="w-3.5 h-3.5" /> Lihat
                  </button>
                  <span v-else class="text-[11px] text-slate-300 font-bold">—</span>
                </td>
                <td class="px-5 py-4 text-[11px] font-bold text-slate-600">
                  <span
                    v-if="b.installed_at"
                    class="bg-emerald-50 text-emerald-700 px-2 py-0.5 rounded border border-emerald-100/80 inline-block whitespace-nowrap"
                  >
                    {{ formatDate(b.installed_at) }}
                  </span>
                  <span v-else class="text-slate-300 font-bold">—</span>
                </td>
                <td class="px-5 py-4">
                  <a
                    v-if="b.design_link"
                    :href="b.design_link"
                    target="_blank"
                    class="flex items-center gap-1 text-[11px] font-bold text-violet-600 hover:underline"
                  >
                    <ExternalLink class="w-3 h-3" /> Drive
                  </a>
                  <span v-else class="text-[11px] text-slate-300 font-bold">—</span>
                </td>
                <td class="px-5 py-4 text-[11px] font-bold text-slate-600">
                  <span v-if="b.editing_finished_at" class="bg-indigo-50 text-indigo-700 px-2 py-0.5 rounded border border-indigo-100/80 inline-block whitespace-nowrap">
                    {{ formatDate(b.editing_finished_at) }}
                  </span>
                  <span v-else class="text-slate-300 font-bold">—</span>
                </td>
                <td class="px-5 py-4 text-center">
                  <div
                    v-if="canManageBanners"
                    class="flex items-center justify-center gap-1.5"
                  >
                    <button
                      @click="openEditModal(b)"
                      class="p-1.5 bg-white border border-slate-200 text-slate-600 rounded-lg hover:bg-primary hover:text-white hover:border-primary transition-all cursor-pointer shadow-xs"
                      title="Edit Data & Status Spanduk"
                    >
                      <Edit2 class="w-3.5 h-3.5" />
                    </button>
                    <button
                      @click="openDeleteModal(b)"
                      class="p-1.5 bg-white border border-slate-200 text-red-500 rounded-lg hover:bg-red-500 hover:text-white hover:border-red-500 transition-all cursor-pointer shadow-xs"
                      title="Hapus Spanduk"
                    >
                      <Trash2 class="w-3.5 h-3.5" />
                    </button>
                  </div>
                  <span v-else class="text-[10px] text-slate-300 font-bold">—</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Grid Cards View for Branding -->
      <div
        v-else-if="viewMode === 'grid'"
        :class="[
          'grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-5 transition-opacity duration-200',
          isSearching ? 'opacity-60 pointer-events-none' : ''
        ]"
      >
        <div
          v-for="b in brandingBanners"
          :key="b.id"
          class="bg-white rounded-[2rem] border border-slate-200 shadow-sm hover:shadow-lg hover:-translate-y-0.5 transition-all duration-300 overflow-hidden flex flex-col"
        >
          <!-- Dual Photo Panel: Foto Toko (kiri) & Foto Bukti Pasang (kanan) -->
          <div class="relative bg-slate-50 shrink-0">
            <div class="grid grid-cols-2 gap-0">
              <!-- ── Panel Kiri: Foto Toko ───────────────────────────────── -->
              <div
                class="relative h-40 bg-slate-100 overflow-hidden border-r border-slate-200/60 group/shop"
                :class="photoUrl(b.shop_photo) ? 'cursor-pointer' : 'cursor-default'"
                @click="photoUrl(b.shop_photo) && openPhotoPreview(b.shop_photo, `Foto Toko - ${b.shop_name || 'Outlet'}`)"
                :title="photoUrl(b.shop_photo) ? `Pratinjau foto toko ${b.shop_name || ''}` : 'Foto toko belum tersedia'"
              >
                <!-- Label mini -->
                <span class="absolute top-1.5 left-1.5 z-10 text-[8px] font-black uppercase tracking-wider bg-white/85 backdrop-blur-sm text-slate-600 px-1.5 py-0.5 rounded-md border border-slate-200/60 shadow-sm">
                  Toko
                </span>
                <!-- Foto Toko ada -->
                <img
                  v-if="photoUrl(b.shop_photo)"
                  :src="photoUrl(b.shop_photo)!"
                  :alt="b.shop_name"
                  class="w-full h-full object-cover group-hover/shop:scale-105 transition-transform duration-500"
                />
                <!-- Foto Toko belum ada -->
                <div
                  v-else
                  class="w-full h-full flex flex-col items-center justify-center text-slate-300 gap-1.5"
                >
                  <Store class="w-8 h-8 opacity-30" />
                  <span class="text-[9px] font-bold uppercase tracking-widest opacity-35">Belum Ada</span>
                </div>
                <!-- Hover zoom overlay -->
                <div
                  v-if="photoUrl(b.shop_photo)"
                  class="absolute inset-0 bg-black/20 opacity-0 group-hover/shop:opacity-100 transition-opacity flex items-center justify-center"
                >
                  <Eye class="w-5 h-5 text-white drop-shadow" />
                </div>
              </div>

              <!-- ── Panel Kanan: Foto Bukti Pasang ─────────────────────── -->
              <div
                class="relative h-40 bg-slate-100 overflow-hidden group/proof"
                :class="photoUrl(b.photo_installed) ? 'cursor-pointer' : 'cursor-default'"
                @click="photoUrl(b.photo_installed) && openPhotoPreview(b.photo_installed, `Bukti Pasang - ${b.shop_name || 'Outlet'}`)"
                :title="photoUrl(b.photo_installed) ? `Pratinjau bukti pemasangan ${b.shop_name || ''}` : 'Bukti pemasangan belum diunggah'"
              >
                <!-- Label mini -->
                <span class="absolute top-1.5 left-1.5 z-10 text-[8px] font-black uppercase tracking-wider bg-white/85 backdrop-blur-sm text-slate-600 px-1.5 py-0.5 rounded-md border border-slate-200/60 shadow-sm">
                  Bukti
                </span>
                <!-- Bukti Pasang → Google Drive -->
                <div
                  v-if="photoUrl(b.photo_installed) && photoUrl(b.photo_installed)!.includes('drive.google.com')"
                  class="w-full h-full flex flex-col items-center justify-center bg-indigo-50/70 text-indigo-600 gap-1.5"
                >
                  <Image class="w-7 h-7 text-indigo-400" />
                  <span class="text-[9px] font-bold text-indigo-500">Google Drive</span>
                  <span class="text-[8px] text-indigo-400 font-semibold">Klik utk lihat</span>
                </div>
                <!-- Bukti Pasang → gambar biasa -->
                <img
                  v-else-if="photoUrl(b.photo_installed)"
                  :src="photoUrl(b.photo_installed)!"
                  alt="Bukti Pasang"
                  class="w-full h-full object-cover group-hover/proof:scale-105 transition-transform duration-500"
                />
                <!-- Bukti Pasang belum ada -->
                <div
                  v-else
                  class="w-full h-full flex flex-col items-center justify-center text-slate-300 gap-1.5"
                >
                  <Image class="w-8 h-8 opacity-30" />
                  <span class="text-[9px] font-bold uppercase tracking-widest opacity-35">Belum Pasang</span>
                </div>
                <!-- Hover zoom overlay -->
                <div
                  v-if="photoUrl(b.photo_installed)"
                  class="absolute inset-0 bg-black/20 opacity-0 group-hover/proof:opacity-100 transition-opacity flex items-center justify-center"
                >
                  <Eye class="w-5 h-5 text-white drop-shadow" />
                </div>
              </div>
            </div>

            <!-- Badge Status (absolute di atas grid, z-20) -->
            <span
              :class="[
                'absolute top-3 left-3 z-20 px-3 py-1 rounded-xl text-[9px] font-black border backdrop-blur-md',
                brandingStatusColor(b.status),
              ]"
            >
              {{ brandingStatusLabel(b.status) }}
            </span>

            <!-- Tombol Aksi (absolute di atas grid, z-20) -->
            <div
              v-if="canManageBanners"
              class="absolute top-3 right-3 z-20 flex gap-1.5"
            >
              <button
                @click="openEditModal(b)"
                class="p-2 bg-white/80 backdrop-blur-md text-slate-700 rounded-xl hover:bg-primary hover:text-white transition-all shadow-sm cursor-pointer"
                title="Edit Data & Status Spanduk"
              >
                <Edit2 class="w-4 h-4" />
              </button>
              <button
                @click="openDeleteModal(b)"
                class="p-2 bg-white/80 backdrop-blur-md text-red-500 rounded-xl hover:bg-red-500 hover:text-white transition-all shadow-sm cursor-pointer"
                title="Hapus Spanduk"
              >
                <Trash2 class="w-4 h-4" />
              </button>
            </div>
          </div>
          <div class="p-5 flex-1 flex flex-col gap-3">
            <div>
              <div class="flex items-center justify-between gap-2">
                <p class="text-[9px] font-black text-primary uppercase tracking-widest">
                  Banner Spanduk
                </p>
                <div class="flex items-center gap-1.5">
                  <span
                    :class="[
                      'px-2 py-0.5 rounded-full text-[9px] font-black uppercase tracking-wider border',
                      (b.type || '').toLowerCase() === 'universal'
                        ? 'bg-sky-50 text-sky-700 border-sky-200'
                        : 'bg-indigo-50 text-indigo-700 border-indigo-200'
                    ]"
                  >
                    {{ b.type || 'Custom' }}
                  </span>
                  <span
                    v-if="b.is_bihun"
                    class="px-2 py-0.5 rounded-full text-[9px] font-black uppercase tracking-wider bg-amber-50 text-amber-700 border border-amber-200"
                  >
                    🍜 Bihun
                  </span>
                </div>
              </div>
              <div class="flex items-center gap-2.5 mt-1.5">
                <div
                  v-if="b.shop_photo"
                  class="relative w-8 h-8 rounded-lg overflow-hidden bg-slate-100 border border-slate-200 shrink-0 cursor-pointer hover:ring-2 hover:ring-primary/40 transition-all shadow-sm"
                  @click="openPhotoPreview(b.shop_photo, `Foto Toko - ${b.shop_name || 'Outlet'}`)"
                  title="Klik untuk pratinjau foto toko"
                >
                  <img
                    :src="photoUrl(b.shop_photo)!"
                    :alt="b.shop_name"
                    class="w-full h-full object-cover"
                  />
                </div>
                <div class="min-w-0 flex-1">
                  <h3 class="font-black text-slate-900 leading-tight truncate">
                    {{ b.shop_name || `Shop #${b.shop_id}` }}
                  </h3>
                  <!-- Hover Card Alamat Toko -->
                  <StoreAddressHoverCard
                    :address="b.shop_address"
                    :shop-name="b.shop_name || `Shop #${b.shop_id}`"
                    :province="b.province"
                    :regency="b.regency"
                    :district="b.district"
                    :village="b.village"
                    :phone="b.shop_phone"
                    :latitude="b.latitude"
                    :longitude="b.longitude"
                    :sales-name="b.sales_name"
                    :sales-branch="b.sales_branch"
                    class="mt-0.5 text-[11px]"
                  />
                </div>
              </div>
              <p v-if="b.notes" class="text-xs text-slate-500 mt-1 line-clamp-2">{{ b.notes }}</p>
            </div>
            <div class="space-y-1 mt-1">
              <div class="flex items-center justify-between text-[11px] font-bold text-slate-500">
                <div class="flex items-center gap-2">
                  <Users class="w-3.5 h-3.5 text-slate-400 shrink-0" />
                  <span
                    >Sales: <strong class="text-slate-700">{{ b.sales_name || '-' }}</strong></span
                  >
                </div>
                <span
                  >Editor: <strong class="text-slate-700">{{ b.editor_name || '-' }}</strong></span
                >
              </div>
              <div class="flex items-center justify-between text-[11px] font-bold text-slate-500">
                <div class="flex items-center gap-1.5 truncate">
                  <UserCheck class="w-3.5 h-3.5 text-emerald-600 shrink-0" />
                  <span class="truncate">Operator: <strong class="text-slate-700">{{ b.operator_name || 'Belum ditugaskan' }}</strong></span>
                </div>
                <span v-if="b.operator_phone" class="text-[10px] text-slate-400 font-medium shrink-0">
                  {{ b.operator_phone }}
                </span>
              </div>
              <div class="flex items-center gap-2 text-[11px] font-bold text-slate-500">
                <Ruler class="w-3.5 h-3.5 text-slate-400 shrink-0" />
                <span>Ukuran:
                  <strong class="text-slate-700">
                    {{ formatBannerDimension(b.size_length, b.size_width, 'dual') }} (Qty: {{ b.qty }})
                  </strong>
                </span>
              </div>
              <!-- Dokumen Pendukung Row -->
              <div v-if="b.documentation_image_url" class="flex items-center justify-between gap-2 pt-1">
                <span class="text-[10px] font-bold text-slate-400 flex items-center gap-1">
                  <Paperclip class="w-3 h-3 text-indigo-500" /> Dokumen:
                </span>
                <template v-if="b.documentation_image_url.toLowerCase().endsWith('.pdf')">
                  <a
                    :href="photoUrl(b.documentation_image_url)!"
                    target="_blank"
                    class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md bg-rose-50 text-rose-700 border border-rose-200 text-[10px] font-extrabold hover:bg-rose-100 transition-colors"
                  >
                    <FileText class="w-3 h-3" /> Buka PDF
                  </a>
                </template>
                <template v-else>
                  <button
                    type="button"
                    @click.stop="openPhotoPreview(b.documentation_image_url, `Dokumen/Foto Pendukung - ${b.shop_name}`)"
                    class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md bg-indigo-50 text-indigo-700 border border-indigo-200 text-[10px] font-extrabold hover:bg-indigo-100 transition-colors cursor-pointer"
                  >
                    <Eye class="w-3 h-3" /> Lihat Foto
                  </button>
                </template>
              </div>
            </div>
            <div class="flex items-center justify-between mt-auto pt-3 border-t border-slate-100">
              <div class="flex items-center gap-2 flex-wrap">
                <span class="text-[10px] font-black text-slate-800">{{
                  formatCurrency(b.total_price)
                }}</span>
                <span
                  v-if="b.installed_at"
                  class="text-[10px] font-bold text-emerald-700 bg-emerald-50 px-1.5 py-0.5 rounded border border-emerald-100"
                  title="Tanggal Pasang"
                >
                  Pasang: {{ formatDate(b.installed_at) }}
                </span>
              </div>
              <a
                v-if="b.design_link"
                :href="b.design_link"
                target="_blank"
                class="flex items-center gap-1 text-[11px] font-bold text-violet-600 hover:underline shrink-0"
              >
                <ExternalLink class="w-3 h-3" /> Desain
              </a>
            </div>
          </div>
        </div>
      </div>

      <!-- Pagination Controls -->
      <div
        v-if="totalBanners > 0"
        class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 sm:gap-4 bg-white p-3.5 sm:p-4 rounded-2xl border border-slate-200 mt-4 shadow-sm"
      >
        <div class="flex items-center justify-between sm:justify-start gap-2 w-full sm:w-auto">
          <span class="text-xs font-bold text-slate-500">Baris per hal:</span>
          <select
            v-model="itemsPerPage"
            @change="currentPage = 1; fetchBrandingBanners()"
            class="bg-slate-50 border border-slate-200 rounded-lg text-xs font-bold px-2 py-1 outline-none cursor-pointer"
          >
            <option :value="10">10</option>
            <option :value="20">20</option>
            <option :value="50">50</option>
            <option :value="100">100</option>
          </select>
        </div>
        <span class="text-xs font-bold text-slate-500 text-center sm:text-left">
          Menampilkan {{ (currentPage - 1) * itemsPerPage + 1 }} -
          {{ Math.min(currentPage * itemsPerPage, totalBanners) }} dari
          {{ totalBanners }} data
        </span>
        <div class="flex items-center justify-center gap-2 w-full sm:w-auto">
          <button
            @click="prevPage"
            :disabled="currentPage === 1"
            class="px-3 sm:px-4 py-1.5 sm:py-2 rounded-xl text-xs font-black border border-slate-200 disabled:opacity-50 hover:bg-slate-50 transition-colors cursor-pointer"
          >
            Prev
          </button>
          <span class="text-xs font-black text-slate-700 px-2"
            >Hal {{ currentPage }} / {{ totalPages }}</span
          >
          <button
            @click="nextPage"
            :disabled="currentPage === totalPages"
            class="px-3 sm:px-4 py-1.5 sm:py-2 rounded-xl text-xs font-black border border-slate-200 disabled:opacity-50 hover:bg-slate-50 transition-colors cursor-pointer"
          >
            Next
          </button>
        </div>
      </div>

      <!-- Footer info (Offline/Online source) -->
      <div v-if="brandingBanners.length > 0" class="mt-4 flex justify-end">
        <span class="text-[10px] font-bold text-slate-400 flex items-center gap-1.5">
          <CheckCircle2 class="w-3 h-3 text-emerald-500" />
          Sumber: <strong class="uppercase">{{ brandingSource }}</strong> ·
          database branding terpisah
        </span>
      </div>
    </div>

    <!-- ── TAB: INTERNAL HRIS ──────────────────────────────────────────── -->
    <div v-if="activeTab === 'internal'">
      <div v-if="isInternalLoading" class="py-20 flex justify-center">
        <span
          class="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin"
        ></span>
      </div>

      <div
        v-else-if="internalBanners.length === 0"
        class="bg-white rounded-[2.5rem] border border-slate-200 p-20 text-center shadow-sm"
      >
        <Package class="w-16 h-16 text-slate-200 mx-auto mb-4" />
        <h3 class="text-xl font-black text-slate-900">Belum Ada Order Internal</h3>
      </div>

      <div v-else class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-5">
        <div
          v-for="b in internalBanners"
          :key="b.id"
          class="bg-white rounded-[2rem] border border-slate-200 shadow-sm hover:shadow-lg hover:-translate-y-0.5 transition-all duration-300 overflow-hidden flex flex-col"
        >
          <div class="relative h-44 bg-slate-100 shrink-0">
            <img v-if="b.image" :src="b.image" class="w-full h-full object-cover" />
            <div v-else class="w-full h-full flex items-center justify-center text-slate-200">
              <Image class="w-12 h-12" />
            </div>
            <span
              :class="[
                'absolute top-3 left-3 px-3 py-1 rounded-xl text-[9px] font-black border backdrop-blur-md',
                internalStatusColor(b.status),
              ]"
            >
              {{ b.status.replace('_', ' ') }}
            </span>
          </div>
          <div class="p-5 flex-1 flex flex-col gap-3">
            <div>
              <p class="text-[9px] font-black text-primary uppercase tracking-widest">
                {{ b.type }}
              </p>
              <h3 class="font-black text-slate-900 leading-tight mt-0.5">{{ b.title }}</h3>
            </div>
            <div class="flex items-center gap-2 text-xs text-slate-500 font-medium">
              <MapPin class="w-3.5 h-3.5 text-slate-400 shrink-0" />
              <span class="truncate">{{ b.location }}</span>
            </div>
            <div class="flex items-center justify-between mt-auto pt-3 border-t border-slate-100">
              <span class="text-[10px] font-bold text-slate-400">{{ b.date }}</span>
              <a
                v-if="b.designURL"
                :href="b.designURL"
                target="_blank"
                class="flex items-center gap-1 text-[11px] font-bold text-violet-600 hover:underline"
              >
                <ExternalLink class="w-3 h-3" /> Desain
              </a>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Photo Preview Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-if="previewPhoto"
          class="fixed inset-0 z-[110] flex items-center justify-center p-4"
          @click.self="previewPhoto = null"
        >
          <div
            class="absolute inset-0 bg-black/80 backdrop-blur-sm"
            @click="previewPhoto = null"
          ></div>
          <div class="relative z-10 max-w-3xl w-full bg-slate-900 rounded-[2rem] overflow-hidden shadow-2xl flex flex-col border border-white/10">
            <!-- Modal Header -->
            <div class="px-6 py-4 bg-slate-900/90 border-b border-white/10 flex items-center justify-between text-white">
              <div class="flex items-center gap-2.5">
                <Store class="w-4 h-4 text-primary" />
                <h4 class="font-extrabold text-sm text-slate-100 truncate">{{ previewTitle }}</h4>
              </div>
              <div class="flex items-center gap-2">
                <a
                  :href="previewPhoto"
                  target="_blank"
                  class="p-2 hover:bg-white/10 rounded-xl text-slate-300 hover:text-white transition-colors text-xs font-bold flex items-center gap-1.5"
                  title="Buka foto asli"
                >
                  <ExternalLink class="w-4 h-4" />
                </a>
                <button
                  @click="previewPhoto = null"
                  class="p-2 hover:bg-white/10 rounded-xl text-slate-300 hover:text-white transition-colors"
                >
                  <X class="w-5 h-5" />
                </button>
              </div>
            </div>

            <!-- Modal Content -->
            <div class="p-4 flex items-center justify-center bg-black/40 min-h-[300px]">
              <iframe
                v-if="previewPhoto.includes('drive.google.com')"
                :src="previewPhoto"
                class="w-full h-[65vh] rounded-xl border-0 bg-white"
                allow="autoplay"
                allowfullscreen
              ></iframe>
              <img
                v-else
                :src="previewPhoto"
                class="w-full rounded-xl shadow-lg object-contain max-h-[75vh]"
              />
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>

  <!-- Modal Edit Banner -->
  <div v-if="isEditModalOpen" class="fixed inset-0 z-[100] flex items-center justify-center p-4">
    <div
      class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
      @click="isEditModalOpen = false"
    ></div>
    <div
      class="bg-white rounded-[32px] w-full max-w-lg relative z-10 shadow-2xl flex flex-col max-h-[90vh] overflow-hidden animate-in fade-in zoom-in duration-300"
    >
      <div
        class="p-6 border-b border-slate-100 flex items-center justify-between bg-slate-50/50 shrink-0"
      >
        <div>
          <h3 class="text-xl font-black text-slate-900">Edit Data Spanduk</h3>
          <p class="text-xs font-bold text-slate-400 mt-1">Perbarui detail permintaan spanduk</p>
        </div>
        <button
          @click="isEditModalOpen = false"
          class="p-3 hover:bg-white rounded-2xl transition-all text-slate-400 hover:text-slate-900 shadow-sm"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <div class="p-6 overflow-y-auto space-y-5 flex-1">
        <div>
          <label class="block text-xs font-bold text-slate-500 mb-2">Salesman *</label>
          <SearchSelect
            :model-value="currentBanner.sales_name"
            :options="salesSelectOptions"
            placeholder="Pilih atau cari Salesman..."
            search-placeholder="Ketik nama atau jabatan sales..."
            empty-message="Salesman tidak ditemukan"
            clearable
            @update:model-value="onSalesChangeEdit"
            @change="onSalesChangeEdit"
          />
        </div>
        <div>
          <div class="flex items-center justify-between mb-2">
            <label class="block text-xs font-bold text-slate-500">Nama Toko *</label>
            <span v-if="currentBanner.sales_name && shopSelectOptionsForEdit.length > 0" class="text-[10px] font-bold text-emerald-600 bg-emerald-50 px-2 py-0.5 rounded-full border border-emerald-200">
              {{ shopSelectOptionsForEdit.length }} Toko Tersedia
            </span>
          </div>
          <SearchSelect
            v-model="currentBanner.shop_name"
            :options="shopSelectOptionsForEdit"
            :disabled="!currentBanner.sales_name"
            :placeholder="currentBanner.sales_name ? 'Pilih atau cari Toko binaan sales...' : 'Pilih Salesman terlebih dahulu...'"
            :search-placeholder="currentBanner.sales_name ? 'Ketik nama toko atau alamat...' : 'Pilih salesman dahulu'"
            allow-custom
            custom-label-prefix="Pakai Nama Toko: "
            :empty-message="currentBanner.sales_name ? 'Toko tidak ditemukan. Anda dapat mengetik untuk memakai nama toko baru.' : 'Pilih Salesman terlebih dahulu untuk menampilkan daftar toko.'"
            @select="onSelectShopEdit"
            @custom-input="currentBanner.shop_id = ''"
          />
          <!-- Loading indicator when fetching shops -->
          <div
            v-if="currentBanner.sales_name && isShopsBySalesLoading"
            class="mt-2 p-2.5 bg-blue-50/80 border border-blue-200/80 rounded-xl flex items-center gap-2 text-xs font-medium text-blue-700 animate-in fade-in duration-200"
          >
            <Loader2 class="w-4 h-4 text-blue-600 animate-spin shrink-0" />
            <span>Memuat seluruh toko binaan <strong>{{ currentBanner.sales_name }}</strong>...</span>
          </div>
          <!-- Hint when sales is not selected -->
          <div
            v-else-if="!currentBanner.sales_name"
            class="mt-2 p-2.5 bg-amber-50/80 border border-amber-200/80 rounded-xl flex items-start sm:items-center gap-2 text-xs font-medium text-amber-800 animate-in fade-in duration-200"
          >
            <AlertCircle class="w-4 h-4 text-amber-600 shrink-0 mt-0.5 sm:mt-0" />
            <span>Pilih <strong>Salesman</strong> terlebih dahulu untuk menampilkan daftar toko yang menjadi tanggung jawabnya.</span>
          </div>
          <p v-else class="text-[11px] font-medium text-slate-400 mt-1.5 flex items-center gap-1">
            <span>💡 Menampilkan {{ shopSelectOptionsForEdit.length }} toko binaan <strong>{{ currentBanner.sales_name }}</strong>. Bisa ketik nama baru jika toko belum terdaftar.</span>
          </p>
        </div>
        <div class="grid grid-cols-2 gap-3 sm:gap-4">
          <div>
            <label class="block text-xs font-bold text-slate-500 mb-2">Tipe Spanduk</label>
            <select
              v-model="currentBanner.type"
              class="w-full bg-slate-50 border-none rounded-xl sm:rounded-2xl px-3 sm:px-4 py-3 sm:py-3.5 text-xs sm:text-sm font-bold text-slate-700 focus:ring-2 focus:ring-primary/20 cursor-pointer"
            >
              <option value="Custom">Custom (Desain Khusus Outlet)</option>
              <option value="Universal">Universal (Desain Standar Pabrik)</option>
            </select>
          </div>
          <div>
            <label class="block text-xs font-bold text-slate-500 mb-2">Bihun?</label>
            <select
              v-model="currentBanner.is_bihun"
              class="w-full bg-slate-50 border-none rounded-xl sm:rounded-2xl px-3 sm:px-4 py-3 sm:py-3.5 text-xs sm:text-sm font-bold text-slate-700 focus:ring-2 focus:ring-primary/20"
            >
              <option :value="true">Ya, Bihun</option>
              <option :value="false">Tidak</option>
            </select>
          </div>
        </div>
        <!-- Ukuran Dimensi & Satuan -->
        <div>
          <div class="flex items-center justify-between mb-2">
            <label class="block text-xs font-bold text-slate-700">Ukuran Dimensi</label>
            <div class="inline-flex p-0.5 bg-slate-100 rounded-lg border border-slate-200">
              <button
                type="button"
                @click="setEditBannerUnit('cm')"
                :class="[
                  'px-2.5 py-0.5 text-[11px] font-bold rounded-md transition-all',
                  currentBannerUnit === 'cm'
                    ? 'bg-white text-indigo-600 shadow-xs'
                    : 'text-slate-500 hover:text-slate-800'
                ]"
              >
                cm
              </button>
              <button
                type="button"
                @click="setEditBannerUnit('m')"
                :class="[
                  'px-2.5 py-0.5 text-[11px] font-bold rounded-md transition-all',
                  currentBannerUnit === 'm'
                    ? 'bg-white text-indigo-600 shadow-xs'
                    : 'text-slate-500 hover:text-slate-800'
                ]"
              >
                meter (m)
              </button>
            </div>
          </div>
          <div class="grid grid-cols-3 gap-2 sm:gap-4">
            <div>
              <label class="block text-[11px] sm:text-xs font-bold text-slate-500 mb-1.5 sm:mb-2 truncate">
                Panjang ({{ currentBannerUnit }})
              </label>
              <input
                v-model.number="currentBanner.size_length"
                type="number"
                :step="currentBannerUnit === 'm' ? '0.1' : '1'"
                class="w-full bg-slate-50 border-none rounded-xl sm:rounded-2xl px-2.5 sm:px-4 py-2.5 sm:py-3.5 text-xs sm:text-sm font-bold text-slate-700 focus:ring-2 focus:ring-primary/20"
              />
            </div>
            <div>
              <label class="block text-[11px] sm:text-xs font-bold text-slate-500 mb-1.5 sm:mb-2 truncate">
                Lebar ({{ currentBannerUnit }})
              </label>
              <input
                v-model.number="currentBanner.size_width"
                type="number"
                :step="currentBannerUnit === 'm' ? '0.1' : '1'"
                class="w-full bg-slate-50 border-none rounded-xl sm:rounded-2xl px-2.5 sm:px-4 py-2.5 sm:py-3.5 text-xs sm:text-sm font-bold text-slate-700 focus:ring-2 focus:ring-primary/20"
              />
            </div>
            <div>
              <label class="block text-[11px] sm:text-xs font-bold text-slate-500 mb-1.5 sm:mb-2 truncate">Qty</label>
              <input
                v-model.number="currentBanner.qty"
                type="number"
                min="1"
                class="w-full bg-slate-50 border-none rounded-xl sm:rounded-2xl px-2.5 sm:px-4 py-2.5 sm:py-3.5 text-xs sm:text-sm font-bold text-slate-700 focus:ring-2 focus:ring-primary/20"
              />
            </div>
          </div>
          <p v-if="currentBanner.size_length && currentBanner.size_width" class="text-[11px] font-semibold text-indigo-600 mt-1.5 flex items-center gap-1">
            <Ruler class="w-3.5 h-3.5" />
            Setara: {{ formatBannerDimension(currentBannerUnit === 'm' ? currentBanner.size_length * 100 : currentBanner.size_length, currentBannerUnit === 'm' ? currentBanner.size_width * 100 : currentBanner.size_width, 'dual') }}
            • Luas: {{ ((currentBannerUnit === 'm' ? currentBanner.size_length : currentBanner.size_length / 100) * (currentBannerUnit === 'm' ? currentBanner.size_width : currentBanner.size_width / 100)).toFixed(2) }} m²
          </p>
        </div>
        <!-- Status & Editor Controls -->
        <div class="p-3.5 sm:p-4 rounded-2xl bg-indigo-50/60 border border-indigo-100 space-y-3 sm:space-y-4">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 sm:gap-4">
            <div>
              <label class="block text-[11px] sm:text-xs font-black text-indigo-900 uppercase tracking-wider mb-1.5 sm:mb-2">
                Status Progress Banner *
              </label>
              <select
                v-model="currentBanner.status"
                class="w-full bg-white border border-indigo-200 rounded-xl sm:rounded-2xl px-3 sm:px-4 py-2.5 sm:py-3 text-xs sm:text-sm font-bold text-slate-800 focus:ring-2 focus:ring-primary/20 cursor-pointer shadow-xs"
              >
                <option value="request">Antrian Desain (Request Baru)</option>
                <option value="editing">Sedang Dikerjakan (Desain)</option>
                <option value="validation">Menunggu Validasi Desain</option>
                <option value="sent_to_vendor">Di Vendor / Percetakan</option>
                <option value="ready_for_operator">Siap Pasang (Di Operator)</option>
                <option value="installed">Terpasang (Selesai)</option>
              </select>
            </div>
            <div>
              <label class="block text-[11px] sm:text-xs font-black text-indigo-900 uppercase tracking-wider mb-1.5 sm:mb-2">
                Editor / Desainer
              </label>
              <select
                v-model="currentBanner.editor_id"
                class="w-full bg-white border border-indigo-200 rounded-xl sm:rounded-2xl px-3 sm:px-4 py-2.5 sm:py-3 text-xs sm:text-sm font-bold text-slate-800 focus:ring-2 focus:ring-primary/20 cursor-pointer shadow-xs"
              >
                <option value="">-- Belum Ditugaskan --</option>
                <option v-for="ed in editors" :key="ed.id" :value="ed.id">
                  {{ ed.name }} {{ ed.area_code ? `(${ed.area_code})` : '' }}
                </option>
              </select>
            </div>
          </div>
        </div>

        <!-- ── Dynamic Section: Di Vendor / Percetakan ────────────────────── -->
        <div
          v-if="['sent_to_vendor', 'purchasing_queue', 'percetakan', 'vendor'].includes(currentBanner.status)"
          class="p-3.5 sm:p-4 rounded-2xl bg-amber-50/70 border border-amber-200/80 space-y-3 sm:space-y-4 animate-in fade-in duration-200"
        >
          <div class="flex items-center gap-2 text-amber-900 font-extrabold text-xs uppercase tracking-wider">
            <Package class="w-4 h-4 text-amber-600" />
            Data Percetakan & Vendor
          </div>
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 sm:gap-4">
            <div>
              <label class="block text-xs font-bold text-slate-600 mb-1.5">Pilih Vendor Percetakan</label>
              <select
                v-model="currentBanner.vendor_id"
                class="w-full bg-white border border-amber-200 rounded-xl sm:rounded-2xl px-3 sm:px-4 py-2.5 sm:py-3 text-xs sm:text-sm font-bold text-slate-800 focus:ring-2 focus:ring-primary/20 cursor-pointer shadow-xs"
              >
                <option value="">-- Pilih Vendor --</option>
                <option v-for="v in vendors" :key="v.id" :value="v.id">
                  {{ v.name }}
                </option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-bold text-slate-600 mb-1.5">Tanggal SPK Cetak</label>
              <input
                v-model="currentBanner.spk_date"
                type="date"
                class="w-full bg-white border border-amber-200 rounded-xl sm:rounded-2xl px-3 sm:px-4 py-2.5 sm:py-3 text-xs sm:text-sm font-bold text-slate-800 focus:ring-2 focus:ring-primary/20 shadow-xs"
              />
            </div>
          </div>
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 sm:gap-4">
            <div>
              <label class="block text-xs font-bold text-slate-600 mb-1.5">Total Biaya Cetak (Rp)</label>
              <input
                v-model.number="currentBanner.total_price"
                type="number"
                placeholder="Contoh: 150000"
                class="w-full bg-white border border-amber-200 rounded-xl sm:rounded-2xl px-3 sm:px-4 py-2.5 sm:py-3 text-xs sm:text-sm font-bold text-slate-800 focus:ring-2 focus:ring-primary/20 shadow-xs"
              />
            </div>
            <div>
              <label class="block text-xs font-bold text-slate-600 mb-1.5">Link File Desain Cetak (GDrive)</label>
              <input
                v-model="currentBanner.design_link"
                type="text"
                placeholder="https://drive.google.com/..."
                class="w-full bg-white border border-amber-200 rounded-xl sm:rounded-2xl px-3 sm:px-4 py-2.5 sm:py-3 text-xs sm:text-sm font-bold text-slate-800 focus:ring-2 focus:ring-primary/20 shadow-xs"
              />
            </div>
          </div>
        </div>

        <!-- ── Dynamic Section: Siap Pasang / Di Operator ────────────────── -->
        <div
          v-if="['ready_for_operator', 'siap_pasang', 'installation', 'diterima_operator'].includes(currentBanner.status)"
          class="p-3.5 sm:p-4 rounded-2xl bg-sky-50/70 border border-sky-200/80 space-y-3 sm:space-y-4 animate-in fade-in duration-200"
        >
          <div class="flex items-center gap-2 text-sky-900 font-extrabold text-xs uppercase tracking-wider">
            <Users class="w-4 h-4 text-sky-600" />
            Data Penugasan Operator Lapangan
          </div>
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 sm:gap-4">
            <div>
              <label class="block text-xs font-bold text-slate-600 mb-1.5">Pilih Operator Pemasang</label>
              <select
                v-model="currentBanner.operator_id"
                class="w-full bg-white border border-sky-200 rounded-xl sm:rounded-2xl px-3 sm:px-4 py-2.5 sm:py-3 text-xs sm:text-sm font-bold text-slate-800 focus:ring-2 focus:ring-primary/20 cursor-pointer shadow-xs"
              >
                <option value="">-- Pilih Operator --</option>
                <option v-for="op in operators" :key="op.id" :value="op.id">
                  {{ op.name }} {{ op.branch_name ? `(${op.branch_name})` : '' }}
                </option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-bold text-slate-600 mb-1.5">Upah / Reward Pasang (Rp)</label>
              <input
                v-model.number="currentBanner.operator_reward"
                type="number"
                placeholder="Contoh: 25000"
                class="w-full bg-white border border-sky-200 rounded-xl sm:rounded-2xl px-3 sm:px-4 py-2.5 sm:py-3 text-xs sm:text-sm font-bold text-slate-800 focus:ring-2 focus:ring-primary/20 shadow-xs"
              />
            </div>
          </div>
          <div>
            <label class="block text-xs font-bold text-slate-600 mb-1.5">Tanggal SPK</label>
            <input
              v-model="currentBanner.spk_date"
              type="date"
              class="w-full bg-white border border-sky-200 rounded-xl sm:rounded-2xl px-3 sm:px-4 py-2.5 sm:py-3 text-xs sm:text-sm font-bold text-slate-800 focus:ring-2 focus:ring-primary/20 shadow-xs"
            />
          </div>
        </div>

        <!-- ── Dynamic Section: Terpasang (Selesai) ───────────────────────── -->
        <div
          v-if="['installed', 'terpasang', 'done'].includes(currentBanner.status)"
          class="p-3.5 sm:p-4 rounded-2xl bg-emerald-50/70 border border-emerald-200/80 space-y-3 sm:space-y-4 animate-in fade-in duration-200"
        >
          <div class="flex items-center gap-2 text-emerald-900 font-extrabold text-xs uppercase tracking-wider">
            <CheckCircle2 class="w-4 h-4 text-emerald-600" />
            Data Bukti Pemasangan Selesai
          </div>
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 sm:gap-4">
            <div>
              <label class="block text-xs font-bold text-slate-600 mb-1.5">Petugas Pemasang</label>
              <select
                v-model="currentBanner.operator_id"
                class="w-full bg-white border border-emerald-200 rounded-xl sm:rounded-2xl px-3 sm:px-4 py-2.5 sm:py-3 text-xs sm:text-sm font-bold text-slate-800 focus:ring-2 focus:ring-primary/20 cursor-pointer shadow-xs"
              >
                <option value="">-- Pilih Petugas --</option>
                <option v-for="op in operators" :key="op.id" :value="op.id">
                  {{ op.name }}
                </option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-bold text-slate-600 mb-1.5">Tanggal Terpasang</label>
              <input
                v-model="currentBanner.installed_at"
                type="date"
                class="w-full bg-white border border-emerald-200 rounded-xl sm:rounded-2xl px-3 sm:px-4 py-2.5 sm:py-3 text-xs sm:text-sm font-bold text-slate-800 focus:ring-2 focus:ring-primary/20 shadow-xs"
              />
            </div>
          </div>
          <div>
            <label class="block text-xs font-bold text-slate-600 mb-1.5">Foto Bukti Terpasang (URL / Path)</label>
            <input
              v-model="currentBanner.photo_installed"
              type="text"
              placeholder="Contoh: https://... atau uploads/..."
              class="w-full bg-white border border-emerald-200 rounded-xl sm:rounded-2xl px-3 sm:px-4 py-2.5 sm:py-3 text-xs sm:text-sm font-bold text-slate-800 focus:ring-2 focus:ring-primary/20 shadow-xs"
            />
          </div>
        </div>

        <!-- ── Dynamic Section: Menunggu Validasi ─────────────────────────── -->
        <div
          v-if="['validation', 'val_design', 'val_store'].includes(currentBanner.status)"
          class="p-3.5 sm:p-4 rounded-2xl bg-violet-50/70 border border-violet-200/80 space-y-3 sm:space-y-4 animate-in fade-in duration-200"
        >
          <div class="flex items-center gap-2 text-violet-900 font-extrabold text-xs uppercase tracking-wider">
            <Eye class="w-4 h-4 text-violet-600" />
            Data Validasi Desain
          </div>
          <div>
            <label class="block text-xs font-bold text-slate-600 mb-1.5">Link Desain Google Drive</label>
            <input
              v-model="currentBanner.design_link"
              type="text"
              placeholder="https://drive.google.com/..."
              class="w-full bg-white border border-violet-200 rounded-xl sm:rounded-2xl px-3 sm:px-4 py-2.5 sm:py-3 text-xs sm:text-sm font-bold text-slate-800 focus:ring-2 focus:ring-primary/20 shadow-xs"
            />
          </div>
          <div class="grid grid-cols-2 gap-3 sm:gap-4">
            <label class="flex items-center gap-2 sm:gap-2.5 p-2.5 sm:p-3 bg-white border border-violet-200 rounded-xl sm:rounded-2xl cursor-pointer">
              <input
                v-model="currentBanner.is_design_approved"
                type="checkbox"
                class="w-4 h-4 text-primary rounded"
              />
              <span class="text-xs font-bold text-slate-700">ACC Desain Visual</span>
            </label>
            <label class="flex items-center gap-2 sm:gap-2.5 p-2.5 sm:p-3 bg-white border border-violet-200 rounded-xl sm:rounded-2xl cursor-pointer">
              <input
                v-model="currentBanner.is_store_approved"
                type="checkbox"
                class="w-4 h-4 text-primary rounded"
              />
              <span class="text-xs font-bold text-slate-700">ACC Toko</span>
            </label>
          </div>
        </div>

        <!-- ── Section: Foto / Logo / Dokumen Pendukung ──────────────────── -->
        <div class="p-3.5 sm:p-4 rounded-2xl bg-slate-50 border border-slate-200/80 space-y-3">
          <div class="flex items-center justify-between">
            <label class="block text-xs font-black text-slate-700 uppercase tracking-wider flex items-center gap-1.5">
              <Paperclip class="w-4 h-4 text-indigo-600" />
              Foto / Logo / Dokumen Pendukung
            </label>
            <span class="text-[10px] font-bold text-slate-400 bg-white px-2 py-0.5 rounded-full border border-slate-200">
              Opsional
            </span>
          </div>

          <!-- Case 1: Ada file baru yang dipilih (belum disimpan) -->
          <div
            v-if="editAttachmentFile"
            class="p-3 bg-white rounded-2xl border border-indigo-200 flex items-center gap-3 shadow-xs"
          >
            <!-- Preview: Gambar -->
            <div
              v-if="!isEditPdfAttachment && editAttachmentPreview"
              class="w-12 h-12 rounded-xl overflow-hidden bg-slate-100 border border-slate-200 shrink-0 cursor-pointer"
              @click="openPhotoPreview(editAttachmentPreview, 'File Baru')"
            >
              <img :src="editAttachmentPreview" class="w-full h-full object-cover" />
            </div>
            <!-- Preview: PDF -->
            <div
              v-else
              class="w-12 h-12 rounded-xl bg-rose-50 border border-rose-200 flex items-center justify-center shrink-0 text-rose-600"
            >
              <FileText class="w-6 h-6" />
            </div>
            <div class="min-w-0 flex-1">
              <p class="text-xs font-black text-slate-800 truncate">{{ editAttachmentFile.name }}</p>
              <p class="text-[11px] font-bold text-indigo-600">
                {{ isEditPdfAttachment ? 'Dokumen PDF' : 'File Foto/Gambar' }} • {{ (editAttachmentFile.size / 1024).toFixed(0) }} KB (File Baru)
              </p>
            </div>
            <button
              type="button"
              @click="clearEditAttachmentFile"
              class="p-2 text-rose-500 hover:bg-rose-50 rounded-xl transition-colors cursor-pointer"
              title="Batalkan file baru"
            >
              <X class="w-4 h-4" />
            </button>
          </div>

          <!-- Case 2: Ada dokumen pendukung yang sudah tersimpan sebelumnya (dan belum ada file baru) -->
          <div
            v-else-if="currentBanner.documentation_image_url"
            class="p-3 bg-white rounded-2xl border border-slate-200 flex items-center gap-3 shadow-xs"
          >
            <!-- Thumbnail Gambar -->
            <template v-if="!currentBanner.documentation_image_url.toLowerCase().endsWith('.pdf')">
              <div
                class="w-12 h-12 rounded-xl overflow-hidden bg-slate-100 border border-slate-200 shrink-0 cursor-pointer group relative"
                @click="openPhotoPreview(currentBanner.documentation_image_url, `Dokumen Pendukung - ${currentBanner.shop_name}`)"
              >
                <img :src="photoUrl(currentBanner.documentation_image_url)!" class="w-full h-full object-cover" />
                <div class="absolute inset-0 bg-black/20 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
                  <Eye class="w-3.5 h-3.5 text-white" />
                </div>
              </div>
              <div class="min-w-0 flex-1">
                <p class="text-xs font-black text-slate-800 truncate">Foto / Logo Pendukung</p>
                <button
                  type="button"
                  @click="openPhotoPreview(currentBanner.documentation_image_url, `Dokumen Pendukung - ${currentBanner.shop_name}`)"
                  class="text-[11px] font-bold text-indigo-600 hover:underline inline-flex items-center gap-1 cursor-pointer"
                >
                  <Eye class="w-3 h-3" /> Lihat Pratinjau
                </button>
              </div>
            </template>

            <!-- Thumbnail PDF -->
            <template v-else>
              <div class="w-12 h-12 rounded-xl bg-rose-50 border border-rose-200 flex items-center justify-center shrink-0 text-rose-600">
                <FileText class="w-6 h-6" />
              </div>
              <div class="min-w-0 flex-1">
                <p class="text-xs font-black text-slate-800 truncate">Dokumen PDF Terlampir</p>
                <a
                  :href="photoUrl(currentBanner.documentation_image_url)!"
                  target="_blank"
                  class="text-[11px] font-bold text-rose-600 hover:underline inline-flex items-center gap-1"
                >
                  <FileDown class="w-3 h-3" /> Buka PDF di Tab Baru
                </a>
              </div>
            </template>

            <!-- Action: Ganti atau Hapus -->
            <div class="flex items-center gap-1">
              <label
                class="p-2 text-indigo-600 hover:bg-indigo-50 rounded-xl transition-colors cursor-pointer"
                title="Ganti File"
              >
                <Edit2 class="w-4 h-4" />
                <input
                  type="file"
                  accept="image/png,image/jpeg,image/webp,application/pdf"
                  class="hidden"
                  @change="handleEditFileSelect"
                />
              </label>
              <button
                type="button"
                @click="removeExistingEditAttachment"
                class="p-2 text-rose-500 hover:bg-rose-50 rounded-xl transition-colors cursor-pointer"
                title="Hapus Dokumen"
              >
                <Trash2 class="w-4 h-4" />
              </button>
            </div>
          </div>

          <!-- Case 3: Belum ada file sama sekali -->
          <div v-else>
            <label
              class="border-2 border-dashed border-slate-200 hover:border-indigo-400 bg-white hover:bg-indigo-50/20 rounded-2xl p-4 flex flex-col items-center justify-center gap-1.5 transition-all cursor-pointer group text-center"
            >
              <UploadCloud class="w-7 h-7 text-slate-400 group-hover:text-indigo-600 transition-colors" />
              <p class="text-xs font-black text-slate-700">
                Klik untuk unggah Foto, Logo, atau PDF
              </p>
              <p class="text-[10px] font-bold text-slate-400">
                Format JPG, PNG, WEBP, atau PDF (Maks. 15MB)
              </p>
              <input
                type="file"
                accept="image/png,image/jpeg,image/webp,application/pdf"
                class="hidden"
                @change="handleEditFileSelect"
              />
            </label>
          </div>
        </div>

        <div>
          <label class="block text-xs font-bold text-slate-500 mb-2">Catatan</label>
          <textarea
            v-model="currentBanner.notes"
            rows="3"
            class="w-full bg-slate-50 border-none rounded-xl sm:rounded-2xl px-3 sm:px-4 py-3 sm:py-3.5 text-xs sm:text-sm font-bold text-slate-700 focus:ring-2 focus:ring-primary/20"
          ></textarea>
        </div>
      </div>

      <div class="p-4 sm:p-6 border-t border-slate-100 bg-slate-50/50 shrink-0 flex flex-col-reverse sm:flex-row justify-end gap-2 sm:gap-3">
        <button
          @click="isEditModalOpen = false"
          class="w-full sm:w-auto px-5 sm:px-6 py-3 sm:py-3.5 rounded-xl sm:rounded-2xl text-xs sm:text-sm font-bold text-slate-500 hover:bg-slate-200 transition-all cursor-pointer"
        >
          Batal
        </button>
        <button
          @click="handleSaveBanner"
          :disabled="isSaving"
          class="w-full sm:w-auto px-6 sm:px-8 py-3 sm:py-3.5 rounded-xl sm:rounded-2xl text-xs sm:text-sm font-bold text-white bg-primary hover:bg-primary/90 shadow-md sm:shadow-lg shadow-primary/20 transition-all disabled:opacity-50 cursor-pointer"
        >
          {{ isSaving ? 'Menyimpan...' : 'Simpan Perubahan' }}
        </button>
      </div>
    </div>
  </div>

  <!-- Modal Tambah Banner -->
  <div v-if="isAddModalOpen" class="fixed inset-0 z-[100] flex items-center justify-center p-4">
    <div
      class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
      @click="isAddModalOpen = false"
    ></div>
    <div
      class="bg-white rounded-[32px] w-full max-w-lg relative z-10 shadow-2xl flex flex-col max-h-[90vh] overflow-hidden animate-in fade-in zoom-in duration-300"
    >
      <div
        class="p-4 sm:p-6 border-b border-slate-100 flex items-center justify-between bg-slate-50/50 shrink-0"
      >
        <div>
          <h3 class="text-lg sm:text-xl font-black text-slate-900">Tambah Data Spanduk</h3>
          <p class="text-xs font-bold text-slate-400 mt-0.5 sm:mt-1">Buat permintaan spanduk baru</p>
        </div>
        <button
          @click="isAddModalOpen = false"
          class="p-2 sm:p-3 hover:bg-white rounded-xl sm:rounded-2xl transition-all text-slate-400 hover:text-slate-900 shadow-sm cursor-pointer"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <div class="p-4 sm:p-6 overflow-y-auto space-y-4 sm:space-y-5 flex-1">
        <div>
          <label class="block text-xs font-bold text-slate-500 mb-2">Salesman *</label>
          <SearchSelect
            :model-value="newBanner.sales_name"
            :options="salesSelectOptions"
            placeholder="Pilih atau cari Salesman..."
            search-placeholder="Ketik nama atau jabatan sales..."
            empty-message="Salesman tidak ditemukan"
            clearable
            @update:model-value="onSalesChangeAdd"
            @change="onSalesChangeAdd"
          />
        </div>
        <div>
          <div class="flex items-center justify-between mb-2">
            <label class="block text-xs font-bold text-slate-500">Nama Toko *</label>
            <span v-if="newBanner.sales_name && shopSelectOptionsForAdd.length > 0" class="text-[10px] font-bold text-emerald-600 bg-emerald-50 px-2 py-0.5 rounded-full border border-emerald-200">
              {{ shopSelectOptionsForAdd.length }} Toko Tersedia
            </span>
          </div>
          <SearchSelect
            v-model="newBanner.shop_name"
            :options="shopSelectOptionsForAdd"
            :disabled="!newBanner.sales_name"
            :placeholder="newBanner.sales_name ? 'Pilih atau cari Toko binaan sales...' : 'Pilih Salesman terlebih dahulu...'"
            :search-placeholder="newBanner.sales_name ? 'Ketik nama toko atau alamat...' : 'Pilih salesman dahulu'"
            allow-custom
            custom-label-prefix="Pakai Nama Toko: "
            :empty-message="newBanner.sales_name ? 'Toko tidak ditemukan. Anda dapat mengetik untuk memakai nama toko baru.' : 'Pilih Salesman terlebih dahulu untuk menampilkan daftar toko.'"
            @select="onSelectShopAdd"
            @custom-input="newBanner.shop_id = ''"
          />
          <!-- Loading indicator when fetching shops -->
          <div
            v-if="newBanner.sales_name && isShopsBySalesLoading"
            class="mt-2 p-2.5 bg-blue-50/80 border border-blue-200/80 rounded-xl flex items-center gap-2 text-xs font-medium text-blue-700 animate-in fade-in duration-200"
          >
            <Loader2 class="w-4 h-4 text-blue-600 animate-spin shrink-0" />
            <span>Memuat seluruh toko binaan <strong>{{ newBanner.sales_name }}</strong>...</span>
          </div>
          <!-- Hint when sales is not selected -->
          <div
            v-else-if="!newBanner.sales_name"
            class="mt-2 p-2.5 bg-amber-50/80 border border-amber-200/80 rounded-xl flex items-start sm:items-center gap-2 text-xs font-medium text-amber-800 animate-in fade-in duration-200"
          >
            <AlertCircle class="w-4 h-4 text-amber-600 shrink-0 mt-0.5 sm:mt-0" />
            <span>Pilih <strong>Salesman</strong> terlebih dahulu untuk menampilkan daftar toko yang menjadi tanggung jawabnya.</span>
          </div>
          <p v-else class="text-[11px] font-medium text-slate-400 mt-1.5 flex items-center gap-1">
            <span>💡 Menampilkan {{ shopSelectOptionsForAdd.length }} toko binaan <strong>{{ newBanner.sales_name }}</strong>. Bisa ketik nama baru jika toko belum terdaftar.</span>
          </p>
        </div>
        <div class="grid grid-cols-2 gap-3 sm:gap-4">
          <div>
            <label class="block text-xs font-bold text-slate-500 mb-2">Tipe Spanduk</label>
            <select
              v-model="newBanner.type"
              class="w-full bg-slate-50 border-none rounded-xl sm:rounded-2xl px-3 sm:px-4 py-3 sm:py-3.5 text-xs sm:text-sm font-bold text-slate-700 focus:ring-2 focus:ring-primary/20 cursor-pointer"
            >
              <option value="Custom">Custom (Desain Khusus Outlet)</option>
              <option value="Universal">Universal (Desain Standar Pabrik)</option>
            </select>
          </div>
          <div>
            <label class="block text-xs font-bold text-slate-500 mb-2">Bihun?</label>
            <select
              v-model="newBanner.is_bihun"
              class="w-full bg-slate-50 border-none rounded-xl sm:rounded-2xl px-3 sm:px-4 py-3 sm:py-3.5 text-xs sm:text-sm font-bold text-slate-700 focus:ring-2 focus:ring-primary/20"
            >
              <option :value="true">Ya, Bihun</option>
              <option :value="false">Tidak</option>
            </select>
          </div>
        </div>
        <!-- Ukuran Dimensi & Satuan -->
        <div>
          <div class="flex items-center justify-between mb-2">
            <label class="block text-xs font-bold text-slate-700">Ukuran Dimensi</label>
            <div class="inline-flex p-0.5 bg-slate-100 rounded-lg border border-slate-200">
              <button
                type="button"
                @click="setAddBannerUnit('cm')"
                :class="[
                  'px-2.5 py-0.5 text-[11px] font-bold rounded-md transition-all',
                  newBannerUnit === 'cm'
                    ? 'bg-white text-blue-600 shadow-xs'
                    : 'text-slate-500 hover:text-slate-800'
                ]"
              >
                cm
              </button>
              <button
                type="button"
                @click="setAddBannerUnit('m')"
                :class="[
                  'px-2.5 py-0.5 text-[11px] font-bold rounded-md transition-all',
                  newBannerUnit === 'm'
                    ? 'bg-white text-blue-600 shadow-xs'
                    : 'text-slate-500 hover:text-slate-800'
                ]"
              >
                meter (m)
              </button>
            </div>
          </div>
          <div class="grid grid-cols-3 gap-2 sm:gap-4">
            <div>
              <label class="block text-[11px] sm:text-xs font-bold text-slate-500 mb-1.5 sm:mb-2 truncate">
                Panjang ({{ newBannerUnit }})
              </label>
              <input
                v-model.number="newBanner.size_length"
                type="number"
                :step="newBannerUnit === 'm' ? '0.1' : '1'"
                class="w-full bg-slate-50 border-none rounded-xl sm:rounded-2xl px-2.5 sm:px-4 py-2.5 sm:py-3.5 text-xs sm:text-sm font-bold text-slate-700 focus:ring-2 focus:ring-primary/20"
              />
            </div>
            <div>
              <label class="block text-[11px] sm:text-xs font-bold text-slate-500 mb-1.5 sm:mb-2 truncate">
                Lebar ({{ newBannerUnit }})
              </label>
              <input
                v-model.number="newBanner.size_width"
                type="number"
                :step="newBannerUnit === 'm' ? '0.1' : '1'"
                class="w-full bg-slate-50 border-none rounded-xl sm:rounded-2xl px-2.5 sm:px-4 py-2.5 sm:py-3.5 text-xs sm:text-sm font-bold text-slate-700 focus:ring-2 focus:ring-primary/20"
              />
            </div>
            <div>
              <label class="block text-[11px] sm:text-xs font-bold text-slate-500 mb-1.5 sm:mb-2 truncate">Qty</label>
              <input
                v-model.number="newBanner.qty"
                type="number"
                min="1"
                class="w-full bg-slate-50 border-none rounded-xl sm:rounded-2xl px-2.5 sm:px-4 py-2.5 sm:py-3.5 text-xs sm:text-sm font-bold text-slate-700 focus:ring-2 focus:ring-primary/20"
              />
            </div>
          </div>
          <p v-if="newBanner.size_length && newBanner.size_width" class="text-[11px] font-semibold text-blue-600 mt-1.5 flex items-center gap-1">
            <Ruler class="w-3.5 h-3.5" />
            Setara: {{ formatBannerDimension(newBannerUnit === 'm' ? newBanner.size_length * 100 : newBanner.size_length, newBannerUnit === 'm' ? newBanner.size_width * 100 : newBanner.size_width, 'dual') }}
            • Luas: {{ ((newBannerUnit === 'm' ? newBanner.size_length : newBanner.size_length / 100) * (newBannerUnit === 'm' ? newBanner.size_width : newBanner.size_width / 100)).toFixed(2) }} m²
          </p>
        </div>
        <!-- ── Section: Foto / Logo / Dokumen Pendukung ──────────────────── -->
        <div class="p-3.5 sm:p-4 rounded-2xl bg-slate-50 border border-slate-200/80 space-y-3">
          <div class="flex items-center justify-between">
            <label class="block text-xs font-black text-slate-700 uppercase tracking-wider flex items-center gap-1.5">
              <Paperclip class="w-4 h-4 text-blue-600" />
              Foto / Logo / Dokumen Pendukung
            </label>
            <span class="text-[10px] font-bold text-slate-400 bg-white px-2 py-0.5 rounded-full border border-slate-200">
              Opsional
            </span>
          </div>

          <!-- Case 1: Ada file yang dipilih -->
          <div
            v-if="addAttachmentFile"
            class="p-3 bg-white rounded-2xl border border-blue-200 flex items-center gap-3 shadow-xs"
          >
            <!-- Preview: Gambar -->
            <div
              v-if="!isAddPdfAttachment && addAttachmentPreview"
              class="w-12 h-12 rounded-xl overflow-hidden bg-slate-100 border border-slate-200 shrink-0 cursor-pointer"
              @click="openPhotoPreview(addAttachmentPreview, 'File Baru')"
            >
              <img :src="addAttachmentPreview" class="w-full h-full object-cover" />
            </div>
            <!-- Preview: PDF -->
            <div
              v-else
              class="w-12 h-12 rounded-xl bg-rose-50 border border-rose-200 flex items-center justify-center shrink-0 text-rose-600"
            >
              <FileText class="w-6 h-6" />
            </div>
            <div class="min-w-0 flex-1">
              <p class="text-xs font-black text-slate-800 truncate">{{ addAttachmentFile.name }}</p>
              <p class="text-[11px] font-bold text-blue-600">
                {{ isAddPdfAttachment ? 'Dokumen PDF' : 'File Foto/Gambar' }} • {{ (addAttachmentFile.size / 1024).toFixed(0) }} KB
              </p>
            </div>
            <button
              type="button"
              @click="clearAddAttachmentFile"
              class="p-2 text-rose-500 hover:bg-rose-50 rounded-xl transition-colors cursor-pointer"
              title="Batalkan file"
            >
              <X class="w-4 h-4" />
            </button>
          </div>

          <!-- Case 2: Belum ada file -->
          <div v-else>
            <label
              class="border-2 border-dashed border-slate-200 hover:border-blue-400 bg-white hover:bg-blue-50/20 rounded-2xl p-4 flex flex-col items-center justify-center gap-1.5 transition-all cursor-pointer group text-center"
            >
              <UploadCloud class="w-7 h-7 text-slate-400 group-hover:text-blue-600 transition-colors" />
              <p class="text-xs font-black text-slate-700">
                Klik untuk unggah Foto, Logo, atau PDF
              </p>
              <p class="text-[10px] font-bold text-slate-400">
                Format JPG, PNG, WEBP, atau PDF (Maks. 15MB)
              </p>
              <input
                type="file"
                accept="image/png,image/jpeg,image/webp,application/pdf"
                class="hidden"
                @change="handleAddFileSelect"
              />
            </label>
          </div>
        </div>

        <div>
          <label class="block text-xs font-bold text-slate-500 mb-2">Catatan</label>
          <textarea
            v-model="newBanner.notes"
            rows="3"
            placeholder="Opsional"
            class="w-full bg-slate-50 border-none rounded-xl sm:rounded-2xl px-3 sm:px-4 py-3 sm:py-3.5 text-xs sm:text-sm font-bold text-slate-700 focus:ring-2 focus:ring-primary/20"
          ></textarea>
        </div>
      </div>

      <div class="p-4 sm:p-6 border-t border-slate-100 bg-slate-50/50 shrink-0 flex flex-col-reverse sm:flex-row justify-end gap-2 sm:gap-3">
        <button
          @click="isAddModalOpen = false"
          class="w-full sm:w-auto px-5 sm:px-6 py-3 sm:py-3.5 rounded-xl sm:rounded-2xl text-xs sm:text-sm font-bold text-slate-500 hover:bg-slate-200 transition-all cursor-pointer"
        >
          Batal
        </button>
        <button
          @click="handleCreateBanner"
          :disabled="isSaving"
          class="w-full sm:w-auto px-6 sm:px-8 py-3 sm:py-3.5 rounded-xl sm:rounded-2xl text-xs sm:text-sm font-bold text-white bg-blue-600 hover:bg-blue-700 shadow-md sm:shadow-lg shadow-blue-600/20 transition-all disabled:opacity-50 cursor-pointer"
        >
          {{ isSaving ? 'Menyimpan...' : 'Tambah Banner' }}
        </button>
      </div>
    </div>
  </div>

  <!-- Modal Konfirmasi Hapus Banner -->
  <div
    v-if="isDeleteModalOpen && bannerToDelete"
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs"
  >
    <div
      class="bg-white rounded-3xl max-w-lg w-full overflow-hidden shadow-2xl border border-slate-100 animate-in fade-in zoom-in-95 duration-200"
    >
      <div class="p-6">
        <div class="flex items-center gap-4 mb-4">
          <div class="w-12 h-12 rounded-2xl bg-rose-50 border border-rose-100 flex items-center justify-center text-rose-600 shrink-0">
            <AlertTriangle class="w-6 h-6" />
          </div>
          <div>
            <h3 class="text-lg font-black text-slate-800">
              Konfirmasi Hapus Spanduk
            </h3>
            <p class="text-xs text-slate-500 font-medium">
              Pastikan data banner yang dipilih benar sebelum dihapus.
            </p>
          </div>
        </div>

        <div class="bg-slate-50 rounded-2xl p-4 border border-slate-100 space-y-2.5 mb-4 text-xs">
          <div class="flex justify-between items-start gap-2">
            <span class="text-slate-400 font-semibold shrink-0">Nama Toko</span>
            <span class="font-bold text-slate-800 text-right">{{ bannerToDelete.shop_name || bannerToDelete.shop_id || '—' }}</span>
          </div>
          <div v-if="bannerToDelete.shop_id" class="flex justify-between items-center gap-2">
            <span class="text-slate-400 font-semibold shrink-0">ID Toko</span>
            <span class="font-bold text-slate-600 font-mono">{{ bannerToDelete.shop_id }}</span>
          </div>
          <div class="flex justify-between items-center gap-2">
            <span class="text-slate-400 font-semibold shrink-0">Salesman / Pemohon</span>
            <span class="font-bold text-slate-700">{{ bannerToDelete.sales_name || '—' }}</span>
          </div>
          <div class="flex justify-between items-center gap-2">
            <span class="text-slate-400 font-semibold shrink-0">Cabang</span>
            <span class="font-bold text-slate-700">{{ bannerToDelete.sales_branch || bannerToDelete.branch_name || 'Semua Cabang' }}</span>
          </div>
          <div class="flex justify-between items-center gap-2">
            <span class="text-slate-400 font-semibold shrink-0">Ukuran & Tipe</span>
            <span class="font-bold text-blue-600">
              {{ formatBannerDimension(bannerToDelete.size_length, bannerToDelete.size_width) }}
              <span v-if="bannerToDelete.banner_type" class="text-slate-500 font-normal">({{ bannerToDelete.banner_type }})</span>
            </span>
          </div>
          <div class="flex justify-between items-center gap-2">
            <span class="text-slate-400 font-semibold shrink-0">Status Saat Ini</span>
            <span
              :class="internalStatusColor(bannerToDelete.status)"
              class="px-2.5 py-0.5 rounded-lg text-[10px] font-black uppercase tracking-wider border"
            >
              {{ bannerToDelete.status }}
            </span>
          </div>
        </div>

        <div class="p-3 bg-rose-50/70 border border-rose-200/70 rounded-xl text-rose-700 text-xs flex items-start gap-2.5">
          <AlertCircle class="w-4 h-4 shrink-0 mt-0.5 text-rose-500" />
          <p class="leading-relaxed font-medium">
            <strong>Peringatan:</strong> Tindakan ini tidak dapat dibatalkan. Seluruh file desain, progres validasi, dan data terkait spanduk ini akan dihapus permanen.
          </p>
        </div>
      </div>

      <div class="p-4 sm:p-6 border-t border-slate-100 bg-slate-50/50 flex flex-col-reverse sm:flex-row justify-end gap-2 sm:gap-3">
        <button
          @click="closeDeleteModal"
          :disabled="isDeleting"
          class="w-full sm:w-auto px-5 sm:px-6 py-2.5 sm:py-3 rounded-xl sm:rounded-2xl text-xs sm:text-sm font-bold text-slate-600 hover:bg-slate-200 transition-all cursor-pointer disabled:opacity-50"
        >
          Batal
        </button>
        <button
          @click="confirmDeleteBanner"
          :disabled="isDeleting"
          class="w-full sm:w-auto inline-flex items-center justify-center gap-2 px-6 sm:px-8 py-2.5 sm:py-3 rounded-xl sm:rounded-2xl text-xs sm:text-sm font-bold text-white bg-rose-600 hover:bg-rose-700 shadow-md sm:shadow-lg shadow-rose-600/20 transition-all disabled:opacity-50 cursor-pointer"
        >
          <Loader2 v-if="isDeleting" class="w-4 h-4 animate-spin" />
          <Trash2 v-else class="w-4 h-4" />
          {{ isDeleting ? 'Menghapus...' : 'Ya, Hapus Banner Ini' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
</style>
