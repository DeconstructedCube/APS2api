import { createRouter, createWebHashHistory } from 'vue-router'
import { api } from '@/api/client'

const routes = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/LoginView.vue'),
    meta: { public: true },
  },
  {
    path: '/',
    name: 'dashboard',
    component: () => import('@/views/DashboardView.vue'),
  },
  {
    path: '/nodes',
    name: 'nodes',
    component: () => import('@/views/NodesView.vue'),
  },
  {
    path: '/proxies',
    name: 'proxies',
    component: () => import('@/views/ProxiesView.vue'),
  },
  {
    path: '/subscriptions',
    name: 'subscriptions',
    component: () => import('@/views/SubscriptionsView.vue'),
  },
  {
    path: '/keys',
    name: 'keys',
    component: () => import('@/views/KeysView.vue'),
  },
  {
    path: '/settings',
    name: 'settings',
    component: () => import('@/views/SettingsView.vue'),
  },
  {
    path: '/logs',
    name: 'logs',
    component: () => import('@/views/LogsView.vue'),
  },
]

export const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

router.beforeEach(async (to, _from, next) => {
  if (to.meta.public) {
    return next()
  }
  try {
    const auth = await api.checkAuth()
    if (auth.ok) {
      next()
    } else {
      next({ name: 'login' })
    }
  } catch {
    next({ name: 'login' })
  }
})
