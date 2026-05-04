<script setup lang="ts">
const { user } = useAuth()

interface LandingLink {
  label: string
  to: string
  variant: 'solid' | 'outline' | 'ghost'
  color?: 'primary' | 'neutral'
  icon?: string
}

const primaryLinks = computed<LandingLink[]>(() => {
  if (user.value) {
    return [
      {
        label: 'My List',
        to: '/list',
        variant: 'solid',
        color: 'primary',
        icon: 'i-lucide-bookmark',
      },
      {
        label: 'Discover',
        to: '/discover',
        variant: 'ghost',
        color: 'neutral',
        icon: 'i-lucide-globe',
      },
    ]
  }
  return [
    {
      label: "Get Started — It's Free",
      to: '/register',
      variant: 'solid',
      color: 'primary',
    },
    {
      label: 'Discover',
      to: '/discover',
      variant: 'ghost',
      color: 'neutral',
      icon: 'i-lucide-globe',
    },
  ]
})
</script>

<template>
  <div class="bg-black">
    <LandingHero :links="primaryLinks" />
    <LandingTrending />
    <LandingFeatures />
    <LandingCTA :links="primaryLinks" />
  </div>
</template>
