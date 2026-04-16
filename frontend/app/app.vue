<script setup lang="ts">
useHead({
  htmlAttrs: {
    lang: 'en',
  },
  meta: [{ name: 'viewport', content: 'width=device-width, initial-scale=1' }],
  link: [
    { rel: 'icon', href: '/favicon.ico' },
    { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
    { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' },
    {
      rel: 'stylesheet',
      href: 'https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap',
    },
  ],
})

const { user, fetchMe, logout, authInitialized } = useAuth()
const route = useRoute()

const discoverItems = [
  [
    {
      label: 'Popular',
      icon: 'i-lucide-flame',
      to: '/discover?sort_by=popularity.desc',
    },
    {
      label: 'Best Rated',
      icon: 'i-lucide-star',
      to: '/discover?sort_by=vote_average.desc&vote_average_min=7.0',
    },
    {
      label: 'New Releases',
      icon: 'i-lucide-sparkles',
      to: `/discover?sort_by=release_date.desc&year_from=${new Date().getFullYear() - 1}`,
    },
  ],
  [
    {
      label: 'All media',
      icon: 'i-lucide-layout-grid',
      to: '/discover',
    },
  ],
]

const isDiscoverActive = computed(() => route.path === '/discover')

const userMenuItems = computed(() => [
  [
    { label: 'My List', icon: 'i-lucide-bookmark', to: '/list' },
    { label: 'Settings', icon: 'i-lucide-settings', to: '/settings' },
  ],
  [
    { label: 'Logout', icon: 'i-lucide-log-out', onSelect: logout },
  ],
])

onMounted(() => {
  fetchMe()
})

const title = 'Rate-It — Track Films You Watch'
const description = 'Track films you watch, log your ratings, and discover great cinema.'

useSeoMeta({
  title,
  description,
  ogTitle: title,
  ogDescription: description,
})
</script>

<template>
  <UApp>
    <header class="sticky top-0 z-50 w-full border-b border-default bg-background/80 backdrop-blur-md">
      <UContainer>
        <div class="flex items-center h-14 gap-6">

          <!-- Logo -->
          <NuxtLink to="/" class="shrink-0 text-lg font-bold tracking-tight">
            Rate<span class="text-primary">-It</span>
          </NuxtLink>

          <!-- Nav -->
          <nav class="hidden md:flex items-center gap-1">
            <UDropdownMenu :items="discoverItems">
              <button
                class="flex items-center gap-1 px-3 py-1.5 text-sm font-medium rounded-md transition-colors"
                :class="isDiscoverActive
                  ? 'text-primary bg-primary/10'
                  : 'text-muted hover:text-default hover:bg-elevated'"
              >
                Discover
                <UIcon name="i-lucide-chevron-down" class="size-3.5 opacity-60" />
              </button>
            </UDropdownMenu>
          </nav>

          <!-- Spacer -->
          <div class="flex-1" />

          <!-- Search -->
          <AppSearch />

          <!-- Controls -->
          <div class="flex items-center gap-2 shrink-0">
            <UColorModeButton />

            <template v-if="!authInitialized">
              <USkeleton class="h-8 w-24 rounded-lg" />
            </template>
            <template v-else-if="user">
              <UDropdownMenu :items="userMenuItems">
                <UButton color="primary" variant="soft" trailing-icon="i-lucide-chevron-down" size="sm">
                  <UIcon name="i-lucide-user" class="size-4" />
                  <span class="ml-1.5 hidden sm:inline">{{ user.username }}</span>
                </UButton>
              </UDropdownMenu>
            </template>
            <template v-else>
              <UButton to="/login" color="primary" variant="solid" size="sm">Sign in</UButton>
            </template>
          </div>

        </div>
      </UContainer>
    </header>

    <main>
      <NuxtPage />
    </main>

    <footer class="border-t border-default mt-auto py-6">
      <UContainer>
        <p class="text-sm text-muted">
          &copy; {{ new Date().getFullYear() }} Rate-It
        </p>
      </UContainer>
    </footer>
  </UApp>
</template>
