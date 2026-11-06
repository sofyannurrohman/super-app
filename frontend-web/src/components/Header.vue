<script setup lang="ts">
import { useLayoutStore } from '@/stores/layout'
import { useMasterDataStore } from '@/stores/masterData'
import { marketplaceApi } from '@/api/marketplace'
import {
  Menu,
  Search,
  Bell,
  Briefcase,
  Users,
  PieChart,
  ChevronDown,
  Factory,
  Warehouse,
  Building2,
  Globe,
  Banknote,
  MapPin,
  UserCheck,
  Check,
  Megaphone,
  Lock,
  Unlock,
  ShoppingBag,
  ShoppingCart,
  Truck,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { computed, onMounted, ref } from 'vue'
import { isModuleEnabled } from '@/config/modules'
import { onClickOutside } from '@vueuse/core'
import NotificationDropdown from '@/components/NotificationDropdown.vue'
import RealtimeNotification from '@/components/RealtimeNotification.vue'
import AnnouncementHubDropdown from '@/components/AnnouncementHubDropdown.vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'

const router = useRouter()

const layoutStore = useLayoutStore()
const masterDataStore = useMasterDataStore()
const authStore = useAuthStore()
const marketplaceBranchIds = ref<string[]>([])
const isMarketplaceBranchesLoading = ref(false)
const marketplaceBranchesError = ref(false)

const marketplaceBranches = computed(() => {
  const linkedBranchIds = new Set(marketplaceBranchIds.value)
  return masterDataStore.branches.filter((branch) => linkedBranchIds.has(branch.id))
})

async function fetchMarketplaceBranches() {
  if (isMarketplaceBranchesLoading.value) return
  isMarketplaceBranchesLoading.value = true
  marketplaceBranchesError.value = false
  try {
    const { data } = await marketplaceApi.getMarketplaceAccounts()
    marketplaceBranchIds.value = data
      .filter((branch) => branch.accounts?.length > 0)
      .map((branch) => branch.branch_id)
  } catch (error) {
    console.error('Gagal memuat cabang marketplace yang tertaut', error)
    marketplaceBranchIds.value = []
    marketplaceBranchesError.value = true
  } finally {
    isMarketplaceBranchesLoading.value = false
  }
}

const handleHamburger = () => {
  if (window.innerWidth < 768) {
    layoutStore.openMobileDrawer()
  } else {
    layoutStore.toggleSidebar()
  }
}

onMounted(async () => {
  // 1. Wait completely for the API call/store logic to finish
  await masterDataStore.fetchBranches()

  // 2. Set default branch based on user branch restrictions
  const user = authStore.user
  const userBranch = masterDataStore.branches.find((b: any) => b.id === user?.branch_id)
  const isBoosterAuditor = user?.role === 'auditor' && userBranch && (userBranch.name || '').toLowerCase().includes('booster') && !user?.has_global_access

  const hasRestrictedBranch =
    user &&
    user.branch_id &&
    user.role !== 'superadmin' &&
    user.role !== 'customer_support' &&
    user.role !== 'hr_admin' &&
    (user.role !== 'auditor' || isBoosterAuditor) &&
    user.role !== 'admin_designer' &&
    user.role !== 'validator_desain' &&
    user.role !== 'validator_toko' &&
    user.role !== 'admin_purchasing' &&
    user.role !== 'purchasing' &&
    user.role !== 'supervisor_gudang' &&
    !user.has_global_access

  if (hasRestrictedBranch) {
    masterDataStore.setSelectedBranchId(user.branch_id)
  } else {
    // For global roles or branding roles, default to ALL_BRANCHES unless previously selected
    if (!localStorage.getItem('selectedBranchId')) {
      masterDataStore.setSelectedBranchId('ALL_BRANCHES')
    }
  }

  // 3. Set default module based on restricted roles
  if (user?.role === 'admin_marketplace') {
    layoutStore.setModule('marketplace')
  } else if (user?.role === 'admin_purchasing') {
    layoutStore.setModule('purchasing')
  } else if (user?.role === 'admin_gudang' || user?.role === 'admin_operasional' || user?.role === 'supervisor_gudang') {
    layoutStore.setModule('warehouse')
  } else if (user?.role === 'admin_pabrik') {
    layoutStore.setModule('factory')
  } else if (user?.role === 'cashier' || user?.role === 'admin_kasir') {
    layoutStore.setModule('finance')
  } else if (user?.role === 'supervisor_sales' || user?.role === 'admin_nota') {
    layoutStore.setModule('sales')
  } else if (user?.role === 'admin_designer' || user?.role === 'validator_desain' || user?.role === 'validator_toko') {
    layoutStore.setModule('branding')
  }

  if (layoutStore.currentModule === 'marketplace') {
    void fetchMarketplaceBranches()
  }
})

const modules = [
  { id: 'hris', name: 'HRIS', icon: Users, color: 'text-primary' },
  { id: 'factory', name: 'Pabrik', icon: Factory, color: 'text-slate-600' },
  { id: 'warehouse', name: 'Gudang', icon: Warehouse, color: 'text-slate-600' },
  { id: 'purchasing', name: 'Purchasing', icon: ShoppingCart, color: 'text-slate-600' },
  { id: 'sales', name: 'Sales', icon: PieChart, color: 'text-slate-600' },
  { id: 'marketplace', name: 'Marketplace', icon: ShoppingBag, color: 'text-slate-600' },
  { id: 'branding', name: 'Branding', icon: Megaphone, color: 'text-slate-600' },
  { id: 'finance', name: 'Keuangan', icon: Banknote, color: 'text-slate-600' },
  { id: 'rajatransport', name: 'Transport', icon: Truck, color: 'text-amber-600' },
]

const filteredModules = computed(() => {
  const role = authStore.user?.role
  const allowedModules = modules.filter((m) => isModuleEnabled(m.id))
  if (!role) return []

  if (role === 'supervisor_sales' || role === 'supervisor_sales_mt') {
    return allowedModules.filter((m) => m.id === 'sales' || m.id === 'branding')
  }
  if (role === 'admin_nota') {
    return allowedModules.filter((m) => m.id === 'sales')
  }
  if (role === 'cashier' || role === 'cashier_pusat') {
    return allowedModules.filter((m) => m.id === 'finance')
  }
  if (role === 'admin_kasir') {
    return allowedModules.filter((m) => m.id === 'finance' || m.id === 'hris')
  }
  if (role === 'admin_gudang') {
    return allowedModules.filter((m) => m.id === 'warehouse' || m.id === 'rajatransport')
  }
  if (role === 'hr_admin') {
    return allowedModules.filter((m) => m.id === 'hris')
  }
  if (role === 'admin_marketplace') {
    return allowedModules.filter((m) => m.id === 'marketplace')
  }
  if (role === 'admin_designer' || role === 'validator_desain' || role === 'validator_toko') {
    return allowedModules.filter((m) => m.id === 'branding')
  }
  if (role === 'admin_purchasing') {
    return allowedModules.filter((m) => m.id === 'purchasing' || m.id === 'factory' || m.id === 'branding' || m.id === 'finance')
  }
  if (role === 'admin_operasional') {
    return allowedModules.filter((m) => m.id === 'warehouse' || m.id === 'sales' || m.id === 'finance' || m.id === 'hris' || m.id === 'rajatransport')
  }
  if (role === 'admin_pabrik') {
    return allowedModules.filter((m) => m.id === 'factory' || m.id === 'warehouse' || m.id === 'rajatransport')
  }
  if (role === 'supervisor_gudang') {
    return allowedModules.filter((m) => m.id === 'warehouse' || m.id === 'factory' || m.id === 'rajatransport')
  }

  // Super Admin, Manager, Customer Support, etc. can see all modules
  return allowedModules
})

const currentModuleName = computed(() => {
  return filteredModules.value.find((m) => m.id === layoutStore.currentModule)?.name || 'HRIS'
})

const moduleDefaultRoutes: Record<string, string> = {
  hris: '/',
  purchasing: '/purchasing/requests',
  sales: '/sales',
  warehouse: '/warehouse',
  rajatransport: '/rajatransport',
  marketplace: '/marketplace',
  branding: '/sales/spanduk',
  finance: '/finance',
  factory: '/factory',
}

const handleModuleClick = (mod: any) => {
  if (mod.url) {
    window.open(mod.url, '_blank')
    return
  }
  layoutStore.setModule(mod.id)
  
  let defaultRoute = moduleDefaultRoutes[mod.id]
  
  // Custom logic for admin_kasir and admin_operasional to land on attendance instead of HRIS dashboard
  if (mod.id === 'hris' && (authStore.user?.role === 'admin_kasir' || authStore.user?.role === 'admin_operasional')) {
    defaultRoute = '/attendance'
  }

  if (defaultRoute) {
    router.push(defaultRoute)
  }
}

const handleModuleChange = (moduleId: string) => {
  const mod = filteredModules.value.find((m) => m.id === moduleId)
  if (mod) handleModuleClick(mod)
}

const isBranchSelectorLocked = computed(() => {
  const user = authStore.user
  if (!user || !user.branch_id || user.has_global_access) return false
  if (
    user.role === 'superadmin' ||
    user.role === 'customer_support' ||
    user.role === 'hr_admin' ||
    user.role === 'admin_designer' ||
    user.role !== 'validator_desain' &&
    user.role !== 'validator_toko' &&
    user.role !== 'admin_purchasing' &&
    user.role !== 'purchasing' &&
    user.role !== 'supervisor_gudang'
  ) {
    if (user.role === 'superadmin' || user.role === 'customer_support' || user.role === 'hr_admin' || user.role === 'admin_designer' || user.role === 'validator_desain' || user.role === 'validator_toko' || user.role === 'admin_purchasing' || user.role === 'purchasing' || user.role === 'supervisor_gudang') {
      return false
    }
  }
  if (user.role === 'auditor') {
    const b = masterDataStore.branches.find((br: any) => br.id === user.branch_id)
    const isBooster = b ? (b.name || '').toLowerCase().includes('booster') : false
    return isBooster
  }
  return true
})

// --- Sales Hover Submenu ---
const isSalesHovered = ref(false)
let salesHoverTimer: ReturnType<typeof setTimeout> | null = null

const salesSubOptions = [
  { value: 'ALL_SALES', label: 'Semua Tipe Sales', shortLabel: 'Semua Sales' },
  { value: 'sales_motoris_booster', label: 'Sales Motoris & Booster', shortLabel: 'Motoris' },
  { value: 'sales_to', label: 'Sales TO', shortLabel: 'TO' },
  { value: 'sales_mt', label: 'Sales MT', shortLabel: 'MT' },
]

const activeSalesLabel = computed(() => {
  const found = salesSubOptions.find((o) => o.value === masterDataStore.selectedSalesType)
  return found ? found.shortLabel : 'Semua Sales'
})

const hasSalesFilter = computed(() => masterDataStore.selectedSalesType !== 'ALL_SALES')

function onSalesMouseEnter() {
  if (salesHoverTimer) clearTimeout(salesHoverTimer)
  isSalesHovered.value = true
}

function onSalesMouseLeave() {
  salesHoverTimer = setTimeout(() => {
    isSalesHovered.value = false
  }, 180)
}

function selectSalesType(value: string) {
  masterDataStore.setSelectedSalesType(value)
  layoutStore.setModule('sales')
  isSalesHovered.value = false
}

function selectSalesBranch(branchId: string) {
  masterDataStore.toggleBranchSelection(branchId)
  layoutStore.setModule('sales')
}

// --- Warehouse Hover Branch Submenu ---
const isWarehouseHovered = ref(false)
let warehouseHoverTimer: ReturnType<typeof setTimeout> | null = null

const activeWarehouseBranchName = computed(() => {
  const id = masterDataStore.selectedWarehouseBranchId
  if (!id || id === 'ALL_BRANCHES') return null
  return masterDataStore.branches.find((b) => b.id === id)?.name ?? null
})

const activeWarehouseBranchShortName = computed(() => {
  if (!activeWarehouseBranchName.value) return null
  const parts = activeWarehouseBranchName.value.split(' ')
  return parts.length > 1 ? parts[parts.length - 1] : activeWarehouseBranchName.value.substring(0, 8)
})

const hasWarehouseBranchFilter = computed(
  () =>
    masterDataStore.selectedWarehouseBranchId !== 'ALL_BRANCHES' &&
    !!masterDataStore.selectedWarehouseBranchId,
)

function onWarehouseMouseEnter() {
  if (warehouseHoverTimer) clearTimeout(warehouseHoverTimer)
  isWarehouseHovered.value = true
}

function onWarehouseMouseLeave() {
  warehouseHoverTimer = setTimeout(() => {
    isWarehouseHovered.value = false
  }, 180)
}

function selectWarehouseBranch(branchId: string) {
  masterDataStore.setSelectedWarehouseBranchId(branchId)
  layoutStore.setModule('warehouse')
  isWarehouseHovered.value = false
}

// --- Finance Hover Branch Submenu ---
const isFinanceHovered = ref(false)
let financeHoverTimer: ReturnType<typeof setTimeout> | null = null

function onFinanceMouseEnter() {
  if (isBranchSelectorLocked.value) return
  if (financeHoverTimer) clearTimeout(financeHoverTimer)
  isFinanceHovered.value = true
}

function onFinanceMouseLeave() {
  financeHoverTimer = setTimeout(() => {
    isFinanceHovered.value = false
  }, 180)
}

function selectFinanceBranch(branchId: string) {
  masterDataStore.toggleBranchSelection(branchId)
  layoutStore.setModule('finance')
}

// --- Marketplace Hover Branch Submenu ---
const isMarketplaceHovered = ref(false)
let marketplaceHoverTimer: ReturnType<typeof setTimeout> | null = null

function onMarketplaceMouseEnter() {
  if (isBranchSelectorLocked.value) return
  if (marketplaceHoverTimer) clearTimeout(marketplaceHoverTimer)
  isMarketplaceHovered.value = true
  void fetchMarketplaceBranches()
}

function onMarketplaceMouseLeave() {
  marketplaceHoverTimer = setTimeout(() => {
    isMarketplaceHovered.value = false
  }, 180)
}

function selectMarketplaceBranch(branchId: string) {
  masterDataStore.toggleBranchSelection(branchId)
  layoutStore.setModule('marketplace')
}

// --- Branding Hover Branch Submenu ---
const isBrandingHovered = ref(false)
let brandingHoverTimer: ReturnType<typeof setTimeout> | null = null

const activeBrandingBranchName = computed(() => {
  return masterDataStore.selectedBrandingBranchName
})

const activeBrandingBranchShortName = computed(() => {
  return masterDataStore.selectedBrandingBranchShortName
})

const hasBrandingBranchFilter = computed(
  () =>
    !isBranchSelectorLocked.value &&
    masterDataStore.selectedBrandingBranchId !== 'ALL_BRANCHES' &&
    !!masterDataStore.selectedBrandingBranchId,
)

function onBrandingMouseEnter() {
  if (isBranchSelectorLocked.value) return
  if (brandingHoverTimer) clearTimeout(brandingHoverTimer)
  isBrandingHovered.value = true
}

function onBrandingMouseLeave() {
  brandingHoverTimer = setTimeout(() => {
    isBrandingHovered.value = false
  }, 180)
}

function selectBrandingBranch(branchId: string) {
  masterDataStore.setSelectedBrandingBranchId(branchId)
  layoutStore.setModule('branding')
  isBrandingHovered.value = false
}

// --- HRIS Hover Branch Submenu ---
const isHrisHovered = ref(false)
let hrisHoverTimer: ReturnType<typeof setTimeout> | null = null

const activeBranchName = computed(() => {
  return masterDataStore.selectedBranchName
})

const activeBranchShortName = computed(() => {
  return masterDataStore.selectedBranchShortName
})

const hasBranchFilter = computed(
  () =>
    !isBranchSelectorLocked.value &&
    !masterDataStore.isAllBranchesSelected,
)

function onHrisMouseEnter() {
  if (isBranchSelectorLocked.value) return
  if (hrisHoverTimer) clearTimeout(hrisHoverTimer)
  isHrisHovered.value = true
}

function onHrisMouseLeave() {
  hrisHoverTimer = setTimeout(() => {
    isHrisHovered.value = false
  }, 180)
}

function selectBranch(branchId: string) {
  masterDataStore.toggleBranchSelection(branchId)
  layoutStore.setModule('hris')
}

const selectAllGlobalBranches = () => {
  masterDataStore.selectAllBranches()
}

const selectAllSpecificBranches = () => {
  const allIds = masterDataStore.branches.map((b: any) => b.id)
  masterDataStore.setSelectedBranchIds(allIds)
}

const selectAllMarketplaceBranches = () => {
  if (marketplaceBranches.value.length === 0) return
  masterDataStore.setSelectedBranchIds(marketplaceBranches.value.map((branch) => branch.id))
}

const currentBranchOptions = computed(() =>
  layoutStore.currentModule === 'marketplace' ? marketplaceBranches.value : masterDataStore.branches,
)

function toggleMobileBranchDropdown() {
  isMobileBranchDropdownOpen.value = !isMobileBranchDropdownOpen.value
  if (isMobileBranchDropdownOpen.value && layoutStore.currentModule === 'marketplace') {
    void fetchMarketplaceBranches()
  }
}

// --- Global & Mobile Branch Filter Handlers ---
const isBranchFilterSupported = computed(() => {
  return ['hris', 'sales', 'warehouse', 'finance', 'marketplace', 'branding', 'purchasing'].includes(
    layoutStore.currentModule,
  )
})

const currentModuleBranchName = computed(() => {
  if (layoutStore.currentModule === 'warehouse') {
    const branchId = masterDataStore.selectedWarehouseBranchId
    if (!branchId || branchId === 'ALL_BRANCHES') return 'Semua Cabang'
    return masterDataStore.branches.find((b) => b.id === branchId)?.name || 'Semua Cabang'
  }
  if (layoutStore.currentModule === 'branding') {
    return masterDataStore.selectedBrandingBranchName
  }
  return masterDataStore.selectedBranchName
})

const currentModuleBranchShortName = computed(() => {
  if (layoutStore.currentModule === 'warehouse') {
    const branchId = masterDataStore.selectedWarehouseBranchId
    if (!branchId || branchId === 'ALL_BRANCHES') return 'Semua'
    const branch = masterDataStore.branches.find((b) => b.id === branchId)
    if (!branch) return 'Semua'
    const parts = branch.name.split(' ')
    return parts.length > 1 ? parts[parts.length - 1] : branch.name.substring(0, 8)
  }
  if (layoutStore.currentModule === 'branding') {
    return masterDataStore.selectedBrandingBranchShortName || 'Semua'
  }
  return masterDataStore.selectedBranchShortName || 'Semua'
})

const isMobileBranchDropdownOpen = ref(false)
const mobileBranchDropdownRef = ref<HTMLElement | null>(null)

onClickOutside(mobileBranchDropdownRef, () => {
  if (isMobileBranchDropdownOpen.value) {
    isMobileBranchDropdownOpen.value = false
  }
})

function handleGlobalBranchToggle(branchId: string) {
  if (layoutStore.currentModule === 'warehouse') {
    masterDataStore.setSelectedWarehouseBranchId(branchId)
    isMobileBranchDropdownOpen.value = false
  } else if (layoutStore.currentModule === 'branding') {
    masterDataStore.setSelectedBrandingBranchId(branchId)
    isMobileBranchDropdownOpen.value = false
  } else {
    masterDataStore.toggleBranchSelection(branchId)
  }
}
</script>

<template>
  <header
    class="sticky top-0 z-30 flex h-18 w-full items-center justify-between border-b border-slate-200/60 bg-white/70 px-3 sm:px-4 backdrop-blur-xl lg:px-8"
  >
    <div class="flex items-center gap-2 sm:gap-6">
      <Button
        variant="ghost"
        size="icon"
        class="md:hidden h-10 w-10 text-slate-500 hover:bg-slate-100/80 hover:text-primary transition-all duration-300 rounded-xl"
        @click="handleHamburger"
      >
        <Menu
          class="h-5.5 w-5.5 transition-transform duration-500"
          :class="{ 'rotate-180': layoutStore.isSidebarCollapsed }"
        />
        <span class="sr-only">Toggle Sidebar</span>
      </Button>

      <div
        class="hidden items-center gap-1 rounded-2xl bg-slate-100/50 p-1.5 md:flex border border-slate-200/50"
      >
        <template v-for="mod in filteredModules" :key="mod.id">
          <!-- Special HRIS Tab with hover branch submenu -->
          <div
            v-if="mod.id === 'hris'"
            class="relative"
            @mouseenter="onHrisMouseEnter"
            @mouseleave="onHrisMouseLeave"
          >
            <button
              @click="handleModuleClick(mod)"
              class="group flex items-center gap-2 rounded-xl px-3.5 py-1.5 transition-all duration-300 ease-out"
              :class="[
                layoutStore.currentModule === mod.id
                  ? 'bg-white text-primary shadow-lg shadow-primary/5 ring-1 ring-slate-200/50'
                  : 'text-slate-500 hover:bg-white/50 hover:text-slate-900',
              ]"
            >
              <component
                :is="mod.icon"
                class="h-4.5 w-4.5 shrink-0 transition-colors duration-300"
                :class="[
                  layoutStore.currentModule === mod.id
                    ? 'text-primary'
                    : 'text-slate-400 group-hover:text-slate-600',
                ]"
              />

              <!-- Tampilan Bertingkat (Atas-Bawah) saat Cabang Aktif -->
              <div
                v-if="isBranchSelectorLocked || hasBranchFilter"
                class="flex flex-col items-start text-left leading-tight"
              >
                <span class="text-xs font-bold leading-tight">{{ mod.name }}</span>
                <span
                  class="text-[10px] font-semibold leading-tight truncate max-w-[80px]"
                  :class="isBranchSelectorLocked ? 'text-amber-600' : 'text-blue-600'"
                >
                  {{ isBranchSelectorLocked ? (activeBranchName ?? 'Cabang') : activeBranchShortName }}
                </span>
              </div>

              <!-- Tampilan Normal (1 Baris) saat Semua Cabang -->
              <span v-else class="text-sm font-semibold">{{ mod.name }}</span>

              <!-- Chevron hanya jika bisa memilih cabang -->
              <ChevronDown
                v-if="!isBranchSelectorLocked"
                class="h-3.5 w-3.5 shrink-0 transition-transform duration-300"
                :class="[
                  isHrisHovered ? 'rotate-180 text-primary' : 'text-slate-400 group-hover:text-slate-600',
                  layoutStore.currentModule === mod.id ? 'text-primary' : '',
                ]"
              />
            </button>

            <!-- HRIS hover branch submenu panel -->
            <Transition
              enter-active-class="transition-all duration-200 ease-out"
              enter-from-class="opacity-0 translate-y-1 scale-95"
              enter-to-class="opacity-100 translate-y-0 scale-100"
              leave-active-class="transition-all duration-150 ease-in"
              leave-from-class="opacity-100 translate-y-0 scale-100"
              leave-to-class="opacity-0 translate-y-1 scale-95"
            >
              <div
                v-if="isHrisHovered && !isBranchSelectorLocked"
                class="absolute left-0 top-full mt-2 w-[340px] sm:w-[480px] max-w-[calc(100vw-32px)] origin-top-left rounded-2xl border border-slate-200/70 bg-white/95 backdrop-blur-xl shadow-xl shadow-slate-200/60 p-2.5 z-50 max-h-[80vh] overflow-y-auto before:absolute before:-top-3 before:left-0 before:right-0 before:h-3 before:content-['']"
              >
                <!-- Header label & Actions -->
                <div class="px-2 pt-1 pb-2 mb-1 flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <div class="flex h-7 w-7 items-center justify-center rounded-xl bg-primary/10">
                      <MapPin class="h-4 w-4 text-primary" />
                    </div>
                    <div>
                      <div class="flex items-center gap-1.5">
                        <p class="text-[11px] font-black uppercase tracking-wider text-slate-800">Filter Cabang HRIS</p>
                        <span
                          v-if="masterDataStore.selectedBranchesCount > 0"
                          class="bg-blue-100 text-blue-700 text-[10px] font-black px-1.5 py-0.2 rounded-md"
                        >
                          {{ masterDataStore.selectedBranchesCount }} Terpilih
                        </span>
                      </div>
                      <p class="text-[10px] text-slate-400 font-medium">Centang 1 atau beberapa cabang sekaligus</p>
                    </div>
                  </div>

                  <div class="flex items-center gap-1">
                    <button
                      type="button"
                      @click="selectAllSpecificBranches"
                      class="text-[10px] font-bold text-blue-600 hover:text-blue-800 hover:bg-blue-50 px-2 py-1 rounded-lg transition-colors cursor-pointer"
                    >
                      Pilih Semua
                    </button>
                    <span class="text-slate-200">|</span>
                    <button
                      type="button"
                      @click="selectAllGlobalBranches"
                      class="text-[10px] font-bold text-slate-500 hover:text-slate-800 hover:bg-slate-100 px-2 py-1 rounded-lg transition-colors cursor-pointer"
                    >
                      Reset
                    </button>
                  </div>
                </div>

                <div class="h-px bg-slate-100 mx-1 mb-2" />

                <!-- Semua Cabang -->
                <button
                  @click="selectAllGlobalBranches"
                  class="group/item flex w-full items-center gap-2.5 rounded-xl px-3 py-2 text-xs transition-all duration-150 mb-1.5"
                  :class="[
                    masterDataStore.isAllBranchesSelected
                      ? 'bg-primary/10 text-primary font-bold ring-1 ring-primary/20'
                      : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900',
                  ]"
                >
                  <span
                    class="flex h-4 w-4 shrink-0 items-center justify-center rounded-md border-2 transition-all duration-150"
                    :class="[
                      masterDataStore.isAllBranchesSelected
                        ? 'border-primary bg-primary text-white'
                        : 'border-slate-300 group-hover/item:border-slate-400',
                    ]"
                  >
                    <Check
                      v-if="masterDataStore.isAllBranchesSelected"
                      class="h-3 w-3 text-white"
                      :stroke-width="3"
                    />
                  </span>
                  <span class="flex-1 text-left font-bold">Semua Cabang (Global)</span>
                  <Unlock class="h-3.5 w-3.5 text-slate-300 group-hover/item:text-slate-500 transition-colors" />
                </button>

                <!-- Daftar Cabang (Grid 2 Kolom dengan Multi-Select Checkboxes) -->
                <div class="grid grid-cols-2 gap-1.5">
                  <button
                    v-for="branch in masterDataStore.branches"
                    :key="branch.id"
                    @click="selectBranch(branch.id)"
                    class="group/item flex items-center gap-2 rounded-xl px-2.5 py-2 text-xs transition-all duration-150 min-w-0"
                    :class="[
                      masterDataStore.isBranchSelected(branch.id)
                        ? 'bg-blue-50 text-blue-700 font-bold ring-1 ring-blue-200'
                        : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900',
                    ]"
                  >
                    <span
                      class="flex h-4 w-4 shrink-0 items-center justify-center rounded-md border-2 transition-all duration-150"
                      :class="[
                        masterDataStore.isBranchSelected(branch.id)
                          ? 'border-blue-600 bg-blue-600 text-white'
                          : 'border-slate-300 group-hover/item:border-slate-400 bg-white',
                      ]"
                    >
                      <Check
                        v-if="masterDataStore.isBranchSelected(branch.id)"
                        class="h-3 w-3 text-white"
                        :stroke-width="3"
                      />
                    </span>
                    <span class="flex-1 text-left truncate" :title="branch.name">{{ branch.name }}</span>
                  </button>
                </div>
              </div>
            </Transition>
          </div>

          <!-- Special Sales Tab with hover submenu -->
          <div
            v-else-if="mod.id === 'sales'"
            class="relative"
            @mouseenter="onSalesMouseEnter"
            @mouseleave="onSalesMouseLeave"
          >
            <button
              @click="handleModuleClick(mod)"
              class="group flex items-center gap-2 rounded-xl px-3.5 py-1.5 transition-all duration-300 ease-out"
              :class="[
                layoutStore.currentModule === mod.id
                  ? 'bg-white text-primary shadow-lg shadow-primary/5 ring-1 ring-slate-200/50'
                  : 'text-slate-500 hover:bg-white/50 hover:text-slate-900',
              ]"
            >
              <component
                :is="mod.icon"
                class="h-4.5 w-4.5 shrink-0 transition-colors duration-300"
                :class="[
                  layoutStore.currentModule === mod.id
                    ? 'text-primary'
                    : 'text-slate-400 group-hover:text-slate-600',
                ]"
              />

              <!-- Tampilan Bertingkat saat Cabang / Tipe Sales Aktif -->
              <div
                v-if="(!isBranchSelectorLocked && !masterDataStore.isAllBranchesSelected) || hasSalesFilter"
                class="flex flex-col items-start text-left leading-tight"
              >
                <span class="text-xs font-bold leading-tight">{{ mod.name }}</span>
                <span class="text-[10px] font-semibold leading-tight text-blue-600 truncate max-w-[85px]">
                  {{ !masterDataStore.isAllBranchesSelected ? activeBranchShortName : activeSalesLabel }}
                </span>
              </div>
              <span v-else class="text-sm font-semibold">{{ mod.name }}</span>

              <ChevronDown
                class="h-3.5 w-3.5 shrink-0 transition-transform duration-300"
                :class="[
                  isSalesHovered ? 'rotate-180 text-primary' : 'text-slate-400 group-hover:text-slate-600',
                  layoutStore.currentModule === mod.id ? 'text-primary' : '',
                ]"
              />
            </button>

            <!-- Hover submenu panel -->
            <Transition
              enter-active-class="transition-all duration-200 ease-out"
              enter-from-class="opacity-0 translate-y-1 scale-95"
              enter-to-class="opacity-100 translate-y-0 scale-100"
              leave-active-class="transition-all duration-150 ease-in"
              leave-from-class="opacity-100 translate-y-0 scale-100"
              leave-to-class="opacity-0 translate-y-1 scale-95"
            >
              <div
                v-if="isSalesHovered"
                class="absolute left-0 top-full mt-2 w-[340px] sm:w-[480px] max-w-[calc(100vw-32px)] origin-top-left rounded-2xl border border-slate-200/70 bg-white/95 backdrop-blur-xl shadow-xl shadow-slate-200/60 p-2.5 z-50 max-h-[80vh] overflow-y-auto before:absolute before:-top-3 before:left-0 before:right-0 before:h-3 before:content-['']"
              >
                <!-- Filter Cabang -->
                <template v-if="!isBranchSelectorLocked">
                  <div class="px-2 pt-1 pb-2 mb-1 flex items-center justify-between">
                    <div class="flex items-center gap-2">
                      <div class="flex h-7 w-7 items-center justify-center rounded-xl bg-blue-100">
                        <MapPin class="h-4 w-4 text-blue-600" />
                      </div>
                      <div>
                        <div class="flex items-center gap-1.5">
                          <p class="text-[11px] font-black uppercase tracking-wider text-slate-800">Filter Cabang Sales</p>
                          <span
                            v-if="masterDataStore.selectedBranchesCount > 0"
                            class="bg-blue-100 text-blue-700 text-[10px] font-black px-1.5 py-0.2 rounded-md"
                          >
                            {{ masterDataStore.selectedBranchesCount }} Terpilih
                          </span>
                        </div>
                        <p class="text-[10px] text-slate-400 font-medium">Pilih 1 atau beberapa cabang (dobel)</p>
                      </div>
                    </div>

                    <div class="flex items-center gap-1">
                      <button
                        type="button"
                        @click="selectAllSpecificBranches"
                        class="text-[10px] font-bold text-blue-600 hover:text-blue-800 hover:bg-blue-50 px-2 py-1 rounded-lg transition-colors cursor-pointer"
                      >
                        Pilih Semua
                      </button>
                      <span class="text-slate-200">|</span>
                      <button
                        type="button"
                        @click="selectAllGlobalBranches"
                        class="text-[10px] font-bold text-slate-500 hover:text-slate-800 hover:bg-slate-100 px-2 py-1 rounded-lg transition-colors cursor-pointer"
                      >
                        Reset
                      </button>
                    </div>
                  </div>

                  <!-- Semua Cabang -->
                  <button
                    @click="selectAllGlobalBranches"
                    class="group/item flex w-full items-center gap-2.5 rounded-xl px-3 py-2 text-xs transition-all duration-150 mb-1.5"
                    :class="[
                      masterDataStore.isAllBranchesSelected
                        ? 'bg-blue-50 text-blue-700 font-bold ring-1 ring-blue-200'
                        : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900',
                    ]"
                  >
                    <span
                      class="flex h-4 w-4 shrink-0 items-center justify-center rounded-md border-2 transition-all duration-150"
                      :class="[
                        masterDataStore.isAllBranchesSelected
                          ? 'border-blue-600 bg-blue-600 text-white'
                          : 'border-slate-300 group-hover/item:border-slate-400 bg-white',
                      ]"
                    >
                      <Check
                        v-if="masterDataStore.isAllBranchesSelected"
                        class="h-3 w-3 text-white"
                        :stroke-width="3"
                      />
                    </span>
                    <span class="flex-1 text-left font-bold">Semua Cabang (Global)</span>
                  </button>

                  <!-- Grid Daftar Cabang -->
                  <div class="grid grid-cols-2 gap-1.5 mb-3">
                    <button
                      v-for="branch in masterDataStore.branches"
                      :key="branch.id"
                      @click="selectSalesBranch(branch.id)"
                      class="group/item flex items-center gap-2 rounded-xl px-2.5 py-2 text-xs transition-all duration-150 min-w-0"
                      :class="[
                        masterDataStore.isBranchSelected(branch.id)
                          ? 'bg-blue-50 text-blue-700 font-bold ring-1 ring-blue-200'
                          : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900',
                      ]"
                    >
                      <span
                        class="flex h-4 w-4 shrink-0 items-center justify-center rounded-md border-2 transition-all duration-150"
                        :class="[
                          masterDataStore.isBranchSelected(branch.id)
                            ? 'border-blue-600 bg-blue-600 text-white'
                            : 'border-slate-300 group-hover/item:border-slate-400 bg-white',
                        ]"
                      >
                        <Check
                          v-if="masterDataStore.isBranchSelected(branch.id)"
                          class="h-3 w-3 text-white"
                          :stroke-width="3"
                        />
                      </span>
                      <span class="flex-1 text-left truncate" :title="branch.name">{{ branch.name }}</span>
                    </button>
                  </div>
                  
                  <div class="h-px bg-slate-100 mx-1 my-2" />
                </template>

                <!-- Filter Tipe Sales -->
                <div class="px-3 pt-1 pb-1 mb-1 flex items-center gap-2">
                  <div class="flex h-6 w-6 items-center justify-center rounded-lg bg-emerald-100">
                    <UserCheck class="h-3.5 w-3.5 text-emerald-600" />
                  </div>
                  <div>
                    <p class="text-[10px] font-bold uppercase tracking-widest text-slate-400">Filter Tipe Sales</p>
                  </div>
                </div>
                <div class="grid grid-cols-2 gap-1">
                  <button
                    v-for="opt in salesSubOptions"
                    :key="opt.value"
                    @click="selectSalesType(opt.value)"
                    class="group/item flex items-center gap-2 rounded-xl px-2.5 py-1.5 text-xs transition-all duration-150 min-w-0"
                    :class="[
                      masterDataStore.selectedSalesType === opt.value
                        ? 'bg-emerald-50 text-emerald-700 font-semibold ring-1 ring-emerald-100'
                        : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900',
                    ]"
                  >
                    <span
                      class="flex h-3.5 w-3.5 shrink-0 items-center justify-center rounded-full border-2 transition-all duration-150"
                      :class="[
                        masterDataStore.selectedSalesType === opt.value
                          ? 'border-emerald-500 bg-emerald-500'
                          : 'border-slate-200 group-hover/item:border-slate-400',
                      ]"
                    >
                      <Check
                        v-if="masterDataStore.selectedSalesType === opt.value"
                        class="h-2 w-2 text-white"
                        :stroke-width="3"
                      />
                    </span>
                    <span class="flex-1 text-left truncate">{{ opt.label }}</span>
                  </button>
                </div>
              </div>
            </Transition>
          </div>

          <!-- Special Warehouse Tab with hover branch submenu -->
          <div
            v-else-if="mod.id === 'warehouse'"
            class="relative"
            @mouseenter="onWarehouseMouseEnter"
            @mouseleave="onWarehouseMouseLeave"
          >
            <button
              @click="handleModuleClick(mod)"
              class="group flex items-center gap-2 rounded-xl px-3.5 py-1.5 transition-all duration-300 ease-out"
              :class="[
                layoutStore.currentModule === mod.id
                  ? 'bg-white text-primary shadow-lg shadow-primary/5 ring-1 ring-slate-200/50'
                  : 'text-slate-500 hover:bg-white/50 hover:text-slate-900',
              ]"
            >
              <component
                :is="mod.icon"
                class="h-4.5 w-4.5 shrink-0 transition-colors duration-300"
                :class="[
                  layoutStore.currentModule === mod.id
                    ? 'text-primary'
                    : 'text-slate-400 group-hover:text-slate-600',
                ]"
              />
              
              <!-- Tampilan Bertingkat saat Cabang Gudang Aktif -->
              <div
                v-if="hasWarehouseBranchFilter"
                class="flex flex-col items-start text-left leading-tight"
              >
                <span class="text-xs font-bold leading-tight">{{ mod.name }}</span>
                <span class="text-[10px] font-semibold leading-tight text-amber-700 truncate max-w-[80px]">
                  {{ activeWarehouseBranchShortName }}
                </span>
              </div>
              <span v-else class="text-sm font-semibold">{{ mod.name }}</span>

              <ChevronDown
                class="h-3.5 w-3.5 shrink-0 transition-transform duration-300"
                :class="[
                  isWarehouseHovered ? 'rotate-180 text-primary' : 'text-slate-400 group-hover:text-slate-600',
                  layoutStore.currentModule === mod.id ? 'text-primary' : '',
                ]"
              />
            </button>

            <!-- Warehouse hover branch submenu panel -->
            <Transition
              enter-active-class="transition-all duration-200 ease-out"
              enter-from-class="opacity-0 translate-y-1 scale-95"
              enter-to-class="opacity-100 translate-y-0 scale-100"
              leave-active-class="transition-all duration-150 ease-in"
              leave-from-class="opacity-100 translate-y-0 scale-100"
              leave-to-class="opacity-0 translate-y-1 scale-95"
            >
              <div
                v-if="isWarehouseHovered"
                class="absolute left-0 top-full mt-2 w-[340px] sm:w-[480px] max-w-[calc(100vw-32px)] origin-top-left rounded-2xl border border-slate-200/70 bg-white/95 backdrop-blur-xl shadow-xl shadow-slate-200/60 p-2 z-50 max-h-[80vh] overflow-y-auto before:absolute before:-top-3 before:left-0 before:right-0 before:h-3 before:content-['']"
              >
                <!-- Header label -->
                <div class="px-3 pt-1.5 pb-1 mb-1 flex items-center gap-2">
                  <div class="flex h-6 w-6 items-center justify-center rounded-lg bg-amber-100">
                    <Building2 class="h-3.5 w-3.5 text-amber-600" />
                  </div>
                  <div>
                    <p class="text-[10px] font-bold uppercase tracking-widest text-slate-400">Filter Cabang Gudang</p>
                    <p class="text-[11px] text-slate-500 font-medium">Pilih cabang untuk filter data</p>
                  </div>
                </div>

                <div class="h-px bg-slate-100 mx-2 mb-1.5" />

                <!-- Semua Cabang -->
                <button
                  @click="selectWarehouseBranch('ALL_BRANCHES')"
                  class="group/item flex w-full items-center gap-2.5 rounded-xl px-3 py-2 text-xs transition-all duration-150 mb-1"
                  :class="[
                    masterDataStore.selectedWarehouseBranchId === 'ALL_BRANCHES'
                      ? 'bg-amber-50 text-amber-700 font-semibold ring-1 ring-amber-100'
                      : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900',
                  ]"
                >
                  <span
                    class="flex h-3.5 w-3.5 shrink-0 items-center justify-center rounded-full border-2 transition-all duration-150"
                    :class="[
                      masterDataStore.selectedWarehouseBranchId === 'ALL_BRANCHES'
                        ? 'border-amber-500 bg-amber-500'
                        : 'border-slate-200 group-hover/item:border-slate-400',
                    ]"
                  >
                    <Check
                      v-if="masterDataStore.selectedWarehouseBranchId === 'ALL_BRANCHES'"
                      class="h-2 w-2 text-white"
                      :stroke-width="3"
                    />
                  </span>
                  <span class="flex-1 text-left font-medium">Semua Cabang</span>
                </button>

                <!-- Grid Daftar Cabang -->
                <div class="grid grid-cols-2 gap-1">
                  <button
                    v-for="branch in masterDataStore.branches"
                    :key="branch.id"
                    @click="selectWarehouseBranch(branch.id)"
                    class="group/item flex items-center gap-2 rounded-xl px-2.5 py-1.5 text-xs transition-all duration-150 min-w-0"
                    :class="[
                      masterDataStore.selectedWarehouseBranchId === branch.id
                        ? 'bg-amber-50 text-amber-700 font-semibold ring-1 ring-amber-100'
                        : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900',
                    ]"
                  >
                    <span
                      class="flex h-3.5 w-3.5 shrink-0 items-center justify-center rounded-full border-2 transition-all duration-150"
                      :class="[
                        masterDataStore.selectedWarehouseBranchId === branch.id
                          ? 'border-amber-500 bg-amber-500'
                          : 'border-slate-200 group-hover/item:border-slate-400',
                      ]"
                    >
                      <Check
                        v-if="masterDataStore.selectedWarehouseBranchId === branch.id"
                        class="h-2 w-2 text-white"
                        :stroke-width="3"
                      />
                    </span>
                    <span class="flex-1 text-left truncate" :title="branch.name">{{ branch.name }}</span>
                  </button>
                </div>
              </div>
            </Transition>
          </div>

          <!-- Special Finance Tab with hover branch submenu -->
          <div
            v-else-if="mod.id === 'finance'"
            class="relative"
            @mouseenter="onFinanceMouseEnter"
            @mouseleave="onFinanceMouseLeave"
          >
            <button
              @click="handleModuleClick(mod)"
              class="group flex items-center gap-2 rounded-xl px-3.5 py-1.5 transition-all duration-300 ease-out"
              :class="[
                layoutStore.currentModule === mod.id
                  ? 'bg-white text-primary shadow-lg shadow-primary/5 ring-1 ring-slate-200/50'
                  : 'text-slate-500 hover:bg-white/50 hover:text-slate-900',
              ]"
            >
              <component
                :is="mod.icon"
                class="h-4.5 w-4.5 shrink-0 transition-colors duration-300"
                :class="[
                  layoutStore.currentModule === mod.id
                    ? 'text-primary'
                    : 'text-slate-400 group-hover:text-slate-600',
                ]"
              />

              <!-- Tampilan Bertingkat saat Cabang Keuangan Aktif -->
              <div
                v-if="isBranchSelectorLocked || hasBranchFilter"
                class="flex flex-col items-start text-left leading-tight"
              >
                <span class="text-xs font-bold leading-tight">{{ mod.name }}</span>
                <span
                  class="text-[10px] font-semibold leading-tight truncate max-w-[80px]"
                  :class="isBranchSelectorLocked ? 'text-amber-600' : 'text-emerald-600'"
                >
                  {{ isBranchSelectorLocked ? (activeBranchName ?? 'Cabang') : activeBranchShortName }}
                </span>
              </div>
              <span v-else class="text-sm font-semibold">{{ mod.name }}</span>

              <ChevronDown
                v-if="!isBranchSelectorLocked"
                class="h-3.5 w-3.5 shrink-0 transition-transform duration-300"
                :class="[
                  isFinanceHovered ? 'rotate-180 text-primary' : 'text-slate-400 group-hover:text-slate-600',
                  layoutStore.currentModule === mod.id ? 'text-primary' : '',
                ]"
              />
            </button>

            <!-- Finance hover branch submenu panel -->
            <Transition
              enter-active-class="transition-all duration-200 ease-out"
              enter-from-class="opacity-0 translate-y-1 scale-95"
              enter-to-class="opacity-100 translate-y-0 scale-100"
              leave-active-class="transition-all duration-150 ease-in"
              leave-from-class="opacity-100 translate-y-0 scale-100"
              leave-to-class="opacity-0 translate-y-1 scale-95"
            >
              <div
                v-if="isFinanceHovered && !isBranchSelectorLocked"
                class="absolute right-0 top-full mt-2 w-[340px] sm:w-[460px] max-w-[calc(100vw-32px)] origin-top-right rounded-2xl border border-slate-200/70 bg-white/95 backdrop-blur-xl shadow-xl shadow-slate-200/60 p-2.5 z-50 max-h-[80vh] overflow-y-auto before:absolute before:-top-3 before:left-0 before:right-0 before:h-3 before:content-['']"
              >
                <!-- Header label & Actions -->
                <div class="px-2 pt-1 pb-2 mb-1 flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <div class="flex h-7 w-7 items-center justify-center rounded-xl bg-emerald-100">
                      <Banknote class="h-4 w-4 text-emerald-600" />
                    </div>
                    <div>
                      <div class="flex items-center gap-1.5">
                        <p class="text-[11px] font-black uppercase tracking-wider text-slate-800">Filter Cabang Keuangan</p>
                        <span
                          v-if="masterDataStore.selectedBranchesCount > 0"
                          class="bg-emerald-100 text-emerald-700 text-[10px] font-black px-1.5 py-0.2 rounded-md"
                        >
                          {{ masterDataStore.selectedBranchesCount }} Terpilih
                        </span>
                      </div>
                      <p class="text-[10px] text-slate-400 font-medium">Pilih 1 atau beberapa cabang sekaligus</p>
                    </div>
                  </div>

                  <div class="flex items-center gap-1">
                    <button
                      type="button"
                      @click="selectAllSpecificBranches"
                      class="text-[10px] font-bold text-emerald-600 hover:text-emerald-800 hover:bg-emerald-50 px-2 py-1 rounded-lg transition-colors cursor-pointer"
                    >
                      Pilih Semua
                    </button>
                    <span class="text-slate-200">|</span>
                    <button
                      type="button"
                      @click="selectAllGlobalBranches"
                      class="text-[10px] font-bold text-slate-500 hover:text-slate-800 hover:bg-slate-100 px-2 py-1 rounded-lg transition-colors cursor-pointer"
                    >
                      Reset
                    </button>
                  </div>
                </div>

                <div class="h-px bg-slate-100 mx-1 mb-2" />

                <!-- Semua Cabang -->
                <button
                  @click="selectAllGlobalBranches"
                  class="group/item flex w-full items-center gap-2.5 rounded-xl px-3 py-2 text-xs transition-all duration-150 mb-1.5"
                  :class="[
                    masterDataStore.isAllBranchesSelected
                      ? 'bg-emerald-50 text-emerald-700 font-bold ring-1 ring-emerald-200'
                      : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900',
                  ]"
                >
                  <span
                    class="flex h-4 w-4 shrink-0 items-center justify-center rounded-md border-2 transition-all duration-150"
                    :class="[
                      masterDataStore.isAllBranchesSelected
                        ? 'border-emerald-600 bg-emerald-600 text-white'
                        : 'border-slate-300 group-hover/item:border-slate-400 bg-white',
                    ]"
                  >
                    <Check
                      v-if="masterDataStore.isAllBranchesSelected"
                      class="h-3 w-3 text-white"
                      :stroke-width="3"
                    />
                  </span>
                  <span class="flex-1 text-left font-bold">Semua Cabang (Global)</span>
                </button>

                <!-- Grid Daftar Cabang -->
                <div class="grid grid-cols-2 gap-1.5">
                  <button
                    v-for="branch in masterDataStore.branches"
                    :key="branch.id"
                    @click="selectFinanceBranch(branch.id)"
                    class="group/item flex items-center gap-2 rounded-xl px-2.5 py-2 text-xs transition-all duration-150 min-w-0"
                    :class="[
                      masterDataStore.isBranchSelected(branch.id)
                        ? 'bg-emerald-50 text-emerald-700 font-bold ring-1 ring-emerald-200'
                        : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900',
                    ]"
                  >
                    <span
                      class="flex h-4 w-4 shrink-0 items-center justify-center rounded-md border-2 transition-all duration-150"
                      :class="[
                        masterDataStore.isBranchSelected(branch.id)
                          ? 'border-emerald-600 bg-emerald-600 text-white'
                          : 'border-slate-300 group-hover/item:border-slate-400 bg-white',
                      ]"
                    >
                      <Check
                        v-if="masterDataStore.isBranchSelected(branch.id)"
                        class="h-3 w-3 text-white"
                        :stroke-width="3"
                      />
                    </span>
                    <span class="flex-1 text-left truncate" :title="branch.name">{{ branch.name }}</span>
                  </button>
                </div>
              </div>
            </Transition>
          </div>

          <!-- Special Marketplace Tab with hover branch submenu -->
          <div
            v-else-if="mod.id === 'marketplace'"
            class="relative"
            @mouseenter="onMarketplaceMouseEnter"
            @mouseleave="onMarketplaceMouseLeave"
          >
            <button
              @click="handleModuleClick(mod)"
              class="group flex items-center gap-2 rounded-xl px-3.5 py-1.5 transition-all duration-300 ease-out"
              :class="[
                layoutStore.currentModule === mod.id
                  ? 'bg-white text-primary shadow-lg shadow-primary/5 ring-1 ring-slate-200/50'
                  : 'text-slate-500 hover:bg-white/50 hover:text-slate-900',
              ]"
            >
              <component
                :is="mod.icon"
                class="h-4.5 w-4.5 shrink-0 transition-colors duration-300"
                :class="[
                  layoutStore.currentModule === mod.id
                    ? 'text-primary'
                    : 'text-slate-400 group-hover:text-slate-600',
                ]"
              />

              <!-- Tampilan Bertingkat saat Cabang Marketplace Aktif -->
              <div
                v-if="isBranchSelectorLocked || hasBranchFilter"
                class="flex flex-col items-start text-left leading-tight"
              >
                <span class="text-xs font-bold leading-tight">{{ mod.name }}</span>
                <span
                  class="text-[10px] font-semibold leading-tight truncate max-w-[80px]"
                  :class="isBranchSelectorLocked ? 'text-amber-600' : 'text-orange-600'"
                >
                  {{ isBranchSelectorLocked ? (activeBranchName ?? 'Cabang') : activeBranchShortName }}
                </span>
              </div>
              <span v-else class="text-sm font-semibold">{{ mod.name }}</span>

              <ChevronDown
                v-if="!isBranchSelectorLocked"
                class="h-3.5 w-3.5 shrink-0 transition-transform duration-300"
                :class="[
                  isMarketplaceHovered ? 'rotate-180 text-primary' : 'text-slate-400 group-hover:text-slate-600',
                  layoutStore.currentModule === mod.id ? 'text-primary' : '',
                ]"
              />
            </button>

            <!-- Marketplace hover branch submenu panel -->
            <Transition
              enter-active-class="transition-all duration-200 ease-out"
              enter-from-class="opacity-0 translate-y-1 scale-95"
              enter-to-class="opacity-100 translate-y-0 scale-100"
              leave-active-class="transition-all duration-150 ease-in"
              leave-from-class="opacity-100 translate-y-0 scale-100"
              leave-to-class="opacity-0 translate-y-1 scale-95"
            >
              <div
                v-if="isMarketplaceHovered && !isBranchSelectorLocked"
                class="absolute left-0 top-full mt-2 w-[340px] sm:w-[480px] max-w-[calc(100vw-32px)] origin-top-left rounded-2xl border border-slate-200/70 bg-white/95 backdrop-blur-xl shadow-xl shadow-slate-200/60 p-2.5 z-50 max-h-[80vh] overflow-y-auto before:absolute before:-top-3 before:left-0 before:right-0 before:h-3 before:content-['']"
              >
                <!-- Header label & Actions -->
                <div class="px-2 pt-1 pb-2 mb-1 flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <div class="flex h-7 w-7 items-center justify-center rounded-xl bg-orange-100">
                      <ShoppingBag class="h-4 w-4 text-orange-600" />
                    </div>
                    <div>
                      <div class="flex items-center gap-1.5">
                        <p class="text-[11px] font-black uppercase tracking-wider text-slate-800">Filter Cabang Marketplace</p>
                        <span
                          v-if="masterDataStore.selectedBranchesCount > 0"
                          class="bg-orange-100 text-orange-700 text-[10px] font-black px-1.5 py-0.2 rounded-md"
                        >
                          {{ masterDataStore.selectedBranchesCount }} Terpilih
                        </span>
                      </div>
                      <p class="text-[10px] text-slate-400 font-medium">Pilih 1 atau beberapa cabang sekaligus</p>
                    </div>
                  </div>

                  <div class="flex items-center gap-1">
                    <button
                      type="button"
                      @click="selectAllMarketplaceBranches"
                      class="text-[10px] font-bold text-orange-600 hover:text-orange-800 hover:bg-orange-50 px-2 py-1 rounded-lg transition-colors cursor-pointer"
                    >
                      Pilih Semua
                    </button>
                    <span class="text-slate-200">|</span>
                    <button
                      type="button"
                      @click="selectAllGlobalBranches"
                      class="text-[10px] font-bold text-slate-500 hover:text-slate-800 hover:bg-slate-100 px-2 py-1 rounded-lg transition-colors cursor-pointer"
                    >
                      Reset
                    </button>
                  </div>
                </div>

                <div class="h-px bg-slate-100 mx-1 mb-2" />

                <!-- Semua Cabang -->
                <button
                  @click="selectAllGlobalBranches"
                  class="group/item flex w-full items-center gap-2.5 rounded-xl px-3 py-2 text-xs transition-all duration-150 mb-1.5"
                  :class="[
                    masterDataStore.isAllBranchesSelected
                      ? 'bg-orange-50 text-orange-700 font-bold ring-1 ring-orange-200'
                      : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900',
                  ]"
                >
                  <span
                    class="flex h-4 w-4 shrink-0 items-center justify-center rounded-md border-2 transition-all duration-150"
                    :class="[
                      masterDataStore.isAllBranchesSelected
                        ? 'border-orange-600 bg-orange-600 text-white'
                        : 'border-slate-300 group-hover/item:border-slate-400 bg-white',
                    ]"
                  >
                    <Check
                      v-if="masterDataStore.isAllBranchesSelected"
                      class="h-3 w-3 text-white"
                      :stroke-width="3"
                    />
                  </span>
                  <span class="flex-1 text-left font-bold">Semua Cabang (Global)</span>
                </button>

                <!-- Grid Daftar Cabang -->
                <div class="grid grid-cols-2 gap-1.5">
                  <button
                    v-for="branch in marketplaceBranches"
                    :key="branch.id"
                    @click="selectMarketplaceBranch(branch.id)"
                    class="group/item flex items-center gap-2 rounded-xl px-2.5 py-2 text-xs transition-all duration-150 min-w-0"
                    :class="[
                      masterDataStore.isBranchSelected(branch.id)
                        ? 'bg-orange-50 text-orange-700 font-bold ring-1 ring-orange-200'
                        : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900',
                    ]"
                  >
                    <span
                      class="flex h-4 w-4 shrink-0 items-center justify-center rounded-md border-2 transition-all duration-150"
                      :class="[
                        masterDataStore.isBranchSelected(branch.id)
                          ? 'border-orange-600 bg-orange-600 text-white'
                          : 'border-slate-300 group-hover/item:border-slate-400 bg-white',
                      ]"
                    >
                      <Check
                        v-if="masterDataStore.isBranchSelected(branch.id)"
                        class="h-3 w-3 text-white"
                        :stroke-width="3"
                      />
                    </span>
                    <span class="flex-1 text-left truncate" :title="branch.name">{{ branch.name }}</span>
                  </button>
                </div>
                <p v-if="isMarketplaceBranchesLoading" class="px-3 py-2 text-xs text-slate-500">Memuat cabang marketplace...</p>
                <p v-else-if="marketplaceBranchesError" class="px-3 py-2 text-xs text-red-600">Gagal memuat cabang marketplace. Buka kembali pilihan cabang untuk mencoba lagi.</p>
                <p v-else-if="marketplaceBranches.length === 0" class="px-3 py-2 text-xs text-slate-500">Belum ada cabang dengan akun marketplace tertaut.</p>
              </div>
            </Transition>
          </div>

          <!-- Special Branding Tab with hover branch submenu -->
          <div
            v-else-if="mod.id === 'branding'"
            class="relative"
            @mouseenter="onBrandingMouseEnter"
            @mouseleave="onBrandingMouseLeave"
          >
            <button
              @click="handleModuleClick(mod)"
              class="group flex items-center gap-2 rounded-xl px-3.5 py-1.5 transition-all duration-300 ease-out"
              :class="[
                layoutStore.currentModule === mod.id
                  ? 'bg-white text-primary shadow-lg shadow-primary/5 ring-1 ring-slate-200/50'
                  : 'text-slate-500 hover:bg-white/50 hover:text-slate-900',
              ]"
            >
              <component
                :is="mod.icon"
                class="h-4.5 w-4.5 shrink-0 transition-colors duration-300"
                :class="[
                  layoutStore.currentModule === mod.id
                    ? 'text-primary'
                    : 'text-slate-400 group-hover:text-slate-600',
                ]"
              />

              <!-- Tampilan Bertingkat saat Cabang Branding Aktif -->
              <div
                v-if="isBranchSelectorLocked || hasBrandingBranchFilter"
                class="flex flex-col items-start text-left leading-tight"
              >
                <span class="text-xs font-bold leading-tight">{{ mod.name }}</span>
                <span
                  class="text-[10px] font-semibold leading-tight truncate max-w-[80px]"
                  :class="isBranchSelectorLocked ? 'text-amber-600' : 'text-indigo-600'"
                >
                  {{ isBranchSelectorLocked ? (activeBranchName ?? 'Cabang') : activeBrandingBranchShortName }}
                </span>
              </div>
              <span v-else class="text-sm font-semibold">{{ mod.name }}</span>

              <ChevronDown
                v-if="!isBranchSelectorLocked"
                class="h-3.5 w-3.5 shrink-0 transition-transform duration-300"
                :class="[
                  isBrandingHovered ? 'rotate-180 text-primary' : 'text-slate-400 group-hover:text-slate-600',
                  layoutStore.currentModule === mod.id ? 'text-primary' : '',
                ]"
              />
            </button>

            <!-- Branding hover branch submenu panel -->
            <Transition
              enter-active-class="transition-all duration-200 ease-out"
              enter-from-class="opacity-0 translate-y-1 scale-95"
              enter-to-class="opacity-100 translate-y-0 scale-100"
              leave-active-class="transition-all duration-150 ease-in"
              leave-from-class="opacity-100 translate-y-0 scale-100"
              leave-to-class="opacity-0 translate-y-1 scale-95"
            >
              <div
                v-if="isBrandingHovered && !isBranchSelectorLocked"
                class="absolute right-0 top-full mt-2 w-[340px] sm:w-[460px] max-w-[calc(100vw-32px)] origin-top-right rounded-2xl border border-slate-200/70 bg-white/95 backdrop-blur-xl shadow-xl shadow-slate-200/60 p-2.5 z-50 max-h-[80vh] overflow-y-auto before:absolute before:-top-3 before:left-0 before:right-0 before:h-3 before:content-['']"
              >
                <!-- Header label & Actions -->
                <div class="px-2 pt-1 pb-2 mb-1 flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <div class="flex h-7 w-7 items-center justify-center rounded-xl bg-indigo-100">
                      <Megaphone class="h-4 w-4 text-indigo-600" />
                    </div>
                    <div>
                      <div class="flex items-center gap-1.5">
                        <p class="text-[11px] font-black uppercase tracking-wider text-slate-800">Filter Cabang Branding</p>
                      </div>
                      <p class="text-[10px] text-slate-400 font-medium">Pilih cabang untuk filter data branding</p>
                    </div>
                  </div>

                  <div class="flex items-center gap-1">
                    <button
                      type="button"
                      @click="selectBrandingBranch('ALL_BRANCHES')"
                      class="text-[10px] font-bold text-slate-500 hover:text-slate-800 hover:bg-slate-100 px-2 py-1 rounded-lg transition-colors cursor-pointer"
                    >
                      Reset
                    </button>
                  </div>
                </div>

                <div class="h-px bg-slate-100 mx-1 mb-2" />

                <!-- Semua Cabang -->
                <button
                  @click="selectBrandingBranch('ALL_BRANCHES')"
                  class="group/item flex w-full items-center gap-2.5 rounded-xl px-3 py-2 text-xs transition-all duration-150 mb-1.5"
                  :class="[
                    masterDataStore.selectedBrandingBranchId === 'ALL_BRANCHES'
                      ? 'bg-indigo-50 text-indigo-700 font-bold ring-1 ring-indigo-200'
                      : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900',
                  ]"
                >
                  <span
                    class="flex h-4 w-4 shrink-0 items-center justify-center rounded-md border-2 transition-all duration-150"
                    :class="[
                      masterDataStore.selectedBrandingBranchId === 'ALL_BRANCHES'
                        ? 'border-indigo-600 bg-indigo-600 text-white'
                        : 'border-slate-300 group-hover/item:border-slate-400 bg-white',
                    ]"
                  >
                    <Check
                      v-if="masterDataStore.selectedBrandingBranchId === 'ALL_BRANCHES'"
                      class="h-3 w-3 text-white"
                      :stroke-width="3"
                    />
                  </span>
                  <span class="flex-1 text-left font-bold">Semua Cabang (Global)</span>
                </button>

                <!-- Grid Daftar Cabang -->
                <div class="grid grid-cols-2 gap-1.5">
                  <button
                    v-for="branch in masterDataStore.branches"
                    :key="branch.id"
                    @click="selectBrandingBranch(branch.id)"
                    class="group/item flex items-center gap-2 rounded-xl px-2.5 py-2 text-xs transition-all duration-150 min-w-0"
                    :class="[
                      masterDataStore.selectedBrandingBranchId === branch.id
                        ? 'bg-indigo-50 text-indigo-700 font-bold ring-1 ring-indigo-200'
                        : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900',
                    ]"
                  >
                    <span
                      class="flex h-4 w-4 shrink-0 items-center justify-center rounded-md border-2 transition-all duration-150"
                      :class="[
                        masterDataStore.selectedBrandingBranchId === branch.id
                          ? 'border-indigo-600 bg-indigo-600 text-white'
                          : 'border-slate-300 group-hover/item:border-slate-400 bg-white',
                      ]"
                    >
                      <Check
                        v-if="masterDataStore.selectedBrandingBranchId === branch.id"
                        class="h-3 w-3 text-white"
                        :stroke-width="3"
                      />
                    </span>
                    <span class="flex-1 text-left truncate" :title="branch.name">{{ branch.name }}</span>
                  </button>
                </div>
              </div>
            </Transition>
          </div>

          <!-- Regular tabs (non-hris, non-sales, non-warehouse, non-finance, non-branding) -->
          <button
            v-else
            @click="handleModuleClick(mod)"
            class="group flex items-center gap-2 rounded-xl px-3.5 py-1.5 text-sm font-semibold transition-all duration-300 ease-out"
            :class="[
              layoutStore.currentModule === mod.id
                ? 'bg-white text-primary shadow-lg shadow-primary/5 ring-1 ring-slate-200/50'
                : 'text-slate-500 hover:bg-white/50 hover:text-slate-900',
            ]"
          >
            <component
              :is="mod.icon"
              class="h-4.5 w-4.5 shrink-0 transition-colors duration-300"
              :class="[
                layoutStore.currentModule === mod.id
                  ? 'text-primary'
                  : 'text-slate-400 group-hover:text-slate-600',
              ]"
            />
            <span>{{ mod.name }}</span>
          </button>
        </template>
      </div>

      <div class="flex md:hidden">
        <Select
          :model-value="layoutStore.currentModule"
          @update:model-value="(val) => handleModuleChange(val as string)"
        >
          <SelectTrigger
            class="flex h-10 items-center gap-1.5 rounded-xl border border-slate-200 bg-white px-2.5 py-2 text-xs sm:text-sm font-bold text-slate-800 shadow-sm active:scale-95 transition-transform max-w-[105px] sm:max-w-none"
          >
            <div class="flex items-center gap-2 truncate">
              <SelectValue :placeholder="currentModuleName" class="truncate" />
            </div>
          </SelectTrigger>
          <SelectContent class="rounded-xl border-slate-200 shadow-xl">
            <SelectItem v-for="mod in filteredModules" :key="mod.id" :value="mod.id" class="rounded-lg m-1">
              <div class="flex items-center gap-2">
                <component :is="mod.icon" class="h-4 w-4" :class="mod.color" />
                {{ mod.name }}
              </div>
            </SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>

    <div class="flex items-center gap-1.5 sm:gap-3 md:gap-6">
      <!-- Global Branch & Sales Filter (Muncul di Mobile untuk semua modul terkait cabang, dan di Desktop untuk modul Sales) -->
      <div
        v-if="isBranchFilterSupported"
        class="flex items-center gap-1.5 sm:gap-2"
        :class="{ 'md:hidden': layoutStore.currentModule !== 'sales' }"
      >
        <!-- Interactive Multi-Select Branch Dropdown Popover -->
        <div ref="mobileBranchDropdownRef" class="relative">
          <button
            type="button"
            @click="toggleMobileBranchDropdown"
            :disabled="isBranchSelectorLocked"
            class="flex items-center justify-between gap-1 sm:gap-2 w-[110px] min-[390px]:w-[130px] sm:w-[160px] md:w-[210px] h-9 sm:h-10 rounded-xl border border-slate-200 bg-white text-xs sm:text-sm font-semibold text-slate-700 shadow-sm transition-all hover:bg-slate-50 focus:ring-2 focus:ring-primary/20 px-2 sm:px-3 cursor-pointer"
          >
            <div class="flex items-center gap-1 sm:gap-1.5 truncate">
              <MapPin class="h-3.5 w-3.5 sm:h-4 sm:w-4 text-primary shrink-0" />
              <span class="truncate font-bold">
                {{ currentModuleBranchName }}
              </span>
            </div>
            <div class="flex items-center shrink-0">
              <Lock v-if="isBranchSelectorLocked" class="h-3 w-3 text-amber-500 mr-1" />
              <ChevronDown class="h-3.5 w-3.5 text-slate-400 transition-transform duration-200" :class="{ 'rotate-180': isMobileBranchDropdownOpen }" />
            </div>
          </button>

          <!-- Dropdown Popover Content -->

          <Transition
            enter-active-class="transition-all duration-200 ease-out"
            enter-from-class="opacity-0 translate-y-1 scale-95"
            enter-to-class="opacity-100 translate-y-0 scale-100"
            leave-active-class="transition-all duration-150 ease-in"
            leave-from-class="opacity-100 translate-y-0 scale-100"
            leave-to-class="opacity-0 translate-y-1 scale-95"
          >
            <div
              v-if="isMobileBranchDropdownOpen"
              class="absolute right-0 sm:left-0 top-full mt-2 w-[280px] sm:w-[340px] origin-top-right sm:origin-top-left rounded-2xl border border-slate-200/80 bg-white/98 backdrop-blur-xl shadow-2xl p-2.5 z-50 max-h-[70vh] overflow-y-auto"
            >
              <div class="px-2 pt-1 pb-2 mb-1 flex items-center justify-between border-b border-slate-100">
                <div class="flex items-center gap-1.5">
                  <span class="text-[11px] font-black uppercase text-slate-700">Pilih Cabang</span>
                  <span
                    v-if="layoutStore.currentModule === 'branding' ? (masterDataStore.selectedBrandingBranchId !== 'ALL_BRANCHES') : (layoutStore.currentModule === 'warehouse' ? (masterDataStore.selectedWarehouseBranchId !== 'ALL_BRANCHES') : masterDataStore.selectedBranchesCount > 0)"
                    class="bg-blue-100 text-blue-700 text-[10px] font-black px-1.5 py-0.2 rounded-md"
                  >
                    1 Terpilih
                  </span>
                </div>
                <div class="flex items-center gap-1">
                  <button
                    v-if="layoutStore.currentModule !== 'branding' && layoutStore.currentModule !== 'warehouse'"
                    type="button"
                    @click="layoutStore.currentModule === 'marketplace' ? selectAllMarketplaceBranches() : selectAllSpecificBranches()"
                    class="text-[10px] font-bold text-blue-600 hover:underline px-1 py-0.5 cursor-pointer"
                  >
                    Pilih Semua
                  </button>
                  <span v-if="layoutStore.currentModule !== 'branding' && layoutStore.currentModule !== 'warehouse'" class="text-slate-200">|</span>
                  <button
                    type="button"
                    @click="layoutStore.currentModule === 'branding' ? (masterDataStore.setSelectedBrandingBranchId('ALL_BRANCHES'), isMobileBranchDropdownOpen = false) : (layoutStore.currentModule === 'warehouse' ? (masterDataStore.setSelectedWarehouseBranchId('ALL_BRANCHES'), isMobileBranchDropdownOpen = false) : selectAllGlobalBranches())"
                    class="text-[10px] font-bold text-slate-500 hover:underline px-1 py-0.5 cursor-pointer"
                  >
                    Reset
                  </button>
                </div>
              </div>

              <!-- Semua Cabang Button -->
              <button
                type="button"
                @click="layoutStore.currentModule === 'branding' ? (masterDataStore.setSelectedBrandingBranchId('ALL_BRANCHES'), isMobileBranchDropdownOpen = false) : (layoutStore.currentModule === 'warehouse' ? (masterDataStore.setSelectedWarehouseBranchId('ALL_BRANCHES'), isMobileBranchDropdownOpen = false) : selectAllGlobalBranches())"
                class="flex w-full items-center gap-2.5 rounded-xl px-2.5 py-2 text-xs font-bold transition-all mb-1"
                :class="[
                  (layoutStore.currentModule === 'branding' ? masterDataStore.selectedBrandingBranchId === 'ALL_BRANCHES' : (layoutStore.currentModule === 'warehouse' ? masterDataStore.selectedWarehouseBranchId === 'ALL_BRANCHES' : masterDataStore.isAllBranchesSelected))
                    ? 'bg-blue-50 text-blue-700 ring-1 ring-blue-200'
                    : 'text-slate-600 hover:bg-slate-50',
                ]"
              >
                <span
                  class="flex h-4 w-4 shrink-0 items-center justify-center rounded-md border-2 transition-all"
                  :class="[
                    (layoutStore.currentModule === 'branding' ? masterDataStore.selectedBrandingBranchId === 'ALL_BRANCHES' : (layoutStore.currentModule === 'warehouse' ? masterDataStore.selectedWarehouseBranchId === 'ALL_BRANCHES' : masterDataStore.isAllBranchesSelected))
                      ? 'border-blue-600 bg-blue-600 text-white'
                      : 'border-slate-300 bg-white',
                  ]"
                >
                  <Check
                    v-if="layoutStore.currentModule === 'branding' ? masterDataStore.selectedBrandingBranchId === 'ALL_BRANCHES' : (layoutStore.currentModule === 'warehouse' ? masterDataStore.selectedWarehouseBranchId === 'ALL_BRANCHES' : masterDataStore.isAllBranchesSelected)"
                    class="h-3 w-3 text-white"
                    :stroke-width="3"
                  />
                </span>
                <span>Semua Cabang (Global)</span>
              </button>

              <!-- Branch Items -->
              <div class="space-y-1">
                <button
                  v-for="branch in currentBranchOptions"
                  :key="branch.id"
                  type="button"
                  @click="handleGlobalBranchToggle(branch.id)"
                  class="flex w-full items-center gap-2.5 rounded-xl px-2.5 py-1.5 text-xs font-medium transition-all"
                  :class="[
                    (layoutStore.currentModule === 'branding' ? masterDataStore.selectedBrandingBranchId === branch.id : (layoutStore.currentModule === 'warehouse' ? masterDataStore.selectedWarehouseBranchId === branch.id : masterDataStore.isBranchSelected(branch.id)))
                      ? 'bg-blue-50 text-blue-700 font-bold ring-1 ring-blue-200'
                      : 'text-slate-600 hover:bg-slate-50',
                  ]"
                >
                  <span
                    class="flex h-4 w-4 shrink-0 items-center justify-center rounded-md border-2 transition-all"
                    :class="[
                      (layoutStore.currentModule === 'branding' ? masterDataStore.selectedBrandingBranchId === branch.id : (layoutStore.currentModule === 'warehouse' ? masterDataStore.selectedWarehouseBranchId === branch.id : masterDataStore.isBranchSelected(branch.id)))
                        ? 'border-blue-600 bg-blue-600 text-white'
                        : 'border-slate-300 bg-white',
                    ]"
                  >
                    <Check
                      v-if="layoutStore.currentModule === 'branding' ? masterDataStore.selectedBrandingBranchId === branch.id : (layoutStore.currentModule === 'warehouse' ? masterDataStore.selectedWarehouseBranchId === branch.id : masterDataStore.isBranchSelected(branch.id))"
                      class="h-3 w-3 text-white"
                      :stroke-width="3"
                    />
                  </span>
                  <span class="truncate text-left">{{ branch.name }}</span>
                </button>
                <p v-if="layoutStore.currentModule === 'marketplace' && isMarketplaceBranchesLoading" class="px-2 py-2 text-xs text-slate-500">Memuat cabang marketplace...</p>
                <p v-else-if="layoutStore.currentModule === 'marketplace' && marketplaceBranchesError" class="px-2 py-2 text-xs text-red-600">Gagal memuat cabang marketplace. Buka kembali pilihan cabang untuk mencoba lagi.</p>
                <p v-else-if="layoutStore.currentModule === 'marketplace' && marketplaceBranches.length === 0" class="px-2 py-2 text-xs text-slate-500">Belum ada cabang dengan akun marketplace tertaut.</p>
              </div>
            </div>
          </Transition>
        </div>

        <!-- Tipe Sales Selector (Hanya muncul di modul Sales) -->
        <Select
          v-if="layoutStore.currentModule === 'sales'"
          :model-value="masterDataStore.selectedSalesType"
          @update:model-value="(val) => masterDataStore.setSelectedSalesType(val as string)"
        >
          <SelectTrigger
            class="w-[95px] min-[390px]:w-[115px] sm:w-[150px] md:w-[210px] h-9 sm:h-10 rounded-xl border border-slate-200 bg-white text-xs sm:text-sm font-semibold text-slate-700 shadow-sm transition-all hover:bg-slate-50 focus:ring-2 focus:ring-primary/20 px-2 sm:px-3"
          >
            <div class="flex items-center gap-1 sm:gap-2 truncate">
              <UserCheck class="h-3.5 w-3.5 sm:h-4 sm:w-4 text-emerald-500 shrink-0" />
              <SelectValue placeholder="Semua Sales" class="truncate">
                <span class="hidden sm:inline">
                  {{
                    masterDataStore.selectedSalesType === 'sales_motoris_booster'
                      ? 'Sales Motoris & Booster'
                      : masterDataStore.selectedSalesType === 'sales_to'
                      ? 'Sales TO'
                      : masterDataStore.selectedSalesType === 'sales_mt'
                      ? 'Sales MT'
                      : 'Semua Tipe Sales'
                  }}
                </span>
                <span class="sm:hidden truncate">
                  {{
                    masterDataStore.selectedSalesType === 'sales_motoris_booster'
                      ? 'Motoris'
                      : masterDataStore.selectedSalesType === 'sales_to'
                      ? 'TO'
                      : masterDataStore.selectedSalesType === 'sales_mt'
                      ? 'MT'
                      : 'Semua'
                  }}
                </span>
              </SelectValue>
            </div>
          </SelectTrigger>
          <SelectContent class="rounded-xl border-slate-200 shadow-xl max-h-[300px] z-50">
            <SelectItem value="ALL_SALES" class="rounded-lg m-1 cursor-pointer font-medium text-primary text-xs sm:text-sm">
              Semua Tipe Sales
            </SelectItem>
            <SelectItem value="sales_motoris_booster" class="rounded-lg m-1 cursor-pointer text-xs sm:text-sm">
              Sales Motoris & Booster
            </SelectItem>
            <SelectItem value="sales_to" class="rounded-lg m-1 cursor-pointer text-xs sm:text-sm">
              Sales TO
            </SelectItem>
            <SelectItem value="sales_mt" class="rounded-lg m-1 cursor-pointer text-xs sm:text-sm">
              Sales MT
            </SelectItem>
          </SelectContent>
        </Select>
      </div>

      <!-- Hub Pengumuman Perusahaan (Megaphone) -->
      <AnnouncementHubDropdown
        v-if="authStore.user"
      />

      <!-- Realtime Notification — untuk semua user yang terotentikasi -->
      <RealtimeNotification
        v-if="authStore.user"
      />
    </div>
  </header>
</template>

<style scoped>
button {
  -webkit-tap-highlight-color: transparent;
}
</style>
