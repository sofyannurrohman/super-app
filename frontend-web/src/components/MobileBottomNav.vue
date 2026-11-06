<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useLayoutStore } from '@/stores/layout'
import { useAuthStore } from '@/stores/auth'
import {
  LayoutDashboard,
  Users,
  Clock,
  Wallet,
  ShoppingCart,
  Truck,
  Warehouse,
  Factory,
  Banknote,
  Menu,
  MapPin,
  FileText,
  Target,
  History,
  ClipboardList,
  RotateCcw,
  Receipt,
} from 'lucide-vue-next'

const layoutStore = useLayoutStore()
const authStore = useAuthStore()
const router = useRouter()
const route = useRoute()

// Bottom nav items per module — max 4 + "More" button
const hrisNav = [
  { to: '/', name: 'Dashboard', icon: LayoutDashboard },
  { to: '/employees', name: 'Karyawan', icon: Users },
  { to: '/attendance', name: 'Kehadiran', icon: Clock },
  { to: '/payroll', name: 'Penggajian', icon: Wallet },
]

const salesNav = [
  { to: '/sales', name: 'Dashboard', icon: LayoutDashboard },
  { to: '/sales/kunjungan', name: 'Kunjungan', icon: MapPin },
  { to: '/sales/pesanan', name: 'Pesanan', icon: ShoppingCart },
  { to: '/sales/delivery', name: 'Pengiriman', icon: Truck },
]

const factoryNav = [
  { to: '/factory', name: 'Dashboard', icon: LayoutDashboard },
  { to: '/factory/stock', name: 'Stok', icon: Factory },
  { to: '/factory/transfers', name: 'Pengiriman', icon: Truck },
  { to: '/factory/raw-materials', name: 'Bahan Baku', icon: ClipboardList },
]

const warehouseNav = [
  { to: '/warehouse', name: 'Dashboard', icon: LayoutDashboard },
  { to: '/warehouse/stock', name: 'Persediaan', icon: Warehouse },
  { to: '/warehouse/incoming-orders', name: 'Pesanan', icon: ClipboardList },
  { to: '/warehouse/logs', name: 'Log', icon: History },
]

const financeNav = [
  { to: '/finance', name: 'Dashboard', icon: LayoutDashboard },
  { to: '/reimbursements', name: 'Reimburse', icon: Receipt },
  { to: '/finance/approval', name: 'Persetujuan', icon: ClipboardList },
  { to: '/finance/profit-loss', name: 'Laporan', icon: Banknote },
]

const employeeNav = [{ to: '/kehadiran', name: 'Absensi', icon: Clock }]

const bottomNavItems = computed(() => {
  if (authStore.user?.role === 'employee') {
    return employeeNav
  }
  switch (layoutStore.currentModule) {
    case 'sales':
      return salesNav
    case 'factory':
      return factoryNav
    case 'warehouse':
      return warehouseNav
    case 'finance':
      return financeNav
    default:
      return hrisNav
  }
})

const isActive = (to: string) => {
  if (
    to === '/' ||
    to === '/sales' ||
    to === '/factory' ||
    to === '/warehouse' ||
    to === '/finance'
  ) {
    return route.path === to
  }
  return route.path.startsWith(to)
}
</script>

<template>
  <nav
    class="fixed bottom-0 left-0 right-0 z-40 bg-white/95 backdrop-blur-xl border-t border-slate-200/60 safe-area-bottom shadow-[0_-4px_24px_rgba(0,0,0,0.06)]"
  >
    <div class="flex items-stretch h-[64px]">
      <!-- Nav Items -->
      <button
        v-for="item in bottomNavItems"
        :key="item.to"
        @click="router.push(item.to)"
        class="flex-1 flex flex-col items-center justify-center gap-1 relative transition-all duration-200 active:scale-95"
        :class="isActive(item.to) ? 'text-primary' : 'text-slate-400'"
      >
        <!-- Active indicator pill -->
        <div
          v-if="isActive(item.to)"
          class="absolute top-0 left-1/2 -translate-x-1/2 w-8 h-[3px] bg-primary rounded-b-full"
        />
        <!-- Active background blob -->
        <div
          v-if="isActive(item.to)"
          class="absolute inset-x-2 inset-y-1 rounded-xl bg-primary/5"
        />
        <component
          :is="item.icon"
          class="w-5 h-5 relative z-10 transition-transform duration-200"
          :class="isActive(item.to) ? 'scale-110' : 'scale-100'"
        />
        <span
          class="text-[10px] font-bold relative z-10 leading-none"
          :class="isActive(item.to) ? 'text-primary' : 'text-slate-400'"
        >
          {{ item.name }}
        </span>
      </button>

      <!-- More / Menu button -->
      <button
        @click="layoutStore.openMobileDrawer"
        class="flex-1 flex flex-col items-center justify-center gap-1 relative transition-all duration-200 active:scale-95"
        :class="layoutStore.isMobileDrawerOpen ? 'text-primary' : 'text-slate-400'"
      >
        <div
          v-if="layoutStore.isMobileDrawerOpen"
          class="absolute inset-x-2 inset-y-1 rounded-xl bg-primary/5"
        />
        <Menu class="w-5 h-5 relative z-10" />
        <span class="text-[10px] font-bold relative z-10 leading-none">Menu</span>
      </button>
    </div>
  </nav>
</template>

<style scoped>
.safe-area-bottom {
  padding-bottom: env(safe-area-inset-bottom, 0px);
  height: calc(64px + env(safe-area-inset-bottom, 0px));
}
</style>
