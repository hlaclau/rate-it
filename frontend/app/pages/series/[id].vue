<script setup lang="ts">
interface Genre {
  id: number
  name: string
}

interface Series {
  id: number
  name: string
  tagline: string
  overview: string
  poster_path: string | null
  backdrop_path: string | null
  vote_average: number
  vote_count: number
  first_air_date: string
  last_air_date: string
  episode_run_time: number[]
  number_of_seasons: number
  number_of_episodes: number
  genres: Genre[]
  status: string
  original_language: string
}

const route = useRoute()
const config = useRuntimeConfig()

const {
  data: series,
  status,
  error,
} = useFetch<Series>(
  () => `${config.public.apiBase}/api/media/series/${route.params.id}`
)

useSeoMeta({
  title: () => (series.value ? `${series.value.name} — Rate It` : 'Rate It'),
  description: () => series.value?.overview ?? '',
  ogImage: () =>
    series.value?.poster_path
      ? `https://image.tmdb.org/t/p/w500${series.value.poster_path}`
      : undefined,
})

const posterUrl = computed(() =>
  series.value?.poster_path
    ? `https://image.tmdb.org/t/p/w500${series.value.poster_path}`
    : null
)

const backdropUrl = computed(() =>
  series.value?.backdrop_path
    ? `https://image.tmdb.org/t/p/original${series.value.backdrop_path}`
    : null
)

const airYear = computed(() => series.value?.first_air_date?.slice(0, 4))

const episodeRuntime = computed(() => {
  const times = series.value?.episode_run_time
  if (!times?.length) return null
  const mins = times[0]
  const h = Math.floor(mins / 60)
  const m = mins % 60
  return h > 0 ? `${h}h ${m}m` : `${m}m`
})

const rating = computed(() => series.value?.vote_average?.toFixed(1))

const firstAirDate = computed(() => {
  if (!series.value?.first_air_date) return null
  return new Date(series.value.first_air_date).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
})

</script>

<template>
  <div>
    <!-- Loading skeleton -->
    <div v-if="status === 'pending'">
      <USkeleton class="h-[420px] w-full rounded-none" />
      <UContainer class="py-8">
        <div class="flex flex-col sm:flex-row gap-8">
          <USkeleton class="w-48 h-72 rounded-xl shrink-0" />
          <div class="flex-1 space-y-4 pt-2">
            <USkeleton class="h-10 w-3/4" />
            <USkeleton class="h-5 w-1/2" />
            <div class="flex gap-2">
              <USkeleton v-for="i in 3" :key="i" class="h-6 w-20 rounded-full" />
            </div>
            <USkeleton class="h-5 w-2/3" />
          </div>
        </div>
        <USkeleton class="mt-10 h-24 w-full max-w-3xl" />
      </UContainer>
    </div>

    <!-- Error state -->
    <UContainer
      v-else-if="error"
      class="py-24 flex flex-col items-center gap-6 text-center"
    >
      <UIcon name="i-lucide-tv" class="size-16 text-muted" />
      <div>
        <h2 class="text-2xl font-semibold">Series not found</h2>
        <p class="text-muted mt-1">This series doesn't exist or could not be loaded.</p>
      </div>
      <UButton to="/" color="neutral" variant="soft" leading-icon="i-lucide-arrow-left">
        Back to home
      </UButton>
    </UContainer>

    <!-- Series content -->
    <div v-else-if="series">
      <!-- Backdrop -->
      <div class="relative h-[420px] overflow-hidden bg-muted">
        <img
          v-if="backdropUrl"
          :src="backdropUrl"
          :alt="series.name"
          class="w-full h-full object-cover"
        />
        <div class="absolute inset-0 bg-gradient-to-t from-background via-background/60 to-transparent" />
        <UButton
          to="/"
          color="neutral"
          variant="soft"
          leading-icon="i-lucide-arrow-left"
          class="absolute top-4 left-4"
        >
          Back
        </UButton>
      </div>

      <!-- Main content -->
      <UContainer class="-mt-28 relative pb-20">
        <div class="flex flex-col sm:flex-row gap-8">
          <!-- Poster -->
          <div class="shrink-0">
            <img
              v-if="posterUrl"
              :src="posterUrl"
              :alt="series.name"
              class="w-44 rounded-xl shadow-2xl ring-1 ring-white/10"
            />
            <div
              v-else
              class="w-44 h-64 rounded-xl bg-elevated flex items-center justify-center ring-1 ring-white/10"
            >
              <UIcon name="i-lucide-tv" class="size-12 text-muted" />
            </div>
          </div>

          <!-- Info -->
          <div class="flex-1 sm:self-end space-y-4 pb-1">
            <div>
              <h1 class="text-3xl sm:text-4xl font-bold tracking-tight">
                {{ series.name }}
              </h1>
              <p v-if="series.tagline" class="text-muted italic mt-1">
                "{{ series.tagline }}"
              </p>
            </div>

            <!-- Genres -->
            <div v-if="series.genres?.length" class="flex flex-wrap gap-2">
              <UBadge
                v-for="genre in series.genres"
                :key="genre.id"
                color="primary"
                variant="soft"
              >
                {{ genre.name }}
              </UBadge>
            </div>

            <!-- Meta row -->
            <div class="flex flex-wrap items-center gap-x-5 gap-y-2 text-sm">
              <span v-if="airYear" class="flex items-center gap-1.5 text-muted">
                <UIcon name="i-lucide-calendar" class="size-4 shrink-0" />
                {{ airYear }}
              </span>
              <span v-if="episodeRuntime" class="flex items-center gap-1.5 text-muted">
                <UIcon name="i-lucide-clock" class="size-4 shrink-0" />
                {{ episodeRuntime }} / ep
              </span>
              <span v-if="series.number_of_seasons" class="flex items-center gap-1.5 text-muted">
                <UIcon name="i-lucide-layers" class="size-4 shrink-0" />
                {{ series.number_of_seasons }} season{{ series.number_of_seasons > 1 ? 's' : '' }}
              </span>
              <span v-if="series.vote_average" class="flex items-center gap-1.5">
                <UIcon name="i-lucide-star" class="size-4 shrink-0 text-yellow-400 fill-yellow-400" />
                <span class="font-semibold">{{ rating }}</span>
                <span class="text-muted">/10</span>
                <span class="text-muted">({{ series.vote_count.toLocaleString() }} votes)</span>
              </span>
            </div>
          </div>
        </div>

        <!-- Overview -->
        <div v-if="series.overview" class="mt-12">
          <h2 class="text-lg font-semibold mb-3">Overview</h2>
          <p class="text-muted leading-relaxed max-w-3xl">{{ series.overview }}</p>
        </div>

        <!-- List actions -->
        <div class="mt-10 pt-8 border-t border-default">
          <MediaListActions :external-id="String(route.params.id)" media-type="series" />
        </div>

        <!-- Details grid -->
        <div class="mt-10 grid grid-cols-2 sm:grid-cols-4 gap-6 pt-8 border-t border-default">
          <div v-if="series.status">
            <p class="text-xs uppercase tracking-widest text-muted mb-1">Status</p>
            <p class="font-medium">{{ series.status }}</p>
          </div>
          <div v-if="series.original_language">
            <p class="text-xs uppercase tracking-widest text-muted mb-1">Language</p>
            <p class="font-medium uppercase">{{ series.original_language }}</p>
          </div>
          <div v-if="firstAirDate">
            <p class="text-xs uppercase tracking-widest text-muted mb-1">First aired</p>
            <p class="font-medium">{{ firstAirDate }}</p>
          </div>
          <div v-if="series.number_of_episodes">
            <p class="text-xs uppercase tracking-widest text-muted mb-1">Episodes</p>
            <p class="font-medium">{{ series.number_of_episodes }}</p>
          </div>
        </div>
      </UContainer>
    </div>
  </div>
</template>
