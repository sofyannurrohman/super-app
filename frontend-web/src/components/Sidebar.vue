<script setup lang="ts">
import { RouterLink, useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { useAuthStore } from '@/stores/auth'
import { useLayoutStore } from '@/stores/layout'
import { useMasterDataStore } from '@/stores/masterData'
import { computed } from 'vue'
import { brand } from '@/config/brand'
import { isPkwtEnabled } from '@/config/modules'
import {
  LayoutDashboard,
  BarChart3,
  Users,
  Clock,
  Wallet,
  FileText,
  Settings,
  Briefcase,
  Calculator,
  Calendar,
  UserCircle,
  Building2,
  Globe,
  MapPin,
  Network,
  ClipboardList,
  CalendarCheck,
  History,
  Banknote,
  Store,
  Target,
  Receipt,
  Megaphone,
  Timer,
  Factory,
  Package,
  Truck,
  Warehouse,
  Boxes,
  ArrowLeftRight,
  RotateCcw,
  ShoppingCart,
  Video,
  PackageCheck,
  LogOut,
  ClipboardCheck,
  FileCheck2,
  Layers,
  PenTool,
  CheckCircle2,
  Brush,
  Printer,
  X,
  Menu,
  Sparkles,
  Landmark,
  Tag,
  PackageSearch,
  Box,
  AlertTriangle,
  TrendingUp,
  TrendingDown,
  BookOpen,
  Activity,
  Coins,
  Component,
  Image as ImageIcon,
  Lock,
  Check,
  UserPlus,
  Wrench,
  FlaskConical,
  Rocket,
  Award,
} from 'lucide-vue-next'

const authStore = useAuthStore()
const layoutStore = useLayoutStore()
const masterDataStore = useMasterDataStore()
const router = useRouter()

const isBranchSelectorLocked = computed(() => {
  const user = authStore.user
  if (!user || !user.branch_id || user.has_global_access) return false
  if (
    user.role === 'superadmin' ||
    user.role === 'customer_support' ||
    user.role === 'hr_admin' ||
    user.role === 'admin_designer' ||
    user.role === 'validator_desain' ||
    user.role === 'validator_toko' ||
    user.role === 'admin_purchasing' ||
    user.role === 'purchasing' ||
    user.role === 'supervisor_gudang' ||
    user.role === 'cashier_pusat'
  ) {
    return false
  }
  if (user.role === 'auditor') {
    const b = masterDataStore.branches.find((br: any) => br.id === user.branch_id)
    const isBooster = b ? (b.name || '').toLowerCase().includes('booster') : false
    return isBooster
  }
  return true
})

const isBranchFilterSupported = computed(() => {
  return ['hris', 'sales', 'warehouse', 'finance', 'marketplace', 'branding', 'purchasing'].includes(
    layoutStore.currentModule,
  )
})

const isFactoryAdminNota = computed(() => {
  const u = authStore.user
  if (!u) return false

  // 1. Dapatkan ID cabang aktif (dari filter header atau dari branch user)
  let activeBranchId = masterDataStore.selectedBranchId
  if (!activeBranchId || activeBranchId === 'ALL_BRANCHES') {
    activeBranchId = u.branch_id || ''
  }

  // 2. Ambil nama cabang aktif
  let branchName = ''
  if (activeBranchId && activeBranchId !== 'ALL_BRANCHES') {
    const b = masterDataStore.branches.find((br: any) => br.id === activeBranchId)
    if (b) branchName = (b.name || '').toLowerCase()
  } else if (u.branch?.name) {
    branchName = (u.branch.name || '').toLowerCase()
  }

  // 3. Wajib merupakan cabang Pabrik
  const isPabrik = branchName.includes('pabrik')

  // 4. Hanya tampil jika cabang aktif adalah Pabrik Utama
  if (u.role === 'admin_nota' || u.role === 'superadmin' || u.role === 'manager') {
    return isPabrik
  }

  return false
})

const isPabrikBranchActive = computed(() => {
  const u = authStore.user
  if (!u) return false

  let activeBranchId = masterDataStore.selectedBranchId
  if (!activeBranchId || activeBranchId === 'ALL_BRANCHES') {
    activeBranchId = u.branch_id || ''
  }

  let branchName = ''
  if (activeBranchId && activeBranchId !== 'ALL_BRANCHES') {
    const b = masterDataStore.branches.find((br: any) => br.id === activeBranchId)
    if (b) branchName = (b.name || '').toLowerCase()
  } else if (u.branch?.name) {
    branchName = (u.branch.name || '').toLowerCase()
  }

  return branchName.includes('pabrik')
})

const currentModuleBranchId = computed(() => {
  if (layoutStore.currentModule === 'warehouse') {
    return masterDataStore.selectedWarehouseBranchId || 'ALL_BRANCHES'
  }
  return masterDataStore.selectedBranchId || 'ALL_BRANCHES'
})

const currentModuleBranchName = computed(() => {
  const branchId = currentModuleBranchId.value
  if (!branchId || branchId === 'ALL_BRANCHES') return 'Semua Cabang'
  return masterDataStore.branches.find((b) => b.id === branchId)?.name || 'Semua Cabang'
})

function handleDrawerBranchChange(branchId: string) {
  if (layoutStore.currentModule === 'warehouse') {
    masterDataStore.setSelectedWarehouseBranchId(branchId)
  } else {
    masterDataStore.setSelectedBranchId(branchId)
  }
}

const closeDrawer = () => layoutStore.closeMobileDrawer()

const user = computed(() => authStore.user)

const userInitials = computed(() => {
  if (!user.value || !user.value.email) return 'AD'
  return user.value.email.substring(0, 2).toUpperCase()
})

const userDisplayName = computed(() => {
  if (!user.value || !user.value.email) return 'Admin HR'
  const namePart = user.value.email.split('@')[0]
  return namePart.charAt(0).toUpperCase() + namePart.slice(1)
})

const userRoleName = computed(() => {
  if (!user.value || !user.value.role) return 'Super Admin'
  const role = user.value.role
  if (role === 'superadmin') return 'Super Admin'
  if (role === 'customer_support') return 'Customer Support'
  if (role === 'hr_admin') return 'HR Admin'
  if (role === 'manager') return 'Manager'
  if (role === 'supervisor_sales' || role === 'supervisor_sales_mt') return 'Supervisor Sales'
  if (role === 'cashier') return 'Cashier Cabang'
  if (role === 'cashier_pusat') return 'Cashier Pusat'
  if (role === 'admin_kasir') return 'Admin Kasir'
  if (role === 'admin_nota') return 'Admin Nota'
  if (role === 'admin_gudang') return 'Admin Gudang'
  if (role === 'admin_marketplace') return 'Admin Marketplace'
  if (role === 'admin_designer') return 'Admin Designer'
  if (role === 'validator_desain') return 'Validator Desain'
  if (role === 'validator_toko') return 'Validator Toko'
  if (role === 'admin_purchasing') return 'Admin Purchasing'
  if (role === 'admin_operasional') return 'Admin Operasional'
  if (role === 'admin_pabrik') return 'Admin Pabrik'
  if (role === 'supervisor_gudang') return 'Supervisor Gudang'
  if (role === 'employee') return 'Karyawan'
  return role.charAt(0).toUpperCase() + role.slice(1)
})

// Logout handler
const handleLogout = () => {
  authStore.logout()
  router.push('/login')
}

const hrisGroups = [
  {
    title: 'Utama',
    items: [
      { to: '/', name: 'Dashboard', icon: LayoutDashboard },
      { to: '/employees', name: 'Karyawan', icon: Users },
      { to: '/daily-employees', name: 'Karyawan Harian', icon: Users },
      { to: '/karyawan-bermasalah', name: 'Karyawan Perlu Tindakan', icon: AlertTriangle },
      { to: '/attendance', name: 'Kehadiran Karyawan', icon: Clock },
      { to: '/attendance-daily', name: 'Kehadiran Harian', icon: Clock },
      { to: '/leaves', name: 'Cuti & Izin', icon: FileText },
      { to: '/employee-requests', name: 'Permintaan Karyawan', icon: UserPlus },
      { to: '/memo-requests', name: 'Request Memo', icon: FileText },
      { to: '/overtimes', name: 'Lembur', icon: Timer },
      { to: '/performance-kpi', name: 'Performa & KPI', icon: Target },
      { to: '/pkwt', name: 'Surat PKWT', icon: FileCheck2 },
    ],
  },
  {
    title: 'Master Data',
    items: [
      { to: '/users', name: 'Pengguna Akun', icon: UserCircle },
      { to: '/active-users', name: 'Pengguna Aktif', icon: Activity },
      { to: '/leave-types', name: 'Tipe Cuti', icon: Calendar },
      { to: '/organization', name: 'Manajemen Organisasi', icon: Network },
      { to: '/job-positions', name: 'Jabatan', icon: Briefcase },
      { to: '/shifts', name: 'Shift Kerja', icon: Clock },
      { to: '/announcements', name: 'Pengumuman', icon: Megaphone },
    ],
  },
  {
    title: 'Pengaturan',
    items: [
      { to: '/system-modules', name: 'Access Control', icon: Settings },
      { to: '/access-roles', name: 'Hak Akses', icon: Settings },
    ],
  },
]

const salesGroups = [
  {
    title: 'Operasional Sales',
    items: [
      { to: '/sales', name: 'Dashboard Sales', icon: LayoutDashboard },
      { to: '/sales/proyeksi-omzet', name: 'Proyeksi Omzet', icon: TrendingUp },
      { to: '/sales/spv-approval', name: 'Persetujuan SPV', icon: ClipboardCheck },
      { to: '/sales/kunjungan', name: 'Kunjungan Sales', icon: MapPin },
      { to: '/sales/pesanan', name: 'Pesanan SO', icon: ShoppingCart },
      { to: '/sales/delivery', name: 'Pengiriman', icon: Truck },
      { to: '/sales/surat-jalan', name: 'Data Surat Jalan', icon: FileText },
      { to: '/sales/penjualan', name: 'Faktur Penjualan', icon: Banknote },
      { to: '/sales/retur', name: 'Retur', icon: RotateCcw },
      { to: '/sales/lps', name: 'Lembar LPS', icon: ClipboardList },
      { to: '/sales/penerimaan', name: 'Penerimaan', icon: FileCheck2 },
      { to: '/sales/spanduk', name: 'Spanduk & Banner', icon: Megaphone },
      { to: '/sales/distributor-portal', name: 'Portal Distributor', icon: PackageCheck },
    ],
  },
  {
    title: 'Master Data',
    items: [
      { to: '/sales/salesman', name: 'Kelola Salesman', icon: Users },
      { to: '/sales/customers', name: 'Kelola Toko', icon: Store },
      { to: '/sales/transfer-toko', name: 'Transfer Toko', icon: ArrowLeftRight },
    ],
  },
  {
    title: 'Reporting',
    items: [
      { to: '/sales/monitoring', name: 'Live Monitoring', icon: Globe },
      { to: '/sales/tracking-noo-ec', name: 'Tracking NOO & EC', icon: TrendingUp },
      { to: '/sales/kpi', name: 'KPI Manager', icon: Target },
      { to: '/sales/laporan-kinerja-harian', name: 'Laporan Kinerja Harian', icon: Award },
      { to: '/sales/audit-logs', name: 'Audit Log', icon: History },
    ],
  },
]

const brandingGroups = [
  {
    title: 'Alur Spanduk & Banner',
    items: [
      { to: '/sales/spanduk', name: 'Spanduk & Banner', icon: Megaphone },
      { to: '/sales/spanduk-editor', name: 'Editor Desain', icon: Brush },
      { to: '/sales/spanduk-validator', name: 'Validator Spanduk', icon: CheckCircle2 },
      { to: '/sales/spanduk-vendor', name: 'Vendor & Purchasing', icon: Printer },
      { to: '/sales/spanduk-pemasangan', name: 'Bukti Tayang', icon: ImageIcon },
    ],
  },
  {
    title: 'Master & Manajemen',
    items: [
      { to: '/sales/video-kampanye', name: 'Video Kampanye', icon: Video },
      { to: '/sales/customers', name: 'Kelola Toko', icon: Store },
      { to: '/sales/vendor', name: 'Kelola Vendor', icon: Building2 },
      { to: '/sales/editor', name: 'Kelola Editor', icon: PenTool },
      { to: '/sales/operator', name: 'Kelola Operator', icon: UserCircle },
      { to: '/sales/komisi-operator', name: 'Komisi Operator', icon: Wallet },
      { to: '/sales/branding-kpi', name: 'KPI Manager', icon: Target },
    ],
  },
]

const modernTradeSalesGroups = [
  {
    title: 'Operasional & Logistik MT',
    items: [
      { to: '/sales', name: 'Dashboard MT', icon: LayoutDashboard },
      { to: '/sales/modern-trade/kunjungan', name: 'Kunjungan & Merchandising', icon: MapPin },
      { to: '/sales/modern-trade/po', name: 'PO & Pengiriman MT', icon: ShoppingCart },
      { to: '/sales/modern-trade/inventory', name: 'Stok & Inventaris MT', icon: Boxes },
      { to: '/sales/modern-trade/store-stocks', name: 'Perbandingan Stok Outlet', icon: Boxes },
      { to: '/sales/modern-trade/spg-reports', name: 'Laporan Penjualan SPG', icon: Receipt },
      { to: '/sales/modern-trade/settlement', name: 'Settlement Omzet MT', icon: ClipboardCheck },
    ],
  },
  {
    title: 'Master Data & Reporting',
    items: [
      { to: '/sales/salesman', name: 'Kelola Sales MT', icon: Users },
      { to: '/sales/modern-trade/spg', name: 'Kelola SPG', icon: Users },
      { to: '/sales/modern-trade/kam', name: 'Key Accounts (KAM)', icon: Building2 },
      { to: '/sales/customers', name: 'Outlet MT (Retailer)', icon: Store },
      { to: '/sales/tracking-noo-ec', name: 'Tracking NOO & EC', icon: TrendingUp },
      { to: '/sales/kpi', name: 'Performa & KPI MT', icon: Target },
    ],
  },
]

const factoryGroups = [
  {
    title: 'Operasional Pabrik',
    items: [
      { to: '/factory', name: 'Dashboard Pabrik', icon: LayoutDashboard },
      { to: '/factory/stock', name: 'Manajemen Stok', icon: Factory },
      { to: '/factory/transfers', name: 'Manajemen Pengiriman', icon: Truck },
      { to: '/factory/returns', name: 'Barang Returan', icon: RotateCcw },
      { to: '/factory/qc-samples', name: 'Sampel Lab & QC', icon: FlaskConical },
    ],
  },
  {
    title: 'Bahan Baku & Pengadaan',
    items: [
      { to: '/factory/raw-materials', name: 'Stok Bahan Baku', icon: Component },
      { to: '/factory/raw-materials/requests', name: 'Purchase Request (PR)', icon: ClipboardList },
      { to: '/factory/raw-materials/orders', name: 'PO Pengadaan Bahan', icon: ShoppingCart },
      { to: '/factory/raw-materials/receipts', name: 'Penerimaan Supplier', icon: PackageCheck },
      { to: '/factory/technician-stock', name: 'Stok Barang Teknisi', icon: Wrench },
    ],
  },
  {
    title: 'Master Data',
    items: [
      { to: '/factory/management', name: 'Manajemen Pabrik', icon: Building2 },
      { to: '/factory/warehouses', name: 'Gudang Pabrik', icon: Warehouse },
      { to: '/factory/products', name: 'Produk', icon: Package },
      { to: '/factory/suppliers', name: 'Data Supplier', icon: Building2 },
      { to: '/vehicles', name: 'Manajemen Kendaraan', icon: Truck },
    ],
  },
]

const warehouseGroups = [
  {
    title: 'Operasional Gudang',
    items: [
      { to: '/warehouse', name: 'Dashboard Gudang', icon: LayoutDashboard },
      { to: '/warehouse/stock', name: 'Persediaan', icon: Warehouse },
      { to: '/warehouse/incoming-orders', name: 'SJ Masuk Gudang', icon: ClipboardList },
      { to: '/warehouse/supplier-receipts', name: 'Penerimaan Supplier', icon: Package },
      { to: '/warehouse/do-management', name: 'Kelola DO & Cadangan', icon: FileCheck2 },
      { to: '/warehouse/sales-transfers', name: 'Pindah Barang', icon: Boxes },
      { to: '/warehouse/surat-jalan', name: 'Data Surat Jalan', icon: FileText },
      { to: '/warehouse/lps', name: 'Lembar LPS', icon: ClipboardList },
      { to: '/warehouse/retur', name: 'Retur', icon: RotateCcw },
      { to: '/warehouse/daily-workers-salary', name: 'Gaji Pekerja Harian', icon: Wallet },
      { to: '/warehouse/tagihan-pabrik', name: 'Tagihan Pabrik', icon: Banknote },
      { to: '/warehouse/logs', name: 'Log Mutasi', icon: History },
    ],
  },
  {
    title: 'Master Data',
    items: [
      { to: '/warehouse/het-pricing', name: 'Harga HET Pelanggan', icon: Tag },
      { to: '/vehicles', name: 'Manajemen Kendaraan', icon: Truck },
      { to: '/warehouse/bonus-inventory', name: 'Inventaris Tambahan', icon: Sparkles },
      { to: '/warehouse/ancillary-goods', name: 'Stok Wadah Dijual', icon: Package },
    ],
  },
]

const financeGroups = [
  {
    title: 'Operasional Kasir',
    items: [
      { to: '/finance/pendapatan-sales-harian', name: 'Pendapatan Sales Harian', icon: Banknote },
      { to: '/finance/approval', name: 'Persetujuan Kasir', icon: ClipboardCheck },
      { to: '/finance/pelaporan-kasir', name: 'Laporan Kasir & Bank', icon: Landmark },
      { to: '/finance/kpi-sales', name: 'KPI Omzet Sales', icon: Target },
      { to: '/finance/tracking-noo-ec', name: 'Tracking NOO & EC', icon: TrendingUp },
      { to: '/finance/kas-pabrik', name: 'Manajemen KAS Pabrik', icon: Coins },
      { to: '/finance/export', name: 'Ekspor Data Finansial', icon: FileText },
    ],
  },
  {
    title: 'Operasional Keuangan',
    items: [
      { to: '/finance', name: 'Dashboard Keuangan', icon: LayoutDashboard },
      { to: '/bi-dashboard', name: 'BI Dashboard', icon: BarChart3 },
      { to: '/payroll', name: 'Penggajian', icon: Wallet },
      { to: '/finance/penggajian-harian', name: 'Gaji Pekerja Harian', icon: Wallet },
      { to: '/reimbursements', name: 'Reimbursement', icon: Receipt },
    ],
  },
  {
    title: 'Akuntansi & Pembukuan',
    items: [
      { to: '/finance/accounting', name: 'Akuntansi & GL', icon: BookOpen },
      { to: '/finance/profit-loss', name: 'Laporan Laba Rugi', icon: Calculator },
      { to: '/finance/bank-accounts', name: 'Rekening Bank', icon: Landmark },
    ],
  },
  {
    title: 'Hutang & Piutang',
    items: [
      { to: '/finance/purchase-invoices', name: 'Faktur Pembelian', icon: Receipt },
      { to: '/finance/piutang-cabang', name: 'Piutang Cabang (Pabrik)', icon: Banknote },
      { to: '/finance/ar-aging', name: 'Piutang (AR Aging)', icon: TrendingUp },
      { to: '/finance/ap-aging', name: 'Hutang (AP Aging)', icon: TrendingDown },
    ],
  },
  {
    title: 'Aset Tetap',
    items: [
      { to: '/finance/fixed-assets', name: 'Manajemen Aset Tetap', icon: Building2 },
    ],
  },
  {
    title: 'Reporting',
    items: [
      { to: '/finance/audit-logs', name: 'Audit Log', icon: History },
    ],
  },
]


const marketplaceGroups = [
  {
    title: 'Operasional Marketplace',
    items: [
      { to: '/marketplace', name: 'Dashboard', icon: LayoutDashboard },
      { to: '/marketplace/shopee-boost', name: 'Shopee Auto Boost', icon: Rocket },
      { to: '/marketplace/stock', name: 'Stok Marketplace', icon: Package },
      { to: '/marketplace/orders', name: 'Order Marketplace', icon: ShoppingCart },
      { to: '/marketplace/settlements', name: 'Pencairan & Keuangan', icon: Wallet },
      { to: '/marketplace/profitability', name: 'Profitability Analytics', icon: BarChart3 },
      { to: '/marketplace/packing/multi', name: 'Packing Multi-Order', icon: PackageSearch },
      { to: '/marketplace/packing', name: 'Order Packing', icon: Box },
      { to: '/marketplace/history', name: 'History Scan', icon: History },
      { to: '/marketplace/audit-trail', name: 'Riwayat Pergerakan Stok', icon: History },
    ],
  },
  {
    title: 'Master Data & Integrasi',
    items: [
      { to: '/marketplace/products', name: 'Produk & SKU', icon: Tag },
      { to: '/marketplace/mapping', name: 'Product Mapping', icon: Network },
      { to: '/marketplace/pricing', name: 'Harga Marketplace', icon: Tag },
      { to: '/marketplace/inventories', name: 'Inventaris & ATK', icon: Boxes },
      { to: '/marketplace/integrations', name: 'Integrasi Akun', icon: Store },
    ],
  },
]

const rajaTransportGroups = [
  {
    title: 'Operasional Ekspedisi',
    items: [
      { to: '/rajatransport', name: 'Dashboard Transport', icon: LayoutDashboard },
      { to: '/rajatransport/tracking', name: 'Tracking Pengiriman (GPS)', icon: MapPin },
      { to: '/rajatransport/trips', name: 'Riwayat & Penugasan Trip', icon: Truck },
      { to: '/rajatransport/backhauls', name: 'Pengiriman Muatan Balik', icon: RotateCcw },
      { to: '/rajatransport/claims', name: 'Klaim BBM & Part', icon: Receipt },
    ],
  },
  {
    title: 'Armada & Driver',
    items: [
      { to: '/rajatransport/fleet-condition', name: 'Kondisi & Perawatan Armada', icon: Wrench },
      { to: '/rajatransport/driver-attendance', name: 'Absensi Driver', icon: Clock },
      { to: '/vehicles', name: 'Master Kendaraan', icon: Truck },
    ],
  },
]

const purchasingGroups = [
  {
    title: 'Operasional Purchasing',
    items: [
      { to: '/purchasing/requests', name: 'Purchase Request & Tindak Lanjut', icon: ClipboardList },
      { to: '/purchasing/history', name: 'History Purchasing', icon: History },
    ],
  },
  {
    title: 'Master Data & Rekanan',
    items: [
      { to: '/purchasing/suppliers', name: 'Data Supplier', icon: Building2 },
      { to: '/factory/raw-materials', name: 'Katalog Bahan Baku', icon: Component },
    ],
  },
]

const employeeGroups = [
  {
    title: 'Karyawan',
    items: [{ to: '/kehadiran', name: 'Portal Absensi & Cuti', icon: LayoutDashboard }],
  },
]

const menuGroups = computed(() => {
  const filterGroups = (groups: any[]) => {
    if (user.value?.role === 'superadmin' || user.value?.role === 'customer_support') return groups.map(group => ({ ...group, items: group.items.filter((item: any) => isPkwtEnabled || item.to !== '/pkwt') })).filter(group => group.items.length > 0)
    return groups.map(group => ({
      ...group,
      items: group.items.filter((item: any) => (isPkwtEnabled || item.to !== '/pkwt') && authStore.hasPermission(item.to))
    })).filter(group => group.items.length > 0)
  }

  if (user.value?.role === 'employee') {
    return filterGroups(employeeGroups)
  }
  if (layoutStore.currentModule === 'purchasing') {
    return filterGroups(purchasingGroups)
  }
  if (layoutStore.currentModule === 'sales') {
    if (user.value?.role === 'distributor_admin' || user.value?.role === 'distributor_sales') {
      return [
        {
          title: 'Portal Distributor',
          items: [
            { to: '/sales/distributor-portal', name: 'Penjualan & Import', icon: PackageCheck },
          ],
        },
      ]
    }
    if (masterDataStore.selectedSalesType === 'sales_mt') {
      return filterGroups(modernTradeSalesGroups)
    }
    let groups = salesGroups.map((group) => {
      if (group.title === 'Operasional Sales' && isFactoryAdminNota.value) {
        return {
          ...group,
          items: [
            ...group.items,
            { to: '/sales/qc-samples', name: 'Sampel Lab & QC', icon: FlaskConical },
          ],
        }
      }
      return group
    })
    return filterGroups(groups)
  }
  if (layoutStore.currentModule === 'branding') {
    let groups = brandingGroups
    if (user.value?.role === 'supervisor_sales' || user.value?.role === 'supervisor_sales_mt') {
      groups = groups.filter(g => g.title === 'Alur Spanduk & Banner')
    }
    return filterGroups(groups)
  }
  if (layoutStore.currentModule === 'factory') {
    let groups = JSON.parse(JSON.stringify(factoryGroups))
    const userRole = user.value?.role
    // Data supplier bersifat sensitif/riskan - hanya boleh diakses oleh superadmin & admin_purchasing
    if (userRole !== 'superadmin' && userRole !== 'admin_purchasing') {
      groups = groups.map((g: any) => {
        g.items = g.items.filter((item: any) => item.to !== '/factory/suppliers')
        return g
      }).filter((g: any) => g.items.length > 0)
    }
    // Menu PO Pengadaan Bahan (dan rincian harga PO) hanya boleh diakses oleh Purchasing & Manajemen
    if (userRole !== 'superadmin' && userRole !== 'admin_purchasing' && userRole !== 'manager') {
      groups = groups.map((g: any) => {
        g.items = g.items.filter((item: any) => item.to !== '/factory/raw-materials/orders')
        return g
      }).filter((g: any) => g.items.length > 0)
    }
    return filterGroups(groups)
  }
  if (layoutStore.currentModule === 'warehouse') {
    return filterGroups(warehouseGroups)
  }
  if (layoutStore.currentModule === 'finance') {
    let groups = JSON.parse(JSON.stringify(financeGroups))
    const userRole = user.value?.role
    // Jika BUKAN cashier_pusat, superadmin, hr_admin, manager, atau admin_purchasing, sembunyikan Kas Pabrik dan Penggajian
    if (userRole !== 'cashier_pusat' && userRole !== 'superadmin' && userRole !== 'hr_admin' && userRole !== 'manager' && userRole !== 'admin_purchasing') {
      groups = groups.map((g: any) => {
        g.items = g.items.filter((item: any) => item.to !== '/finance/kas-pabrik' && item.to !== '/payroll')
        return g
      }).filter((g: any) => g.items.length > 0)
    }

    // Menu Pendapatan Sales Harian hanya untuk Pabrik Utama
    if (!isPabrikBranchActive.value) {
      groups = groups.map((g: any) => {
        g.items = g.items.filter((item: any) => item.to !== '/finance/pendapatan-sales-harian')
        return g
      }).filter((g: any) => g.items.length > 0)
    }

    return filterGroups(groups)
  }
  if (layoutStore.currentModule === 'marketplace') {
    return filterGroups(marketplaceGroups)
  }
  if (layoutStore.currentModule === 'rajatransport') {
    return filterGroups(rajaTransportGroups)
  }
  
  if (user.value?.role === 'admin_operasional') {
    return filterGroups(
      hrisGroups.map((group: any) => ({
        ...group,
        items: group.items.filter((item: any) => item.to === '/attendance'),
      })).filter((group: any) => group.items.length > 0)
    )
  }

  if (user.value?.role !== 'superadmin' && user.value?.role !== 'auditor' && user.value?.role !== 'customer_support') {
    return filterGroups(hrisGroups.filter((group: any) => group.title !== 'Master Data'))
  }
  
  return filterGroups(hrisGroups)
})
</script>

<template>
  <aside
    class="bg-white border-r border-slate-200/60 flex flex-col h-full shrink-0 transition-all duration-500 ease-in-out z-40 relative shadow-[1px_0_10px_rgba(0,0,0,0.02)]"
    :class="[layoutStore.isSidebarCollapsed ? 'w-[88px]' : 'w-[280px]']"
  >
    <!-- Mobile close button -->
    <button
      @click="closeDrawer"
      class="md:hidden absolute top-4 right-4 z-10 p-1.5 rounded-xl text-slate-400 hover:text-slate-700 hover:bg-slate-100 transition-all"
    >
      <X class="w-5 h-5" />
    </button>
    <!-- Logo Area -->
    <div
      class="py-8 transition-all duration-500"
      :class="[layoutStore.isSidebarCollapsed ? 'px-0' : 'px-6']"
    >
      <div
        class="flex items-center overflow-hidden whitespace-nowrap transition-all duration-500"
        :class="[layoutStore.isSidebarCollapsed ? 'justify-center gap-0 flex-col' : 'justify-between gap-3']"
      >
        <div class="flex items-center gap-3.5 overflow-hidden">
          <div
            class="bg-white/5 p-2 rounded-2xl border border-primary/10 shadow-sm shadow-primary/5 shrink-0 transition-all duration-500"
            :class="{ 'scale-90': layoutStore.isSidebarCollapsed }"
          >
            <img :src="brand.logoUrl" alt="Logo" class="w-8 h-8 shrink-0" />
          </div>
          <div
            class="flex flex-col transition-all duration-500 origin-left"
            :class="[
              layoutStore.isSidebarCollapsed
                ? 'opacity-0 w-0 -translate-x-4 scale-75'
                : 'opacity-100 w-auto translate-x-0 scale-100',
            ]"
          >
            <span class="text-xl font-extrabold text-slate-900 tracking-tight leading-tight"
              >{{ brand.name }}</span
            >
            <span class="text-[10px] font-bold text-primary uppercase tracking-[0.2em] leading-none"
              >{{ authStore.user?.company_name || 'People & Operations' }}</span>
            >
          </div>
        </div>

        <!-- Desktop Hamburger Toggle in Sidebar -->
        <button
          type="button"
          @click="layoutStore.toggleSidebar"
          class="hidden md:flex items-center justify-center h-8 w-8 rounded-xl text-slate-400 hover:text-primary hover:bg-slate-100 transition-all shrink-0 cursor-pointer"
          :class="{ 'mt-3': layoutStore.isSidebarCollapsed }"
          :title="layoutStore.isSidebarCollapsed ? 'Perluas Sidebar' : 'Ciutkan Sidebar'"
        >
          <Menu class="w-4.5 h-4.5 transition-transform duration-300" :class="{ 'rotate-180': layoutStore.isSidebarCollapsed }" />
        </button>
      </div>
    </div>

    <!-- Mobile Drawer Branch Filter (Hanya muncul di HP / Mobile Drawer saat modul terkait cabang aktif) -->
    <div
      v-if="isBranchFilterSupported"
      class="md:hidden px-6 pb-5 -mt-2"
    >
      <div class="rounded-2xl border border-slate-200/80 bg-slate-50/90 p-3 shadow-sm">
        <div class="flex items-center justify-between mb-2">
          <div class="flex items-center gap-1.5 text-xs font-bold text-slate-700">
            <MapPin class="w-3.5 h-3.5 text-primary" />
            <span>Filter Cabang</span>
          </div>
          <span
            v-if="isBranchSelectorLocked"
            class="text-[10px] font-semibold text-amber-600 bg-amber-50 px-1.5 py-0.5 rounded-md border border-amber-200/60 flex items-center gap-1"
          >
            <Lock class="w-2.5 h-2.5" /> Terkunci
          </span>
        </div>

        <Select
          :model-value="currentModuleBranchId"
          @update:model-value="(val) => handleDrawerBranchChange(val as string)"
          :disabled="isBranchSelectorLocked"
        >
          <SelectTrigger class="w-full h-9 rounded-xl border border-slate-200 bg-white text-xs font-semibold text-slate-800 shadow-none">
            <SelectValue placeholder="Pilih Cabang" class="truncate">
              {{ currentModuleBranchName }}
            </SelectValue>
          </SelectTrigger>
          <SelectContent class="rounded-xl border-slate-200 shadow-xl max-h-[260px] z-[9999]">
            <SelectItem value="ALL_BRANCHES" class="rounded-lg m-1 cursor-pointer font-medium text-primary text-xs">
              Semua Cabang (Global)
            </SelectItem>
            <SelectItem
              v-for="branch in masterDataStore.branches"
              :key="branch.id"
              :value="branch.id"
              class="rounded-lg m-1 cursor-pointer text-xs"
            >
              {{ branch.name }}
            </SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>

    <!-- Navigation Links -->
    <nav class="flex-1 px-4 space-y-7 overflow-y-auto no-scrollbar pb-10">
      <div v-for="group in menuGroups" :key="group.title" class="space-y-2">
        <p
          v-if="group.title"
          class="px-4 text-[10px] font-extrabold uppercase tracking-[0.15em] text-slate-400 transition-all duration-500 whitespace-nowrap overflow-hidden"
          :class="[
            layoutStore.isSidebarCollapsed ? 'opacity-0 h-0 my-0' : 'opacity-100 h-auto my-2',
          ]"
        >
          {{ group.title }}
        </p>

        <div class="space-y-1">
          <RouterLink
            v-for="item in group.items"
            :key="item.to"
            :to="item.to"
            @click="closeDrawer"
            class="group flex items-center px-4 py-3 text-slate-500 rounded-2xl transition-all duration-300 hover:bg-slate-50 hover:text-slate-900 relative overflow-hidden"
            :class="[
              layoutStore.isSidebarCollapsed ? 'justify-center gap-0 px-2' : 'gap-3.5',
              $route.path === item.to
                ? '!bg-primary/5 !text-primary shadow-sm ring-1 ring-primary/10'
                : '',
            ]"
          >
            <!-- Vertical Indicator for Active Item -->
            <div
              class="absolute left-0 top-1/2 -translate-y-1/2 w-1 h-6 bg-primary rounded-r-full transition-transform duration-300 scale-y-0 group-[.router-link-exact-active]:scale-y-100"
            ></div>

            <component
              :is="item.icon"
              class="h-5 w-5 transition-all duration-300 group-hover:scale-110 group-[.router-link-exact-active]:scale-110 shrink-0"
            />

            <span
              class="text-[14px] font-bold transition-all duration-500 origin-left whitespace-nowrap"
              :class="[
                layoutStore.isSidebarCollapsed
                  ? 'opacity-0 w-0 scale-90 invisible ml-0'
                  : 'opacity-100 w-auto scale-100 visible ml-0',
              ]"
            >
              {{ item.name }}
            </span>

            <!-- Tooltip for collapsed state (native title) -->
            <div v-if="layoutStore.isSidebarCollapsed" class="sr-only">{{ item.name }}</div>
          </RouterLink>
        </div>
      </div>

      <!-- Link Portal Absensi for Admin/HR/Manager/Supervisor (Personal Attendance) -->
      <div v-if="user && user.role !== 'employee'" class="space-y-2 pt-2 border-t border-slate-100">
        <p
          class="px-4 text-[10px] font-extrabold uppercase tracking-[0.15em] text-slate-400 transition-all duration-500 whitespace-nowrap overflow-hidden"
          :class="[
            layoutStore.isSidebarCollapsed ? 'opacity-0 h-0 my-0' : 'opacity-100 h-auto my-2',
          ]"
        >
          Pribadi
        </p>
        <div class="space-y-1">
          <RouterLink
            to="/kehadiran"
            @click="closeDrawer"
            class="group flex items-center px-4 py-3 text-slate-500 rounded-2xl transition-all duration-300 hover:bg-slate-50 hover:text-slate-900 relative overflow-hidden"
            :class="[
              layoutStore.isSidebarCollapsed ? 'justify-center gap-0 px-2' : 'gap-3.5',
              $route.path === '/kehadiran'
                ? '!bg-primary/5 !text-primary shadow-sm ring-1 ring-primary/10'
                : '',
            ]"
          >
            <!-- Vertical Indicator for Active Item -->
            <div
              class="absolute left-0 top-1/2 -translate-y-1/2 w-1 h-6 bg-primary rounded-r-full transition-transform duration-300 scale-y-0 group-[.router-link-exact-active]:scale-y-100"
              :class="{ 'scale-y-100': $route.path === '/kehadiran' }"
            ></div>

            <Clock
              class="h-5 w-5 transition-all duration-300 group-hover:scale-110 shrink-0"
              :class="{ 'text-primary': $route.path === '/kehadiran' }"
            />

            <span
              class="text-[14px] font-bold transition-all duration-500 origin-left whitespace-nowrap"
              :class="[
                layoutStore.isSidebarCollapsed
                  ? 'opacity-0 w-0 scale-90 invisible ml-0'
                  : 'opacity-100 w-auto scale-100 visible ml-0',
              ]"
            >
              Absen Saya
            </span>
            <div v-if="layoutStore.isSidebarCollapsed" class="sr-only">Absen Saya</div>
          </RouterLink>
        </div>
      </div>
    </nav>

    <!-- User Profile & Logout Area -->
    <div
      class="p-4 mt-auto border-t border-slate-100 bg-slate-50/30 transition-all duration-500"
      :class="[layoutStore.isSidebarCollapsed ? 'px-2 py-4' : 'p-5']"
    >
      <div
        class="flex items-center justify-between"
        :class="{ 'flex-col gap-3.5 items-center': layoutStore.isSidebarCollapsed }"
      >
        <!-- Profile info -->
        <div class="flex items-center gap-3 min-w-0">
          <div
            class="flex flex-col min-w-0 transition-all duration-500 origin-left"
            :class="[
              layoutStore.isSidebarCollapsed
                ? 'opacity-0 w-0 scale-90 hidden'
                : 'opacity-100 w-auto scale-100 block',
            ]"
          >
            <p class="text-xs font-bold text-slate-800 leading-none truncate max-w-[130px]">
              {{ userDisplayName }}
            </p>
            <span
              class="text-[10px] text-primary font-bold uppercase tracking-wider mt-1 truncate max-w-[130px]"
              >{{ userRoleName }}</span
            >
          </div>
        </div>

        <!-- Logout button -->
        <Button
          @click="handleLogout"
          variant="ghost"
          size="icon"
          class="h-9 w-9 text-slate-400 hover:text-destructive hover:bg-destructive/5 rounded-xl transition-all duration-300 shrink-0"
          title="Log Out"
        >
          <LogOut class="w-4.5 h-4.5" />
        </Button>
      </div>
    </div>
  </aside>
</template>

<style scoped>
/* Custom scrollbar for a cleaner sidebar look */
.no-scrollbar::-webkit-scrollbar {
  display: none;
}

.no-scrollbar {
  -ms-overflow-style: none;
  scrollbar-width: none;
}

aside {
  -webkit-tap-highlight-color: transparent;
}
</style>

<style scoped>
/* Custom scrollbar for a cleaner sidebar look */
nav::-webkit-scrollbar {
  width: 5px;
}

nav::-webkit-scrollbar-track {
  background: transparent;
}

nav::-webkit-scrollbar-thumb {
  background: #e5e7eb;
  border-radius: 10px;
}

nav::-webkit-scrollbar-thumb:hover {
  background: #d1d5db;
}

/* For Firefox */
nav {
  scrollbar-width: thin;
  scrollbar-color: #e5e7eb transparent;
}
</style>
