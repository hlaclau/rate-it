<script setup lang="ts">
import type { UserSummary } from '~/composables/useUsers'

const { searchUsers } = useUsers()

const query = ref('')
const results = ref<UserSummary[]>([])
const loading = ref(false)
const searched = ref(false)

let debounceTimer: ReturnType<typeof setTimeout> | null = null

watch(query, (val) => {
  if (debounceTimer) clearTimeout(debounceTimer)
  if (val.length < 2) {
    results.value = []
    searched.value = false
    return
  }
  debounceTimer = setTimeout(async () => {
    loading.value = true
    searched.value = true
    try {
      results.value = await searchUsers(val)
    } finally {
      loading.value = false
    }
  }, 300)
})

useSeoMeta({ title: 'Community — Rate It' })
</script>

<template>
  <UContainer class="py-10 max-w-2xl">
    <h1 class="text-3xl font-bold mb-2">Community</h1>
    <p class="text-muted mb-8">Find users and explore their lists.</p>

    <UInput
      v-model="query"
      placeholder="Search by username…"
      leading-icon="i-lucide-search"
      size="lg"
      class="mb-8"
    />

    <!-- Loading -->
    <div v-if="loading" class="space-y-3">
      <USkeleton v-for="i in 4" :key="i" class="h-14 rounded-xl" />
    </div>

    <!-- Results -->
    <div v-else-if="results.length > 0" class="space-y-2">
      <NuxtLink
        v-for="u in results"
        :key="u.id"
        :to="`/list/${u.username}`"
        class="flex items-center gap-4 p-4 rounded-xl bg-elevated ring-1 ring-default hover:ring-primary transition-all"
      >
        <UAvatar
          :src="u.avatar_url ?? undefined"
          :alt="u.username"
          size="md"
        />
        <span class="font-medium">{{ u.username }}</span>
        <UIcon
          name="i-lucide-chevron-right"
          class="ml-auto size-4 text-muted"
        />
      </NuxtLink>
    </div>

    <!-- No results -->
    <div
      v-else-if="searched && query.length >= 2"
      class="flex flex-col items-center gap-3 py-16 text-center"
    >
      <UIcon name="i-lucide-user-search" class="size-12 text-muted" />
      <p class="font-semibold">No users found</p>
      <p class="text-muted text-sm">Try a different username.</p>
    </div>
  </UContainer>
</template>
