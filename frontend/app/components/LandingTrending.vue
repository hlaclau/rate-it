<!-- app/components/LandingTrending.vue -->
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
}

interface MediaResponse {
  results: MediaResult[]
}

const config = useRuntimeConfig()

const { data, status } = await useFetch<MediaResponse>(
  `${config.public.apiBase}/api/media/search?sort_by=popularity.desc&page=1`,
  { lazy: true }
)

const movies = computed(() =>
  (data.value?.results ?? []).filter((r) => r.poster_path).slice(0, 20)
)

function displayTitle(r: MediaResult) {
  return r.title ?? r.name ?? ''
}

function displayYear(r: MediaResult) {
  return (r.release_date ?? r.first_air_date ?? '').slice(0, 4)
}

function mediaRoute(r: MediaResult) {
  return r.media_type === 'tv' ? `/series/${r.id}` : `/movie/${r.id}`
}
</script>

<template>
  <section class="py-20 overflow-hidden bg-black">
    <UContainer>
      <!-- Section header -->
      <div class="flex items-end justify-between mb-8">
        <div>
          <p class="text-xs font-semibold tracking-widest uppercase text-purple-400 mb-2">
            What's hot right now
          </p>
          <h2 class="font-display text-5xl sm:text-6xl text-white leading-none">
            TRENDING
          </h2>
        </div>
        <NuxtLink
          to="/discover?sort_by=popularity.desc"
          class="hidden sm:flex items-center gap-1.5 text-sm font-medium text-white/50 hover:text-purple-400 transition-colors"
        >
          View all
          <UIcon name="i-lucide-arrow-right" class="size-4" />
        </NuxtLink>
      </div>
    </UContainer>

    <!-- Scrollable poster strip (no UContainer — bleeds to edges) -->
    <div
      class="flex gap-4 px-4 sm:px-8 overflow-x-auto pb-4 scrollbar-none"
      style="scroll-snap-type: x mandatory; -webkit-overflow-scrolling: touch;"
    >
      <!-- Skeletons while loading -->
      <template v-if="status === 'pending'">
        <div
          v-for="i in 12"
          :key="i"
          class="flex-none w-36 sm:w-44"
          style="scroll-snap-align: start;"
        >
          <USkeleton class="aspect-[2/3] w-full rounded-xl" />
        </div>
      </template>

      <!-- Posters -->
      <NuxtLink
        v-for="(movie, i) in movies"
        :key="movie.id"
        :to="mediaRoute(movie)"
        class="group flex-none w-36 sm:w-44 animate-fade-up"
        :style="`scroll-snap-align: start; animation-delay: ${i * 40}ms;`"
      >
        <div class="relative aspect-[2/3] rounded-xl overflow-hidden poster-glow">
          <img
            :src="`https://image.tmdb.org/t/p/w342${movie.poster_path}`"
            :alt="displayTitle(movie)"
            class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110"
            loading="lazy"
          />
          <!-- Hover overlay -->
          <div
            class="absolute inset-0 bg-gradient-to-t from-black via-black/60 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300 flex flex-col justify-end p-3"
          >
            <p class="text-white text-xs font-semibold line-clamp-2 leading-snug">
              {{ displayTitle(movie) }}
            </p>
            <div class="flex items-center justify-between mt-1.5">
              <span class="text-white/60 text-xs">{{ displayYear(movie) }}</span>
              <span v-if="movie.vote_average" class="flex items-center gap-1 text-xs text-yellow-400 font-medium">
                <UIcon name="i-lucide-star" class="size-3 fill-yellow-400" />
                {{ movie.vote_average.toFixed(1) }}
              </span>
            </div>
          </div>
        </div>
      </NuxtLink>
    </div>

    <!-- Mobile "View all" link -->
    <div class="sm:hidden mt-4 px-4">
      <NuxtLink
        to="/discover?sort_by=popularity.desc"
        class="flex items-center gap-1.5 text-sm font-medium text-white/50 hover:text-purple-400 transition-colors"
      >
        View all trending
        <UIcon name="i-lucide-arrow-right" class="size-4" />
      </NuxtLink>
    </div>

  </section>
</template>
