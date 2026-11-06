<script setup lang="ts">
import { ref, shallowRef, computed, onMounted, onUnmounted, markRaw, toRaw, watch } from 'vue'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import 'leaflet-defaulticon-compatibility'
import 'leaflet-defaulticon-compatibility/dist/leaflet-defaulticon-compatibility.css'

import markerIcon2x from 'leaflet/dist/images/marker-icon-2x.png'
import markerIcon from 'leaflet/dist/images/marker-icon.png'
import markerShadow from 'leaflet/dist/images/marker-shadow.png'

delete (L.Icon.Default.prototype as any)._getIconUrl
L.Icon.Default.mergeOptions({
  iconRetinaUrl: markerIcon2x,
  iconUrl: markerIcon,
  shadowUrl: markerShadow
})
import apiClient from '@/api/axios'
import { useMasterDataStore } from '@/stores/masterData'
import { useAuthStore } from '@/stores/auth'
import { useSalesVisitStore } from '@/stores/salesVisit'
import { toast } from 'vue-sonner'
import {
  Users,
  MapPin,
  Clock,
  Navigation,
  RefreshCcw,
  Maximize2,
  Search,
  Store as StoreIcon,
  Phone,
  Truck,
  Briefcase,
  Megaphone,
  Focus,
  ZoomIn,
  ZoomOut,
  RotateCcw,
  X,
  AlertTriangle,
  Radio,
  PauseCircle,
  WifiOff
} from 'lucide-vue-next'

export type FreshnessTier = 'LIVE' | 'STANDBY' | 'STALE' | 'OFFLINE' | 'NO_SIGNAL'

interface FieldStaffMonitoringItem {
  id: string
  name: string
  jobTitle: string
  departmentName: string
  category: 'SALES' | 'DRIVER' | 'MARKETING'
  branchName: string
  lat: number | null
  lng: number | null
  status: string
  dutyStatus?: string
  locationName: string
  storeName?: string
  storeAddress?: string
  storeCategory?: string
  storePhone?: string
  time: string | null
  timeAgo: string
  isLive: boolean
  isStandby: boolean
  isLostSignal: boolean
  isOffline: boolean
  hasCoordinates: boolean
  freshnessTier: FreshnessTier
  markerColor: string
  visitResult: string
  transactionsHtml: string
}

const map = shallowRef<any>(null)
const markerMap = shallowRef<Map<string, L.Marker>>(new Map())
const consolidatedList = ref<FieldStaffMonitoringItem[]>([])
const searchQuery = ref('')
const selectedCategoryFilter = ref<'ALL' | 'SALES' | 'DRIVER' | 'MARKETING'>('ALL')
const selectedStatusFilter = ref<'ALL' | 'LIVE' | 'STANDBY' | 'STALE' | 'OFFLINE'>('ALL')
const isLoading = ref(true)
const isInitialBoundsSet = ref(false)
const masterDataStore = useMasterDataStore()
const authStore = useAuthStore()
const visitStore = useSalesVisitStore()
const lastUpdated = ref(new Date().toLocaleTimeString())
let eventSource: EventSource | null = null
const lastKnownState = ref<Record<string, string>>({})

// Image Preview State
const showImageModal = ref(false)
const previewImageUrl = ref('')
const zoomLevel = ref(1)

const openImagePreview = (url: string) => {
  previewImageUrl.value = url
  zoomLevel.value = 1
  showImageModal.value = true
}

const handleZoom = (val: number) => {
  zoomLevel.value = Math.max(0.5, Math.min(3, zoomLevel.value + val))
}

// Global function for popup clicks
;(window as any).previewReceipt = (url: string) => {
  openImagePreview(url)
}

const initMap = () => {
  if (map.value) return

  // Default view to Indonesia coordinates
  const mapInstance = L.map('monitoring-map', {
    zoomControl: true,
  }).setView([-6.2088, 106.8456], 12)
  map.value = markRaw(mapInstance)

  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '&copy; OpenStreetMap contributors',
  }).addTo(toRaw(map.value) as any)
}

const getStaffTypeInfo = (title: string, dept: string = '') => {
  const t = (title || '').toLowerCase()
  const d = (dept || '').toLowerCase()

  if (
    t.includes('driver') ||
    t.includes('supir') ||
    t.includes('sopir') ||
    t.includes('pengemudi') ||
    t.includes('kurir') ||
    t.includes('delivery') ||
    d.includes('driver') ||
    d.includes('logistik') ||
    d.includes('delivery') ||
    d.includes('pengiriman')
  ) {
    return { category: 'DRIVER' as const, label: 'Driver', color: '#0284c7', icon: Truck }
  }
  if (t.includes('motoris')) {
    return { category: 'SALES' as const, label: 'Sales Motoris', color: '#3b82f6', icon: Briefcase }
  }
  if (t.includes('task order') || t.includes('take order') || /\bto\b/i.test(t)) {
    return { category: 'SALES' as const, label: 'Task Order', color: '#f59e0b', icon: Briefcase }
  }
  if (t.includes('booster')) {
    return { category: 'SALES' as const, label: 'Sales Booster', color: '#8b5cf6', icon: Briefcase }
  }
  if (t.includes('mt') || t.includes('modern trade') || t.includes('kam') || t.includes('key account') || /\bmt\b/i.test(t)) {
    return { category: 'SALES' as const, label: 'Modern Trade', color: '#ec4899', icon: StoreIcon }
  }
  if (
    t.includes('marketing') ||
    t.includes('promosi') ||
    t.includes('branding') ||
    t.includes('spanduk') ||
    d.includes('marketing') ||
    d.includes('promosi')
  ) {
    return { category: 'MARKETING' as const, label: 'Marketing', color: '#10b981', icon: Megaphone }
  }

  return { category: 'SALES' as const, label: 'Salesman', color: '#6366f1', icon: Briefcase }
}

const isFieldStaff = (title: string, dept: string = '') => {
  const pos = (title || '').toLowerCase().trim()
  const d = (dept || '').toLowerCase().trim()

  // Blacklist office, finance, factory, security, admin, HR roles
  const blacklist = [
    'accounting', 'finance', 'keuangan', 'kasir', 'cashier', 'pabrik', 'factory',
    'operator mesin', 'security', 'satpam', 'cleaning', 'ob', 'hrd', 'legal',
    'direktur', 'director', 'owner', 'auditor', 'validator', 'editor', 'customer service',
    'kepala gudang'
  ]
  for (const b of blacklist) {
    if (pos.includes(b) || d.includes(b)) {
      return false
    }
  }

  // Check Driver keywords
  const isDriver =
    pos.includes('driver') || pos.includes('supir') || pos.includes('sopir') ||
    pos.includes('pengemudi') || pos.includes('kurir') || pos.includes('delivery') ||
    d.includes('driver') || d.includes('logistik') || d.includes('delivery') || d.includes('pengiriman')

  // Check Sales / Marketing keywords
  const isSalesOrMkt =
    pos.includes('sales') || pos.includes('marketing') || pos.includes('motoris') ||
    pos.includes('booster') || pos.includes('take order') || pos.includes('task order') ||
    pos.includes('modern trade') || pos.includes('kam') || pos.includes('key account') ||
    pos.includes('kanvas') || pos.includes('canvas') || pos.includes('spanduk') ||
    pos.includes('promosi') || pos.includes('branding') || d.includes('sales') ||
    d.includes('marketing') || d.includes('promosi') || d.includes('branding') ||
    /\b(to|mt|spv)\b/i.test(pos)

  return isDriver || isSalesOrMkt
}

// Helpers for relative time & today validation
const formatRelativeTime = (timeStr: string | null): string => {
  if (!timeStr) return '-'
  const time = new Date(timeStr).getTime()
  if (isNaN(time)) return '-'
  const diffSec = Math.floor((Date.now() - time) / 1000)
  if (diffSec < 0) return 'Baru saja'
  if (diffSec < 60) return `${diffSec} dtk lalu`
  const diffMin = Math.floor(diffSec / 60)
  if (diffMin < 60) return `${diffMin} mnt lalu`
  const diffHour = Math.floor(diffMin / 60)
  if (diffHour < 24) return `${diffHour} jam lalu`
  return 'Kemarin'
}

const isToday = (timeStr: string | null): boolean => {
  if (!timeStr) return false
  const d = new Date(timeStr)
  if (isNaN(d.getTime())) return false
  const now = new Date()
  return (
    d.getFullYear() === now.getFullYear() &&
    d.getMonth() === now.getMonth() &&
    d.getDate() === now.getDate()
  )
}

// Recalculates freshness tier, status, colors, and timeAgo for an item
const applyItemFreshness = (item: FieldStaffMonitoringItem) => {
  const staffInfo = getStaffTypeInfo(item.jobTitle, item.departmentName)
  item.timeAgo = formatRelativeTime(item.time)

  if (item.isOffline) {
    item.freshnessTier = 'OFFLINE'
    item.status = 'OFFLINE (ABSEN KELUAR)'
    item.isLive = false
    item.isStandby = false
    item.isLostSignal = false
    item.markerColor = '#94a3b8'
    if (!item.locationName || item.locationName.includes('Menunggu')) {
      item.locationName = 'Lokasi Terakhir saat Absen Keluar'
    }
    return
  }

  // If item has no coordinates or coordinate is NOT recorded today
  if (!item.lat || !item.lng || isNaN(item.lat) || isNaN(item.lng) || !isToday(item.time)) {
    item.hasCoordinates = false
    item.freshnessTier = 'NO_SIGNAL'
    item.isLive = false
    item.isStandby = false
    item.isLostSignal = false
    item.markerColor = '#94a3b8'
    if (!item.dutyStatus || item.dutyStatus.includes('Belum Ada Absen')) {
      item.status = 'BELUM AKTIF / OFFLINE'
      item.locationName = 'Menunggu absensi & GPS HP'
    } else {
      item.status = 'ABSEN MASUK (MENUNGGU GPS HP)'
      item.locationName = 'Menunggu sinyal GPS handphone'
    }
    return
  }

  // Item has coordinates recorded TODAY
  const timeMs = item.time ? new Date(item.time).getTime() : 0
  const diffSec = Math.floor((Date.now() - timeMs) / 1000)

  if (diffSec < 600) {
    // Under 10 minutes = LIVE REAL-TIME
    item.freshnessTier = 'LIVE'
    item.isLive = true
    item.isStandby = false
    item.isLostSignal = false
    item.hasCoordinates = true

    if (item.storeName && item.visitResult === 'Sedang di lokasi toko...') {
      item.status = 'SEDANG BERKUNJUNG'
      item.locationName = item.storeName
      item.markerColor = '#10b981'
    } else {
      item.status = 'LIVE TRACKING (DI JALAN)'
      item.locationName = item.category === 'DRIVER' ? 'Perjalanan Pengiriman' : 'Perjalanan Lapangan'
      item.markerColor = item.category === 'DRIVER' ? '#0284c7' : staffInfo.color
    }
  } else if (diffSec < 3600) {
    // 10 to 60 minutes = STANDBY / IDLE
    item.freshnessTier = 'STANDBY'
    item.isLive = false
    item.isStandby = true
    item.isLostSignal = false
    item.hasCoordinates = true
    item.markerColor = '#f59e0b' // Amber/Yellow

    if (item.storeName && item.visitResult === 'Sedang di lokasi toko...') {
      item.status = 'BERKUNJUNG (STANDBY)'
      item.locationName = item.storeName
    } else {
      item.status = 'STANDBY / DIAM DI LOKASI'
      item.locationName = `Posisi ${item.timeAgo}`
    }
  } else {
    // Over 60 minutes today = STALE / LOST SIGNAL
    item.freshnessTier = 'STALE'
    item.isLive = false
    item.isStandby = false
    item.isLostSignal = true
    item.hasCoordinates = true
    item.markerColor = '#f97316' // Orange warning

    item.status = `SINYAL TERPUTUS (${item.timeAgo})`
    item.locationName = `Sinyal Terakhir ${item.time ? new Date(item.time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : ''}`
  }
}

const fetchMonitoringData = async (forceFitBounds = false) => {
  isLoading.value = true
  try {
    const branchId = masterDataStore.selectedBranchId
    const isGlobalBranch = !branchId || branchId === 'ALL_BRANCHES' || branchId === 'ALL'
    const branchQueryParam = !isGlobalBranch ? `&branch_id=${branchId}` : ''
    const today = new Date().toISOString().split('T')[0]

    // Fetch multi-sumber data secara paralel
    const liveLocBranchQuery = !isGlobalBranch ? `?branch_id=${branchId}` : ''
    const [empRes, liveLocRes, attRes] = await Promise.allSettled([
      apiClient.get(`/employees?limit=all${branchQueryParam}`),
      apiClient.get(`/admin/sales/live-locations${liveLocBranchQuery}`),
      apiClient.get(`/attendance/all?limit=500${branchQueryParam}`),
      visitStore.fetchVisits({
        start_date: today,
        end_date: today,
        branch_id: isGlobalBranch ? '' : branchId,
        sales_type: masterDataStore.selectedSalesType === 'ALL_SALES' ? '' : masterDataStore.selectedSalesType,
        limit: 1000
      })
    ])

    const allEmployees = empRes.status === 'fulfilled' ? (empRes.value.data?.data || empRes.value.data || []) : []
    const liveLocData = liveLocRes.status === 'fulfilled' ? (liveLocRes.value.data?.data || liveLocRes.value.data || []) : []
    const attData = attRes.status === 'fulfilled' ? (attRes.value.data?.data || attRes.value.data || []) : []
    const visits = visitStore.visits || []

    const consolidatedMap = new Map<string, FieldStaffMonitoringItem>()

    // 1. Inisialisasi HANYA Karyawan Lapangan (Sales, Driver, Marketing) yang aktif
    allEmployees.forEach((emp: any) => {
      const isActive = emp.user?.is_active !== false && emp.employment_status !== 'RESIGNED' && emp.employment_status !== 'TERMINATED'
      if (!isActive) return

      const jobTitle = emp.job_position?.title || emp.job_title || ''
      const deptName = emp.department?.name || ''
      if (!isFieldStaff(jobTitle, deptName)) return

      const staffInfo = getStaffTypeInfo(jobTitle, deptName)

      // Sales Type Filter jika dipilih
      const type = masterDataStore.selectedSalesType
      if (type && type !== 'ALL_SALES' && staffInfo.category === 'SALES') {
        let matchSalesType = true
        const lowTitle = jobTitle.toLowerCase()
        if (type === 'sales_mt') {
          matchSalesType = /\b(mt|modern trade|kam|key account)\b/i.test(lowTitle)
        } else if (type === 'sales_to') {
          matchSalesType = /\b(to|task order|take order)\b/i.test(lowTitle)
        } else if (type === 'sales_motoris_booster') {
          matchSalesType = /\b(motoris|booster)\b/i.test(lowTitle)
        }
        if (!matchSalesType) return
      }

      consolidatedMap.set(emp.id, {
        id: emp.id,
        name: `${emp.first_name || ''} ${emp.last_name || ''}`.trim(),
        jobTitle: jobTitle || staffInfo.label,
        departmentName: deptName,
        category: staffInfo.category,
        branchName: emp.branch?.name || '',
        lat: null,
        lng: null,
        status: 'BELUM AKTIF / OFFLINE',
        dutyStatus: 'Belum Ada Absen Hari Ini',
        locationName: 'Menunggu sinyal GPS handphone',
        time: null,
        timeAgo: '-',
        isLive: false,
        isStandby: false,
        isLostSignal: false,
        isOffline: true,
        hasCoordinates: false,
        freshnessTier: 'NO_SIGNAL',
        markerColor: '#94a3b8',
        visitResult: '',
        transactionsHtml: ''
      })
    })

    // 2. Terapkan Data Absensi Hari Ini
    const latestAttByEmp = new Map<string, any>()
    attData.forEach((att: any) => {
      const empId = att.employee?.id || att.employee_id
      if (!empId) return

      const existing = latestAttByEmp.get(empId)
      if (!existing) {
        latestAttByEmp.set(empId, att)
      } else {
        const existingTime = new Date(existing.clock_in_time || 0).getTime()
        const newTime = new Date(att.clock_in_time || 0).getTime()
        if (newTime > existingTime) {
          latestAttByEmp.set(empId, att)
        }
      }
    })

    latestAttByEmp.forEach((att: any, empId: string) => {
      let item = consolidatedMap.get(empId)
      if (!item && att.employee) {
        const emp = att.employee
        const jobTitle = emp.job_position?.title || emp.job_title || ''
        const deptName = emp.department?.name || ''
        if (isFieldStaff(jobTitle, deptName)) {
          const staffInfo = getStaffTypeInfo(jobTitle, deptName)
          item = {
            id: empId,
            name: `${emp.first_name || ''} ${emp.last_name || ''}`.trim(),
            jobTitle: jobTitle || staffInfo.label,
            departmentName: deptName,
            category: staffInfo.category,
            branchName: emp.branch?.name || '',
            lat: null,
            lng: null,
            status: 'BELUM AKTIF / OFFLINE',
            dutyStatus: 'Belum Ada Absen Hari Ini',
            locationName: 'Menunggu sinyal GPS handphone',
            time: null,
            timeAgo: '-',
            isLive: false,
            isStandby: false,
            isLostSignal: false,
            isOffline: true,
            hasCoordinates: false,
            freshnessTier: 'NO_SIGNAL',
            markerColor: '#94a3b8',
            visitResult: '',
            transactionsHtml: ''
          }
          consolidatedMap.set(empId, item)
        }
      }

      if (item) {
        if (att.clock_out_time) {
          item.status = 'OFFLINE (ABSEN KELUAR)'
          item.dutyStatus = 'Sudah Absen Keluar'
          item.isOffline = true
        } else if (att.clock_in_time) {
          item.status = 'ABSEN MASUK (MENUNGGU GPS HP)'
          item.dutyStatus = 'Aktif (Sudah Absen Masuk)'
          item.isOffline = false
        }
      }
    })

    // 3. Terapkan Riwayat Kunjungan Toko Hari Ini
    visits.forEach((v: any) => {
      const empId = v.employee_id
      if (!empId) return

      let item = consolidatedMap.get(empId)
      const storeName = v.store?.name || ''
      const storeAddress = v.store?.address || [v.store?.district, v.store?.regency, v.store?.province].filter(Boolean).join(', ') || ''
      const storeCategory = v.store?.customer_type || v.store?.category || ''
      const storePhone = v.store?.phone_number || ''

      if (item) {
        item.storeName = storeName
        item.storeAddress = storeAddress
        item.storeCategory = storeCategory
        item.storePhone = storePhone

        if (!v.check_out_time) {
          item.visitResult = 'Sedang di lokasi toko...'
        } else {
          const trxs = (v.transactions?.length || 0) + (v.sales_orders?.length || 0)
          if (trxs > 0) item.visitResult = `Berhasil (${trxs} Transaksi)`
          else item.visitResult = 'Selesai (Tidak ada transaksi)'
        }

        if (v.transactions && v.transactions.length > 0) {
          item.transactionsHtml = `
            <div class="mt-3 pt-3 border-t border-slate-100">
              <p class="text-[9px] font-black text-slate-400 uppercase tracking-widest mb-2 flex items-center gap-1">
                Transaksi Terdeteksi
              </p>
              ${v.transactions.map((t: any) => `
                <div class="bg-slate-50 rounded-xl p-2.5 mb-2 border border-slate-100">
                  <div class="flex justify-between items-center mb-1">
                    <span class="text-[11px] font-black text-slate-900">${t.store?.name || item?.locationName}</span>
                    <span class="text-[10px] font-black text-emerald-600">Rp ${new Intl.NumberFormat('id-ID').format(t.total_amount)}</span>
                  </div>
                </div>
              `).join('')}
            </div>
          `
        } else if (v.notes) {
          item.transactionsHtml = `
            <div class="mt-3 pt-3 border-t border-slate-100">
              <p class="text-[9px] font-black text-slate-400 uppercase tracking-widest mb-1 flex items-center gap-1">
                Catatan Kunjungan
              </p>
              <p class="text-[10px] font-medium text-slate-600 leading-relaxed italic">"${v.notes}"</p>
            </div>
          `
        }
      }
    })

    // 4. Terapkan Data Live GPS Terkini dari Handphone
    liveLocData.forEach((loc: any) => {
      const empId = loc.employee_id
      if (!empId) return

      let item = consolidatedMap.get(empId)
      if (!item) {
        const jobTitle = loc.job_title || ''
        const deptName = loc.department || ''
        if (isFieldStaff(jobTitle, deptName)) {
          const staffInfo = getStaffTypeInfo(jobTitle, deptName)
          item = {
            id: empId,
            name: loc.employee_name || 'Petugas Lapangan',
            jobTitle: jobTitle || staffInfo.label,
            departmentName: deptName,
            category: loc.category || staffInfo.category,
            branchName: loc.branch_name || '',
            lat: null,
            lng: null,
            status: 'STANDBY',
            dutyStatus: loc.is_on_duty ? 'Aktif' : 'Standby',
            locationName: 'Menunggu sinyal GPS handphone',
            time: null,
            timeAgo: '-',
            isLive: false,
            isStandby: false,
            isLostSignal: false,
            freshnessTier: 'NO_SIGNAL',
            markerColor: staffInfo.color,
            isOffline: false,
            hasCoordinates: false,
            visitResult: '',
            transactionsHtml: ''
          }
          consolidatedMap.set(empId, item)
        }
      }

      if (item && loc.latitude && loc.longitude) {
        const lat = Number(loc.latitude)
        const lng = Number(loc.longitude)
        const locTime = loc.recorded_at || loc.received_at || loc.updated_at

        // Strictly validate coordinates and verify that the timestamp is from TODAY
        if (!isNaN(lat) && !isNaN(lng) && isToday(locTime)) {
          item.lat = lat
          item.lng = lng
          item.time = locTime
          item.hasCoordinates = true
          if (loc.is_on_duty !== undefined && !loc.is_on_duty) {
            item.isOffline = true
          }
        }
      }
    })

    // 5. Terapkan Kalkulasi Freshness Tingkat Lanjut ke Seluruh Item
    consolidatedMap.forEach((item) => {
      applyItemFreshness(item)
    })

    // Sort: Live di atas -> Standby -> Sinyal Terputus -> Belum Aktif -> Offline Selesai
    const tierPriority: Record<FreshnessTier, number> = {
      LIVE: 1,
      STANDBY: 2,
      STALE: 3,
      NO_SIGNAL: 4,
      OFFLINE: 5
    }

    consolidatedList.value = Array.from(consolidatedMap.values()).sort((a, b) => {
      const pA = tierPriority[a.freshnessTier] || 99
      const pB = tierPriority[b.freshnessTier] || 99
      if (pA !== pB) return pA - pB
      return a.name.localeCompare(b.name)
    })

    updateMarkers(forceFitBounds)
  } catch (error) {
    console.error('Failed to fetch monitoring data:', error)
  } finally {
    isLoading.value = false
    lastUpdated.value = new Date().toLocaleTimeString()
  }
}

// Background auto-refresh of freshness timer every 10 seconds
const tickFreshnessTimer = () => {
  consolidatedList.value.forEach((item) => {
    applyItemFreshness(item)
  })
}

const filteredConsolidatedList = computed(() => {
  let list = consolidatedList.value

  // Category filter
  if (selectedCategoryFilter.value !== 'ALL') {
    list = list.filter((item) => item.category === selectedCategoryFilter.value)
  }

  // Status tier filter
  if (selectedStatusFilter.value !== 'ALL') {
    if (selectedStatusFilter.value === 'LIVE') {
      list = list.filter((item) => item.freshnessTier === 'LIVE')
    } else if (selectedStatusFilter.value === 'STANDBY') {
      list = list.filter((item) => item.freshnessTier === 'STANDBY')
    } else if (selectedStatusFilter.value === 'STALE') {
      list = list.filter((item) => item.freshnessTier === 'STALE')
    } else if (selectedStatusFilter.value === 'OFFLINE') {
      list = list.filter((item) => item.freshnessTier === 'OFFLINE' || item.freshnessTier === 'NO_SIGNAL')
    }
  }

  // Search query filter
  const q = searchQuery.value.toLowerCase().trim()
  if (!q) return list

  return list.filter((item) => {
    return (
      item.name.toLowerCase().includes(q) ||
      item.jobTitle.toLowerCase().includes(q) ||
      item.locationName.toLowerCase().includes(q) ||
      (item.storeName && item.storeName.toLowerCase().includes(q)) ||
      (item.storeAddress && item.storeAddress.toLowerCase().includes(q)) ||
      item.status.toLowerCase().includes(q)
    )
  })
})

const countsByStatus = computed(() => {
  const all = consolidatedList.value
  return {
    ALL: all.length,
    LIVE: all.filter((i) => i.freshnessTier === 'LIVE').length,
    STANDBY: all.filter((i) => i.freshnessTier === 'STANDBY').length,
    STALE: all.filter((i) => i.freshnessTier === 'STALE').length,
    OFFLINE: all.filter((i) => i.freshnessTier === 'OFFLINE' || i.freshnessTier === 'NO_SIGNAL').length,
  }
})

const countsByCategory = computed(() => {
  const all = consolidatedList.value
  return {
    ALL: all.length,
    SALES: all.filter((i) => i.category === 'SALES').length,
    DRIVER: all.filter((i) => i.category === 'DRIVER').length,
    MARKETING: all.filter((i) => i.category === 'MARKETING').length,
  }
})

// Update Marker Positions without Rebuilding Map & without Displacing User View
const updateMarkers = (forceFitBounds = false) => {
  if (!map.value) return

  const currentEmpIds = new Set<string>()
  const validBounds: [number, number][] = []

  consolidatedList.value.forEach((item) => {
    const prevState = lastKnownState.value[item.id]
    const currentStateStr = `${item.status}_${item.locationName}_${item.time}`

    // Notify status change on background updates
    if (prevState && prevState !== currentStateStr && !isLoading.value && isInitialBoundsSet.value) {
      let verb = 'berada di'
      if (item.status === 'SEDANG BERKUNJUNG') verb = 'memulai kunjungan di'
      if (item.status === 'SELESAI KUNJUNGAN') verb = 'selesai kunjungan di'
      if (item.status === 'OFFLINE (ABSEN KELUAR)') verb = 'telah Absen Keluar di'
      if (item.freshnessTier === 'LIVE') {
        toast.info(`${item.name} (${item.jobTitle}) ${verb} ${item.locationName}`)
      }
    }
    lastKnownState.value[item.id] = currentStateStr

    // If item has NO coordinates or is NOT active today, remove existing marker
    if (!item.hasCoordinates || !item.lat || !item.lng || isNaN(item.lat) || isNaN(item.lng) || item.freshnessTier === 'NO_SIGNAL') {
      if (markerMap.value.has(item.id)) {
        markerMap.value.get(item.id)?.remove()
        markerMap.value.delete(item.id)
      }
      return
    }

    currentEmpIds.add(item.id)
    validBounds.push([item.lat, item.lng])

    const staffInfo = getStaffTypeInfo(item.jobTitle, item.departmentName)
    const markerColor = item.markerColor || staffInfo.color
    const timeText = item.time
      ? new Date(item.time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
      : '-'

    // Distinct Visual Styling per Tier
    let pingHtml = ''
    let markerRingClass = 'border-2 border-white'
    let opacityClass = 'opacity-100'

    if (item.freshnessTier === 'LIVE') {
      pingHtml = `<div class="absolute w-full h-full rounded-full opacity-35 animate-ping" style="background-color: ${markerColor}"></div>`
      markerRingClass = 'border-2 border-white ring-2 ring-emerald-500/50 ring-offset-1'
    } else if (item.freshnessTier === 'STANDBY') {
      markerRingClass = 'border-2 border-amber-300'
    } else if (item.freshnessTier === 'STALE') {
      markerRingClass = 'border-2 border-dashed border-orange-400'
      opacityClass = 'opacity-85'
    } else if (item.freshnessTier === 'OFFLINE') {
      markerRingClass = 'border-2 border-slate-300'
      opacityClass = 'opacity-65'
    }

    const customIcon = L.divIcon({
      className: 'custom-sales-marker',
      html: `
        <div class="relative flex items-center justify-center w-10 h-10 ${opacityClass}">
          ${pingHtml}
          <div class="relative w-8 h-8 rounded-full ${markerRingClass} shadow-lg flex items-center justify-center text-white font-black text-[10px]" style="background-color: ${markerColor}">
            ${item.name.charAt(0)}
          </div>
          ${item.freshnessTier === 'LIVE' ? '<span class="absolute -top-1 -right-1 w-3 h-3 bg-emerald-500 rounded-full border-2 border-white animate-pulse"></span>' : ''}
          ${item.freshnessTier === 'STANDBY' ? '<span class="absolute -top-1 -right-1 w-3 h-3 bg-amber-500 rounded-full border-2 border-white"></span>' : ''}
          ${item.freshnessTier === 'STALE' ? '<span class="absolute -top-1 -right-1 w-3 h-3 bg-orange-500 rounded-full border-2 border-white"></span>' : ''}
        </div>
      `,
      iconSize: [40, 40],
      iconAnchor: [20, 20],
      popupAnchor: [0, -20],
    })

    // Status Badge Styling in Popup
    let popupStatusBadge = `bg-emerald-100 text-emerald-700`
    let freshnessLabel = `🟢 LIVE (${item.timeAgo})`
    if (item.freshnessTier === 'STANDBY') {
      popupStatusBadge = `bg-amber-100 text-amber-700`
      freshnessLabel = `🟡 STANDBY (${item.timeAgo})`
    } else if (item.freshnessTier === 'STALE') {
      popupStatusBadge = `bg-orange-100 text-orange-700`
      freshnessLabel = `🟠 SINYAL TERPUTUS (${item.timeAgo})`
    } else if (item.freshnessTier === 'OFFLINE') {
      popupStatusBadge = `bg-slate-200 text-slate-700`
      freshnessLabel = `⚫ SELESAI SHIFT`
    }

    const popupHtml = `
      <div class="w-72 font-sans p-1">
        <div class="flex items-center gap-3 mb-3">
          <div class="w-10 h-10 rounded-xl flex items-center justify-center font-black text-white shadow-sm shrink-0" style="background-color: ${markerColor}">
            ${item.name.charAt(0)}
          </div>
          <div class="min-w-0 flex-1">
            <p class="font-black text-slate-900 text-sm truncate leading-tight">${item.name}</p>
            <div class="flex items-center gap-1.5 mt-0.5">
              <span class="px-1.5 py-0.5 rounded-md text-[8px] font-black text-white uppercase tracking-tighter" style="background-color: ${staffInfo.color}">${staffInfo.label}</span>
              <span v-if="item.branchName" class="text-[8px] font-bold text-slate-400 truncate">${item.branchName}</span>
            </div>
          </div>
        </div>
        
        <div class="space-y-2 mb-3">
          <div class="bg-slate-50 rounded-xl p-3 border border-slate-100">
            <div class="flex justify-between items-center mb-1.5">
              <span class="text-[9px] font-black text-slate-400 uppercase tracking-widest">Status Terkini</span>
              <span class="text-[9px] font-black px-2 py-0.5 rounded-md ${popupStatusBadge}">${freshnessLabel}</span>
            </div>

            ${item.storeName ? `
              <div class="mt-2 pt-2 border-t border-slate-200/60">
                <div class="flex items-center gap-1.5">
                  <span class="text-[11px] font-black text-slate-900 leading-snug">${item.storeName}</span>
                  ${item.storeCategory ? `<span class="text-[8px] font-black px-1 py-0.2 bg-blue-100 text-blue-700 rounded uppercase">${item.storeCategory}</span>` : ''}
                </div>
                ${item.storeAddress ? `
                  <p class="text-[10px] text-slate-500 font-medium mt-0.5 leading-relaxed flex items-start gap-1">
                    <span class="shrink-0 text-slate-400">📍</span>
                    <span>${item.storeAddress}</span>
                  </p>
                ` : ''}
                ${item.storePhone ? `
                  <p class="text-[9px] text-slate-400 font-medium mt-0.5">📞 ${item.storePhone}</p>
                ` : ''}
              </div>
            ` : `
              <p class="text-xs font-black text-slate-900 leading-tight mt-1">${item.locationName}</p>
            `}

            <div class="flex justify-between items-center mt-2.5 pt-2 border-t border-slate-100">
              <p class="text-[10px] font-bold text-primary">🕒 ${timeText} <span class="text-slate-400 text-[9px] font-normal">(${item.timeAgo})</span></p>
              <p class="text-[10px] font-bold text-slate-600">${item.visitResult || ''}</p>
            </div>
          </div>
        </div>

        ${item.transactionsHtml || ''}
      </div>
    `

    if (markerMap.value.has(item.id)) {
      const existing = markerMap.value.get(item.id)!
      existing.setLatLng([item.lat, item.lng])
      existing.setIcon(customIcon)
      const popup = existing.getPopup()
      if (popup) {
        popup.setContent(popupHtml)
      } else {
        existing.bindPopup(popupHtml, { maxWidth: 320, className: 'custom-monitoring-popup' })
      }
    } else {
      const newMarker = L.marker([item.lat, item.lng], { icon: customIcon })
        .addTo(toRaw(map.value) as any)
        .bindPopup(popupHtml, {
          maxWidth: 320,
          className: 'custom-monitoring-popup',
        })
      markerMap.value.set(item.id, markRaw(newMarker))
    }
  })

  // Clean up markers of staff no longer present or without coordinates
  for (const [id, marker] of markerMap.value.entries()) {
    if (!currentEmpIds.has(id)) {
      marker.remove()
      markerMap.value.delete(id)
    }
  }

  // Only fit bounds on the FIRST load or when explicitly forced by user
  if ((!isInitialBoundsSet.value || forceFitBounds) && validBounds.length > 0) {
    try {
      const bounds = L.latLngBounds(validBounds)
      map.value.fitBounds(bounds, { padding: [40, 40], maxZoom: 15 })
      isInitialBoundsSet.value = true
    } catch (e) {
      console.error('Error fitting bounds:', e)
    }
  }
}

const fitAllMarkers = () => {
  if (!map.value) return
  const validBounds: [number, number][] = []
  consolidatedList.value.forEach((i) => {
    if (i.hasCoordinates && i.lat && i.lng && !isNaN(i.lat) && !isNaN(i.lng) && i.freshnessTier !== 'NO_SIGNAL') {
      validBounds.push([i.lat, i.lng])
    }
  })
  if (validBounds.length > 0) {
    try {
      const bounds = L.latLngBounds(validBounds)
      map.value.fitBounds(bounds, { padding: [40, 40], maxZoom: 15 })
      toast.success('Peta telah dipusatkan ke seluruh petugas aktif')
    } catch (e) {
      console.error(e)
    }
  } else {
    toast.warning('Belum ada koordinat aktif hari ini untuk dipusatkan.')
  }
}

const updateRealtimeStaffLocation = (payload: any) => {
  if (!payload || !payload.employee_id || !payload.latitude || !payload.longitude) return

  const empId = payload.employee_id
  const lat = Number(payload.latitude)
  const lng = Number(payload.longitude)
  if (isNaN(lat) || isNaN(lng)) return

  const locTime = payload.recorded_at || payload.received_at || new Date().toISOString()
  if (!isToday(locTime)) return

  let item = consolidatedList.value.find((i) => i.id === empId)

  // If item doesn't exist yet, try to create from payload if eligible
  if (!item) {
    const jobTitle = payload.job_title || ''
    const deptName = payload.department || ''
    if (isFieldStaff(jobTitle, deptName)) {
      const staffInfo = getStaffTypeInfo(jobTitle, deptName)
      item = {
        id: empId,
        name: payload.employee_name || 'Petugas Lapangan',
        jobTitle: jobTitle || staffInfo.label,
        departmentName: deptName,
        category: payload.category || staffInfo.category,
        branchName: '',
        lat,
        lng,
        status: 'LIVE TRACKING (DI JALAN)',
        locationName: (payload.category === 'DRIVER' || staffInfo.category === 'DRIVER') ? 'Perjalanan Pengiriman' : 'Perjalanan Lapangan',
        time: locTime,
        timeAgo: 'Baru saja',
        isLive: true,
        isStandby: false,
        isLostSignal: false,
        freshnessTier: 'LIVE',
        markerColor: (payload.category === 'DRIVER' || staffInfo.category === 'DRIVER') ? '#0284c7' : '#06b6d4',
        isOffline: false,
        hasCoordinates: true,
        visitResult: '',
        transactionsHtml: ''
      }
      consolidatedList.value.push(item)
    }
  }

  if (item) {
    item.lat = lat
    item.lng = lng
    item.time = locTime
    item.hasCoordinates = true
    if (payload.is_on_duty !== undefined) {
      item.isOffline = !payload.is_on_duty
    }
    applyItemFreshness(item)
    updateMarkers(false)
    lastUpdated.value = new Date().toLocaleTimeString()
  }
}

let refreshInterval: any = null
let fastTimerInterval: any = null

onMounted(() => {
  initMap()
  fetchMonitoringData(true)

  // Auto refresh data every 30 seconds as background fallback
  refreshInterval = setInterval(() => fetchMonitoringData(false), 30000)

  // Tick freshness calculation every 10 seconds
  fastTimerInterval = setInterval(tickFreshnessTimer, 10000)

  // Setup EventSource for real-time live location updates via Redis/SSE
  if (authStore.token) {
    const baseUrl = import.meta.env.VITE_API_BASE_URL || '/api'
    const url = `${baseUrl}/v1/sse/notifications?token=${authStore.token}`
    eventSource = new EventSource(url)
    eventSource.addEventListener('salesman_location', (event: MessageEvent) => {
      try {
        if (event.data) {
          const payload = JSON.parse(event.data)
          updateRealtimeStaffLocation(payload)
        }
      } catch (e) {
        console.error('Failed to parse realtime location update:', e)
        fetchMonitoringData(false)
      }
    })
  }
})

watch(
  () => masterDataStore.selectedBranchId,
  () => {
    isInitialBoundsSet.value = false
    fetchMonitoringData(true)
  }
)

watch(
  () => masterDataStore.selectedSalesType,
  () => {
    fetchMonitoringData(false)
  }
)

onUnmounted(() => {
  if (refreshInterval) clearInterval(refreshInterval)
  if (fastTimerInterval) clearInterval(fastTimerInterval)
  if (eventSource) {
    eventSource.close()
    eventSource = null
  }
  if (map.value) {
    map.value.remove()
    map.value = null
  }
})

const panToMarker = (item: FieldStaffMonitoringItem) => {
  if (!map.value) return
  if (item.hasCoordinates && item.lat && item.lng && !isNaN(item.lat) && !isNaN(item.lng) && item.freshnessTier !== 'NO_SIGNAL') {
    map.value.flyTo([item.lat, item.lng], 16)
    const targetMarker = markerMap.value.get(item.id)
    if (targetMarker) {
      targetMarker.openPopup()
    }
  } else {
    toast.warning(`Belum ada koordinat GPS aktif hari ini untuk ${item.name}.`)
  }
}
</script>

<template>
  <div class="h-[calc(100vh-120px)] flex flex-col gap-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-6 shrink-0">
      <div>
        <h1 class="text-3xl font-extrabold text-slate-900 tracking-tight flex items-center flex-wrap gap-3">
          <Navigation class="w-8 h-8 text-primary" />
          <span>Live Monitoring Petugas Lapangan</span>
          <span v-if="masterDataStore.selectedBranchName" class="text-primary text-2xl font-bold">
            — {{ masterDataStore.selectedBranchName }}
          </span>
        </h1>
        <p class="text-[15px] font-medium text-slate-500 mt-1.5 flex items-center gap-2">
          <RefreshCcw class="w-4 h-4 text-emerald-500" :class="{ 'animate-spin': isLoading }" />
          Monitoring real-time tim Driver, Sales, dan Marketing. Terakhir diperbarui: {{ lastUpdated }}
        </p>
      </div>
      <div class="flex items-center gap-3">
        <button
          @click="fitAllMarkers"
          class="bg-white border border-slate-200 px-4 py-2.5 rounded-2xl text-xs font-bold text-slate-700 hover:bg-slate-50 transition-all shadow-sm flex items-center gap-2"
          title="Pusatkan peta ke semua petugas aktif"
        >
          <Focus class="w-4 h-4 text-primary" /> Pusatkan Peta
        </button>
        <button
          @click="fetchMonitoringData(false)"
          class="bg-primary text-white px-5 py-2.5 rounded-2xl text-xs font-bold hover:bg-primary/90 transition-all shadow-sm flex items-center gap-2"
        >
          <RefreshCcw class="w-4 h-4" /> Refresh Data
        </button>
      </div>
    </div>

    <!-- Main Content -->
    <div class="flex-1 flex flex-col lg:flex-row gap-6 min-h-0">
      <!-- Left: Sidebar List -->
      <div
        class="lg:w-96 bg-white rounded-3xl border border-slate-200 shadow-sm flex flex-col overflow-hidden shrink-0"
      >
        <!-- Sidebar Header & Filter -->
        <div class="p-4 border-b border-slate-100 bg-slate-50/50 space-y-3">
          <div class="flex items-center justify-between">
            <h3 class="font-extrabold text-slate-900 flex items-center gap-2 text-sm">
              <Users class="w-4 h-4 text-slate-500" />
              <span>Petugas Lapangan</span>
              <span class="text-xs bg-slate-200 text-slate-700 px-2 py-0.5 rounded-full font-black">
                {{ filteredConsolidatedList.length }}
              </span>
            </h3>
            <div class="flex items-center gap-1.5">
              <span class="text-[10px] font-black bg-emerald-100 text-emerald-700 px-2 py-0.5 rounded-full flex items-center gap-1">
                <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
                {{ countsByStatus.LIVE }} Live
              </span>
              <span v-if="countsByStatus.STANDBY > 0" class="text-[10px] font-black bg-amber-100 text-amber-700 px-2 py-0.5 rounded-full">
                {{ countsByStatus.STANDBY }} Standby
              </span>
            </div>
          </div>

          <!-- Status Quick Filter Pills -->
          <div class="flex items-center gap-1.5 overflow-x-auto no-scrollbar pb-0.5">
            <button
              @click="selectedStatusFilter = 'ALL'"
              :class="selectedStatusFilter === 'ALL' ? 'bg-slate-900 text-white font-black' : 'bg-slate-100 text-slate-600 hover:bg-slate-200 font-bold'"
              class="px-2.5 py-1 rounded-xl text-[10px] transition-all shrink-0"
            >
              Semua Status ({{ countsByStatus.ALL }})
            </button>
            <button
              @click="selectedStatusFilter = 'LIVE'"
              :class="selectedStatusFilter === 'LIVE' ? 'bg-emerald-600 text-white font-black shadow-sm' : 'bg-emerald-50 text-emerald-700 hover:bg-emerald-100 font-bold'"
              class="px-2.5 py-1 rounded-xl text-[10px] transition-all shrink-0 flex items-center gap-1"
            >
              <span class="w-1.5 h-1.5 rounded-full bg-emerald-400"></span>
              Live ({{ countsByStatus.LIVE }})
            </button>
            <button
              @click="selectedStatusFilter = 'STANDBY'"
              :class="selectedStatusFilter === 'STANDBY' ? 'bg-amber-600 text-white font-black shadow-sm' : 'bg-amber-50 text-amber-700 hover:bg-amber-100 font-bold'"
              class="px-2.5 py-1 rounded-xl text-[10px] transition-all shrink-0 flex items-center gap-1"
            >
              <span class="w-1.5 h-1.5 rounded-full bg-amber-400"></span>
              Standby ({{ countsByStatus.STANDBY }})
            </button>
            <button
              @click="selectedStatusFilter = 'STALE'"
              :class="selectedStatusFilter === 'STALE' ? 'bg-orange-600 text-white font-black shadow-sm' : 'bg-orange-50 text-orange-700 hover:bg-orange-100 font-bold'"
              class="px-2.5 py-1 rounded-xl text-[10px] transition-all shrink-0 flex items-center gap-1"
            >
              <span class="w-1.5 h-1.5 rounded-full bg-orange-400"></span>
              Terputus ({{ countsByStatus.STALE }})
            </button>
            <button
              @click="selectedStatusFilter = 'OFFLINE'"
              :class="selectedStatusFilter === 'OFFLINE' ? 'bg-slate-600 text-white font-black shadow-sm' : 'bg-slate-100 text-slate-600 hover:bg-slate-200 font-bold'"
              class="px-2.5 py-1 rounded-xl text-[10px] transition-all shrink-0"
            >
              Offline ({{ countsByStatus.OFFLINE }})
            </button>
          </div>

          <!-- Category Filter Tabs -->
          <div class="grid grid-cols-4 gap-1 p-1 bg-slate-200/60 rounded-xl text-[11px] font-bold text-slate-600">
            <button
              @click="selectedCategoryFilter = 'ALL'"
              :class="{ 'bg-white text-slate-900 shadow-sm': selectedCategoryFilter === 'ALL' }"
              class="py-1 rounded-lg transition-all text-center"
            >
              Semua ({{ countsByCategory.ALL }})
            </button>
            <button
              @click="selectedCategoryFilter = 'SALES'"
              :class="{ 'bg-white text-primary shadow-sm': selectedCategoryFilter === 'SALES' }"
              class="py-1 rounded-lg transition-all text-center"
            >
              Sales ({{ countsByCategory.SALES }})
            </button>
            <button
              @click="selectedCategoryFilter = 'DRIVER'"
              :class="{ 'bg-white text-sky-600 shadow-sm': selectedCategoryFilter === 'DRIVER' }"
              class="py-1 rounded-lg transition-all text-center"
            >
              Driver ({{ countsByCategory.DRIVER }})
            </button>
            <button
              @click="selectedCategoryFilter = 'MARKETING'"
              :class="{ 'bg-white text-emerald-600 shadow-sm': selectedCategoryFilter === 'MARKETING' }"
              class="py-1 rounded-lg transition-all text-center"
            >
              Mkt ({{ countsByCategory.MARKETING }})
            </button>
          </div>

          <!-- Search Box -->
          <div class="relative w-full">
            <Search class="w-3.5 h-3.5 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Cari nama, toko, status, alamat..."
              class="w-full bg-white border border-slate-200 rounded-xl pl-9 pr-3 py-1.5 text-xs font-bold text-slate-900 outline-none focus:border-primary transition-colors"
            />
          </div>
        </div>

        <!-- Staff List -->
        <div class="flex-1 overflow-y-auto p-3 space-y-2.5 no-scrollbar">
          <div v-if="filteredConsolidatedList.length === 0" class="py-16 text-center px-6">
            <MapPin class="w-10 h-10 mx-auto mb-3 text-slate-300" />
            <p class="text-xs font-bold text-slate-400">Tidak ada petugas lapangan yang sesuai dengan filter.</p>
          </div>

          <div
            v-for="item in filteredConsolidatedList"
            :key="item.id"
            @click="panToMarker(item)"
            :class="item.hasCoordinates && item.freshnessTier !== 'NO_SIGNAL' ? 'hover:border-primary/40 hover:bg-primary/[0.03] cursor-pointer' : 'opacity-65 cursor-default'"
            class="p-3 rounded-2xl border border-slate-200 transition-all bg-white relative group shadow-sm"
          >
            <div class="flex items-start gap-3">
              <!-- Avatar with Status Dot -->
              <div class="relative shrink-0">
                <div
                  :style="`background-color: ${item.markerColor}`"
                  class="w-10 h-10 rounded-xl flex items-center justify-center font-black text-white text-xs shadow-sm"
                >
                  {{ item.name.charAt(0) }}
                </div>
                <span
                  v-if="item.freshnessTier === 'LIVE'"
                  class="absolute -bottom-1 -right-1 w-3.5 h-3.5 rounded-full bg-emerald-500 border-2 border-white animate-pulse"
                  title="Sinyal aktif (< 10 menit)"
                ></span>
                <span
                  v-else-if="item.freshnessTier === 'STANDBY'"
                  class="absolute -bottom-1 -right-1 w-3.5 h-3.5 rounded-full bg-amber-500 border-2 border-white"
                  title="Standby / Tidak bergerak"
                ></span>
                <span
                  v-else-if="item.freshnessTier === 'STALE'"
                  class="absolute -bottom-1 -right-1 w-3.5 h-3.5 rounded-full bg-orange-500 border-2 border-white"
                  title="Sinyal terputus (> 1 jam)"
                ></span>
              </div>

              <!-- Info -->
              <div class="flex-1 min-w-0">
                <div class="flex items-center justify-between gap-1">
                  <p class="font-extrabold text-slate-900 text-xs truncate">{{ item.name }}</p>
                  <span
                    v-if="item.time && isToday(item.time)"
                    class="text-[9px] font-bold shrink-0"
                    :class="{
                      'text-emerald-600': item.freshnessTier === 'LIVE',
                      'text-amber-600': item.freshnessTier === 'STANDBY',
                      'text-orange-500': item.freshnessTier === 'STALE',
                      'text-slate-400': item.freshnessTier === 'OFFLINE'
                    }"
                  >
                    {{ item.timeAgo }}
                  </span>
                  <span v-else class="text-[9px] font-bold text-slate-400 shrink-0">
                    -
                  </span>
                </div>

                <div class="flex items-center gap-1.5 mt-0.5">
                  <span class="text-[9px] font-bold text-slate-500 truncate">{{ item.jobTitle }}</span>
                  <span v-if="item.branchName" class="text-[8px] font-bold text-slate-400 truncate">• {{ item.branchName }}</span>
                </div>

                <!-- Status Badge -->
                <div class="mt-2 flex items-center gap-1.5 flex-wrap">
                  <!-- Live Badge -->
                  <span
                    v-if="item.freshnessTier === 'LIVE'"
                    class="text-[8px] font-black uppercase tracking-wider px-2 py-0.5 rounded-md border bg-emerald-50 text-emerald-700 border-emerald-200 flex items-center gap-1"
                  >
                    <Radio class="w-2.5 h-2.5 text-emerald-600 animate-pulse shrink-0" />
                    {{ item.status }}
                  </span>

                  <!-- Standby Badge -->
                  <span
                    v-else-if="item.freshnessTier === 'STANDBY'"
                    class="text-[8px] font-black uppercase tracking-wider px-2 py-0.5 rounded-md border bg-amber-50 text-amber-700 border-amber-200 flex items-center gap-1"
                  >
                    <PauseCircle class="w-2.5 h-2.5 text-amber-600 shrink-0" />
                    {{ item.status }}
                  </span>

                  <!-- Stale / Lost Signal Badge -->
                  <span
                    v-else-if="item.freshnessTier === 'STALE'"
                    class="text-[8px] font-black uppercase tracking-wider px-2 py-0.5 rounded-md border bg-orange-50 text-orange-700 border-orange-200 flex items-center gap-1"
                  >
                    <WifiOff class="w-2.5 h-2.5 text-orange-600 shrink-0" />
                    {{ item.status }}
                  </span>

                  <!-- Clock In Waiting GPS Badge -->
                  <span
                    v-else-if="!item.isOffline && item.dutyStatus?.includes('Aktif')"
                    class="text-[8px] font-black uppercase tracking-wider px-2 py-0.5 rounded-md border bg-blue-50 text-blue-700 border-blue-200 flex items-center gap-1"
                  >
                    ABSEN MASUK (MENUNGGU GPS)
                  </span>

                  <!-- Offline / Done Badge -->
                  <span
                    v-else
                    class="text-[8px] font-black uppercase tracking-wider px-2 py-0.5 rounded-md border bg-slate-100 text-slate-500 border-slate-200"
                  >
                    {{ item.status }}
                  </span>

                  <!-- Non-store location info -->
                  <span v-if="!item.storeName && item.hasCoordinates && item.freshnessTier !== 'NO_SIGNAL'" class="text-[9px] font-medium text-slate-600 truncate flex-1 flex items-center gap-1">
                    <MapPin class="w-2.5 h-2.5 text-slate-400 shrink-0" />
                    <span class="truncate">{{ item.locationName }}</span>
                  </span>
                  <span v-else-if="!item.hasCoordinates" class="text-[9px] font-medium text-slate-400 truncate flex-1 flex items-center gap-1">
                    <MapPin class="w-2.5 h-2.5 text-slate-300 shrink-0" />
                    <span class="truncate">Menunggu sinyal GPS HP</span>
                  </span>
                </div>

                <!-- Toko & Alamat Detail Card (Jika Kunjungan Toko) -->
                <div v-if="item.storeName" class="mt-2.5 p-2 bg-slate-50/80 rounded-xl border border-slate-100 space-y-1">
                  <div class="flex items-center gap-1.5">
                    <StoreIcon class="w-3 h-3 text-primary shrink-0" />
                    <span class="text-[11px] font-black text-slate-800 truncate">{{ item.storeName }}</span>
                    <span v-if="item.storeCategory" class="text-[7.5px] font-black px-1.5 py-0.2 bg-blue-100 text-blue-700 rounded shrink-0 uppercase">
                      {{ item.storeCategory }}
                    </span>
                  </div>
                  <div v-if="item.storeAddress" class="text-[9.5px] text-slate-500 font-medium flex items-start gap-1 leading-snug">
                    <MapPin class="w-2.5 h-2.5 text-slate-400 shrink-0 mt-0.5" />
                    <span class="line-clamp-2">{{ item.storeAddress }}</span>
                  </div>
                  <div v-if="item.storePhone" class="text-[8.5px] text-slate-400 font-medium flex items-center gap-1">
                    <Phone class="w-2.5 h-2.5 text-slate-400 shrink-0" />
                    <span>{{ item.storePhone }}</span>
                  </div>
                </div>

                <div v-if="item.visitResult" class="text-[9px] font-bold text-emerald-600 mt-1.5 truncate">
                  {{ item.visitResult }}
                </div>
              </div>

              <!-- Button FlyTo -->
              <button
                v-if="item.hasCoordinates && item.freshnessTier !== 'NO_SIGNAL'"
                class="opacity-0 group-hover:opacity-100 p-1 text-primary hover:bg-primary/10 rounded-lg transition-all self-center"
                title="Lihat di peta"
              >
                <Maximize2 class="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Right: Map -->
      <div
        class="flex-1 bg-white rounded-3xl border border-slate-200 shadow-sm overflow-hidden relative min-h-[450px]"
      >
        <div id="monitoring-map" class="w-full h-full z-10 min-h-[450px]"></div>

        <!-- Legend Overlay -->
        <div
          class="absolute bottom-6 left-6 z-20 bg-white/95 backdrop-blur-md p-4 rounded-3xl border border-slate-200 shadow-xl flex flex-col gap-2.5 min-w-[240px]"
        >
          <p class="text-[10px] font-black text-slate-400 uppercase tracking-[0.2em] mb-1">
            Status Kesegaran GPS
          </p>
          <div class="grid grid-cols-2 gap-x-3 gap-y-2 pb-2 border-b border-slate-100">
            <div class="flex items-center gap-2">
              <span class="w-3 h-3 rounded-full bg-emerald-500 animate-pulse border-2 border-white shadow-sm shrink-0"></span>
              <span class="text-[11px] font-black text-slate-700">Live (&lt; 10m)</span>
            </div>
            <div class="flex items-center gap-2">
              <span class="w-3 h-3 rounded-full bg-amber-500 border-2 border-white shadow-sm shrink-0"></span>
              <span class="text-[11px] font-black text-slate-700">Standby (10-60m)</span>
            </div>
            <div class="flex items-center gap-2">
              <span class="w-3 h-3 rounded-full bg-orange-500 border-2 border-white shadow-sm shrink-0"></span>
              <span class="text-[11px] font-black text-slate-700">Terputus (&gt; 1j)</span>
            </div>
            <div class="flex items-center gap-2">
              <span class="w-3 h-3 rounded-full bg-slate-400 border-2 border-white shadow-sm shrink-0"></span>
              <span class="text-[11px] font-black text-slate-700">Selesai Shift</span>
            </div>
          </div>

          <p class="text-[10px] font-black text-slate-400 uppercase tracking-[0.2em] mb-0.5">
            Kategori Petugas
          </p>
          <div class="grid grid-cols-2 gap-x-3 gap-y-1.5">
            <div class="flex items-center gap-2">
              <span class="w-2.5 h-2.5 rounded-full bg-[#0284c7] shrink-0"></span>
              <span class="text-[10px] font-bold text-slate-600">Driver</span>
            </div>
            <div class="flex items-center gap-2">
              <span class="w-2.5 h-2.5 rounded-full bg-[#3b82f6] shrink-0"></span>
              <span class="text-[10px] font-bold text-slate-600">Sales Motoris</span>
            </div>
            <div class="flex items-center gap-2">
              <span class="w-2.5 h-2.5 rounded-full bg-[#f59e0b] shrink-0"></span>
              <span class="text-[10px] font-bold text-slate-600">Task Order</span>
            </div>
            <div class="flex items-center gap-2">
              <span class="w-2.5 h-2.5 rounded-full bg-[#8b5cf6] shrink-0"></span>
              <span class="text-[10px] font-bold text-slate-600">Sales Booster</span>
            </div>
            <div class="flex items-center gap-2">
              <span class="w-2.5 h-2.5 rounded-full bg-[#ec4899] shrink-0"></span>
              <span class="text-[10px] font-bold text-slate-600">Modern Trade</span>
            </div>
            <div class="flex items-center gap-2">
              <span class="w-2.5 h-2.5 rounded-full bg-[#10b981] shrink-0"></span>
              <span class="text-[10px] font-bold text-slate-600">Marketing</span>
            </div>
          </div>

          <div class="mt-1 pt-2 border-t border-slate-100 flex items-center justify-between">
            <div class="flex items-center gap-2">
              <span class="text-[10px] font-bold text-slate-400">Pusatkan Peta:</span>
            </div>
            <button
              @click="fitAllMarkers"
              class="text-[10px] font-bold text-primary hover:underline flex items-center gap-1"
            >
              <Focus class="w-3 h-3" /> Reset View
            </button>
          </div>
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
          @click="zoomLevel = 1"
          class="p-3 bg-white/10 hover:bg-white/20 text-white rounded-2xl backdrop-blur-xl border border-white/10 transition-all"
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
          :style="`transform: scale(${zoomLevel})`"
          class="max-w-full max-h-full object-contain transition-transform duration-200 cursor-grab active:cursor-grabbing"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
#monitoring-map {
  background: #f8fafc;
}

/* Custom Leaflet Popup Styling */
:deep(.leaflet-popup-content-wrapper) {
  border-radius: 16px;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
  border: 1px solid #f1f5f9;
}

:deep(.leaflet-popup-tip) {
  box-shadow: none;
}

.no-scrollbar::-webkit-scrollbar {
  display: none;
}
.no-scrollbar {
  -ms-overflow-style: none;
  scrollbar-width: none;
}
</style>
