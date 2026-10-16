<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api/client'
import Button from '@/components/ui/Button.vue'
import { Lock, Eye, EyeOff, ShieldCheck, AlertCircle } from 'lucide-vue-next'

const router = useRouter()
const password = ref('')
const showPassword = ref(false)
const loading = ref(false)
const errorMsg = ref('')

async function handleLogin() {
  const pw = password.value.trim()
  if (!pw) {
    errorMsg.value = '请输入控制台访问密码'
    return
  }

  errorMsg.value = ''
  loading.value = true

  try {
    const res = await api.login(pw)
    if (res.ok) {
      await router.replace({ name: 'dashboard' })
    } else {
      errorMsg.value = res.error?.message || '密码错误或无权访问'
    }
  } catch {
    errorMsg.value = '网络请求异常，请检查服务状态'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen bg-gray-100/90 flex flex-col justify-center items-center p-4 sm:p-6">
    <div class="w-full max-w-md bg-white rounded-3xl shadow-xl border border-gray-100 p-8 sm:p-10 transition-all">
      <!-- Header / Logo -->
      <div class="text-center mb-8">
        <div class="w-14 h-14 bg-emerald-600 text-white rounded-2xl flex items-center justify-center font-bold text-2xl mx-auto mb-4 shadow-md shadow-emerald-600/20">
          V
        </div>
        <h1 class="text-2xl font-extrabold text-gray-900 tracking-tight">
          APS2api 控制台
        </h1>
        <p class="text-xs text-gray-500 mt-1.5">
          请输入控制台管理密码
        </p>
      </div>

      <!-- Error Alert -->
      <div
        v-if="errorMsg"
        class="mb-6 p-3.5 bg-rose-50 border border-rose-200/80 rounded-xl text-xs font-medium text-rose-800 flex items-start gap-2.5"
      >
        <AlertCircle class="w-4 h-4 text-rose-600 shrink-0 mt-0.5" />
        <span>{{ errorMsg }}</span>
      </div>

      <!-- Form -->
      <form
        class="space-y-5"
        @submit.prevent="handleLogin"
      >
        <div class="space-y-1.5">
          <label class="block text-xs font-semibold text-gray-700 uppercase tracking-wider">
            访问密码
          </label>
          <div class="relative">
            <div class="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-gray-400">
              <Lock class="w-4 h-4" />
            </div>
            <input
              v-model="password"
              :type="showPassword ? 'text' : 'password'"
              placeholder="输入访问密码"
              autocomplete="current-password"
              autofocus
              class="w-full pl-10 pr-10 py-3 bg-gray-50 border border-gray-200 rounded-xl text-sm outline-none focus:bg-white focus:border-emerald-600 focus:ring-2 focus:ring-emerald-600/20 transition-all font-mono"
            >
            <button
              type="button"
              class="absolute inset-y-0 right-0 pr-3.5 flex items-center text-gray-400 hover:text-gray-600 transition-colors"
              @click="showPassword = !showPassword"
            >
              <EyeOff
                v-if="showPassword"
                class="w-4 h-4"
              />
              <Eye
                v-else
                class="w-4 h-4"
              />
            </button>
          </div>
        </div>

        <Button
          type="submit"
          variant="primary"
          size="lg"
          class="w-full justify-center text-sm font-semibold shadow-md shadow-emerald-600/10"
          :loading="loading"
          :disabled="!password.trim()"
        >
          <ShieldCheck class="w-4 h-4" />
          <span>登录</span>
        </Button>
      </form>
    </div>
  </div>
</template>
