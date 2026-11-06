<template>
  <div class="p-6 space-y-6 bg-[#f8fafc] min-h-screen">
    <!-- Header -->
    <div
      class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 bg-white p-6 rounded-3xl shadow-sm border border-slate-100"
    >
      <div>
        <h1 class="text-2xl font-black tracking-tight text-slate-900">
          Stok & Mutasi Bahan Baku
        </h1>
        <p class="text-sm text-slate-500 mt-0.5">
          Kelola saldo stok fisik bahan baku, stock opname / saldo awal, dan riwayat mutasi gudang pabrik.
        </p>
      </div>
      <div class="flex flex-wrap items-center gap-3">
        <select
          v-model="selectedFactoryId"
          class="text-sm bg-slate-50 border border-slate-200 rounded-xl px-3 py-2 font-medium focus:outline-none focus:ring-2 focus:ring-primary/20"
        >
          <option v-for="f in factoryStore.factories" :key="f.id" :value="f.id">{{ f.name }}</option>
        </select>
        <Button variant="outline" class="rounded-xl border-slate-200" @click="refreshData">
          <RefreshCw class="mr-2 h-4 w-4" :class="{ 'animate-spin': factoryStore.loading }" />
          Segarkan
        </Button>
        <Button
          @click="openCreateMaster"
          variant="outline"
          class="rounded-xl font-bold bg-white"
        >
          <Plus class="mr-2 h-4 w-4" /> Master Bahan Baku
        </Button>
        <Button
          @click="openStockOpnameGlobal"
          :disabled="!selectedFactoryId"
          class="rounded-xl font-bold shadow-md shadow-amber-500/10 bg-amber-600 hover:bg-amber-700 text-white"
        >
          <ClipboardCheck class="mr-2 h-4 w-4" /> Stock Opname / Saldo Awal
        </Button>
        <Button
          @click="openMutation"
          :disabled="!selectedFactoryId"
          variant="secondary"
          class="rounded-xl font-bold text-slate-700 bg-slate-100 hover:bg-slate-200"
        >
          <Plus class="mr-2 h-4 w-4" /> Mutasi Manual
        </Button>
      </div>
    </div>

    <!-- Summary KPIs Bar -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <Card class="p-4 bg-white border border-slate-100 rounded-2xl shadow-xs flex items-center gap-3">
        <div class="h-10 w-10 rounded-xl bg-amber-500/10 flex items-center justify-center text-amber-600">
          <Boxes class="h-5 w-5" />
        </div>
        <div>
          <div class="text-[10px] font-bold uppercase text-slate-400">Total Varian Master</div>
          <div class="text-lg font-black text-slate-900">{{ rawMaterialMasterList.length }} Item</div>
        </div>
      </Card>
      <Card class="p-4 bg-white border border-slate-100 rounded-2xl shadow-xs flex items-center gap-3">
        <div class="h-10 w-10 rounded-xl bg-emerald-500/10 flex items-center justify-center text-emerald-600">
          <Scale class="h-5 w-5" />
        </div>
        <div>
          <div class="text-[10px] font-bold uppercase text-slate-400">Total Stok Bobot (KG)</div>
          <div class="text-lg font-black text-slate-900 font-mono">{{ stockStats.totalKg.toLocaleString('id-ID', { maximumFractionDigits: 1 }) }} KG</div>
        </div>
      </Card>
      <Card class="p-4 bg-white border border-slate-100 rounded-2xl shadow-xs flex items-center gap-3">
        <div class="h-10 w-10 rounded-xl bg-blue-500/10 flex items-center justify-center text-blue-600">
          <Droplets class="h-5 w-5" />
        </div>
        <div>
          <div class="text-[10px] font-bold uppercase text-slate-400">Total Stok Cairan</div>
          <div class="text-lg font-black text-slate-900 font-mono">{{ stockStats.totalLiter.toLocaleString('id-ID', { maximumFractionDigits: 1 }) }} Liter</div>
        </div>
      </Card>
      <Card class="p-4 bg-white border border-slate-100 rounded-2xl shadow-xs flex items-center gap-3">
        <div class="h-10 w-10 rounded-xl bg-purple-500/10 flex items-center justify-center text-purple-600">
          <PackageOpen class="h-5 w-5" />
        </div>
        <div>
          <div class="text-[10px] font-bold uppercase text-slate-400">Total Kemasan & Wadah</div>
          <div class="text-lg font-black text-slate-900 font-mono">{{ stockStats.totalPackagingPcs.toLocaleString('id-ID') }} Pcs</div>
        </div>
      </Card>
    </div>

    <!-- Tabs & Mode Switcher -->
    <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
      <Tabs v-model="activeTab" class="w-full sm:w-auto">
        <TabsList class="bg-white border border-slate-100 p-1 rounded-2xl shadow-sm inline-flex h-12 items-center justify-center text-slate-500 w-full sm:w-auto overflow-x-auto">
          <TabsTrigger
            value="stock"
            class="rounded-xl px-5 py-2.5 text-xs sm:text-sm font-semibold transition-all data-[state=active]:bg-primary data-[state=active]:text-white data-[state=active]:shadow-md"
          >
            Stok Bahan Baku
          </TabsTrigger>
          <TabsTrigger
            value="logs"
            class="rounded-xl px-5 py-2.5 text-xs sm:text-sm font-semibold transition-all data-[state=active]:bg-primary data-[state=active]:text-white data-[state=active]:shadow-md"
          >
            Riwayat Mutasi
          </TabsTrigger>
          <TabsTrigger
            value="master"
            class="rounded-xl px-5 py-2.5 text-xs sm:text-sm font-semibold transition-all data-[state=active]:bg-primary data-[state=active]:text-white data-[state=active]:shadow-md"
          >
            Daftar Master Bahan Baku
          </TabsTrigger>
        </TabsList>
      </Tabs>

      <!-- Unit Display Switcher (When on stock tab) -->
      <div v-if="activeTab === 'stock'" class="flex items-center gap-1.5 bg-white p-1 rounded-2xl border border-slate-100 shadow-sm text-xs">
        <span class="text-[10px] font-bold uppercase text-slate-400 px-2">Format Satuan:</span>
        <button
          type="button"
          @click="unitDisplayMode = 'highest'"
          :class="[
            'px-3 py-1.5 rounded-xl font-bold transition-all text-xs',
            unitDisplayMode === 'highest'
              ? 'bg-slate-900 text-white shadow-xs'
              : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100'
          ]"
        >
          ⭐ Satuan Tertinggi (KG / L / DUS)
        </button>
        <button
          type="button"
          @click="unitDisplayMode = 'breakdown'"
          :class="[
            'px-3 py-1.5 rounded-xl font-bold transition-all text-xs',
            unitDisplayMode === 'breakdown'
              ? 'bg-slate-900 text-white shadow-xs'
              : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100'
          ]"
        >
          Pecahan (12 kg 500 gr)
        </button>
        <button
          type="button"
          @click="unitDisplayMode = 'base'"
          :class="[
            'px-3 py-1.5 rounded-xl font-bold transition-all text-xs',
            unitDisplayMode === 'base'
              ? 'bg-slate-900 text-white shadow-xs'
              : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100'
          ]"
        >
          Satuan Asli
        </button>
      </div>
    </div>

    <!-- TAB 1: STOK BAHAN BAKU -->
    <div v-if="activeTab === 'stock'" class="mt-2">
      <Card class="border-none shadow-sm bg-white rounded-3xl overflow-hidden">
        <DataTable
          :columns="stockColumns"
          :data="factoryStore.rawMaterialStock"
          :isLoading="factoryStore.loading"
        >
          <template #cell-product="{ row }">
            <div class="flex items-center gap-3 py-1">
              <div class="h-10 w-10 rounded-xl bg-primary/10 flex items-center justify-center text-primary border border-primary/20">
                <Package class="h-5 w-5" />
              </div>
              <div class="flex flex-col">
                <span class="font-bold text-slate-900 text-sm">{{ (row as any).product?.name }}</span>
                <span class="text-xs text-slate-500 font-medium">{{ (row as any).product?.sku }} • {{ (row as any).product?.category || 'BAHAN BAKU' }}</span>
              </div>
            </div>
          </template>
          <template #cell-factory_warehouse="{ row }">
            <span class="text-sm font-semibold text-slate-700">
              {{ (row as any).factory_warehouse?.name || 'Gudang Utama' }}
            </span>
          </template>
          <template #cell-quantity="{ row }">
            <div class="flex flex-col items-start gap-0.5 py-1">
              <!-- Mode: Highest (Default) -->
              <template v-if="unitDisplayMode === 'highest'">
                <Badge variant="secondary" class="font-mono font-black text-xs bg-slate-900 text-white px-2.5 py-1 rounded-lg">
                  {{ formatMultiUnit((row as any).quantity, (row as any).product?.unit, (row as any).product?.pcs_per_unit).formattedHighest }}
                </Badge>
                <span
                  v-if="formatMultiUnit((row as any).quantity, (row as any).product?.unit, (row as any).product?.pcs_per_unit).category !== 'other'"
                  class="text-[11px] text-slate-500 font-mono font-medium"
                >
                  ({{ formatMultiUnit((row as any).quantity, (row as any).product?.unit, (row as any).product?.pcs_per_unit).baseFormatted }})
                </span>
              </template>

              <!-- Mode: Breakdown -->
              <template v-else-if="unitDisplayMode === 'breakdown'">
                <Badge variant="secondary" class="font-mono font-black text-xs bg-amber-100 text-amber-950 border border-amber-200 px-2.5 py-1 rounded-lg">
                  {{ formatMultiUnit((row as any).quantity, (row as any).product?.unit, (row as any).product?.pcs_per_unit).breakdown }}
                </Badge>
                <span class="text-[11px] text-slate-500 font-mono font-medium">
                  ({{ formatMultiUnit((row as any).quantity, (row as any).product?.unit, (row as any).product?.pcs_per_unit).baseFormatted }})
                </span>
              </template>

              <!-- Mode: Base -->
              <template v-else>
                <Badge variant="secondary" class="font-mono font-bold text-xs bg-slate-100 text-slate-800 px-2.5 py-1 rounded-lg">
                  {{ (row as any).quantity.toLocaleString('id-ID') }} {{ (row as any).product?.unit || 'PCS' }}
                </Badge>
              </template>
            </div>
          </template>
          <template #cell-actions="{ row }">
            <div class="flex items-center justify-end gap-1">
              <Button
                variant="ghost"
                size="icon"
                @click="openStockCardRow(row)"
                title="Lihat Riwayat Mutasi / Kartu Stok"
                class="h-8 w-8 text-blue-600 hover:text-blue-700 hover:bg-blue-50 rounded-lg"
              >
                <History class="h-4 w-4" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                @click="openStockOpnameRow(row)"
                title="Koreksi / Opname Stok"
                class="h-8 w-8 text-amber-600 hover:text-amber-700 hover:bg-amber-50 rounded-lg"
              >
                <ClipboardCheck class="h-4 w-4" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                @click="handleDeleteStock((row as any).id)"
                title="Hapus Saldo Stok"
                class="h-8 w-8 text-slate-400 hover:text-destructive hover:bg-destructive/5 rounded-lg"
              >
                <Trash2 class="h-3.5 w-3.5" />
              </Button>
            </div>
          </template>
        </DataTable>
      </Card>
    </div>

    <!-- TAB 2: RIWAYAT MUTASI -->
    <div v-if="activeTab === 'logs'" class="mt-2 space-y-3">
      <!-- Filter Bar for Logs -->
      <div class="flex flex-col sm:flex-row items-center justify-between gap-3 bg-white p-4 rounded-3xl border border-slate-100 shadow-2xs">
        <div class="flex flex-wrap items-center gap-2.5 w-full sm:w-auto">
          <div class="flex items-center gap-1.5 text-xs font-bold text-slate-600">
            <Filter class="w-3.5 h-3.5 text-amber-600" />
            <span>Filter Bahan Baku:</span>
          </div>
          <select
            v-model="selectedLogProductId"
            @change="loadData"
            class="text-xs font-semibold bg-slate-50 border border-slate-200 rounded-xl px-3 py-1.5 text-slate-700 focus:outline-none focus:ring-2 focus:ring-amber-500/20 max-w-[260px] truncate"
          >
            <option value="">Semua Bahan Baku & Kemasan</option>
            <option v-for="p in rawMaterialMasterList" :key="p.id" :value="p.id">
              {{ p.name }} ({{ p.sku }})
            </option>
          </select>
          <Button
            v-if="selectedLogProductId"
            variant="ghost"
            size="sm"
            @click="selectedLogProductId = ''; loadData()"
            class="h-7 px-2 text-[11px] font-bold text-slate-500 hover:text-red-600 rounded-lg"
          >
            Reset Filter
          </Button>
        </div>
        <div class="text-xs text-slate-400 font-medium">
          Ditemukan: <strong class="text-slate-800">{{ factoryStore.rawMaterialLogs.length }}</strong> Catatan Mutasi
        </div>
      </div>

      <Card class="border-none shadow-sm bg-white rounded-3xl overflow-hidden">
        <DataTable
          :columns="logColumns"
          :data="factoryStore.rawMaterialLogs"
          :isLoading="factoryStore.loading"
        >
          <template #cell-type="{ row }">
            <Badge
              :variant="
                (row as any).type === 'IN' ? 'default' :
                (row as any).type === 'OUT' ? 'destructive' : 'secondary'
              "
              class="rounded-lg text-[10px] py-0 px-2 font-bold uppercase"
            >
              {{ (row as any).type }}
            </Badge>
          </template>
          <template #cell-product="{ row }">
            <div class="flex flex-col">
              <span class="font-bold text-slate-900 text-sm">{{ (row as any).product?.name }}</span>
              <span class="text-xs text-slate-500 font-medium">{{ (row as any).product?.sku }}</span>
            </div>
          </template>
          <template #cell-quantity="{ row }">
            <div class="flex flex-col">
              <span class="font-mono font-bold text-xs" :class="
                (row as any).type === 'IN' ? 'text-green-600' :
                (row as any).type === 'OUT' ? 'text-red-600' : 'text-slate-700'
              ">
                {{ (row as any).type === 'IN' ? '+' : (row as any).type === 'OUT' ? '-' : '' }}{{ formatMultiUnit(Math.abs((row as any).quantity), (row as any).product?.unit, (row as any).product?.pcs_per_unit).formattedHighest }}
              </span>
              <span class="text-[10px] text-slate-400 font-mono font-medium">
                {{ (row as any).type === 'IN' ? '+' : (row as any).type === 'OUT' ? '-' : '' }}{{ Math.abs((row as any).quantity).toLocaleString('id-ID') }} {{ (row as any).product?.unit || 'PCS' }}
              </span>
            </div>
          </template>
          <template #cell-created_at="{ row }">
            <span class="text-xs font-medium text-slate-600">
              {{ new Date((row as any).created_at).toLocaleString('id-ID') }}
            </span>
          </template>
          <template #cell-actions="{ row }">
            <div class="flex items-center justify-end">
              <Button
                variant="ghost"
                size="icon"
                @click="handleDeleteLog((row as any).id)"
                title="Hapus Catatan Log"
                class="h-8 w-8 text-slate-400 hover:text-destructive hover:bg-destructive/5 rounded-lg"
              >
                <Trash2 class="h-3.5 w-3.5" />
              </Button>
            </div>
          </template>
        </DataTable>
      </Card>
    </div>

    <!-- TAB 3: DAFTAR MASTER BAHAN BAKU -->
    <div v-if="activeTab === 'master'" class="mt-2">
      <Card class="border-none shadow-sm bg-white rounded-3xl overflow-hidden">
        <DataTable
          :columns="masterColumns"
          :data="rawMaterialMasterList"
          :isLoading="factoryStore.loading"
        >
          <template #cell-sku="{ row }">
            <span class="font-mono text-[11px] font-bold text-primary bg-primary/5 px-2 py-1 rounded-lg ring-1 ring-primary/10">
              {{ (row as any).sku }}
            </span>
          </template>
          <template #cell-name="{ row }">
            <div class="flex items-center gap-3 py-1">
              <div class="h-9 w-9 rounded-xl bg-amber-500/10 flex items-center justify-center text-amber-600 border border-amber-500/20">
                <Package class="h-4 w-4" />
              </div>
              <div class="flex flex-col">
                <span class="font-bold text-slate-900 text-sm">{{ (row as any).name }}</span>
                <span class="text-xs text-slate-400 font-medium">{{ (row as any).brand || 'Tanpa Brand' }}</span>
              </div>
            </div>
          </template>
          <template #cell-category="{ row }">
            <Badge variant="secondary" class="font-semibold text-xs bg-slate-100 text-slate-700">
              {{ (row as any).category || 'BAHAN BAKU' }}
            </Badge>
          </template>
          <template #cell-unit="{ row }">
            <span class="font-bold text-xs text-slate-700 uppercase">
              {{ (row as any).unit || 'KG' }}
            </span>
          </template>
          <template #cell-actions="{ row }">
            <div class="flex items-center justify-end gap-1">
              <Button
                variant="ghost"
                size="icon"
                @click="openEditMaster(row)"
                title="Edit Master Bahan Baku"
                class="h-8 w-8 text-slate-400 hover:text-primary hover:bg-primary/5 rounded-lg"
              >
                <Edit2 class="h-3.5 w-3.5" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                @click="handleDeleteMaster((row as any).id)"
                title="Hapus Master Bahan Baku"
                class="h-8 w-8 text-slate-400 hover:text-destructive hover:bg-destructive/5 rounded-lg"
              >
                <Trash2 class="h-3.5 w-3.5" />
              </Button>
            </div>
          </template>
        </DataTable>
      </Card>
    </div>

    <!-- Dialogs -->
    <RawMaterialStockCardDialog
      v-model:open="showStockCardDialog"
      :factoryId="selectedFactoryId"
      :stockItem="selectedStockCardItem"
    />

    <RawMaterialOpnameDialog
      v-model:open="showOpnameDialog"
      :factoryId="selectedFactoryId"
      :stockItem="selectedStockItem"
      @success="refreshData"
    />

    <RawMaterialMutationDialog
      v-model:open="showMutationDialog"
      :factoryId="selectedFactoryId"
      :defaultType="mutationDefaultType"
      @success="refreshData"
    />

    <RawMaterialMasterDialog
      v-model:open="showMasterDialog"
      :product="selectedMasterProduct"
      @success="refreshData"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useFactoryStore } from '@/stores/factory'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import DataTable from '@/components/DataTable.vue'
import {
  RefreshCw,
  Plus,
  Package,
  ClipboardCheck,
  Edit2,
  Trash2,
  Boxes,
  Scale,
  Droplets,
  PackageOpen,
  History,
  Filter,
} from 'lucide-vue-next'
import RawMaterialOpnameDialog from '@/components/factory/RawMaterialOpnameDialog.vue'
import RawMaterialMutationDialog from '@/components/factory/RawMaterialMutationDialog.vue'
import RawMaterialMasterDialog from '@/components/factory/RawMaterialMasterDialog.vue'
import RawMaterialStockCardDialog from '@/components/factory/RawMaterialStockCardDialog.vue'
import { groupProductsByCategory, formatMultiUnit } from '@/utils/productUnits'
import { toast } from 'vue-sonner'

const factoryStore = useFactoryStore()
const activeTab = ref('stock')
const selectedFactoryId = ref('')
const selectedLogProductId = ref('')
const unitDisplayMode = ref<'highest' | 'breakdown' | 'base'>('highest')

const stockStats = computed(() => {
  let totalKg = 0
  let totalLiter = 0
  let totalPackagingPcs = 0

  for (const s of factoryStore.rawMaterialStock) {
    const formatted = formatMultiUnit(s.quantity, s.product?.unit, s.product?.pcs_per_unit)
    if (formatted.category === 'weight') {
      totalKg += formatted.highestValue
    } else if (formatted.category === 'volume') {
      totalLiter += formatted.highestValue
    } else if (formatted.category === 'count') {
      totalPackagingPcs += s.quantity
    }
  }

  return {
    totalKg,
    totalLiter,
    totalPackagingPcs,
  }
})

const showStockCardDialog = ref(false)
const selectedStockCardItem = ref<any>(null)

const showOpnameDialog = ref(false)
const selectedStockItem = ref<any>(null)

const showMutationDialog = ref(false)
const mutationDefaultType = ref('IN')
const showMasterDialog = ref(false)
const selectedMasterProduct = ref<any>(null)

const openStockCardRow = (item: any) => {
  selectedStockCardItem.value = item
  showStockCardDialog.value = true
}

const openStockOpnameGlobal = () => {
  selectedStockItem.value = null
  showOpnameDialog.value = true
}

const openStockOpnameRow = (item: any) => {
  selectedStockItem.value = item
  showOpnameDialog.value = true
}

const openMutation = () => {
  mutationDefaultType.value = 'IN'
  showMutationDialog.value = true
}

const openCreateMaster = () => {
  selectedMasterProduct.value = null
  showMasterDialog.value = true
}

const openEditMaster = (item: any) => {
  selectedMasterProduct.value = item
  showMasterDialog.value = true
}

const stockColumns = [
  {
    id: 'product',
    header: 'Produk Bahan Baku',
    accessorFn: (row: any) => `${row.product?.name || ''} ${row.product?.sku || ''} ${row.product?.category || ''}`,
    sortable: true,
  },
  {
    id: 'factory_warehouse',
    header: 'Gudang',
    accessorFn: (row: any) => row.factory_warehouse?.name || 'Gudang Utama',
    sortable: true,
  },
  { key: 'batch_no', label: 'Batch', sortable: true },
  { key: 'quantity', label: 'Stok Tersedia', sortable: true },
  { key: 'actions', label: '', sortable: false },
]

const logColumns = [
  { key: 'created_at', label: 'Waktu Mutasi', sortable: true },
  { key: 'type', label: 'Tipe', sortable: true },
  {
    id: 'product',
    header: 'Produk',
    accessorFn: (row: any) => `${row.product?.name || ''} ${row.product?.sku || ''}`,
    sortable: true,
  },
  { key: 'batch_no', label: 'Batch', sortable: true },
  { key: 'quantity', label: 'Qty', sortable: true },
  { key: 'source', label: 'Sumber', sortable: true },
  { key: 'actions', label: '', sortable: false },
]

const masterColumns = [
  { key: 'sku', label: 'Kode SKU', sortable: true },
  { key: 'name', label: 'Nama Bahan Baku / Barang', sortable: true },
  { key: 'category', label: 'Kategori Pengadaan', sortable: true },
  { key: 'unit', label: 'Satuan (UoM)', sortable: true },
  { key: 'actions', label: '', sortable: false },
]

const rawMaterialMasterList = computed(() => {
  const grouped = groupProductsByCategory(factoryStore.products)
  return [
    ...grouped.rawMaterials,
    ...grouped.packagingMaterials,
    ...grouped.tradingProducts,
    ...grouped.giftProducts,
  ]
})

const handleDeleteStock = async (id: string) => {
  if (!confirm('Hapus baris saldo stok ini?')) {
    return
  }
  try {
    await factoryStore.deleteStock(selectedFactoryId.value, id)
    toast.success('Baris saldo stok berhasil dihapus')
  } catch (err: any) {
    toast.error('Gagal menghapus baris stok: ' + (err.message || String(err)))
  }
}

const handleDeleteLog = async (id: string) => {
  if (!confirm('Hapus catatan mutasi ini?')) {
    return
  }
  try {
    await factoryStore.deleteRawMaterialLog(selectedFactoryId.value, id)
    toast.success('Catatan riwayat mutasi berhasil dihapus')
  } catch (err: any) {
    toast.error('Gagal menghapus log mutasi: ' + (err.message || String(err)))
  }
}

const handleDeleteMaster = async (id: string) => {
  if (!confirm('Hapus master bahan baku ini? Pastikan tidak ada stok atau riwayat transaksi aktif yang terkait.')) {
    return
  }
  try {
    await factoryStore.deleteProduct(id)
    toast.success('Master bahan baku berhasil dihapus')
    await factoryStore.fetchProducts()
  } catch (err: any) {
    toast.error('Gagal menghapus master bahan baku: ' + (err.message || String(err)))
  }
}

async function loadData() {
  if (!selectedFactoryId.value) return
  if (activeTab.value === 'stock') {
    await factoryStore.fetchRawMaterialStock(selectedFactoryId.value)
  } else if (activeTab.value === 'logs') {
    await factoryStore.fetchRawMaterialLogs(selectedFactoryId.value, {
      product_id: selectedLogProductId.value || undefined,
    })
  } else if (activeTab.value === 'master') {
    await factoryStore.fetchProducts()
  }
}

async function refreshData() {
  await loadData()
}

watch(selectedFactoryId, () => {
  loadData()
})

watch(activeTab, () => {
  loadData()
})

onMounted(async () => {
  if (factoryStore.factories.length === 0) {
    await factoryStore.fetchFactories()
  }
  if (factoryStore.products.length === 0) {
    await factoryStore.fetchProducts()
  }
  if (factoryStore.factories.length > 0) {
    selectedFactoryId.value = factoryStore.factories[0]?.id || ''
  }
})
</script>
