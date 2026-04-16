<script setup>
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
const description =
  'Track films you watch, log your ratings, and discover the Midnight Palace of cinema.'

useSeoMeta({
  title,
  description,
  ogTitle: title,
  ogDescription: description,
})
</script>

<template>
  <UApp>
    <UHeader>
      <template #left>
        <NuxtLink
          to="/"
          class="text-xl font-bold tracking-wider text-gray-900 dark:text-purple-50"
        >
          Rate<span class="text-purple-500">-</span>It
        </NuxtLink>
        <nav class="hidden sm:flex items-center gap-1 ml-6">
          <UButton to="/discover" color="neutral" variant="ghost" size="sm">
            Discover
          </UButton>
          <UButton to="/discover?sort_by=popularity.desc" color="neutral" variant="ghost" size="sm">
            Popular
          </UButton>
          <UButton to="/discover?sort_by=vote_average.desc&vote_average_min=7.0" color="neutral" variant="ghost" size="sm">
            Best Rated
          </UButton>
          <UButton :to="`/discover?sort_by=release_date.desc&year_from=${new Date().getFullYear() - 1}`" color="neutral" variant="ghost" size="sm">
            New Releases
          </UButton>
        </nav>
      </template>

      <AppSearch />

      <template #right>
        <UColorModeButton />

        <template v-if="!authInitialized">
          <USkeleton class="h-8 w-20 rounded-lg" />
        </template>
        <template v-else-if="user">
          <UDropdownMenu :items="userMenuItems">
            <UButton color="purple" variant="soft" trailing-icon="i-lucide-chevron-down" size="sm">
              <UIcon name="i-lucide-user" class="size-4" />
              <span class="hidden sm:inline ml-1">{{ user.username }}</span>
            </UButton>
          </UDropdownMenu>
        </template>
        <template v-else>
          <!-- Desktop button -->
          <UButton label="Sign In" color="purple" variant="solid" class="hidden sm:flex" to="/login" />
          <!-- Mobile button -->
          <UButton icon="i-lucide-user" color="purple" variant="solid" class="sm:hidden" aria-label="Sign In"
            to="/login" />
        </template>
      </template>
    </UHeader>

    <UMain>
      <NuxtPage />
    </UMain>

    <UFooter>
      <template #left>
        <p class="text-sm text-purple-300">
          &copy; {{ new Date().getFullYear() }} Rate-It
        </p>
      </template>
    </UFooter>
  </UApp>
</template>
