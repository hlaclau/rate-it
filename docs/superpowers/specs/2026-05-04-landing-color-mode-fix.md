# Landing Page — Semantic Color Mode Fix

## Goal

Replace every hardcoded `bg-black` / `text-white` in the landing components and app shell with Nuxt UI semantic tokens so the whole app — landing page included — responds uniformly to the system dark/light mode preference.

## What stays the same

- Bebas Neue headlines (`.font-display`)
- Purple gradient text (`.text-gradient-purple`, `.text-gradient-gold`)
- Radial purple glow backgrounds (inline `radial-gradient` with `rgba(88,28,135,…)` — purple glow works in both modes)
- Poster mosaic with `opacity-20` (subtle either way)
- Colored gradient icon backgrounds in feature cards (`from-purple-600 to-violet-600` etc.) — always colored, white icon text on them stays `text-white`
- Hover overlays on posters: `bg-gradient-to-t from-black` is always dark (it's on a poster image), so `text-white` inside those overlays stays

## What changes

### 1. `main.css` — update `.glass-card`, add `.hero-overlay`

**`.glass-card`** currently only has dark-mode values. Add light-mode defaults and a `@variant dark` block:

```css
.glass-card {
  background: color-mix(in srgb, var(--color-purple-50) 60%, transparent);
  backdrop-filter: blur(12px);
  border: 1px solid color-mix(in srgb, var(--color-purple-200) 50%, transparent);
}

@variant dark {
  .glass-card {
    background: color-mix(in srgb, var(--color-purple-950) 40%, transparent);
    border: 1px solid color-mix(in srgb, var(--color-purple-700) 30%, transparent);
  }
}
```

**`.hero-overlay`** — new class to replace the inline `style` on the hero gradient overlay div. Light mode uses white washes, dark mode uses black washes:

```css
.hero-overlay {
  background:
    linear-gradient(to right, rgba(255,255,255,0.92) 0%, rgba(255,255,255,0.6) 50%, rgba(255,255,255,0.8) 100%),
    linear-gradient(to bottom, rgba(255,255,255,0.3) 0%, transparent 30%, transparent 70%, rgba(255,255,255,0.7) 100%);
}

@variant dark {
  .hero-overlay {
    background:
      linear-gradient(to right, rgba(0,0,0,0.95) 0%, rgba(0,0,0,0.7) 50%, rgba(0,0,0,0.85) 100%),
      linear-gradient(to bottom, rgba(0,0,0,0.4) 0%, transparent 30%, transparent 70%, rgba(0,0,0,0.9) 100%);
  }
}
```

---

### 2. `app/app.vue` — revert navbar and footer

| Element | Before | After |
|---------|--------|-------|
| `<header>` class | `border-white/8 dark:bg-black/60 bg-white/80` | `border-default bg-background/80` |
| `<footer>` class | `border-white/8 bg-black` | `border-default` |

---

### 3. `LandingHero.vue`

| Location | Before | After |
|----------|--------|-------|
| `<section>` | `bg-black` | `bg-background` |
| Gradient overlay div | `style="background: linear-gradient(…rgba(0,0,0…)…)"` | `class="hero-overlay"` (remove `style`) |
| `<h1>` | `text-white` | `text-default` |
| Subheadline `<p>` | `text-white/60` | `text-muted` |
| Scroll cue wrapper | `text-white/30` | `text-muted/60` |
| Scroll cue line | `bg-white/20` | `bg-muted/30` |

---

### 4. `LandingTrending.vue`

| Location | Before | After |
|----------|--------|-------|
| `<section>` | `bg-black` | `bg-background` |
| `<h2>` "TRENDING" | `text-white` | `text-default` |
| Desktop "View all" link | `text-white/50` | `text-muted` |
| Mobile "View all" link | `text-white/50` | `text-muted` |
| Text inside poster hover overlay | `text-white`, `text-white/60` | **unchanged** (overlay is always `from-black`) |

---

### 5. `LandingFeatures.vue`

| Location | Before | After |
|----------|--------|-------|
| `<section>` | `bg-black` | `bg-background` |
| `<h2>` "MADE FOR FILM LOVERS" | `text-white` | `text-default` |
| Card icon `UIcon` | `text-white` | **unchanged** (always on a colored gradient bg) |
| Card label `<p>` | `text-white/40` | `text-muted/70` |
| Card title `<h3>` | `text-white` | `text-default` |
| Card description `<p>` | `text-white/50` | `text-muted` |

---

### 6. `LandingCTA.vue`

| Location | Before | After |
|----------|--------|-------|
| `<section>` | `bg-black` | `bg-background` |
| Badge `<div>` | `border-purple-700/40 bg-purple-950/40` | `border-primary/30 bg-primary/10` |
| Badge `<span>` text | `text-white/70` | `text-muted` |
| `<h2>` | `text-white` | `text-default` |
| Subtext `<p>` | `text-white/50` | `text-muted` |

---

## Semantic tokens used

All from Nuxt UI v4, auto-flip with color mode:

| Token class | Light value | Dark value |
|-------------|-------------|------------|
| `bg-background` | white | near-black |
| `text-default` | near-black | near-white |
| `text-muted` | mid-grey | mid-grey |
| `border-default` | light grey | dark grey |
| `bg-primary/10` | light purple tint | dark purple tint |
| `border-primary/30` | purple/30 | purple/30 |

## Out of scope

- No changes to non-landing pages (Discover, List, movie detail, login, register)
- No changes to the purple radial glow `rgba` values in inline styles — purple glow reads well in both modes
- No changes to `.text-gradient-purple`, `.font-display`, `.poster-glow`, `.animate-fade-up`
- No changes to scrollbar utilities
