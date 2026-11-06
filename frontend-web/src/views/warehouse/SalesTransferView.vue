<template>
  <div class="p-6 space-y-6">
    <div
      class="flex justify-between items-center bg-white text-slate-900 p-8 rounded-2xl shadow-lg"
    >
      <div>
        <h1 class="text-2xl font-bold tracking-tight">
          Transfer & Retur Barang dari Gudang {{ currentBranchName }} ke Sales
        </h1>
        <p class="text-muted-foreground">
          Manajemen transfer barang dari gudang ke sales dan retur barang dari sales ke gudang.
        </p>
      </div>
      <Button @click="openForm" class="flex gap-2">
        <Plus class="h-5 w-5" />
        Buat Transfer
      </Button>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-12 gap-8">
      <!-- Left: History -->
      <div class="lg:col-span-8 space-y-4">
        <h2 class="text-xl font-bold flex items-center gap-2">
          <History class="h-5 w-5 text-primary" />
          Riwayat Perpindahan & Retur
        </h2>
        <Card class="border-none shadow-sm overflow-hidden">
          <DataTable :columns="columns" :data="filteredTransfers">
            <template #cell-type="{ row }">
              <span
                v-if="row.type === 'TRANSFER'"
                class="flex items-center gap-1 text-blue-600 font-bold"
              >
                <ArrowRight class="h-3 w-3" /> TRANSFER
              </span>
              <span
                v-else-if="row.type === 'BONUS_REQUEST'"
                class="flex items-center gap-1 text-amber-600 font-bold"
              >
                <Gift class="h-3 w-3" /> BONUS
              </span>
              <span v-else class="flex items-center gap-1 text-orange-600 font-bold">
                <ArrowLeft class="h-3 w-3" /> RETUR
              </span>
            </template>
            <template #cell-reference_no="{ row }">
              <span v-if="row.is_bonus_request" class="text-xs font-mono bg-amber-50 px-2 py-1 rounded text-amber-700">REQ-BONUS</span>
              <span v-else class="font-mono text-[11px] bg-slate-100 px-1.5 py-0.5 rounded border border-slate-200">
                {{ row.reference_no || '-' }}
              </span>
            </template>
            <template #cell-product_name="{ row }">
              <div v-if="row.product_id">
                <span class="font-medium">{{ row.product?.name || '-' }}</span>
                <div class="flex items-center gap-1 mt-0.5">
                  <span
                    v-if="row.batch_no && row.batch_no !== 'AUTO'"
                    class="text-[10px] font-mono bg-blue-50 text-blue-700 px-1.5 py-0.5 rounded border border-blue-200"
                  >
                    Batch: {{ row.batch_no }}
                  </span>
                  <span
                    v-else
                    class="text-[10px] font-mono bg-slate-100 text-slate-500 px-1.5 py-0.5 rounded"
                  >
                    FEFO
                  </span>
                </div>
              </div>
              <div v-else-if="row.bonus_inventory_id">
                <span class="font-medium text-amber-600 flex items-center gap-1">
                  {{ row.bonus_inventory?.name || '-' }} <span class="px-1.5 py-0.5 bg-amber-100 text-amber-700 rounded text-[10px] font-bold">BONUS</span>
                </span>
              </div>
              <div v-else>-</div>
            </template>
            <template #cell-employee_name="{ row }">
              {{ row.employee?.first_name }} {{ row.employee?.last_name }}
            </template>
            <template #cell-quantity="{ row }">
              <div>
                <strong>{{ row.quantity }} {{ row.product?.unit || row.bonus_inventory?.unit || 'pcs' }}</strong>
                <div v-if="row.status === 'COMPLETED' && row.product_id" class="text-[10px] text-slate-500 mt-0.5 space-y-0.5">
                  <div class="text-green-600 font-medium">Bagus: {{ row.actual_quantity || row.quantity }}</div>
                  <div v-if="row.damaged_quantity > 0" class="text-red-600 font-bold">Rusak: {{ row.damaged_quantity }}</div>
                </div>
              </div>
            </template>
            <template #cell-status="{ row }">
              <span
                :class="[
                  'px-2 py-1 rounded-full text-[10px] font-bold',
                  row.status === 'PENDING'
                    ? 'bg-slate-100 text-slate-600'
                    : row.status === 'COMPLETED'
                      ? 'bg-green-100 text-green-600'
                      : row.status === 'PROCESSING' || row.status === 'WAITING_SUPERVISOR'
                        ? 'bg-blue-100 text-blue-600'
                        : 'bg-red-100 text-red-600',
                ]"
              >
                {{ row.status }}
              </span>
            </template>
            <template #cell-actions="{ row }">
              <div class="flex items-center gap-1.5">
                <div v-if="row.is_bonus_request && row.status === 'PENDING'" class="flex gap-1">
                  <Button
                    size="sm"
                    class="h-7 px-2 text-[10px] bg-green-600 hover:bg-green-700 text-white"
                    @click="handleApproveBonus(row.id)"
                    >Setujui Bonus (Langsung)</Button
                  >
                  <Button
                    size="sm"
                    class="h-7 px-2 text-[10px] bg-indigo-600 hover:bg-indigo-700 text-white"
                    @click="handleGenerateDO(row)"
                    >Masukkan ke DO</Button
                  >
                  <Button
                    size="sm"
                    variant="outline"
                    class="h-7 px-2 text-[10px] text-red-600 border-red-200 hover:bg-red-50"
                    @click="handleRejectBonus(row.id)"
                    >Tolak</Button
                  >
                </div>
                <div v-else-if="row.status === 'PENDING' && !row.is_bonus_request" class="flex gap-1">
                  <Button
                    size="sm"
                    class="h-7 px-2 text-[10px] bg-green-600 hover:bg-green-700 text-white"
                    @click="handleComplete(row.id)"
                    >Langsung Selesai</Button
                  >
                  <Button
                    size="sm"
                    class="h-7 px-2 text-[10px] bg-indigo-600 hover:bg-indigo-700 text-white"
                    @click="handleGenerateDO(row)"
                    >{{ row.type === 'RETURN' ? 'Sisipkan Retur ke DO' : 'Masukkan ke DO' }}</Button
                  >
                  <Button
                    size="sm"
                    variant="outline"
                    class="h-7 px-2 text-[10px] text-red-600 border-red-200 hover:bg-red-50"
                    @click="handleCancel(row.id)"
                    >Batal</Button
                  >
                </div>
                
                <!-- Edit button for PENDING or COMPLETED -->
                <Button
                  v-if="(row.status === 'PENDING' || row.status === 'COMPLETED') && !row.is_bonus_request"
                  variant="outline"
                  size="icon"
                  class="h-7 w-7 text-amber-600 border-amber-200 hover:bg-amber-50"
                  @click="handleEdit(row)"
                >
                  <Edit class="h-3.5 w-3.5" />
                </Button>

                <Button
                  v-if="row.status !== 'PENDING'"
                  variant="ghost"
                  size="icon"
                  class="h-8 w-8 text-slate-400 hover:text-red-600"
                  @click="handleDelete(row.id)"
                >
                  <Trash2 class="h-4 w-4" />
                </Button>
              </div>
            </template>
          </DataTable>
        </Card>
      </div>

      <!-- Right: Sales Stock Checker -->
      <div class="lg:col-span-4 space-y-4">
        <h2 class="text-xl font-bold flex items-center gap-2">
          <UserSquare2 class="h-5 w-5 text-primary" />
          Cek Stok Salesman
        </h2>
        <Card>
          <CardHeader>
            <CardDescription>Pilih salesman untuk melihat stok yang mereka bawa.</CardDescription>
          </CardHeader>
          <CardContent class="space-y-4">
            <div class="space-y-2">
              <label class="text-sm font-medium">Salesman</label>
              <select
                v-model="selectedSalesman"
                class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                @change="fetchSalesStock"
              >
                <option value="">Pilih Salesman</option>
                <option v-for="emp in salesEmployees" :key="emp.id" :value="emp.id">
                  {{ emp.first_name }} {{ emp.last_name }}
                </option>
              </select>
            </div>

            <div v-if="activeSalesStock.length > 0" class="space-y-2 border-t pt-4">
              <div
                v-for="stock in activeSalesStock"
                :key="stock.id"
                class="flex justify-between items-center p-2 bg-slate-50 rounded-lg"
              >
                <div>
                  <div v-if="stock.product_id">
                    <p class="font-bold text-sm">{{ stock.product?.name }}</p>
                    <p class="text-xs text-slate-400">{{ stock.product?.sku }}</p>
                  </div>
                  <div v-else-if="stock.bonus_inventory_id">
                    <p class="font-bold text-sm text-amber-600 flex items-center gap-1">
                      {{ stock.bonus_inventory?.name || 'Bonus Inventory' }} <span class="px-1.5 py-0.5 bg-amber-100 text-amber-700 rounded text-[10px] font-bold">BONUS</span>
                    </p>
                  </div>
                </div>
                <div class="text-right">
                  <p class="font-black text-primary">
                    {{ stock.quantity }} {{ stock.product?.unit || stock.bonus_inventory?.unit || 'pcs' }}
                  </p>
                </div>
              </div>
            </div>
            <div
              v-else-if="selectedSalesman"
              class="text-center py-8 text-slate-400 text-sm italic"
            >
              Tidak ada stok yang dibawa.
            </div>
          </CardContent>
        </Card>
      </div>
    </div>

    <!-- Create Transfer/Return Modal -->
    <div
      v-if="showForm"
      class="fixed inset-0 bg-background/80 backdrop-blur-sm z-50 flex items-center justify-center p-4"
    >
      <Card class="w-full max-w-lg shadow-2xl border-2">
        <CardHeader>
          <CardTitle>Buat Perpindahan Barang</CardTitle>
          <CardDescription
            >Pindahkan stok dari gudang ke sales atau retur dari sales ke gudang.</CardDescription
          >
        </CardHeader>
        <CardContent class="space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <label class="text-sm font-medium">Tipe Perpindahan</label>
              <select
                v-model="form.type"
                class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              >
                <option value="TRANSFER">TRANSFER (Gudang -> Sales)</option>
                <option value="RETURN">RETUR (Sales -> Gudang)</option>
              </select>
            </div>
            <div class="space-y-2">
              <label class="text-sm font-medium">Salesman</label>
              <select
                v-model="form.employee_id"
                class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              >
                <option value="">Pilih Salesman</option>
                <option v-for="emp in salesEmployees" :key="emp.id" :value="emp.id">
                  {{ emp.first_name }} {{ emp.last_name }} ({{ emp.employee_id_number }})
                </option>
              </select>
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-sm font-medium">Barang</label>
            <select
              v-model="form.item_selection"
              class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
            >
              <option value="">Pilih Barang</option>
              <optgroup label="Produk">
                <option
                  v-for="product in factoryStore.sellableProducts"
                  :key="product.id"
                  :value="'PRODUCT:' + product.id"
                >
                  {{ product.name }} (Gudang:
                  {{ getProductStock(product.id) }}
                  {{ product.unit }})
                </option>
              </optgroup>
              <optgroup label="Bonus Inventory">
                <option
                  v-for="bonus in warehouseStore.bonusInventory"
                  :key="bonus.id"
                  :value="'BONUS:' + bonus.id"
                >
                  {{ bonus.name }} (Gudang: {{ bonus.quantity }} {{ bonus.unit }})
                </option>
              </optgroup>
            </select>
          </div>

          <div v-if="form.item_selection && form.item_selection.startsWith('PRODUCT:')" class="space-y-2">
            <div class="flex justify-between items-center">
              <label class="text-sm font-medium">Pilih Batch Stok</label>
              <span class="text-xs text-slate-500">
                Tersedia: <strong>{{ availableBatchesForSelectedProduct.length }}</strong> batch
              </span>
            </div>
            <select
              v-model="form.batch_no"
              class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
            >
              <option value="">Otomatis (FEFO - Kedaluwarsa Terdekat)</option>
              <option
                v-for="b in availableBatchesForSelectedProduct"
                :key="b.batch_no"
                :value="b.batch_no"
              >
                Batch: {{ b.batch_no || 'DEFAULT' }} (Stok: {{ b.quantity }} pcs{{ b.expiry_date ? ' | Exp: ' + new Date(b.expiry_date).toLocaleDateString('id-ID') : '' }})
              </option>
            </select>
            <p v-if="selectedBatchStockInfo" class="text-[11px] text-blue-600 font-medium">
              Sisa stok pada batch {{ selectedBatchStockInfo.batch_no }}: <strong>{{ selectedBatchStockInfo.quantity }} pcs</strong>
            </p>
          </div>

          <div class="space-y-2">
            <label class="text-sm font-medium">Jumlah</label>
            <input
              type="number"
              v-model="form.quantity"
              class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
            />
          </div>

          <div class="space-y-2">
            <label class="text-sm font-medium">Catatan</label>
            <textarea
              v-model="form.notes"
              placeholder="Contoh: Stok untuk Kanvas Area A"
              class="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
            ></textarea>
          </div>
        </CardContent>
        <CardFooter class="flex justify-end gap-2 bg-slate-50 p-4 border-t rounded-b-lg">
          <Button variant="outline" class="font-bold" @click="showForm = false">Batal</Button>
          <Button class="font-bold px-6" @click="handleSubmit" :loading="loading"
            >Simpan & Ajukan</Button
          >
        </CardFooter>
      </Card>
    </div>

    <!-- Edit/Correction Transfer Modal -->
    <div
      v-if="showEditForm"
      class="fixed inset-0 bg-background/80 backdrop-blur-sm z-50 flex items-center justify-center p-4"
    >
      <Card class="w-full max-w-lg shadow-2xl border-2 overflow-hidden max-h-[90vh] flex flex-col bg-white">
        <CardHeader class="flex-shrink-0">
          <CardTitle>Edit / Koreksi Perpindahan Barang</CardTitle>
          <CardDescription>
            Cabang: {{ currentBranchName }} | Tipe: {{ editForm.type }} | Status: {{ editForm.status }}
          </CardDescription>
        </CardHeader>
        <CardContent class="space-y-4 overflow-y-auto flex-grow pr-6">
          <div class="space-y-2">
            <label class="text-sm font-medium">Salesman</label>
            <input
              type="text"
              disabled
              :value="editForm.employee_name"
              class="flex h-10 w-full rounded-md border border-input bg-slate-50 px-3 py-2 text-sm text-slate-500 cursor-not-allowed"
            />
          </div>

          <div class="space-y-2">
            <label class="text-sm font-medium">Produk</label>
            <input
              type="text"
              disabled
              :value="editForm.product_name"
              class="flex h-10 w-full rounded-md border border-input bg-slate-50 px-3 py-2 text-sm text-slate-500 cursor-not-allowed"
            />
          </div>

          <!-- If status is PENDING/WAITING_SUPERVISOR, allow editing quantity directly -->
          <div v-if="editForm.status === 'PENDING' || editForm.status === 'WAITING_SUPERVISOR'" class="space-y-2">
            <label class="text-sm font-medium">Jumlah</label>
            <input
              type="number"
              v-model="editForm.quantity"
              class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
            />
          </div>

          <!-- If status is COMPLETED, allow correcting Actual and Damaged quantities -->
          <div v-else-if="editForm.status === 'COMPLETED'" class="space-y-4">
            <div class="bg-amber-50 border border-amber-200 rounded-lg p-3 text-xs text-amber-800 flex gap-2">
              <AlertTriangle class="h-5 w-5 text-amber-600 flex-shrink-0" />
              <div>
                Koreksi transaksi yang sudah <strong>SELESAI</strong> akan secara otomatis menyesuaikan stok di gudang dan stok salesman, serta mencatat log audit.
              </div>
            </div>

            <div class="grid grid-cols-3 gap-4">
              <div class="space-y-2">
                <label class="text-sm font-medium">Total Qty</label>
                <input
                  type="number"
                  v-model="editForm.quantity"
                  class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                />
              </div>
              <div class="space-y-2">
                <label class="text-sm font-medium">Bagus (Good)</label>
                <input
                  type="number"
                  v-model="editForm.actual_quantity"
                  class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                />
              </div>
              <div class="space-y-2">
                <label class="text-sm font-medium">Rusak (Damaged)</label>
                <input
                  type="number"
                  v-model="editForm.damaged_quantity"
                  class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                />
              </div>
            </div>

            <div class="space-y-2">
              <label class="text-sm font-medium">Alasan Koreksi <span class="text-red-500">*</span></label>
              <input
                type="text"
                v-model="editForm.reason"
                placeholder="Contoh: Selisih hitung fisik saat return"
                class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              />
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-sm font-medium">Catatan</label>
            <textarea
              v-model="editForm.notes"
              placeholder="Tambahkan informasi tambahan jika ada"
              class="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
            ></textarea>
          </div>

          <!-- Log Riwayat Koreksi / Log Fleksibilitas -->
          <div v-if="editForm.corrections && editForm.corrections.length > 0" class="space-y-2 pt-4 border-t border-slate-100">
            <h3 class="text-sm font-bold text-slate-700">Riwayat Log Koreksi & Barang Rusak</h3>
            <div class="space-y-2 max-h-[150px] overflow-y-auto pr-2">
              <div
                v-for="c in editForm.corrections"
                :key="c.id"
                class="bg-slate-50 border border-slate-200 rounded p-2.5 text-xs text-slate-600 space-y-1"
              >
                <div class="flex justify-between font-semibold text-slate-700">
                  <span>Oleh: {{ c.corrected_by }}</span>
                  <span class="text-[10px] text-slate-400 font-normal">
                    {{ new Date(c.created_at).toLocaleString('id-ID') }}
                  </span>
                </div>
                <div class="grid grid-cols-2 gap-1 text-[11px]">
                  <div>Qty Total: {{ c.old_quantity }} &rarr; {{ c.new_quantity }}</div>
                  <div>Qty Rusak: {{ c.old_damaged_qty }} &rarr; {{ c.new_damaged_qty }}</div>
                </div>
                <div class="text-[11px] italic text-slate-500 bg-white p-1 rounded border border-slate-100 mt-1">
                  Alasan: {{ c.correction_reason }}
                </div>
              </div>
            </div>
          </div>
        </CardContent>
        <CardFooter class="flex justify-end gap-2 bg-slate-50 p-4 border-t rounded-b-lg flex-shrink-0">
          <Button variant="outline" class="font-bold" @click="showEditForm = false">Batal</Button>
          <Button class="font-bold px-6 bg-amber-600 hover:bg-amber-700 text-white" @click="handleEditSubmit" :loading="loading"
            >Simpan Perubahan</Button
          >
        </CardFooter>
      </Card>
    </div>

    <!-- Confirmation Dialog -->
    <ConfirmationDialog
      v-model:open="confirmDialog.open"
      :title="confirmDialog.title"
      :description="confirmDialog.description"
      :variant="confirmDialog.variant"
      :confirm-text="confirmDialog.confirmText"
      :cancel-text="confirmDialog.cancelText"
      :is-loading="confirmDialog.isLoading"
      @confirm="confirmDialog.onConfirm"
      @cancel="confirmDialog.open = false"
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, computed, ref, reactive, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useSalesTransferStore } from '@/stores/salesTransfer'
import { useWarehouseStore } from '@/stores/warehouse'
import { useMasterDataStore } from '@/stores/masterData'
import { useFactoryStore } from '@/stores/factory'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card/index'
import DataTable from '@/components/DataTable.vue'
import ConfirmationDialog from '@/components/ui/ConfirmationDialog.vue'
import { toast } from 'vue-sonner'
import {
  Plus,
  History,
  UserSquare2,
  ArrowRight,
  ArrowLeft,
  Trash2,
  Search,
  X,
  Edit,
  AlertTriangle,
  Gift,
} from 'lucide-vue-next'

const transferStore = useSalesTransferStore()
const warehouseStore = useWarehouseStore()
const masterStore = useMasterDataStore()
const factoryStore = useFactoryStore()
const route = useRoute()

const showForm = ref(false)
const showEditForm = ref(false)
const loading = ref(false)
const selectedSalesman = ref('')
const masterDataStore = useMasterDataStore()

const editForm = reactive({
  id: '',
  employee_name: '',
  product_name: '',
  quantity: 0,
  actual_quantity: 0,
  damaged_quantity: 0,
  type: 'TRANSFER',
  status: 'PENDING',
  notes: '',
  reason: '',
  corrections: [] as any[],
})

const currentBranchName = computed(() => {
  const branch = masterDataStore.branches.find((b) => b.id === masterDataStore.selectedWarehouseBranchId)
  return branch ? branch.name : 'Pusat'
})

const activeSalesStock = computed(() => {
  return transferStore.salesStock.filter((stock: any) => stock.quantity > 0)
})

const getProductStock = (productId: string): number => {
  let branchId = masterDataStore.selectedWarehouseBranchId
  if ((!branchId || branchId === 'ALL_BRANCHES') && form.employee_id) {
    const selectedEmp = masterStore.employees.find((e: any) => e.id === form.employee_id)
    if (selectedEmp?.branch_id) {
      branchId = selectedEmp.branch_id
    }
  }

  return warehouseStore.inventory
    .filter((i: any) => {
      if (i.product_id !== productId) return false
      if (branchId && branchId !== 'ALL_BRANCHES' && i.branch_id && i.branch_id !== branchId) return false
      if (i.stock_type && i.stock_type !== 'GOOD') return false
      if (i.batch_no === 'BACKORDER') return false
      return true
    })
    .reduce((sum: number, item: any) => sum + (Number(item.quantity) || 0), 0)
}

const selectedProductId = computed(() => {
  if (form.item_selection && form.item_selection.startsWith('PRODUCT:')) {
    return form.item_selection.replace('PRODUCT:', '')
  }
  return null
})

const availableBatchesForSelectedProduct = computed(() => {
  if (!selectedProductId.value) return []

  let branchId = masterDataStore.selectedWarehouseBranchId
  if ((!branchId || branchId === 'ALL_BRANCHES') && form.employee_id) {
    const selectedEmp = masterStore.employees.find((e: any) => e.id === form.employee_id)
    if (selectedEmp?.branch_id) {
      branchId = selectedEmp.branch_id
    }
  }

  const items = warehouseStore.inventory.filter((i: any) => {
    if (i.product_id !== selectedProductId.value) return false
    if (branchId && branchId !== 'ALL_BRANCHES' && i.branch_id && i.branch_id !== branchId) return false
    if (i.stock_type && i.stock_type !== 'GOOD') return false
    if (i.batch_no === 'BACKORDER') return false
    return (Number(i.quantity) || 0) > 0
  })

  const batchMap = new Map<string, any>()
  for (const item of items) {
    const bNo = (item.batch_no || 'DEFAULT').trim() || 'DEFAULT'
    if (!batchMap.has(bNo)) {
      batchMap.set(bNo, {
        batch_no: bNo,
        quantity: 0,
        expiry_date: item.expiry_date,
      })
    }
    batchMap.get(bNo).quantity += Number(item.quantity) || 0
  }

  return Array.from(batchMap.values()).sort((a, b) => {
    if (a.expiry_date && b.expiry_date) {
      return new Date(a.expiry_date).getTime() - new Date(b.expiry_date).getTime()
    }
    return 0
  })
})

const selectedBatchStockInfo = computed(() => {
  if (!form.batch_no) return null
  return availableBatchesForSelectedProduct.value.find((b) => b.batch_no === form.batch_no) || null
})

const isSalesPosition = (title?: string, deptName?: string) => {
  const t = (title || '').toLowerCase()
  const d = (deptName || '').toLowerCase()
  const isExcluded = /(operator|driver|sopir|kuli|helper|gudang|warehouse|security|cleaning|harian|validator|editor|designer|desain|purchasing|auditor|kasir|cashier|teknisi|technician)/i.test(t)
  if (isExcluded) return false
  return /(sales|motoris|booster|marketing|take order|task order|\bto\b|\bmt\b|modern trade|key account|\bkam\b|kanvas)/i.test(t) ||
         /(sales|marketing|commercial)/i.test(d)
}

const salesEmployees = computed(() => {
  const branchId = masterDataStore.selectedWarehouseBranchId
  const list = masterStore.employees.filter((emp: any) => {
    const isActive = emp.user?.is_active !== false &&
      emp.employment_status !== 'RESIGNED' &&
      emp.employment_status !== 'TERMINATED' &&
      !String(emp.employment_status || '').toLowerCase().includes('harian')
    if (!isActive) return false

    if (branchId && branchId !== 'ALL_BRANCHES') {
      if (emp.branch_id && emp.branch_id !== branchId) return false
    }
    const jobTitle = emp.job_position?.title || ''
    const deptName = emp.department?.name || ''
    return isSalesPosition(jobTitle, deptName)
  })

  return list.sort((a: any, b: any) => {
    const nameA = `${a.first_name} ${a.last_name || ''}`.trim()
    const nameB = `${b.first_name} ${b.last_name || ''}`.trim()
    return nameA.localeCompare(nameB)
  })
})
const filteredTransfers = computed(() => {
  const normalTransfers = transferStore.transfers.filter((t: any) => {
    if (!t.employee) return true
    const jobTitle = t.employee.job_position?.title || ''
    const deptName = t.employee.department?.name || ''
    return isSalesPosition(jobTitle, deptName)
  })

  const bonusRequestsMapped = transferStore.bonusRequests.map((r: any) => ({
    id: r.id,
    type: 'BONUS_REQUEST',
    employee: r.employee,
    bonus_inventory: r.bonus_inventory,
    bonus_inventory_id: r.bonus_inventory_id,
    quantity: r.quantity,
    status: r.status,
    notes: r.notes,
    transfer_date: r.created_at,
    is_bonus_request: true,
  }))

  return [...bonusRequestsMapped, ...normalTransfers]
})
const confirmDialog = ref({
  open: false,
  title: '',
  description: '',
  variant: 'info' as 'danger' | 'warning' | 'success' | 'info',
  confirmText: '',
  cancelText: '',
  isLoading: false,
  onConfirm: () => {},
})

const form = reactive({
  employee_id: '',
  item_selection: '', // format: "PRODUCT:uuid" or "BONUS:uuid"
  batch_no: '',
  quantity: 1,
  type: 'TRANSFER',
  notes: '',
})

watch(
  () => form.item_selection,
  () => {
    form.batch_no = ''
  },
)

const openForm = async () => {
  showForm.value = true
  loading.value = true
  try {
    await Promise.all([
      masterStore.fetchEmployees(true),
      warehouseStore.fetchInventory(),
      warehouseStore.fetchBonusInventory(),
      factoryStore.fetchProducts(),
    ])
  } finally {
    loading.value = false
  }
}

const columns = [
  { key: 'reference_no', label: 'No. Nota/Ref' },
  { key: 'type', label: 'Tipe' },
  { key: 'employee_name', label: 'Salesman' },
  { key: 'product_name', label: 'Barang' },
  { key: 'quantity', label: 'Jumlah' },
  { key: 'status', label: 'Status' },
  { key: 'transfer_date', label: 'Tanggal', type: 'datetime' },
  { key: 'actions', label: 'Aksi' },
]

const fetchSalesStock = () => {
  if (selectedSalesman.value) {
    transferStore.fetchSalesStock(selectedSalesman.value)
  }
}

const handleSubmit = async () => {
  if (!form.employee_id || !form.item_selection || form.quantity <= 0) {
    toast.error('Mohon lengkapi semua data.')
    return
  }

  const payload: any = {
    employee_id: form.employee_id,
    quantity: form.quantity,
    type: form.type,
    notes: form.notes,
  }

  if (form.item_selection.startsWith('PRODUCT:')) {
    payload.product_id = form.item_selection.replace('PRODUCT:', '')
    if (form.batch_no) {
      payload.batch_no = form.batch_no
    }
  } else if (form.item_selection.startsWith('BONUS:')) {
    payload.bonus_inventory_id = form.item_selection.replace('BONUS:', '')
  }

  loading.value = true
  try {
    await transferStore.createTransfer(payload)
    showForm.value = false
    // Reset form
    form.employee_id = ''
    form.item_selection = ''
    form.batch_no = ''
    form.quantity = 0
    form.notes = ''
    toast.success('Perpindahan barang berhasil diajukan')
  } catch (err: any) {
    toast.error(err || 'Gagal mengajukan perpindahan barang')
  } finally {
    loading.value = false
  }
}

const handleEdit = (row: any) => {
  editForm.id = row.id
  editForm.employee_name = `${row.employee?.first_name || ''} ${row.employee?.last_name || ''}`
  editForm.product_name = row.product?.name || row.bonus_inventory?.name || ''
  editForm.quantity = row.quantity
  editForm.actual_quantity = row.actual_quantity || row.quantity
  editForm.damaged_quantity = row.damaged_quantity || 0
  editForm.type = row.type
  editForm.status = row.status
  editForm.notes = row.notes || ''
  editForm.reason = ''
  editForm.corrections = row.corrections || []
  showEditForm.value = true
}

const handleEditSubmit = async () => {
  if (editForm.status === 'COMPLETED' && !editForm.reason) {
    toast.error('Alasan koreksi wajib diisi untuk transaksi yang sudah selesai.')
    return
  }

  loading.value = true
  try {
    await transferStore.updateTransfer(editForm.id, {
      quantity: editForm.quantity,
      actual_quantity: editForm.actual_quantity,
      damaged_quantity: editForm.damaged_quantity,
      notes: editForm.notes,
      reason: editForm.reason,
    })
    showEditForm.value = false
    warehouseStore.fetchInventory() // Refresh warehouse stock
    if (selectedSalesman.value) fetchSalesStock() // Refresh sales stock check
    toast.success('Transaksi berhasil diperbarui')
  } catch (err: any) {
    toast.error(err || 'Gagal memperbarui transaksi')
  } finally {
    loading.value = false
  }
}

const handleComplete = (id: string) => {
  confirmDialog.value = {
    open: true,
    title: 'Selesaikan Perpindahan',
    description: 'Selesaikan perpindahan ini? Stok akan dipotong/ditambah.',
    variant: 'success',
    confirmText: 'Ya, Selesaikan',
    cancelText: 'Kembali',
    isLoading: false,
    onConfirm: async () => {
      confirmDialog.value.isLoading = true
      try {
        await transferStore.completeTransfer(id)
        warehouseStore.fetchInventory() // Refresh warehouse stock
        if (selectedSalesman.value) fetchSalesStock()
        confirmDialog.value.open = false
        toast.success('Perpindahan barang berhasil diselesaikan')
      } catch (err: any) {
        toast.error(err || 'Gagal menyelesaikan perpindahan barang')
      } finally {
        confirmDialog.value.isLoading = false
      }
    },
  }
}

const handleGenerateDO = (row: any) => {
  const isReturn = row.type === 'RETURN'
  const title = isReturn ? 'Sisipkan Retur ke DO' : 'Buat DO Motoris'
  const description = isReturn 
    ? `Sisipkan barang retur ini ke DO salesman hari ini? Barang akan otomatis masuk ke kolom "Pulang / Retur".`
    : `Buatkan DO Motoris untuk perpindahan ini? DO akan masuk ke menu Kelola DO Spares untuk diisi jumlah muatannya oleh Admin Gudang, lalu di-Finalize untuk menyelesaikan perpindahan.`
  const successMsg = isReturn
    ? 'Retur berhasil disisipkan ke DO! Silakan cek di menu Kelola DO Spares.'
    : 'DO Motoris berhasil dibuat! Silakan cek di menu Kelola DO Spares untuk mengisi muatan dan finalisasi.'

  confirmDialog.value = {
    open: true,
    title: title,
    description: description,
    variant: 'info',
    confirmText: isReturn ? 'Ya, Sisipkan' : 'Ya, Buat DO',
    cancelText: 'Kembali',
    isLoading: false,
    onConfirm: async () => {
      confirmDialog.value.isLoading = true
      try {
        await transferStore.generateDO([row.id])
        confirmDialog.value.open = false
        toast.success(successMsg)
        await transferStore.fetchTransfers(masterStore.selectedBranchId)
      } catch (err: any) {
        toast.error(err || 'Gagal memproses ke DO Motoris')
      } finally {
        confirmDialog.value.isLoading = false
      }
    },
  }
}

const handleCancel = (id: string) => {
  confirmDialog.value = {
    open: true,
    title: 'Batalkan Perpindahan',
    description: 'Apakah Anda yakin ingin membatalkan perpindahan barang ini?',
    variant: 'danger',
    confirmText: 'Ya, Batalkan',
    cancelText: 'Kembali',
    isLoading: false,
    onConfirm: async () => {
      confirmDialog.value.isLoading = true
      try {
        await transferStore.cancelTransfer(id)
        confirmDialog.value.open = false
        toast.success('Perpindahan barang berhasil dibatalkan')
      } catch (err: any) {
        toast.error(err || 'Gagal membatalkan perpindahan barang')
      } finally {
        confirmDialog.value.isLoading = false
      }
    },
  }
}

const handleDelete = (id: string) => {
  confirmDialog.value = {
    open: true,
    title: 'Hapus Riwayat',
    description: 'Yakin ingin menghapus riwayat ini secara permanen?',
    variant: 'danger',
    confirmText: 'Hapus',
    cancelText: 'Tutup',
    isLoading: false,
    onConfirm: async () => {
      try {
        confirmDialog.value.isLoading = true
        await transferStore.deleteTransfer(id)
        toast.success('Riwayat berhasil dihapus')
        confirmDialog.value.open = false
      } catch (e: any) {
        toast.error(e)
      } finally {
        confirmDialog.value.isLoading = false
      }
    },
  }
}

const handleApproveBonus = (id: string) => {
  confirmDialog.value = {
    open: true,
    title: 'Setujui Permintaan Bonus',
    description: 'Yakin ingin menyetujui permintaan bonus ini? Stok bonus di gudang akan otomatis terpotong.',
    variant: 'success',
    confirmText: 'Setujui',
    cancelText: 'Batal',
    isLoading: false,
    onConfirm: async () => {
      try {
        confirmDialog.value.isLoading = true
        await transferStore.approveBonusRequest(id)
        toast.success('Permintaan bonus disetujui')
        confirmDialog.value.open = false
        warehouseStore.fetchBonusInventory()
      } catch (e: any) {
        toast.error(e)
      } finally {
        confirmDialog.value.isLoading = false
      }
    },
  }
}

const handleRejectBonus = (id: string) => {
  confirmDialog.value = {
    open: true,
    title: 'Tolak Permintaan Bonus',
    description: 'Yakin ingin menolak permintaan bonus ini?',
    variant: 'danger',
    confirmText: 'Tolak',
    cancelText: 'Batal',
    isLoading: false,
    onConfirm: async () => {
      try {
        confirmDialog.value.isLoading = true
        await transferStore.rejectBonusRequest(id)
        toast.success('Permintaan bonus ditolak')
        confirmDialog.value.open = false
      } catch (e: any) {
        toast.error(e)
      } finally {
        confirmDialog.value.isLoading = false
      }
    },
  }
}

watch(
  () => masterStore.selectedWarehouseBranchId,
  () => {
    transferStore.fetchTransfers()
    transferStore.fetchBonusRequests()
    warehouseStore.fetchInventory()
  },
)

onMounted(() => {
  masterStore.fetchBranches()
  transferStore.fetchTransfers()
  transferStore.fetchBonusRequests()
  warehouseStore.fetchInventory()
  masterStore.fetchEmployees()
  factoryStore.fetchProducts()

  if (route.path.includes('/retur')) {
    form.type = 'RETURN'
    openForm()
  }
})
</script>
