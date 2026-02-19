import { defineConfig } from 'vite'
import { resolve } from 'path'
import htmlMinifier from 'vite-plugin-html-minifier'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  root: 'web',
  publicDir: 'assets',     
  build: {
    outDir: '../dist',
    emptyOutDir: true,
    rollupOptions: {
      input: {
        main: resolve(__dirname, 'web/index.html'),
        login: resolve(__dirname, 'web/login.html'),
      },
      output: {
        entryFileNames: 'assets/[name].js',
        assetFileNames: 'assets/[name].[ext]',
        chunkFileNames: 'assets/[name]-[hash].js',
      }
    },
    minify: 'esbuild',
    sourcemap: true,
  },
  plugins: [
    tailwindcss(),
    htmlMinifier({
      minify: true,
    })
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