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
        <UForm
          :state="state"
          :validate="validate"
          class="w-full"
          @submit="onSubmit"
        >
          <div class="mb-6">
            <UFormGroup label="Email" name="email" class="w-full">
              <UInput
                v-model="state.email"
                placeholder="you@example.com"
                icon="i-lucide-mail"
                size="lg"
                variant="outline"
                color="primary"
                class="w-full"
              />
            </UFormGroup>
          </div>

          <UFormGroup label="Password" name="password" class="w-full">
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
                  @click="showPassword = !showPassword"
                />
              </template>
            </UInput>
          </UFormGroup>

          <UButton
            type="submit"
            block
            size="lg"
            color="primary"
            :loading="loading"
            class="mt-8 transition-transform hover:scale-[1.02] active:scale-[0.98]"
          >
            Sign in
          </UButton>
        </UForm>

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

const validate = (state: { email: string; password: string }) => {
  const errors = []
  if (!state.email) {
    errors.push({ path: 'email', message: 'Required' })
  } else if (!state.email.includes('@') || !state.email.includes('.')) {
    errors.push({ path: 'email', message: 'Enter a valid email address' })
  }
  if (!state.password) {
    errors.push({ path: 'password', message: 'Required' })
  }
  return errors
}

async function onSubmit() {
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
