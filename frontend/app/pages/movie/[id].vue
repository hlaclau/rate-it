<script setup lang="ts">
interface Genre {
  id: number
  name: string
}

interface Movie {
  id: number
  title: string
  tagline: string
  overview: string
  poster_path: string | null
  backdrop_path: string | null
  vote_average: number
  vote_count: number
  release_date: string
  runtime: number | null
  genres: Genre[]
  status: string
  original_language: string
}

const route = useRoute()
const config = useRuntimeConfig()

const {
  data: movie,
  status,
  error,
} = useFetch<Movie>(
  () => `${config.public.apiBase}/api/media/movie/${route.params.id}`
)

useSeoMeta({
  title: () => (movie.value ? `${movie.value.title} — Rate It` : 'Rate It'),
  description: () => movie.value?.overview ?? '',
  ogImage: () =>
    movie.value?.poster_path
      ? `https://image.tmdb.org/t/p/w500${movie.value.poster_path}`
      : undefined,
})

const posterUrl = computed(() =>
  movie.value?.poster_path
    ? `https://image.tmdb.org/t/p/w500${movie.value.poster_path}`
    : null
)

const backdropUrl = computed(() =>
  movie.value?.backdrop_path
    ? `https://image.tmdb.org/t/p/original${movie.value.backdrop_path}`
    : null
)

const releaseYear = computed(() => movie.value?.release_date?.slice(0, 4))

const runtimeFormatted = computed(() => {
  const mins = movie.value?.runtime
  if (!mins) return null
  const h = Math.floor(mins / 60)
  const m = mins % 60
  return h > 0 ? `${h}h ${m}m` : `${m}m`
})

const rating = computed(() => movie.value?.vote_average?.toFixed(1))

const releaseDate = computed(() => {
  if (!movie.value?.release_date) return null
  return new Date(movie.value.release_date).toLocaleDateString('en-US', {
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
              <USkeleton
                v-for="i in 3"
                :key="i"
                class="h-6 w-20 rounded-full"
              />
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
      <UIcon name="i-lucide-film" class="size-16 text-muted" />
      <div>
        <h2 class="text-2xl font-semibold">Movie not found</h2>
        <p class="text-muted mt-1">
          This movie doesn't exist or could not be loaded.
        </p>
      </div>
      <UButton
        to="/"
        color="neutral"
        variant="soft"
        leading-icon="i-lucide-arrow-left"
      >
        Back to home
      </UButton>
    </UContainer>

    <!-- Movie content -->
    <div v-else-if="movie">
      <!-- Backdrop -->
      <div class="relative h-[420px] overflow-hidden bg-muted">
        <img
          v-if="backdropUrl"
          :src="backdropUrl"
          :alt="movie.title"
          class="w-full h-full object-cover"
        />
        <div
          class="absolute inset-0 bg-gradient-to-t from-background via-background/60 to-transparent"
        />
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
              :alt="movie.title"
              class="w-44 rounded-xl shadow-2xl ring-1 ring-white/10"
            />
            <div
              v-else
              class="w-44 h-64 rounded-xl bg-elevated flex items-center justify-center ring-1 ring-white/10"
            >
              <UIcon name="i-lucide-film" class="size-12 text-muted" />
            </div>
          </div>

          <!-- Info -->
          <div class="flex-1 sm:self-end space-y-4 pb-1">
            <div>
              <h1 class="text-3xl sm:text-4xl font-bold tracking-tight">
                {{ movie.title }}
              </h1>
              <p v-if="movie.tagline" class="text-muted italic mt-1">
                "{{ movie.tagline }}"
              </p>
            </div>

            <!-- Genres -->
            <div v-if="movie.genres?.length" class="flex flex-wrap gap-2">
              <UBadge
                v-for="genre in movie.genres"
                :key="genre.id"
                color="primary"
                variant="soft"
              >
                {{ genre.name }}
              </UBadge>
            </div>

            <!-- Meta row -->
            <div class="flex flex-wrap items-center gap-x-5 gap-y-2 text-sm">
              <span
                v-if="releaseYear"
                class="flex items-center gap-1.5 text-muted"
              >
                <UIcon name="i-lucide-calendar" class="size-4 shrink-0" />
                {{ releaseYear }}
              </span>
              <span
                v-if="runtimeFormatted"
                class="flex items-center gap-1.5 text-muted"
              >
                <UIcon name="i-lucide-clock" class="size-4 shrink-0" />
                {{ runtimeFormatted }}
              </span>
              <span v-if="movie.vote_average" class="flex items-center gap-1.5">
                <UIcon
                  name="i-lucide-star"
                  class="size-4 shrink-0 text-yellow-400 fill-yellow-400"
                />
                <span class="font-semibold">{{ rating }}</span>
                <span class="text-muted">/10</span>
                <span class="text-muted"
                  >({{ movie.vote_count.toLocaleString() }} votes)</span
                >
              </span>
            </div>
          </div>
        </div>

        <!-- Overview -->
        <div v-if="movie.overview" class="mt-12">
          <h2 class="text-lg font-semibold mb-3">Overview</h2>
          <p class="text-muted leading-relaxed max-w-3xl">
            {{ movie.overview }}
          </p>
        </div>

        <!-- List actions -->
        <div class="mt-10 pt-8 border-t border-default">
          <MediaListActions :external-id="String(route.params.id)" media-type="movie" />
        </div>

        <!-- Details grid -->
        <div
          class="mt-10 grid grid-cols-2 sm:grid-cols-4 gap-6 pt-8 border-t border-default"
        >
          <div v-if="movie.status">
            <p class="text-xs uppercase tracking-widest text-muted mb-1">
              Status
            </p>
            <p class="font-medium">{{ movie.status }}</p>
          </div>
          <div v-if="movie.original_language">
            <p class="text-xs uppercase tracking-widest text-muted mb-1">
              Language
            </p>
            <p class="font-medium uppercase">{{ movie.original_language }}</p>
          </div>
          <div v-if="releaseDate">
            <p class="text-xs uppercase tracking-widest text-muted mb-1">
              Release date
            </p>
            <p class="font-medium">{{ releaseDate }}</p>
          </div>
          <div v-if="runtimeFormatted">
            <p class="text-xs uppercase tracking-widest text-muted mb-1">
              Runtime
            </p>
            <p class="font-medium">{{ runtimeFormatted }}</p>
          </div>
        </div>
      </UContainer>
    </div>
  </div>
</template>
