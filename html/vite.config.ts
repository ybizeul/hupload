import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tsconfigPaths from 'vite-tsconfig-paths'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react(),tsconfigPaths()],
  server: {
    proxy: {
      '/api': 'http://127.0.0.1:8080/',
      '/d/': 'http://127.0.0.1:8080/',
      '/login': 'http://127.0.0.1:8080/',
      '/oidc': 'http://127.0.0.1:8080/',
      '/auth': 'http://127.0.0.1:8080/',
    }
  },
  resolve: { alias: { '@tabler/icons-react': '@tabler/icons-react/dist/esm/icons/index.mjs', }, },
})
