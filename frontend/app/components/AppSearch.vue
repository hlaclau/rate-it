<script setup lang="ts">
interface MovieSearchResult {
  id: number
  title: string
  poster_path: string | null
  release_date: string
  overview: string
  vote_average: number
}

interface SearchResponse {
  results: MovieSearchResult[]
}

const config = useRuntimeConfig()
const router = useRouter()

const query = ref('')
const results = ref<MovieSearchResult[]>([])
const isOpen = ref(false)
const isLoading = ref(false)

let debounceTimer: ReturnType<typeof setTimeout> | null = null

watch(query, (q) => {
  if (debounceTimer) clearTimeout(debounceTimer)
  if (!q.trim()) {
    results.value = []
    isOpen.value = false
    return
  }
  debounceTimer = setTimeout(async () => {
    isLoading.value = true
    try {
      const data = await $fetch<SearchResponse>(
        `${config.public.apiBase}/media/search?q=${encodeURIComponent(q)}`
      )
      results.value = data.results
      isOpen.value = data.results.length > 0
    } catch {
      results.value = []
      isOpen.value = false
    } finally {
      isLoading.value = false
    }
  }, 300)
})

function selectMovie(id: number) {
  isOpen.value = false
  query.value = ''
  results.value = []
  router.push(`/movie/${id}`)
}

function close() {
  isOpen.value = false
}

router.afterEach(() => {
  isOpen.value = false
  query.value = ''
  results.value = []
})
</script>

<template>
  <div class="relative w-full max-w-sm hidden sm:block">
    <UInput
      v-model="query"
      icon="i-lucide-search"
      placeholder="Search movies..."
      :loading="isLoading"
      @focus="isOpen = results.length > 0"
    />

    <div
      v-if="isOpen"
      class="absolute top-full mt-1 w-full z-50 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg overflow-hidden"
    >
      <ul>
        <li
          v-for="movie in results.slice(0, 8)"
          :key="movie.id"
          class="flex items-center gap-3 px-3 py-2 hover:bg-gray-100 dark:hover:bg-gray-800 cursor-pointer"
          @click="selectMovie(movie.id)"
          @mousedown.prevent
        >
          <img
            v-if="movie.poster_path"
            :src="`https://image.tmdb.org/t/p/w92${movie.poster_path}`"
            :alt="movie.title"
            class="w-8 h-12 object-cover rounded flex-shrink-0"
          />
          <div
            v-else
            class="w-8 h-12 bg-gray-200 dark:bg-gray-700 rounded flex items-center justify-center flex-shrink-0"
          >
            <UIcon name="i-lucide-film" class="text-gray-400 text-xs" />
          </div>
          <div class="min-w-0">
            <p
              class="text-sm font-medium text-gray-900 dark:text-white truncate"
            >
              {{ movie.title }}
            </p>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              {{ movie.release_date?.slice(0, 4) }}
            </p>
          </div>
        </li>
      </ul>
    </div>

    <div v-if="isOpen" class="fixed inset-0 z-40" @click="close" />
  </div>
</template>
