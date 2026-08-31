export default defineNuxtRouteMiddleware(() => {
  if (useAuth().accessToken) {
    return navigateTo({ name: 'index' })
  }
})
