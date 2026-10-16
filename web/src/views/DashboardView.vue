<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api/client'
import StatCard from '@/components/ui/StatCard.vue'
import { Server, Zap, Activity, AlertTriangle } from 'lucide-vue-next'

const activeNodesCount = ref(0)
const totalNodesCount = ref(0)
const poolSize = ref(1)
const avgLatency = ref('-')
const coolingCount = ref(0)
const nodesList = ref<any[]>([])

async function loadData() {
  try {
    const nodesRes = await api.getNodes()
    const nodes = nodesRes.nodes || []
    totalNodesCount.value = nodes.length
    nodesList.value = nodes

    const now = Math.floor(Date.now() / 1000)
    let cooling = 0
    let totalMs = 0
    let testedCount = 0

    nodes.forEach((n) => {
      const h = nodesRes.health[n.raw_uri]
      if (h) {
        if (h.cooldown_until > now) cooling++
        if (h.last_test_ms > 0) {
          totalMs += h.last_test_ms
          testedCount++
        }
      }
    })

    coolingCount.value = cooling
    activeNodesCount.value = nodes.filter((n) => !n.disabled && (!nodesRes.health[n.raw_uri]?.cooldown_until || nodesRes.health[n.raw_uri].cooldown_until <= now)).length
    avgLatency.value = testedCount > 0 ? `${Math.round(totalMs / testedCount)} ms` : '-'

    const settings = await api.getSettings()
    poolSize.value = settings.parallel_pool_size || 1
  } catch (e) {
    console.error(e)
  }
}

onMounted(() => {
  void loadData()
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h2 class="text-xl font-bold text-gray-900 tracking-tight">控制面板</h2>
    </div>

    <!-- 4 个核心数据卡片 -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <StatCard
        label="可用节点"
        :value="activeNodesCount"
        :sub-value="`总计 ${totalNodesCount}`"
        color="green"
      >
        <template #icon>
          <Server class="w-4 h-4 text-emerald-600" />
        </template>
      </StatCard>

      <StatCard
        label="并发竞速池规模"
        :value="poolSize"
        sub-value="配置并发"
        color="blue"
      >
        <template #icon>
          <Zap class="w-4 h-4 text-blue-600" />
        </template>
      </StatCard>

      <StatCard
        label="平均探测延迟"
        :value="avgLatency"
        color="default"
      >
        <template #icon>
          <Activity class="w-4 h-4 text-gray-400" />
        </template>
      </StatCard>

      <StatCard
        label="429 频控次数"
        :value="coolingCount"
        :color="coolingCount > 0 ? 'amber' : 'default'"
      >
        <template #icon>
          <AlertTriangle class="w-4 h-4 text-amber-500" />
        </template>
      </StatCard>
    </div>

    <!-- 快速概要表格 -->
    <div class="bg-white border border-gray-200/80 rounded-2xl shadow-xs overflow-hidden">
      <div class="px-5 py-4 border-b border-gray-100 flex items-center justify-between">
        <h3 class="text-sm font-bold text-gray-900">节点概览</h3>
        <router-link to="/nodes" class="text-xs font-medium text-emerald-600 hover:text-emerald-700">进入节点管理 &rarr;</router-link>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs">
          <thead class="bg-gray-50 border-b border-gray-100 text-gray-500 font-medium">
            <tr>
              <th class="px-4 py-3">节点名称</th>
              <th class="px-4 py-3">协议</th>
              <th class="px-4 py-3">状态</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100">
            <tr v-if="nodesList.length === 0">
              <td colspan="3" class="px-4 py-6 text-center text-gray-400">暂无节点数据</td>
            </tr>
            <tr v-for="node in nodesList.slice(0, 5)" :key="node.raw_uri" class="hover:bg-gray-50/50">
              <td class="px-4 py-3 font-medium text-gray-800">{{ node.name }}</td>
              <td class="px-4 py-3 font-mono text-gray-500 uppercase">{{ node.type }}</td>
              <td class="px-4 py-3">
                <span v-if="node.disabled" class="text-gray-400">已禁用</span>
                <span v-else class="text-emerald-600 font-medium">正常</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
