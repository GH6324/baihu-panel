import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { viteStaticCopy } from 'vite-plugin-static-copy'
import { VitePWA } from 'vite-plugin-pwa'
import { fileURLToPath, URL } from 'node:url'
import { writeFileSync, readFileSync, readdirSync, statSync, existsSync, unlinkSync } from 'node:fs'
import { join, resolve } from 'node:path'
import { gzipSync } from 'node:zlib'

// Simple inline compression plugin to avoid extra dependencies
const compressionPlugin = () => ({
  name: 'compression-plugin',
  apply: 'build',
  closeBundle: {
    order: 'post',
    handler: async () => {
      const distDir = resolve(process.cwd(), 'dist')
      if (!existsSync(distDir)) return
      
      const compressDir = (dir: string) => {
        const files = readdirSync(dir)
        for (const file of files) {
          const filePath = join(dir, file)
          const stat = statSync(filePath)
          if (stat.isDirectory()) {
            compressDir(filePath)
          } else if (file.match(/\.(js|css|html|svg|json|vs)$/) && !file.endsWith('.gz')) {
            try {
              const content = readFileSync(filePath)
              if (content.length < 1024) continue // Don't compress very small files
              const gzipped = gzipSync(content, { level: 9 })
              writeFileSync(`${filePath}.gz`, gzipped)
              // Remove original file after successful compression to save space in binary
              unlinkSync(filePath)
            } catch (e) {
              console.error(`Compression failed for ${file}:`, e)
            }
          }
        }
      }
      compressDir(distDir)
    }
  }
} as const)

export default defineConfig({
  plugins: [
    vue(),
    tailwindcss(),
    viteStaticCopy({
      targets: [
        {
          src: 'node_modules/monaco-editor/min/vs',
          dest: 'assets'
        }
      ]
    }),
    compressionPlugin(),
    VitePWA({
      registerType: 'autoUpdate',
      injectRegister: 'auto',
      includeAssets: ['favicon.ico', 'logo.svg', 'pwa-icon-192.png', 'pwa-icon-512.png'],
      manifest: {
        name: 'Baihu Panel',
        short_name: 'Baihu',
        description: '白虎面板 - 现代化的服务器管理面板',
        theme_color: '#ffffff',
        background_color: '#ffffff',
        display: 'standalone',
        icons: [
          {
            src: 'pwa-icon-192.png',
            sizes: '192x192',
            type: 'image/png'
          },
          {
            src: 'pwa-icon-512.png',
            sizes: '512x512',
            type: 'image/png'
          },
          {
            src: 'pwa-icon-512.png',
            sizes: '512x512',
            type: 'image/png',
            purpose: 'any maskable'
          }
        ]
      },
      workbox: {
        maximumFileSizeToCacheInBytes: 10 * 1024 * 1024 // 10MB to allow monaco-editor files
      },
      devOptions: {
        enabled: true
      }
    })
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8052',
        changeOrigin: true,
        ws: true
      },
      '/openapi': {
        target: 'http://localhost:8052',
        changeOrigin: true
      }
    }
  },
  // 使用相对路径，这样动态导入的模块也会使用相对路径
  // 浏览器会根据当前页面 URL 解析相对路径
  base: './',
  build: {
    reportCompressedSize: true,
    sourcemap: false,
    rollupOptions: {
      output: {
        // 防止生成以 _ 开头的文件，导致被 Cloudflare Pages 或 Github Pages 等静态托管平台拦截并降级返回 HTML
        chunkFileNames: 'assets/[name]-[hash].js',
        entryFileNames: 'assets/[name]-[hash].js',
        assetFileNames: 'assets/[name]-[hash].[ext]',
        sanitizeFileName(name) {
          // 仿制 Rollup 的默认 sanitizeFileName，将特殊字符替换为 '-'
          let safeName = name.replace(/[\0?*:|"<>\/\\&=$]/g, '-')
          // 去除开头可能引起静态托管平台屏蔽的下划线 '_'
          return safeName.replace(/^_/, '')
        }
      }
    }
  },
  define: {
  }
})
