<script setup lang="ts">
import type { LocationQuery } from '#vue-router'

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
useAuth()
const route = useRoute()
const router = useRouter()

const MIN_YEAR = 1900
const MAX_YEAR = new Date().getFullYear()

// --- State ---
const type = ref('')
const sortBy = ref('popularity.desc')
const yearRange = ref<[number, number]>([MIN_YEAR, MAX_YEAR])
const ratingRange = ref<[number, number]>([0, 10])
const page = ref(1)
const watchProviders = ref('')
const withGenres = ref('')

const MIN_VOTE_COUNT = 100

// Platform definitions — value is the TMDB provider ID string
const platforms = [
  { value: '8', label: 'Netflix' },
  { value: '337', label: 'Disney+' },
  { value: '9', label: 'Amazon Prime' },
  { value: '384', label: 'HBO Max' },
  { value: '2', label: 'Apple TV+' },
  { value: '15', label: 'Hulu' },
  { value: '531', label: 'Paramount+' },
  { value: '283', label: 'Crunchyroll' },
]

// Genre definitions — value is the TMDB genre ID string
const genres = [
  { value: '28', label: 'Action' },
  { value: '12', label: 'Adventure' },
  { value: '16', label: 'Animation' },
  { value: '35', label: 'Comedy' },
  { value: '80', label: 'Crime' },
  { value: '99', label: 'Documentary' },
  { value: '18', label: 'Drama' },
  { value: '14', label: 'Fantasy' },
  { value: '27', label: 'Horror' },
  { value: '10749', label: 'Romance' },
  { value: '878', label: 'Sci-Fi' },
  { value: '53', label: 'Thriller' },
]

// Automatically require a minimum vote count when sorting/filtering by rating
const needsMinVotes = computed(
  () =>
    sortBy.value === 'vote_average.desc' ||
    sortBy.value === 'vote_average.asc' ||
    ratingRange.value[0] > 0
)

// USelectMenu works with full objects; bridge back to ID string on change
const selectedPlatform = computed({
  get: () => platforms.find((p) => watchProviders.value === p.value) ?? undefined,
  set: (item) => {
    watchProviders.value = item?.value ?? ''
  },
})

const selectedGenre = computed({
  get: () => genres.find((g) => withGenres.value === g.value) ?? undefined,
  set: (item) => {
    withGenres.value = item?.value ?? ''
  },
})



// --- Init / sync from URL ---
function syncFromQuery(query: LocationQuery) {
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
  watchProviders.value = query.watch_providers
    ? String(query.watch_providers)
    : ''
  withGenres.value = query.with_genres
    ? String(query.with_genres)
    : ''
}

syncFromQuery(route.query)

let externalNav = false
onBeforeRouteUpdate((to) => {
  externalNav = true
  syncFromQuery(to.query)
  nextTick(() => {
    externalNav = false
  })
})

// --- Reset page when filters change ---
watch(
  [type, sortBy, yearRange, ratingRange, watchProviders, withGenres],
  () => {
    if (externalNav) return
    page.value = 1
  },
  { deep: true }
)

// --- Sync state back to URL ---
function buildQuery() {
  const out: Record<string, string> = {}
  if (type.value) out.type = type.value
  if (sortBy.value !== 'popularity.desc') out.sort_by = sortBy.value
  if (yearRange.value[0] > MIN_YEAR) out.year_from = String(yearRange.value[0])
  if (yearRange.value[1] < MAX_YEAR) out.year_to = String(yearRange.value[1])
  if (ratingRange.value[0] > 0)
    out.vote_average_min = ratingRange.value[0].toFixed(1)
  if (ratingRange.value[1] < 10)
    out.vote_average_max = ratingRange.value[1].toFixed(1)
  if (page.value > 1) out.page = String(page.value)
  if (watchProviders.value) out.watch_providers = watchProviders.value
  if (withGenres.value) out.with_genres = withGenres.value
  return out
}

watch(
  [type, sortBy, yearRange, ratingRange, page, watchProviders, withGenres],
  () => {
    if (externalNav) return
    router.replace({ query: buildQuery() })
  },
  { deep: true }
)

// --- API URL ---
const apiUrl = computed(() => {
  const params = new URLSearchParams()
  if (type.value) params.set('type', type.value)
  params.set('sort_by', sortBy.value)
  if (yearRange.value[0] > MIN_YEAR)
    params.set('year_from', String(yearRange.value[0]))
  if (yearRange.value[1] < MAX_YEAR)
    params.set('year_to', String(yearRange.value[1]))
  if (ratingRange.value[0] > 0)
    params.set('vote_average_min', ratingRange.value[0].toFixed(1))
  if (ratingRange.value[1] < 10)
    params.set('vote_average_max', ratingRange.value[1].toFixed(1))
  if (needsMinVotes.value) params.set('vote_count_min', String(MIN_VOTE_COUNT))
  params.set('page', String(page.value))
  if (watchProviders.value) {
    params.set('watch_providers', watchProviders.value)
    params.set('watch_region', 'US')
  }
  if (withGenres.value)
    params.set('with_genres', withGenres.value)
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
      watchProviders.value = ''
      withGenres.value = ''
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
      watchProviders.value = ''
      withGenres.value = ''
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
      watchProviders.value = ''
      withGenres.value = ''
    },
  },
]

const activePreset = computed(() => {
  if (sortBy.value === 'popularity.desc' && ratingRange.value[0] === 0)
    return 'Popular'
  if (sortBy.value === 'vote_average.desc' && ratingRange.value[0] >= 7)
    return 'Best Rated'
  if (sortBy.value === 'release_date.desc') return 'New Releases'
  return null
})

// Active filter count badge
const activeFilterCount = computed(() => {
  let count = 0
  if (type.value) count++
  if (sortBy.value !== 'popularity.desc') count++
  if (yearRange.value[0] > MIN_YEAR || yearRange.value[1] < MAX_YEAR) count++
  if (ratingRange.value[0] > 0 || ratingRange.value[1] < 10) count++
  count += watchProviders.value.length
  count += withGenres.value.length
  return count
})

function resetFilters() {
  type.value = ''
  sortBy.value = 'popularity.desc'
  yearRange.value = [MIN_YEAR, MAX_YEAR]
  ratingRange.value = [0, 10]
  watchProviders.value = ''
  withGenres.value = ''
}

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

function displayTitle(r: MediaResult) {
  return r.title ?? r.name ?? ''
}
function displayYear(r: MediaResult) {
  return (r.release_date ?? r.first_air_date ?? '').slice(0, 4)
}
function posterUrl(r: MediaResult) {
  return r.poster_path
    ? `https://image.tmdb.org/t/p/w342${r.poster_path}`
    : null
}
function mediaRoute(r: MediaResult) {
  return r.media_type === 'tv' ? `/series/${r.id}` : `/movie/${r.id}`
}
</script>

<template>
  <UContainer class="py-10">
    <!-- Header -->
    <div class="mb-8">
      <h1 class="text-4xl font-bold tracking-tight">Discover</h1>
      <p class="text-muted mt-1.5">
        Browse movies &amp; series with smart filters.
      </p>
    </div>

    <!-- Preset pills -->
    <div class="flex flex-wrap gap-2 mb-6">
      <button
        v-for="preset in presets"
        :key="preset.label"
        class="discover-preset"
        :class="activePreset === preset.label ? 'discover-preset--active' : ''"
        @click="preset.apply()"
      >
        <UIcon :name="preset.icon" class="size-3.5 shrink-0" />
        {{ preset.label }}
      </button>
    </div>

    <!-- ── Filter Panel ── -->
    <div class="discover-panel mb-8">
      <!-- Row 1: Type + Sort + Year + Rating -->
      <div class="discover-row">
        <!-- Type segmented -->
        <div class="discover-field">
          <span class="discover-label">Type</span>
          <div class="discover-segmented">
            <button
              v-for="opt in typeOptions"
              :key="opt.value"
              class="discover-seg-btn"
              :class="type === opt.value ? 'discover-seg-btn--active' : ''"
              @click="type = opt.value"
            >
              {{ opt.label }}
            </button>
          </div>
        </div>

        <!-- Sort -->
        <div class="discover-field discover-field--grow">
          <span class="discover-label">Sort by</span>
          <USelect v-model="sortBy" :items="sortOptions" class="w-full" />
        </div>

        <!-- Year Range -->
        <div class="discover-field">
          <span class="discover-label">Year</span>
          <div class="discover-range-inputs">
            <input
              type="number"
              :value="yearRange[0]"
              :min="MIN_YEAR"
              :max="yearRange[1]"
              class="discover-num-input"
              @change="
                yearRange = [
                  Math.max(
                    MIN_YEAR,
                    Math.min(
                      Number(($event.target as HTMLInputElement).value),
                      yearRange[1]
                    )
                  ),
                  yearRange[1],
                ]
              "
            />
            <span class="discover-range-sep">–</span>
            <input
              type="number"
              :value="yearRange[1]"
              :min="yearRange[0]"
              :max="MAX_YEAR"
              class="discover-num-input"
              @change="
                yearRange = [
                  yearRange[0],
                  Math.min(
                    MAX_YEAR,
                    Math.max(
                      Number(($event.target as HTMLInputElement).value),
                      yearRange[0]
                    )
                  ),
                ]
              "
            />
          </div>
        </div>

        <!-- Rating Range -->
        <div class="discover-field">
          <span class="discover-label">Rating</span>
          <div class="discover-range-inputs">
            <input
              type="number"
              :value="ratingRange[0]"
              :min="0"
              :max="ratingRange[1]"
              :step="0.1"
              class="discover-num-input discover-num-input--sm"
              @change="
                ratingRange = [
                  Math.max(
                    0,
                    Math.min(
                      Number(
                        Number(
                          ($event.target as HTMLInputElement).value
                        ).toFixed(1)
                      ),
                      ratingRange[1]
                    )
                  ),
                  ratingRange[1],
                ]
              "
            />
            <span class="discover-range-sep">–</span>
            <input
              type="number"
              :value="ratingRange[1]"
              :min="ratingRange[0]"
              :max="10"
              :step="0.1"
              class="discover-num-input discover-num-input--sm"
              @change="
                ratingRange = [
                  ratingRange[0],
                  Math.min(
                    10,
                    Math.max(
                      Number(
                        Number(
                          ($event.target as HTMLInputElement).value
                        ).toFixed(1)
                      ),
                      ratingRange[0]
                    )
                  ),
                ]
              "
            />
          </div>
        </div>
      </div>

      <!-- Divider -->
      <div class="discover-divider" />

      <!-- Row 2: Streaming + Genre (multi-select dropdowns) -->
      <div class="discover-row">
        <!-- Streaming -->
        <div class="discover-field discover-field--grow">
          <span class="discover-label">
            <UIcon name="i-lucide-tv" class="size-3.5" />
            Streaming
          </span>
          <USelectMenu
            v-model="selectedPlatform"
            :items="platforms"
            placeholder="All platforms"
            class="w-full"
          />
        </div>

        <!-- Genre -->
        <div class="discover-field discover-field--grow">
          <span class="discover-label">
            <UIcon name="i-lucide-tag" class="size-3.5" />
            Genre
          </span>
          <USelectMenu
            v-model="selectedGenre"
            :items="genres"
            placeholder="All genres"
            class="w-full"
          />
        </div>
      </div>

      <!-- Footer: active count + reset -->
      <div v-if="activeFilterCount > 0" class="discover-panel-footer">
        <span class="discover-active-badge">
          <UIcon name="i-lucide-sliders-horizontal" class="size-3" />
          {{ activeFilterCount }} filter{{ activeFilterCount !== 1 ? 's' : '' }}
          active
        </span>
        <button class="discover-reset-btn" @click="resetFilters">
          <UIcon name="i-lucide-x" class="size-3.5" />
          Reset all
        </button>
      </div>
    </div>

    <!-- Result count -->
    <p v-if="data && status === 'success'" class="text-sm text-muted mb-6">
      <span class="font-semibold text-default">{{
        data.total_results.toLocaleString()
      }}</span>
      results — page {{ data.page }} of {{ totalPages }}
    </p>

    <!-- Loading skeleton -->
    <div
      v-if="status === 'pending'"
      class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4"
    >
      <div v-for="i in 20" :key="i" class="space-y-2">
        <USkeleton class="aspect-[2/3] w-full rounded-xl" />
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
        <div
          class="relative aspect-[2/3] rounded-xl overflow-hidden bg-elevated"
        >
          <img
            v-if="posterUrl(result)"
            :src="posterUrl(result)!"
            :alt="displayTitle(result)"
            class="w-full h-full object-cover transition-transform duration-300 group-hover:scale-105"
          />
          <div v-else class="w-full h-full flex items-center justify-center">
            <UIcon name="i-lucide-film" class="size-10 text-muted" />
          </div>
          <div
            class="absolute inset-0 bg-gradient-to-t from-black/80 via-black/30 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300"
          />
          <div
            class="absolute bottom-0 left-0 right-0 p-3 translate-y-1 opacity-0 group-hover:translate-y-0 group-hover:opacity-100 transition-all duration-300"
          >
            <p class="text-xs text-white/90 line-clamp-4 leading-snug">
              {{ result.overview }}
            </p>
          </div>
          <div class="absolute top-2 right-2">
            <UBadge
              :color="result.media_type === 'tv' ? 'neutral' : 'primary'"
              variant="solid"
              size="xs"
            >
              {{ result.media_type === 'tv' ? 'Series' : 'Movie' }}
            </UBadge>
          </div>
        </div>
        <div class="mt-2 space-y-0.5">
          <p
            class="text-sm font-medium truncate group-hover:text-primary transition-colors"
          >
            {{ displayTitle(result) }}
          </p>
          <div class="flex items-center justify-between">
            <span class="text-xs text-muted">{{ displayYear(result) }}</span>
            <span
              v-if="result.vote_average"
              class="flex items-center gap-1 text-xs text-muted"
            >
              <UIcon
                name="i-lucide-star"
                class="size-3 text-yellow-400 fill-yellow-400"
              />
              {{ result.vote_average.toFixed(1) }}
            </span>
          </div>
        </div>
      </NuxtLink>
    </div>

    <!-- Empty state -->
    <div v-else class="flex flex-col items-center gap-4 py-24 text-center">
      <UIcon name="i-lucide-search-x" class="size-14 text-muted" />
      <div>
        <p class="font-semibold">No results found</p>
        <p class="text-muted text-sm mt-1">Try adjusting your filters.</p>
      </div>
    </div>

    <!-- Pagination -->
    <div
      v-if="totalPages > 1 && status === 'success'"
      class="flex items-center justify-center gap-4 mt-10"
    >
      <UButton
        color="neutral"
        variant="soft"
        leading-icon="i-lucide-chevron-left"
        :disabled="page <= 1"
        @click="page--"
      >
        Previous
      </UButton>
      <span class="text-sm text-muted"
        >Page {{ page }} of {{ totalPages }}</span
      >
      <UButton
        color="neutral"
        variant="soft"
        trailing-icon="i-lucide-chevron-right"
        :disabled="page >= totalPages"
        @click="page++"
      >
        Next
      </UButton>
    </div>
  </UContainer>
</template>

<style scoped>
/* ── Preset pills ── */
.discover-preset {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.875rem;
  border-radius: 9999px;
  font-size: 0.8125rem;
  font-weight: 500;
  border: 1px solid var(--ui-border);
  color: var(--ui-text-muted);
  background: transparent;
  cursor: pointer;
  transition: all 0.18s ease;
}

.discover-preset:hover {
  color: var(--ui-text);
  border-color: var(--ui-primary);
  background: color-mix(in srgb, var(--ui-primary) 8%, transparent);
}

.discover-preset--active {
  background: var(--ui-primary);
  border-color: var(--ui-primary);
  color: white;
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--ui-primary) 20%, transparent);
}

.discover-preset--active:hover {
  background: var(--ui-primary);
  color: white;
}

/* ── Filter Panel ── */
.discover-panel {
  border: 1px solid var(--ui-border);
  border-radius: 1rem;
  padding: 1.25rem;
  background: var(--ui-bg-elevated);
  display: flex;
  flex-direction: column;
  gap: 1.125rem;
}

.discover-row {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  gap: 0.875rem;
}

.discover-field {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.discover-field--grow {
  flex: 1 1 160px;
}

.discover-label {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.6875rem;
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--ui-text-muted);
}

/* ── Segmented control ── */
.discover-segmented {
  display: flex;
  border-radius: 0.625rem;
  border: 1px solid var(--ui-border);
  overflow: hidden;
  background: var(--ui-bg);
}

.discover-seg-btn {
  padding: 0.375rem 0.875rem;
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--ui-text-muted);
  background: transparent;
  border: none;
  cursor: pointer;
  transition: all 0.15s ease;
  white-space: nowrap;
}

.discover-seg-btn:not(:last-child) {
  border-right: 1px solid var(--ui-border);
}

.discover-seg-btn:hover:not(.discover-seg-btn--active) {
  color: var(--ui-text);
  background: var(--ui-bg-muted);
}

.discover-seg-btn--active {
  background: var(--ui-primary);
  color: white;
}

/* ── Range inputs ── */
.discover-range-inputs {
  display: flex;
  align-items: center;
  gap: 0.375rem;
}

.discover-range-sep {
  font-size: 0.8125rem;
  color: var(--ui-text-muted);
}

.discover-num-input {
  width: 5rem;
  border-radius: 0.5rem;
  border: 1px solid var(--ui-border);
  background: var(--ui-bg);
  padding: 0.375rem 0.5rem;
  font-size: 0.8125rem;
  text-align: center;
  color: var(--ui-text);
  transition:
    border-color 0.15s,
    box-shadow 0.15s;
  appearance: textfield;
  -moz-appearance: textfield;
}

.discover-num-input::-webkit-inner-spin-button,
.discover-num-input::-webkit-outer-spin-button {
  -webkit-appearance: none;
}

.discover-num-input:focus {
  outline: none;
  border-color: var(--ui-primary);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--ui-primary) 20%, transparent);
}

.discover-num-input--sm {
  width: 3.75rem;
}

/* ── Dividers ── */
.discover-divider {
  height: 1px;
  background: var(--ui-border);
}

/* ── Panel footer ── */
.discover-panel-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: 0.25rem;
  border-top: 1px solid var(--ui-border);
  margin-top: 0.125rem;
}

.discover-active-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.75rem;
  color: var(--ui-primary);
  font-weight: 500;
}

.discover-reset-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--ui-text-muted);
  background: transparent;
  border: none;
  cursor: pointer;
  padding: 0.25rem 0.5rem;
  border-radius: 0.375rem;
  transition: all 0.15s;
}

.discover-reset-btn:hover {
  color: var(--ui-text);
  background: var(--ui-bg-muted);
}
</style>
