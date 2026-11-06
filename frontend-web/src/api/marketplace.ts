import apiClient from './axios'

export interface MarketplaceOrder {
  id: number
  branch_id?: string
  order_id: string
  order_status: string
  order_substatus: string
  cancelation_return_type: string
  sku_id: string
  seller_sku: string
  product_name: string
  warehouse_name: string
  tracking_id: string
  marketplace: string
  buyer_username: string
  payment_method: string
  total_amount: number
  currency: string
  item_count: number
  shipping_provider: string
  create_time_raw: number
  picked_by?: string
  packed_by?: string
  // Settlement & Finance
  settlement_status?: string
  settlement_date?: string
  estimated_settlement_date?: string
  settlement_amount?: number
  platform_fee?: number
  commission_fee?: number
  shipping_fee_deduction?: number
  refund_amount?: number
  refund_reason?: string
  return_status?: string
  return_item_count?: number
  restocked_at?: string
  restocked_by?: string
  opname_id?: string
  created_at: string
  updated_at: string
}

export interface MarketplaceProduct {
  id: number
  branch_id?: string
  marketplace: string
  product_id: string
  product_name: string
  sku_id: string
  seller_sku: string
  price_amount: number
  price_currency: string
  stock: number
  status: string
  category: string
  sales_regions: string
  last_synced_at: string
  created_at: string
  updated_at: string
}

export interface MarketplaceOrderGroup {
  order_id: string
  branch_id?: string
  order_status: string
  warehouse_name: string
  tracking_id: string
  marketplace: string
  buyer_username: string
  payment_method: string
  total_amount: number
  currency: string
  shipping_provider: string
  picked_by?: string
  packed_by?: string
  is_packed?: boolean
  packing_status?: 'PACKED' | 'UNPACKED' | string
  reconciliation_status?: 'MATCH_SETTLED' | 'IN_TRANSIT_HOLD' | 'RETURN_DISPATCHED' | 'RETURN_UNPACKED' | 'RESTOCKED' | 'ANOMALY_SETTLED_UNPACKED' | string
  // Settlement & Finance
  settlement_status?: string
  settlement_date?: string
  estimated_settlement_date?: string
  settlement_amount?: number
  platform_fee?: number
  commission_fee?: number
  fee_and_tax_amount?: number
  shipping_cost_amount?: number
  adjustment_amount?: number
  fee_tax_breakdown?: string
  hold_reason?: string
  shipping_fee_deduction?: number
  refund_amount?: number
  refund_reason?: string
  return_status?: string
  return_item_count?: number
  restocked_at?: string
  restocked_by?: string
  finance_synced_at?: string
  created_at: string
  items: MarketplaceOrder[]
}

export interface MarketplaceOrdersResponse {
  data: MarketplaceOrderGroup[]
  total: number
  total_omzet?: number
  total_tiktok?: number
  total_shopee?: number
  status_stats?: {
    awaiting_collection: number
    packed: number
    canceled: number
  }
  page: number
  page_size: number
}

export interface MarketplaceSettlementsResponse {
  data: MarketplaceOrderGroup[]
  total: number
  total_settled_amount: number
  total_platform_fee: number
  total_refund_amount: number
}

export interface SettlementSummaryResponse {
  total_settled_amount: number
  total_settled_orders: number
  total_pending_amount: number
  total_pending_orders: number
  total_refund_amount: number
  total_refund_orders: number
  total_platform_fees: number
  total_pending_return_items: number
  total_restocked_items: number
  total_dispatched_return_orders?: number
  total_dispatched_return_value?: number
  total_restocked_return_orders?: number
  total_restocked_return_value?: number
}

export interface UnsettledSnapshot {
  id: number
  marketplace: string
  shop_id: string
  branch_id?: string
  captured_at: string
  currency: string
  awaiting_delivery: number
  after_sales: number
  processing: number
  source: 'SELLER_CENTER' | 'TIKTOK_API'
  transaction_count: number
  unclassified_amount: number
}

export interface UnsettledTransaction {
  id: number
  snapshot_id: number
  transaction_id: string
  order_id: string
  adjustment_id: string
  transaction_type: string
  unsettled_reason: string
  category: string
  estimated_amount: number
  estimated_settlement: string
}

export interface MarketplaceSettlementFilter {
  branch_id?: string
  marketplace?: string
  tab?: 'ALL' | 'SETTLED' | 'HOLD' | 'REFUND_RETURN' | string
  settlement_status?: string
  return_status?: string
  packing_status?: 'ALL' | 'PACKED' | 'UNPACKED' | 'NEEDS_RESTOCK' | 'RESTOCKED' | string
  keyword?: string
  product_name?: string
  warehouse_name?: string
  shipping_provider?: string
  date_type?: 'created_at' | 'settlement_date' | string
  from?: string
  to?: string
  page?: number
  pageSize?: number
}

export interface RestockOrderPayload {
  order_id: string
  notes?: string
}

export interface SettlementImportRecord {
  order_id: string
  settlement_date: string
  settlement_amount: number
  platform_fee: number
  commission_fee: number
  status?: string
}

export interface BatchSettlementImportPayload {
  marketplace?: string
  records: SettlementImportRecord[]
}

export interface UpdateSettlementOrderPayload {
  settlement_status?: string
  settlement_date?: string
  settlement_amount?: number
  platform_fee?: number
  commission_fee?: number
  refund_amount?: number
  refund_reason?: string
  return_status?: string
}

export interface MarketplaceOrderFilter {
  branch_id?: string
  marketplace?: string
  status?: string
  keyword?: string
  product_name?: string
  warehouse_name?: string
  shipping_provider?: string
  from?: string
  to?: string
  page?: number
  pageSize?: number
}

export interface PackOrderPayload {
  box_stock_id?: string
  box_qty?: number
  additional_items?: { stock_id: string; qty: number }[]
  skip_default_box?: boolean
}

export interface MarketplaceStock {
  id: string
  branch_id: string
  product_id?: string | null
  product_name: string
  sku: string
  quantity: number
  min_stock: number
  price: number
  notes: string
  is_active: boolean
  image_url?: string
  created_by: string
  created_at: string
  updated_at: string
  product?: FactoryProduct | null
}

export interface FactoryProduct {
  id: string
  name: string
  sku: string
  category: string
  brand: string
  selling_price: number
  factory_price: number
  marketplace_factory_price: number
  cost_price: number
  unit: string
  image_url?: string
}

export interface MarketplaceOpname {
  id: string
  marketplace_stock_id: string
  type: 'STOCK_IN' | 'STOCK_OUT' | 'ADJUSTMENT'
  quantity: number
  quantity_before: number
  quantity_after: number
  price: number
  total_value: number
  reference: string
  source_type?: string
  marketplace_order_id?: string
  notes: string
  created_by: string
  created_at: string
  marketplace_stock?: MarketplaceStock
}

export interface MarketplaceOpnameSummary {
  total_in: number
  total_out: number
  total_adj: number
  total_transactions: number
  total_omzet?: number
}

export interface MarketplaceDashboardStats {
  total_sku: number
  critical_count: number
  today_in: number
  today_out: number
  today_omzet: number
  chart_data?: { date: string, omzet: number }[]
}

export interface CreateStockPayload {
  product_id?: string | null
  product_name: string
  sku: string
  quantity: number
  min_stock: number
  price: number
  notes?: string
  image?: File | null
}

export interface OpnamePayload {
  type: 'STOCK_IN' | 'STOCK_OUT' | 'ADJUSTMENT'
  quantity: number
  reference?: string
  notes?: string
  date?: string
}

export interface UpdateOpnamePayload {
  date?: string
  quantity?: number
  reference?: string
  notes?: string
}

export const marketplaceApi = {
  // Dashboard
  getDashboard: (branchId?: string | null, period?: string, dateFilter?: string) =>
    apiClient.get<MarketplaceDashboardStats>('/marketplace/dashboard', {
      params: { branch_id: branchId, period, date: dateFilter },
    }),

  // Factory Products (for linking)
  getFactoryProducts: () =>
    apiClient.get<FactoryProduct[]>('/marketplace/factory-products'),

  // Stock CRUD
  getStocks: (branchId?: string | null) =>
    apiClient.get<MarketplaceStock[]>('/marketplace/stocks', {
      params: { branch_id: branchId },
    }),

  getStock: (id: string) =>
    apiClient.get<MarketplaceStock>(`/marketplace/stocks/${id}`),

  createStock: (data: CreateStockPayload, branchId?: string | null) => {
    if (data.image) {
      const formData = new FormData()
      if (data.product_id) formData.append('product_id', data.product_id)
      formData.append('product_name', data.product_name)
      formData.append('sku', data.sku)
      formData.append('quantity', data.quantity.toString())
      formData.append('min_stock', data.min_stock.toString())
      formData.append('price', data.price.toString())
      if (data.notes) formData.append('notes', data.notes)
      formData.append('image', data.image)

      return apiClient.post<MarketplaceStock>('/marketplace/stocks', formData, {
        params: { branch_id: branchId },
        headers: { 'Content-Type': 'multipart/form-data' },
      })
    }
    return apiClient.post<MarketplaceStock>('/marketplace/stocks', data, {
      params: { branch_id: branchId },
    })
  },

  updateStock: (id: string, data: Partial<CreateStockPayload>) => {
    if (data.image) {
      const formData = new FormData()
      if (data.product_id !== undefined) formData.append('product_id', data.product_id || '')
      if (data.product_name) formData.append('product_name', data.product_name)
      if (data.sku) formData.append('sku', data.sku)
      if (data.min_stock !== undefined) formData.append('min_stock', data.min_stock.toString())
      if (data.price !== undefined) formData.append('price', data.price.toString())
      if (data.notes !== undefined) formData.append('notes', data.notes)
      formData.append('image', data.image)

      return apiClient.put(`/marketplace/stocks/${id}`, formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
    }
    return apiClient.put(`/marketplace/stocks/${id}`, data)
  },

  deleteStock: (id: string) =>
    apiClient.delete(`/marketplace/stocks/${id}`),

  // Opname & Packing
  doOpname: (id: string, data: OpnamePayload) =>
    apiClient.post<MarketplaceOpname>(`/marketplace/stocks/${id}/opname`, data),
    

  getOpnames: (params: {
    branchId?: string | null
    stockId?: string
    type?: string
    from?: string
    to?: string
  }) =>
    apiClient.get<MarketplaceOpname[]>('/marketplace/opnames', {
      params: {
        branch_id: params.branchId,
        stock_id: params.stockId,
        type: params.type,
        from: params.from,
        to: params.to,
      },
    }),

  // Audit Trail - pakai endpoint marketplace yang benar
  getAuditTrail: (branchId?: string | null, entity = 'MarketplaceStock') =>
    apiClient.get('/marketplace/audit-logs', { params: { branch_id: branchId, entity } }),

  // Stock Movements - ambil data opnames (barang masuk/keluar/opname)
  getStockMovements: (params: {
    branchId?: string | null
    stockId?: string
    type?: string
    sourceType?: string
    source_type?: string
    from?: string
    to?: string
    search?: string
    page?: number
    limit?: number
  }) =>
    apiClient.get<{ data: MarketplaceOpname[], total: number, page: number, limit: number, summary?: MarketplaceOpnameSummary }>('/marketplace/opnames', {
      params: {
        branch_id: params.branchId,
        stock_id: params.stockId,
        type: params.type,
        source_type: params.source_type || params.sourceType,
        from: params.from,
        to: params.to,
        search: params.search,
        page: params.page,
        limit: params.limit,
      },
    }),

  updateStockMovement: (id: string, data: UpdateOpnamePayload) =>
    apiClient.put<{ message: string; data: MarketplaceOpname }>(`/marketplace/opnames/${id}`, data),

  // ───────────────────────────────────────────────
  // Marketplace Orders (TikTok)
  // ───────────────────────────────────────────────
  getMarketplaceOrders: (filter: MarketplaceOrderFilter) =>
    apiClient.get<MarketplaceOrdersResponse>('/marketplace/orders', {
      params: {
        branch_id: filter.branch_id || undefined,
        marketplace: filter.marketplace || undefined,
        status: filter.status || undefined,
        keyword: filter.keyword || undefined,
        product_name: filter.product_name || undefined,
        warehouse_name: filter.warehouse_name || undefined,
        shipping_provider: filter.shipping_provider || undefined,
        from: filter.from || undefined,
        to: filter.to || undefined,
        page: filter.page ?? 1,
        page_size: filter.pageSize ?? 20,
      },
    }),

  syncMarketplaceOrders: (branchId?: string | null, marketplace?: string) =>
    apiClient.post<{ message: string; synced_count: number }>('/marketplace/sync-orders', {}, {
      params: { branch_id: branchId || undefined, marketplace: marketplace || undefined },
      timeout: 120000,
    }),

  fetchSingleOrder: (payload: { order_id: string; marketplace?: string; as_pending_pack?: boolean; branch_id?: string | null }) =>
    apiClient.post<{ message: string; synced_count: number; order_id: string; marketplace: string }>(
      '/marketplace/orders/fetch-single',
      {
        order_id: payload.order_id,
        marketplace: payload.marketplace,
        as_pending_pack: payload.as_pending_pack ?? true,
      },
      {
        params: { branch_id: payload.branch_id || undefined },
      }
    ),

  getOrderFilters: (branchId?: string | null) =>
    apiClient.get<{ warehouses: string[]; products: string[]; shipping_providers: string[] }>('/marketplace/orders/filters', {
      params: { branch_id: branchId || undefined },
    }),

  // ───────────────────────────────────────────────
  // Marketplace Products / SKU (TikTok & Shopee)
  // ───────────────────────────────────────────────
  getMarketplaceProducts: (status?: string, branchId?: string | null, marketplace?: string) =>
    apiClient.get<MarketplaceProduct[]>('/marketplace/mp-products', {
      params: { status: status || undefined, branch_id: branchId || undefined, marketplace: marketplace || undefined },
    }),

  syncMarketplaceProducts: (branchId?: string | null, marketplace?: string) =>
    apiClient.post<{ message: string; synced_count: number }>('/marketplace/sync-products', {}, {
      params: { branch_id: branchId || undefined, marketplace: marketplace || undefined },
      timeout: 120000,
    }),
  packMarketplaceOrder: (orderId: string, payload?: PackOrderPayload) =>
    apiClient.post<{ message: string }>(`/marketplace/pack/${orderId}`, payload ?? {}),
  verifyMarketplaceOrder: (orderId: string) =>
    apiClient.post<{ message: string }>(`/marketplace/verify/${orderId}`),
  fetchProductMappings: (skuId: string, branchId?: string | null) =>
    apiClient.get<any[]>(`/marketplace/mappings`, { params: { sku_id: skuId, branch_id: branchId || undefined } }),
  saveProductMappings: (skuId: string, mappings: any[], branchId?: string | null) =>
    apiClient.post<{ message: string }>(`/marketplace/mappings`, { mappings }, { params: { sku_id: skuId, branch_id: branchId || undefined } }),

  // ───────────────────────────────────────────────
  // Marketplace Settlements & Financial Reconciliation
  // ───────────────────────────────────────────────
  getSettlementOrders: (filter: MarketplaceSettlementFilter) =>
    apiClient.get<MarketplaceSettlementsResponse>('/marketplace/settlements', {
      params: {
        branch_id: filter.branch_id || undefined,
        marketplace: filter.marketplace || undefined,
        tab: filter.tab || undefined,
        settlement_status: filter.settlement_status || undefined,
        return_status: filter.return_status || undefined,
        packing_status: filter.packing_status || undefined,
        keyword: filter.keyword || undefined,
        product_name: filter.product_name || undefined,
        warehouse_name: filter.warehouse_name || undefined,
        shipping_provider: filter.shipping_provider || undefined,
        date_type: filter.date_type || undefined,
        from: filter.from || undefined,
        to: filter.to || undefined,
        page: filter.page ?? 1,
        page_size: filter.pageSize ?? 20,
      },
    }),

  getSettlementSummary: (params?: { branch_id?: string | null; from?: string; to?: string; marketplace?: string; date_type?: string }) =>
    apiClient.get<{ data: SettlementSummaryResponse }>('/marketplace/settlements/summary', {
      params: {
        branch_id: params?.branch_id || undefined,
        from: params?.from || undefined,
        to: params?.to || undefined,
        marketplace: params?.marketplace || undefined,
        date_type: params?.date_type || undefined,
      },
    }),

  getUnsettledSnapshots: (branchId?: string | null, shopId?: string) =>
    apiClient.get<{ shops: string[]; snapshots: UnsettledSnapshot[] }>('/marketplace/settlements/unsettled-snapshots', {
      params: { branch_id: branchId || undefined, shop_id: shopId || undefined },
    }),

  getUnsettledSnapshot: (id: number, branchId?: string | null) =>
    apiClient.get<{ snapshot: UnsettledSnapshot; transactions: UnsettledTransaction[] }>(`/marketplace/settlements/unsettled-snapshots/${id}`, {
      params: { branch_id: branchId || undefined },
    }),

  captureUnsettledSnapshot: (payload: { shop_id: string; source: 'SELLER_CENTER' | 'TIKTOK_API'; awaiting_delivery?: number; after_sales?: number; processing?: number }, branchId?: string | null) =>
    apiClient.post<{ snapshot?: UnsettledSnapshot; status?: string; job_id?: string }>('/marketplace/settlements/unsettled-snapshots', payload, {
      params: { branch_id: branchId || undefined },
      timeout: 300000,
    }),

  getUnsettledJob: (jobId: string) =>
    apiClient.get<{ status: 'processing' | 'completed' | 'failed'; snapshot_id?: number; error?: string }>(`/marketplace/settlements/unsettled-jobs/${jobId}`),

  restockOrder: (payload: RestockOrderPayload) =>
    apiClient.post<{ message: string; order_id: string }>('/marketplace/settlements/restock', payload),

  importSettlements: (payload: BatchSettlementImportPayload) =>
    apiClient.post<{ message: string; updated_count: number }>('/marketplace/settlements/import', payload),

  updateSettlementOrder: (orderId: string, payload: UpdateSettlementOrderPayload) =>
    apiClient.put<{ message: string }>(`/marketplace/settlements/${orderId}`, payload),

  autoClassifySettlements: (branchId?: string | null) =>
    apiClient.post<{ message: string; updated_count: number }>('/marketplace/settlements/auto-classify', {
      branch_id: branchId || undefined,
    }),

  syncFinanceTransactions: (params?: { branchId?: string | null; days?: number; marketplace?: 'tiktok' | 'shopee' }) =>
    apiClient.post<{ message: string; synced_count?: number; statements_count?: number; in_background?: boolean; sync_id?: string }>(
      params?.marketplace === 'shopee' ? '/marketplace/settlements/sync-shopee-finance' : '/marketplace/settlements/sync-finance',
      null,
      {
        params: { branch_id: params?.branchId || undefined, days: params?.days ?? 30 },
        timeout: 120000,
      }
    ),

  getFinanceSyncStatus: (syncId: string) =>
    apiClient.get<{ status: 'processing' | 'completed' | 'failed'; synced_count: number; statements_count: number; paid_statements_count?: number; settled_statements_count?: number; latest_payment_at?: string; processed_count: number; phase?: string; platform?: string; updated_at?: string; error?: string }>(
      `/marketplace/settlements/sync-finance/${syncId}`,
    ),

  syncOrderFinance: (orderId: string, branchId?: string | null, marketplace?: string) =>
    apiClient.post<{
      message: string
      order_id: string
      settlement_amount: number
      settlement_amount_provided?: boolean
      fee_and_tax_amount: number
      fee_and_tax_provided?: boolean
      shipping_cost_amount: number
      settlement_status: string
      fee_tax_breakdown: string
    }>(`/marketplace/settlements/${marketplace === 'shopee' ? 'sync-shopee-order-finance' : 'sync-order-finance'}/${encodeURIComponent(orderId)}`, null, {
      params: { branch_id: branchId || undefined },
      timeout: 60000,
    }),

  // ───────────────────────────────────────────────
  // Marketplace Accounts / Integration per Branch
  // ───────────────────────────────────────────────
  getMarketplaceAccounts: () =>
    apiClient.get<BranchMarketplaceAccount[]>('/marketplaces/accounts'),
  getTikTokAuthUrl: (branchId: string) =>
    apiClient.get<{ auth_url: string; app_key: string; state: string }>('/marketplaces/tiktok/auth-url', {
      params: { branch_id: branchId },
    }),
  getShopeeAuthUrl: (branchId: string) =>
    apiClient.get<{ auth_url: string; partner_id: string; branch_id: string }>('/marketplaces/shopee/auth-url', {
      params: { branch_id: branchId },
    }),
  saveMarketplaceAccount: (payload: SaveMarketplaceAccountPayload) =>
    apiClient.post<{ message: string; data: any }>('/marketplaces/accounts', payload),
  deleteMarketplaceAccount: (id: number) =>
    apiClient.delete<{ message: string }>(`/marketplaces/accounts/${id}`),
  testTikTokConnection: (branchId: string) =>
    apiClient.post<{ success: boolean; message?: string; error?: string; shop_id?: string; expires_at?: string; warehouse_count?: number }>(
      '/marketplaces/tiktok/test-connection',
      { branch_id: branchId }
    ),
  testShopeeConnection: (branchId: string) =>
    apiClient.post<{ success: boolean; message?: string; error?: string; shop_id?: string; expires_at?: string; product_count?: number }>(
      '/marketplaces/shopee/test-connection',
      { branch_id: branchId }
    ),
  getShopeeConfig: () =>
    apiClient.get<ShopeeConfigResponse>('/marketplaces/shopee/config'),
  saveShopeeConfig: (payload: SaveShopeeConfigPayload) =>
    apiClient.post<{ message: string; partner_id: string; environment: string }>('/marketplaces/shopee/config', payload),
}

export interface ShopeeConfigResponse {
  partner_id: string
  partner_key: string
  has_key: boolean
  environment: 'production' | 'sandbox' | string
  base_url: string
  is_configured: boolean
}

export interface SaveShopeeConfigPayload {
  partner_id: string
  partner_key?: string
  environment: 'production' | 'sandbox' | string
}

export interface MarketplaceAccountDetail {
  id: number
  marketplace: string
  shop_id: string
  is_connected: boolean
  is_expired: boolean
  expires_at?: string
  has_refresh_token: boolean
  updated_at: string
}

export interface BranchMarketplaceAccount {
  branch_id: string
  branch_name: string
  branch_address?: string
  accounts: MarketplaceAccountDetail[]
}

export interface SaveMarketplaceAccountPayload {
  id?: number
  branch_id: string
  marketplace: string
  shop_id: string
  access_token: string
  refresh_token?: string
  expires_at?: string
}


