import { defineConfig } from 'vite'
import { resolve } from 'path'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  root: 'web',    
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
    historyApiFallback: true,
    proxy: {
        '/api': {
            target: 'http://localhost:8080',
            changeOrigin: true,
        },
    }
}
})