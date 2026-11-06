# Karsa HR — instalasi trial terpisah

Repositori ini berisi backend Go, frontend web Vue, aplikasi mobile Flutter, dan stack Docker Compose untuk satu instalasi per klien. Panduan lengkap ada di [README-TRIAL.md](README-TRIAL.md).

## Deploy pertama dari GitHub ke VPS

Clone branch `hris-trial` dari repositori privat ini ke `/apps/hris-client-trial`, lalu jalankan di VPS:

```sh
git clone --branch hris-trial https://github.com/sofyannurrohman/super-app.git /apps/hris-client-trial
```

```sh
cd /apps/hris-client-trial
cp .env.example .env
chmod 600 .env
# Isi password acak, JWT secret, email admin, dan nama perusahaan di .env.
sh deploy-isolated.sh
```

Skrip deploy hanya membuat project Compose `hris-client-trial` pada port `127.0.0.1:13001`; skrip berhenti jika nama project atau port itu sudah digunakan. PostgreSQL dan Redis milik trial memakai volume terpisah dan tidak membuka port host. Siapkan domain dan HTTPS trial secara terpisah jika akses publik diperlukan.

`.env`, data uploads, dan hasil build tidak disimpan di Git. Modul awal yang aktif hanya `hris`; PKWT, integrasi eksternal, dan distribusi mobile memerlukan konfigurasi serta pengujian tersendiri.
