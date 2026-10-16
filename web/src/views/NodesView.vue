<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { api } from '@/api/client'
import type { NodeItem, NodeHealthItem } from '@/api/client'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'
import {
  Play,
  Pause,
  ArrowUpDown,
  Trash2,
  CheckSquare,
  Download,
  Upload,
  Layers,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
} from 'lucide-vue-next'

const nodes = ref<NodeItem[]>([])
const healthMap = ref<Record<string, NodeHealthItem>>({})
const selectedUris = ref<Set<string>>(new Set())
const singleUri = ref('')
const loading = ref(false)

// 分页
const currentPage = ref(1)
const pageSize = 20

const totalPages = computed(() => {
  return Math.ceil(nodes.value.length / pageSize) || 1
})

const paginatedNodes = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return nodes.value.slice(start, start + pageSize)
})

// 测速进度
const progress = ref({
  running: false,
  paused: false,
  done: 0,
  total: 0,
  ok_count: 0,
  fail_count: 0,
  current_node: '',
})
let progressTimer: ReturnType<typeof setInterval> | null = null

// 去重预览弹窗
const dedupModalOpen = ref(false)
const dedupStats = ref({ groups: 0, duplicate_count: 0 })

// 文件导入
const fileInputRef = ref<HTMLInputElement | null>(null)
const jsonFileInputRef = ref<HTMLInputElement | null>(null)
const isOverwriteImport = ref(false)

async function loadNodes() {
  loading.value = true
  try {
    const res = await api.getNodes()
    nodes.value = res.nodes || []
    healthMap.value = res.health || {}
    if (currentPage.value > totalPages.value) {
      currentPage.value = totalPages.value
    }
  } finally {
    loading.value = false
  }
}

async function handleImportSingle() {
  const text = singleUri.value.trim()
  if (!text) return
  await api.importNodes(text)
  singleUri.value = ''
  await loadNodes()
}

async function handleTestAll() {
  await api.testAllNodes()
  startProgressPoll()
}

function startProgressPoll() {
  if (progressTimer) clearInterval(progressTimer)
  progressTimer = setInterval(async () => {
    const p = await api.getTestProgress()
    progress.value = {
      ...p,
      paused: false,
    }
    if (!p.running) {
      if (progressTimer) {
        clearInterval(progressTimer)
        progressTimer = null
      }
      await loadNodes()
    }
  }, 500)
}

async function togglePauseTest() {
  if (progress.value.paused) {
    await api.resumeTest()
    progress.value.paused = false
  } else {
    await api.pauseTest()
    progress.value.paused = true
  }
}

async function handleTerminateTest() {
  await api.terminateTest()
}

async function handleTestSingle(uri: string) {
  await api.testSingleNode(uri)
  await loadNodes()
}

async function handleSort(desc: boolean) {
  await api.sortNodes(desc)
  await loadNodes()
}

async function handleOpenDedupModal() {
  const preview = await api.previewDedupNodes()
  dedupStats.value = preview
  dedupModalOpen.value = true
}

async function confirmDedup() {
  await api.dedupNodes()
  dedupModalOpen.value = false
  await loadNodes()
}

async function handleDeleteDisabled() {
  if (confirm('确定清理所有处于禁用状态的节点吗？')) {
    await api.deleteDisabledNodes()
    await loadNodes()
  }
}

async function handleExportNodes() {
  if (nodes.value.length === 0) return
  const text = nodes.value.map((n) => n.raw_uri).join('\n')
  const blob = new Blob([text], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `nodes-${new Date().toISOString().slice(0, 10)}.txt`
  a.click()
  URL.revokeObjectURL(url)
}

async function handleToggleNode(node: NodeItem) {
  if (node.disabled) {
    await api.enableNode(node.raw_uri)
  } else {
    await api.batchDisableNodes([node.raw_uri])
  }
  await loadNodes()
}

async function handleDeleteNode(uri: string) {
  if (confirm('确认删除该节点吗？')) {
    await api.deleteNode(uri)
    await loadNodes()
  }
}

function toggleSelect(uri: string) {
  if (selectedUris.value.has(uri)) {
    selectedUris.value.delete(uri)
  } else {
    selectedUris.value.add(uri)
  }
}

function toggleSelectAllCurrentPage() {
  const pageUris = paginatedNodes.value.map((n) => n.raw_uri)
  const allSelected = pageUris.every((u) => selectedUris.value.has(u))
  if (allSelected) {
    pageUris.forEach((u) => selectedUris.value.delete(u))
  } else {
    pageUris.forEach((u) => selectedUris.value.add(u))
  }
}

async function handleBatchEnable() {
  if (selectedUris.value.size === 0) return
  await api.batchEnableNodes(Array.from(selectedUris.value))
  await loadNodes()
}

async function handleBatchDisable() {
  if (selectedUris.value.size === 0) return
  await api.batchDisableNodes(Array.from(selectedUris.value))
  await loadNodes()
}

async function handleBatchDelete() {
  if (selectedUris.value.size === 0) return
  if (confirm(`确定删除已选中的 ${selectedUris.value.size} 个节点？`)) {
    await api.batchDeleteNodes(Array.from(selectedUris.value))
    selectedUris.value.clear()
    await loadNodes()
  }
}

async function handleAddSelectedToProxies() {
  if (selectedUris.value.size === 0) return
  await api.importProxyBatch(Array.from(selectedUris.value))
  alert(`已将选中的 ${selectedUris.value.size} 个节点加入入口代理池`)
}

async function handleRemoveSelectedFromProxies() {
  if (selectedUris.value.size === 0) return
  await api.deleteProxyBatch(Array.from(selectedUris.value))
  alert(`已将选中的节点从入口代理池移除`)
}

// 配置文件上传导入
function triggerFileImport(overwrite: boolean) {
  isOverwriteImport.value = overwrite
  fileInputRef.value?.click()
}

async function onConfigFileSelected(e: Event) {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return
  const text = await file.text()
  if (isOverwriteImport.value) {
    if (nodes.value.length > 0) {
      await api.batchDeleteNodes(nodes.value.map((n) => n.raw_uri))
    }
  }
  await api.importNodes(text)
  target.value = ''
  await loadNodes()
}

// nodes.json 导入
function triggerJsonImport(overwrite: boolean) {
  isOverwriteImport.value = overwrite
  jsonFileInputRef.value?.click()
}

async function onJsonFileSelected(e: Event) {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return
  const text = await file.text()
  if (isOverwriteImport.value) {
    if (nodes.value.length > 0) {
      await api.batchDeleteNodes(nodes.value.map((n) => n.raw_uri))
    }
  }
  await api.importNodesJson(text)
  target.value = ''
  await loadNodes()
}

onMounted(() => {
  void loadNodes()
})

onUnmounted(() => {
  if (progressTimer) clearInterval(progressTimer)
})
</script>

<template>
  <div class="space-y-6">
    <!-- Top Action Bar -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h2 class="text-xl font-bold text-gray-900 tracking-tight">节点池管理</h2>
        <p class="text-xs text-gray-500 mt-1">共 {{ nodes.length }} 个节点参与分发调度</p>
      </div>

      <div class="flex items-center gap-2 flex-wrap">
        <Button variant="primary" size="md" :disabled="progress.running" @click="handleTestAll">
          <Play class="w-4 h-4" />
          <span>批量测速</span>
        </Button>
        <Button variant="secondary" size="md" @click="handleSort(false)">
          <ArrowUpDown class="w-4 h-4" />
          <span>升序</span>
        </Button>
        <Button variant="secondary" size="md" @click="handleSort(true)">
          <ArrowUpDown class="w-4 h-4" />
          <span>降序</span>
        </Button>
        <Button variant="secondary" size="md" @click="handleOpenDedupModal">
          <CheckSquare class="w-4 h-4" />
          <span>去重</span>
        </Button>
        <Button variant="secondary" size="md" @click="handleExportNodes">
          <Download class="w-4 h-4" />
          <span>导出</span>
        </Button>
        <Button variant="outline" size="md" @click="handleDeleteDisabled">
          <Trash2 class="w-4 h-4" />
          <span>清理禁用</span>
        </Button>
      </div>
    </div>

    <!-- Hidden file inputs -->
    <input ref="fileInputRef" type="file" accept=".txt,.yaml,.yml,.json,.conf,.list" class="hidden" @change="onConfigFileSelected">
    <input ref="jsonFileInputRef" type="file" accept=".json" class="hidden" @change="onJsonFileSelected">

    <!-- Progress Card if running -->
    <div v-if="progress.running" class="bg-white border border-emerald-200 rounded-2xl p-4 shadow-xs">
      <div class="flex items-center justify-between text-xs font-semibold text-gray-700 mb-2">
        <span>测速进行中: {{ progress.done }} / {{ progress.total }} (成功: {{ progress.ok_count }}, 失败: {{ progress.fail_count }})</span>
        <div class="space-x-3">
          <button type="button" class="text-gray-600 hover:underline cursor-pointer" @click="togglePauseTest">
            {{ progress.paused ? '继续' : '暂停' }}
          </button>
          <button type="button" class="text-rose-600 hover:underline cursor-pointer" @click="handleTerminateTest">
            终止
          </button>
        </div>
      </div>
      <div class="w-full bg-gray-100 h-2 rounded-full overflow-hidden">
        <div class="bg-emerald-600 h-full transition-all" :style="{ width: (progress.total > 0 ? (progress.done / progress.total) * 100 : 0) + '%' }" />
      </div>
      <div class="text-[11px] text-gray-400 mt-1 truncate">{{ progress.current_node }}</div>
    </div>

    <!-- Single Node Import Bar -->
    <div class="bg-white border border-gray-200/80 rounded-2xl p-4 shadow-xs space-y-3">
      <div class="flex gap-3">
        <input
          v-model="singleUri"
          type="text"
          placeholder="粘贴单节点 URI (vmess://, vless://, ss://, trojan://, hy2://)"
          class="flex-1 px-3.5 py-2 bg-gray-50 border border-gray-200 rounded-xl text-xs outline-none focus:bg-white focus:border-emerald-600 font-mono"
          @keydown.enter.prevent="handleImportSingle"
        >
        <Button variant="secondary" size="sm" @click="handleImportSingle">
          导入单节点
        </Button>
      </div>

      <!-- File Import Actions -->
      <div class="flex items-center gap-2 pt-1 border-t border-gray-100 text-xs text-gray-500">
        <span>配置批量导入：</span>
        <button type="button" class="font-medium text-emerald-600 hover:underline" @click="triggerFileImport(false)">导入 Clash/V2Ray 节点文件</button>
        <span class="text-gray-300">|</span>
        <button type="button" class="font-medium text-rose-600 hover:underline" @click="triggerFileImport(true)">清空并替换</button>
        <span class="text-gray-300">|</span>
        <button type="button" class="font-medium text-gray-600 hover:underline" @click="triggerJsonImport(false)">旧版 nodes.json 导入</button>
      </div>
    </div>

    <!-- Batch Selection Floating Bar -->
    <div v-if="selectedUris.size > 0" class="bg-emerald-50 border border-emerald-200 rounded-xl px-4 py-2.5 flex items-center justify-between text-xs text-emerald-900">
      <span>已选中 <strong>{{ selectedUris.size }}</strong> 个节点</span>
      <div class="flex items-center gap-2 flex-wrap">
        <button type="button" class="px-2.5 py-1 bg-white border border-emerald-300 rounded-lg font-medium text-emerald-800 hover:bg-emerald-100" @click="handleBatchEnable">批量启用</button>
        <button type="button" class="px-2.5 py-1 bg-white border border-emerald-300 rounded-lg font-medium text-emerald-800 hover:bg-emerald-100" @click="handleBatchDisable">批量禁用</button>
        <button type="button" class="px-2.5 py-1 bg-white border border-blue-300 rounded-lg font-medium text-blue-800 hover:bg-blue-100" @click="handleAddSelectedToProxies">加入入口代理</button>
        <button type="button" class="px-2.5 py-1 bg-white border border-gray-300 rounded-lg font-medium text-gray-700 hover:bg-gray-100" @click="handleRemoveSelectedFromProxies">移出入口代理</button>
        <button type="button" class="px-2.5 py-1 bg-rose-50 border border-rose-200 rounded-lg font-medium text-rose-700 hover:bg-rose-100" @click="handleBatchDelete">批量删除</button>
      </div>
    </div>

    <!-- Node Table -->
    <div class="bg-white border border-gray-200/80 rounded-2xl shadow-xs overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs">
          <thead class="bg-gray-50 border-b border-gray-100 text-gray-500 font-medium">
            <tr>
              <th class="px-4 py-3 w-10">
                <input
                  type="checkbox"
                  :checked="paginatedNodes.length > 0 && paginatedNodes.every(n => selectedUris.has(n.raw_uri))"
                  @change="toggleSelectAllCurrentPage"
                >
              </th>
              <th class="px-4 py-3">节点名称</th>
              <th class="px-4 py-3">协议</th>
              <th class="px-4 py-3">延迟 / 状态</th>
              <th class="px-4 py-3">成功 / 失败</th>
              <th class="px-4 py-3 text-right">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100">
            <tr v-if="nodes.length === 0">
              <td colspan="6" class="px-4 py-8 text-center text-gray-400">
                暂无节点，请在上方输入框导入单节点或通过文件上传
              </td>
            </tr>
            <tr v-for="node in paginatedNodes" :key="node.raw_uri" class="hover:bg-gray-50/50">
              <td class="px-4 py-3">
                <input
                  type="checkbox"
                  :checked="selectedUris.has(node.raw_uri)"
                  @change="toggleSelect(node.raw_uri)"
                >
              </td>
              <td class="px-4 py-3 font-medium text-gray-800 max-w-xs truncate" :title="node.name">{{ node.name }}</td>
              <td class="px-4 py-3">
                <Badge variant="blue" size="sm">{{ node.type.toUpperCase() }}</Badge>
              </td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-2">
                  <span v-if="node.disabled" class="text-gray-400 font-medium">已禁用</span>
                  <span v-else-if="healthMap[node.raw_uri]?.consecutive_failures > 0" class="text-rose-600 font-medium">
                    连败 {{ healthMap[node.raw_uri].consecutive_failures }}
                  </span>
                  <span v-else class="text-emerald-600 font-medium">正常</span>

                  <span v-if="healthMap[node.raw_uri]?.last_test_ms > 0" class="font-mono text-gray-500 text-[11px]">
                    {{ Math.round(healthMap[node.raw_uri].last_test_ms) }}ms
                  </span>
                </div>
              </td>
              <td class="px-4 py-3 text-gray-500 font-mono">
                {{ healthMap[node.raw_uri]?.success_count || 0 }} / {{ healthMap[node.raw_uri]?.fail_count || 0 }}
              </td>
              <td class="px-4 py-3 text-right space-x-2">
                <button
                  type="button"
                  class="text-blue-600 font-medium hover:underline"
                  @click="handleTestSingle(node.raw_uri)"
                >
                  测速
                </button>
                <button
                  type="button"
                  class="font-medium hover:underline"
                  :class="node.disabled ? 'text-emerald-600' : 'text-gray-500'"
                  @click="handleToggleNode(node)"
                >
                  {{ node.disabled ? '启用' : '禁用' }}
                </button>
                <button
                  type="button"
                  class="text-rose-600 font-medium hover:underline"
                  @click="handleDeleteNode(node.raw_uri)"
                >
                  删除
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination Footer -->
      <div v-if="totalPages > 1" class="px-4 py-3 bg-gray-50 border-t border-gray-100 flex items-center justify-between text-xs text-gray-500">
        <div>第 {{ currentPage }} / {{ totalPages }} 页（共 {{ nodes.length }} 个节点）</div>
        <div class="flex items-center gap-1">
          <button type="button" :disabled="currentPage <= 1" class="p-1 rounded hover:bg-gray-200 disabled:opacity-40" @click="currentPage = 1">
            <ChevronsLeft class="w-4 h-4" />
          </button>
          <button type="button" :disabled="currentPage <= 1" class="p-1 rounded hover:bg-gray-200 disabled:opacity-40" @click="currentPage--">
            <ChevronLeft class="w-4 h-4" />
          </button>
          <span class="px-2 font-medium text-gray-700">{{ currentPage }}</span>
          <button type="button" :disabled="currentPage >= totalPages" class="p-1 rounded hover:bg-gray-200 disabled:opacity-40" @click="currentPage++">
            <ChevronRight class="w-4 h-4" />
          </button>
          <button type="button" :disabled="currentPage >= totalPages" class="p-1 rounded hover:bg-gray-200 disabled:opacity-40" @click="currentPage = totalPages">
            <ChevronsRight class="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>

    <!-- Dedup Preview Modal -->
    <div v-if="dedupModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-gray-900/50 backdrop-blur-xs p-4">
      <div class="bg-white rounded-2xl p-6 w-full max-w-sm shadow-2xl border border-gray-100 space-y-4">
        <h3 class="text-base font-bold text-gray-900">节点去重确认</h3>
        <p class="text-xs text-gray-600 leading-relaxed">
          检测到 <strong>{{ dedupStats.groups }}</strong> 组重复节点，共计 <strong>{{ dedupStats.duplicate_count }}</strong> 个重复项。
          合并操作将保留每个重复组内最具代表性的节点。
        </p>
        <div class="flex justify-end gap-2.5 pt-2">
          <Button variant="ghost" @click="dedupModalOpen = false">取消</Button>
          <Button variant="danger" @click="confirmDedup">确认合并去重</Button>
        </div>
      </div>
    </div>
  </div>
</template>
