import { fileURLToPath, URL } from 'node:url'
import dns from 'node:dns'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import tailwindcss from '@tailwindcss/vite'
import { compression } from 'vite-plugin-compression2'

dns.setDefaultResultOrder('ipv4first')

export default defineConfig({
  plugins: [
    vue(),
    ...(process.env.NODE_ENV !== 'production' ? [vueDevTools()] : []),
    tailwindcss(),
    // Pre-build .gz files — Nginx serve langsung tanpa kompresi real-time
    // Hanya aktif saat production build
    ...(process.env.NODE_ENV === 'production'
      ? [
          compression({
            exclude: [/\.(png|jpe?g|gif|ico|webp|woff2?)$/], // skip binary
            threshold: 1024, // hanya kompres file > 1KB
          }),
        ]
      : []),
  ],
  esbuild: {
    drop: process.env.NODE_ENV === 'production' ? ['console', 'debugger'] : [],
  },
  build: {
    // Ingatkan jika ada chunk > 1000KB setelah splitting
    chunkSizeWarningLimit: 1000,
    rollupOptions: {
      output: {
        manualChunks: (id) => {
          // Library chart besar — hanya dimuat di halaman yang butuh chart
          if (id.includes('node_modules/echarts') || id.includes('node_modules/vue-echarts')) {
            return 'vendor-charts'
          }
          // Library Excel — hanya dimuat saat export/import
          if (id.includes('node_modules/xlsx') || id.includes('node_modules/exceljs')) {
            return 'vendor-excel'
          }
          // Library peta — hanya dimuat di halaman dengan map
          if (id.includes('node_modules/leaflet')) {
            return 'vendor-map'
          }
          // Vue ecosystem — selalu dipakai, pisahkan dari app code
          if (id.includes('node_modules/vue') || id.includes('node_modules/pinia') || id.includes('node_modules/vue-router')) {
            return 'vendor-vue'
          }
          // UI components
          if (id.includes('node_modules/reka-ui') || id.includes('node_modules/lucide-vue-next')) {
            return 'vendor-ui'
          }
          // Modul lainnya dibiarkan di-split secara alami oleh Rollup untuk mencegah circular dependency
        },
      },
    },
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },
  server: {
    host: true,
    proxy: {
      '/api': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
      '/uploads': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
    },
  }
})