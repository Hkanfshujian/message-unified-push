// import { defineConfig } from 'vite'
// import vue from '@vitejs/plugin-vue'

// // https://vite.dev/config/
// export default defineConfig({
//   plugins: [vue()],
// })



import path from 'node:path'
import tailwindcss from '@tailwindcss/vite'
import vue from '@vitejs/plugin-vue'
import { defineConfig, loadEnv } from 'vite'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const proxyTarget = env.VITE_API_URL || 'http://127.0.0.1:8081'

  return {
    // 使用相对路径，这样可以在任何路径下部署
    base: './',
    plugins: [vue(), tailwindcss()],
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
      },
    },
    define: {
      __BUILD_TIME__: JSON.stringify(new Date().toISOString()),
      'globalThis.__BUILD_TIME__': JSON.stringify(new Date().toISOString()),
    },
    server: {
      proxy: {
        '/uploads': {
          target: proxyTarget,
          changeOrigin: true,
          secure: false,
        },
        '/public': {
          target: proxyTarget,
          changeOrigin: true,
          secure: false,
        },
        '/storage': {
          target: proxyTarget,
          changeOrigin: true,
          secure: false,
        },
      },
    },
  }
})
