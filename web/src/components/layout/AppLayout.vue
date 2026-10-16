<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import Sidebar from './Sidebar.vue'
import Topbar from './Topbar.vue'
import Button from '@/components/ui/Button.vue'
import { api } from '@/api/client'

const route = useRoute()
const sidebarOpen = ref(false)
const passwordModalOpen = ref(false)
const refreshing = ref(false)

const oldPassword = ref('')
const newPassword = ref('')
const passwordError = ref('')
const passwordSuccess = ref(false)
const changingPassword = ref(false)

const currentTitle = computed(() => {
  const titles: Record<string, string> = {
    dashboard: '控制面板',
    nodes: '节点池管理',
    proxies: '入口代理池',
    subscriptions: '订阅与自定义 UA',
    keys: 'API 密钥凭证',
    settings: '系统配置',
    logs: '运行日志',
  }
  return titles[route.name as string] || '控制台'
})

async function handleChangePassword() {
  if (!oldPassword.value || !newPassword.value) {
    passwordError.value = '请填写完整原密码和新密码'
    return
  }
  changingPassword.value = true
  passwordError.value = ''
  try {
    const res = await api.changePassword(oldPassword.value, newPassword.value)
    if (res.ok) {
      passwordSuccess.value = true
      setTimeout(() => {
        passwordModalOpen.value = false
        oldPassword.value = ''
        newPassword.value = ''
        passwordSuccess.value = false
      }, 1500)
    } else {
      passwordError.value = res.error?.message || '密码修改失败'
    }
  } catch {
    passwordError.value = '请求失败，请稍后重试'
  } finally {
    changingPassword.value = false
  }
}
</script>

<template>
  <div class="min-h-screen bg-gray-50 flex">
    <!-- Sidebar -->
    <Sidebar
      :open="sidebarOpen"
      @update:open="sidebarOpen = $event"
    />

    <!-- Main Content Area -->
    <div class="flex-1 lg:pl-64 flex flex-col min-w-0">
      <Topbar
        :title="currentTitle"
        :refreshing="refreshing"
        @toggle-sidebar="sidebarOpen = !sidebarOpen"
        @open-password-modal="passwordModalOpen = true"
        @refresh="refreshing = false"
      />

      <main class="flex-1 p-4 sm:p-6 lg:p-8 max-w-7xl w-full mx-auto">
        <slot />
      </main>
    </div>

    <!-- Password Modal -->
    <div v-if="passwordModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-gray-900/50 backdrop-blur-xs p-4">
      <div class="bg-white rounded-2xl p-6 w-full max-w-md shadow-2xl border border-gray-100">
        <h3 class="text-base font-bold text-gray-900 mb-4">修改管理员密码</h3>
        <div class="space-y-3">
          <div>
            <label class="block text-xs font-semibold text-gray-600 mb-1">当前密码</label>
            <input
              v-model="oldPassword"
              type="password"
              placeholder="请输入当前密码"
              class="w-full px-3.5 py-2.5 bg-gray-50 border border-gray-200 rounded-xl text-sm outline-none focus:bg-white focus:border-emerald-600"
            >
          </div>
          <div>
            <label class="block text-xs font-semibold text-gray-600 mb-1">新密码</label>
            <input
              v-model="newPassword"
              type="password"
              placeholder="请输入新密码"
              class="w-full px-3.5 py-2.5 bg-gray-50 border border-gray-200 rounded-xl text-sm outline-none focus:bg-white focus:border-emerald-600"
            >
          </div>
          <div v-if="passwordError" class="text-xs text-rose-600 mt-1">
            {{ passwordError }}
          </div>
          <div v-if="passwordSuccess" class="text-xs text-emerald-600 mt-1">
            密码修改成功！
          </div>
        </div>
        <div class="mt-6 flex justify-end gap-2.5">
          <Button variant="ghost" @click="passwordModalOpen = false">取消</Button>
          <Button variant="primary" :loading="changingPassword" @click="handleChangePassword">确认修改</Button>
        </div>
      </div>
    </div>
  </div>
</template>
