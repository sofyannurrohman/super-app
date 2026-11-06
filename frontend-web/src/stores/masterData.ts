import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import apiClient from '@/api/axios'
import { useAuthStore } from '@/stores/auth'

export const useMasterDataStore = defineStore('masterData', () => {
  const branches = ref<any[]>([])
  const departments = ref<any[]>([])
  const jobPositions = ref<any[]>([])
  const leaveTypes = ref<any[]>([])
  const employees = ref<any[]>([])
  const companies = ref<any[]>([])
  const stores = ref<any[]>([])

  const isBranchesLoading = ref<boolean>(false)
  const isDepartmentsLoading = ref<boolean>(false)
  const isJobPositionsLoading = ref<boolean>(false)
  const isLeaveTypesLoading = ref<boolean>(false)
  const isEmployeesLoading = ref<boolean>(false)
  const isCompaniesLoading = ref<boolean>(false)
  const isStoresLoading = ref<boolean>(false)
  
  const loadInitialBranchIds = (): string[] => {
    try {
      const storedMultiple = localStorage.getItem('selectedBranchIds')
      if (storedMultiple) {
        const parsed = JSON.parse(storedMultiple)
        if (Array.isArray(parsed) && parsed.length > 0) return parsed
      }
      const storedSingle = localStorage.getItem('selectedBranchId')
      if (storedSingle && storedSingle !== 'ALL_BRANCHES') {
        return storedSingle.split(',').map((s) => s.trim()).filter(Boolean)
      }
    } catch (e) {
      console.warn('Error loading initial branch IDs:', e)
    }
    return ['ALL_BRANCHES']
  }

  const rawSelectedBranchIds = ref<string[]>(loadInitialBranchIds())
  const selectedSalesType = ref<string>(localStorage.getItem('selectedSalesType') || 'ALL_SALES')
  const rawSelectedWarehouseBranchId = ref<string>(localStorage.getItem('selectedWarehouseBranchId') || 'ALL_BRANCHES')
  const rawSelectedBrandingBranchId = ref<string>(localStorage.getItem('selectedBrandingBranchId') || 'ALL_BRANCHES')

  const isUserBranchRestricted = (user: any) => {
    if (!user || !user.branch_id || user.has_global_access) return false
    if (
      user.role === 'superadmin' ||
      user.role === 'hr_admin' ||
      user.role === 'customer_support' ||
      user.role === 'admin_designer' ||
      user.role === 'validator_desain' ||
      user.role === 'validator_toko' ||
      user.role === 'admin_purchasing' ||
      user.role === 'purchasing' ||
      user.role === 'supervisor_gudang'
    ) {
      return false
    }
    if (user.role === 'auditor') {
      const b = branches.value.find((br: any) => br.id === user.branch_id)
      const isBooster = b ? (b.name || '').toLowerCase().includes('booster') : false
      return isBooster
    }
    return true
  }

  const selectedBranchIds = computed<string[]>({
    get() {
      const authStore = useAuthStore()
      const user = authStore.user
      if (isUserBranchRestricted(user)) {
        return [user.branch_id]
      }
      if (!rawSelectedBranchIds.value || rawSelectedBranchIds.value.length === 0) {
        return ['ALL_BRANCHES']
      }
      return rawSelectedBranchIds.value
    },
    set(val: string[]) {
      const authStore = useAuthStore()
      const user = authStore.user
      if (isUserBranchRestricted(user)) {
        return
      }
      let cleaned = (val || []).filter(Boolean)
      if (cleaned.length === 0 || cleaned.includes('ALL_BRANCHES')) {
        cleaned = ['ALL_BRANCHES']
      } else {
        // Hapus duplikat dan 'ALL_BRANCHES' jika ada ID spesifik
        cleaned = Array.from(new Set(cleaned.filter((id) => id !== 'ALL_BRANCHES')))
        if (cleaned.length === 0) cleaned = ['ALL_BRANCHES']
      }

      rawSelectedBranchIds.value = cleaned
      localStorage.setItem('selectedBranchIds', JSON.stringify(cleaned))
      if (cleaned.length === 1 && cleaned[0] === 'ALL_BRANCHES') {
        localStorage.setItem('selectedBranchId', 'ALL_BRANCHES')
      } else {
        localStorage.setItem('selectedBranchId', cleaned.join(','))
      }
    }
  })

  // Backwards-compatible computed for single/comma-separated string
  const selectedBranchId = computed<string, string | null>({
    get() {
      const ids = selectedBranchIds.value
      if (ids.length === 1 && ids[0] === 'ALL_BRANCHES') {
        return 'ALL_BRANCHES'
      }
      return ids.join(',')
    },
    set(val: string | null) {
      if (!val || val === 'ALL_BRANCHES') {
        selectedBranchIds.value = ['ALL_BRANCHES']
      } else {
        selectedBranchIds.value = val.split(',').map((s) => s.trim()).filter(Boolean)
      }
    }
  })

  const isAllBranchesSelected = computed(() => {
    return selectedBranchIds.value.length === 1 && selectedBranchIds.value[0] === 'ALL_BRANCHES'
  })

  const selectedBranchesCount = computed(() => {
    if (isAllBranchesSelected.value) return 0
    return selectedBranchIds.value.length
  })

  const isBranchSelected = (branchId: string): boolean => {
    if (branchId === 'ALL_BRANCHES') {
      return isAllBranchesSelected.value
    }
    return selectedBranchIds.value.includes(branchId)
  }

  const toggleBranchSelection = (branchId: string) => {
    if (branchId === 'ALL_BRANCHES') {
      selectedBranchIds.value = ['ALL_BRANCHES']
      return
    }

    let current = [...selectedBranchIds.value].filter((id) => id !== 'ALL_BRANCHES')
    if (current.includes(branchId)) {
      current = current.filter((id) => id !== branchId)
      if (current.length === 0) {
        selectedBranchIds.value = ['ALL_BRANCHES']
      } else {
        selectedBranchIds.value = current
      }
    } else {
      current.push(branchId)
      selectedBranchIds.value = current
    }
  }

  const setSelectedBranchIds = (ids: string[]) => {
    selectedBranchIds.value = ids
  }

  const selectAllBranches = () => {
    selectedBranchIds.value = ['ALL_BRANCHES']
  }

  const selectedBranchCompanyId = computed(() => {
    if (isAllBranchesSelected.value) return ''
    const firstId = selectedBranchIds.value[0]
    const branch = branches.value.find((b) => b.id === firstId)
    return branch?.company_id || ''
  })

  const selectedBranchName = computed(() => {
    if (isAllBranchesSelected.value) {
      return 'Semua Cabang'
    }
    const ids = selectedBranchIds.value
    if (ids.length === 1) {
      const branch = branches.value.find((b) => b.id === ids[0])
      return branch ? (branch.name || 'Semua Cabang') : 'Semua Cabang'
    }
    const names = ids
      .map((id) => branches.value.find((b) => b.id === id)?.name)
      .filter(Boolean)
    if (names.length === 0) {
      return 'Semua Cabang'
    }
    if (names.length === 2) {
      return names.join(', ')
    }
    return `${names.length} Cabang Dipilih`
  })

  const selectedBranchShortName = computed(() => {
    if (isAllBranchesSelected.value) {
      return null
    }
    const ids = selectedBranchIds.value
    if (ids.length === 1) {
      const branch = branches.value.find((b) => b.id === ids[0])
      if (!branch || !branch.name) return null
      const parts = branch.name.trim().split(/\s+/)
      return parts.length > 1 ? parts[parts.length - 1] : branch.name.substring(0, 8)
    }
    if (ids.length === 2) {
      const shortNames = ids.map((id) => {
        const b = branches.value.find((br) => br.id === id)
        if (!b || !b.name) return ''
        const parts = b.name.trim().split(/\s+/)
        return parts.length > 1 ? parts[parts.length - 1] : b.name.substring(0, 5)
      }).filter(Boolean)
      if (shortNames.length === 0) return null
      return shortNames.join(' + ')
    }
    return `${ids.length} Cabang`
  })

  const setSelectedBranchId = (id: string | null) => {
    selectedBranchId.value = id
  }

  const setSelectedSalesType = (type: string) => {
    const value = type || 'ALL_SALES'
    selectedSalesType.value = value
    localStorage.setItem('selectedSalesType', value)
  }

  const selectedWarehouseBranchId = computed<string>({
    get() {
      const authStore = useAuthStore()
      const user = authStore.user
      if (isUserBranchRestricted(user)) {
        return user.branch_id || 'ALL_BRANCHES'
      }
      return rawSelectedWarehouseBranchId.value || 'ALL_BRANCHES'
    },
    set(val: string) {
      const authStore = useAuthStore()
      const user = authStore.user
      if (isUserBranchRestricted(user)) {
        return
      }
      const value = val || 'ALL_BRANCHES'
      rawSelectedWarehouseBranchId.value = value
      localStorage.setItem('selectedWarehouseBranchId', value)
    }
  })

  const setSelectedWarehouseBranchId = (id: string) => {
    selectedWarehouseBranchId.value = id || 'ALL_BRANCHES'
  }

  const selectedBrandingBranchId = computed<string>({
    get() {
      const authStore = useAuthStore()
      const user = authStore.user
      if (isUserBranchRestricted(user)) {
        return user.branch_id || 'ALL_BRANCHES'
      }
      return rawSelectedBrandingBranchId.value || 'ALL_BRANCHES'
    },
    set(val: string) {
      const authStore = useAuthStore()
      const user = authStore.user
      if (isUserBranchRestricted(user)) {
        return
      }
      const value = val || 'ALL_BRANCHES'
      rawSelectedBrandingBranchId.value = value
      localStorage.setItem('selectedBrandingBranchId', value)
    }
  })

  const setSelectedBrandingBranchId = (id: string) => {
    selectedBrandingBranchId.value = id || 'ALL_BRANCHES'
  }

  const selectedBrandingBranchName = computed(() => {
    const id = selectedBrandingBranchId.value
    if (!id || id === 'ALL_BRANCHES') return 'Semua Cabang'
    const branch = branches.value.find((b) => String(b.id) === String(id))
    return branch ? (branch.name || 'Semua Cabang') : 'Semua Cabang'
  })

  const selectedBrandingBranchShortName = computed(() => {
    const id = selectedBrandingBranchId.value
    if (!id || id === 'ALL_BRANCHES') return null
    const branch = branches.value.find((b) => String(b.id) === String(id))
    if (!branch || !branch.name) return null
    const parts = branch.name.trim().split(/\s+/)
    return parts.length > 1 ? parts[parts.length - 1] : branch.name.substring(0, 8)
  })

  async function fetchBranches(force = false) {
    if (branches.value.length > 0 && !force) return
    isBranchesLoading.value = true
    try {
      const res = await apiClient.get('/branches')
      branches.value = res.data.data || []

      // If no branch is selected or selected branch doesn't exist,
      // and it's not ALL_BRANCHES, validate properly for both single and multi branch IDs
      const currentIds = selectedBranchIds.value
      const isAll = isAllBranchesSelected.value
      const isValid = isAll || currentIds.some((id) => branches.value.some((b) => b.id === id))

      if (!isValid && branches.value.length > 0) {
        setSelectedBranchId('ALL_BRANCHES')
      }
    } finally {
      isBranchesLoading.value = false
    }
  }

  async function fetchDepartments(force = false) {
    if (departments.value.length > 0 && !force) return
    isDepartmentsLoading.value = true
    try {
      const res = await apiClient.get('/departments')
      departments.value = res.data.data || []
    } finally {
      isDepartmentsLoading.value = false
    }
  }

  async function fetchJobPositions(force = false) {
    if (jobPositions.value.length > 0 && !force) return
    isJobPositionsLoading.value = true
    try {
      const res = await apiClient.get('/job-positions')
      jobPositions.value = res.data.data || []
    } finally {
      isJobPositionsLoading.value = false
    }
  }

  async function fetchLeaveTypes(force = false) {
    if (leaveTypes.value.length > 0 && !force) return
    isLeaveTypesLoading.value = true
    try {
      const res = await apiClient.get('/leave-types')
      leaveTypes.value = res.data.data || []
    } finally {
      isLeaveTypesLoading.value = false
    }
  }

  async function fetchEmployees(force = false) {
    if (employees.value.length > 0 && !force) return
    isEmployeesLoading.value = true
    try {
      // limit=500 cukup untuk keperluan dropdown/filter di view
      // Data ini hanya untuk UI filter, bukan untuk display tabel penuh
      const res = await apiClient.get('/employees?limit=500&status=Active')
      employees.value = res.data.data || []
    } catch (error) {
      console.error('Failed to fetch employees for master data store:', error)
    } finally {
      isEmployeesLoading.value = false
    }
  }

  async function fetchCompanies(force = false) {
    if (companies.value.length > 0 && !force) return
    isCompaniesLoading.value = true
    try {
      const res = await apiClient.get('/companies')
      companies.value = res.data.data || []
    } finally {
      isCompaniesLoading.value = false
    }
  }

  let lastFetchedStoreBranchKey = ''

  async function fetchStores(companyID?: string, force = false) {
    const branchKey = `${companyID || ''}_${selectedBranchId.value || 'ALL'}`
    if (!force && stores.value.length > 0 && lastFetchedStoreBranchKey === branchKey) {
      return
    }
    if (isStoresLoading.value) return
    isStoresLoading.value = true
    try {
      // minimal=true menghindari kalkulasi 6 subquery finansial yang berat untuk dropdown master data
      let url = `/stores?company_id=${companyID || ''}&limit=2500&status=ACTIVE&minimal=true`
      
      // Auto-filter by branch if selected
      if (selectedBranchId.value && selectedBranchId.value !== 'ALL_BRANCHES') {
        url += `&branch_id=${selectedBranchId.value}`
      }
      
      const res = await apiClient.get(url)
      stores.value = res.data.data || []
      lastFetchedStoreBranchKey = branchKey
    } catch (error) {
      console.error('Failed to fetch stores:', error)
    } finally {
      isStoresLoading.value = false
    }
  }

  return {
    branches,
    departments,
    jobPositions,
    leaveTypes,
    employees,
    companies,
    stores,
    selectedBranchId,
    selectedBranchIds,
    selectedSalesType,
    selectedWarehouseBranchId,
    selectedBrandingBranchId,
    isBranchesLoading,
    isDepartmentsLoading,
    isJobPositionsLoading,
    isLeaveTypesLoading,
    isEmployeesLoading,
    isCompaniesLoading,
    isStoresLoading,
    fetchBranches,
    fetchDepartments,
    fetchJobPositions,
    fetchLeaveTypes,
    fetchEmployees,
    fetchCompanies,
    fetchStores,
    setSelectedBranchId,
    setSelectedBranchIds,
    toggleBranchSelection,
    selectAllBranches,
    isBranchSelected,
    isAllBranchesSelected,
    selectedBranchesCount,
    setSelectedSalesType,
    setSelectedWarehouseBranchId,
    setSelectedBrandingBranchId,
    selectedBranchCompanyId,
    selectedBranchName,
    selectedBranchShortName,
    selectedBrandingBranchName,
    selectedBrandingBranchShortName,
  }
})
