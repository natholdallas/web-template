export default defineNuxtRouteMiddleware(() => {
  if (!useAuth().token) {
    return navigateTo({ name: 'entrance' })
  }
})
