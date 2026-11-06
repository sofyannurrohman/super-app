import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import {
  marketplaceApi,
  type MarketplaceStock,
  type MarketplaceOpname,
  type MarketplaceDashboardStats,
  type CreateStockPayload,
  type OpnamePayload,
  type UpdateOpnamePayload,
  type FactoryProduct,
  type MarketplaceOrder,
  type MarketplaceOrderGroup,
  type MarketplaceProduct,
  type MarketplaceOrderFilter,
  type MarketplaceOpnameSummary,
  type PackOrderPayload,
  type MarketplaceSettlementFilter,
  type SettlementSummaryResponse,
  type RestockOrderPayload,
  type BatchSettlementImportPayload,
  type UpdateSettlementOrderPayload,
} from '@/api/marketplace'
import { useMasterDataStore } from '@/stores/masterData'

export const useMarketplaceStore = defineStore('marketplace', () => {
  const masterDataStore = useMasterDataStore()

  // State
  const stocks = ref<MarketplaceStock[]>([])
  const opnames = ref<MarketplaceOpname[]>([])
  const auditTrail = ref<any[]>([])
  const stockMovements = ref<MarketplaceOpname[]>([]) // history pergerakan barang (opnames)
  const movementsTotal = ref(0)
  const movementsPage = ref(1)
  const movementsLimit = ref(20)
  const movementsSummary = ref<MarketplaceOpnameSummary>({
    total_in: 0,
    total_out: 0,
    total_adj: 0,
    total_transactions: 0,
    total_omzet: 0,
  })
  const dashboardStats = ref<MarketplaceDashboardStats>({
    total_sku: 0,
    critical_count: 0,
    today_in: 0,
    today_out: 0,
    today_omzet: 0,
  })
  const isLoading = ref(false)
  const isOpnameLoading = ref(false)
  const error = ref<string | null>(null)
  const lastUpdated = ref<Date | null>(null)
  const factoryProducts = ref<FactoryProduct[]>([])
  
  // Marketplace Orders & Products
  const marketplaceOrders = ref<MarketplaceOrderGroup[]>([])
  const marketplaceOrdersTotal = ref(0)
  const marketplaceOrdersTotalOmzet = ref(0)
  const marketplaceOrdersTotalTiktok = ref(0)
  const marketplaceOrdersTotalShopee = ref(0)
  const marketplaceProducts = ref<MarketplaceProduct[]>([])
  const isOrdersLoading = ref(false)
  const isProductsLoading = ref(false)
  const isSyncing = ref(false)
  const orderFilters = ref({
    warehouses: [] as string[],
    products: [] as string[],
    shipping_providers: [] as string[],
  })
  
  const statsAwaitingCollection = ref(0)
  const statsPacked = ref(0)
  const statsCanceled = ref(0)
  let latestOrdersRequest = 0

  // Marketplace Settlements & Financial Reconciliation
  const settlementOrders = ref<MarketplaceOrderGroup[]>([])
  const settlementOrdersTotal = ref(0)
  const settlementTotalAmount = ref(0)
  const settlementTotalFee = ref(0)
  const settlementTotalRefund = ref(0)
  const settlementSummary = ref<SettlementSummaryResponse>({
    total_settled_amount: 0,
    total_settled_orders: 0,
    total_pending_amount: 0,
    total_pending_orders: 0,
    total_refund_amount: 0,
    total_refund_orders: 0,
    total_platform_fees: 0,
    total_pending_return_items: 0,
    total_restocked_items: 0,
  })
  const isSettlementLoading = ref(false)
  const isSummaryLoading = ref(false)

  // Computed
  const criticalStocks = computed(() =>
    stocks.value.filter((s) => s.quantity <= s.min_stock && s.quantity > 0),
  )
  const emptyStocks = computed(() => stocks.value.filter((s) => s.quantity === 0))
  const safeStocks = computed(() => stocks.value.filter((s) => s.quantity > s.min_stock))

  function getStockStatus(stock: MarketplaceStock): 'safe' | 'critical' | 'empty' {
    if (stock.quantity === 0) return 'empty'
    if (stock.quantity <= stock.min_stock) return 'critical'
    return 'safe'
  }

  // Actions
  async function fetchDashboard(period: string = 'daily', dateFilter?: string) {
    try {
      const res = await marketplaceApi.getDashboard(masterDataStore.selectedBranchId, period, dateFilter)
      dashboardStats.value = res.data
    } catch (e: any) {
      console.error('Failed to fetch marketplace dashboard:', e)
    }
  }

  async function fetchStocks() {
    isLoading.value = true
    error.value = null
    try {
      const res = await marketplaceApi.getStocks(masterDataStore.selectedBranchId)
      stocks.value = res.data
      lastUpdated.value = new Date()
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Gagal memuat data stok'
    } finally {
      isLoading.value = false
    }
  }

  async function fetchFactoryProducts() {
    try {
      const res = await marketplaceApi.getFactoryProducts()
      factoryProducts.value = res.data
    } catch (e) {
      console.error('Failed to fetch factory products:', e)
    }
  }

  async function createStock(data: CreateStockPayload) {
    const res = await marketplaceApi.createStock(data, masterDataStore.selectedBranchId)
    stocks.value.unshift(res.data)
    await fetchDashboard()
    return res.data
  }

  async function updateStock(id: string, data: Partial<CreateStockPayload>) {
    await marketplaceApi.updateStock(id, data)
    await fetchStocks()
  }

  async function deleteStock(id: string) {
    await marketplaceApi.deleteStock(id)
    stocks.value = stocks.value.filter((s) => s.id !== id)
    await fetchDashboard()
  }

  async function doOpname(id: string, data: OpnamePayload) {
    isOpnameLoading.value = true
    try {
      const res = await marketplaceApi.doOpname(id, data)
      await fetchStocks()
      opnames.value.unshift(res.data)
      await fetchDashboard()
      return res.data
    } finally {
      isOpnameLoading.value = false
    }
  }

  async function fetchOpnames(params: {
    stockId?: string
    type?: string
    from?: string
    to?: string
  } = {}) {
    isLoading.value = true
    try {
      const res = await marketplaceApi.getOpnames({
        branchId: masterDataStore.selectedBranchId,
        ...params,
      })
      opnames.value = res.data
    } finally {
      isLoading.value = false
    }
  }


  // Fetch riwayat pergerakan barang (opnames: STOCK_IN, STOCK_OUT, ADJUSTMENT)
  async function fetchStockMovements(params: {
    stockId?: string
    type?: string
    sourceType?: string
    source_type?: string
    from?: string
    to?: string
    search?: string
    page?: number
    limit?: number
  } = {}) {
    isLoading.value = true
    try {
      const res = await marketplaceApi.getStockMovements({
        branchId: masterDataStore.selectedBranchId,
        ...params,
      })
      stockMovements.value = res.data?.data || []
      movementsTotal.value = res.data?.total || 0
      movementsPage.value = res.data?.page || 1
      movementsLimit.value = res.data?.limit || 20
      if (res.data?.summary) {
        movementsSummary.value = res.data.summary
      } else {
        movementsSummary.value = {
          total_in: 0,
          total_out: 0,
          total_adj: 0,
          total_transactions: res.data?.total || 0,
          total_omzet: 0,
        }
      }
    } catch (e: any) {
      console.error('Failed to fetch stock movements:', e)
      stockMovements.value = []
      movementsTotal.value = 0
      movementsSummary.value = {
        total_in: 0,
        total_out: 0,
        total_adj: 0,
        total_transactions: 0,
        total_omzet: 0,
      }
    } finally {
      isLoading.value = false
    }
  }

  async function fetchAuditTrail() {
    isLoading.value = true
    try {
      const res = await marketplaceApi.getAuditTrail(masterDataStore.selectedBranchId)
      auditTrail.value = Array.isArray(res.data) ? res.data : []
    } catch (e: any) {
      console.error('Failed to fetch audit trail:', e)
      auditTrail.value = []
    } finally {
      isLoading.value = false
    }
  }

  async function updateStockMovement(id: string, data: UpdateOpnamePayload) {
    isLoading.value = true
    try {
      const res = await marketplaceApi.updateStockMovement(id, data)
      await fetchStocks()
      return res.data
    } finally {
      isLoading.value = false
    }
  }

  function getActiveBranchId(): string | undefined {
    const bid = masterDataStore.selectedBranchId
    return bid && bid !== 'ALL_BRANCHES' ? bid : undefined
  }

  async function fetchMarketplaceOrders(filter: MarketplaceOrderFilter) {
    const requestId = ++latestOrdersRequest
    isOrdersLoading.value = true
    try {
      const branchId = filter.branch_id !== undefined ? filter.branch_id : getActiveBranchId()
      const mergedFilter = { ...filter, branch_id: branchId }
      const res = await marketplaceApi.getMarketplaceOrders(mergedFilter)
      if (requestId !== latestOrdersRequest) return
      marketplaceOrders.value = res.data.data
      marketplaceOrdersTotal.value = res.data.total
      marketplaceOrdersTotalOmzet.value = res.data.total_omzet || 0
      marketplaceOrdersTotalTiktok.value = res.data.total_tiktok || 0
      marketplaceOrdersTotalShopee.value = res.data.total_shopee || 0
      statsAwaitingCollection.value = res.data.status_stats?.awaiting_collection || 0
      statsPacked.value = res.data.status_stats?.packed || 0
      statsCanceled.value = res.data.status_stats?.canceled || 0
    } catch (e: any) {
      console.error('Failed to fetch marketplace orders:', e)
    } finally {
      if (requestId === latestOrdersRequest) isOrdersLoading.value = false
    }
  }

  async function syncMarketplaceOrders(marketplace?: string) {
    isSyncing.value = true
    try {
      const res = await marketplaceApi.syncMarketplaceOrders(getActiveBranchId(), marketplace)
      return res.data
    } catch (e: any) {
      console.error('Failed to sync marketplace orders:', e)
      throw e
    } finally {
      isSyncing.value = false
    }
  }

  async function fetchSingleMarketplaceOrder(orderId: string, marketplace?: string, asPendingPack = true) {
    isSyncing.value = true
    try {
      const res = await marketplaceApi.fetchSingleOrder({
        order_id: orderId,
        marketplace,
        as_pending_pack: asPendingPack,
        branch_id: getActiveBranchId(),
      })
      return res.data
    } catch (e: any) {
      console.error('Failed to fetch single marketplace order:', e)
      throw e
    } finally {
      isSyncing.value = false
    }
  }

  async function fetchOrderFilters() {
    try {
      const res = await marketplaceApi.getOrderFilters(getActiveBranchId())
      orderFilters.value = res.data
    } catch (e: any) {
      console.error('Failed to fetch order filters:', e)
    }
  }

  async function fetchMarketplaceProducts(status?: string, marketplace?: string) {
    isProductsLoading.value = true
    try {
      const res = await marketplaceApi.getMarketplaceProducts(status, getActiveBranchId(), marketplace)
      marketplaceProducts.value = res.data
    } catch (e: any) {
      console.error('Failed to fetch marketplace products:', e)
    } finally {
      isProductsLoading.value = false
    }
  }

  async function syncMarketplaceProducts(marketplace?: string) {
    isSyncing.value = true
    try {
      const res = await marketplaceApi.syncMarketplaceProducts(getActiveBranchId(), marketplace)
      await fetchMarketplaceProducts(undefined, marketplace)
      return res.data
    } catch (e: any) {
      console.error('Failed to sync marketplace products:', e)
      throw e
    } finally {
      isSyncing.value = false
    }
  }

  async function packMarketplaceOrder(orderId: string, payload?: PackOrderPayload) {
    try {
      const res = await marketplaceApi.packMarketplaceOrder(orderId, payload)
      return res.data
    } catch (e: any) {
      console.error('Failed to pack marketplace order:', e)
      throw e
    }
  }

  async function verifyMarketplaceOrder(orderId: string) {
    try {
      const res = await marketplaceApi.verifyMarketplaceOrder(orderId)
      return res.data
    } catch (e: any) {
      console.error('Failed to verify marketplace order:', e)
      throw e
    }
  }

  async function fetchProductMappings(skuId: string) {
    try {
      const res = await marketplaceApi.fetchProductMappings(skuId, getActiveBranchId())
      return res.data
    } catch (e: any) {
      console.error('Failed to fetch product mappings:', e)
      throw e
    }
  }

  async function saveProductMappings(skuId: string, mappings: any[]) {
    try {
      const res = await marketplaceApi.saveProductMappings(skuId, mappings, getActiveBranchId())
      return res.data
    } catch (e: any) {
      console.error('Failed to save product mappings:', e)
      throw e
    }
  }

  // ───────────────────────────────────────────────
  // Marketplace Settlements & Financial Actions
  // ───────────────────────────────────────────────
  async function fetchSettlementOrders(filter: MarketplaceSettlementFilter) {
    isSettlementLoading.value = true
    try {
      const res = await marketplaceApi.getSettlementOrders({
        ...filter,
        branch_id: getActiveBranchId() || filter.branch_id,
      })
      settlementOrders.value = res.data.data || []
      settlementOrdersTotal.value = res.data.total || 0
      settlementTotalAmount.value = res.data.total_settled_amount || 0
      settlementTotalFee.value = res.data.total_platform_fee || 0
      settlementTotalRefund.value = res.data.total_refund_amount || 0
    } catch (e: any) {
      console.error('Failed to fetch settlement orders:', e)
      settlementOrders.value = []
      settlementOrdersTotal.value = 0
    } finally {
      isSettlementLoading.value = false
    }
  }

  async function fetchSettlementSummary(params?: { from?: string; to?: string; marketplace?: string; date_type?: string }) {
    isSummaryLoading.value = true
    try {
      const res = await marketplaceApi.getSettlementSummary({
        ...params,
        branch_id: getActiveBranchId(),
      })
      if (res.data?.data) {
        settlementSummary.value = res.data.data
      }
    } catch (e: any) {
      console.error('Failed to fetch settlement summary:', e)
    } finally {
      isSummaryLoading.value = false
    }
  }

  async function restockOrder(orderId: string, notes?: string) {
    try {
      const res = await marketplaceApi.restockOrder({ order_id: orderId, notes })
      return res.data
    } catch (e: any) {
      console.error('Failed to restock order:', e)
      throw e
    }
  }

  async function importSettlements(payload: BatchSettlementImportPayload) {
    try {
      const res = await marketplaceApi.importSettlements(payload)
      return res.data
    } catch (e: any) {
      console.error('Failed to import settlements:', e)
      throw e
    }
  }

  async function updateSettlementOrder(orderId: string, payload: UpdateSettlementOrderPayload) {
    try {
      const res = await marketplaceApi.updateSettlementOrder(orderId, payload)
      return res.data
    } catch (e: any) {
      console.error('Failed to update settlement order:', e)
      throw e
    }
  }

  async function autoClassifySettlements() {
    try {
      const res = await marketplaceApi.autoClassifySettlements(masterDataStore.selectedBranchId)
      return res.data
    } catch (e: any) {
      console.error('Failed to auto-classify settlements:', e)
      throw e
    }
  }

  async function syncFinanceTransactions(days: number = 30, marketplace: 'tiktok' | 'shopee' = 'tiktok') {
    try {
      const res = await marketplaceApi.syncFinanceTransactions({
        branchId: masterDataStore.selectedBranchId,
        days,
        marketplace,
      })
      return res.data
    } catch (e: any) {
      console.error('Failed to sync finance transactions:', e)
      throw e
    }
  }

  async function getFinanceSyncStatus(syncId: string) {
    const res = await marketplaceApi.getFinanceSyncStatus(syncId)
    return res.data
  }

  async function syncOrderFinance(orderId: string, marketplace: string = 'tiktok') {
    try {
      const res = await marketplaceApi.syncOrderFinance(orderId, masterDataStore.selectedBranchId, marketplace)
      return res.data
    } catch (e: any) {
      console.error('Failed to sync order finance:', e)
      throw e
    }
  }

  return {
    stocks,
    opnames,
    auditTrail,
    stockMovements,
    dashboardStats,
    factoryProducts,
    isLoading,
    isOpnameLoading,
    error,
    lastUpdated,
    isProductsLoading,
    isSyncing,
    orderFilters,
    criticalStocks,
    emptyStocks,
    safeStocks,
    getStockStatus,
    fetchDashboard,
    fetchStocks,
    fetchFactoryProducts,
    createStock,
    updateStock,
    deleteStock,
    doOpname,

    fetchOpnames,
    fetchAuditTrail,
    fetchStockMovements,
    updateStockMovement,
    // Additions:
    marketplaceOrders,
    marketplaceOrdersTotal,
    marketplaceOrdersTotalOmzet,
    marketplaceOrdersTotalTiktok,
    marketplaceOrdersTotalShopee,
    marketplaceProducts,
    isOrdersLoading,
    fetchMarketplaceOrders,
    syncMarketplaceOrders,
    fetchSingleMarketplaceOrder,
    fetchMarketplaceProducts,
    syncMarketplaceProducts,
    packMarketplaceOrder,
    verifyMarketplaceOrder,
    fetchOrderFilters,
    fetchProductMappings,
    saveProductMappings,
    movementsTotal,
    movementsPage,
    movementsLimit,
    movementsSummary,
    statsAwaitingCollection,
    statsPacked,
    statsCanceled,

    // Settlements & Finance:
    settlementOrders,
    settlementOrdersTotal,
    settlementTotalAmount,
    settlementTotalFee,
    settlementTotalRefund,
    settlementSummary,
    isSettlementLoading,
    isSummaryLoading,
    fetchSettlementOrders,
    fetchSettlementSummary,
    restockOrder,
    importSettlements,
    updateSettlementOrder,
    autoClassifySettlements,
    syncFinanceTransactions,
    getFinanceSyncStatus,
    syncOrderFinance,
  }
})


