<script setup>
useHead({
  htmlAttrs: {
    lang: 'en',
    class: 'dark'
  },
  meta: [
    { name: 'viewport', content: 'width=device-width, initial-scale=1' }
  ],
  link: [
    { rel: 'icon', href: '/favicon.ico' },
    { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
    { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' },
    { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap' }
  ]
})

const { user, fetchMe, logout } = useAuth()

onMounted(() => {
  fetchMe()
})

const title = 'Rate-It — Track Films You Watch'
const description = 'Track films you watch, log your ratings, and discover the Midnight Palace of cinema.'

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
        <NuxtLink to="/" class="text-xl font-bold tracking-wider text-gray-900 dark:text-purple-50">
          Rate<span class="text-purple-500">-</span>It
        </NuxtLink>
      </template>

      <template #right>
        <UColorModeButton />

        <template v-if="user">
          <div class="flex items-center gap-3">
            <span class="hidden sm:inline-block text-sm font-medium text-purple-200">
              {{ user.username }}
            </span>
            <UButton label="Logout" color="purple" variant="soft" icon="i-lucide-log-out" @click="logout" />
          </div>
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
