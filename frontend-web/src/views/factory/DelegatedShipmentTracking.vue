<template>
  <div class="p-8 space-y-8 bg-[#f8fafc] min-h-screen">
    <!-- Header -->
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 bg-white p-8 rounded-[2rem] shadow-sm border border-slate-100">
      <div>
        <div class="flex items-center gap-2 text-xs font-bold text-orange-600 uppercase tracking-[0.2em] mb-2">
          <Truck class="h-4 w-4" />
          <span>Tracking Pengiriman Titipan</span>
        </div>
        <h1 class="text-3xl font-black tracking-tight text-slate-900">Lacak Delegasi Cabang</h1>
        <p class="text-slate-500 mt-1">
          Pantau status Surat Jalan yang ditugaskan ke cabang dan kelola hutang stok yang harus dikembalikan.
        </p>
      </div>
      <div class="flex gap-3">
        <Button variant="outline" class="rounded-2xl border-slate-200 bg-white hover:bg-slate-50 text-slate-600 font-bold" @click="fetchData">
          <RefreshCw class="mr-2 h-4 w-4" :class="{ 'animate-spin': factoryStore.loading }" />
          Segarkan
        </Button>
      </div>
    </div>

    <!-- Stats / Stock Debt Summary -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
      <Card class="rounded-[2rem] shadow-xl shadow-slate-200/50 border-none overflow-hidden">
        <CardHeader class="bg-gradient-to-br from-orange-50 to-red-50 p-6 border-b border-orange-100/50">
          <CardTitle class="text-xl font-black text-orange-900 flex items-center gap-2">
            <AlertTriangle class="h-5 w-5 text-orange-600" />
            Hutang Stok ke Cabang (Belum Diganti)
          </CardTitle>
          <CardDescription class="text-orange-700/80">Barang yang dipinjam dari cabang saat pengiriman titipan.</CardDescription>
        </CardHeader>
        <CardContent class="p-0">
          <div v-if="factoryStore.delegatedSummary.length === 0" class="p-8 text-center text-slate-500">
            Tidak ada hutang stok ke cabang.
          </div>
          <div v-else class="divide-y divide-slate-100">
            <div v-for="item in factoryStore.delegatedSummary" :key="item.product_id" class="p-4 flex items-center justify-between hover:bg-slate-50 transition-colors">
              <div class="flex items-center gap-3">
                <div class="h-10 w-10 rounded-xl bg-orange-100 flex items-center justify-center flex-shrink-0">
                  <Package class="h-5 w-5 text-orange-600" />
                </div>
                <div>
                  <p class="font-bold text-slate-900 text-sm">{{ item.product_name }}</p>
                  <p class="text-xs text-slate-500">Menunggu Replenishment</p>
                </div>
              </div>
              <div class="text-right">
                <span class="text-lg font-black text-orange-600">{{ item.total_pieces }}</span>
                <span class="text-xs font-bold text-slate-400 ml-1">PCS</span>
              </div>
            </div>
          </div>
        </CardContent>
        <CardFooter v-if="factoryStore.delegatedSummary.length > 0" class="p-4 bg-slate-50 border-t border-slate-100">
          <Button class="w-full rounded-xl bg-orange-600 hover:bg-orange-700 text-white shadow-lg shadow-orange-600/20">
            Proses Pengembalian (Replenish)
          </Button>
        </CardFooter>
      </Card>

      <!-- Recent Delegated Shipments -->
      <Card class="rounded-[2rem] shadow-xl shadow-slate-200/50 border-none overflow-hidden">
        <CardHeader class="bg-white p-6 border-b border-slate-100">
          <CardTitle class="text-xl font-black text-slate-900">Riwayat Pengiriman Delegasi</CardTitle>
          <CardDescription>Daftar Surat Jalan yang ditugaskan ke cabang.</CardDescription>
        </CardHeader>
        <CardContent class="p-0 max-h-[400px] overflow-y-auto">
          <div v-if="factoryStore.delegatedShipments.length === 0" class="p-8 text-center text-slate-500">
            Belum ada data pengiriman delegasi.
          </div>
          <div v-else class="divide-y divide-slate-100">
            <div v-for="shipment in factoryStore.delegatedShipments" :key="shipment.id" class="p-4 hover:bg-slate-50 transition-colors">
              <div class="flex justify-between items-start mb-2">
                <div>
                  <p class="font-black text-slate-900 text-sm font-mono">{{ shipment.delivery_order_no || shipment.id.substring(0, 8) }}</p>
                  <p class="text-xs text-slate-500 flex items-center gap-1 mt-1">
                    <MapPin class="h-3 w-3" /> Tujuan: {{ shipment.to_branch?.name }}
                  </p>
                </div>
                <Badge :variant="shipment.status === 'SHIPPED' || shipment.status === 'ARRIVED' ? 'default' : 'secondary'" 
                  :class="[
                    'rounded-full text-[10px]',
                    (shipment.status === 'SHIPPED' || shipment.status === 'ARRIVED') ? 'bg-green-100 text-green-700' : '',
                    (shipment.status === 'APPROVED' || shipment.status === 'ASSIGNED_TO_BRANCH') ? 'bg-amber-100 text-amber-700' : ''
                  ].join(' ')">
                  {{ shipment.status }}
                </Badge>
              </div>
              <div class="mt-3 flex items-center gap-2 text-xs font-medium text-slate-600 bg-slate-100/50 p-2 rounded-lg">
                <Truck class="h-3.5 w-3.5 text-slate-400" />
                Didelegasikan ke: Cabang
                <Badge v-if="shipment.is_stock_borrowed" variant="outline" class="ml-auto text-[9px] border-orange-200 text-orange-600 bg-orange-50">Stok Dipinjam</Badge>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useFactoryStore } from '@/stores/factory'
import { useAuthStore } from '@/stores/auth'
import { Button } from '@/components/ui/button'
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Truck, RefreshCw, AlertTriangle, Package, MapPin } from 'lucide-vue-next'

const factoryStore = useFactoryStore()
const authStore = useAuthStore()

const fetchData = async () => {
  if (authStore.user?.company_id) {
    await factoryStore.fetchDelegatedShipments()
  }
}

onMounted(() => {
  fetchData()
})
</script>
