<script setup lang="ts">
import { ref, onMounted, onUnmounted, h } from 'vue'
import apiClient from '@/api/axios'
import DataTable from '@/components/DataTable.vue'
import { toast } from 'vue-sonner'

const activeUsers = ref<any[]>([])
const isLoading = ref(true)
let intervalId: any = null

const fetchActiveUsers = async () => {
  try {
    const response = await apiClient.get('/users/active-users')
    const rawData = response.data.data
    activeUsers.value = Array.isArray(rawData) ? rawData : rawData?.data || []
  } catch (error: any) {
    console.error('Failed to fetch active users:', error)
    toast.error(
      'Gagal mengambil data pengguna aktif: ' + (error.response?.data?.message || error.message),
    )
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  fetchActiveUsers()
  // Poll every 15 seconds
  intervalId = setInterval(fetchActiveUsers, 15000)
})

onUnmounted(() => {
  if (intervalId) clearInterval(intervalId)
})

const formatTime = (timestamp: number) => {
  if (!timestamp) return '-'
  const date = new Date(timestamp * 1000)
  return date.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' }) + ' (' + date.toLocaleDateString('id-ID') + ')'
}

const getTimeAgo = (timestamp: number) => {
  if (!timestamp) return 'Tidak diketahui'
  const diffInSeconds = Math.floor(Date.now() / 1000) - timestamp
  
  if (diffInSeconds <= 45) return 'Online'
  if (diffInSeconds < 60) return `${diffInSeconds} detik yang lalu`
  if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)} menit yang lalu`
  if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)} jam yang lalu`
  return `${Math.floor(diffInSeconds / 86400)} hari yang lalu`
}

const columns = [
  {
    accessorKey: 'name',
    header: 'Nama',
    cell: ({ row }: any) => {
      const name = row.getValue('name') || 'Unknown'
      return name
    },
  },
  {
    accessorKey: 'role',
    header: 'Role',
    cell: ({ row }: any) => {
      const val = row.getValue('role')
      let label = val
      switch (val) {
        case 'superadmin': label = 'Super Admin'; break
        case 'employee': label = 'Karyawan'; break
        case 'manager': label = 'Manager'; break
        case 'hr_admin': label = 'HR Admin'; break
        case 'supervisor_sales': label = 'Supervisor Sales'; break
        case 'cashier': label = 'Cashier'; break
        case 'admin_kasir': label = 'Admin Kasir'; break
        case 'admin_gudang': label = 'Admin Gudang'; break
      }
      return h('div', { class: 'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold bg-blue-100 text-blue-800' }, label)
    },
  },
  {
    accessorKey: 'last_active',
    header: 'Terakhir Aktif',
    cell: ({ row }: any) => {
      const ts = row.getValue('last_active')
      return formatTime(ts)
    },
  },
  {
    id: 'status',
    header: 'Status',
    cell: ({ row }: any) => {
      const ts = row.original.last_active
      const timeAgoStr = getTimeAgo(ts)
      const diffInSeconds = Math.floor(Date.now() / 1000) - ts

      if (diffInSeconds <= 45) {
        // Online (Hijau Berkedip)
        return h('div', { class: 'flex items-center' }, [
          h('span', { class: 'relative flex h-3 w-3 mr-2' }, [
            h('span', { class: 'animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75' }),
            h('span', { class: 'relative inline-flex rounded-full h-3 w-3 bg-green-500' })
          ]),
          h('span', { class: 'text-sm font-medium text-green-700' }, 'Online')
        ])
      } else {
        // Offline (Abu-abu)
        return h('div', { class: 'flex items-center' }, [
          h('span', { class: 'relative flex h-2 w-2 mr-2' }, [
            h('span', { class: 'relative inline-flex rounded-full h-2 w-2 bg-slate-400' })
          ]),
          h('span', { class: 'text-sm font-medium text-slate-500' }, timeAgoStr)
        ])
      }
    },
  },
]
</script>

<template>
  <div class="space-y-6">
    <div class="flex justify-between items-center">
      <div>
        <h2 class="text-2xl font-bold tracking-tight text-slate-900">Pengguna Aktif</h2>
        <p class="text-sm text-slate-500 mt-1">Daftar pengguna yang sedang aktif menggunakan aplikasi Super App (Real-time).</p>
      </div>
    </div>

    <div class="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
      <DataTable
        :data="activeUsers"
        :columns="columns"
        :isLoading="isLoading"
        searchKey="name"
        searchPlaceholder="Cari nama pengguna..."
      />
    </div>
  </div>
</template>
