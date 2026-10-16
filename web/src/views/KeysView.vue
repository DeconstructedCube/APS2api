<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api/client'
import type { ApiKeyItem } from '@/api/client'
import Button from '@/components/ui/Button.vue'
import { Plus, Copy, Check, Trash2 } from 'lucide-vue-next'

const keys = ref<ApiKeyItem[]>([])
const newName = ref('')
const newKey = ref('')
const copiedKey = ref('')

async function loadKeys() {
  const res = await api.getKeys()
  keys.value = res.keys || []
}

async function handleAdd() {
  if (!newName.value.trim()) return
  await api.addKey(newName.value.trim(), newKey.value.trim())
  newName.value = ''
  newKey.value = ''
  await loadKeys()
}

async function handleDelete(name: string) {
  if (confirm(`确认删除 API 密钥 "${name}" 吗？`)) {
    await api.deleteKey(name)
    await loadKeys()
  }
}

async function copyKey(k: string) {
  await navigator.clipboard.writeText(k)
  copiedKey.value = k
  setTimeout(() => { copiedKey.value = '' }, 2000)
}

onMounted(() => {
  void loadKeys()
})
</script>

<template>
  <div class="space-y-6">
    <div>
      <h2 class="text-xl font-bold text-gray-900 tracking-tight">API 密钥管理</h2>
      <p class="text-xs text-gray-500 mt-1">对外服务调用的身份鉴权凭证 (sk-...)</p>
    </div>

    <!-- Create Card -->
    <div class="bg-white border border-gray-200/80 rounded-2xl p-5 shadow-xs space-y-3 max-w-2xl">
      <h3 class="text-xs font-semibold text-gray-700 uppercase tracking-wider">新建 API 密钥</h3>
      <div class="flex gap-3 flex-wrap">
        <input
          v-model="newName"
          type="text"
          placeholder="密钥名称（例如 client-app）"
          class="flex-1 min-w-44 px-3.5 py-2.5 bg-gray-50 border border-gray-200 rounded-xl text-xs outline-none focus:bg-white focus:border-emerald-600"
        >
        <input
          v-model="newKey"
          type="text"
          placeholder="留空自动生成 (sk-...)"
          class="flex-1 min-w-64 px-3.5 py-2.5 bg-gray-50 border border-gray-200 rounded-xl text-xs outline-none focus:bg-white focus:border-emerald-600 font-mono"
        >
        <Button variant="primary" size="md" @click="handleAdd">
          <Plus class="w-4 h-4" />
          <span>添加密钥</span>
        </Button>
      </div>
    </div>

    <!-- Keys Table -->
    <div class="bg-white border border-gray-200/80 rounded-2xl shadow-xs overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs">
          <thead class="bg-gray-50 border-b border-gray-100 text-gray-500 font-medium">
            <tr>
              <th class="px-4 py-3">名称</th>
              <th class="px-4 py-3">密钥内容</th>
              <th class="px-4 py-3 text-right">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100">
            <tr v-if="keys.length === 0">
              <td colspan="3" class="px-4 py-8 text-center text-gray-400">暂无 API 密钥</td>
            </tr>
            <tr v-for="k in keys" :key="k.name" class="hover:bg-gray-50/50">
              <td class="px-4 py-3 font-medium text-gray-800">{{ k.name }}</td>
              <td class="px-4 py-3 font-mono text-gray-600">{{ k.key }}</td>
              <td class="px-4 py-3 text-right space-x-2">
                <button
                  type="button"
                  class="inline-flex items-center gap-1 text-gray-500 hover:text-gray-800 font-medium"
                  @click="copyKey(k.key)"
                >
                  <component :is="copiedKey === k.key ? Check : Copy" class="w-3.5 h-3.5" :class="{ 'text-emerald-600': copiedKey === k.key }" />
                  <span>{{ copiedKey === k.key ? '已复制' : '复制' }}</span>
                </button>
                <button
                  type="button"
                  class="text-rose-600 font-medium hover:underline inline-flex items-center gap-1"
                  @click="handleDelete(k.name)"
                >
                  <Trash2 class="w-3.5 h-3.5" />
                  <span>删除</span>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
