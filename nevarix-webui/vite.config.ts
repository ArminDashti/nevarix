import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { VitePWA } from 'vite-plugin-pwa'
import path from 'node:path'

const basePath = process.env.VITE_BASE_PATH || '/'
const baseNoSlash = basePath.replace(/\/$/, '') || ''
const apiProxyTarget = process.env.VITE_API_PROXY || 'http://127.0.0.1:8090'
const hmrClientPort = Number(process.env.VITE_HMR_CLIENT_PORT || 5173)

export default defineConfig({
  base: basePath,
  plugins: [
    vue(),
    tailwindcss(),
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: ['favicon.svg', 'icons/icon-192.png', 'icons/icon-512.png'],
      manifest: {
        name: 'Nevarix',
        short_name: 'Nevarix',
        description: 'Host and SoftEther monitoring',
        theme_color: '#0b1220',
        background_color: '#0b1220',
        display: 'standalone',
        start_url: basePath,
        icons: [
          { src: '/icons/icon-192.png', sizes: '192x192', type: 'image/png', purpose: 'any maskable' },
          { src: '/icons/icon-512.png', sizes: '512x512', type: 'image/png', purpose: 'any maskable' },
        ],
      },
      workbox: {
        navigateFallback: '/index.html',
        runtimeCaching: [
          {
            urlPattern: ({ url }) => url.pathname.includes('/api/'),
            handler: 'NetworkFirst',
            options: {
              cacheName: 'nevarix-api',
              networkTimeoutSeconds: 8,
              expiration: { maxEntries: 64, maxAgeSeconds: 60 },
            },
          },
        ],
      },
      devOptions: { enabled: false },
    }),
  ],
  resolve: {
    alias: { '@': path.resolve(__dirname, './src') },
  },
  server: {
    host: true,
    port: 5173,
    allowedHosts: true,
    hmr: { clientPort: hmrClientPort },
    proxy: {
      [`${baseNoSlash}/api`]: {
        target: apiProxyTarget,
        changeOrigin: true,
        rewrite: (p) => p.replace(new RegExp(`^${baseNoSlash}/api`), '/api'),
      },
    },
  },
})
