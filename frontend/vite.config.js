import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    host: '0.0.0.0',  // Allows Vite to be accessible over the network
    port: 5173,       // Default port; adjust if necessary
  }
})
