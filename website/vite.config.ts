import { tanstackStart } from '@tanstack/react-start/plugin/vite'
import react from '@vitejs/plugin-react'
import { nitro } from 'nitro/vite'
import { defineConfig } from 'vite'

export default defineConfig({
  server: { port: 3010 },
  plugins: [
    tanstackStart({ prerender: { enabled: true, crawlLinks: true } }),
    nitro(),
    react(),
  ],
})
