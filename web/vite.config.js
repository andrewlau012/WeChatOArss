import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src')
    }
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/feed': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/login': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/add': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/del': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/pause': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/list': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/version': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/opml': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/img-proxy': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/video-proxy': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/link-proxy': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
})
