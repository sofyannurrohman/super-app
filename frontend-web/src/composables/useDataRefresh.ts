import { ref, computed, onMounted, onUnmounted } from 'vue'
import { toast } from 'vue-sonner'

export interface UseDataRefreshOptions {
  /** Pesan toast ketika pengguna menekan tombol Segarkan Data secara manual */
  successMessage?: string
  /** Cooldown minimum (ms) untuk auto-refresh saat kembali ke tab/window */
  focusCooldownMs?: number
  /** Aktifkan auto-refresh saat tab kembali fokus / visible (default: true) */
  autoRefreshOnFocus?: boolean
  /** Guard opsional untuk mencegah auto-refresh saat modal form / proses submit sedang berjalan */
  canAutoRefresh?: () => boolean
}

export function useDataRefresh(
  refreshFn: (context: { isManual: boolean }) => Promise<unknown>,
  options: UseDataRefreshOptions = {},
) {
  const {
    successMessage = 'Data berhasil diperbarui',
    focusCooldownMs = 15000,
    autoRefreshOnFocus = true,
    canAutoRefresh,
  } = options

  const isRefreshing = ref(false)
  const lastRefreshedAt = ref<Date | null>(new Date())

  const lastRefreshedTimeStr = computed(() => {
    if (!lastRefreshedAt.value) return ''
    return lastRefreshedAt.value.toLocaleTimeString('id-ID', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    })
  })

  const markRefreshed = () => {
    lastRefreshedAt.value = new Date()
  }

  const refresh = async (opts?: { silent?: boolean }) => {
    if (isRefreshing.value) return
    const isManual = !opts?.silent

    isRefreshing.value = true
    try {
      await refreshFn({ isManual })
      lastRefreshedAt.value = new Date()
      if (isManual && successMessage) {
        toast.success(successMessage)
      }
    } catch (err: any) {
      console.error('Failed to refresh data:', err)
      if (isManual) {
        toast.error(err?.response?.data?.error || 'Gagal menyegarkan data terkini')
      }
    } finally {
      isRefreshing.value = false
    }
  }

  const handleVisibilityOrFocus = () => {
    if (!autoRefreshOnFocus) return
    if (typeof document !== 'undefined' && document.visibilityState === 'hidden') return
    if (isRefreshing.value) return
    if (canAutoRefresh && !canAutoRefresh()) return

    const now = Date.now()
    const lastTime = lastRefreshedAt.value ? lastRefreshedAt.value.getTime() : 0
    if (now - lastTime < focusCooldownMs) return

    void refresh({ silent: true })
  }

  onMounted(() => {
    if (autoRefreshOnFocus && typeof window !== 'undefined') {
      window.addEventListener('focus', handleVisibilityOrFocus)
      document.addEventListener('visibilitychange', handleVisibilityOrFocus)
    }
  })

  onUnmounted(() => {
    if (typeof window !== 'undefined') {
      window.removeEventListener('focus', handleVisibilityOrFocus)
      document.removeEventListener('visibilitychange', handleVisibilityOrFocus)
    }
  })

  return {
    isRefreshing,
    lastRefreshedAt,
    lastRefreshedTimeStr,
    markRefreshed,
    refresh,
  }
}
