<script setup lang="ts">
import { LayoutDashboard, Zap, Globe, Layers, Key, Sliders, Terminal, LogOut } from 'lucide-vue-next'
import { useRouter, useRoute } from 'vue-router'
import { api } from '@/api/client'

defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  'update:open': [val: boolean]
}>()

const router = useRouter()
const route = useRoute()

const navItems = [
  { name: '控制面板', path: '/', icon: LayoutDashboard },
  { name: '节点池管理', path: '/nodes', icon: Zap },
  { name: '入口代理池', path: '/proxies', icon: Globe },
  { name: '订阅与 UA', path: '/subscriptions', icon: Layers },
  { name: 'API 密钥', path: '/keys', icon: Key },
  { name: '系统设置', path: '/settings', icon: Sliders },
  { name: '系统日志', path: '/logs', icon: Terminal },
]

function handleNav(path: string) {
  router.push(path)
  emit('update:open', false)
}

async function handleLogout() {
  await api.logout()
  router.push('/login')
}
</script>

<template>
  <!-- Mobile Backdrop -->
  <div
    v-if="open"
    class="fixed inset-0 z-40 bg-gray-900/40 backdrop-blur-xs lg:hidden transition-opacity"
    @click="emit('update:open', false)"
  />

  <!-- Sidebar -->
  <aside
    :class="[
      'fixed top-0 bottom-0 left-0 z-40 w-64 bg-white border-r border-gray-200/80 flex flex-col transition-transform duration-200 ease-in-out lg:translate-x-0',
      open ? 'translate-x-0 shadow-2xl lg:shadow-none' : '-translate-x-full',
    ]"
  >
    <!-- Logo Header -->
    <div class="h-16 px-6 border-b border-gray-100 flex items-center gap-3">
      <div class="w-8 h-8 rounded-xl bg-emerald-600 text-white font-bold text-sm flex items-center justify-center shadow-xs">
        V
      </div>
      <div class="flex flex-col">
        <span class="font-bold text-sm text-gray-900 leading-none">APS2api</span>
        <span class="text-[11px] font-medium text-gray-400 mt-1">控制台</span>
      </div>
    </div>

    <!-- Nav Items -->
    <nav class="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
      <button
        v-for="item in navItems"
        :key="item.path"
        type="button"
        :class="[
          'w-full flex items-center gap-3 px-3.5 py-2.5 rounded-xl text-sm font-medium transition-colors text-left cursor-pointer',
          route.path === item.path
            ? 'bg-emerald-50 text-emerald-700 font-semibold'
            : 'text-gray-600 hover:bg-gray-100/80 hover:text-gray-900',
        ]"
        @click="handleNav(item.path)"
      >
        <component
          :is="item.icon"
          class="w-4 h-4 shrink-0"
          :class="route.path === item.path ? 'text-emerald-600' : 'text-gray-400'"
        />
        <span>{{ item.name }}</span>
      </button>
    </nav>

    <!-- Footer -->
    <div class="p-3 border-t border-gray-100 flex items-center justify-between px-3 py-2 bg-gray-50 rounded-xl m-3 text-xs text-gray-500">
      <div class="flex items-center gap-2">
        <span class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
        <span class="font-medium text-gray-700">服务正常</span>
      </div>

      <button
        type="button"
        class="text-gray-400 hover:text-rose-600 transition-colors p-1 rounded hover:bg-gray-200 cursor-pointer"
        title="退出登录"
        @click="handleLogout"
      >
        <LogOut class="w-4 h-4" />
      </button>
    </div>
  </aside>
</template>
