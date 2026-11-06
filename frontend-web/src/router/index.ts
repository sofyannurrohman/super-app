// src/router/index.ts
import { createRouter, createWebHistory } from 'vue-router'
import DashboardLayout from '@/layouts/DashboardLayout.vue'
import { useAuthStore } from '@/stores/auth'
import { useLayoutStore } from '@/stores/layout'
import { resetLoadingBar } from '@/api/axios'
import { isModuleEnabled, isPkwtEnabled, moduleForPath } from '@/config/modules'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/views/RegisterView.vue'),
    },
    {
      path: '/invoice/:receiptNo(.*)',
      name: 'publicInvoice',
      component: () => import('@/views/sales/PublicInvoiceView.vue'),
    },
    {
      path: '/public/tracking',
      name: 'publicTracking',
      component: () => import('@/views/sales/PublicTrackingView.vue'),
    },
    {
      path: '/kehadiran',
      name: 'employeePortal',
      component: () => import('@/views/EmployeePortalView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/lps-print',
      name: 'lpsPrintStandalone',
      component: () => import('@/views/sales/LPSView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/absensi/pekerja-harian',
      alias: ['/kiosk-absensi', '/kiosk'],
      name: 'kioskAbsensi',
      component: () => import('@/views/attendance/TabletKioskPage.vue'),
    },
    {
      path: '/',
      component: DashboardLayout,
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'home',
          component: () => import('@/views/DashboardView.vue'),
        },
        {
          path: 'bi-dashboard',
          name: 'bi-dashboard',
          component: () => import('@/views/BIDashboardView.vue'),
          meta: { requiresAuth: true }
        },
        {
          path: 'employees',
          name: 'employees',
          component: () => import('@/views/EmployeeView.vue'),
        },
        {
          path: 'daily-employees',
          name: 'dailyEmployees',
          component: () => import('@/views/employee/DailyEmployeePage.vue'),
        },
        {
          path: 'karyawan-bermasalah',
          name: 'problematicEmployees',
          component: () => import('@/views/ProblematicEmployeeView.vue'),
        },
        {
          path: 'attendance',
          name: 'attendance',
          component: () => import('@/views/AttendanceView.vue'),
        },
        {
          path: 'attendance-daily',
          name: 'attendance-daily',
          component: () => import('@/views/attendance/DailyWorkerAttendanceView.vue'),
        },
        {
          path: 'leaves',
          name: 'leaves',
          component: () => import('@/views/LeaveView.vue'),
        },
        {
          path: 'payroll',
          name: 'payroll',
          component: () => import('@/views/PayrollView.vue'),
        },
        {
          path: 'overtimes',
          name: 'overtimes',
          component: () => import('@/views/OvertimeView.vue'),
        },
        {
          path: 'job-positions',
          name: 'jobPositions',
          component: () => import('@/views/JobPositionView.vue'),
        },
        {
          path: 'leave-types',
          name: 'leaveTypes',
          component: () => import('@/views/LeaveTypeView.vue'),
        },
        {
          path: 'shifts',
          name: 'shifts',
          component: () => import('@/views/ShiftView.vue'),
        },
        {
          path: 'users',
          name: 'users',
          component: () => import('@/views/UserView.vue'),
        },
        {
          path: 'active-users',
          name: 'activeUsers',
          component: () => import('@/views/ActiveUsersView.vue'),
        },
        {
          path: 'access-roles',
          name: 'accessRoles',
          component: () => import('@/views/AccessRoleView.vue'),
        },
        {
          path: 'system-modules',
          name: 'systemModules',
          component: () => import('@/views/SystemModuleView.vue'),
        },
        {
          path: 'organization',
          name: 'organization',
          component: () => import('@/views/OrganizationView.vue'),
        },
        {
          path: 'leave-requests',
          name: 'leaveRequests',
          component: () => import('@/views/LeaveRequestView.vue'),
        },
        {
          path: 'reimbursements',
          name: 'reimbursements',
          component: () => import('@/views/ReimbursementView.vue'),
        },
        {
          path: 'performance-kpi',
          name: 'performanceKpi',
          component: () => import('@/views/PerformanceView.vue'),
        },
        {
          path: 'roles',
          name: 'roles',
          component: () => import('@/views/RoleView.vue'),
        },
        {
          path: 'audit-logs',
          name: 'auditLogs',
          component: () => import('@/views/AuditLogView.vue'),
        },
        {
          path: 'announcements',
          name: 'announcements',
          component: () => import('@/views/AnnouncementView.vue'),
        },
        {
          path: 'memo-requests',
          name: 'memoRequests',
          component: () => import('@/views/MemoRequestView.vue'),
        },
        {
          path: 'employee-requests',
          name: 'employeeRequests',
          component: () => import('@/views/EmployeeRequestView.vue'),
        },
        {
          path: 'pkwt',
          name: 'pkwt',
          component: () => import('@/views/PKWTView.vue'),
        },
        {
          path: 'vehicles',
          name: 'vehicles',
          component: () => import('@/views/VehicleView.vue'),
          meta: { title: 'Manajemen Kendaraan', requiresAuth: true }
        },
        {
          path: 'vehicles/live-tracking',
          name: 'live-tracking',
          component: () => import('@/views/LiveTrackingView.vue'),
          meta: { title: 'Live Tracking Armada', requiresAuth: true }
        },
        {
          path: 'branding-integration',
          redirect: '/sales/branding-integration',
        },
        {
          path: 'komisi-operator',
          name: 'komisiOperator',
          component: () => import('@/views/BrandingIntegrationView.vue'),
          meta: { title: 'Komisi Operator Spanduk' },
        },
        { path: 'analytics', redirect: '/' },
        // Sales & Marketing Module
        {
          path: 'sales',
          children: [
            {
              path: '',
              name: 'salesDashboard',
              component: () => import('@/views/sales/SalesDashboard.vue'),
            },
            {
              path: 'salesman',
              name: 'salesman',
              component: () => import('@/views/sales/SalesmanView.vue'),
            },
            {
              path: 'operator',
              name: 'operator',
              component: () => import('@/views/sales/OperatorView.vue'),
            },
            {
              path: 'editor',
              name: 'editor',
              component: () => import('@/views/sales/EditorView.vue'),
            },
            {
              path: 'pesanan',
              name: 'salesOrders',
              component: () => import('@/views/sales/SalesOrderView.vue'),
            },
            {
              path: 'discounts',
              name: 'discountRules',
              component: () => import('@/views/sales/DiscountRulesView.vue'),
            },

            {
              path: 'delivery',
              name: 'deliveryManagement',
              component: () => import('@/views/sales/LogisticsDashboard.vue'),
            },
            {
              path: 'surat-jalan',
              name: 'suratJalan',
              component: () => import('@/views/sales/SuratJalanView.vue'),
            },
            {
              path: 'penjualan',
              name: 'penjualan',
              component: () => import('@/views/sales/PenjualanView.vue'),
            },
            {
              path: 'penerimaan',
              name: 'penerimaan',
              component: () => import('@/views/sales/PenerimaanView.vue'),
            },
            {
              path: 'qc-samples',
              name: 'salesQcSamples',
              component: () => import('@/views/factory/FactoryQcSampleView.vue'),
            },
            {
              path: 'pindah-barang',
              alias: 'retur_pindah',
              name: 'salesTransfers',
              component: () => import('@/views/warehouse/SalesTransferView.vue'),
            },
            {
              path: 'spv-approval',
              name: 'supervisorApproval',
              component: () => import('@/views/sales/SupervisorApprovalView.vue'),
            },
            {
              path: 'spanduk',
              name: 'spanduk',
              component: () => import('@/views/sales/SpandukView.vue'),
            },
            {
              path: 'spanduk-editor',
              name: 'spandukEditor',
              component: () => import('@/views/sales/SpandukEditorView.vue'),
            },
            {
              path: 'spanduk-validator',
              name: 'spanduk-validator',
              component: () => import('@/views/sales/SpandukValidatorView.vue'),
            },
            {
              path: 'spanduk-vendor',
              name: 'spanduk-vendor',
              component: () => import('@/views/sales/SpandukVendorView.vue'),
            },
            {
              path: 'vendor',
              name: 'vendorManagement',
              component: () => import('@/views/sales/VendorView.vue'),
            },
            {
              path: 'spanduk-pemasangan',
              name: 'spanduk-pemasangan',
              component: () => import('@/views/sales/SpandukPemasanganView.vue'),
            },
            {
              path: 'kunjungan',
              name: 'salesVisits',
              component: () => import('@/views/sales/VisitManagementView.vue'),
            },
            {
              path: 'lps',
              name: 'lpsReport',
              component: () => import('@/views/sales/LPSView.vue'),
            },
            {
              path: 'proyeksi-omzet',
              name: 'proyeksiOmzet',
              component: () => import('@/views/sales/ProyeksiOmzetView.vue'),
            },
            {
              path: 'monitoring',
              name: 'monitoring',
              component: () => import('@/views/sales/Live-Monitoring.vue'),
            },
            {
              path: 'kpi',
              name: 'kpi',
              component: () => import('@/views/sales/KPI-Manager.vue'),
            },
            {
              path: 'tracking-noo-ec',
              name: 'salesTrackingNooEc',
              component: () => import('@/views/sales/TrackingNooEcView.vue'),
            },
            {
              path: 'laporan-kinerja-harian',
              name: 'salesDailyKpiReport',
              component: () => import('@/views/sales/DailyKpiReportView.vue'),
              meta: { title: 'Laporan Kinerja Harian', requiresAuth: true },
            },
            {
              path: 'branding-kpi',
              name: 'brandingKpi',
              component: () => import('@/views/sales/BrandingKPIManager.vue'),
            },
            {
              path: 'video-kampanye',
              name: 'brandingCampaignVideos',
              component: () => import('@/views/sales/BrandingCampaignVideoView.vue'),
              meta: { title: 'Video Kampanye', requiresAuth: true },
            },
            {
              path: 'branding-integration',
              name: 'brandingIntegration',
              component: () => import('@/views/BrandingIntegrationView.vue'),
            },
            {
              path: 'komisi-operator',
              redirect: '/sales/branding-integration',
            },
            {
              path: 'customers',
              name: 'customers',
              component: () => import('@/views/sales/StoreView.vue'),
            },
            {
              path: 'transfer-toko',
              name: 'transferToko',
              component: () => import('@/views/sales/TransferStoreView.vue'),
            },
            {
              path: 'audit-logs',
              name: 'salesAuditLogs',
              component: () => import('@/views/sales/AuditLogView.vue'),
            },
            {
              path: 'distributor-portal',
              alias: ['distributor-sales', 'distributor'],
              name: 'distributorPortal',
              component: () => import('@/views/sales/DistributorPortalView.vue'),
            },
            {
              path: 'verification',
              name: 'cashierVerification',
              component: () => import('@/views/warehouse/CashierVerificationView.vue'),
            },
            {
              path: 'modern-trade/kunjungan',
              name: 'mtSalesVisits',
              component: () => import('@/views/sales/VisitManagementView.vue'),
              props: { isModernTrade: true },
            },
            {
              path: 'modern-trade/settlement',
              name: 'mtSettlement',
              component: () => import('@/views/sales/MTSettlementView.vue'),
            },
            {
              path: 'mt-delivery',
              name: 'mtDelivery',
              component: () => import('@/views/factory/MTDeliveryView.vue'),
            },
            {
              path: 'modern-trade/kam',
              name: 'mtKAM',
              component: () => import('@/views/sales/modern-trade/KAMView.vue'),
            },
            {
              path: 'modern-trade/po',
              name: 'mtPO',
              component: () => import('@/views/sales/modern-trade/POView.vue'),
            },
            {
              path: 'modern-trade/merchandising',
              redirect: { path: '/sales/modern-trade/kunjungan', query: { tab: 'merchandising' } },
            },
            {
              path: 'modern-trade/promotion',
              name: 'mtPromotion',
              component: () => import('@/views/sales/modern-trade/PromotionView.vue'),
            },
            {
              path: 'modern-trade/spg',
              name: 'mtSpg',
              component: () => import('@/views/sales/modern-trade/SpgView.vue'),
            },
            {
              path: 'modern-trade/inventory',
              name: 'mtInventory',
              component: () => import('@/views/sales/modern-trade/MTInventoryView.vue'),
            },
            {
              path: 'modern-trade/spg-reports',
              name: 'mtSpgReports',
              component: () => import('@/views/sales/modern-trade/SpgReportView.vue'),
            },
            {
              path: 'modern-trade/store-stocks',
              name: 'mtStoreStocks',
              component: () => import('@/views/sales/modern-trade/MTStoreStockComparisonView.vue'),
            },
          ],
        },
        // Factory & Logistics Module
        {
          path: 'factory',
          children: [
            {
              path: '',
              name: 'factoryDashboard',
              component: () => import('@/views/factory/FactoryDashboard.vue'),
            },
            {
              path: 'management',
              name: 'factoryManagement',
              component: () => import('@/views/factory/FactoryManagementView.vue'),
            },
            {
              path: 'warehouses',
              name: 'factoryWarehouses',
              component: () => import('@/views/factory/FactoryWarehouseView.vue'),
            },
            {
              path: 'products',
              name: 'factoryProducts',
              component: () => import('@/views/factory/FactoryProductView.vue'),
            },
            {
              path: 'raw-materials',
              name: 'factoryRawMaterials',
              component: () => import('@/views/factory/RawMaterialManagementView.vue'),
            },
            {
              path: 'raw-materials/requests',
              name: 'factoryRawMaterialRequests',
              component: () => import('@/views/factory/RawMaterialRequestsView.vue'),
              meta: { title: 'Purchase Request Bahan Baku' },
            },
            {
              path: 'raw-materials/orders',
              name: 'factoryRawMaterialOrders',
              component: () => import('@/views/factory/RawMaterialOrdersView.vue'),
            },
            {
              path: 'raw-materials/receipts',
              name: 'factoryRawMaterialReceipts',
              component: () => import('@/views/factory/RawMaterialReceiptsView.vue'),
            },
            {
              path: 'technician-stock',
              name: 'factoryTechnicianStock',
              component: () => import('@/views/factory/TechnicianInventoryView.vue'),
              meta: { title: 'Stok Barang Teknisi' },
            },
            {
              path: 'stock',
              name: 'factoryStockGlobal',
              component: () => import('@/views/factory/FactoryStockListView.vue'),
            },
            {
              path: 'transfers',
              name: 'factoryTransfersGlobal',
              component: () => import('@/views/factory/FactoryTransferListView.vue'),
            },
            {
              path: 'returns',
              name: 'factoryReturns',
              component: () => import('@/views/factory/FactoryReturnsView.vue'),
            },
            {
              path: 'qc-samples',
              name: 'factoryQcSamples',
              component: () => import('@/views/factory/FactoryQcSampleView.vue'),
            },
            {
              path: 'delegated-shipments',
              name: 'factoryDelegatedShipments',
              component: () => import('@/views/factory/DelegatedShipmentTracking.vue'),
            },
            {
              path: 'suppliers',
              name: 'factorySuppliers',
              component: () => import('@/views/warehouse/SupplierView.vue'),
            },
            {
              path: ':id',
              name: 'factoryInventory',
              component: () => import('@/views/factory/FactoryInventoryView.vue'),
            },
          ],
        },
        // Purchasing Module
        {
          path: 'purchasing',
          children: [
            {
              path: '',
              redirect: '/purchasing/requests',
            },
            {
              path: 'requests',
              name: 'purchasingRequests',
              component: () => import('@/views/purchasing/PurchasingRequestsView.vue'),
              meta: { title: 'Purchase Request & Tindak Lanjut' },
            },
            {
              path: 'history',
              name: 'purchasingHistory',
              component: () => import('@/views/purchasing/PurchasingHistoryView.vue'),
              meta: { title: 'History Purchasing & PO' },
            },
            {
              path: 'suppliers',
              name: 'purchasingSuppliers',
              component: () => import('@/views/warehouse/SupplierView.vue'),
              meta: { title: 'Data Supplier' },
            },
          ],
        },
        // Warehouse Module (Gudang Cabang)
        {
          path: 'warehouse',
          children: [
            {
              path: 'setoran-harian',
              name: 'warehouseSetoranHarian',
              component: () => import('@/views/warehouse/BranchDepositView.vue'),
            },
            {
              path: '',
              name: 'warehouseDashboard',
              component: () => import('@/views/warehouse/WarehouseDashboard.vue'),
            },
            {
              path: 'daily-workers-salary',
              name: 'warehouseDailyWorkersSalary',
              component: () => import('@/views/warehouse/DailyWorkerSalaryView.vue'),
            },
            {
              path: 'stock',
              alias: 'adjustment',
              name: 'warehouseStock',
              component: () => import('@/views/warehouse/WarehouseManagementView.vue'),
            },
            {
              path: 'het-pricing',
              name: 'warehouseHETPricing',
              component: () => import('@/views/warehouse/HETPricingView.vue'),
            },
            {
              path: 'logs',
              name: 'warehouseLogs',
              component: () => import('@/views/warehouse/WarehouseLogView.vue'),
            },
            {
              path: 'tagihan-pabrik',
              name: 'warehouseTagihanPabrik',
              component: () => import('@/views/warehouse/BranchReceivablesView.vue'),
            },
            {
              path: 'incoming-orders',
              name: 'warehouseIncomingOrders',
              component: () => import('@/views/warehouse/IncomingOrdersView.vue'),
            },
            {
              path: 'do-management',
              name: 'warehouseDOManagement',
              component: () => import('@/views/warehouse/DOManagementView.vue'),
            },
            {
              path: 'surat-jalan',
              name: 'warehouseSuratJalan',
              component: () => import('@/views/sales/SuratJalanView.vue'),
            },
            {
              path: 'lps',
              name: 'warehouseLps',
              component: () => import('@/views/sales/LPSView.vue'),
            },
            {
              path: 'sales-transfers',
              name: 'warehouseSalesTransfers',
              component: () => import('@/views/warehouse/SalesTransferView.vue'),
            },
            {
              path: 'distributor-requests',
              name: 'warehouseDistributorRequests',
              component: () => import('@/views/warehouse/DistributorRequestView.vue'),
            },
            {
              path: 'distributor-approval',
              name: 'warehouseDistributorApproval',
              component: () => import('@/views/warehouse/DistributorApprovalView.vue'),
            },
            {
              path: 'retur',
              alias: '/sales/retur',
              name: 'salesReturns',
              component: () => import('@/views/sales/SalesReturnView.vue'),
            },
            {
              path: 'suppliers',
              redirect: '/factory/suppliers',
            },
            {
              path: 'supplier-receipts',
              name: 'warehouseSupplierReceipts',
              component: () => import('@/views/warehouse/SupplierReceiptView.vue'),
            },

            {
              path: 'leftover',
              name: 'warehouseLeftover',
              component: () => import('@/views/warehouse/LeftoverDeliveryView.vue'),
            },
            {
              path: 'bonus-inventory',
              name: 'bonusInventory',
              component: () => import('@/views/warehouse/BonusInventoryView.vue'),
            },
            {
              path: 'ancillary-goods',
              name: 'warehouseAncillaryGoods',
              component: () => import('@/views/sales/AncillaryGoodsView.vue'),
            },
          ],
        },
        // Finance Module
        {
          path: 'finance',
          children: [
            {
              path: 'pendapatan-sales-harian',
              name: 'financeDailySalesIncome',
              component: () => import('@/views/finance/DailySalesIncomeView.vue'),
              meta: { title: 'Pendapatan Sales Harian', requiresAuth: true },
            },
            {
              path: 'setoran-faktur',
              name: 'financeInvoiceDeposit',
              redirect: '/finance/pelaporan-kasir',
            },
            {
              path: 'penggajian-harian',
              name: 'financeDailyWorkerPayroll',
              component: () => import('@/views/finance/DailyWorkerPayrollView.vue'),
            },
            {
              path: 'piutang-cabang',
              name: 'financePiutangCabang',
              component: () => import('@/views/finance/BranchReceivablesView.vue'),
            },
            {
              path: 'validasi-setoran',
              name: 'financeSetoranCabang',
              component: () => import('@/views/finance/FinanceBranchDepositView.vue'),
            },
            {
              path: 'pelaporan-kasir',
              name: 'financeCashierReport',
              component: () => import('@/views/finance/CashierFinancialReportView.vue'),
            },
            {
              path: 'kpi-sales',
              name: 'financeKpiSales',
              component: () => import('@/views/sales/KPI-Manager.vue'),
              meta: { title: 'KPI Omzet Sales', requiresAuth: true },
            },
            {
              path: 'tracking-noo-ec',
              name: 'financeTrackingNooEc',
              component: () => import('@/views/sales/TrackingNooEcView.vue'),
              meta: { title: 'Tracking NOO & EC', requiresAuth: true },
            },
            {
              path: '',
              name: 'financeDashboard',
              component: () => import('@/views/finance/FinanceDashboard.vue'),
            },
            {
              path: 'approval',
              name: 'financeApproval',
              component: () => import('@/views/finance/FinanceApprovalView.vue'),
            },
            {
              path: 'profit-loss',
              name: 'profitLoss',
              component: () => import('@/views/finance/ProfitLossView.vue'),
            },
            {
              path: 'export',
              name: 'financeExport',
              component: () => import('@/views/finance/FinanceExportView.vue'),
            },
            {
              path: 'bank-accounts',
              name: 'bankAccounts',
              component: () => import('@/views/finance/BankAccountView.vue'),
            },
            {
              path: 'accounting',
              name: 'accounting',
              component: () => import('@/views/finance/AccountingView.vue'),
            },
            {
              path: 'purchase-invoices',
              name: 'financePurchaseInvoices',
              component: () => import('@/views/finance/PurchaseInvoicesView.vue'),
            },
            {
              path: 'ar-aging',
              name: 'arAging',
              component: () => import('@/views/finance/ARAgingView.vue'),
            },
            {
              path: 'ap-aging',
              name: 'apAging',
              component: () => import('@/views/finance/APAgingView.vue'),
            },
            {
              path: 'fixed-assets',
              name: 'fixedAssets',
              component: () => import('@/views/finance/FixedAssetView.vue'),
            },
            {
              path: 'audit-logs',
              name: 'financeAuditLogs',
              component: () => import('@/views/finance/AuditLogView.vue'),
            },
            {
              path: 'kas-pabrik',
              name: 'kasManagement',
              component: () => import('@/views/finance/KasPabrikView.vue'),
            },
          ],
        },
        // Marketplace Module
        {
          path: 'marketplace',
          children: [
            {
              path: '',
              name: 'marketplaceDashboard',
              component: () => import('@/views/marketplace/MarketplaceDashboard.vue'),
            },
            {
              path: 'stock',
              name: 'marketplaceStock',
              component: () => import('@/views/marketplace/MarketplaceStockView.vue'),
            },
            {
              path: 'orders',
              name: 'marketplaceOrders',
              component: () => import('@/views/marketplace/MarketplaceOrdersView.vue'),
            },
            {
              path: 'settlements',
              name: 'marketplaceSettlements',
              component: () => import('@/views/marketplace/MarketplaceSettlementView.vue'),
              meta: { title: 'Pencairan & Keuangan Marketplace' }
            },
            {
              path: 'profitability',
              name: 'marketplaceProfitability',
              component: () => import('@/views/marketplace/MarketplaceProfitabilityView.vue'),
              meta: { title: 'Profitability Analytics' },
            },
            {
              path: 'products',
              name: 'marketplaceProducts',
              component: () => import('@/views/marketplace/MarketplaceProductsView.vue'),
              meta: { title: 'Produk & SKU' }
            },
            {
              path: 'mapping',
              name: 'marketplaceMapping',
              component: () => import('@/views/marketplace/MarketplaceMappingView.vue'),
              meta: { title: 'Product Mapping' }
            },
            {
              path: 'pricing',
              name: 'marketplace-pricing',
              component: () => import('@/views/marketplace/MarketplacePricingView.vue'),
            },
            {
              path: 'audit-trail',
              name: 'marketplaceAuditTrail',
              component: () => import('@/views/marketplace/MarketplaceAuditTrailView.vue'),
            },
            {
              path: 'inventories',
              name: 'marketplaceInventories',
              component: () => import('@/views/marketplace/MarketplaceInventoryView.vue'),
            },
            {
              path: 'picking',
              name: 'marketplacePicking',
              component: () => import('@/views/marketplace/MarketplacePickingView.vue'),
            },
            {
              path: 'packing',
              name: 'marketplacePacking',
              component: () => import('@/views/marketplace/MarketplacePackingView.vue'),
            },
            {
              path: 'packing/multi',
              name: 'marketplacePackingMulti',
              component: () => import('@/views/marketplace/MarketplacePackingView.vue'),
            },
            {
              path: 'history',
              name: 'marketplaceHistory',
              component: () => import('@/views/marketplace/MarketplaceScanHistoryView.vue'),
              meta: { title: 'History Scan Picking & Packing' }
            },
            {
              path: 'integrations',
              name: 'marketplaceIntegrations',
              component: () => import('@/views/marketplace/MarketplaceIntegrationView.vue'),
              meta: { title: 'Integrasi Akun Marketplace' }
            },
            {
              path: 'shopee-boost',
              name: 'marketplaceShopeeBoost',
              component: () => import('@/views/marketplace/ShopeeAutoBoostView.vue'),
              meta: { title: 'Shopee Auto Boost' }
            }
          ],
        },
        // RajaTransport Module (Logistik & Ekspedisi Antar Provinsi)
        {
          path: 'rajatransport',
          children: [
            {
              path: '',
              name: 'rajatransport',
              component: () => import('@/views/raja-transport/RajaTransportDashboardView.vue'),
              meta: { requiresAuth: true, title: 'Dashboard Ekspedisi RajaTransport' },
            },
            {
              path: 'trips',
              name: 'rajatransportTrips',
              component: () => import('@/views/raja-transport/TripManagementView.vue'),
              meta: { requiresAuth: true, title: 'Riwayat & Penugasan Trip' },
            },
            {
              path: 'backhauls',
              name: 'rajatransportBackhauls',
              component: () => import('@/views/raja-transport/BackhaulView.vue'),
              meta: { requiresAuth: true, title: 'Pengiriman Muatan Balik' },
            },
            {
              path: 'claims',
              name: 'rajatransportClaims',
              component: () => import('@/views/raja-transport/ExpenseClaimView.vue'),
              meta: { requiresAuth: true, title: 'Klaim BBM & Part' },
            },
            {
              path: 'fleet-condition',
              name: 'rajatransportFleetCondition',
              component: () => import('@/views/raja-transport/FleetConditionView.vue'),
              meta: { requiresAuth: true, title: 'Kondisi & Perawatan Armada' },
            },
            {
              path: 'driver-attendance',
              name: 'rajatransportDriverAttendance',
              component: () => import('@/views/raja-transport/DriverAttendanceView.vue'),
              meta: { requiresAuth: true, title: 'Absensi Driver Ekspedisi' },
            },
            {
              path: 'tracking',
              name: 'rajatransportTracking',
              component: () => import('@/views/raja-transport/DeliveryTrackingView.vue'),
              meta: { requiresAuth: true, title: 'Tracking Pengiriman (GPS)' },
            },
          ],
        },
      ],
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'notFound',
      redirect: '/',
    },
  ],
})

router.beforeEach((to, from) => {
  const authStore = useAuthStore()
  const layoutStore = useLayoutStore()

  // 1. Check if route requires auth
  if (to.meta.requiresAuth) {
    if (!authStore.isAuthenticated) {
      // Not logged in -> Login
      return '/login'
    }

    if (!authStore.isAuthorized) {
      // Logged in but not authorized (not superadmin/admin/HR)
      // For now, logout and redirect to login, or we can show an error
      authStore.logout()
      return '/login'
    }

    const role = authStore.user?.role
    const path = to.path
    if (!isModuleEnabled(moduleForPath(path)) || (path === '/pkwt' && !isPkwtEnabled)) return '/'


    // Redirect employee role to portal view
    if (role === 'employee' && to.name !== 'employeePortal') {
      return { name: 'employeePortal' }
    }

    // Restrict HRIS Master Data access to superadmin only
    const hrisMasterDataPaths = ['/users', '/active-users', '/leave-types', '/organization', '/job-positions', '/shifts', '/announcements']
    if (hrisMasterDataPaths.includes(path) && role !== 'superadmin' && role !== 'auditor' && role !== 'customer_support') {
      return '/'
    }

    // Restrict Supplier Master Data access to superadmin and admin_purchasing only
    if ((path === '/factory/suppliers' || path.startsWith('/factory/suppliers/') || path === '/warehouse/suppliers') && role !== 'superadmin' && role !== 'admin_purchasing') {
      return '/'
    }

    // Restrict Factory PO access to superadmin, admin_purchasing, and manager only
    if ((path === '/factory/raw-materials/orders' || path.startsWith('/factory/raw-materials/orders/')) && role !== 'superadmin' && role !== 'admin_purchasing' && role !== 'manager') {
      return '/factory/raw-materials/requests'
    }

    // Role-based route redirection & protection
    if (role === 'distributor_admin' || role === 'distributor_sales') {
      layoutStore.setModule('sales')
      if (path !== '/sales/distributor-portal') {
        return '/sales/distributor-portal'
      }
      return true
    }

    if (role === 'supervisor_sales' || role === 'supervisor_sales_mt' || role === 'admin_nota') {
      if (role === 'admin_nota') {
        if (
          path === '/' ||
          (!path.startsWith('/sales') &&
            path !== '/factory/qc-samples' &&
            !path.startsWith('/factory/qc-samples/') &&
            path !== '/kehadiran' &&
            path !== '/lps-print')
        ) {
          layoutStore.setModule('sales')
          return '/sales'
        }
      } else {
        const restrictedBrandingMasters = ['/sales/vendor', '/sales/editor', '/sales/operator', '/sales/branding-kpi', '/sales/branding-integration', '/sales/komisi-operator', '/sales/video-kampanye']
        if (restrictedBrandingMasters.includes(path)) {
          layoutStore.setModule('branding')
          return '/sales/spanduk'
        }
        if (path === '/' || (!path.startsWith('/sales') && path !== '/kehadiran' && path !== '/lps-print')) {
          layoutStore.setModule('sales')
          return '/sales'
        }
      }
    } else if (role === 'cashier') {
      if (path === '/finance/kas-pabrik' || path === '/payroll') {
        layoutStore.setModule('finance')
        return '/finance'
      }
      if (path === '/' || (!path.startsWith('/finance') && !path.startsWith('/reimbursements') && path !== '/kehadiran' && path !== '/lps-print')) {
        layoutStore.setModule('finance')
        return '/finance'
      }
    } else if (role === 'cashier_pusat') {
      if (path === '/' || (!path.startsWith('/finance') && !path.startsWith('/reimbursements') && path !== '/payroll' && path !== '/kehadiran' && path !== '/lps-print')) {
        layoutStore.setModule('finance')
        return '/finance'
      }
    } else if (role === 'admin_kasir') {
      if (path === '/finance/kas-pabrik' || path === '/payroll') {
        layoutStore.setModule('finance')
        return '/finance'
      }
      const isHrisAllowed = path === '/' || path.startsWith('/attendance') || path === '/employees' || path.startsWith('/daily-employees')
      if (!path.startsWith('/finance') && !path.startsWith('/reimbursements') && path !== '/kehadiran' && path !== '/lps-print' && !isHrisAllowed) {
        layoutStore.setModule('finance')
        return '/finance'
      }
    } else if (role === 'admin_gudang') {
      const warehouseAllowedSalesPaths = ['/sales/surat-jalan', '/sales/lps', '/lps-print']
      if (
        path === '/' ||
        (!path.startsWith('/warehouse') &&
          !path.startsWith('/vehicles') &&
          !warehouseAllowedSalesPaths.includes(path) &&
          path !== '/kehadiran')
      ) {
        layoutStore.setModule('warehouse')
        return '/warehouse'
      }
    } else if (role === 'admin_marketplace') {
      if (path === '/' || (!path.startsWith('/marketplace') && path !== '/kehadiran')) {
        layoutStore.setModule('marketplace')
        return '/marketplace'
      }
    } else if (role === 'admin_designer' || role === 'validator_desain' || role === 'validator_toko') {
      const allowedBranding = [
        '/sales/spanduk',
        '/sales/spanduk-editor',
        '/sales/spanduk-validator',
        '/sales/spanduk-vendor',
        '/sales/spanduk-pemasangan',
        '/sales/customers',
        '/sales/vendor',
        '/sales/editor',
        '/sales/operator',
        '/sales/branding-kpi',
        '/sales/video-kampanye',
        '/sales/branding-integration',
        '/sales/komisi-operator',
      ]
      let defaultPath = '/sales/spanduk'
      if (role === 'admin_designer') {
        defaultPath = '/sales/spanduk-editor'
      } else if (role === 'validator_desain' || role === 'validator_toko') {
        defaultPath = '/sales/spanduk-validator'
      }

      if (path === '/' || (!allowedBranding.includes(path) && path !== '/kehadiran')) {
        layoutStore.setModule('branding')
        return defaultPath
      }
    } else if (role === 'admin_purchasing') {
      const allowedBranding = [
        '/sales/spanduk',
        '/sales/spanduk-editor',
        '/sales/spanduk-validator',
        '/sales/spanduk-vendor',
        '/sales/spanduk-pemasangan',
        '/sales/customers',
        '/sales/vendor',
        '/sales/editor',
        '/sales/operator',
        '/sales/branding-kpi',
        '/sales/video-kampanye',
        '/sales/branding-integration',
        '/sales/komisi-operator',
      ]
      const isAllowed =
        path.startsWith('/purchasing') ||
        allowedBranding.includes(path) ||
        path.startsWith('/factory') ||
        path.startsWith('/finance') ||
        path.startsWith('/reimbursements') ||
        path === '/payroll' ||
        path === '/bi-dashboard' ||
        path.startsWith('/vehicles') ||
        path === '/kehadiran' ||
        path.startsWith('/kehadiran/')

      if (path === '/' || !isAllowed) {
        layoutStore.setModule('purchasing')
        return '/purchasing/requests'
      }
    } else if (role === 'admin_pabrik') {
      const allowedFactoryMenus = authStore.user?.menus || []
      let defaultFactoryPath = '/factory'
      if (allowedFactoryMenus.length > 0 && !allowedFactoryMenus.includes('/factory')) {
        const firstFactory = allowedFactoryMenus.find((m: string) => m.startsWith('/factory') || m === '/vehicles')
        if (firstFactory) defaultFactoryPath = firstFactory
      }

      if (
        path === '/' ||
        (!path.startsWith('/factory') &&
          !path.startsWith('/vehicles') &&
          !path.startsWith('/warehouse') &&
          !path.startsWith('/rajatransport') &&
          path !== '/kehadiran')
      ) {
        layoutStore.setModule('factory')
        return defaultFactoryPath
      }
    } else if (role === 'admin_operasional') {
      const isAllowed =
        path.startsWith('/warehouse') ||
        path.startsWith('/sales') ||
        path.startsWith('/finance') ||
        path.startsWith('/reimbursements') ||
        path.startsWith('/vehicles') ||
        path.startsWith('/attendance') ||
        path === '/kehadiran' ||
        path === '/lps-print'
      if (path === '/' || !isAllowed) {
        layoutStore.setModule('warehouse')
        return '/warehouse'
      }
    } else if (role === 'supervisor_gudang') {
      const warehouseAllowedSalesPaths = ['/sales/surat-jalan', '/sales/lps', '/lps-print']
      if (
        path === '/' ||
        (!path.startsWith('/warehouse') &&
          !path.startsWith('/factory') &&
          !path.startsWith('/vehicles') &&
          !warehouseAllowedSalesPaths.includes(path) &&
          path !== '/kehadiran')
      ) {
        layoutStore.setModule('warehouse')
        return '/warehouse'
      }
    }

    // Update active module state to match URL path prefix (keeps header & sidebar in sync)
    const brandingExclusivePaths = [
      '/sales/operator',
      '/sales/editor',
      '/sales/vendor',
      '/sales/spanduk-editor',
      '/sales/spanduk-validator',
      '/sales/spanduk-vendor',
      '/sales/spanduk-pemasangan',
      '/sales/branding-kpi',
      '/sales/video-kampanye',
      '/sales/branding-integration',
      '/sales/komisi-operator',
    ]

    const isSharedBrandingRoute = path === '/sales/spanduk' || path === '/sales/customers'

    if (brandingExclusivePaths.includes(path)) {
      layoutStore.setModule('branding')
    } else if (isSharedBrandingRoute) {
      if (role === 'admin_nota' || layoutStore.currentModule === 'sales') {
        layoutStore.setModule('sales')
      } else if (
        ['admin_designer', 'validator_desain', 'validator_toko', 'admin_purchasing'].includes(
          role || '',
        )
      ) {
        layoutStore.setModule('branding')
      } else if (layoutStore.currentModule !== 'branding') {
        layoutStore.setModule('sales')
      }
    } else if (path.startsWith('/sales')) {
      layoutStore.setModule('sales')
    } else if (path.startsWith('/finance') || path.startsWith('/reimbursements') || path === '/payroll' || path === '/bi-dashboard') {
      layoutStore.setModule('finance')
    } else if (
      path.startsWith('/warehouse') ||
      ((role === 'admin_gudang' || role === 'supervisor_gudang') && (path === '/sales/surat-jalan' || path === '/sales/lps'))
    ) {
      layoutStore.setModule('warehouse')
    } else if (path.startsWith('/purchasing')) {
      layoutStore.setModule('purchasing')
    } else if (path.startsWith('/factory')) {
      if (role === 'admin_nota') {
        layoutStore.setModule('sales')
      } else {
        layoutStore.setModule('factory')
      }
    } else if (path.startsWith('/marketplace')) {
      layoutStore.setModule('marketplace')
    } else if (path.startsWith('/rajatransport')) {
      layoutStore.setModule('rajatransport')
    } else if (path.startsWith('/vehicles')) {
      if (layoutStore.currentModule !== 'factory') {
        layoutStore.setModule('warehouse')
      }
    } else {
      // For personal pages like /kehadiran, retain their restricted module or fallback to hris
      if (role === 'supervisor_sales' || role === 'supervisor_sales_mt' || role === 'admin_nota') {
        layoutStore.setModule('sales')
      } else if (role === 'cashier' || role === 'cashier_pusat' || (role === 'admin_kasir' && path === '/kehadiran')) {
        layoutStore.setModule('finance')
      } else if (role === 'admin_gudang' || role === 'supervisor_gudang') {
        layoutStore.setModule('warehouse')
      } else if (role === 'admin_marketplace') {
        layoutStore.setModule('marketplace')
      } else if (role === 'admin_pabrik') {
        layoutStore.setModule('factory')
      } else if (role === 'admin_purchasing' && path === '/kehadiran') {
        layoutStore.setModule('purchasing')
      } else if (role === 'admin_designer' || role === 'validator_desain' || role === 'validator_toko') {
        layoutStore.setModule('branding')
      } else if (role === 'admin_operasional' && path === '/kehadiran') {
        layoutStore.setModule('warehouse')
      } else {
        layoutStore.setModule('hris')
      }
    }

    if (!authStore.hasPermission(path)) {
      if (path !== '/kehadiran' && !path.startsWith('/kehadiran/')) {
        return '/kehadiran'
      }
    }
  }

  // 2. Prevent logged in authorized users from seeing login or register page
  if (
    (to.name === 'login' || to.name === 'register') &&
    authStore.isAuthenticated &&
    authStore.isAuthorized
  ) {
    if (authStore.user?.role === 'employee') {
      return '/kehadiran'
    }
    const role = authStore.user?.role
    if (role === 'supervisor_sales' || role === 'admin_nota') {
      return '/sales'
    }
    if (role === 'cashier' || role === 'cashier_pusat') {
      return '/finance'
    }
    if (role === 'admin_gudang' || role === 'supervisor_gudang') {
      return '/warehouse'
    }
    if (role === 'admin_marketplace') {
      return '/marketplace'
    }
    if (role === 'admin_designer') {
      return '/sales/spanduk'
    }
    if (role === 'validator_desain' || role === 'validator_toko') {
      return '/sales/spanduk-validator'
    }
    if (role === 'admin_purchasing') {
      return '/sales/spanduk-vendor'
    }
    if (role === 'admin_pabrik') {
      const allowedFactoryMenus = authStore.user?.menus || []
      let defaultFactoryPath = '/factory'
      if (allowedFactoryMenus.length > 0 && !allowedFactoryMenus.includes('/factory')) {
        const firstFactory = allowedFactoryMenus.find((m: string) => m.startsWith('/factory') || m === '/vehicles')
        if (firstFactory) defaultFactoryPath = firstFactory
      }
      return defaultFactoryPath
    }
    return '/'
  }

  return true
})

router.afterEach(() => {
  resetLoadingBar()
})

router.onError((error, to) => {
  resetLoadingBar()
  const isChunkLoadFailed =
    error.name === 'ChunkLoadError' ||
    (error.message && (
      error.message.includes('Failed to fetch dynamically imported module') ||
      error.message.includes('Importing a module script failed') ||
      error.message.includes('error loading dynamically imported module')
    ))

  if (isChunkLoadFailed) {
    console.warn('Chunk load error detected, checking reload guard...', error)
    const reloadKey = 'last_chunk_reload_attempt'
    const lastAttempt = sessionStorage.getItem(reloadKey)
    const now = Date.now()

    // Hanya reload otomatis maksimal 1 kali dalam 15 detik untuk mencegah infinite reload loop
    if (!lastAttempt || now - parseInt(lastAttempt, 10) > 15000) {
      sessionStorage.setItem(reloadKey, now.toString())
      console.warn('Reloading page once to fetch latest app bundle...')
      window.location.reload()
    } else {
      console.error('Persistent chunk loading failure. Halting reload loop to prevent browser freeze.', error)
    }
  } else {
    console.error('Router error:', error)
  }
})

export default router
