import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'

// Filter harmless browser notifications/warnings (e.g. during zoom or responsive resize)
const isBenignError = (msg: any) => {
  const str = String(msg || '')
  return (
    str.includes('ResizeObserver loop') ||
    str.includes('ResizeObserver loop completed with undelivered notifications') ||
    str.includes('ResizeObserver loop limit exceeded') ||
    str === 'Script error.'
  )
}

// Global catch-all: prevent blank white screen by showing an error dialog
window.onerror = (msg, src, line, col, err) => {
  if (isBenignError(msg) || isBenignError(err?.message)) {
    return true // Ignore benign notifications so browser does not interrupt the app
  }
  console.error('[Global Error Caught]:', msg, err)
  const div = document.getElementById('app')
  // Only display fatal dialog if the app failed during initial boot/mounting
  const isInitialBoot = !div || !div.children.length || !!document.getElementById('app-initial-loader')
  if (div && isInitialBoot) {
    div.innerHTML = `<div style="padding:32px 24px;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;color:#b91c1c;background:#fef2f2;border:1px solid #fca5a5;border-radius:16px;margin:40px auto;max-width:550px;box-shadow:0 20px 25px -5px rgba(0,0,0,0.1)">
      <div style="display:flex;align-items:center;gap:12px;margin-bottom:12px;">
        <span style="font-size:24px;">⚠️</span>
        <h2 style="margin:0;font-size:18px;font-weight:800;color:#991b1b">Gagal Memuat Halaman</h2>
      </div>
      <p style="margin:0 0 16px;color:#475569;font-size:13px;line-height:1.5">
        Terjadi kendala saat memuat berkas sistem di peramban Anda. Hal ini biasanya terjadi jika browser menyimpan cache versi lama setelah server diperbarui.
      </p>
      <div style="display:flex;gap:8px;margin-bottom:20px;">
        <button onclick="localStorage.clear();sessionStorage.clear();window.location.reload(true);" style="padding:10px 18px;background:#b91c1c;color:#fff;border:none;border-radius:10px;font-size:13px;font-weight:bold;cursor:pointer;box-shadow:0 4px 6px -1px rgba(185,28,28,0.2);">
          Bersihkan Cache &amp; Muat Ulang
        </button>
        <button onclick="window.location.href='/login'" style="padding:10px 18px;background:#f1f5f9;color:#334155;border:1px solid #cbd5e1;border-radius:10px;font-size:13px;font-weight:bold;cursor:pointer;">
          Halaman Login
        </button>
      </div>
      <details style="cursor:pointer;border-top:1px dashed #fca5a5;padding-top:12px">
        <summary style="color:#64748b;font-size:12px;font-weight:600">Detail Teknis (Error Trace)</summary>
        <pre style="font-size:11px;color:#334155;white-space:pre-wrap;margin-top:8px;background:#ffffff;padding:12px;border-radius:8px;border:1px solid #e2e8f0;overflow-x:auto">${String(err?.stack || err || msg)}</pre>
      </details>
    </div>`
  }
}

window.addEventListener('unhandledrejection', (e) => {
  if (isBenignError(e.reason) || isBenignError(e.reason?.message)) {
    return
  }
  console.error('[Unhandled Rejection]:', e.reason)
  const isChunkError = e.reason && (
    String(e.reason).includes('Failed to fetch dynamically imported module') ||
    String(e.reason).includes('Importing a module script failed')
  )
  if (isChunkError) {
    const reloadKey = 'last_chunk_reload_attempt'
    const lastAttempt = sessionStorage.getItem(reloadKey)
    const now = Date.now()
    if (!lastAttempt || now - parseInt(lastAttempt, 10) > 15000) {
      sessionStorage.setItem(reloadKey, now.toString())
      window.location.reload()
      return
    }
  }
  const div = document.getElementById('app')
  if (div && (!div.children.length || document.getElementById('app-initial-loader'))) {
    window.onerror && window.onerror(String(e.reason), '', 0, 0, e.reason)
  }
})

const app = createApp(App)

app.config.errorHandler = (err, _instance, info) => {
  // Use Function constructor to bypass esbuild console drop in production
  const log = new Function('e', 'i', 'console.error("[Vue Error]:", e, i)')
  log(err, info)
}

app.use(createPinia())
app.use(router)

app.mount('#app')

