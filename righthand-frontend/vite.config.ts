import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'
import type { UserConfigExport } from 'vite'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src')
    }
  },
  server: {
    host: '0.0.0.0',
    port: 5173,
    proxy: {
      // 所有 API 请求统一转发到 Go 后端服务
      '/api': {
        target: 'http://localhost:9080',
        changeOrigin: true
        // 不需要 rewrite,保持 /api 前缀
      }
    }
  },
  test: {
    globals: true,
    environment: 'happy-dom',
    setupFiles: ['./tests/setup.ts'],
    server: {
      deps: {
        // element-plus 与 async-validator 必须内联交由 Vite 处理：
        // async-validator 无 exports 字段，Node ESM 解析其 CJS 入口(main)得到 { default: fn } 包装对象，
        // 导致 ElForm 内部 new AsyncValidator() 抛 TypeError、表单校验静默失效（校验恒通过）。
        // 内联后 Vite 走 module 字段(dist-web ESM)，default 导出为构造函数。
        inline: [/element-plus/, /async-validator/]
      }
    },
    // 真实组件（Element Plus）挂载较慢，放宽超时避免环境差异导致误报
    testTimeout: 30000,
    hookTimeout: 30000
  }
})
