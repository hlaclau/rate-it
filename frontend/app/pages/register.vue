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
        <p class="text-zinc-500 dark:text-zinc-400">Create your account.</p>
      </div>

      <div
        class="bg-white dark:bg-zinc-900/50 backdrop-blur-xl border border-zinc-200 dark:border-zinc-800 p-8 rounded-2xl shadow-xl dark:shadow-2xl w-full"
      >
        <form class="w-full" @submit.prevent="onSubmit">
          <div class="mb-6">
            <UFormField
              label="Username"
              name="username"
              :error="errors.username || undefined"
              class="w-full"
            >
              <UInput
                v-model="state.username"
                placeholder="username"
                icon="i-lucide-user"
                size="lg"
                variant="outline"
                color="primary"
                class="w-full"
              />
            </UFormField>
          </div>

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

          <div class="mb-6">
            <UFormField
              label="Password"
              name="password"
              :error="errors.password || undefined"
              help="Min. 8 characters, including 1 uppercase, 1 lowercase and 1 number"
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
          </div>

          <div class="mb-6">
            <UFormField
              label="Confirm Password"
              name="confirmPassword"
              :error="errors.confirmPassword || undefined"
              class="w-full"
            >
              <UInput
                v-model="state.confirmPassword"
                :type="showConfirmPassword ? 'text' : 'password'"
                placeholder="••••••••"
                icon="i-lucide-shield-check"
                size="lg"
                variant="outline"
                color="primary"
                class="w-full"
              >
                <template #trailing>
                  <UButton
                    :icon="
                      showConfirmPassword ? 'i-lucide-eye-off' : 'i-lucide-eye'
                    "
                    color="neutral"
                    variant="ghost"
                    size="xs"
                    tabindex="-1"
                    type="button"
                    @click="showConfirmPassword = !showConfirmPassword"
                  />
                </template>
              </UInput>
            </UFormField>
          </div>

          <button
            type="submit"
            :disabled="loading"
            class="mt-2 w-full flex items-center justify-center px-4 py-2.5 rounded-lg text-sm font-semibold text-white bg-purple-600 hover:bg-purple-500 disabled:opacity-60 disabled:cursor-not-allowed transition-transform hover:scale-[1.02] active:scale-[0.98]"
          >
            <span v-if="loading" class="flex items-center gap-2">
              <UIcon
                name="i-lucide-loader-circle"
                class="size-4 animate-spin"
              />
              Creating account…
            </span>
            <span v-else>Create Account</span>
          </button>
        </form>

        <div class="mt-6 text-center text-sm text-zinc-500 dark:text-zinc-400">
          Already have an account?
          <NuxtLink
            to="/login"
            class="text-purple-600 dark:text-purple-400 hover:text-purple-500 dark:hover:text-purple-300 font-medium transition-colors"
          >
            Sign in
          </NuxtLink>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const { register } = useAuth()
const loading = ref(false)
const showPassword = ref(false)
const showConfirmPassword = ref(false)

const state = reactive({
  username: '',
  email: '',
  password: '',
  confirmPassword: '',
})

const errors = reactive({
  username: '',
  email: '',
  password: '',
  confirmPassword: '',
})

function validate() {
  errors.username = ''
  errors.email = ''
  errors.password = ''
  errors.confirmPassword = ''
  let valid = true
  if (!state.username) {
    errors.username = 'Required'
    valid = false
  }
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
  } else {
    if (state.password.length < 8) {
      errors.password = 'At least 8 characters'
      valid = false
    } else if (!/[A-Z]/.test(state.password)) {
      errors.password = 'At least one uppercase letter'
      valid = false
    } else if (!/[a-z]/.test(state.password)) {
      errors.password = 'At least one lowercase letter'
      valid = false
    } else if (!/[0-9]/.test(state.password)) {
      errors.password = 'At least one number'
      valid = false
    }
  }
  if (state.password && state.password !== state.confirmPassword) {
    errors.confirmPassword = 'Passwords do not match'
    valid = false
  }
  return valid
}

async function onSubmit() {
  if (!validate()) return
  loading.value = true
  try {
    await register(state.username, state.email, state.password)
  } catch (err) {
    const error = err as { data?: { message?: string } }
    useToast().add({
      title: 'Registration failed',
      description: error.data?.message || 'Something went wrong.',
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
