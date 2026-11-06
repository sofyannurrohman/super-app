import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import apiClient from '@/api/axios'

const getInitialUser = () => {
  try {
    const raw = localStorage.getItem('user')
    if (!raw || raw === 'undefined' || raw === 'null') return null
    return JSON.parse(raw)
  } catch {
    return null
  }
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref(getInitialUser())

  const isAuthenticated = computed(() => !!token.value)

  const isAuthorized = computed(() => {
    const roles = [
      'superadmin',
      'admin',
      'HR',
      'hr_admin',
      'employee',
      'manager',
      'admin_nota',
      'supervisor_sales',
      'supervisor_sales_mt',
      'cashier',
      'admin_kasir',
      'admin_gudang',
      'auditor',
      'admin_marketplace',
      'admin_designer',
      'validator_desain',
      'validator_toko',
      'admin_purchasing',
      'customer_support',
      'admin_operasional',
      'admin_pabrik',
      'supervisor_gudang',
      'cashier_pusat',
    ]
    return user.value && roles.includes(user.value.role)
  })

  const hasPermission = (path: string) => {
    if (!user.value) return false
    if (user.value.role === 'superadmin' || user.value.role === 'customer_support') return true
    if (path.startsWith('/kehadiran')) return true // Personal attendance route

    if (!user.value.menus || !Array.isArray(user.value.menus) || user.value.menus.length === 0) {
      // Fallback sementara jika belum ada konfigurasi access_role
      return true
    }

    const dashboardRoots = ['/', '/factory', '/sales', '/warehouse', '/finance', '/marketplace', '/purchasing']

    return user.value.menus.some((menuPath: string) => {
      if (menuPath === path) return true
      if (dashboardRoots.includes(menuPath)) {
        return false // Root dashboards require exact path match
      }
      return path.startsWith(menuPath + '/')
    })
  }

  async function login(credentials: any) {
    try {
      // Assuming backend route: POST /auth/login
      const response = await apiClient.post('/login', credentials)
      token.value = response.data.data.token
      user.value = response.data.data.user
      localStorage.setItem('token', token.value)
      localStorage.setItem('user', JSON.stringify(user.value))
      return true
    } catch (error) {
      console.error('Login failed', error)
      return false
    }
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  return { token, user, isAuthenticated, isAuthorized, hasPermission, login, logout }
})
