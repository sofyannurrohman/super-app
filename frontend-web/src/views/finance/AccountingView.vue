<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import {
  BookOpen,
  FileText,
  BarChart3,
  Plus,
  RefreshCw,
  Search,
  ChevronLeft,
  ChevronRight,
  ChevronDown,
  X,
  Trash2,
  Layers,
  TrendingUp,
  TrendingDown,
  Scale,
  Activity,
  CheckCircle,
  Clock,
  AlertCircle,
  Database,
} from 'lucide-vue-next'

// ─── Types ────────────────────────────────────────────────────────────────────

interface Account {
  id: string
  code: string
  name: string
  type: 'ASSET' | 'LIABILITY' | 'EQUITY' | 'REVENUE' | 'EXPENSE'
  normal_balance: 'DEBIT' | 'CREDIT'
  is_active: boolean
  description?: string
  parent_id?: string | null
}

interface JournalEntry {
  id: string
  journal_number: string
  date: string
  description: string
  reference_type: string
  reference_id?: string
  status: 'DRAFT' | 'POSTED' | 'CANCELLED'
  total_debit: number
  total_credit: number
  created_at: string
}

interface JournalLine {
  account_id: string
  description: string
  debit: number | string
  credit: number | string
}

interface TrialBalanceRow {
  account_code: string
  account_name: string
  account_type?: string
  debit: number
  credit: number
  balance?: number
}

// ─── State: tabs ──────────────────────────────────────────────────────────────

const activeTab = ref<'accounts' | 'journal' | 'reports'>('accounts')
const activeReportTab = ref<'balance-sheet' | 'profit-loss' | 'cash-flow' | 'trial-balance'>('balance-sheet')

// ─── Filters / pagination ─────────────────────────────────────────────────────

const months = [
  { value: 1, label: 'Januari' },
  { value: 2, label: 'Februari' },
  { value: 3, label: 'Maret' },
  { value: 4, label: 'April' },
  { value: 5, label: 'Mei' },
  { value: 6, label: 'Juni' },
  { value: 7, label: 'Juli' },
  { value: 8, label: 'Agustus' },
  { value: 9, label: 'September' },
  { value: 10, label: 'Oktober' },
  { value: 11, label: 'November' },
  { value: 12, label: 'Desember' },
]

const now = new Date()
const selectedMonth = ref(now.getMonth() + 1)
const selectedYear = ref(now.getFullYear())
const yearOptions = computed(() => {
  const y = now.getFullYear()
  return [y - 2, y - 1, y, y + 1]
})

// ─── Accounts (Tab 1) ─────────────────────────────────────────────────────────

const accounts = ref<Account[]>([])
const accountsLoading = ref(false)
const accountSearch = ref('')

const filteredAccounts = computed(() => {
  const q = accountSearch.value.toLowerCase()
  if (!q) return accounts.value
  return accounts.value.filter(
    (a) =>
      a.code.toLowerCase().includes(q) ||
      a.name.toLowerCase().includes(q) ||
      a.type.toLowerCase().includes(q),
  )
})

const fetchAccounts = async () => {
  accountsLoading.value = true
  try {
    const res = await apiClient.get('/accounting/accounts')
    accounts.value = res.data?.data ?? res.data ?? []
  } catch (e: any) {
    toast.error('Gagal memuat bagan akun')
    console.error(e)
  } finally {
    accountsLoading.value = false
  }
}

const seedAccounts = async () => {
  if (!confirm('Seed akun default PSAK? Akun yang sudah ada tidak akan ditimpa.')) return
  try {
    await apiClient.post('/accounting/accounts/seed')
    toast.success('Akun PSAK berhasil di-seed!')
    fetchAccounts()
  } catch (e: any) {
    toast.error(e?.response?.data?.message ?? 'Gagal seed akun PSAK')
  }
}

// Add Account Modal
const showAddAccountModal = ref(false)
const addAccountForm = ref({
  code: '',
  name: '',
  type: 'ASSET' as Account['type'],
  description: '',
  parent_id: '',
})
const addAccountLoading = ref(false)

const openAddAccountModal = () => {
  addAccountForm.value = { code: '', name: '', type: 'ASSET', description: '', parent_id: '' }
  showAddAccountModal.value = true
}

const submitAddAccount = async () => {
  if (!addAccountForm.value.code || !addAccountForm.value.name) {
    toast.error('Kode dan nama akun wajib diisi')
    return
  }
  addAccountLoading.value = true
  try {
    const payload = {
      ...addAccountForm.value,
      parent_id: addAccountForm.value.parent_id || null,
    }
    await apiClient.post('/accounting/accounts', payload)
    toast.success('Akun berhasil ditambahkan!')
    showAddAccountModal.value = false
    fetchAccounts()
  } catch (e: any) {
    toast.error(e?.response?.data?.message ?? 'Gagal menambahkan akun')
  } finally {
    addAccountLoading.value = false
  }
}

// ─── Journal Entries (Tab 2) ──────────────────────────────────────────────────

const journals = ref<JournalEntry[]>([])
const journalLoading = ref(false)
const journalPage = ref(1)
const journalLimit = 15
const journalTotal = ref(0)
const journalTotalPages = computed(() => Math.max(1, Math.ceil(journalTotal.value / journalLimit)))

const fetchJournals = async () => {
  journalLoading.value = true
  try {
    const res = await apiClient.get('/accounting/journal-entries', {
      params: {
        month: selectedMonth.value,
        year: selectedYear.value,
        page: journalPage.value,
        limit: journalLimit,
      },
    })
    const data = res.data?.data ?? res.data ?? {}
    journals.value = data.entries ?? data.items ?? (Array.isArray(data) ? data : [])
    journalTotal.value = data.total ?? journals.value.length
  } catch (e: any) {
    toast.error('Gagal memuat jurnal umum')
    console.error(e)
  } finally {
    journalLoading.value = false
  }
}

// Create Journal Modal
const showJournalModal = ref(false)
const journalForm = ref({
  date: new Date().toISOString().slice(0, 10),
  description: '',
  lines: [] as JournalLine[],
})
const journalSubmitLoading = ref(false)

const addJournalLine = () => {
  journalForm.value.lines.push({ account_id: '', description: '', debit: '', credit: '' })
}

const removeJournalLine = (idx: number) => {
  journalForm.value.lines.splice(idx, 1)
}

const journalTotalDebit = computed(() =>
  journalForm.value.lines.reduce((s, l) => s + (parseFloat(String(l.debit)) || 0), 0),
)
const journalTotalCredit = computed(() =>
  journalForm.value.lines.reduce((s, l) => s + (parseFloat(String(l.credit)) || 0), 0),
)
const journalBalanced = computed(() =>
  journalTotalDebit.value === journalTotalCredit.value && journalTotalDebit.value > 0,
)

const openJournalModal = () => {
  journalForm.value = {
    date: new Date().toISOString().slice(0, 10),
    description: '',
    lines: [
      { account_id: '', description: '', debit: '', credit: '' },
      { account_id: '', description: '', debit: '', credit: '' },
    ],
  }
  showJournalModal.value = true
}

const submitJournal = async () => {
  if (!journalForm.value.description) {
    toast.error('Deskripsi jurnal wajib diisi')
    return
  }
  if (journalForm.value.lines.length < 2) {
    toast.error('Minimal 2 baris jurnal')
    return
  }
  if (!journalBalanced.value) {
    toast.error('Total debit dan kredit harus seimbang')
    return
  }
  const hasEmptyAccount = journalForm.value.lines.some((l) => !l.account_id)
  if (hasEmptyAccount) {
    toast.error('Semua baris harus memiliki akun')
    return
  }

  journalSubmitLoading.value = true
  try {
    const payload = {
      date: journalForm.value.date,
      description: journalForm.value.description,
      lines: journalForm.value.lines.map((l) => ({
        account_id: l.account_id,
        description: l.description,
        debit: parseFloat(String(l.debit)) || 0,
        credit: parseFloat(String(l.credit)) || 0,
      })),
    }
    await apiClient.post('/accounting/journal-entries', payload)
    toast.success('Jurnal berhasil dibuat!')
    showJournalModal.value = false
    fetchJournals()
  } catch (e: any) {
    toast.error(e?.response?.data?.message ?? 'Gagal membuat jurnal')
  } finally {
    journalSubmitLoading.value = false
  }
}

// ─── Financial Reports (Tab 3) ────────────────────────────────────────────────

const reportLoading = ref(false)

const balanceSheet = ref<any>(null)
const fetchBalanceSheet = async () => {
  reportLoading.value = true
  try {
    const res = await apiClient.get('/accounting/reports/balance-sheet', {
      params: { month: selectedMonth.value, year: selectedYear.value },
    })
    balanceSheet.value = res.data?.data ?? res.data
  } catch {
    toast.error('Gagal memuat neraca')
  } finally {
    reportLoading.value = false
  }
}

const profitLoss = ref<any>(null)
const fetchProfitLoss = async () => {
  reportLoading.value = true
  try {
    const res = await apiClient.get('/accounting/reports/profit-loss', {
      params: { month: selectedMonth.value, year: selectedYear.value },
    })
    profitLoss.value = res.data?.data ?? res.data
  } catch {
    toast.error('Gagal memuat laporan laba/rugi')
  } finally {
    reportLoading.value = false
  }
}

const cashFlow = ref<any>(null)
const fetchCashFlow = async () => {
  reportLoading.value = true
  try {
    const res = await apiClient.get('/accounting/reports/cash-flow', {
      params: { month: selectedMonth.value, year: selectedYear.value },
    })
    cashFlow.value = res.data?.data ?? res.data
  } catch {
    toast.error('Gagal memuat arus kas')
  } finally {
    reportLoading.value = false
  }
}

const trialBalance = ref<TrialBalanceRow[]>([])
const trialBalanceTotals = ref({ debit: 0, credit: 0 })
const fetchTrialBalance = async () => {
  reportLoading.value = true
  try {
    const res = await apiClient.get('/accounting/reports/trial-balance', {
      params: { month: selectedMonth.value, year: selectedYear.value },
    })
    const data = res.data?.data ?? res.data ?? {}
    trialBalance.value = data.rows ?? data.accounts ?? (Array.isArray(data) ? data : [])
    trialBalanceTotals.value = {
      debit: data.total_debit ?? trialBalance.value.reduce((s: number, r: TrialBalanceRow) => s + r.debit, 0),
      credit: data.total_credit ?? trialBalance.value.reduce((s: number, r: TrialBalanceRow) => s + r.credit, 0),
    }
  } catch {
    toast.error('Gagal memuat neraca saldo')
  } finally {
    reportLoading.value = false
  }
}

const loadCurrentReport = () => {
  switch (activeReportTab.value) {
    case 'balance-sheet': fetchBalanceSheet(); break
    case 'profit-loss': fetchProfitLoss(); break
    case 'cash-flow': fetchCashFlow(); break
    case 'trial-balance': fetchTrialBalance(); break
  }
}

// ─── Watchers ─────────────────────────────────────────────────────────────────

watch(activeTab, (tab) => {
  if (tab === 'journal') fetchJournals()
  if (tab === 'reports') loadCurrentReport()
})

watch(activeReportTab, () => loadCurrentReport())

watch([selectedMonth, selectedYear], () => {
  fetchTrialBalance()
  if (activeTab.value === 'journal') { journalPage.value = 1; fetchJournals() }
  if (activeTab.value === 'reports') loadCurrentReport()
})

watch(journalPage, () => fetchJournals())

// ─── Lifecycle ────────────────────────────────────────────────────────────────

onMounted(() => {
  fetchAccounts()
  fetchTrialBalance()
})

// ─── Helpers ──────────────────────────────────────────────────────────────────

const expandedJournalId = ref<string | null>(null)
const expandedJournalLines = ref<any[]>([])
const expandLoading = ref(false)

const toggleExpand = async (id: string) => {
  if (expandedJournalId.value === id) {
    expandedJournalId.value = null
    expandedJournalLines.value = []
    return
  }
  expandedJournalId.value = id
  expandLoading.value = true
  try {
    const res = await apiClient.get(`/accounting/journal-entries/${id}`)
    const entry = res.data?.data ?? res.data
    expandedJournalLines.value = entry.lines ?? []
  } catch { expandedJournalLines.value = [] }
  finally { expandLoading.value = false }
}

const getAccountName = (accountId: string) => {
  const a = accounts.value.find(acc => acc.id === accountId)
  return a ? `${a.code} — ${a.name}` : accountId
}

const postJournal = async (id: string) => {
  if (!confirm('Post jurnal ini? Tidak bisa dibatalkan setelah POSTED.')) return
  try {
    await apiClient.post(`/accounting/journal-entries/${id}/post`)
    toast.success('Jurnal berhasil diposting!')
    fetchJournals()
    fetchTrialBalance()
  } catch (e: any) {
    toast.error(e?.response?.data?.error ?? 'Gagal posting jurnal')
  }
}

const cancelJournal = async (id: string) => {
  if (!confirm('Batalkan jurnal ini?')) return
  try {
    await apiClient.post(`/accounting/journal-entries/${id}/cancel`)
    toast.success('Jurnal berhasil dibatalkan')
    fetchJournals()
  } catch (e: any) {
    toast.error(e?.response?.data?.error ?? 'Gagal membatalkan jurnal')
  }
}

const kpiBalances = computed(() => {
  const totals = { ASSET: 0, LIABILITY: 0, EQUITY: 0, REVENUE: 0, EXPENSE: 0 }
  trialBalance.value.forEach(r => {
    if (r.account_type && r.account_type in totals) {
      totals[r.account_type as keyof typeof totals] += (r.balance || 0)
    }
  })
  return totals
})

const formatCurrency = (val: number | undefined | null) =>
  new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(val ?? 0)

const formatDate = (d: string) =>
  d ? new Date(d).toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' }) : '-'

const accountTypeLabel: Record<string, string> = {
  ASSET: 'Aset',
  LIABILITY: 'Liabilitas',
  EQUITY: 'Ekuitas',
  REVENUE: 'Pendapatan',
  EXPENSE: 'Beban',
}

const accountTypeColor: Record<string, string> = {
  ASSET: 'bg-blue-50 text-blue-700 border-blue-100',
  LIABILITY: 'bg-orange-50 text-orange-700 border-orange-100',
  EQUITY: 'bg-violet-50 text-violet-700 border-violet-100',
  REVENUE: 'bg-emerald-50 text-emerald-700 border-emerald-100',
  EXPENSE: 'bg-rose-50 text-rose-700 border-rose-100',
}

const journalStatusColor: Record<string, string> = {
  POSTED: 'bg-emerald-50 text-emerald-700 border-emerald-100',
  DRAFT: 'bg-amber-50 text-amber-700 border-amber-100',
  CANCELLED: 'bg-red-50 text-red-600 border-red-100',
}

const journalStatusLabel: Record<string, string> = {
  POSTED: 'Diposting',
  DRAFT: 'Draf',
  CANCELLED: 'Dibatalkan',
}

const accountTypeCounts = computed(() => {
  const counts: Record<string, number> = { ASSET: 0, LIABILITY: 0, EQUITY: 0, REVENUE: 0, EXPENSE: 0 }
  accounts.value.forEach((a) => { counts[a.type] = (counts[a.type] || 0) + 1 })
  return counts
})
</script>

<template>
  <div class="space-y-6 pb-10">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-3xl font-extrabold text-slate-900 tracking-tight">Akuntansi</h1>
        <p class="text-[15px] font-medium text-slate-500 mt-1.5 flex items-center gap-2">
          <BookOpen class="w-4 h-4 text-primary" />
          Bagan akun, jurnal umum, dan laporan keuangan perusahaan
        </p>
      </div>
      <div class="flex items-center gap-2">
        <select
          v-model="selectedMonth"
          class="bg-white border border-slate-200 px-3 py-2 rounded-xl text-sm font-semibold text-slate-700 shadow-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
        >
          <option v-for="m in months" :key="m.value" :value="m.value">{{ m.label }}</option>
        </select>
        <select
          v-model="selectedYear"
          class="bg-white border border-slate-200 px-3 py-2 rounded-xl text-sm font-semibold text-slate-700 shadow-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
        >
          <option v-for="y in yearOptions" :key="y" :value="y">{{ y }}</option>
        </select>
      </div>
    </div>

    <!-- Tab Navigation -->
    <div class="flex gap-1 bg-slate-100/70 p-1 rounded-2xl w-fit">
      <button
        v-for="tab in [
          { key: 'accounts', label: 'Bagan Akun', icon: Layers },
          { key: 'journal', label: 'Jurnal Umum', icon: FileText },
          { key: 'reports', label: 'Laporan Keuangan', icon: BarChart3 },
        ]"
        :key="tab.key"
        @click="activeTab = tab.key as any"
        class="flex items-center gap-2 px-5 py-2.5 rounded-xl text-sm font-bold transition-all duration-200"
        :class="
          activeTab === tab.key
            ? 'bg-white text-slate-900 shadow-sm'
            : 'text-slate-500 hover:text-slate-700'
        "
      >
        <component :is="tab.icon" class="w-4 h-4" />
        {{ tab.label }}
      </button>
    </div>

    <!-- ══════════════════════════════════════════════
         TAB 1: BAGAN AKUN
    ══════════════════════════════════════════════ -->
    <div v-if="activeTab === 'accounts'" class="space-y-5">
      <!-- KPI Summary Cards -->
      <div class="grid grid-cols-2 md:grid-cols-5 gap-4">
        <div class="bg-white p-4 rounded-2xl border border-slate-200/60 shadow-sm flex flex-col">
          <span class="text-xs font-bold text-slate-400 mb-1">💼 Total Aset</span>
          <span class="text-lg font-extrabold text-blue-700">{{ formatCurrency(kpiBalances.ASSET) }}</span>
        </div>
        <div class="bg-white p-4 rounded-2xl border border-slate-200/60 shadow-sm flex flex-col">
          <span class="text-xs font-bold text-slate-400 mb-1">⚖️ Total Liabilitas</span>
          <span class="text-lg font-extrabold text-orange-600">{{ formatCurrency(kpiBalances.LIABILITY) }}</span>
        </div>
        <div class="bg-white p-4 rounded-2xl border border-slate-200/60 shadow-sm flex flex-col">
          <span class="text-xs font-bold text-slate-400 mb-1">🏦 Total Ekuitas</span>
          <span class="text-lg font-extrabold text-violet-600">{{ formatCurrency(kpiBalances.EQUITY) }}</span>
        </div>
        <div class="bg-white p-4 rounded-2xl border border-slate-200/60 shadow-sm flex flex-col">
          <span class="text-xs font-bold text-slate-400 mb-1">📈 Pendapatan</span>
          <span class="text-lg font-extrabold text-emerald-600">{{ formatCurrency(kpiBalances.REVENUE) }}</span>
        </div>
        <div class="bg-white p-4 rounded-2xl border border-slate-200/60 shadow-sm flex flex-col">
          <span class="text-xs font-bold text-slate-400 mb-1">📉 Beban</span>
          <span class="text-lg font-extrabold text-rose-600">{{ formatCurrency(kpiBalances.EXPENSE) }}</span>
        </div>
      </div>

      <!-- Summary chips -->
      <div class="flex flex-wrap gap-3">
        <div
          v-for="(count, type) in accountTypeCounts"
          :key="type"
          class="flex items-center gap-2 px-4 py-2 rounded-2xl border text-xs font-bold"
          :class="accountTypeColor[type as string]"
        >
          <span>{{ accountTypeLabel[type as string] }}</span>
          <span class="bg-white/60 px-1.5 py-0.5 rounded-lg">{{ count }}</span>
        </div>
      </div>

      <!-- Toolbar -->
      <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div class="relative">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input
            v-model="accountSearch"
            type="text"
            placeholder="Cari kode atau nama akun..."
            class="pl-9 pr-4 py-2.5 rounded-xl border border-slate-200 bg-white text-sm font-medium text-slate-700 focus:outline-none focus:ring-2 focus:ring-primary/20 w-72"
          />
        </div>
        <div class="flex items-center gap-2">
          <button
            @click="fetchAccounts"
            class="p-2.5 bg-white border border-slate-200 rounded-xl text-slate-500 hover:bg-slate-50 shadow-sm hover:text-primary transition-all"
          >
            <RefreshCw class="w-4 h-4" />
          </button>
          <button
            @click="seedAccounts"
            class="flex items-center gap-2 px-4 py-2.5 bg-slate-900 text-white rounded-xl text-sm font-bold hover:bg-slate-800 transition-all shadow-sm"
          >
            <Database class="w-4 h-4" />
            Seed PSAK Default
          </button>
          <button
            @click="openAddAccountModal"
            class="flex items-center gap-2 px-4 py-2.5 bg-primary text-white rounded-xl text-sm font-bold hover:bg-primary/90 transition-all shadow-sm shadow-primary/20"
          >
            <Plus class="w-4 h-4" />
            Tambah Akun
          </button>
        </div>
      </div>

      <!-- Table -->
      <div class="bg-white rounded-3xl border border-slate-200/60 shadow-[0_4px_20px_rgba(0,0,0,0.02)] overflow-hidden">
        <div v-if="accountsLoading" class="flex items-center justify-center py-20">
          <span class="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin" />
        </div>
        <div v-else-if="filteredAccounts.length === 0" class="flex flex-col items-center justify-center py-20 text-slate-400">
          <Layers class="w-14 h-14 mb-4 opacity-20" />
          <p class="font-bold text-slate-500">Belum ada akun</p>
          <p class="text-sm mt-1">Klik "Tambah Akun" atau "Seed PSAK Default" untuk memulai</p>
        </div>
        <div v-else class="overflow-x-auto">
          <table class="w-full">
            <thead>
              <tr class="border-b border-slate-100 bg-slate-50/50">
                <th class="text-left px-6 py-4 text-xs font-extrabold text-slate-400 uppercase tracking-wider">Kode</th>
                <th class="text-left px-6 py-4 text-xs font-extrabold text-slate-400 uppercase tracking-wider">Nama Akun</th>
                <th class="text-left px-6 py-4 text-xs font-extrabold text-slate-400 uppercase tracking-wider">Tipe</th>
                <th class="text-left px-6 py-4 text-xs font-extrabold text-slate-400 uppercase tracking-wider">Saldo Normal</th>
                <th class="text-left px-6 py-4 text-xs font-extrabold text-slate-400 uppercase tracking-wider">Status</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="account in filteredAccounts"
                :key="account.id"
                class="border-b border-slate-50 hover:bg-slate-50/50 transition-colors"
              >
                <td class="px-6 py-4">
                  <span class="font-mono text-sm font-bold text-slate-700">{{ account.code }}</span>
                </td>
                <td class="px-6 py-4">
                  <div>
                    <p class="text-sm font-semibold text-slate-800">{{ account.name }}</p>
                    <p v-if="account.description" class="text-xs text-slate-400 mt-0.5">{{ account.description }}</p>
                  </div>
                </td>
                <td class="px-6 py-4">
                  <span
                    class="inline-flex items-center px-2.5 py-1 rounded-lg text-xs font-bold border"
                    :class="accountTypeColor[account.type]"
                  >
                    {{ accountTypeLabel[account.type] }}
                  </span>
                </td>
                <td class="px-6 py-4">
                  <span
                    class="inline-flex items-center gap-1 text-xs font-bold"
                    :class="account.normal_balance === 'DEBIT' ? 'text-blue-600' : 'text-rose-600'"
                  >
                    <component :is="account.normal_balance === 'DEBIT' ? TrendingUp : TrendingDown" class="w-3.5 h-3.5" />
                    {{ account.normal_balance === 'DEBIT' ? 'Debit' : 'Kredit' }}
                  </span>
                </td>
                <td class="px-6 py-4">
                  <span
                    class="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg text-xs font-bold border"
                    :class="account.is_active ? 'bg-emerald-50 text-emerald-700 border-emerald-100' : 'bg-slate-100 text-slate-500 border-slate-200'"
                  >
                    <component :is="account.is_active ? CheckCircle : AlertCircle" class="w-3.5 h-3.5" />
                    {{ account.is_active ? 'Aktif' : 'Nonaktif' }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- ══════════════════════════════════════════════
         TAB 2: JURNAL UMUM
    ══════════════════════════════════════════════ -->
    <div v-else-if="activeTab === 'journal'" class="space-y-5">
      <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <p class="text-sm font-semibold text-slate-500">
          {{ months.find((m) => m.value === selectedMonth)?.label }} {{ selectedYear }} —
          <span class="text-slate-800 font-bold">{{ journalTotal }} jurnal</span>
        </p>
        <div class="flex items-center gap-2">
          <button
            @click="fetchJournals"
            class="p-2.5 bg-white border border-slate-200 rounded-xl text-slate-500 hover:bg-slate-50 shadow-sm hover:text-primary transition-all"
          >
            <RefreshCw class="w-4 h-4" />
          </button>
          <button
            @click="openJournalModal"
            class="flex items-center gap-2 px-4 py-2.5 bg-primary text-white rounded-xl text-sm font-bold hover:bg-primary/90 transition-all shadow-sm shadow-primary/20"
          >
            <Plus class="w-4 h-4" />
            Buat Jurnal Manual
          </button>
        </div>
      </div>

      <div class="bg-white rounded-3xl border border-slate-200/60 shadow-[0_4px_20px_rgba(0,0,0,0.02)] overflow-hidden">
        <div v-if="journalLoading" class="flex items-center justify-center py-20">
          <span class="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin" />
        </div>
        <div v-else-if="journals.length === 0" class="flex flex-col items-center justify-center py-20 text-slate-400">
          <FileText class="w-14 h-14 mb-4 opacity-20" />
          <p class="font-bold text-slate-500">Belum ada jurnal</p>
          <p class="text-sm mt-1">Pilih bulan/tahun lain atau buat jurnal manual</p>
        </div>
        <div v-else class="overflow-x-auto">
          <table class="w-full">
            <thead>
              <tr class="border-b border-slate-100 bg-slate-50/50">
                <th class="w-10 px-4 py-4 text-center"></th>
                <th class="text-left px-6 py-4 text-xs font-extrabold text-slate-400 uppercase tracking-wider">No. Jurnal</th>
                <th class="text-left px-6 py-4 text-xs font-extrabold text-slate-400 uppercase tracking-wider">Tanggal</th>
                <th class="text-left px-6 py-4 text-xs font-extrabold text-slate-400 uppercase tracking-wider">Deskripsi</th>
                <th class="text-left px-6 py-4 text-xs font-extrabold text-slate-400 uppercase tracking-wider">Tipe Referensi</th>
                <th class="text-right px-6 py-4 text-xs font-extrabold text-slate-400 uppercase tracking-wider">Debit</th>
                <th class="text-right px-6 py-4 text-xs font-extrabold text-slate-400 uppercase tracking-wider">Kredit</th>
                <th class="text-left px-6 py-4 text-xs font-extrabold text-slate-400 uppercase tracking-wider">Status</th>
                <th class="text-right px-6 py-4 text-xs font-extrabold text-slate-400 uppercase tracking-wider">Aksi</th>
              </tr>
            </thead>
            <tbody>
              <template v-for="journal in journals" :key="journal.id">
                <tr
                  class="border-b border-slate-50 hover:bg-slate-50/50 transition-colors"
                >
                  <td class="px-4 py-4 text-center">
                    <button @click="toggleExpand(journal.id)" class="p-1 rounded hover:bg-slate-200 text-slate-500 transition-colors">
                      <ChevronDown v-if="expandedJournalId === journal.id" class="w-4 h-4" />
                      <ChevronRight v-else class="w-4 h-4" />
                    </button>
                  </td>
                  <td class="px-6 py-4">
                    <span class="font-mono text-sm font-bold text-primary">{{ journal.journal_number }}</span>
                  </td>
                  <td class="px-6 py-4 text-sm font-semibold text-slate-600">{{ formatDate(journal.date) }}</td>
                  <td class="px-6 py-4">
                    <p class="text-sm font-semibold text-slate-800 max-w-xs truncate">{{ journal.description }}</p>
                  </td>
                  <td class="px-6 py-4">
                    <span class="text-xs font-bold text-slate-500 bg-slate-100 px-2.5 py-1 rounded-lg">
                      {{ journal.reference_type || 'MANUAL' }}
                    </span>
                  </td>
                  <td class="px-6 py-4 text-right text-sm font-semibold text-blue-700">
                    {{ formatCurrency(journal.total_debit) }}
                  </td>
                  <td class="px-6 py-4 text-right text-sm font-semibold text-rose-600">
                    {{ formatCurrency(journal.total_credit) }}
                  </td>
                  <td class="px-6 py-4">
                    <span
                      class="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg text-xs font-bold border"
                      :class="journalStatusColor[journal.status] ?? 'bg-slate-50 text-slate-600 border-slate-200'"
                    >
                      <component
                        :is="journal.status === 'POSTED' ? CheckCircle : journal.status === 'DRAFT' ? Clock : AlertCircle"
                        class="w-3.5 h-3.5"
                      />
                      {{ journalStatusLabel[journal.status] ?? journal.status }}
                    </span>
                  </td>
                  <td class="px-6 py-4 text-right">
                    <div class="flex items-center justify-end gap-2" v-if="journal.status === 'DRAFT'">
                      <button @click="postJournal(journal.id)" class="px-2.5 py-1.5 bg-blue-50 text-blue-600 hover:bg-blue-100 rounded-lg text-xs font-bold flex items-center gap-1 transition-colors">
                        ▶ Post
                      </button>
                      <button @click="cancelJournal(journal.id)" class="px-2.5 py-1.5 border border-rose-200 text-rose-500 hover:bg-rose-50 rounded-lg text-xs font-bold flex items-center gap-1 transition-colors">
                        ✕ Batal
                      </button>
                    </div>
                  </td>
                </tr>
                <tr v-if="expandedJournalId === journal.id" class="bg-slate-50">
                  <td colspan="9" class="px-8 py-5 border-b border-slate-100">
                    <div v-if="expandLoading" class="flex items-center justify-center py-4">
                      <span class="w-5 h-5 border-2 border-primary border-t-transparent rounded-full animate-spin" />
                    </div>
                    <div v-else class="border border-slate-200 rounded-xl overflow-hidden bg-white shadow-sm">
                      <table class="w-full">
                        <thead class="bg-slate-50/50">
                          <tr>
                            <th class="text-left px-4 py-2.5 text-xs font-extrabold text-slate-500">Akun</th>
                            <th class="text-right px-4 py-2.5 text-xs font-extrabold text-slate-500">Debit</th>
                            <th class="text-right px-4 py-2.5 text-xs font-extrabold text-slate-500">Kredit</th>
                          </tr>
                        </thead>
                        <tbody>
                          <tr v-for="(line, lIdx) in expandedJournalLines" :key="lIdx" class="border-t border-slate-100">
                            <td class="px-4 py-2.5 text-sm font-medium text-slate-700">{{ getAccountName(line.account_id) }}</td>
                            <td class="px-4 py-2.5 text-sm font-semibold text-blue-700 text-right">{{ formatCurrency(Number(line.debit_amount || line.debit || 0)) }}</td>
                            <td class="px-4 py-2.5 text-sm font-semibold text-rose-600 text-right">{{ formatCurrency(Number(line.credit_amount || line.credit || 0)) }}</td>
                          </tr>
                        </tbody>
                        <tfoot class="border-t-2 border-slate-100 bg-slate-50/30">
                          <tr>
                            <td class="px-4 py-3 text-xs font-extrabold text-slate-500 uppercase tracking-wider text-right">Total</td>
                            <td class="px-4 py-3 text-sm font-extrabold text-blue-700 text-right">
                              {{ formatCurrency(expandedJournalLines.reduce((s, l) => s + Number(l.debit_amount || l.debit || 0), 0)) }}
                            </td>
                            <td class="px-4 py-3 text-sm font-extrabold text-rose-600 text-right">
                              {{ formatCurrency(expandedJournalLines.reduce((s, l) => s + Number(l.credit_amount || l.credit || 0), 0)) }}
                            </td>
                          </tr>
                        </tfoot>
                      </table>
                    </div>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>

          <!-- Pagination -->
          <div class="flex items-center justify-between px-6 py-4 border-t border-slate-100">
            <p class="text-xs font-semibold text-slate-400">
              Halaman {{ journalPage }} dari {{ journalTotalPages }}
            </p>
            <div class="flex items-center gap-1">
              <button
                @click="journalPage--"
                :disabled="journalPage <= 1"
                class="p-2 rounded-lg border border-slate-200 text-slate-500 hover:bg-slate-50 disabled:opacity-40 disabled:cursor-not-allowed transition-all"
              >
                <ChevronLeft class="w-4 h-4" />
              </button>
              <span class="px-3 py-1.5 text-xs font-bold text-slate-700">{{ journalPage }}</span>
              <button
                @click="journalPage++"
                :disabled="journalPage >= journalTotalPages"
                class="p-2 rounded-lg border border-slate-200 text-slate-500 hover:bg-slate-50 disabled:opacity-40 disabled:cursor-not-allowed transition-all"
              >
                <ChevronRight class="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- ══════════════════════════════════════════════
         TAB 3: LAPORAN KEUANGAN
    ══════════════════════════════════════════════ -->
    <div v-else-if="activeTab === 'reports'" class="space-y-5">
      <!-- Report sub-tabs -->
      <div class="flex gap-1 border-b border-slate-200">
        <button
          v-for="sub in [
            { key: 'balance-sheet', label: 'Neraca', icon: Scale },
            { key: 'profit-loss', label: 'Laba / Rugi', icon: TrendingUp },
            { key: 'cash-flow', label: 'Arus Kas', icon: Activity },
            { key: 'trial-balance', label: 'Neraca Saldo', icon: BarChart3 },
          ]"
          :key="sub.key"
          @click="activeReportTab = sub.key as any"
          class="flex items-center gap-2 px-5 py-3 text-sm font-bold border-b-2 transition-all duration-200 -mb-px"
          :class="
            activeReportTab === sub.key
              ? 'border-primary text-primary'
              : 'border-transparent text-slate-500 hover:text-slate-700'
          "
        >
          <component :is="sub.icon" class="w-4 h-4" />
          {{ sub.label }}
        </button>
      </div>

      <div class="flex items-center justify-between">
        <p class="text-sm font-semibold text-slate-500">
          Periode:
          <span class="text-slate-800 font-bold">
            {{ months.find((m) => m.value === selectedMonth)?.label }} {{ selectedYear }}
          </span>
        </p>
        <button
          @click="loadCurrentReport"
          class="flex items-center gap-2 px-4 py-2 bg-white border border-slate-200 rounded-xl text-sm font-bold text-slate-600 hover:bg-slate-50 shadow-sm transition-all"
        >
          <RefreshCw class="w-4 h-4" />
          Muat Ulang
        </button>
      </div>

      <!-- Loading -->
      <div v-if="reportLoading" class="flex items-center justify-center py-24">
        <span class="w-10 h-10 border-4 border-primary border-t-transparent rounded-full animate-spin" />
      </div>

      <!-- NERACA -->
      <div v-else-if="activeReportTab === 'balance-sheet'" class="space-y-6">
        <div v-if="!balanceSheet || (balanceSheet.total_assets === 0 && balanceSheet.total_liabilities === 0 && balanceSheet.total_equity === 0)" class="flex flex-col items-center justify-center py-16 text-slate-400">
          <BarChart3 class="w-16 h-16 mx-auto mb-4 opacity-20" />
          <p class="font-bold text-slate-500">Belum ada data jurnal diposting</p>
          <p class="text-sm mt-1 max-w-sm mx-auto text-center">Buat jurnal manual di tab "Jurnal Umum" dan klik Post, atau tunggu transaksi dari modul Sales/Payroll yang otomatis tercatat.</p>
        </div>
        <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <!-- Assets -->
          <div class="bg-white rounded-3xl border border-slate-200/60 shadow-[0_4px_20px_rgba(0,0,0,0.02)] overflow-hidden">
            <div class="px-6 py-4 border-b border-slate-100 bg-blue-50/40 flex items-center gap-3">
              <div class="w-8 h-8 bg-blue-100 rounded-xl flex items-center justify-center">
                <TrendingUp class="w-4 h-4 text-blue-600" />
              </div>
              <h3 class="text-base font-extrabold text-slate-800">Aset</h3>
            </div>
            <div class="p-6 space-y-2">
              <template v-if="balanceSheet.assets">
                <template v-for="(section, key) in balanceSheet.assets" :key="String(key)">
                  <p class="text-xs font-extrabold text-slate-400 uppercase tracking-wider mt-4 first:mt-0">{{ key }}</p>
                  <template v-if="typeof section === 'object' && section !== null">
                    <div
                      v-for="(amount, name) in section as Record<string, number>"
                      :key="String(name)"
                      class="flex items-center justify-between py-2 border-b border-slate-50 last:border-0"
                    >
                      <span class="text-sm font-medium text-slate-600">{{ name }}</span>
                      <span class="text-sm font-bold text-slate-800">{{ formatCurrency(Number(amount)) }}</span>
                    </div>
                  </template>
                </template>
              </template>
              <div class="pt-4 border-t-2 border-blue-100 flex items-center justify-between">
                <span class="text-sm font-extrabold text-blue-700">Total Aset</span>
                <span class="text-base font-extrabold text-blue-700">{{ formatCurrency(balanceSheet.total_assets ?? 0) }}</span>
              </div>
            </div>
          </div>

          <!-- Liabilities + Equity -->
          <div class="bg-white rounded-3xl border border-slate-200/60 shadow-[0_4px_20px_rgba(0,0,0,0.02)] overflow-hidden">
            <div class="px-6 py-4 border-b border-slate-100 bg-orange-50/40 flex items-center gap-3">
              <div class="w-8 h-8 bg-orange-100 rounded-xl flex items-center justify-center">
                <TrendingDown class="w-4 h-4 text-orange-600" />
              </div>
              <h3 class="text-base font-extrabold text-slate-800">Liabilitas & Ekuitas</h3>
            </div>
            <div class="p-6 space-y-2">
              <template v-if="balanceSheet.liabilities">
                <p class="text-xs font-extrabold text-slate-400 uppercase tracking-wider">Liabilitas</p>
                <template v-if="typeof balanceSheet.liabilities === 'object'">
                  <div
                    v-for="(amount, name) in balanceSheet.liabilities as Record<string, number>"
                    :key="String(name)"
                    class="flex items-center justify-between py-2 border-b border-slate-50 last:border-0"
                  >
                    <span class="text-sm font-medium text-slate-600">{{ name }}</span>
                    <span class="text-sm font-bold text-slate-800">{{ formatCurrency(Number(amount)) }}</span>
                  </div>
                </template>
                <div class="flex items-center justify-between pt-2 pb-4">
                  <span class="text-xs font-extrabold text-orange-600">Total Liabilitas</span>
                  <span class="text-sm font-extrabold text-orange-600">{{ formatCurrency(balanceSheet.total_liabilities ?? 0) }}</span>
                </div>
              </template>
              <template v-if="balanceSheet.equity">
                <p class="text-xs font-extrabold text-slate-400 uppercase tracking-wider">Ekuitas</p>
                <template v-if="typeof balanceSheet.equity === 'object'">
                  <div
                    v-for="(amount, name) in balanceSheet.equity as Record<string, number>"
                    :key="String(name)"
                    class="flex items-center justify-between py-2 border-b border-slate-50 last:border-0"
                  >
                    <span class="text-sm font-medium text-slate-600">{{ name }}</span>
                    <span class="text-sm font-bold text-slate-800">{{ formatCurrency(Number(amount)) }}</span>
                  </div>
                </template>
                <div class="flex items-center justify-between pt-2">
                  <span class="text-xs font-extrabold text-violet-600">Total Ekuitas</span>
                  <span class="text-sm font-extrabold text-violet-600">{{ formatCurrency(balanceSheet.total_equity ?? 0) }}</span>
                </div>
              </template>
              <div class="pt-4 border-t-2 border-orange-100 flex items-center justify-between">
                <span class="text-sm font-extrabold text-orange-700">Total Liabilitas + Ekuitas</span>
                <span class="text-base font-extrabold text-orange-700">
                  {{ formatCurrency((balanceSheet.total_liabilities ?? 0) + (balanceSheet.total_equity ?? 0)) }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- LAPORAN LABA/RUGI -->
      <div v-else-if="activeReportTab === 'profit-loss'">
        <div v-if="!profitLoss || (profitLoss.total_revenue === 0 && profitLoss.total_cogs === 0 && profitLoss.total_payroll === 0 && profitLoss.total_expenses === 0)" class="flex flex-col items-center justify-center py-16 text-slate-400">
          <BarChart3 class="w-16 h-16 mx-auto mb-4 opacity-20" />
          <p class="font-bold text-slate-500">Belum ada data jurnal diposting</p>
          <p class="text-sm mt-1 max-w-sm mx-auto text-center">Buat jurnal manual di tab "Jurnal Umum" dan klik Post, atau tunggu transaksi dari modul Sales/Payroll yang otomatis tercatat.</p>
        </div>
        <div v-else class="bg-white rounded-3xl border border-slate-200/60 shadow-[0_4px_20px_rgba(0,0,0,0.02)] overflow-hidden">
          <div class="px-6 py-4 border-b border-slate-100 bg-[#0f172a] text-white flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 bg-white/10 rounded-xl flex items-center justify-center">
                <TrendingUp class="w-4 h-4 text-white" />
              </div>
              <h3 class="text-base font-extrabold">Laporan Laba / Rugi</h3>
            </div>
            <span class="text-xs font-bold text-slate-300">
              {{ months.find((m) => m.value === selectedMonth)?.label }} {{ selectedYear }}
            </span>
          </div>
          <div class="p-8 max-w-2xl mx-auto space-y-4 font-mono">
            <div class="flex items-center justify-between py-2 border-b border-slate-100">
              <span class="text-sm font-semibold text-slate-700">Pendapatan Penjualan:</span>
              <span class="text-sm font-bold text-slate-800">{{ formatCurrency(profitLoss.total_revenue) }}</span>
            </div>
            <div class="flex items-center justify-between py-2">
              <span class="text-sm font-semibold text-slate-700">Harga Pokok Penjualan:</span>
              <span class="text-sm font-bold text-slate-800">({{ formatCurrency(profitLoss.total_cogs) }})</span>
            </div>
            <div class="border-t-2 border-slate-300 my-2"></div>
            <div class="flex items-center justify-between py-2">
              <span class="text-sm font-bold text-slate-800">Laba Kotor:</span>
              <span class="text-sm font-bold text-slate-800">{{ formatCurrency(profitLoss.gross_profit) }}</span>
            </div>
            <div class="border-t-2 border-slate-300 my-2"></div>
            <div class="flex items-center justify-between py-2">
              <span class="text-sm font-semibold text-slate-700">Beban Gaji & Tunjangan:</span>
              <span class="text-sm font-bold text-slate-800">({{ formatCurrency(profitLoss.total_payroll) }})</span>
            </div>
            <div class="flex items-center justify-between py-2">
              <span class="text-sm font-semibold text-slate-700">Beban Operasional Lain:</span>
              <span class="text-sm font-bold text-slate-800">({{ formatCurrency(profitLoss.total_expenses) }})</span>
            </div>
            <div class="border-t-2 border-slate-300 my-2"></div>
            <div class="flex items-center justify-between py-2">
              <span class="text-base font-extrabold text-slate-800">Laba Bersih:</span>
              <span class="text-base font-extrabold" :class="(profitLoss.net_profit ?? 0) >= 0 ? 'text-emerald-700' : 'text-red-600'">
                {{ formatCurrency(profitLoss.net_profit) }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- ARUS KAS -->
      <div v-else-if="activeReportTab === 'cash-flow'">
        <div v-if="!cashFlow || (cashFlow.net_operating === 0 && cashFlow.net_investing === 0 && cashFlow.net_financing === 0)" class="flex flex-col items-center justify-center py-16 text-slate-400">
          <BarChart3 class="w-16 h-16 mx-auto mb-4 opacity-20" />
          <p class="font-bold text-slate-500">Belum ada data jurnal diposting</p>
          <p class="text-sm mt-1 max-w-sm mx-auto text-center">Buat jurnal manual di tab "Jurnal Umum" dan klik Post, atau tunggu transaksi dari modul Sales/Payroll yang otomatis tercatat.</p>
        </div>
        <div v-else class="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div
            v-for="section in [
              { key: 'operating', label: 'Aktivitas Operasi', colorClass: 'bg-blue-50/40', dotClass: 'bg-blue-100', iconClass: 'text-blue-600', netKey: 'net_operating' },
              { key: 'investing', label: 'Aktivitas Investasi', colorClass: 'bg-violet-50/40', dotClass: 'bg-violet-100', iconClass: 'text-violet-600', netKey: 'net_investing' },
              { key: 'financing', label: 'Aktivitas Pendanaan', colorClass: 'bg-emerald-50/40', dotClass: 'bg-emerald-100', iconClass: 'text-emerald-600', netKey: 'net_financing' },
            ]"
            :key="section.key"
            class="bg-white rounded-3xl border border-slate-200/60 shadow-[0_4px_20px_rgba(0,0,0,0.02)] overflow-hidden"
          >
            <div class="px-5 py-4 border-b border-slate-100 flex items-center gap-3" :class="section.colorClass">
              <div class="w-7 h-7 rounded-lg flex items-center justify-center" :class="section.dotClass">
                <Activity class="w-3.5 h-3.5" :class="section.iconClass" />
              </div>
              <h3 class="text-sm font-extrabold text-slate-800">{{ section.label }}</h3>
            </div>
            <div class="p-5 space-y-2">
              <template v-if="cashFlow[section.key] && typeof cashFlow[section.key] === 'object'">
                <div
                  v-for="(amount, name) in cashFlow[section.key] as Record<string, number>"
                  :key="String(name)"
                  class="flex items-center justify-between py-2 border-b border-slate-50"
                >
                  <span class="text-xs font-medium text-slate-600">{{ name }}</span>
                  <span
                    class="text-xs font-bold"
                    :class="Number(amount) >= 0 ? 'text-emerald-700' : 'text-rose-600'"
                  >
                    {{ formatCurrency(Number(amount)) }}
                  </span>
                </div>
              </template>
              <div class="pt-3 flex items-center justify-between">
                <span class="text-xs font-extrabold text-slate-600">Net Arus Kas</span>
                <span
                  class="text-sm font-extrabold"
                  :class="Number(cashFlow[section.netKey] ?? 0) >= 0 ? 'text-emerald-700' : 'text-rose-600'"
                >
                  {{ formatCurrency(Number(cashFlow[section.netKey] ?? 0)) }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- NERACA SALDO -->
      <div v-else-if="activeReportTab === 'trial-balance'">
        <div v-if="trialBalance.length === 0 || (trialBalanceTotals.debit === 0 && trialBalanceTotals.credit === 0)" class="flex flex-col items-center justify-center py-16 text-slate-400">
          <BarChart3 class="w-16 h-16 mx-auto mb-4 opacity-20" />
          <p class="font-bold text-slate-500">Belum ada data jurnal diposting</p>
          <p class="text-sm mt-1 max-w-sm mx-auto text-center">Buat jurnal manual di tab "Jurnal Umum" dan klik Post, atau tunggu transaksi dari modul Sales/Payroll yang otomatis tercatat.</p>
        </div>
        <div v-else class="bg-white rounded-3xl border border-slate-200/60 shadow-[0_4px_20px_rgba(0,0,0,0.02)] overflow-hidden">
          <div class="overflow-x-auto">
            <table class="w-full">
              <thead>
                <tr class="border-b border-slate-100 bg-slate-50/50">
                  <th class="text-left px-6 py-4 text-xs font-extrabold text-slate-400 uppercase tracking-wider">Kode</th>
                  <th class="text-left px-6 py-4 text-xs font-extrabold text-slate-400 uppercase tracking-wider">Nama Akun</th>
                  <th class="text-right px-6 py-4 text-xs font-extrabold text-slate-400 uppercase tracking-wider">Debit</th>
                  <th class="text-right px-6 py-4 text-xs font-extrabold text-slate-400 uppercase tracking-wider">Kredit</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="row in trialBalance"
                  :key="row.account_code"
                  class="border-b border-slate-50 hover:bg-slate-50/50 transition-colors"
                >
                  <td class="px-6 py-3.5">
                    <span class="font-mono text-sm font-bold text-slate-600">{{ row.account_code }}</span>
                  </td>
                  <td class="px-6 py-3.5 text-sm font-medium text-slate-700">{{ row.account_name }}</td>
                  <td class="px-6 py-3.5 text-right text-sm font-semibold" :class="row.debit > 0 ? 'text-blue-700' : 'text-slate-300'">
                    {{ row.debit > 0 ? formatCurrency(row.debit) : '—' }}
                  </td>
                  <td class="px-6 py-3.5 text-right text-sm font-semibold" :class="row.credit > 0 ? 'text-rose-600' : 'text-slate-300'">
                    {{ row.credit > 0 ? formatCurrency(row.credit) : '—' }}
                  </td>
                </tr>
              </tbody>
              <tfoot>
                <tr class="bg-slate-50 border-t-2 border-slate-200">
                  <td colspan="2" class="px-6 py-4 text-sm font-extrabold text-slate-700">Total</td>
                  <td class="px-6 py-4 text-right text-sm font-extrabold text-blue-700">{{ formatCurrency(trialBalanceTotals.debit) }}</td>
                  <td class="px-6 py-4 text-right text-sm font-extrabold text-rose-600">{{ formatCurrency(trialBalanceTotals.credit) }}</td>
                </tr>
                <tr v-if="trialBalanceTotals.debit !== trialBalanceTotals.credit">
                  <td colspan="4" class="px-6 py-3 text-center text-xs font-bold text-amber-600 bg-amber-50">
                    ⚠ Neraca saldo tidak seimbang — selisih
                    {{ formatCurrency(Math.abs(trialBalanceTotals.debit - trialBalanceTotals.credit)) }}
                  </td>
                </tr>
              </tfoot>
            </table>
          </div>
        </div>
      </div>
    </div>

    <!-- ══════════════════════════════════════════════
         MODAL: TAMBAH AKUN
    ══════════════════════════════════════════════ -->
    <Teleport to="body">
      <Transition name="modal-fade">
        <div v-if="showAddAccountModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="showAddAccountModal = false" />
          <div class="relative bg-white rounded-3xl shadow-2xl w-full max-w-lg overflow-hidden">
            <div class="flex items-center justify-between px-7 py-5 border-b border-slate-100">
              <div class="flex items-center gap-3">
                <div class="w-9 h-9 bg-primary/10 rounded-2xl flex items-center justify-center">
                  <Layers class="w-5 h-5 text-primary" />
                </div>
                <h2 class="text-lg font-extrabold text-slate-900">Tambah Akun Baru</h2>
              </div>
              <button @click="showAddAccountModal = false" class="p-2 hover:bg-slate-100 rounded-xl transition-colors">
                <X class="w-5 h-5 text-slate-400" />
              </button>
            </div>
            <div class="px-7 py-6 space-y-5">
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-xs font-extrabold text-slate-500 uppercase tracking-wider mb-2">Kode Akun *</label>
                  <input
                    v-model="addAccountForm.code"
                    type="text"
                    placeholder="Contoh: 1-1100"
                    class="w-full px-4 py-2.5 border border-slate-200 rounded-xl text-sm font-medium text-slate-700 focus:outline-none focus:ring-2 focus:ring-primary/20"
                  />
                </div>
                <div>
                  <label class="block text-xs font-extrabold text-slate-500 uppercase tracking-wider mb-2">Tipe Akun *</label>
                  <select
                    v-model="addAccountForm.type"
                    class="w-full px-4 py-2.5 border border-slate-200 rounded-xl text-sm font-medium text-slate-700 focus:outline-none focus:ring-2 focus:ring-primary/20"
                  >
                    <option value="ASSET">Aset</option>
                    <option value="LIABILITY">Liabilitas</option>
                    <option value="EQUITY">Ekuitas</option>
                    <option value="REVENUE">Pendapatan</option>
                    <option value="EXPENSE">Beban</option>
                  </select>
                </div>
              </div>
              <div>
                <label class="block text-xs font-extrabold text-slate-500 uppercase tracking-wider mb-2">Nama Akun *</label>
                <input
                  v-model="addAccountForm.name"
                  type="text"
                  placeholder="Contoh: Kas dan Setara Kas"
                  class="w-full px-4 py-2.5 border border-slate-200 rounded-xl text-sm font-medium text-slate-700 focus:outline-none focus:ring-2 focus:ring-primary/20"
                />
              </div>
              <div>
                <label class="block text-xs font-extrabold text-slate-500 uppercase tracking-wider mb-2">Deskripsi</label>
                <textarea
                  v-model="addAccountForm.description"
                  rows="2"
                  placeholder="Deskripsi singkat akun (opsional)"
                  class="w-full px-4 py-2.5 border border-slate-200 rounded-xl text-sm font-medium text-slate-700 focus:outline-none focus:ring-2 focus:ring-primary/20 resize-none"
                />
              </div>
              <div>
                <label class="block text-xs font-extrabold text-slate-500 uppercase tracking-wider mb-2">Akun Induk</label>
                <select
                  v-model="addAccountForm.parent_id"
                  class="w-full px-4 py-2.5 border border-slate-200 rounded-xl text-sm font-medium text-slate-700 focus:outline-none focus:ring-2 focus:ring-primary/20"
                >
                  <option value="">— Tidak ada (akun utama) —</option>
                  <option v-for="a in accounts" :key="a.id" :value="a.id">{{ a.code }} — {{ a.name }}</option>
                </select>
              </div>
            </div>
            <div class="px-7 py-5 border-t border-slate-100 flex items-center justify-end gap-3">
              <button
                @click="showAddAccountModal = false"
                class="px-5 py-2.5 text-sm font-bold text-slate-600 bg-slate-100 hover:bg-slate-200 rounded-xl transition-colors"
              >
                Batal
              </button>
              <button
                @click="submitAddAccount"
                :disabled="addAccountLoading"
                class="flex items-center gap-2 px-6 py-2.5 bg-primary text-white text-sm font-bold rounded-xl hover:bg-primary/90 transition-all disabled:opacity-60 disabled:cursor-not-allowed"
              >
                <span v-if="addAccountLoading" class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                <Plus v-else class="w-4 h-4" />
                Simpan Akun
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- ══════════════════════════════════════════════
         MODAL: BUAT JURNAL MANUAL
    ══════════════════════════════════════════════ -->
    <Teleport to="body">
      <Transition name="modal-fade">
        <div v-if="showJournalModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="showJournalModal = false" />
          <div class="relative bg-white rounded-3xl shadow-2xl w-full max-w-4xl max-h-[90vh] flex flex-col overflow-hidden">
            <div class="flex items-center justify-between px-7 py-5 border-b border-slate-100 shrink-0">
              <div class="flex items-center gap-3">
                <div class="w-9 h-9 bg-primary/10 rounded-2xl flex items-center justify-center">
                  <FileText class="w-5 h-5 text-primary" />
                </div>
                <div>
                  <h2 class="text-lg font-extrabold text-slate-900">Buat Jurnal Manual</h2>
                  <p class="text-xs font-semibold text-slate-400 mt-0.5">Pastikan total debit = total kredit</p>
                </div>
              </div>
              <button @click="showJournalModal = false" class="p-2 hover:bg-slate-100 rounded-xl transition-colors">
                <X class="w-5 h-5 text-slate-400" />
              </button>
            </div>
            <div class="px-7 py-6 overflow-y-auto flex-1 space-y-5">
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-xs font-extrabold text-slate-500 uppercase tracking-wider mb-2">Tanggal *</label>
                  <input
                    v-model="journalForm.date"
                    type="date"
                    class="w-full px-4 py-2.5 border border-slate-200 rounded-xl text-sm font-medium text-slate-700 focus:outline-none focus:ring-2 focus:ring-primary/20"
                  />
                </div>
                <div>
                  <label class="block text-xs font-extrabold text-slate-500 uppercase tracking-wider mb-2">Deskripsi *</label>
                  <input
                    v-model="journalForm.description"
                    type="text"
                    placeholder="Contoh: Pembelian perlengkapan kantor"
                    class="w-full px-4 py-2.5 border border-slate-200 rounded-xl text-sm font-medium text-slate-700 focus:outline-none focus:ring-2 focus:ring-primary/20"
                  />
                </div>
              </div>

              <!-- Lines Table -->
              <div class="border border-slate-200 rounded-2xl overflow-hidden">
                <div class="bg-slate-50 px-4 py-3 flex items-center justify-between border-b border-slate-200">
                  <h3 class="text-xs font-extrabold text-slate-500 uppercase tracking-wider">Baris Jurnal</h3>
                  <button
                    @click="addJournalLine"
                    class="flex items-center gap-1.5 text-xs font-bold text-primary hover:text-primary/80 transition-colors"
                  >
                    <Plus class="w-3.5 h-3.5" /> Tambah Baris
                  </button>
                </div>
                <table class="w-full">
                  <thead>
                    <tr class="border-b border-slate-100">
                      <th class="text-left px-4 py-3 text-xs font-bold text-slate-400 w-[35%]">Akun</th>
                      <th class="text-left px-4 py-3 text-xs font-bold text-slate-400 w-[25%]">Keterangan</th>
                      <th class="text-right px-4 py-3 text-xs font-bold text-slate-400 w-[17%]">Debit (Rp)</th>
                      <th class="text-right px-4 py-3 text-xs font-bold text-slate-400 w-[17%]">Kredit (Rp)</th>
                      <th class="px-4 py-3 w-[6%]" />
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="(line, idx) in journalForm.lines" :key="idx" class="border-b border-slate-50 last:border-0">
                      <td class="px-3 py-2">
                        <select
                          v-model="line.account_id"
                          class="w-full px-3 py-2 text-xs border border-slate-200 rounded-lg font-medium text-slate-700 focus:outline-none focus:ring-2 focus:ring-primary/20"
                        >
                          <option value="">— Pilih akun —</option>
                          <option v-for="a in accounts" :key="a.id" :value="a.id">{{ a.code }} — {{ a.name }}</option>
                        </select>
                      </td>
                      <td class="px-3 py-2">
                        <input
                          v-model="line.description"
                          type="text"
                          placeholder="Keterangan..."
                          class="w-full px-3 py-2 text-xs border border-slate-200 rounded-lg font-medium text-slate-700 focus:outline-none focus:ring-2 focus:ring-primary/20"
                        />
                      </td>
                      <td class="px-3 py-2">
                        <input
                          v-model="line.debit"
                          type="number"
                          min="0"
                          placeholder="0"
                          class="w-full px-3 py-2 text-xs border border-slate-200 rounded-lg font-mono text-right text-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-200"
                          @input="() => { if (Number(line.debit) > 0) line.credit = '' }"
                        />
                      </td>
                      <td class="px-3 py-2">
                        <input
                          v-model="line.credit"
                          type="number"
                          min="0"
                          placeholder="0"
                          class="w-full px-3 py-2 text-xs border border-slate-200 rounded-lg font-mono text-right text-rose-600 focus:outline-none focus:ring-2 focus:ring-rose-200"
                          @input="() => { if (Number(line.credit) > 0) line.debit = '' }"
                        />
                      </td>
                      <td class="px-3 py-2 text-center">
                        <button
                          @click="removeJournalLine(idx)"
                          :disabled="journalForm.lines.length <= 2"
                          class="p-1.5 text-slate-300 hover:text-rose-500 hover:bg-rose-50 rounded-lg transition-colors disabled:opacity-30 disabled:cursor-not-allowed"
                        >
                          <Trash2 class="w-3.5 h-3.5" />
                        </button>
                      </td>
                    </tr>
                  </tbody>
                  <tfoot>
                    <tr class="bg-slate-50 border-t-2 border-slate-200">
                      <td colspan="2" class="px-4 py-3 text-xs font-extrabold text-slate-600">Total</td>
                      <td class="px-4 py-3 text-right text-sm font-extrabold text-blue-700">{{ formatCurrency(journalTotalDebit) }}</td>
                      <td class="px-4 py-3 text-right text-sm font-extrabold text-rose-600">{{ formatCurrency(journalTotalCredit) }}</td>
                      <td />
                    </tr>
                    <tr v-if="journalTotalDebit > 0 || journalTotalCredit > 0">
                      <td colspan="5" class="px-4 py-2.5">
                        <div
                          class="flex items-center gap-2 text-xs font-bold px-3 py-2 rounded-xl"
                          :class="journalBalanced ? 'bg-emerald-50 text-emerald-700' : 'bg-amber-50 text-amber-700'"
                        >
                          <component :is="journalBalanced ? CheckCircle : AlertCircle" class="w-4 h-4" />
                          <span v-if="journalBalanced">Seimbang ✓ — Jurnal siap diposting</span>
                          <span v-else>
                            Belum seimbang — Selisih {{ formatCurrency(Math.abs(journalTotalDebit - journalTotalCredit)) }}
                          </span>
                        </div>
                      </td>
                    </tr>
                  </tfoot>
                </table>
              </div>
            </div>
            <div class="px-7 py-5 border-t border-slate-100 flex items-center justify-end gap-3 shrink-0">
              <button
                @click="showJournalModal = false"
                class="px-5 py-2.5 text-sm font-bold text-slate-600 bg-slate-100 hover:bg-slate-200 rounded-xl transition-colors"
              >
                Batal
              </button>
              <button
                @click="submitJournal"
                :disabled="journalSubmitLoading || !journalBalanced"
                class="flex items-center gap-2 px-6 py-2.5 bg-primary text-white text-sm font-bold rounded-xl hover:bg-primary/90 transition-all disabled:opacity-60 disabled:cursor-not-allowed"
              >
                <span v-if="journalSubmitLoading" class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                <CheckCircle v-else class="w-4 h-4" />
                Posting Jurnal
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 0.2s ease;
}
.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}
</style>
