<!-- app/components/LandingFeatures.vue -->
<script setup lang="ts">
const features = [
  {
    icon: 'i-lucide-list-checks',
    label: 'Watchlist',
    title: 'Build your watchlist',
    description: "Add films and series you want to watch. Track what you've seen — all in one place.",
    accent: 'from-purple-600 to-violet-600',
  },
  {
    icon: 'i-lucide-star',
    label: 'Rate',
    title: 'Rate & review',
    description: 'Score every title out of 10 and write your own critic-style review.',
    accent: 'from-yellow-500 to-amber-500',
  },
  {
    icon: 'i-lucide-compass',
    label: 'Discover',
    title: 'Discover new titles',
    description: 'Browse by genre, year, or rating. Find your next favourite film.',
    accent: 'from-sky-500 to-blue-600',
  },
  {
    icon: 'i-lucide-tv-2',
    label: 'Movies & Series',
    title: 'Movies & series',
    description: 'Films and TV shows — your entire watchlist in one place.',
    accent: 'from-rose-500 to-pink-600',
  },
]
</script>

<template>
  <section class="py-24 bg-background relative overflow-hidden">
    <!-- Faint background glow -->
    <div
      class="absolute inset-0 pointer-events-none"
      style="background: radial-gradient(ellipse 80% 40% at 50% 100%, rgba(88,28,135,0.25) 0%, transparent 70%);"
    />

    <UContainer class="relative z-10">
      <!-- Section header -->
      <div class="text-center mb-16">
        <p class="text-xs font-semibold tracking-widest uppercase text-purple-400 mb-3">
          Everything you need
        </p>
        <h2 class="font-display text-5xl sm:text-6xl md:text-7xl text-default leading-none">
          MADE FOR FILM LOVERS
        </h2>
      </div>

      <!-- Feature grid -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
        <div
          v-for="feature in features"
          :key="feature.label"
          class="glass-card rounded-2xl p-6 flex flex-col gap-4 group hover:border-purple-600/50 transition-colors duration-300"
        >
          <!-- Icon -->
          <div
            class="size-12 rounded-xl flex items-center justify-center bg-gradient-to-br shrink-0"
            :class="feature.accent"
          >
            <UIcon :name="feature.icon" class="size-6 text-white" />
          </div>

          <!-- Text -->
          <div>
            <p class="text-xs font-semibold tracking-widest uppercase text-muted/70 mb-1">
              {{ feature.label }}
            </p>
            <h3 class="text-lg font-bold text-default mb-2">
              {{ feature.title }}
            </h3>
            <p class="text-sm text-muted leading-relaxed">
              {{ feature.description }}
            </p>
          </div>
        </div>
      </div>
    </UContainer>
  </section>
</template>
