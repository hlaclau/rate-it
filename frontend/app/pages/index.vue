<script setup lang="ts">
const { user } = useAuth()

const primaryLinks = computed(() => {
  if (user.value) {
    return [
      {
        label: 'My List',
        size: 'xl',
        color: 'purple',
        variant: 'solid',
        to: '/list',
        icon: 'i-lucide-bookmark',
      },
      {
        label: 'Discover',
        size: 'xl',
        color: 'neutral',
        variant: 'ghost',
        to: '/discover',
        icon: 'i-lucide-globe',
      },
    ]
  }
  return [
    {
      label: "Get Started — It's Free",
      size: 'xl',
      color: 'purple',
      variant: 'solid',
      to: '/register',
    },
    {
      label: 'Discover',
      size: 'xl',
      color: 'neutral',
      variant: 'ghost',
      to: '/discover',
      icon: 'i-lucide-globe',
    },
  ]
})

const features = [
  {
    icon: 'i-lucide-list-checks',
    title: 'Build your watchlist',
    description:
      "Add movies and series you want to watch. Keep track of what you've already seen, all in one place.",
  },
  {
    icon: 'i-lucide-star',
    title: 'Rate & review',
    description:
      'Score every title out of 10 and write your own critic-style review to capture your thoughts.',
  },
  {
    icon: 'i-lucide-compass',
    title: 'Discover new titles',
    description:
      'Browse by genre, year, or rating to find your next favourite film or binge-worthy series.',
  },
  {
    icon: 'i-lucide-tv-2',
    title: 'Movies & series',
    description:
      'Rate-It covers both films and TV series so your entire watchlist lives in one place.',
  },
]
</script>

<template>
  <div>
    <UPageHero
      description="Your personal space to track every movie and series you watch, rate them, and write reviews you'll actually care about."
      :links="primaryLinks"
    >
      <template #title>
        Track what you watch.<br />
        <span class="text-purple-500">Own your taste.</span>
      </template>
    </UPageHero>

    <UPageSection :features="features" />

    <UPageSection>
      <UPageCTA
        title="Start tracking in seconds."
        description="Free forever. No algorithm. Just your list."
        variant="subtle"
        :links="primaryLinks"
      />
    </UPageSection>
  </div>
</template>
