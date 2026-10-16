export interface ApiResponse<T = unknown> {
  ok?: boolean
  error?: { message: string }
  data?: T
}

export interface NodeItem {
  type: string
  name: string
  raw_uri: string
  disabled: boolean
}

export interface NodeHealthItem {
  success_count: number
  fail_count: number
  consecutive_failures: number
  last_test_ms: number
  last_test_error: string
  cooldown_until: number
}

export interface ProxyCandidate {
  name: string
  raw_uri: string
  normalized_uri: string
  type: string
  disabled: boolean
  last_test_ms: number
  last_test_ok: boolean
}

export interface SubscriptionItem {
  id: string
  name: string
  url: string
  user_agent?: string
  custom_ua_id?: string
  update_interval: number
  adopt_manual?: boolean
  last_update_time?: number
  last_error?: string
}

export interface CustomUAItem {
  id: string
  name: string
  user_agent: string
}

export interface ApiKeyItem {
  name: string
  key: string
  description?: string
}

export interface SystemSettings {
  parallel_pool_size: number
  race_timeout: number
  stream_idle_timeout_seconds: number
  max_retries: number
  [key: string]: unknown
}

async function request<T = unknown>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers || {})
  if (!headers.has('Content-Type') && options.body && typeof options.body === 'string') {
    headers.set('Content-Type', 'application/json')
  }

  const res = await fetch('/api/admin' + path, {
    ...options,
    headers,
  })

  const data = (await res.json()) as T
  return data
}

export const api = {
  // 鉴权
  checkAuth: () => request<{ ok: boolean }>('/check-auth'),
  login: (password: string) => request<{ ok: boolean }>('/login', {
    method: 'POST',
    body: JSON.stringify({ password }),
  }),
  logout: () => request<{ ok: boolean }>('/logout', { method: 'POST' }),
  changePassword: (old_password: string, new_password: string) =>
    request<{ ok: boolean }>('/password', {
      method: 'POST',
      body: JSON.stringify({ old_password, new_password }),
    }),

  // 仪表盘与配置
  getStats: () => request<Record<string, unknown>>('/stats'),
  getSettings: () => request<{ settings: SystemSettings }>('/settings'),
  updateSettings: (payload: SystemSettings) => request<{ ok: boolean }>('/settings', {
    method: 'PUT',
    body: JSON.stringify({ settings: payload }),
  }),

  // 节点池
  getNodes: () => request<{ nodes: NodeItem[]; health: Record<string, NodeHealthItem> }>('/nodes'),
  deleteNode: (uri: string) => request<{ ok: boolean }>('/nodes', {
    method: 'DELETE',
    body: JSON.stringify({ raw_uri: uri }),
  }),
  enableNode: (uri: string) => request<{ ok: boolean }>('/nodes/enable', {
    method: 'POST',
    body: JSON.stringify({ raw_uri: uri }),
  }),
  importNodes: (text: string) => request('/nodes/import', {
    method: 'POST',
    body: JSON.stringify({ text }),
  }),
  testSingleNode: (uri: string) => request<{ ok: boolean; elapsed_ms: number; error: string; disabled: boolean }>('/nodes/test', {
    method: 'POST',
    body: JSON.stringify({ raw_uri: uri }),
  }),
  testAllNodes: () => request('/nodes/test-all', { method: 'POST' }),
  getTestProgress: () => request<{
    running: boolean
    done: number
    total: number
    ok_count: number
    fail_count: number
    current_node: string
  }>('/nodes/test-progress'),
  pauseTest: () => request('/nodes/test-pause', { method: 'POST' }),
  resumeTest: () => request('/nodes/test-resume', { method: 'POST' }),
  terminateTest: () => request('/nodes/test-terminate', { method: 'POST' }),
  sortNodes: (desc: boolean) => request<{ ok: boolean }>('/nodes/sort', {
    method: 'POST',
    body: JSON.stringify({ desc }),
  }),
  dedupNodes: () => request('/nodes/deduplicate', { method: 'POST' }),
  previewDedupNodes: () => request<{ groups: number; duplicate_count: number }>('/nodes/deduplicate/preview'),
  importNodesJson: (text: string) => request('/nodes/import-json', {
    method: 'POST',
    body: JSON.stringify({ text }),
  }),
  deleteDisabledNodes: () => request('/nodes/disabled', { method: 'DELETE' }),
  batchEnableNodes: (uris: string[]) => request('/nodes/batch-enable', {
    method: 'POST',
    body: JSON.stringify({ uris }),
  }),
  batchDisableNodes: (uris: string[]) => request('/nodes/batch-disable', {
    method: 'POST',
    body: JSON.stringify({ uris }),
  }),
  batchDeleteNodes: (uris: string[]) => request('/nodes/batch-delete', {
    method: 'POST',
    body: JSON.stringify({ uris }),
  }),

  // 入口代理
  getProxies: () => request<{ candidates: ProxyCandidate[] }>('/proxy-nodes'),
  importProxy: (raw_uri: string) => request('/proxy-nodes/import', {
    method: 'POST',
    body: JSON.stringify({ raw_uri }),
  }),
  importProxyBatch: (uris: string[]) => request('/proxy-nodes/import-batch', {
    method: 'POST',
    body: JSON.stringify({ uris }),
  }),
  deleteProxyBatch: (uris: string[]) => request('/proxy-nodes/delete-batch', {
    method: 'POST',
    body: JSON.stringify({ uris }),
  }),
  deleteProxy: (uri: string) => request<{ ok: boolean; was_active: boolean }>('/proxy-nodes', {
    method: 'DELETE',
    body: JSON.stringify({ raw_uri: uri }),
  }),
  testBatchProxies: () => request('/proxy-nodes/test-batch', { method: 'POST' }),
  batchEnableProxies: () => request('/proxy-nodes/enable-batch', { method: 'POST' }),
  batchDisableProxies: () => request('/proxy-nodes/disable-batch', { method: 'POST' }),
  deleteDisabledProxies: () => request('/proxy-nodes/disabled', { method: 'DELETE' }),

  // 订阅与 UA
  getSubscriptions: () => request<{ subscriptions: SubscriptionItem[]; custom_uas: CustomUAItem[] }>('/subscriptions/list'),
  saveSubscription: (payload: { name: string; url: string; user_agent?: string; update_interval: number; adopt_manual?: boolean }) =>
    request<{ ok: boolean; id: string }>('/subscriptions/save', {
      method: 'POST',
      body: JSON.stringify(payload),
    }),
  deleteSubscription: (id: string, delete_nodes: boolean = true) => request('/subscriptions/delete', {
    method: 'POST',
    body: JSON.stringify({ id, delete_nodes }),
  }),
  updateSubscriptions: (id?: string) => request('/subscriptions/update', {
    method: 'POST',
    body: JSON.stringify(id ? { id } : {}),
  }),
  saveCustomUA: (name: string, user_agent: string) => request<{ ok: boolean; id: string }>('/subscriptions/custom_ua/save', {
    method: 'POST',
    body: JSON.stringify({ name, user_agent }),
  }),
  deleteCustomUA: (id: string) => request('/subscriptions/custom_ua/delete', {
    method: 'POST',
    body: JSON.stringify({ id }),
  }),

  // API 密钥
  getKeys: () => request<{ keys: ApiKeyItem[] }>('/keys'),
  addKey: (name: string, key: string) => request('/keys', {
    method: 'POST',
    body: JSON.stringify({ name, key }),
  }),
  deleteKey: (name: string) => request('/keys/' + encodeURIComponent(name), { method: 'DELETE' }),

  // 日志
  getLogs: () => request<{ ok: boolean; content: string }>('/log'),
}
