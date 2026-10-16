<script setup lang="ts">
import { Loader2 } from 'lucide-vue-next'

interface Props {
  variant?: 'primary' | 'secondary' | 'danger' | 'ghost' | 'outline'
  size?: 'sm' | 'md' | 'lg'
  loading?: boolean
  disabled?: boolean
  type?: 'button' | 'submit' | 'reset'
}

withDefaults(defineProps<Props>(), {
  variant: 'secondary',
  size: 'md',
  loading: false,
  disabled: false,
  type: 'button',
})
</script>

<template>
  <button
    :type="type"
    :disabled="disabled || loading"
    class="inline-flex items-center justify-center font-medium rounded-xl transition-all select-none disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer active:scale-[0.98]"
    :class="[
      size === 'sm' ? 'px-3 py-1.5 text-xs gap-1.5' : size === 'lg' ? 'px-5 py-2.5 text-base gap-2.5' : 'px-4 py-2 text-sm gap-2',
      {
        'bg-emerald-600 hover:bg-emerald-700 text-white shadow-xs focus:ring-2 focus:ring-emerald-400/40': variant === 'primary',
        'bg-white hover:bg-gray-50 text-gray-700 border border-gray-200 shadow-xs focus:ring-2 focus:ring-gray-200': variant === 'secondary',
        'bg-rose-50 hover:bg-rose-100 text-rose-700 border border-rose-200 focus:ring-2 focus:ring-rose-200': variant === 'danger',
        'hover:bg-gray-100 text-gray-600 focus:ring-2 focus:ring-gray-200': variant === 'ghost',
        'border border-gray-300 text-gray-700 hover:bg-gray-50': variant === 'outline',
      },
    ]"
  >
    <Loader2
      v-if="loading"
      class="w-4 h-4 animate-spin shrink-0"
    />
    <slot />
  </button>
</template>
