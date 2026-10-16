<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api/client'
import type { SystemSettings } from '@/api/client'
import Button from '@/components/ui/Button.vue'
import { Sliders, Save, Check } from 'lucide-vue-next'

const settings = ref<SystemSettings>({
  parallel_pool_size: 15,
  race_timeout: 60,
  stream_idle_timeout_seconds: 20,
  max_retries: 1,
})

const saving = ref(false)
const saved = ref(false)

async function loadSettings() {
  const res = await api.getSettings()
  if (res?.settings) {
    settings.value = {
      ...settings.value,
      ...res.settings,
    }
  }
}

async function handleSave() {
  saving.value = true
  try {
    await api.updateSettings(settings.value)
    saved.value = true
    setTimeout(() => { saved.value = false }, 2000)
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  void loadSettings()
})
</script>

<template>
  <div class="space-y-6 max-w-3xl">
    <div>
      <h2 class="text-xl font-bold text-gray-900 tracking-tight">系统运行配置</h2>
      <p class="text-xs text-gray-500 mt-1">配置并发竞速规模与网络超时阈值</p>
    </div>

    <div class="bg-white border border-gray-200/80 rounded-2xl p-6 shadow-xs space-y-6">
      <div class="flex items-center gap-2 text-xs font-semibold text-gray-500 uppercase tracking-wider pb-3 border-b border-gray-100">
        <Sliders class="w-4 h-4 text-emerald-600" />
        <span>并发与超时控制</span>
      </div>

      <div class="space-y-4 text-xs">
        <div>
          <label class="block font-semibold text-gray-700 mb-1">并发竞速池规模 (Parallel Pool Size)</label>
          <input
            v-model.number="settings.parallel_pool_size"
            type="number"
            min="1"
            max="64"
            class="w-full max-w-xs px-3.5 py-2.5 bg-gray-50 border border-gray-200 rounded-xl outline-none focus:bg-white focus:border-emerald-600 font-mono text-sm"
          >
          <p class="text-gray-400 mt-1">单请求在滑动窗口内允许同时在途的候选节点数量</p>
        </div>

        <div>
          <label class="block font-semibold text-gray-700 mb-1">单节点竞速超时 (Race Timeout, 秒)</label>
          <input
            v-model.number="settings.race_timeout"
            type="number"
            min="1"
            max="60"
            class="w-full max-w-xs px-3.5 py-2.5 bg-gray-50 border border-gray-200 rounded-xl outline-none focus:bg-white focus:border-emerald-600 font-mono text-sm"
          >
          <p class="text-gray-400 mt-1">单个节点等待首包判定的最大等待时间，超时自动淘汰切入下一候选</p>
        </div>

        <div>
          <label class="block font-semibold text-gray-700 mb-1">流式包间超时 (Stream Idle Timeout, 秒)</label>
          <input
            v-model.number="settings.stream_idle_timeout_seconds"
            type="number"
            min="5"
            max="300"
            class="w-full max-w-xs px-3.5 py-2.5 bg-gray-50 border border-gray-200 rounded-xl outline-none focus:bg-white focus:border-emerald-600 font-mono text-sm"
          >
          <p class="text-gray-400 mt-1">流式输出过程中相邻数据块之间的静默超时阈值</p>
        </div>

        <div>
          <label class="block font-semibold text-gray-700 mb-1">换批重试次数 (Max Retries)</label>
          <input
            v-model.number="settings.max_retries"
            type="number"
            min="0"
            max="20"
            class="w-full max-w-xs px-3.5 py-2.5 bg-gray-50 border border-gray-200 rounded-xl outline-none focus:bg-white focus:border-emerald-600 font-mono text-sm"
          >
          <p class="text-gray-400 mt-1">当在途候选全部遇阻时的轮次重试预算</p>
        </div>
      </div>

      <div class="pt-4 border-t border-gray-100 flex items-center gap-3">
        <Button variant="primary" :loading="saving" @click="handleSave">
          <component :is="saved ? Check : Save" class="w-4 h-4" />
          <span>{{ saved ? '保存成功' : '保存设置' }}</span>
        </Button>
        <span v-if="saved" class="text-xs text-emerald-600 font-medium">配置已生效，不影响在途请求</span>
      </div>
    </div>
  </div>
</template>
