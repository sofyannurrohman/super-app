## Goal Description
Membuat role `supervisor_gudang` yang memiliki hak akses untuk:
- Mengakses tab **Pabrik** dan **Gudang**.
- Melihat data gudang semua cabang maupun tiap-tiap cabang secara spesifik (Branch Selector di-unlock).

## Proposed Changes

### Backend
1. **`internal/domain/user.go`**: Menambahkan `RoleSupervisorGudang Role = "supervisor_gudang"`.
2. **`internal/delivery/http/middleware.go`**: Memodifikasi `RoleMiddleware` agar role `supervisor_gudang` secara otomatis mewarisi akses dari `admin_gudang` dan `admin_pabrik` (sehingga dapat mengakses seluruh API terkait Gudang dan Pabrik).
3. **`internal/delivery/http/factory_handler.go`**: Mengizinkan `supervisor_gudang` melihat HPP/Harga Pabrik di API `GetProducts` (sama seperti `admin_gudang`).
4. **`internal/delivery/http/vehicle_handler.go`**: Memodifikasi pengecekan `isWarehouseAdmin` agar `supervisor_gudang` dapat mem-bypass filter cabang kendaraan.
5. **`internal/delivery/http/employee_handler.go`**: Menambahkan `RoleSupervisorGudang` ke dalam middleware `readOnlyAdmin` agar bisa melihat daftar karyawan/driver.

### Frontend
1. **`src/stores/auth.ts`**: Menambahkan `'supervisor_gudang'` ke array `isAuthorized` agar user bisa login.
2. **`src/components/Header.vue`**:
   - Menambahkan kondisi di `filteredModules` agar role ini melihat modul `warehouse` dan `factory`.
   - Mengubah kondisi `isBranchSelectorLocked` dengan menambahkan pengecualian untuk `supervisor_gudang` agar bisa bebas berpindah cabang/melihat semua cabang.
   - Menyetel modul default ke `warehouse` pada saat `onMounted`.
3. **`src/router/index.ts`**:
   - Menambahkan guard untuk membatasi navigasi `supervisor_gudang` hanya ke path yang berawalan `/warehouse`, `/factory`, `/vehicles`, atau `/kehadiran`.
   - Memastikan default redirect saat login mengarah ke `/warehouse`.
4. **`src/components/Sidebar.vue`**: Mengatur `userRoleName` agar merender label "Supervisor Gudang".
5. **`src/views/UserView.vue`**:
   - Menambahkan pilihan "Supervisor Gudang" pada dropdown role saat membuat/mengedit pengguna.
   - Memberikan badge warna khusus (misal: teal/emerald) di data table untuk memudahkan identifikasi.
   - Memasukkan role ini ke dalam kondisi `v-if` untuk menampilkan checkbox **Global Access (view_all_branches)**.

## Verification Plan
### Manual Verification
1. Buat pengguna baru dengan role `Supervisor Gudang` dan centang opsi **Global Access**.
2. Login menggunakan akun tersebut.
3. Pastikan halaman pertama adalah Dashboard Gudang.
4. Verifikasi hanya dua tab modul yang muncul: **Gudang** dan **Pabrik**.
5. Coba pilih cabang tertentu dan juga "Semua Cabang" di header, pastikan UI tidak di-lock.
6. Coba navigasi ke Dashboard Pabrik dan pastikan data bisa termuat tanpa error 403.
