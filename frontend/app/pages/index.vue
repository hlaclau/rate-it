<script setup lang="ts">
const { user, logout } = useAuth()

const links = computed(() => {
  if (user.value) {
    return [
      {
        label: 'Go to Diary',
        size: 'xl',
        color: 'purple',
        variant: 'solid',
        to: '/diary'
      },
      {
        label: 'Logout',
        size: 'xl',
        color: 'gray',
        variant: 'ghost',
        click: logout
      }
    ]
  }
  return [
    {
      label: 'Get Started — It\'s Free',
      size: 'xl',
      color: 'purple',
      variant: 'solid',
      to: '/register'
    }
  ]
})
</script>

<template>
  <div>
    <UPageHero
      description="Rate-It lets you keep a diary of every film you see, rate them, and share your taste with the community."
      :links="links">
      <template #title>
        Track films you watch.<br />
        Tell friends what's good.
      </template>
    </UPageHero>

    <UPageSection :features="[
      {
        icon: 'i-lucide-notebook-pen',
        title: 'Log your watches',
        description:
          'Keep track of every film you\'ve seen, when you watched it, and what you thought.',
      },
      {
        icon: 'i-lucide-star',
        title: 'Rate & review',
        description:
          'Score films your way and write reviews the community can discover.',
      },
      {
        icon: 'i-lucide-users',
        title: 'Join the conversation',
        description:
          'See what friends are watching, follow film lovers, and build your taste profile.',
      },
    ]" />

    <UPageSection>
      <UPageCTA title="Ready to start tracking?" variant="subtle" :links="links" />
    </UPageSection>
  </div>
</template>
