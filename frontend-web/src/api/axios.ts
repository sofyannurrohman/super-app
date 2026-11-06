import axios from 'axios'

const apiClient = axios.create({
  // This combines your .env variable (http://localhost:8081/api) with /v1
  // Result: http://localhost:8081/api/v1
  baseURL: `${import.meta.env.VITE_API_BASE_URL}/v1`,
  timeout: 30000, // 30s default timeout to prevent hanging requests
})

let activeRequests = 0
let safetyTimer: ReturnType<typeof setTimeout> | null = null

export const resetLoadingBar = () => {
  activeRequests = 0
  if (safetyTimer) {
    clearTimeout(safetyTimer)
    safetyTimer = null
  }
  const bar = document.getElementById('loading-bar')
  if (!bar) return
  bar.classList.remove('loading')
  bar.classList.add('done')
  setTimeout(() => {
    if (activeRequests === 0) bar.classList.remove('done')
  }, 500)
}

const updateLoadingBar = () => {
  const bar = document.getElementById('loading-bar')
  if (!bar) return

  if (activeRequests > 0) {
    bar.classList.remove('done')
    bar.classList.add('loading')

    // Safety fallback: if requests get stuck for more than 15 seconds, force-dismiss the loading bar
    if (!safetyTimer) {
      safetyTimer = setTimeout(() => {
        resetLoadingBar()
      }, 15000)
    }
  } else {
    if (safetyTimer) {
      clearTimeout(safetyTimer)
      safetyTimer = null
    }
    bar.classList.remove('loading')
    bar.classList.add('done')
    setTimeout(() => {
      if (activeRequests === 0) bar.classList.remove('done')
    }, 500)
  }
}

declare module 'axios' {
  export interface AxiosRequestConfig {
    skipLoadingBar?: boolean
  }
}

apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }

  // Auto inject sales_type and status for sales endpoints
  const salesType = localStorage.getItem('selectedSalesType')
  if (config.method?.toLowerCase() === 'get') {
    config.headers['Cache-Control'] = 'no-cache, no-store, must-revalidate'
    config.headers['Pragma'] = 'no-cache'
    config.headers['Expires'] = '0'

    if (salesType && salesType !== 'ALL_SALES' && (config.url?.includes('/admin/sales') || config.url?.includes('/sales'))) {
      config.params = config.params || {}
      if (!config.params.sales_type) {
        config.params.sales_type = salesType
      }
    }

    // Auto inject status=Active for fetching employees in sales tab
    if (config.url?.includes('/employees') && window.location.pathname.startsWith('/sales')) {
      config.params = config.params || {}
      if (!config.params.status) {
        config.params.status = 'Active'
      }
    }
  }

  if (!config.skipLoadingBar && !config.headers?.['X-Skip-Loading']) {
    activeRequests++
    updateLoadingBar()
  }
  return config
})

apiClient.interceptors.response.use(
  (response) => {
    if (!response.config?.skipLoadingBar && !response.config?.headers?.['X-Skip-Loading']) {
      activeRequests = Math.max(0, activeRequests - 1)
      updateLoadingBar()
    }
    return response
  },
  (error) => {
    if (!error.config?.skipLoadingBar && !error.config?.headers?.['X-Skip-Loading']) {
      activeRequests = Math.max(0, activeRequests - 1)
      updateLoadingBar()
    }

    // Handle 401 Unauthorized - Token expired or invalid
    if (error.response && error.response.status === 401) {
      const isPublicPage = window.location.pathname.includes('/login') || window.location.pathname.includes('/absensi/pekerja-harian') || window.location.pathname.includes('/public/')

      if (!isPublicPage) {
        localStorage.removeItem('token')
        localStorage.removeItem('user')
        window.location.href = '/login?session=expired'
      }
    }

    return Promise.reject(error)
  },
)

export default apiClient
