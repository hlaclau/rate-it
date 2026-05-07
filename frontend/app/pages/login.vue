<template>
  <div
    class="flex-1 flex items-center justify-center py-20 px-4 transition-colors duration-300"
  >
    <div class="w-full max-w-md">
      <div class="text-center mb-8">
        <h1
          class="text-4xl font-bold text-zinc-900 dark:text-white tracking-tight mb-2"
        >
          Rate<span class="text-purple-600 dark:text-purple-500">-It</span>
        </h1>
        <p class="text-zinc-500 dark:text-zinc-400">Welcome back.</p>
      </div>

      <div
        class="bg-white dark:bg-zinc-900/50 backdrop-blur-xl border border-zinc-200 dark:border-zinc-800 p-8 rounded-2xl shadow-xl dark:shadow-2xl w-full"
      >
        <form class="w-full" @submit.prevent="onSubmit">
          <div class="mb-6">
            <UFormField
              label="Email"
              name="email"
              :error="errors.email || undefined"
              class="w-full"
            >
              <UInput
                v-model="state.email"
                placeholder="you@example.com"
                icon="i-lucide-mail"
                size="lg"
                variant="outline"
                color="primary"
                class="w-full"
              />
            </UFormField>
          </div>

          <UFormField
            label="Password"
            name="password"
            :error="errors.password || undefined"
            class="w-full"
          >
            <UInput
              v-model="state.password"
              :type="showPassword ? 'text' : 'password'"
              placeholder="••••••••"
              icon="i-lucide-lock"
              size="lg"
              variant="outline"
              color="primary"
              class="w-full"
            >
              <template #trailing>
                <UButton
                  :icon="showPassword ? 'i-lucide-eye-off' : 'i-lucide-eye'"
                  color="neutral"
                  variant="ghost"
                  size="xs"
                  tabindex="-1"
                  type="button"
                  @click="showPassword = !showPassword"
                />
              </template>
            </UInput>
          </UFormField>

          <button
            type="submit"
            :disabled="loading"
            class="mt-8 w-full flex items-center justify-center px-4 py-2.5 rounded-lg text-sm font-semibold text-white bg-purple-600 hover:bg-purple-500 disabled:opacity-60 disabled:cursor-not-allowed transition-transform hover:scale-[1.02] active:scale-[0.98]"
          >
            <span v-if="loading" class="flex items-center gap-2">
              <UIcon
                name="i-lucide-loader-circle"
                class="size-4 animate-spin"
              />
              Signing in…
            </span>
            <span v-else>Sign in</span>
          </button>
        </form>

        <div class="mt-6 text-center text-sm text-zinc-500 dark:text-zinc-400">
          Don't have an account?
          <NuxtLink
            to="/register"
            class="text-purple-600 dark:text-purple-400 hover:text-purple-500 dark:hover:text-purple-300 font-medium transition-colors"
          >
            Create an account
          </NuxtLink>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const { login } = useAuth()
const loading = ref(false)
const showPassword = ref(false)

const state = reactive({
  email: '',
  password: '',
})

const errors = reactive({
  email: '',
  password: '',
})

function validate() {
  errors.email = ''
  errors.password = ''
  let valid = true
  if (!state.email) {
    errors.email = 'Required'
    valid = false
  } else if (!state.email.includes('@') || !state.email.includes('.')) {
    errors.email = 'Enter a valid email address'
    valid = false
  }
  if (!state.password) {
    errors.password = 'Required'
    valid = false
  }
  return valid
}

async function onSubmit() {
  if (!validate()) return
  loading.value = true
  try {
    await login(state.email, state.password)
  } catch (err) {
    const error = err as { data?: { message?: string } }
    useToast().add({
      title: 'Login failed',
      description: error.data?.message || 'Check your credentials.',
      color: 'error',
    })
  } finally {
    loading.value = false
  }
}

definePageMeta({
  layout: false,
})
</script>
