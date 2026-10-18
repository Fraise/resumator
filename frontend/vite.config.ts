import {defineConfig} from 'vite'
import {svelte} from '@sveltejs/vite-plugin-svelte'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [svelte()],
  server: {
    // Wails' dev assetserver proxies webview requests with the
    // "wails.localhost" host header; Vite 8+ rejects unknown hosts.
    allowedHosts: ['wails.localhost', 'localhost', '127.0.0.1']
  }
})
