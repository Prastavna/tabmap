import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import ui from '@nuxt/ui/vite'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue(), tailwindcss(), ui()],
  resolve: {
    alias: {
      '@': '/src'
    }
  }
})
