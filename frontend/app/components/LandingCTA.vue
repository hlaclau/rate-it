<!-- app/components/LandingCTA.vue -->
<script setup lang="ts">
interface CTALink {
  label: string
  to: string
  variant: 'solid' | 'outline' | 'ghost'
  color?: 'primary' | 'neutral'
  icon?: string
}

defineProps<{ links: CTALink[] }>()
</script>

<template>
  <section class="py-28 bg-black relative overflow-hidden">
    <!-- Animated gradient background -->
    <div
      class="absolute inset-0 pointer-events-none"
      style="background: radial-gradient(ellipse 70% 60% at 50% 50%, rgba(88,28,135,0.35) 0%, rgba(59,7,100,0.15) 50%, transparent 70%);"
    />

    <!-- Subtle grid pattern -->
    <div
      class="absolute inset-0 opacity-5 pointer-events-none"
      style="background-image: linear-gradient(rgba(168,85,247,0.5) 1px, transparent 1px), linear-gradient(90deg, rgba(168,85,247,0.5) 1px, transparent 1px); background-size: 48px 48px;"
    />

    <UContainer class="relative z-10 text-center">
      <!-- Badge -->
      <div class="inline-flex items-center gap-2 px-4 py-1.5 rounded-full border border-purple-700/40 bg-purple-950/40 mb-8">
        <span class="size-1.5 rounded-full bg-green-400 animate-pulse" />
        <span class="text-xs font-medium text-white/70 tracking-wide">Free forever · No credit card</span>
      </div>

      <h2 class="font-display text-6xl sm:text-7xl md:text-8xl text-white leading-none mb-6">
        START TRACKING<br />
        <span class="text-gradient-purple">IN SECONDS.</span>
      </h2>

      <p class="text-white/50 text-lg max-w-md mx-auto mb-10 leading-relaxed">
        No algorithm pushing you what to watch. Just your list, your ratings, your taste.
      </p>

      <div class="flex flex-wrap justify-center gap-3">
        <UButton
          v-for="link in links"
          :key="link.label"
          :to="link.to"
          :variant="link.variant"
          :color="link.color ?? 'primary'"
          :icon="link.icon"
          size="xl"
        >
          {{ link.label }}
        </UButton>
      </div>
    </UContainer>
  </section>
</template>
