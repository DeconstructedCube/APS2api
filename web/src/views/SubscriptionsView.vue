<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api/client'
import type { SubscriptionItem, CustomUAItem } from '@/api/client'
import Button from '@/components/ui/Button.vue'
import { Plus, RefreshCw, Trash2, Shield } from 'lucide-vue-next'

const subscriptions = ref<SubscriptionItem[]>([])
const customUAs = ref<CustomUAItem[]>([])

// 添加订阅弹窗
const showAddSub = ref(false)
const subName = ref('')
const subUrl = ref('')
const subUa = ref('')
const subInterval = ref(0)
const subAdoptManual = ref(false)

// 删除订阅弹窗
const showDeleteModal = ref(false)
const deletingSubId = ref('')
const deleteExclusiveNodes = ref(true)

// 自定义 UA 弹窗
const showAddUa = ref(false)
const uaName = ref('')
const uaContent = ref('')

async function loadData() {
  const res = await api.getSubscriptions()
  subscriptions.value = res.subscriptions || []
  customUAs.value = res.custom_uas || []
}

async function handleSaveSub() {
  if (!subName.value.trim() || !subUrl.value.trim()) return
  await api.saveSubscription({
    name: subName.value.trim(),
    url: subUrl.value.trim(),
    user_agent: subUa.value,
    interval_minutes: Number(subInterval.value) || 0,
  })
  showAddSub.value = false
  subName.value = ''
  subUrl.value = ''
  subInterval.value = 0
  subAdoptManual.value = false
  await loadData()
}

async function handleUpdateAll() {
  await api.updateSubscriptions()
  await loadData()
}

async function handleUpdateSingle(id: string) {
  await api.updateSubscriptions(id)
  await loadData()
}

function openDeleteModal(id: string) {
  deletingSubId.value = id
  deleteExclusiveNodes.value = true
  showDeleteModal.value = true
}

async function confirmDeleteSub() {
  if (!deletingSubId.value) return
  await api.deleteSubscription(deletingSubId.value, deleteExclusiveNodes.value)
  showDeleteModal.value = false
  deletingSubId.value = ''
  await loadData()
}

async function handleSaveUA() {
  if (!uaName.value.trim() || !uaContent.value.trim()) return
  await api.saveCustomUA(uaName.value.trim(), uaContent.value.trim())
  showAddUa.value = false
  uaName.value = ''
  uaContent.value = ''
  await loadData()
}

async function handleDeleteUA(id: string) {
  if (confirm('确认删除该自定义 UA 吗？')) {
    await api.deleteCustomUA(id)
    await loadData()
  }
}

onMounted(() => {
  void loadData()
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h2 class="text-xl font-bold text-gray-900 tracking-tight">订阅与 UA 管理</h2>
        <p class="text-xs text-gray-500 mt-1">管理自动更新订阅链接与自定义 User-Agent</p>
      </div>

      <div class="flex items-center gap-2">
        <Button variant="primary" size="md" @click="showAddSub = true">
          <Plus class="w-4 h-4" />
          <span>添加订阅</span>
        </Button>
        <Button variant="secondary" size="md" @click="handleUpdateAll">
          <RefreshCw class="w-4 h-4" />
          <span>更新全部订阅</span>
        </Button>
        <Button variant="outline" size="md" @click="showAddUa = true">
          <Shield class="w-4 h-4" />
          <span>添加自定义 UA</span>
        </Button>
      </div>
    </div>

    <!-- Subscriptions Table -->
    <div class="bg-white border border-gray-200/80 rounded-2xl shadow-xs overflow-hidden">
      <div class="px-5 py-4 border-b border-gray-100 flex items-center justify-between">
        <h3 class="text-sm font-bold text-gray-900">订阅列表</h3>
        <span class="text-xs text-gray-400">共 {{ subscriptions.length }} 个订阅</span>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs">
          <thead class="bg-gray-50 border-b border-gray-100 text-gray-500 font-medium">
            <tr>
              <th class="px-4 py-3">名称</th>
              <th class="px-4 py-3">订阅 URL</th>
              <th class="px-4 py-3">绑定 User-Agent</th>
              <th class="px-4 py-3">自动更新周期</th>
              <th class="px-4 py-3">最后更新时间</th>
              <th class="px-4 py-3 text-right">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100">
            <tr v-if="subscriptions.length === 0">
              <td colspan="6" class="px-4 py-8 text-center text-gray-400">
                暂无订阅配置，点击右上角“添加订阅”录入链接
              </td>
            </tr>
            <tr v-for="s in subscriptions" :key="s.id" class="hover:bg-gray-50/50">
              <td class="px-4 py-3 font-medium text-gray-800">{{ s.name }}</td>
              <td class="px-4 py-3 font-mono text-gray-500 max-w-xs truncate" :title="s.url">{{ s.url }}</td>
              <td class="px-4 py-3 text-gray-600">{{ s.user_agent || '默认 (Chrome)' }}</td>
              <td class="px-4 py-3 text-gray-600">{{ s.interval_minutes > 0 ? `${s.interval_minutes} 分钟` : '手动触发' }}</td>
              <td class="px-4 py-3 text-gray-400">{{ s.last_updated_at ? new Date(s.last_updated_at * 1000).toLocaleString() : '未更新' }}</td>
              <td class="px-4 py-3 text-right space-x-2">
                <button type="button" class="text-emerald-600 font-medium hover:underline" @click="handleUpdateSingle(s.id)">更新</button>
                <button type="button" class="text-rose-600 font-medium hover:underline" @click="openDeleteModal(s.id)">删除</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Custom UA Table -->
    <div class="bg-white border border-gray-200/80 rounded-2xl shadow-xs overflow-hidden">
      <div class="px-5 py-4 border-b border-gray-100 flex items-center justify-between">
        <h3 class="text-sm font-bold text-gray-900">自定义 User-Agent 库</h3>
        <span class="text-xs text-gray-400">拉取订阅时使用的防屏蔽请求头</span>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs">
          <thead class="bg-gray-50 border-b border-gray-100 text-gray-500 font-medium">
            <tr>
              <th class="px-4 py-3">标识名称</th>
              <th class="px-4 py-3">User-Agent 标头值</th>
              <th class="px-4 py-3 text-right">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100">
            <tr v-if="customUAs.length === 0">
              <td colspan="3" class="px-4 py-6 text-center text-gray-400">暂无自定义 UA，系统默认使用 Chrome/Clash/V2Ray 轮询</td>
            </tr>
            <tr v-for="u in customUAs" :key="u.id" class="hover:bg-gray-50/50">
              <td class="px-4 py-3 font-medium text-gray-800">{{ u.name }}</td>
              <td class="px-4 py-3 font-mono text-gray-600 max-w-md truncate" :title="u.content">{{ u.content }}</td>
              <td class="px-4 py-3 text-right">
                <button type="button" class="text-rose-600 font-medium hover:underline" @click="handleDeleteUA(u.id)">删除</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Modal: Add Subscription -->
    <div v-if="showAddSub" class="fixed inset-0 z-50 flex items-center justify-center bg-gray-900/50 backdrop-blur-xs p-4">
      <div class="bg-white rounded-2xl p-6 w-full max-w-md shadow-2xl border border-gray-100 space-y-4">
        <h3 class="text-base font-bold text-gray-900">添加订阅</h3>
        <div>
          <label class="block text-xs font-semibold text-gray-600 mb-1">订阅名称</label>
          <input v-model="subName" type="text" placeholder="例如：主力订阅" class="w-full px-3.5 py-2.5 bg-gray-50 border border-gray-200 rounded-xl text-sm outline-none focus:bg-white focus:border-emerald-600">
        </div>
        <div>
          <label class="block text-xs font-semibold text-gray-600 mb-1">订阅链接</label>
          <input v-model="subUrl" type="text" placeholder="https://..." class="w-full px-3.5 py-2.5 bg-gray-50 border border-gray-200 rounded-xl text-sm outline-none focus:bg-white focus:border-emerald-600 font-mono">
        </div>
        <div>
          <label class="block text-xs font-semibold text-gray-600 mb-1">绑定 User-Agent</label>
          <select v-model="subUa" class="w-full px-3.5 py-2.5 bg-gray-50 border border-gray-200 rounded-xl text-sm outline-none focus:bg-white focus:border-emerald-600">
            <option value="">默认 (Chrome / Clash 轮询)</option>
            <option value="clash-verge/v2.5.2">clash-verge/v2.5.2</option>
            <option value="Clash.Meta">Clash.Meta</option>
            <option value="v2rayNG/1.8.5">v2rayNG/1.8.5</option>
            <option v-for="u in customUAs" :key="u.id" :value="u.content">{{ u.name }}</option>
          </select>
        </div>
        <div>
          <label class="block text-xs font-semibold text-gray-600 mb-1">自动更新周期 (分钟，0 为手动)</label>
          <input v-model="subInterval" type="number" min="0" class="w-full px-3.5 py-2.5 bg-gray-50 border border-gray-200 rounded-xl text-sm outline-none focus:bg-white focus:border-emerald-600">
        </div>
        <div class="pt-1">
          <label class="flex items-start gap-2 cursor-pointer text-xs text-gray-600 select-none">
            <input v-model="subAdoptManual" type="checkbox" class="mt-0.5 rounded border-gray-300 text-emerald-600">
            <span><strong>接管手动节点：</strong>相同 URI 的手动节点将采用该订阅元数据，随订阅更新和删除</span>
          </label>
        </div>
        <div class="flex justify-end gap-2.5 pt-2">
          <Button variant="ghost" @click="showAddSub = false">取消</Button>
          <Button variant="primary" @click="handleSaveSub">保存订阅</Button>
        </div>
      </div>
    </div>

    <!-- Modal: Delete Subscription with Option -->
    <div v-if="showDeleteModal" class="fixed inset-0 z-50 flex items-center justify-center bg-gray-900/50 backdrop-blur-xs p-4">
      <div class="bg-white rounded-2xl p-6 w-full max-w-sm shadow-2xl border border-gray-100 space-y-4">
        <h3 class="text-base font-bold text-gray-900">确认删除订阅</h3>
        <p class="text-xs text-gray-600 leading-relaxed">确定要删除该订阅吗？</p>
        <div class="p-3 bg-gray-50 border border-gray-200 rounded-xl">
          <label class="flex items-start gap-2 cursor-pointer text-xs text-gray-700 select-none">
            <input v-model="deleteExclusiveNodes" type="checkbox" class="mt-0.5 rounded border-gray-300 text-emerald-600">
            <span><strong>同时删除独占节点</strong><br><span class="text-gray-400 text-[11px]">取消勾选后，该订阅独占的节点将转为手动节点保留。</span></span>
          </label>
        </div>
        <div class="flex justify-end gap-2.5 pt-2">
          <Button variant="ghost" @click="showDeleteModal = false">取消</Button>
          <Button variant="danger" @click="confirmDeleteSub">确认删除</Button>
        </div>
      </div>
    </div>

    <!-- Modal: Add Custom UA -->
    <div v-if="showAddUa" class="fixed inset-0 z-50 flex items-center justify-center bg-gray-900/50 backdrop-blur-xs p-4">
      <div class="bg-white rounded-2xl p-6 w-full max-w-md shadow-2xl border border-gray-100 space-y-4">
        <h3 class="text-base font-bold text-gray-900">添加自定义 User-Agent</h3>
        <div>
          <label class="block text-xs font-semibold text-gray-600 mb-1">标识名称</label>
          <input v-model="uaName" type="text" placeholder="例如：ClashVerge" class="w-full px-3.5 py-2.5 bg-gray-50 border border-gray-200 rounded-xl text-sm outline-none focus:bg-white focus:border-emerald-600">
        </div>
        <div>
          <label class="block text-xs font-semibold text-gray-600 mb-1">User-Agent 内容</label>
          <input v-model="uaContent" type="text" placeholder="clash-verge/v1.0.0" class="w-full px-3.5 py-2.5 bg-gray-50 border border-gray-200 rounded-xl text-sm outline-none focus:bg-white focus:border-emerald-600 font-mono">
        </div>
        <div class="flex justify-end gap-2.5 pt-2">
          <Button variant="ghost" @click="showAddUa = false">取消</Button>
          <Button variant="primary" @click="handleSaveUA">添加 UA</Button>
        </div>
      </div>
    </div>
  </div>
</template>
