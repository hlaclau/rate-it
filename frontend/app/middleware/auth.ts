export default defineNuxtRouteMiddleware((to) => {
  const { isAuthenticated, authInitialized } = useAuth()

  // Defer until fetchMe() has resolved — prevents false redirects on hard refresh
  if (!authInitialized.value) return

  if (!isAuthenticated.value) {
    return navigateTo(`/login?redirect=${encodeURIComponent(to.fullPath)}`)
  }
})
