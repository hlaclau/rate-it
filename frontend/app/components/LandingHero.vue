<!-- app/components/LandingHero.vue -->
<script setup lang="ts">
interface HeroLink {
  label: string
  to: string
  variant: 'solid' | 'outline' | 'ghost'
  color?: 'primary' | 'neutral'
  icon?: string
}

interface MediaResult {
  id: number
  media_type: 'movie' | 'tv'
  poster_path: string | null
}

interface MediaResponse {
  results: MediaResult[]
}

defineProps<{ links: HeroLink[] }>()

const config = useRuntimeConfig()

const { data, status } = await useFetch<MediaResponse>(
  `${config.public.apiBase}/api/media/search?sort_by=popularity.desc&page=1`,
  { lazy: true }
)

const posters = computed(() =>
  (data.value?.results ?? [])
    .filter((r) => r.poster_path)
    .slice(0, 16)
    .map((r) => `https://image.tmdb.org/t/p/w342${r.poster_path}`)
)
</script>

<template>
  <section class="relative min-h-screen flex items-center overflow-hidden bg-background">
    <!-- Poster mosaic background -->
    <div
      v-show="status === 'success' && posters.length > 0"
      class="absolute inset-0 grid opacity-20"
      style="grid-template-columns: repeat(8, 1fr); grid-template-rows: repeat(2, 1fr);"
    >
      <div
        v-for="(src, i) in posters"
        :key="i"
        class="overflow-hidden"
      >
        <img
          :src="src"
          alt=""
          class="w-full h-full object-cover scale-110"
          loading="lazy"
        />
      </div>
    </div>

    <!-- Gradient overlay -->
    <div class="absolute inset-0 hero-overlay" />

    <!-- Purple radial glow -->
    <div
      class="absolute inset-0 pointer-events-none"
      style="background: radial-gradient(ellipse 60% 50% at 30% 60%, rgba(168,85,247,0.18) 0%, transparent 70%);"
    />

    <!-- Content -->
    <UContainer class="relative z-10 py-32">
      <div class="max-w-3xl">
        <!-- Eyebrow -->
        <div class="flex items-center gap-2 mb-6">
          <span class="inline-block w-8 h-px bg-purple-400" />
          <span class="text-xs font-semibold tracking-widest uppercase text-purple-400">
            Your personal film diary
          </span>
        </div>

        <!-- Headline -->
        <h1 class="font-display text-7xl sm:text-8xl md:text-9xl leading-none text-default mb-4">
          TRACK EVERY&nbsp;FILM.<br />
          <span class="text-gradient-purple">OWN YOUR TASTE.</span>
        </h1>

        <!-- Subheadline -->
        <p class="text-lg sm:text-xl text-muted max-w-xl mt-6 mb-10 leading-relaxed">
          Build your watchlist, rate films out of 10, write reviews you'll
          actually care about. Free, forever — no algorithm, just your list.
        </p>

        <!-- CTA buttons -->
        <div class="flex flex-wrap gap-3">
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
      </div>
    </UContainer>

    <!-- Scroll cue -->
    <div class="absolute bottom-8 left-1/2 -translate-x-1/2 flex flex-col items-center gap-1 text-muted/60 text-xs tracking-widest uppercase">
      <span>Scroll</span>
      <div class="w-px h-8 bg-muted/30 animate-pulse" />
    </div>
  </section>
</template>
