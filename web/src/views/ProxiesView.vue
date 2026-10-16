<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api/client'
import type { ProxyCandidate } from '@/api/client'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'
import { Plus, Play, Trash2 } from 'lucide-vue-next'

const proxies = ref<ProxyCandidate[]>([])
const newProxyUri = ref('')
const loading = ref(false)

async function loadProxies() {
  loading.value = true
  try {
    const res = await api.getProxies()
    proxies.value = res.candidates || []
  } finally {
    loading.value = false
  }
}

async function handleImport() {
  const uri = newProxyUri.value.trim()
  if (!uri) return
  await api.importProxy(uri)
  newProxyUri.value = ''
  await loadProxies()
}

async function handleTestAll() {
  await api.testBatchProxies()
  await loadProxies()
}

async function handleDelete(uri: string) {
  if (confirm('确认删除该入口代理吗？')) {
    await api.deleteProxy(uri)
    await loadProxies()
  }
}

async function handleDeleteDisabled() {
  if (confirm('确定清理所有禁用的入口代理吗？')) {
    await api.deleteDisabledProxies()
    await loadProxies()
  }
}

onMounted(() => {
  void loadProxies()
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h2 class="text-xl font-bold text-gray-900 tracking-tight">入口代理池</h2>
        <p class="text-xs text-gray-500 mt-1">全局多入口候选轮询管理</p>
      </div>

      <div class="flex items-center gap-2">
        <Button variant="primary" size="md" @click="handleTestAll">
          <Play class="w-4 h-4" />
          <span>批量测试入口</span>
        </Button>
        <Button variant="outline" size="md" @click="handleDeleteDisabled">
          <Trash2 class="w-4 h-4" />
          <span>清理禁用</span>
        </Button>
      </div>
    </div>

    <!-- Import Input -->
    <div class="bg-white border border-gray-200/80 rounded-2xl p-4 shadow-xs flex gap-3">
      <input
        v-model="newProxyUri"
        type="text"
        placeholder="添加全局入口代理 (socks5://127.0.0.1:1080 或 vless://...)"
        class="flex-1 px-3.5 py-2 bg-gray-50 border border-gray-200 rounded-xl text-xs outline-none focus:bg-white focus:border-emerald-600 font-mono"
        @keydown.enter.prevent="handleImport"
      >
      <Button variant="secondary" size="sm" @click="handleImport">
        <Plus class="w-3.5 h-3.5" />
        <span>添加候选</span>
      </Button>
    </div>

    <!-- Proxies Table -->
    <div class="bg-white border border-gray-200/80 rounded-2xl shadow-xs overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs">
          <thead class="bg-gray-50 border-b border-gray-100 text-gray-500 font-medium">
            <tr>
              <th class="px-4 py-3">入口名称</th>
              <th class="px-4 py-3">类型</th>
              <th class="px-4 py-3">状态</th>
              <th class="px-4 py-3">测试延迟</th>
              <th class="px-4 py-3 text-right">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100">
            <tr v-if="proxies.length === 0">
              <td colspan="5" class="px-4 py-8 text-center text-gray-400">
                暂无入口代理配置
              </td>
            </tr>
            <tr v-for="p in proxies" :key="p.raw_uri" class="hover:bg-gray-50/50">
              <td class="px-4 py-3 font-medium text-gray-800">{{ p.name }}</td>
              <td class="px-4 py-3">
                <Badge variant="blue" size="sm">{{ p.type.toUpperCase() }}</Badge>
              </td>
              <td class="px-4 py-3">
                <span v-if="p.disabled" class="text-gray-400 font-medium">已禁用</span>
                <span v-else class="text-emerald-600 font-medium">启用中</span>
              </td>
              <td class="px-4 py-3 font-mono text-gray-500">
                {{ p.last_test_ms > 0 ? `${Math.round(p.last_test_ms)}ms` : '-' }}
              </td>
              <td class="px-4 py-3 text-right">
                <button
                  type="button"
                  class="text-rose-600 font-medium hover:underline"
                  @click="handleDelete(p.raw_uri)"
                >
                  删除
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
