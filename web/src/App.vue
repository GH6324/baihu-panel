<script setup lang="ts">
import { onMounted } from 'vue'
import { RouterView } from 'vue-router'
import { Toaster } from '@/components/ui/sonner'

onMounted(() => {
  // 全局应用设备差异化垂直抗锯齿
  const isMac = /Mac|iPod|iPhone|iPad/.test(navigator.userAgent)
  document.documentElement.classList.add(isMac ? 'antialiased' : 'subpixel-antialiased')

  // Windows 平台检测，用于在 CSS 中调整字体优先级
  const isWindows = /Win/.test(navigator.userAgent)
  if (isWindows) {
    document.documentElement.classList.add('is-windows')
  }
})
</script>

<template>
  <RouterView />
  <Toaster position="bottom-right" :duration="2000" />
</template>
