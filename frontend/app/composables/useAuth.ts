export const useAuth = () => {
  const config = useRuntimeConfig()
  const user = useState<any>('auth-user', () => null)
  const isAuthenticated = computed(() => !!user.value)
  const authInitialized = useState<boolean>('auth-initialized', () => false)

  // Custom fetch function that handles credentials (cookies) and refresh logic
  const apiFetch = $fetch.create({
    baseURL: `${config.public.apiBase}/api/`,
    onRequest({ options }) {
      options.credentials = 'include'
    },
    async onResponseError({ response, options }) {
      if (response.status === 401 && !options._retry) {
        options._retry = true
        try {
          await $fetch(`${config.public.apiBase}/api/refresh`, {
            method: 'POST',
            credentials: 'include'
          })
          return apiFetch(response.url, options)
        } catch (error) {
          user.value = null
          return navigateTo('/login')
        }
      }
    }
  })

  const login = async (email: string, password: string) => {
    const route = useRoute()
    const data: any = await apiFetch('login', {
      method: 'POST',
      body: { email, password }
    })
    user.value = data
    const redirectPath = route.query.redirect as string | undefined
    return navigateTo(redirectPath || '/')
  }

  const register = async (username: string, email: string, password: string) => {
    await apiFetch('register', {
      method: 'POST',
      body: { username, email, password }
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
      const data: any = await apiFetch('me')
      user.value = data
    } catch (e) {
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
    apiFetch
  }
}
