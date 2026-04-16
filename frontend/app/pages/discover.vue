<script setup lang="ts">
interface MediaResult {
  id: number
  media_type: 'movie' | 'tv'
  title?: string
  name?: string
  poster_path: string | null
  release_date?: string
  first_air_date?: string
  vote_average: number
  overview: string
}

interface MediaResponse {
  page: number
  results: MediaResult[]
  total_pages: number
  total_results: number
}

const config = useRuntimeConfig()
const route = useRoute()
const router = useRouter()

const MIN_YEAR = 1900
const MAX_YEAR = new Date().getFullYear()

// --- State ---
const type = ref('')
const sortBy = ref('popularity.desc')
const yearRange = ref([MIN_YEAR, MAX_YEAR])
const ratingRange = ref([0, 10])
const page = ref(1)

const MIN_VOTE_COUNT = 100

// Automatically require a minimum vote count when sorting/filtering by rating
const needsMinVotes = computed(() =>
  sortBy.value === 'vote_average.desc' ||
  sortBy.value === 'vote_average.asc' ||
  ratingRange.value[0] > 0
)

// --- Init / sync from URL ---
function syncFromQuery(query: Record<string, string | string[]>) {
  type.value = String(query.type ?? '')
  sortBy.value = String(query.sort_by ?? 'popularity.desc')
  yearRange.value = [
    Number(query.year_from ?? MIN_YEAR),
    Number(query.year_to ?? MAX_YEAR),
  ]
  ratingRange.value = [
    Number(query.vote_average_min ?? 0),
    Number(query.vote_average_max ?? 10),
  ]
  page.value = Number(query.page ?? 1)
}

syncFromQuery(route.query)

// Handle preset nav link clicks (same route, different query)
let externalNav = false
onBeforeRouteUpdate((to) => {
  externalNav = true
  syncFromQuery(to.query)
  nextTick(() => { externalNav = false })
})

// --- Reset page when filters change ---
watch([type, sortBy, yearRange, ratingRange], () => {
  if (externalNav) return
  page.value = 1
}, { deep: true })

// --- Sync state back to URL ---
function buildQuery() {
  const out: Record<string, string> = {}
  if (type.value) out.type = type.value
  if (sortBy.value !== 'popularity.desc') out.sort_by = sortBy.value
  if (yearRange.value[0] > MIN_YEAR) out.year_from = String(yearRange.value[0])
  if (yearRange.value[1] < MAX_YEAR) out.year_to = String(yearRange.value[1])
  if (ratingRange.value[0] > 0) out.vote_average_min = ratingRange.value[0].toFixed(1)
  if (ratingRange.value[1] < 10) out.vote_average_max = ratingRange.value[1].toFixed(1)
  if (page.value > 1) out.page = String(page.value)
  return out
}

watch([type, sortBy, yearRange, ratingRange, page], () => {
  if (externalNav) return
  router.replace({ query: buildQuery() })
}, { deep: true })

// --- API URL ---
const apiUrl = computed(() => {
  const params = new URLSearchParams()
  if (type.value) params.set('type', type.value)
  params.set('sort_by', sortBy.value)
  if (yearRange.value[0] > MIN_YEAR) params.set('year_from', String(yearRange.value[0]))
  if (yearRange.value[1] < MAX_YEAR) params.set('year_to', String(yearRange.value[1]))
  if (ratingRange.value[0] > 0) params.set('vote_average_min', ratingRange.value[0].toFixed(1))
  if (ratingRange.value[1] < 10) params.set('vote_average_max', ratingRange.value[1].toFixed(1))
  if (needsMinVotes.value) params.set('vote_count_min', String(MIN_VOTE_COUNT))
  params.set('page', String(page.value))
  return `${config.public.apiBase}/api/media/search?${params.toString()}`
})

const { data, status } = useFetch<MediaResponse>(() => apiUrl.value)

useSeoMeta({
  title: 'Discover — Rate It',
  description: 'Browse movies & series with filters.',
})

// --- Presets ---
const presets = [
  {
    label: 'Popular',
    icon: 'i-lucide-flame',
    apply() {
      type.value = ''
      sortBy.value = 'popularity.desc'
      yearRange.value = [MIN_YEAR, MAX_YEAR]
      ratingRange.value = [0, 10]
    },
  },
  {
    label: 'Best Rated',
    icon: 'i-lucide-star',
    apply() {
      type.value = ''
      sortBy.value = 'vote_average.desc'
      yearRange.value = [MIN_YEAR, MAX_YEAR]
      ratingRange.value = [7.0, 10]
    },
  },
  {
    label: 'New Releases',
    icon: 'i-lucide-sparkles',
    apply() {
      type.value = ''
      sortBy.value = 'release_date.desc'
      yearRange.value = [MAX_YEAR - 1, MAX_YEAR]
      ratingRange.value = [0, 10]
    },
  },
]

const activePreset = computed(() => {
  if (sortBy.value === 'popularity.desc' && ratingRange.value[0] === 0) return 'Popular'
  if (sortBy.value === 'vote_average.desc' && ratingRange.value[0] >= 7) return 'Best Rated'
  if (sortBy.value === 'release_date.desc') return 'New Releases'
  return null
})

// --- Helpers ---
const totalPages = computed(() => Math.min(data.value?.total_pages ?? 1, 500))

const sortOptions = [
  { label: 'Most Popular', value: 'popularity.desc' },
  { label: 'Least Popular', value: 'popularity.asc' },
  { label: 'Highest Rated', value: 'vote_average.desc' },
  { label: 'Lowest Rated', value: 'vote_average.asc' },
  { label: 'Newest First', value: 'release_date.desc' },
  { label: 'Oldest First', value: 'release_date.asc' },
]

const typeOptions = [
  { label: 'All', value: '' },
  { label: 'Movies', value: 'movie' },
  { label: 'Series', value: 'series' },
]

function displayTitle(r: MediaResult) { return r.title ?? r.name ?? '' }
function displayYear(r: MediaResult) { return (r.release_date ?? r.first_air_date ?? '').slice(0, 4) }
function posterUrl(r: MediaResult) { return r.poster_path ? `https://image.tmdb.org/t/p/w342${r.poster_path}` : null }
function mediaRoute(r: MediaResult) { return r.media_type === 'tv' ? `/series/${r.id}` : `/movie/${r.id}` }
</script>

<template>
  <UContainer class="py-10">
    <!-- Header + presets -->
    <div class="mb-6">
      <h1 class="text-3xl font-bold tracking-tight">Discover</h1>
      <p class="text-muted mt-1">
        Browse movies & series with filters.
      </p>
    </div>

    <!-- Preset pills -->
    <div class="flex flex-wrap gap-2 mb-6">
      <button
        v-for="preset in presets"
        :key="preset.label"
        class="flex items-center gap-1.5 px-3 py-1.5 rounded-full text-sm font-medium border transition-colors"
        :class="activePreset === preset.label
          ? 'bg-primary-500 border-primary-500 text-white'
          : 'border-default text-muted hover:text-default hover:border-primary-400'"
        @click="preset.apply()"
      >
        <UIcon :name="preset.icon" class="size-3.5" />
        {{ preset.label }}
      </button>
    </div>

    <!-- Filters -->
    <div class="flex flex-wrap items-end gap-3 mb-8 p-4 rounded-xl border border-default bg-elevated">
      <!-- Type toggle -->
      <div class="flex rounded-lg overflow-hidden border border-default shrink-0">
        <button
          v-for="opt in typeOptions"
          :key="opt.value"
          class="px-3 py-1.5 text-sm font-medium transition-colors"
          :class="type === opt.value
            ? 'bg-primary-500 text-white'
            : 'text-muted hover:text-default hover:bg-muted'"
          @click="type = opt.value"
        >
          {{ opt.label }}
        </button>
      </div>

      <!-- Sort -->
      <USelect v-model="sortBy" :items="sortOptions" class="w-44" />

      <!-- Year range -->
      <div class="flex flex-col gap-1 shrink-0">
        <span class="text-xs text-muted font-medium">Year</span>
        <div class="flex items-center gap-1.5">
          <input
            type="number"
            :value="yearRange[0]"
            :min="MIN_YEAR"
            :max="yearRange[1]"
            class="w-20 rounded-md border border-default bg-default px-2 py-1.5 text-sm tabular-nums text-center focus:outline-none focus:ring-2 focus:ring-primary-500"
            @change="yearRange = [Math.max(MIN_YEAR, Math.min(Number(($event.target as HTMLInputElement).value), yearRange[1])), yearRange[1]]"
          />
          <span class="text-muted text-xs">—</span>
          <input
            type="number"
            :value="yearRange[1]"
            :min="yearRange[0]"
            :max="MAX_YEAR"
            class="w-20 rounded-md border border-default bg-default px-2 py-1.5 text-sm tabular-nums text-center focus:outline-none focus:ring-2 focus:ring-primary-500"
            @change="yearRange = [yearRange[0], Math.min(MAX_YEAR, Math.max(Number(($event.target as HTMLInputElement).value), yearRange[0]))]"
          />
        </div>
      </div>

      <!-- Rating range -->
      <div class="flex flex-col gap-1 shrink-0">
        <span class="text-xs text-muted font-medium">Rating</span>
        <div class="flex items-center gap-1.5">
          <input
            type="number"
            :value="ratingRange[0]"
            :min="0"
            :max="ratingRange[1]"
            :step="0.1"
            class="w-16 rounded-md border border-default bg-default px-2 py-1.5 text-sm tabular-nums text-center focus:outline-none focus:ring-2 focus:ring-primary-500"
            @change="ratingRange = [Math.max(0, Math.min(Number(Number(($event.target as HTMLInputElement).value).toFixed(1)), ratingRange[1])), ratingRange[1]]"
          />
          <span class="text-muted text-xs">—</span>
          <input
            type="number"
            :value="ratingRange[1]"
            :min="ratingRange[0]"
            :max="10"
            :step="0.1"
            class="w-16 rounded-md border border-default bg-default px-2 py-1.5 text-sm tabular-nums text-center focus:outline-none focus:ring-2 focus:ring-primary-500"
            @change="ratingRange = [ratingRange[0], Math.min(10, Math.max(Number(Number(($event.target as HTMLInputElement).value).toFixed(1)), ratingRange[0]))]"
          />
        </div>
      </div>
    </div>

    <!-- Result count -->
    <p v-if="data && status === 'success'" class="text-sm text-muted mb-6">
      {{ data.total_results.toLocaleString() }} results — page {{ data.page }} of {{ totalPages }}
    </p>

    <!-- Loading skeleton -->
    <div v-if="status === 'pending'" class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
      <div v-for="i in 20" :key="i" class="space-y-2">
        <USkeleton class="aspect-[2/3] w-full rounded-lg" />
        <USkeleton class="h-4 w-3/4" />
        <USkeleton class="h-3 w-1/2" />
      </div>
    </div>

    <!-- Results grid -->
    <div
      v-else-if="data?.results?.length"
      class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4"
    >
      <NuxtLink
        v-for="result in data.results"
        :key="result.id"
        :to="mediaRoute(result)"
        class="group"
      >
        <div class="relative aspect-[2/3] rounded-lg overflow-hidden bg-elevated">
          <img
            v-if="posterUrl(result)"
            :src="posterUrl(result)!"
            :alt="displayTitle(result)"
            class="w-full h-full object-cover transition-transform duration-300 group-hover:scale-105"
          />
          <div v-else class="w-full h-full flex items-center justify-center">
            <UIcon name="i-lucide-film" class="size-10 text-muted" />
          </div>
          <div class="absolute inset-0 bg-gradient-to-t from-black/80 via-black/30 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
          <div class="absolute bottom-0 left-0 right-0 p-3 translate-y-1 opacity-0 group-hover:translate-y-0 group-hover:opacity-100 transition-all duration-300">
            <p class="text-xs text-white/90 line-clamp-4 leading-snug">{{ result.overview }}</p>
          </div>
          <div class="absolute top-2 right-2">
            <UBadge :color="result.media_type === 'tv' ? 'neutral' : 'primary'" variant="solid" size="xs">
              {{ result.media_type === 'tv' ? 'Series' : 'Movie' }}
            </UBadge>
          </div>
        </div>
        <div class="mt-2 space-y-0.5">
          <p class="text-sm font-medium truncate group-hover:text-primary transition-colors">
            {{ displayTitle(result) }}
          </p>
          <div class="flex items-center justify-between">
            <span class="text-xs text-muted">{{ displayYear(result) }}</span>
            <span v-if="result.vote_average" class="flex items-center gap-1 text-xs text-muted">
              <UIcon name="i-lucide-star" class="size-3 text-yellow-400 fill-yellow-400" />
              {{ result.vote_average.toFixed(1) }}
            </span>
          </div>
        </div>
      </NuxtLink>
    </div>

    <!-- Empty state -->
    <div v-else-if="status !== 'pending'" class="flex flex-col items-center gap-4 py-24 text-center">
      <UIcon name="i-lucide-search-x" class="size-14 text-muted" />
      <div>
        <p class="font-semibold">No results found</p>
        <p class="text-muted text-sm mt-1">Try adjusting your filters.</p>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1 && status === 'success'" class="flex items-center justify-center gap-4 mt-10">
      <UButton color="neutral" variant="soft" leading-icon="i-lucide-chevron-left" :disabled="page <= 1" @click="page--">
        Previous
      </UButton>
      <span class="text-sm text-muted">Page {{ page }} of {{ totalPages }}</span>
      <UButton color="neutral" variant="soft" trailing-icon="i-lucide-chevron-right" :disabled="page >= totalPages" @click="page++">
        Next
      </UButton>
    </div>
  </UContainer>
</template>
