import { defineConfig } from 'vite'
import react, { reactCompilerPreset } from '@vitejs/plugin-react'
import babel from '@rolldown/plugin-babel'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    babel({ presets: [reactCompilerPreset()] }),
    tailwindcss(),
  ],
  server: {
    proxy: {
      '/api': 'http://localhost:5050',
      '/ws': { target: 'ws://localhost:5050', ws: true },
      '/auth/google': 'http://localhost:5050',
      '/auth/github': 'http://localhost:5050',
      '/auth/telegram': 'http://localhost:5050',
      '/auth/me': 'http://localhost:5050',
    },
  },
})
