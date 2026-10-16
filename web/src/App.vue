<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'

const route = useRoute()
const isLoginPage = computed(() => route.meta.public === true)
</script>

<template>
  <div class="min-h-screen bg-gray-50 text-gray-900">
    <router-view v-if="isLoginPage" />
    <AppLayout v-else>
      <router-view />
    </AppLayout>
  </div>
</template>
