<script setup lang="ts">
interface MediaSearchResult {
  id: number
  media_type: 'movie' | 'tv' | 'person'
  title?: string
  name?: string
  poster_path: string | null
  release_date?: string
  first_air_date?: string
  overview: string
  vote_average: number
}

interface SearchResponse {
  results: MediaSearchResult[]
}

const config = useRuntimeConfig()
const router = useRouter()

const query = ref('')
const results = ref<MediaSearchResult[]>([])
const isOpen = ref(false)
const isLoading = ref(false)
const inputRef = ref<HTMLInputElement | null>(null)

let debounceTimer: ReturnType<typeof setTimeout> | null = null

const mediaResults = computed(() =>
  results.value
    .filter((r) => r.media_type === 'movie' || r.media_type === 'tv')
    .slice(0, 6)
)

function displayTitle(r: MediaSearchResult) {
  return r.title ?? r.name ?? ''
}
function displayYear(r: MediaSearchResult) {
  return (r.release_date ?? r.first_air_date ?? '').slice(0, 4)
}
function mediaRoute(r: MediaSearchResult) {
  return r.media_type === 'tv' ? `/series/${r.id}` : `/movie/${r.id}`
}

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
        `${config.public.apiBase}/api/media/search?q=${encodeURIComponent(q)}`
      )
      results.value = data.results
      isOpen.value = mediaResults.value.length > 0
    } catch {
      results.value = []
      isOpen.value = false
    } finally {
      isLoading.value = false
    }
  }, 300)
})

function selectMedia(result: MediaSearchResult) {
  isOpen.value = false
  query.value = ''
  results.value = []
  router.push(mediaRoute(result))
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
  <div class="relative hidden sm:block">
    <UInput
      ref="inputRef"
      v-model="query"
      icon="i-lucide-search"
      placeholder="Search…"
      :loading="isLoading"
      size="sm"
      class="w-40 focus-within:w-56 transition-[width] duration-200"
      @focus="isOpen = mediaResults.length > 0"
    />

    <Transition
      enter-active-class="transition duration-150 ease-out"
      enter-from-class="opacity-0 translate-y-1"
      enter-to-class="opacity-100 translate-y-0"
      leave-active-class="transition duration-100 ease-in"
      leave-from-class="opacity-100 translate-y-0"
      leave-to-class="opacity-0 translate-y-1"
    >
      <div
        v-if="isOpen"
        class="absolute top-full mt-2 w-80 z-50 rounded-xl border border-default bg-elevated shadow-xl overflow-hidden"
      >
        <ul class="py-1">
          <li
            v-for="result in mediaResults"
            :key="result.id"
            class="group flex items-center gap-3 px-3 py-2.5 hover:bg-muted cursor-pointer transition-colors"
            @click="selectMedia(result)"
            @mousedown.prevent
          >
            <!-- Poster -->
            <div
              class="shrink-0 w-9 h-[54px] rounded-md overflow-hidden bg-muted"
            >
              <img
                v-if="result.poster_path"
                :src="`https://image.tmdb.org/t/p/w92${result.poster_path}`"
                :alt="displayTitle(result)"
                class="w-full h-full object-cover"
              />
              <div
                v-else
                class="w-full h-full flex items-center justify-center"
              >
                <UIcon name="i-lucide-film" class="size-4 text-muted" />
              </div>
            </div>

            <!-- Info -->
            <div class="min-w-0 flex-1">
              <p
                class="text-sm font-medium truncate group-hover:text-primary transition-colors"
              >
                {{ displayTitle(result) }}
              </p>
              <div class="flex items-center gap-1.5 mt-0.5">
                <UBadge
                  :color="result.media_type === 'tv' ? 'neutral' : 'primary'"
                  variant="subtle"
                  size="xs"
                >
                  {{ result.media_type === 'tv' ? 'Series' : 'Movie' }}
                </UBadge>
                <span v-if="displayYear(result)" class="text-xs text-muted">{{
                  displayYear(result)
                }}</span>
                <span
                  v-if="result.vote_average"
                  class="flex items-center gap-0.5 text-xs text-muted ml-auto"
                >
                  <UIcon
                    name="i-lucide-star"
                    class="size-3 text-yellow-400 fill-yellow-400"
                  />
                  {{ result.vote_average.toFixed(1) }}
                </span>
              </div>
            </div>
          </li>
        </ul>
      </div>
    </Transition>

    <div v-if="isOpen" class="fixed inset-0 z-40" @click="close" />
  </div>
</template>
