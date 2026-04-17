<script setup lang="ts">
import type { ListEntry } from '~/composables/useList'

definePageMeta({ middleware: 'auth' })

const { list, listLoading, fetchList, remove } = useList()

const activeFilter = ref<'all' | 'watched' | 'plan_to_watch'>('all')

type SortValue = 'added_at_desc' | 'added_at_asc' | 'title_asc' | 'title_desc' | 'rating_desc' | 'rating_asc' | 'release_year_desc' | 'release_year_asc'
const sortValue = ref<SortValue>('added_at_desc')

const sortOptions: { label: string; value: SortValue }[] = [
  { label: 'Date added (newest)', value: 'added_at_desc' },
  { label: 'Date added (oldest)', value: 'added_at_asc' },
  { label: 'Title A → Z', value: 'title_asc' },
  { label: 'Title Z → A', value: 'title_desc' },
  { label: 'Rating (highest)', value: 'rating_desc' },
  { label: 'Rating (lowest)', value: 'rating_asc' },
  { label: 'Year (newest)', value: 'release_year_desc' },
  { label: 'Year (oldest)', value: 'release_year_asc' },
]

const filteredList = computed<ListEntry[]>(() => {
  const items = activeFilter.value === 'all'
    ? list.value
    : list.value.filter(e => e.status === activeFilter.value)

  const lastUnderscore = sortValue.value.lastIndexOf('_')
  const key = sortValue.value.slice(0, lastUnderscore)
  const dir = sortValue.value.slice(lastUnderscore + 1)

  return [...items].sort((a, b) => {
    if (key === 'title') {
      return dir === 'asc' ? a.title.localeCompare(b.title) : b.title.localeCompare(a.title)
    }
    if (key === 'rating') {
      const ra = a.rating ?? (dir === 'asc' ? Infinity : -Infinity)
      const rb = b.rating ?? (dir === 'asc' ? Infinity : -Infinity)
      return dir === 'asc' ? ra - rb : rb - ra
    }
    if (key === 'release_year') {
      const ya = a.release_year ?? (dir === 'asc' ? Infinity : -Infinity)
      const yb = b.release_year ?? (dir === 'asc' ? Infinity : -Infinity)
      return dir === 'asc' ? ya - yb : yb - ya
    }
    // added_at (default)
    const da = new Date(a.added_at).getTime()
    const db = new Date(b.added_at).getTime()
    return dir === 'asc' ? da - db : db - da
  })
})

onMounted(fetchList)

const posterUrl = (path: string | null) =>
  path ? `https://image.tmdb.org/t/p/w300${path}` : null

// Remove confirmation
const pendingRemove = ref<ListEntry | null>(null)
const removing = ref(false)

const confirmRemove = async () => {
  if (!pendingRemove.value) return
  removing.value = true
  try {
    await remove(pendingRemove.value.media_id)
    pendingRemove.value = null
  } finally {
    removing.value = false
  }
}

useSeoMeta({ title: 'My List — Rate It' })
</script>

<template>
  <UContainer class="py-10">
    <div class="flex flex-wrap items-center justify-between gap-4 mb-8">
      <h1 class="text-3xl font-bold">My List</h1>
      <div class="flex flex-wrap items-center gap-2">
        <div class="flex gap-2">
          <UButton
            v-for="f in [
              { label: 'All', value: 'all' },
              { label: 'Watched', value: 'watched' },
              { label: 'Plan to watch', value: 'plan_to_watch' },
            ]"
            :key="f.value"
            :variant="activeFilter === f.value ? 'solid' : 'ghost'"
            color="neutral"
            size="sm"
            @click="activeFilter = (f.value as 'all' | 'watched' | 'plan_to_watch')"
          >
            {{ f.label }}
          </UButton>
        </div>
        <USelect
          v-model="sortValue"
          :items="sortOptions"
          value-key="value"
          label-key="label"
          size="sm"
          class="w-52"
        />
      </div>
    </div>

    <!-- Loading -->
    <div
      v-if="listLoading"
      class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-6"
    >
      <USkeleton v-for="i in 10" :key="i" class="aspect-[2/3] rounded-xl" />
    </div>

    <!-- Empty state -->
    <div
      v-else-if="filteredList.length === 0"
      class="flex flex-col items-center gap-4 py-24 text-center"
    >
      <UIcon name="i-lucide-bookmark" class="size-16 text-muted" />
      <div>
        <p class="font-semibold text-lg">Nothing here yet</p>
        <p class="text-muted text-sm mt-1">
          {{ activeFilter === 'all' ? 'Add movies or series to your list from their detail page.' : 'No entries with this status.' }}
        </p>
      </div>
      <UButton to="/discover" variant="soft" leading-icon="i-lucide-compass">
        Browse media
      </UButton>
    </div>

    <!-- List grid -->
    <div
      v-else
      class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-6"
    >
      <div
        v-for="entry in filteredList"
        :key="entry.media_id"
        class="group relative rounded-xl overflow-hidden bg-elevated ring-1 ring-default hover:ring-primary transition-all"
      >
        <NuxtLink :to="entry.type === 'series' ? `/series/${entry.external_id}` : `/movie/${entry.external_id}`">
          <img
            v-if="posterUrl(entry.poster_path)"
            :src="posterUrl(entry.poster_path)!"
            :alt="entry.title"
            class="w-full aspect-[2/3] object-cover"
          />
          <div
            v-else
            class="w-full aspect-[2/3] bg-muted flex items-center justify-center"
          >
            <UIcon name="i-lucide-film" class="size-10 text-muted" />
          </div>
        </NuxtLink>

        <div class="p-3 space-y-1.5">
          <p class="font-semibold text-sm leading-tight line-clamp-2">{{ entry.title }}</p>
          <div class="flex items-center justify-between gap-2">
            <UBadge
              :color="entry.status === 'watched' ? 'success' : 'info'"
              size="xs"
            >
              {{ entry.status === 'watched' ? 'Watched' : 'Plan to watch' }}
            </UBadge>
            <span
              v-if="entry.rating"
              class="flex items-center gap-0.5 text-xs font-medium shrink-0"
            >
              <UIcon name="i-lucide-star" class="size-3 text-yellow-400 fill-yellow-400" />
              {{ entry.rating }}
            </span>
          </div>
          <p
            v-if="entry.review"
            class="text-xs text-muted line-clamp-2 italic"
          >
            "{{ entry.review }}"
          </p>
        </div>

        <!-- Remove button (visible on hover) -->
        <UButton
          icon="i-lucide-x"
          size="xs"
          color="neutral"
          variant="solid"
          class="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity"
          @click.prevent="pendingRemove = entry"
        />
      </div>
    </div>

    <!-- Remove confirmation modal -->
    <UModal
      :open="!!pendingRemove"
      title="Remove from list"
      :description="pendingRemove ? `Remove &quot;${pendingRemove.title}&quot; from your list?` : ''"
      @update:open="(v) => { if (!v) pendingRemove = null }"
    >
      <template #footer>
        <div class="flex justify-end gap-2">
          <UButton color="neutral" variant="ghost" @click="pendingRemove = null">Cancel</UButton>
          <UButton color="error" :loading="removing" leading-icon="i-lucide-trash-2" @click="confirmRemove">
            Remove
          </UButton>
        </div>
      </template>
    </UModal>

  </UContainer>
</template>
