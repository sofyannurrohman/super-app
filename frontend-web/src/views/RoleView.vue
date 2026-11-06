<script setup lang="ts">
import { ref } from 'vue'

const roles = ref([
  { id: 'superadmin', label: 'Super Admin', desc: 'Akses penuh ke semua modul sistem' },
  {
    id: 'hr_admin',
    label: 'HR Admin',
    desc: 'Hak wewenang kelola absensi, data pegawai, dan payroll',
  },
  { id: 'manager', label: 'Manager', desc: 'Akses Dashboard, KPI, dan evaluasi staff bawahan' },
  { id: 'supervisor_sales', label: 'Supervisor Sales', desc: 'Hanya boleh mengakses modul Sales & Marketing' },
  { id: 'cashier', label: 'Cashier / Kasir', desc: 'Hanya boleh mengakses modul Keuangan' },
  { id: 'admin_kasir', label: 'Admin Kasir', desc: 'Akses modul Keuangan dan Absensi Karyawan' },
  { id: 'admin_nota', label: 'Admin Nota', desc: 'Hanya boleh mengakses modul Sales & Marketing (konfirmasi nota)' },
  { id: 'admin_gudang', label: 'Admin Gudang', desc: 'Hanya boleh mengakses modul Gudang Cabang' },
  { id: 'admin_marketplace', label: 'Admin Marketplace', desc: 'Hanya boleh mengakses modul Marketplace Cabang' },
  { id: 'admin_designer', label: 'Admin Designer', desc: 'Hanya boleh mengakses modul Branding' },
  { id: 'validator_desain', label: 'Validator Desain', desc: 'Hanya boleh mengakses modul Branding (Validator Spanduk - ACC Desain & Revisi Desain)' },
  { id: 'validator_toko', label: 'Validator Toko', desc: 'Hanya boleh mengakses modul Branding (Validator Spanduk - ACC Toko & Revisi Toko)' },
  { id: 'admin_purchasing', label: 'Admin Purchasing', desc: 'Hak wewenang modul Branding (Vendor & Purchasing), Pabrik, dan Keuangan Pusat' },
  { id: 'admin_operasional', label: 'Admin Operasional', desc: 'Hak wewenang operasional gudang, sales, kasir, dan absensi cabang' },
  { id: 'admin_pabrik', label: 'Admin Pabrik', desc: 'Hak wewenang modul Pabrik, RajaTransport, serta modul Gudang Cabang sesuai unit cabang penugasan' },
  { id: 'supervisor_gudang', label: 'Supervisor Gudang', desc: 'Hak wewenang modul Gudang Cabang, Pabrik, dan RajaTransport' },
  { id: 'auditor', label: 'Auditor', desc: 'Hak wewenang read-only untuk penarikan data (Khusus cabang Booster: otomatis membuka akses Admin Gudang, Admin Nota, dan Kasir)' },
  { id: 'customer_support', label: 'Customer Support', desc: 'Hak wewenang read-only untuk memonitoring data ke semua cabang' },
  { id: 'employee', label: 'Karyawan', desc: 'Hak wewenang default / aplikasi Mobile' },
])
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold tracking-tight text-slate-900">Manajemen Hak Akses (Role)</h1>
        <p class="text-slate-500 mt-1">
          Daftar ID akses pengguna yang mendefinisikan batas kemampuan manipulasi data.
        </p>
      </div>
      <button
        class="px-4 py-2 bg-slate-900 text-white font-medium rounded-lg hover:bg-slate-800 transition-colors"
      >
        Tambah Role Ekstra
      </button>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div
        v-for="role in roles"
        :key="role.id"
        class="bg-white rounded-xl border border-slate-200 shadow-sm p-6 hover:shadow-md transition-shadow relative overflow-hidden group"
      >
        <!-- Decoration -->
        <div
          class="absolute top-0 right-0 w-24 h-24 bg-gradient-to-br from-indigo-50 to-indigo-100 rounded-bl-full -z-10 opacity-50 group-hover:opacity-100 transition-opacity"
        ></div>

        <h3 class="text-xl font-bold text-slate-900">{{ role.label }}</h3>
        <p
          class="text-xs font-mono text-indigo-500 mt-1 bg-indigo-50 inline-block px-2 py-0.5 rounded"
        >
          ID: {{ role.id }}
        </p>
        <p class="text-slate-600 mt-4 text-sm">{{ role.desc }}</p>

        <div class="mt-6 flex justify-end">
          <button
            class="text-sm font-medium text-slate-500 hover:text-indigo-600 transition-colors"
          >
            Lihat Daftar Pengguna &rarr;
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
