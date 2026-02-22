import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import { visualizer } from 'rollup-plugin-visualizer'
import Components from 'unplugin-vue-components/vite'
import { AntDesignVueResolver } from 'unplugin-vue-components/resolvers'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    // 自动按需导入组件
    Components({
      resolvers: [
        AntDesignVueResolver({
          importStyle: false // 不自动导入样式，我们在 main.ts 中统一导入
        })
      ]
    }),
    // 构建分析插件（仅在构建时启用）
    visualizer({
      open: false, // 构建后不自动打开
      filename: 'dist/stats.html', // 分析文件输出路径
      gzipSize: true,
      brotliSize: true
    })
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  build: {
    // 启用 CSS 代码分割
    cssCodeSplit: true,
    // 设置 chunk 大小警告限制为 500kb
    chunkSizeWarningLimit: 500,
    rollupOptions: {
      output: {
        // 手动分块策略 - 使用函数进行更细粒度的分割
        manualChunks: (id) => {
          // Vue 核心库
          if (id.includes('node_modules/vue/') || 
              id.includes('node_modules/vue-router/') || 
              id.includes('node_modules/pinia/')) {
            return 'vue-vendor'
          }
          
          // Ant Design Vue 图标单独分块
          if (id.includes('@ant-design/icons-vue')) {
            return 'antd-icons'
          }
          
          // Ant Design Vue 核心组件 - 按使用频率和大小分组
          if (id.includes('ant-design-vue')) {
            // 表格组件（最大，单独分块）
            if (id.includes('/table/') || id.includes('/vc-table/')) {
              return 'antd-table'
            }
            // 表单相关组件（第二大）
            if (id.includes('/form/') || 
                id.includes('/input/') || 
                id.includes('/select/') ||
                id.includes('/vc-select/') ||
                id.includes('/vc-input/')) {
              return 'antd-form'
            }
            // 导航组件（包含 menu 和 tabs）
            if (id.includes('/menu/') || 
                id.includes('/tabs/') || 
                id.includes('/dropdown/') ||
                id.includes('/vc-menu/') ||
                id.includes('/vc-tabs/') ||
                id.includes('/vc-dropdown/')) {
              return 'antd-navigation'
            }
            // 反馈组件
            if (id.includes('/modal/') || 
                id.includes('/message/') || 
                id.includes('/notification/') ||
                id.includes('/spin/') || 
                id.includes('/alert/') || 
                id.includes('/result/') ||
                id.includes('/popconfirm/') ||
                id.includes('/vc-dialog/')) {
              return 'antd-feedback'
            }
            // 数据展示组件
            if (id.includes('/card/') || 
                id.includes('/descriptions/') || 
                id.includes('/tag/') || 
                id.includes('/badge/') || 
                id.includes('/avatar/') ||
                id.includes('/list/') ||
                id.includes('/statistic/') ||
                id.includes('/tooltip/') ||
                id.includes('/popover/')) {
              return 'antd-display'
            }
            // 布局组件（较小，可以合并）
            if (id.includes('/layout/') || 
                id.includes('/grid/') || 
                id.includes('/space/') ||
                id.includes('/divider/') ||
                id.includes('/row/') ||
                id.includes('/col/')) {
              return 'antd-layout'
            }
            // 其他基础组件和工具（包括 _util, config-provider 等）
            return 'antd-core'
          }
          
          // 工具库
          if (id.includes('axios') || id.includes('@emotion/css')) {
            return 'utils-vendor'
          }
          
          // node_modules 中的其他依赖
          if (id.includes('node_modules')) {
            return 'vendor'
          }
        },
        // 用于从入口点创建的块的打包输出格式
        entryFileNames: 'assets/[name]-[hash].js',
        // 用于命名代码拆分时创建的共享块的输出命名
        chunkFileNames: 'assets/[name]-[hash].js',
        // 用于输出静态资源的命名
        assetFileNames: 'assets/[name]-[hash].[ext]'
      }
    },
    // 压缩选项
    minify: 'terser',
    terserOptions: {
      compress: {
        // 生产环境移除 console
        drop_console: true,
        drop_debugger: true
      }
    }
  },
  server: {
    proxy: {
      '/oauth': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/.well-known': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
})
