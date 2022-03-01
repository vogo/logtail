import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'
import { viteSingleFile } from "vite-plugin-singlefile"
import { resolve } from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  build: {
    target: 'esnext',
    assetsInlineLimit: 999999999,
    chunkSizeWarningLimit: 999999999,
    cssCodeSplit: false,
    rollupOptions: {
      output: {
        inlineDynamicImports: true,
        manualChunks: undefined,
      }
    }
  },
  server: {
    proxy: {
      '/manage': 'http://localhost:54321',
    }
  },
  plugins: [
    vue(),
    AutoImport({
      resolvers: [ElementPlusResolver()],
    }),
    Components({
      resolvers: [ElementPlusResolver()],
    }),
    viteSingleFile()
  ]
})
