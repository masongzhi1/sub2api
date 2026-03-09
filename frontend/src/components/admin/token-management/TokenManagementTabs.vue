<template>
  <div class="rounded-2xl border border-gray-200 bg-white p-1.5 shadow-sm dark:border-dark-700 dark:bg-dark-800">
    <div class="flex flex-wrap gap-2">
      <router-link
        v-for="item in items"
        :key="item.path"
        :to="item.path"
        class="inline-flex min-h-[44px] items-center rounded-xl px-4 py-2 text-sm font-medium transition-colors"
        :class="route.path === item.path
          ? 'bg-primary-600 text-white shadow-sm'
          : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900 dark:text-dark-300 dark:hover:bg-dark-700 dark:hover:text-white'"
        :aria-current="route.path === item.path ? 'page' : undefined"
      >
        {{ item.label }}
      </router-link>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'

const route = useRoute()
const { t } = useI18n()

const items = computed(() => [
  { path: '/admin/subscriptions', label: t('nav.subscriptions') },
  { path: '/admin/tokens', label: t('nav.tokenManagement') }
])
</script>
