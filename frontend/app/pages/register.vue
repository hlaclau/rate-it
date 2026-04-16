<template>
  <div class="flex items-center justify-center py-20 px-4 transition-colors duration-300">
    <div class="w-full max-w-md">
      <!-- Header -->
      <div class="text-center mb-8">
        <h1 class="text-4xl font-bold text-zinc-900 dark:text-white tracking-tight mb-2">
          Rate<span class="text-purple-600 dark:text-purple-500">-It</span>
        </h1>
        <p class="text-zinc-500 dark:text-zinc-400">Join the Palace Curators.</p>
      </div>

      <!-- Card -->
      <div class="bg-white dark:bg-zinc-900/50 backdrop-blur-xl border border-zinc-200 dark:border-zinc-800 p-8 rounded-2xl shadow-xl dark:shadow-2xl w-full">
        <UForm :state="state" :validate="validate" class="w-full" @submit="onSubmit">

          <div class="mb-6">
            <UFormGroup label="Username" name="username" class="w-full">
              <UInput v-model="state.username" placeholder="curator01" icon="i-lucide-user" size="lg" variant="outline"
                color="primary" class="w-full" />
            </UFormGroup>
          </div>

          <div class="mb-6">
            <UFormGroup label="Email" name="email" class="w-full">
              <UInput v-model="state.email" placeholder="curator@midnight.com" icon="i-lucide-mail" size="lg"
                variant="outline" color="primary" class="w-full" />
            </UFormGroup>
          </div>

          <div class="mb-6">
            <UFormGroup label="Password" name="password" class="w-full">
              <UInput v-model="state.password" :type="showPassword ? 'text' : 'password'" placeholder="••••••••"
                icon="i-lucide-lock" size="lg" variant="outline" color="primary" class="w-full">
                <template #trailing>
                  <UButton :icon="showPassword ? 'i-lucide-eye-off' : 'i-lucide-eye'" color="neutral" variant="ghost"
                    size="xs" tabindex="-1" @click="showPassword = !showPassword" />
                </template>
              </UInput>
            </UFormGroup>
          </div>

          <div class="mb-6">
            <UFormGroup label="Confirm Password" name="confirmPassword" class="w-full">
              <UInput v-model="state.confirmPassword" :type="showConfirmPassword ? 'text' : 'password'"
                placeholder="••••••••" icon="i-lucide-shield-check" size="lg" variant="outline" color="primary"
                class="w-full">
                <template #trailing>
                  <UButton :icon="showConfirmPassword ? 'i-lucide-eye-off' : 'i-lucide-eye'" color="neutral"
                    variant="ghost" size="xs" tabindex="-1" @click="showConfirmPassword = !showConfirmPassword" />
                </template>
              </UInput>
            </UFormGroup>
          </div>

          <UButton type="submit" block size="lg" color="primary" :loading="loading" class="mt-2 transition-transform hover:scale-[1.02] active:scale-[0.98]">
            Create Account
          </UButton>
        </UForm>

        <div class="mt-6 text-center text-sm text-zinc-500 dark:text-zinc-400">
          Already a curator?
          <NuxtLink to="/login" class="text-purple-600 dark:text-purple-400 hover:text-purple-500 dark:hover:text-purple-300 font-medium transition-colors">
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
  confirmPassword: ''
})

const validate = (state: any) => {
  const errors = []
  if (!state.username) errors.push({ path: 'username', message: 'Required' })
  if (!state.email) {
    errors.push({ path: 'email', message: 'Required' })
  } else if (!state.email.includes('@') || !state.email.includes('.')) {
    errors.push({ path: 'email', message: 'Enter a valid email address' })
  }
  if (!state.password) {
    errors.push({ path: 'password', message: 'Required' })
  } else if (state.password.length < 8) {
    errors.push({ path: 'password', message: 'Must be at least 8 characters' })
  }
  if (state.password && state.password !== state.confirmPassword) {
    errors.push({ path: 'confirmPassword', message: 'Passwords do not match' })
  }
  return errors
}

async function onSubmit() {
  loading.value = true
  try {
    await register(state.username, state.email, state.password)
  } catch (err: any) {
    useToast().add({
      title: 'Registration failed',
      description: err.data?.message || 'Something went wrong.',
      color: 'error'
    })
  } finally {
    loading.value = false
  }
}

definePageMeta({
  layout: false
})
</script>
