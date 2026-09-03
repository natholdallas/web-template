export default defineNuxtRouteMiddleware(() => {
  if (!useAuthStore().accessToken) {
    return navigateTo({ name: 'entrance' })
  }
})
