# HRIS client trial

Paket ini adalah salinan kode terpisah dari commit yang tercatat di `SOURCE_COMMIT`. Database, Redis, dan uploads hanya untuk trial. Tidak ada dump database produksi atau kredensial integrasi produksi di paket ini.

## Menjalankan

Butuh Docker dan Docker Compose v2. Setelah clone, salin `.env.example` ke `.env`. Ubah `POSTGRES_PASSWORD`, `JWT_SECRET`, `TRIAL_COMPANY_NAME`, `TRIAL_ADMIN_EMAIL`, `TRIAL_ADMIN_PASSWORD`, `TRIAL_STAFF_PASSWORD`, dan `TRIAL_HTTP_PORT` sebelum diberikan ke klien. Buat password dan JWT secret acak yang kuat. Simpan `.env` secara privat; file ini diabaikan Git.

```sh
docker compose -f compose.yaml build
docker compose -f compose.yaml up -d postgres redis
docker compose -f compose.yaml run --rm --no-deps --entrypoint ./seeder backend
docker compose -f compose.yaml up -d backend frontend
docker compose -f compose.yaml ps
```

Akses lokal: `http://127.0.0.1:13001` (sesuaikan port `.env`). Login memakai email dan password `TRIAL_ADMIN_*` di `.env`. Jalankan seeder sekali saja untuk database baru. Seeder standar mengisi perusahaan dan cabang fiktif, 7 karyawan HR, jadwal dan riwayat absensi, saldo serta permohonan cuti, dan lembur. Lima akun operasional, toko, KPI sales, dan transaksi contoh hanya dibuat jika `TRIAL_SEED_OPERATIONS=true`. Akun `hr@demo.local`, `manager@demo.local`, dan `staff@demo.local` memakai `TRIAL_STAFF_PASSWORD`. Seeder dapat dijalankan ulang tanpa menggandakan transaksi contoh.

Untuk akses klien dari internet, pasang di host trial tersendiri dengan HTTPS, reverse proxy ke port lokal `TRIAL_HTTP_PORT`, dan pembatasan akses. Jangan mengekspos port PostgreSQL atau Redis. Gunakan nama domain trial milik klien dan pisahkan backup dari produksi.

Jika dipasang pada host yang juga menjalankan aplikasi lain, gunakan folder `/apps/hris-client-trial`, `COMPOSE_PROJECT_NAME=hris-client-trial`, dan `TRIAL_HTTP_PORT=13001`. Skrip `deploy-isolated.sh` menolak menjalankan deploy jika project Compose atau port itu sudah dipakai. Stack ini memiliki jaringan dan volume sendiri; port PostgreSQL dan Redis tidak dipublikasikan. Skrip hanya membuat stack baru dan tidak mengubah konfigurasi Nginx host. Jalankan dari folder tersebut dengan `sh deploy-isolated.sh`, lalu pasang domain dan HTTPS khusus trial setelah stack sehat.

Untuk menghentikan: `docker compose -f compose.yaml down`. Jangan menambahkan `-v` kecuali memang ingin menghapus data trial. Integrasi email, pembayaran, marketplace, FCM, dan AI belum dikonfigurasi; fitur yang memerlukannya perlu kredensial sandbox milik klien. Aplikasi mobile perlu build terpisah dengan URL API trial.

## Identitas produk

Nama demo saat ini **Karsa HR**. Ganti `VITE_APP_NAME`, `VITE_APP_TAGLINE`, dan `VITE_APP_LOGO_URL` di `.env`, lalu bangun ulang frontend untuk memakai identitas lain. Untuk logo lokal, simpan aset di `frontend-web/public` dan isi URL seperti `/logo-klien.svg`. Isi kanal dukungan hanya dengan kontak yang memang akan melayani calon klien. Nama perusahaan demo di database diatur oleh `TRIAL_COMPANY_NAME`. Nama produk ini adalah identitas sementara; lakukan pemeriksaan merek sebelum penggunaan komersial.

Sebelum trial dibuka ke publik, isi kontak bantuan dan kebijakan penghapusan akun dengan kanal yang benar. Modul operasional khusus pabrik, marketplace, dan dokumen legal masih perlu peninjauan per calon klien.

Aplikasi mobile memakai ID paket sementara `com.example.karsahrtrial`. Sebelum distribusi resmi, ganti dengan domain aplikasi milik Anda, konfigurasi Firebase baru, dan URL API trial lewat `--dart-define=BASE_URL=... --dart-define=COMPANY_NAME=...`.

## Produk modular

Instalasi demo standar mengaktifkan semua modul pada tab header. Atur `VITE_ENABLED_MODULES` di `.env` dengan daftar dipisahkan koma untuk membatasi modul, lalu bangun ulang frontend dan buat ulang backend. Modul PKWT dimatikan secara default karena naskah kontrak harus ditinjau dan disesuaikan per perusahaan; `VITE_ENABLE_PKWT=true` menampilkannya kembali setelah template legal disetujui. Pengaturan ini mengatur tampilan dan navigasi web serta menghentikan beberapa pekerjaan latar vertikal yang tidak aktif. Otorisasi API tetap mengikuti hak akses backend; audit dan pemisahan endpoint per modul masih diperlukan sebelum produk dijual sebagai SaaS bersama.

`TRIAL_SEED_OPERATIONS=false` hanya membuat data HR. Setel `true` untuk menambah contoh sales dan toko pada database trial baru. Identitas perusahaan pada dokumen web diambil dari perusahaan akun yang masuk. Untuk menawarkan model SaaS multi-perusahaan dalam satu database, perlu audit isolasi data di semua endpoint sebelum penjualan; paket ini dirancang sebagai instalasi terpisah per klien.

## Satu instalasi per klien

Salin paket ini ke folder klien yang berbeda. Beri nilai unik pada `COMPOSE_PROJECT_NAME`, `TRIAL_HTTP_PORT`, `TRIAL_COMPANY_NAME`, akun admin, dan seluruh rahasia di `.env` tiap folder. Compose akan memberi nama volume PostgreSQL dan Redis sesuai project sehingga data antar-klien terpisah. Simpan uploads dan backup per folder. Jangan menggunakan satu `.env` untuk dua klien.

Utilitas migrasi dan perbaikan yang terikat data perusahaan asal tidak disertakan dalam paket trial. Integrasi Firebase dan Google Drive hanya aktif jika kredensial serta folder milik klien disetel melalui `FCM_CREDENTIALS_FILE`, `GDRIVE_CREDENTIALS_FILE`, dan `GDRIVE_FOLDER_ID`.
