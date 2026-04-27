export interface User {
  id: string
  username: string
  email: string
  bio?: string
  avatar_url?: string
}

export const useAuth = () => {
  const config = useRuntimeConfig()
  const user = useState<User | null>('auth-user', () => null)
  const isAuthenticated = computed(() => !!user.value)
  const authInitialized = useState<boolean>('auth-initialized', () => false)

  // Custom fetch function that handles credentials (cookies) and refresh logic
  // We explicitly type apiFetch to avoid the "implicitly has any" error during circular initialization
  const apiFetch: typeof $fetch = $fetch.create({
    baseURL: `${config.public.apiBase}/api/`,
    onRequest({ options }) {
      options.credentials = 'include'
    },
    async onResponseError({ response, options }) {
      // Create a local type-safe version of options to use our custom _retry property
      const fetchOptions = options as typeof options & { _retry?: boolean }

      if (response.status === 401 && !fetchOptions._retry) {
        fetchOptions._retry = true
        try {
          await $fetch(`${config.public.apiBase}/api/refresh`, {
            method: 'POST',
            credentials: 'include',
          })
          // We cast to Parameters<typeof apiFetch>[1] to bridge the gap between 
          // ResolvedFetchOptions and NitroFetchOptions without using 'any'
          await apiFetch(response.url, fetchOptions as Parameters<typeof apiFetch>[1])
        } catch {
          user.value = null
          await navigateTo('/login')
        }
      }
    },
  })

  const login = async (email: string, password: string) => {
    const route = useRoute()
    const data = await apiFetch<User>('login', {
      method: 'POST',
      body: { email, password },
    })
    user.value = data
    const redirectPath = route.query.redirect as string | undefined
    return navigateTo(redirectPath || '/')
  }

  const register = async (
    username: string,
    email: string,
    password: string
  ) => {
    await apiFetch('register', {
      method: 'POST',
      body: { username, email, password },
    })
    await login(email, password)
  }

  const logout = async () => {
    await apiFetch('logout', { method: 'POST' })
    user.value = null
    return navigateTo('/login')
  }

  const fetchMe = async () => {
    try {
      const data = await apiFetch<User>('me')
      user.value = data
    } catch {
      user.value = null
    } finally {
      authInitialized.value = true
    }
  }

  return {
    user,
    isAuthenticated,
    authInitialized,
    login,
    register,
    logout,
    fetchMe,
    apiFetch,
  }
}
