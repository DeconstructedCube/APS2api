<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { api } from '@/api/client'
import Button from '@/components/ui/Button.vue'
import { RefreshCw } from 'lucide-vue-next'

const logsText = ref('加载日志中...')
const autoRefresh = ref(true)
const logContainerRef = ref<HTMLPreElement | null>(null)
let refreshTimer: ReturnType<typeof setInterval> | null = null

async function loadLogs() {
  try {
    const res = await api.getLogs()
    logsText.value = res.content || '暂无日志记录'
    await nextTick()
    if (logContainerRef.value) {
      logContainerRef.value.scrollTop = logContainerRef.value.scrollHeight
    }
  } catch {
    logsText.value = '无法获取系统运行日志'
  }
}

function handleAutoRefreshChange() {
  if (autoRefresh.value) {
    startPolling()
  } else if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

function startPolling() {
  if (refreshTimer) clearInterval(refreshTimer)
  refreshTimer = setInterval(() => {
    if (autoRefresh.value) {
      void loadLogs()
    }
  }, 3000)
}

onMounted(() => {
  void loadLogs()
  startPolling()
})

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer)
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h2 class="text-xl font-bold text-gray-900 tracking-tight">系统运行日志</h2>
        <p class="text-xs text-gray-500 mt-1">展示系统最近 200 条日志流</p>
      </div>

      <div class="flex items-center gap-4">
        <label class="flex items-center gap-2 text-xs font-medium text-gray-600 cursor-pointer select-none">
          <input
            v-model="autoRefresh"
            type="checkbox"
            class="rounded border-gray-300 text-emerald-600"
            @change="handleAutoRefreshChange"
          >
          <span>自动轮询刷新 (3s)</span>
        </label>
        <Button variant="secondary" size="md" @click="loadLogs">
          <RefreshCw class="w-4 h-4" />
          <span>刷新</span>
        </Button>
      </div>
    </div>

    <!-- Logs Box -->
    <div class="bg-white border border-gray-200/80 rounded-2xl p-4 shadow-xs">
      <pre
        ref="logContainerRef"
        class="w-full h-[68vh] bg-gray-50 border border-gray-200/80 rounded-xl p-4 text-xs font-mono text-gray-800 overflow-y-auto leading-relaxed select-text whitespace-pre-wrap break-all"
      >{{ logsText }}</pre>
    </div>
  </div>
</template>
