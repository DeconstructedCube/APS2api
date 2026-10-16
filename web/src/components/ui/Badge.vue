<script setup lang="ts">
interface Props {
  variant?: 'green' | 'red' | 'amber' | 'blue' | 'gray'
  size?: 'sm' | 'md'
}

withDefaults(defineProps<Props>(), {
  variant: 'gray',
  size: 'md',
})
</script>

<template>
  <span
    class="inline-flex items-center justify-center font-medium rounded-full transition-colors"
    :class="[
      size === 'sm' ? 'px-2 py-0.5 text-xs' : 'px-2.5 py-1 text-xs',
      {
        'bg-emerald-50 text-emerald-700 border border-emerald-200/60': variant === 'green',
        'bg-rose-50 text-rose-700 border border-rose-200/60': variant === 'red',
        'bg-amber-50 text-amber-700 border border-amber-200/60': variant === 'amber',
        'bg-blue-50 text-blue-700 border border-blue-200/60': variant === 'blue',
        'bg-gray-100 text-gray-700 border border-gray-200/60': variant === 'gray',
      },
    ]"
  >
    <slot />
  </span>
</template>
