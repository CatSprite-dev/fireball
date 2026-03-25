import { defineConfig } from 'vite'
import { resolve } from 'path'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  root: 'web',
  publicDir: 'assets',     
  build: {
    outDir: '../dist',
    emptyOutDir: true,
    rollupOptions: {
      input: {
        main: resolve(__dirname, 'web/index.html'),
      },
    },
    minify: 'esbuild',
    sourcemap: true,
  },
  plugins: [
    vue(),
  ],
  server: {
    port: 3000,
    open: true,
    proxy: {
        '/auth': {
            target: 'http://localhost:8080',
            changeOrigin: true,
            rewrite: (path) => path,
            configure: (proxy) => {
                proxy.on('proxyReq', (proxyReq, req, res) => {
                    console.log(`---> Proxy: ${req.method} ${req.url} -> 8080${req.url}`);
                })
            }
        }
    }
  }
})