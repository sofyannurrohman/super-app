<template>
  <div class="bank-account-view">
    <!-- Header -->
    <div class="view-header">
      <div class="header-content">
        <div class="header-icon">
          <Landmark :size="24" />
        </div>
        <div>
          <h1 class="view-title">Kelola Rekening Bank</h1>
          <p class="view-subtitle">Kelola data rekening bank perusahaan untuk transaksi transfer ke pelanggan di aplikasi mobile</p>
        </div>
      </div>
      <div class="header-actions">
        <select v-model="selectedFilterCompanyId" class="company-filter" @change="fetchAccounts">
          <option value="ALL">Semua Perusahaan (PT)</option>
          <option v-for="c in masterDataStore.companies" :key="c.id" :value="c.id">{{ c.name }}</option>
        </select>
        <select v-model="selectedFilterBranchId" class="company-filter" @change="fetchAccounts">
          <option value="ALL_BRANCHES">Semua Cabang</option>
          <option v-for="b in masterDataStore.branches" :key="b.id" :value="b.id">{{ b.name }}</option>
        </select>
        <button class="btn-primary" @click="openAddModal">
          <Plus :size="18" />
          Tambah Rekening
        </button>
      </div>
    </div>

    <!-- Info Banner -->
    <div class="info-banner">
      <Info :size="16" />
      <span>Rekening dapat diatur berdasarkan <strong>PT dan Cabang</strong>. Aplikasi mobile akan menyesuaikan rekening yang aktif sesuai dengan profil salesman / driver terkait.</span>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="loading-state">
      <div class="spinner" />
      <p>Memuat data rekening...</p>
    </div>

    <!-- Empty State -->
    <div v-else-if="accounts.length === 0" class="empty-state">
      <div class="empty-icon"><Landmark :size="48" /></div>
      <h3>Belum Ada Rekening Bank</h3>
      <p>Tambahkan rekening bank perusahaan untuk ditampilkan ke salesman &amp; driver</p>
      <button class="btn-primary" @click="openAddModal">
        <Plus :size="18" /> Tambah Rekening Pertama
      </button>
    </div>

    <!-- Account Cards Grid -->
    <div v-else class="accounts-grid">
      <div
        v-for="account in accounts"
        :key="account.id"
        class="account-card"
        :class="{ 'account-card--active': account.is_active }"
      >
        <!-- Active Badge -->
        <div v-if="account.is_active" class="active-badge">
          <CheckCircle2 :size="12" />
          Rekening Aktif
        </div>

        <!-- Card Header -->
        <div class="card-bank-header">
          <div class="bank-logo-wrap">
            <img
              v-if="account.bank_logo_url"
              :src="account.bank_logo_url"
              :alt="account.bank_name"
              class="bank-logo"
              @error="(e) => (e.target as HTMLImageElement).style.display='none'"
            />
            <div v-else class="bank-logo-placeholder">
              <Landmark :size="20" />
            </div>
          </div>
          <div class="bank-info">
            <h3 class="bank-name">{{ account.bank_name }}</h3>
            <p class="account-holder">{{ account.account_holder }}</p>
            <div class="account-company-wrapper">
              <div v-if="account.company_id" class="account-company">
                <Building2 :size="12" />
                <span>{{ getCompanyName(account.company_id) }}</span>
              </div>
              <div v-else class="account-company account-company--global">
                <Building2 :size="12" />
                <span>Global (Semua PT)</span>
              </div>

              <div v-if="account.branch_id" class="account-company" style="margin-top: 4px;">
                <MapPin :size="12" />
                <span>{{ getBranchName(account.branch_id) }}</span>
              </div>
              <div v-else class="account-company account-company--global" style="margin-top: 4px;">
                <MapPin :size="12" />
                <span>Semua Cabang</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Account Number -->
        <div class="account-number-block">
          <span class="account-number-label">NOMOR REKENING</span>
          <div class="account-number-row">
            <span class="account-number">{{ formatAccountNumber(account.account_number) }}</span>
            <button class="btn-copy" @click="copyAccountNumber(account.account_number)" title="Salin nomor rekening">
              <Copy :size="14" />
            </button>
          </div>
        </div>

        <!-- Notes -->
        <p v-if="account.notes" class="account-notes">{{ account.notes }}</p>

        <!-- Actions -->
        <div class="card-actions">
          <button
            v-if="!account.is_active"
            class="btn-set-active"
            @click="setActive(account)"
            :disabled="settingActive === account.id"
          >
            <CheckCircle2 :size="14" />
            {{ settingActive === account.id ? 'Memproses...' : 'Aktifkan' }}
          </button>
          <button
            v-else
            class="btn-set-active btn-set-active--on"
            @click="setActive(account)"
            :disabled="settingActive === account.id"
          >
            <Wifi :size="14" /> 
            {{ settingActive === account.id ? 'Memproses...' : 'Sedang Aktif' }}
          </button>
          <div class="action-buttons">
            <button class="btn-icon btn-edit" @click="openEditModal(account)" title="Edit">
              <Pencil :size="14" />
            </button>
            <button class="btn-icon btn-delete" @click="confirmDelete(account)" title="Hapus">
              <Trash2 :size="14" />
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Add/Edit Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="modal-overlay" @click.self="closeModal">
        <div class="modal-box">
          <div class="modal-header">
            <h2>{{ editingAccount ? 'Edit Rekening Bank' : 'Tambah Rekening Bank' }}</h2>
            <button class="btn-close" @click="closeModal"><X :size="20" /></button>
          </div>
          <form @submit.prevent="submitForm" class="modal-form">
            <div class="form-group">
              <label>Perusahaan (PT) <span class="optional">(opsional)</span></label>
              <select v-model="form.company_id">
                <option value="">Berlaku Global (Semua PT)</option>
                <option v-for="c in masterDataStore.companies" :key="c.id" :value="c.id">{{ c.name }}</option>
              </select>
              <span class="form-hint">Pilih PT untuk menyesuaikan rekening bank di aplikasi mobile sesuai data PT karyawan</span>
            </div>

            <div class="form-group">
              <label>Cabang (Branch) <span class="optional">(opsional)</span></label>
              <select v-model="form.branch_id">
                <option value="">Semua Cabang / Berlaku Global</option>
                <option v-for="b in masterDataStore.branches" :key="b.id" :value="b.id">{{ b.name }}</option>
              </select>
              <span class="form-hint">Pilih Cabang untuk membuat rekening spesifik per cabang</span>
            </div>
            <div class="form-group">
              <label>Nama Bank <span class="required">*</span></label>
              <input v-model="form.bank_name" type="text" placeholder="contoh: Bank Central Asia (BCA)" required />
            </div>
            <div class="form-group">
              <label>Nomor Rekening <span class="required">*</span></label>
              <input v-model="form.account_number" type="text" placeholder="contoh: 8175293021" required />
            </div>
            <div class="form-group">
              <label>Nama Pemilik Rekening <span class="required">*</span></label>
              <input v-model="form.account_holder" type="text" placeholder="contoh: Perusahaan Demo" required />
            </div>
            <div class="form-group">
              <label>URL Logo Bank <span class="optional">(opsional)</span></label>
              <input v-model="form.bank_logo_url" type="url" placeholder="https://..." />
              <span class="form-hint">URL gambar logo bank untuk ditampilkan di aplikasi</span>
            </div>
            <div class="form-group">
              <label>Catatan <span class="optional">(opsional)</span></label>
              <textarea v-model="form.notes" rows="2" placeholder="Catatan tambahan..." />
            </div>
            <div class="form-group form-group--checkbox">
              <label class="checkbox-label">
                <input v-model="form.is_active" type="checkbox" />
                <span>Jadikan rekening ini sebagai rekening aktif</span>
              </label>
              <span class="form-hint">Hanya satu rekening yang bisa aktif sekaligus. Rekening aktif sebelumnya akan dinonaktifkan.</span>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn-secondary" @click="closeModal">Batal</button>
              <button type="submit" class="btn-primary" :disabled="submitting">
                <Loader2 v-if="submitting" :size="16" class="spin" />
                {{ submitting ? 'Menyimpan...' : (editingAccount ? 'Perbarui' : 'Simpan') }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>

    <!-- Delete Confirm Modal -->
    <Teleport to="body">
      <div v-if="showDeleteConfirm" class="modal-overlay" @click.self="showDeleteConfirm = false">
        <div class="modal-box modal-box--sm">
          <div class="delete-confirm">
            <div class="delete-icon"><Trash2 :size="32" /></div>
            <h3>Hapus Rekening Bank?</h3>
            <p>Rekening <strong>{{ deletingAccount?.bank_name }}</strong> dengan nomor <strong>{{ deletingAccount?.account_number }}</strong> akan dihapus permanen jika belum digunakan dalam transaksi. Jika sudah digunakan, nonaktifkan rekening agar riwayat transaksi tetap tersimpan.</p>
            <div class="modal-footer">
              <button class="btn-secondary" @click="showDeleteConfirm = false">Batal</button>
              <button class="btn-danger" @click="deleteAccount" :disabled="deleting">
                {{ deleting ? 'Menghapus...' : 'Ya, Hapus' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { toast } from 'vue-sonner'
import {
  Landmark, Plus, CheckCircle2, Pencil, Trash2, Copy,
  Info, X, Loader2, Wifi, Building2, MapPin
} from 'lucide-vue-next'
import apiClient from '@/api/axios'
import { useMasterDataStore } from '@/stores/masterData'

interface BankAccount {
  id: string
  company_id: string | null
  branch_id: string | null
  bank_name: string
  account_number: string
  account_holder: string
  bank_logo_url: string
  is_active: boolean
  notes: string
  created_at: string
  updated_at: string
  company?: any
  branch?: any
}

const accounts = ref<BankAccount[]>([])
const loading = ref(false)
const submitting = ref(false)
const deleting = ref(false)
const settingActive = ref<string | null>(null)

const masterDataStore = useMasterDataStore()
const selectedFilterCompanyId = ref<string>('ALL')
const selectedFilterBranchId = ref<string>('ALL_BRANCHES')

const showModal = ref(false)
const showDeleteConfirm = ref(false)
const editingAccount = ref<BankAccount | null>(null)
const deletingAccount = ref<BankAccount | null>(null)

const form = ref({
  company_id: '',
  branch_id: '',
  bank_name: '',
  account_number: '',
  account_holder: '',
  bank_logo_url: '',
  notes: '',
  is_active: false,
})

function formatAccountNumber(num: string): string {
  const clean = num.replace(/[^\d]/g, '')
  if (clean.length === 10) {
    return `${clean.slice(0,3)}-${clean.slice(3,6)}-${clean.slice(6)}`
  }
  return num
}

async function fetchAccounts() {
  loading.value = true
  try {
    const params = new URLSearchParams()
    if (selectedFilterCompanyId.value !== 'ALL') {
      params.set('company_id', selectedFilterCompanyId.value)
    }
    if (selectedFilterBranchId.value !== 'ALL_BRANCHES') {
      params.set('branch_id', selectedFilterBranchId.value)
    }
    const query = params.toString()
    const res = await apiClient.get(`/finance/bank-accounts${query ? `?${query}` : ''}`)
    accounts.value = res.data.data || []
  } catch (e: any) {
    toast.error('Gagal memuat data rekening bank')
  } finally {
    loading.value = false
  }
}

function openAddModal() {
  editingAccount.value = null
  form.value = {
    company_id: '',
    branch_id: '',
    bank_name: '',
    account_number: '',
    account_holder: '',
    bank_logo_url: '',
    notes: '',
    is_active: false,
  }
  showModal.value = true
}

function openEditModal(account: BankAccount) {
  editingAccount.value = account
  form.value = {
    company_id: account.company_id || '',
    branch_id: account.branch_id || '',
    bank_name: account.bank_name,
    account_number: account.account_number,
    account_holder: account.account_holder,
    bank_logo_url: account.bank_logo_url || '',
    notes: account.notes || '',
    is_active: account.is_active,
  }
  showModal.value = true
}

function closeModal() {
  showModal.value = false
  editingAccount.value = null
}

async function submitForm() {
  submitting.value = true
  try {
    const payload = { 
      ...form.value, 
      company_id: form.value.company_id || null,
      branch_id: form.value.branch_id || null
    }
    if (editingAccount.value) {
      await apiClient.put(`/finance/bank-accounts/${editingAccount.value.id}`, payload)
      toast.success('Rekening bank berhasil diperbarui')
    } else {
      await apiClient.post('/finance/bank-accounts', payload)
      toast.success('Rekening bank berhasil ditambahkan')
    }
    closeModal()
    await fetchAccounts()
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Gagal menyimpan rekening bank')
  } finally {
    submitting.value = false
  }
}

function confirmDelete(account: BankAccount) {
  deletingAccount.value = account
  showDeleteConfirm.value = true
}

async function deleteAccount() {
  if (!deletingAccount.value) return
  deleting.value = true
  try {
    await apiClient.delete(`/finance/bank-accounts/${deletingAccount.value.id}`)
    toast.success('Rekening bank berhasil dihapus')
    showDeleteConfirm.value = false
    deletingAccount.value = null
    await fetchAccounts()
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Gagal menghapus rekening bank')
  } finally {
    deleting.value = false
  }
}

async function setActive(account: BankAccount) {
  settingActive.value = account.id
  try {
    await apiClient.post(`/finance/bank-accounts/${account.id}/set-active`)
    toast.success(`Rekening ${account.bank_name} berhasil diaktifkan`)
    await fetchAccounts()
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Gagal mengaktifkan rekening')
  } finally {
    settingActive.value = null
  }
}

async function copyAccountNumber(num: string) {
  const clean = num.replace(/[^\d]/g, '')
  await navigator.clipboard.writeText(clean)
  toast.success('Nomor rekening disalin!')
}

function getCompanyName(id: string): string {
  const c = masterDataStore.companies.find(c => c.id === id)
  return c ? c.name : 'Unknown PT'
}

function getBranchName(id: string): string {
  const b = masterDataStore.branches.find(b => b.id === id)
  return b ? b.name : 'Unknown Branch'
}

onMounted(async () => {
  await masterDataStore.fetchCompanies()
  await masterDataStore.fetchBranches()
  fetchAccounts()
})
</script>

<style scoped>
.bank-account-view {
  padding: 24px;
  max-width: 1100px;
  margin: 0 auto;
}

.view-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
  gap: 16px;
  flex-wrap: wrap;
}
.header-content {
  display: flex;
  align-items: center;
  gap: 14px;
}
.header-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: linear-gradient(135deg, #6366f1, #4f46e5);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  flex-shrink: 0;
}
.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.company-filter {
  padding: 10px 14px;
  border-radius: 10px;
  border: 1.5px solid #e2e8f0;
  background-color: white;
  color: #1e293b;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  outline: none;
  transition: all 0.2s;
}
.company-filter:focus, .company-filter:hover {
  border-color: #6366f1;
}
.view-title {
  font-size: 20px;
  font-weight: 700;
  color: #1e293b;
  margin: 0;
}
.view-subtitle {
  font-size: 13px;
  color: #64748b;
  margin: 2px 0 0;
}

.info-banner {
  display: flex;
  align-items: center;
  gap: 10px;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  border-radius: 10px;
  padding: 12px 16px;
  font-size: 13px;
  color: #1d4ed8;
  margin-bottom: 24px;
}
.info-banner strong { font-weight: 700; }

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 60px;
  color: #64748b;
}
.spinner {
  width: 36px;
  height: 36px;
  border: 3px solid #e2e8f0;
  border-top-color: #6366f1;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

.empty-state {
  text-align: center;
  padding: 60px 24px;
  color: #64748b;
}
.empty-icon {
  width: 80px;
  height: 80px;
  border-radius: 20px;
  background: #f1f5f9;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 16px;
  color: #94a3b8;
}
.empty-state h3 { font-size: 18px; font-weight: 700; color: #1e293b; margin: 0 0 8px; }
.empty-state p { margin: 0 0 20px; }

.accounts-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
}

.account-card {
  background: white;
  border: 1.5px solid #e2e8f0;
  border-radius: 16px;
  padding: 20px;
  position: relative;
  transition: box-shadow 0.2s, border-color 0.2s;
}
.account-card:hover {
  box-shadow: 0 4px 20px rgba(0,0,0,0.08);
}
.account-card--active {
  border-color: #22c55e;
  background: linear-gradient(135deg, #f0fdf4 0%, #ffffff 100%);
  box-shadow: 0 4px 20px rgba(34,197,94,0.12);
}

.active-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  background: #22c55e;
  color: white;
  font-size: 11px;
  font-weight: 700;
  padding: 3px 10px;
  border-radius: 20px;
  margin-bottom: 14px;
  letter-spacing: 0.3px;
}

.card-bank-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}
.bank-logo-wrap {
  width: 52px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.bank-logo {
  max-width: 52px;
  max-height: 32px;
  object-fit: contain;
}
.bank-logo-placeholder {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: #f1f5f9;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #94a3b8;
}
.bank-name { font-size: 15px; font-weight: 700; color: #1e293b; margin: 0 0 2px; }
.account-holder { font-size: 12px; color: #64748b; margin: 0 0 4px; }
.account-company {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  font-weight: 600;
  color: #6366f1;
  background: #eff6ff;
  padding: 3px 8px;
  border-radius: 6px;
}
.account-company--global {
  color: #d97706;
  background: #fef3c7;
}

.account-number-block {
  background: #f8fafc;
  border-radius: 10px;
  padding: 12px 14px;
  margin-bottom: 12px;
}
.account-number-label {
  font-size: 9px;
  font-weight: 800;
  color: #94a3b8;
  letter-spacing: 0.8px;
  display: block;
  margin-bottom: 4px;
}
.account-number-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.account-number {
  font-size: 20px;
  font-weight: 900;
  color: #1e293b;
  letter-spacing: 1px;
}
.btn-copy {
  width: 30px;
  height: 30px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: white;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #6366f1;
  transition: all 0.15s;
}
.btn-copy:hover { background: #ede9fe; border-color: #6366f1; }

.account-notes {
  font-size: 12px;
  color: #64748b;
  margin: 0 0 12px;
  padding: 8px 10px;
  background: #f8fafc;
  border-radius: 8px;
  border-left: 3px solid #e2e8f0;
}

.card-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: 14px;
  border-top: 1px solid #f1f5f9;
  margin-top: 4px;
}
.btn-set-active {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 600;
  color: var(--primary-color, #0f172a);
  background: #f1f5f9;
  border: none;
  cursor: pointer;
  transition: all 0.2s ease;
}
.btn-set-active:hover:not(:disabled) {
  background: #e2e8f0;
}
.btn-set-active--on {
  color: #10b981;
  background: #ecfdf5;
}
.btn-set-active--on:hover:not(:disabled) {
  background: #d1fae5;
  color: #059669;
}
.btn-set-active:disabled { opacity: 0.6; cursor: not-allowed; }
.active-label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 600;
  color: #16a34a;
}
.action-buttons { display: flex; gap: 8px; }
.btn-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  border: 1.5px solid #e2e8f0;
  background: white;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.15s;
}
.btn-edit { color: #6366f1; }
.btn-edit:hover { background: #ede9fe; border-color: #6366f1; }
.btn-delete { color: #ef4444; }
.btn-delete:hover { background: #fef2f2; border-color: #ef4444; }

.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: linear-gradient(135deg, #6366f1, #4f46e5);
  color: white;
  border: none;
  border-radius: 10px;
  padding: 10px 18px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.15s, transform 0.15s;
}
.btn-primary:hover { opacity: 0.92; transform: translateY(-1px); }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; transform: none; }
.btn-secondary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: white;
  color: #64748b;
  border: 1.5px solid #e2e8f0;
  border-radius: 10px;
  padding: 10px 18px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}
.btn-secondary:hover { background: #f8fafc; }
.btn-danger {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: #ef4444;
  color: white;
  border: none;
  border-radius: 10px;
  padding: 10px 18px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}
.btn-danger:hover { background: #dc2626; }
.btn-danger:disabled { opacity: 0.6; cursor: not-allowed; }

.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
  backdrop-filter: blur(4px);
}
.modal-box {
  background: white;
  border-radius: 20px;
  width: 100%;
  max-width: 520px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 20px 60px rgba(0,0,0,0.2);
}
.modal-box--sm { max-width: 420px; }
.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px 0;
}
.modal-header h2 { font-size: 17px; font-weight: 700; color: #1e293b; margin: 0; }
.btn-close {
  width: 32px;
  height: 32px;
  border: none;
  background: #f1f5f9;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #64748b;
}
.btn-close:hover { background: #e2e8f0; }

.modal-form { padding: 20px 24px 24px; }
.form-group {
  margin-bottom: 16px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.form-group label {
  font-size: 13px;
  font-weight: 600;
  color: #374151;
}
.required { color: #ef4444; }
.optional { color: #94a3b8; font-weight: 400; font-size: 12px; }
.form-group select {
  width: 100%;
  padding: 12px 14px;
  border: 1.5px solid #e2e8f0;
  border-radius: 10px;
  font-size: 14px;
  color: #1e293b;
  transition: all 0.2s;
  background: #f8fafc;
  outline: none;
  font-family: inherit;
}
.form-group select:focus {
  border-color: #6366f1;
  background: white;
}
.form-group input {
  width: 100%;
  padding: 12px 14px;
  border: 1.5px solid #e2e8f0;
  border-radius: 10px;
  font-size: 14px;
  color: #1e293b;
  transition: all 0.2s;
  background: #f8fafc;
  outline: none;
  font-family: inherit;
}
.form-group input:focus,
.form-group textarea:focus { border-color: #6366f1; }
.form-hint { font-size: 11px; color: #94a3b8; }
.form-group--checkbox { flex-direction: column; }
.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-weight: 400 !important;
}
.checkbox-label input[type=checkbox] { width: 16px; height: 16px; accent-color: #6366f1; }

.modal-footer {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid #f1f5f9;
}

.delete-confirm {
  padding: 32px 24px;
  text-align: center;
}
.delete-icon {
  width: 64px;
  height: 64px;
  border-radius: 16px;
  background: #fef2f2;
  color: #ef4444;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 16px;
}
.delete-confirm h3 { font-size: 18px; font-weight: 700; color: #1e293b; margin: 0 0 8px; }
.delete-confirm p { font-size: 13px; color: #64748b; margin: 0 0 4px; line-height: 1.5; }
.delete-confirm .modal-footer { justify-content: center; }

.spin { animation: spin 0.8s linear infinite; }
</style>
