import { defineStore } from 'pinia'
import { factoryApi } from '@/api/factory'
import {
  filterSellableProducts,
  isRawOrPackagingMaterial,
  isRawMaterial,
  isPackagingMaterial,
} from '@/utils/productUnits'

export interface FactoryWarehouse {
  id: string
  factory_id: string
  name: string
  code: string
  type: string
  description: string
  created_at: string
}

export interface Factory {
  id: string
  company_id: string
  branch_id: string
  name: string
  location: string
  warehouse_code?: string
  branch?: {
    id: string
    name: string
  }
}

export interface Product {
  id: string
  company_id: string
  name: string
  sku: string
  unit: string
  pcs_per_unit: number
  weight: number
  weight_unit: string
  length?: number
  width?: number
  height?: number
  volume?: number
  cost_price: number
  factory_price: number
  selling_price: number
  dbp: number
  wbp: number
  rbp: number
  category?: string
  brand?: string
  specs?: string
  description: string
  image_url?: string
  shelf_life_days?: number
  bpom_number?: string
  bpom_expiry_date?: string
  halal_number?: string
  halal_expiry_date?: string
}

export interface FactoryStock {
  id: string
  factory_id: string
  product_id: string
  factory_warehouse_id?: string
  batch_no: string
  expiry_date?: string
  quantity: number
  updated_at: string
  product?: Product
}

export interface ProductionLogMaterial {
  id?: string
  production_log_id?: string
  product_id: string
  batch_no: string
  quantity: number
  notes?: string
  product?: Product
}

export interface ProductionLog {
  id: string
  factory_id: string
  product_id: string
  batch_no: string
  expiry_date?: string
  employee_id: string
  quantity: number
  carton_count?: number
  pieces_per_carton?: number
  production_date: string
  notes: string
  product?: Product
  employee?: {
    first_name: string
    last_name: string
  }
  materials?: ProductionLogMaterial[]
}

export interface FactoryQcSample {
  id: string
  factory_id: string
  product_id: string
  production_log_id?: string | null
  batch_no: string
  expiry_date?: string | null
  production_date?: string | null
  quantity: number
  sample_type: string
  purpose: string
  status: 'PENDING' | 'APPROVED' | 'REJECTED'
  requested_by_id: string
  requested_by_name: string
  requested_at: string
  approved_by_id?: string | null
  approved_by_name?: string
  approved_at?: string | null
  rejection_reason?: string
  notes?: string
  created_at: string
  updated_at: string
  product?: Product
  factory?: Factory
  production_log?: ProductionLog
}

export interface ProductTransfer {
  id: string
  from_factory_id: string
  to_branch_id: string
  product_id: string
  quantity: number
  total_weight: number
  total_volume?: number
  status: string
  batch_no?: string
  delivery_order_no?: string
  notes: string
  target_shipment_date?: string
  estimated_arrival?: string
  initiated_by?: string
  created_at: string
  product?: Product
  from_factory?: {
    name: string
  }
  fulfillment_branch?: {
    name: string
  }
  to_branch?: {
    name: string
  }
  unit?: string
  vehicle_id?: string
  driver_id?: string
  vehicle?: {
    name: string
    license_plate: string
    capacity: number
    bed_length?: number
    bed_width?: number
    bed_height?: number
    bed_volume?: number
  }
  driver?: {
    first_name: string
    last_name: string
  }
  pcs_per_unit?: number
  fulfillment_branch_id?: string | null
  is_stock_borrowed?: boolean
  replenishment_transfer_id?: string
  transfer_type?: string
  is_billable?: boolean
  payment_method?: string
  payment_status?: string
  total_amount?: number
  payment_due_at?: string
}

export interface LowStockAlert {
  warehouse_stock_id: string
  branch_id: string
  branch_name: string
  product_id: string
  product_name: string
  product_sku: string
  product_unit: string
  pcs_per_unit: number
  quantity: number
  min_limit: number
  batch_no: string
}

export interface FactoryReturnItem {
  id?: string
  factory_return_id?: string
  product_id: string
  batch_no: string
  quantity: number
  unit: string
  pcs_per_unit: number
  return_type?: string // 'GOOD', 'DAMAGED'
  damage_type?: string // 'BOTTLE_PACKAGING', 'PRODUCT_CONTENT', 'EXPIRED', 'TOTAL_DAMAGE', 'OTHER'
  damage_reason?: string
  expiry_date?: string
  received_quantity?: number
  received_good_qty?: number
  received_bottle_damaged_qty?: number
  received_content_damaged_qty?: number
  disposition_action?: string // 'RESTOCK_GOOD', 'REPACK_BOTTLE', 'SCRAP_DISPOSAL', 'PARTIAL_MIXED'
  inspection_notes?: string
  target_factory_warehouse_id?: string | null
  product?: Product
  factory_warehouse?: FactoryWarehouse
}

export interface FactoryReturn {
  id: string
  return_no: string
  from_branch_id: string
  to_factory_id: string
  status: string // 'REQUESTED', 'APPROVED', 'SHIPPED', 'RECEIVED'
  notes?: string
  is_received: boolean
  received_by?: string
  received_at?: string
  receiver_name?: string
  inspected_by_name?: string
  inspection_notes?: string
  total_received_qty?: number
  total_good_qty?: number
  total_bottle_damaged_qty?: number
  total_content_damaged_qty?: number
  shipped_at?: string
  created_at: string
  updated_at: string
  from_branch?: {
    id: string
    name: string
  }
  to_factory?: {
    id: string
    name: string
  }
  vehicle?: any
  driver?: any
  items: FactoryReturnItem[]
}

export interface ReceiveReturnPayload {
  receiver_name?: string
  inspected_by_name?: string
  inspection_notes?: string
  items?: {
    item_id: string
    received_good_qty: number
    received_bottle_damaged_qty: number
    received_content_damaged_qty: number
    disposition_action?: string
    inspection_notes?: string
    target_factory_warehouse_id?: string | null
  }[]
}

import { factoryReturnApi } from '@/api/factoryReturn'

export interface RawMaterialPRItem {
  id?: string
  purchase_request_id?: string
  product_id?: string
  item_name?: string
  factory_warehouse_id?: string | null
  requested_qty: number
  unit: string
  notes?: string
  product?: Product
  factory_warehouse?: FactoryWarehouse
}

export interface RawMaterialPR {
  id: string
  company_id: string
  factory_id?: string
  pr_number: string
  request_date: string
  required_date?: string
  status: 'PENDING' | 'PROCESSED' | 'REJECTED' | 'CANCELLED'
  notes: string
  created_by_name?: string
  reviewed_by_name?: string
  reviewed_at?: string
  review_notes?: string
  rejection_reason?: string
  purchase_order_id?: string | null
  purchase_order?: any
  created_at: string
  factory?: Factory
  items?: RawMaterialPRItem[]
}

export interface RawMaterialPOItem {
  id?: string
  purchase_order_id?: string
  product_id: string
  factory_warehouse_id?: string | null
  requested_qty: number
  received_qty?: number
  unit: string
  estimated_unit_price: number
  subtotal?: number
  notes?: string
  product?: Product
  factory_warehouse?: FactoryWarehouse
}

export interface RawMaterialPO {
  id: string
  company_id: string
  factory_id: string
  supplier_id: string
  purchase_request_id?: string | null
  purchase_request?: RawMaterialPR
  po_number: string
  order_date: string
  estimated_arrival?: string
  status: 'DRAFT' | 'PENDING_SPV' | 'PENDING_PURCHASING' | 'APPROVED' | 'PARTIAL_RECEIVED' | 'RECEIVED' | 'REJECTED' | 'CANCELLED'
  total_estimated_amount: number
  notes: string
  created_by_name?: string
  spv_approved_by_name?: string
  spv_approved_by_position?: string
  spv_approved_at?: string
  spv_notes?: string
  purchasing_approved_by_name?: string
  purchasing_approved_by_position?: string
  purchasing_approved_at?: string
  purchasing_notes?: string
  progress_status?: string
  progress_notes?: string
  progress_updated_by_name?: string
  progress_updated_at?: string
  payment_status?: string
  payment_notes?: string
  rejection_reason?: string
  rejected_by_name?: string
  rejected_at?: string
  created_at: string
  factory?: Factory
  supplier?: {
    id: string
    name: string
    contact_name?: string
    phone?: string
  }
  items?: RawMaterialPOItem[]
  receipts?: any[]
}

export interface RawMaterialReceiptItem {
  id?: string
  receipt_id?: string
  po_item_id?: string | null
  product_id: string
  factory_warehouse_id?: string | null
  batch_no: string
  expiry_date?: string
  received_qty: number
  rejected_qty: number
  unit: string
  unit_price?: number
  notes?: string
  product?: Product
  factory_warehouse?: FactoryWarehouse
}

export interface RawMaterialReceipt {
  id: string
  company_id: string
  factory_id: string
  purchase_order_id?: string | null
  supplier_id: string
  receipt_no: string
  supplier_delivery_note_no?: string
  delivery_date?: string
  received_date: string
  received_by_name?: string
  proof_image_url?: string
  notes?: string
  unloading_cost?: number
  unloading_notes?: string
  kas_pabrik_id?: string | null
  kas_transaction_id?: string | null
  kas_pabrik?: any
  created_at: string
  factory?: Factory
  supplier?: {
    id: string
    name: string
  }
  purchase_order?: RawMaterialPO
  items?: RawMaterialReceiptItem[]
}

export const useFactoryStore = defineStore('factory', {
  state: () => ({
    factories: [] as Factory[],
    factoryWarehouses: [] as FactoryWarehouse[],
    products: [] as Product[],
    // Cache metadata for products
    _productsFetchedAt: 0 as number,
    _productsFetching: null as Promise<void> | null,
    productsLoading: false,
    currentFactory: null as Factory | null,
    inventory: [] as FactoryStock[],
    allInventory: [] as FactoryStock[],
    inventoryLogs: [] as any[],
    allInventoryLogs: [] as any[],
    productionHistory: [] as ProductionLog[],
    allProductionHistory: [] as ProductionLog[],
    transferHistory: [] as ProductTransfer[],
    allTransfers: [] as ProductTransfer[],
    delegatedShipments: [] as ProductTransfer[],
    delegatedSummary: [] as any[],
    factoryReturns: [] as any[],
    backorderDemand: [] as any[],
    lowStockAlerts: [] as LowStockAlert[],
    dashboardStats: {
      total_factories: 0,
      total_products: 0,
      pending_shipments: 0,
      today_production: 0,
      recent_transfers: [] as ProductTransfer[],
    },
    rawMaterialStock: [] as any[],
    rawMaterialLogs: [] as any[],
    rawMaterialPRs: [] as RawMaterialPR[],
    rawMaterialPOs: [] as RawMaterialPO[],
    rawMaterialReceipts: [] as RawMaterialReceipt[],
    qcSamples: [] as FactoryQcSample[],
    loading: false,
    error: null as string | null,
  }),

  getters: {
    sellableProducts: (state): Product[] => filterSellableProducts(state.products),
    rawMaterialProducts: (state): Product[] => state.products.filter((p) => isRawOrPackagingMaterial(p)),
    rawMaterialsOnly: (state): Product[] => state.products.filter((p) => isRawMaterial(p)),
    packagingMaterialsOnly: (state): Product[] => state.products.filter((p) => isPackagingMaterial(p)),
  },

  actions: {
    async fetchRawMaterialStock(factoryId: string) {
      this.loading = true
      try {
        const response = await factoryApi.getRawMaterialStock(factoryId)
        this.rawMaterialStock = response.data?.data || response.data || []
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch raw material stock'
      } finally {
        this.loading = false
      }
    },
    async fetchRawMaterialLogs(factoryId: string, params?: { product_id?: string }) {
      this.loading = true
      try {
        const response = await factoryApi.getRawMaterialLogs(factoryId, params)
        const logs = response.data?.data || response.data || []
        if (!params?.product_id) {
          this.rawMaterialLogs = logs
        }
        return logs
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch raw material logs'
        return []
      } finally {
        this.loading = false
      }
    },
    async mutateRawMaterial(factoryId: string, data: any) {
      try {
        await factoryApi.mutateRawMaterial(factoryId, data)
        await this.fetchRawMaterialStock(factoryId)
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to mutate raw material'
      }
    },
    async deleteStock(factoryId: string, id: string) {
      try {
        await factoryApi.deleteStock(id)
        await this.fetchRawMaterialStock(factoryId)
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to delete stock'
      }
    },
    async deleteRawMaterialLog(factoryId: string, id: string) {
      try {
        await factoryApi.deleteRawMaterialLog(id)
        await this.fetchRawMaterialLogs(factoryId)
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to delete log'
      }
    },
    async fetchRawMaterialPRs(params?: { factory_id?: string; status?: string }) {
      this.loading = true
      try {
        const response = await factoryApi.getRawMaterialPRs(params)
        this.rawMaterialPRs = response.data?.data || []
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch raw material purchase requests'
      } finally {
        this.loading = false
      }
    },
    async createRawMaterialPR(data: any) {
      try {
        const res = await factoryApi.createRawMaterialPR(data)
        if (data.factory_id) {
          await this.fetchRawMaterialPRs({ factory_id: data.factory_id })
        }
        return res.data
      } catch (err: any) {
        throw err.response?.data?.error || 'Gagal membuat Purchase Request Bahan Baku'
      }
    },
    async updateRawMaterialPR(id: string, data: any) {
      try {
        const res = await factoryApi.updateRawMaterialPR(id, data)
        if (data.factory_id) {
          await this.fetchRawMaterialPRs({ factory_id: data.factory_id })
        }
        return res.data
      } catch (err: any) {
        throw err.response?.data?.error || 'Gagal memperbarui Purchase Request'
      }
    },
    async rejectRawMaterialPR(id: string, reason: string, factoryId?: string) {
      try {
        const res = await factoryApi.rejectRawMaterialPR(id, { reason })
        if (factoryId) {
          await this.fetchRawMaterialPRs({ factory_id: factoryId })
        }
        return res.data
      } catch (err: any) {
        throw err.response?.data?.error || 'Gagal menolak Purchase Request'
      }
    },
    async completeRawMaterialPR(id: string, notes: string) {
      try {
        const res = await factoryApi.completeRawMaterialPR(id, { notes })
        return res.data
      } catch (err: any) {
        throw err.response?.data?.error || 'Gagal menindaklanjuti Purchase Request'
      }
    },
    async deleteRawMaterialPR(id: string, factoryId?: string) {
      try {
        await factoryApi.deleteRawMaterialPR(id)
        if (factoryId) {
          await this.fetchRawMaterialPRs({ factory_id: factoryId })
        }
      } catch (err: any) {
        throw err.response?.data?.error || 'Gagal menghapus Purchase Request'
      }
    },
    async fetchRawMaterialPOs(params?: { factory_id?: string; status?: string }) {
      this.loading = true
      try {
        const response = await factoryApi.getRawMaterialPOs(params)
        this.rawMaterialPOs = response.data?.data || []
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch raw material purchase orders'
      } finally {
        this.loading = false
      }
    },
    async fetchRawMaterialReceipts(params?: { factory_id?: string }) {
      this.loading = true
      try {
        const response = await factoryApi.getRawMaterialReceipts(params)
        this.rawMaterialReceipts = response.data?.data || []
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch raw material receipts'
      } finally {
        this.loading = false
      }
    },
    async createRawMaterialPO(data: any) {
      try {
        const res = await factoryApi.createRawMaterialPO(data)
        if (data.factory_id) {
          await this.fetchRawMaterialPOs({ factory_id: data.factory_id })
        }
        return res.data
      } catch (err: any) {
        throw err.response?.data?.error || 'Gagal membuat Purchase Order Bahan Baku'
      }
    },
    async updateRawMaterialPO(id: string, data: any) {
      try {
        const res = await factoryApi.updateRawMaterialPO(id, data)
        if (data.factory_id) {
          await this.fetchRawMaterialPOs({ factory_id: data.factory_id })
        }
        return res.data
      } catch (err: any) {
        throw err.response?.data?.error || 'Gagal memperbarui Purchase Order'
      }
    },
    async deleteRawMaterialPO(id: string, factoryId?: string) {
      try {
        await factoryApi.deleteRawMaterialPO(id)
        if (factoryId) {
          await this.fetchRawMaterialPOs({ factory_id: factoryId })
        }
      } catch (err: any) {
        throw err.response?.data?.error || 'Gagal menghapus Purchase Order'
      }
    },
    async approvePOBySPV(id: string, notes?: string, factoryId?: string) {
      try {
        const res = await factoryApi.approvePOBySPV(id, { notes })
        if (factoryId) {
          await this.fetchRawMaterialPOs({ factory_id: factoryId })
        }
        return res.data
      } catch (err: any) {
        throw err.response?.data?.error || 'Gagal menyetujui PO'
      }
    },
    async approvePOByPurchasing(id: string, notes?: string, factoryId?: string) {
      try {
        const res = await factoryApi.approvePOByPurchasing(id, { notes })
        if (factoryId) {
          await this.fetchRawMaterialPOs({ factory_id: factoryId })
        }
        return res.data
      } catch (err: any) {
        throw err.response?.data?.error || 'Gagal menyetujui PO'
      }
    },
    async updatePOProgress(id: string, data: { progress_status: string; progress_notes?: string; estimated_arrival?: string }, factoryId?: string) {
      try {
        const res = await factoryApi.updatePOProgress(id, data)
        if (factoryId) {
          await this.fetchRawMaterialPOs({ factory_id: factoryId })
        }
        return res.data
      } catch (err: any) {
        throw err.response?.data?.error || 'Gagal memperbarui progress PO'
      }
    },
    async updatePOPayment(id: string, data: { payment_status: string; payment_notes?: string }, factoryId?: string) {
      try {
        const res = await factoryApi.updatePOPayment(id, data)
        if (factoryId) {
          await this.fetchRawMaterialPOs({ factory_id: factoryId })
        }
        return res.data
      } catch (err: any) {
        throw err.response?.data?.error || 'Gagal memperbarui status pembayaran'
      }
    },
    async rejectRawMaterialPO(id: string, reason: string, factoryId?: string) {
      try {
        const res = await factoryApi.rejectRawMaterialPO(id, { reason })
        if (factoryId) {
          await this.fetchRawMaterialPOs({ factory_id: factoryId })
        }
        return res.data
      } catch (err: any) {
        throw err.response?.data?.error || 'Gagal menolak PO'
      }
    },
    async createRawMaterialReceipt(data: any) {
      try {
        const res = await factoryApi.createRawMaterialReceipt(data)
        if (data.factory_id) {
          await this.fetchRawMaterialReceipts({ factory_id: data.factory_id })
          await this.fetchRawMaterialStock(data.factory_id)
          await this.fetchRawMaterialLogs(data.factory_id)
          await this.fetchRawMaterialPOs({ factory_id: data.factory_id })
        }
        return res.data
      } catch (err: any) {
        throw err.response?.data?.error || 'Gagal mencatat penerimaan bahan baku'
      }
    },
    async fetchLowStockAlerts() {
      this.loading = true
      try {
        const response = await factoryApi.getLowStockAlerts()
        this.lowStockAlerts = response.data?.data || response.data || []
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch low stock alerts'
      } finally {
        this.loading = false
      }
    },
    async fetchBackorderDemand(companyId: string) {
      try {
        const response = await factoryApi.getBackorderDemand(companyId)
        this.backorderDemand = response.data?.data || response.data || []
      } catch (err: any) {
        console.error('Failed to fetch backorder demand:', err)
      }
    },
    async fetchDashboardStats() {
      try {
        const response = await factoryApi.getDashboardStats()
        this.dashboardStats = response.data?.data ||
          response.data || {
            total_factories: 0,
            total_products: 0,
            pending_shipments: 0,
            today_production: 0,
            recent_transfers: [],
          }
        if (!this.dashboardStats.recent_transfers) {
          this.dashboardStats.recent_transfers = []
        }
      } catch (err: any) {
        console.error('Failed to fetch dashboard stats:', err)
      }
    },
    async fetchFactories() {
      this.loading = true
      try {
        const response = await factoryApi.getFactories()
        this.factories = response.data?.data || response.data || []
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch factories'
      } finally {
        this.loading = false
      }
    },

    async createFactory(data: any) {
      try {
        await factoryApi.createFactory(data)
        await this.fetchFactories()
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to create factory'
      }
    },

    async updateFactory(id: string, data: any) {
      try {
        await factoryApi.updateFactory(id, data)
        await this.fetchFactories()
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to update factory'
      }
    },

    async deleteFactory(id: string) {
      try {
        await factoryApi.deleteFactory(id)
        await this.fetchFactories()
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to delete factory'
      }
    },

    async fetchFactoryWarehouses(factoryId: string) {
      this.loading = true
      try {
        const response = await factoryApi.getFactoryWarehouses(factoryId)
        this.factoryWarehouses = response.data || []
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch factory warehouses'
      } finally {
        this.loading = false
      }
    },

    async createFactoryWarehouse(factoryId: string, data: any) {
      try {
        await factoryApi.createFactoryWarehouse(factoryId, data)
        await this.fetchFactoryWarehouses(factoryId)
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to create factory warehouse'
      }
    },

    async updateFactoryWarehouse(factoryId: string, warehouseId: string, data: any) {
      try {
        await factoryApi.updateFactoryWarehouse(warehouseId, data)
        await this.fetchFactoryWarehouses(factoryId)
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to update factory warehouse'
      }
    },

    async deleteFactoryWarehouse(factoryId: string, warehouseId: string) {
      try {
        await factoryApi.deleteFactoryWarehouse(warehouseId)
        await this.fetchFactoryWarehouses(factoryId)
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to delete factory warehouse'
      }
    },

    async fetchProducts(force = false) {
      const CACHE_TTL_MS = 5 * 60 * 1000 // 5 minutes
      const now = Date.now()

      // Return cached data if within TTL and not forced
      if (!force && this.products.length > 0 && now - this._productsFetchedAt < CACHE_TTL_MS) {
        return
      }

      // Deduplicate in-flight requests: if already fetching, wait for same promise
      if (this._productsFetching) {
        return this._productsFetching
      }

      this.productsLoading = true
      this._productsFetching = (async () => {
        try {
          const response = await factoryApi.getProducts()
          this.products = response.data?.data || response.data || []
          this._productsFetchedAt = Date.now()
        } catch (err: any) {
          this.error = err.response?.data?.error || 'Failed to fetch products'
        } finally {
          this.productsLoading = false
          this._productsFetching = null
        }
      })()

      return this._productsFetching
    },

    /** Invalidate product cache (call after create/update/delete product) */
    invalidateProductsCache() {
      this._productsFetchedAt = 0
    },

    async createProduct(data: any) {
      try {
        await factoryApi.createProduct(data)
        this.invalidateProductsCache()
        await this.fetchProducts()
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to create product'
      }
    },

    async updateProduct(id: string, data: any) {
      try {
        await factoryApi.updateProduct(id, data)
        this.invalidateProductsCache()
        await this.fetchProducts()
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to update product'
      }
    },

    async deleteProduct(id: string) {
      try {
        await factoryApi.deleteProduct(id)
        this.invalidateProductsCache()
        await this.fetchProducts()
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to delete product'
      }
    },

    async fetchFactoryDetail(id: string) {
      this.loading = true
      try {
        const response = await factoryApi.getFactory(id)
        this.currentFactory = response.data
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch factory detail'
      } finally {
        this.loading = false
      }
    },

    async fetchInventory(factoryId: string) {
      try {
        const response = await factoryApi.getInventory(factoryId)
        this.inventory = response.data?.data || response.data || []
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch inventory'
      }
    },

    async fetchInventoryLogs(factoryId: string) {
      try {
        const response = await factoryApi.getInventoryLogs(factoryId)
        this.inventoryLogs = response.data?.data || response.data || []
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch inventory logs'
      }
    },

    async fetchAllInventoryLogs() {
      this.loading = true
      try {
        const response = await factoryApi.getAllInventoryLogs()
        this.allInventoryLogs = response.data?.data || response.data || []
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch all inventory logs'
      } finally {
        this.loading = false
      }
    },

    async fetchAllInventory() {
      this.loading = true
      try {
        const response = await factoryApi.getAllInventory()
        this.allInventory = response.data?.data || response.data || []
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch all inventory'
      } finally {
        this.loading = false
      }
    },

    async fetchProductionHistory(factoryId: string) {
      try {
        const response = await factoryApi.getProductionHistory(factoryId)
        this.productionHistory = response.data?.data || response.data || []
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch production history'
      }
    },

    async fetchAllProductionHistory() {
      this.loading = true
      try {
        const response = await factoryApi.getAllProductionHistory()
        this.allProductionHistory = response.data?.data || response.data || []
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch all production history'
      } finally {
        this.loading = false
      }
    },

    async fetchTransferHistory(factoryId: string) {
      try {
        const response = await factoryApi.getTransferHistory(factoryId)
        this.transferHistory = response.data?.data || response.data || []
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch transfer history'
      }
    },

    async fetchAllTransfers() {
      this.loading = true
      try {
        const response = await factoryApi.getAllTransfers()
        this.allTransfers = response.data?.data || response.data || []
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch all transfers'
      } finally {
        this.loading = false
      }
    },

    async fetchDelegatedShipments() {
      this.loading = true
      try {
        const response = await factoryApi.getDelegatedShipments()
        this.delegatedShipments = response.data?.shipments || []
        this.delegatedSummary = response.data?.summary || []
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch delegated shipments'
      } finally {
        this.loading = false
      }
    },

    async logProduction(factoryId: string, data: any) {
      try {
        await factoryApi.logProduction(factoryId, data)
        await this.fetchInventory(factoryId)
        await this.fetchProductionHistory(factoryId)
      } catch (err: any) {
        throw err
      }
    },

    async logBatchProduction(factoryId: string, data: any) {
      try {
        await factoryApi.logBatchProduction(factoryId, data)
        await this.fetchAllInventory()
        await this.fetchAllProductionHistory()
        await this.fetchAllInventoryLogs()
      } catch (err: any) {
        throw err
      }
    },

    async updateProductionLog(factoryId: string, id: string, data: any) {
      try {
        await factoryApi.updateProductionLog(id, data)
        await this.fetchInventory(factoryId)
        await this.fetchProductionHistory(factoryId)
      } catch (err: any) {
        throw err
      }
    },

    async deleteProductionLog(factoryId: string, id: string) {
      try {
        await factoryApi.deleteProductionLog(id)
        await this.fetchInventory(factoryId)
        await this.fetchProductionHistory(factoryId)
      } catch (err: any) {
        throw err
      }
    },

    async adjustStock(factoryId: string, data: any) {
      try {
        await factoryApi.adjustStock(factoryId, data)
        await this.fetchInventory(factoryId)
      } catch (err: any) {
        throw err
      }
    },

    async syncWarehouseStock(factoryId: string) {
      try {
        await factoryApi.syncWarehouseStock(factoryId)
        await this.fetchInventory(factoryId)
      } catch (err: any) {
        throw err
      }
    },

    async syncAllWarehouseStock() {
      try {
        await factoryApi.syncAllWarehouseStock()
        await this.fetchAllInventory()
      } catch (err: any) {
        throw err
      }
    },

    async sendToBranch(factoryId: string, data: any) {
      try {
        await factoryApi.sendToBranch(factoryId, data)
        await this.fetchInventory(factoryId)
        await this.fetchAllTransfers()
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to initiate transfer'
      }
    },

    async shipTransfer(id: string) {
      try {
        await factoryApi.shipTransfer(id)
        await this.fetchAllTransfers()
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to execute shipment'
      }
    },

    async approveTransfer(id: string) {
      try {
        await factoryApi.approveTransfer(id)
        await this.fetchAllTransfers()
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to approve transfer'
      }
    },

    async updateTransfer(id: string, data: any) {
      try {
        await factoryApi.updateTransfer(id, data)
        await this.fetchAllTransfers()
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to update transfer'
      }
    },

    async deleteTransfer(id: string) {
      try {
        await factoryApi.deleteTransfer(id)
        await this.fetchAllTransfers()
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to delete transfer'
      }
    },

    async createTransfer(data: any) {
      try {
        const res = await factoryApi.createTransfer(data)
        await this.fetchAllTransfers()
        return res.data
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to create transfer'
      }
    },
    async fetchFactoryReturns(factoryId: string) {
      try {
        const response = await factoryReturnApi.getByFactory(factoryId)
        this.factoryReturns = response.data
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch factory returns'
      }
    },
    async fetchAllFactoryReturns() {
      try {
        const response = await factoryReturnApi.getByCompany()
        this.factoryReturns = response.data
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch all factory returns'
      }
    },
    async receiveFactoryReturn(id: string, payload?: ReceiveReturnPayload) {
      try {
        await factoryReturnApi.receiveReturn(id, payload)
        await this.fetchAllFactoryReturns()
        await this.fetchAllInventory()
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to receive return'
      }
    },
    async fetchQcSamples(factoryId: string, params?: { status?: string }) {
      this.loading = true
      try {
        const response = await factoryApi.getQcSamples(factoryId, params)
        this.qcSamples = response.data?.data || response.data || []
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch QC samples'
      } finally {
        this.loading = false
      }
    },
    async fetchAllQcSamples(params?: { company_id?: string; status?: string }) {
      this.loading = true
      try {
        const response = await factoryApi.getAllQcSamples(params)
        this.qcSamples = response.data?.data || response.data || []
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch all QC samples'
      } finally {
        this.loading = false
      }
    },
    async requestQcSample(factoryId: string, data: any) {
      try {
        const response = await factoryApi.requestQcSample(factoryId, data)
        await this.fetchQcSamples(factoryId)
        return response.data
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to request QC sample'
      }
    },
    async approveQcSample(id: string, factoryId?: string) {
      try {
        const response = await factoryApi.approveQcSample(id)
        if (factoryId) {
          await this.fetchQcSamples(factoryId)
          await this.fetchInventory(factoryId)
        } else {
          await this.fetchAllQcSamples()
          await this.fetchAllInventory()
        }
        return response.data
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to approve QC sample'
      }
    },
    async rejectQcSample(id: string, reason: string, factoryId?: string) {
      try {
        const response = await factoryApi.rejectQcSample(id, { reason })
        if (factoryId) {
          await this.fetchQcSamples(factoryId)
        } else {
          await this.fetchAllQcSamples()
        }
        return response.data
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to reject QC sample'
      }
    },
  },
})
