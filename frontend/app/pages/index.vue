<template>
  <div>
    <UPageHero
      description="Rate-It lets you keep a diary of every film you see, rate them, and share your taste with the community."
      :links="[
        {
          label: 'Get Started — It\'s Free',
          size: 'xl',
          color: 'primary',
          variant: 'solid',
        }
      ]"
    >
      <template #title>
        Track films you watch.<br />
        Tell friends what's good.
      </template>
    </UPageHero>

    <UPageSection
      :features="[
        {
          icon: 'i-lucide-notebook-pen',
          title: 'Log your watches',
          description: 'Keep track of every film you\'ve seen, when you watched it, and what you thought.',
        },
        {
          icon: 'i-lucide-star',
          title: 'Rate & review',
          description: 'Score films your way and write reviews the community can discover.',
        },
        {
          icon: 'i-lucide-users',
          title: 'Join the conversation',
          description: 'See what friends are watching, follow film lovers, and build your taste profile.',
        }
      ]"
    />

    <UPageSection>
      <UPageCTA
        title="Ready to start tracking?"
        variant="subtle"
        :links="[
          {
            label: 'Create Your Free Account',
            size: 'xl',
            color: 'primary',
            variant: 'solid',
          }
        ]"
      />
    </UPageSection>
  </div>
</template>
