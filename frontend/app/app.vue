<script setup>
useHead({
  htmlAttrs: {
    lang: 'en',
    class: 'dark',
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

        <!-- Desktop button -->
        <UButton
          label="Sign In / Create Account"
          color="primary"
          variant="solid"
          class="hidden sm:flex"
        />
        <!-- Mobile button -->
        <UButton
          icon="i-lucide-user"
          color="primary"
          variant="solid"
          class="sm:hidden"
          aria-label="Sign In / Create Account"
        />
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
