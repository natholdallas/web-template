export default defineNuxtRouteMiddleware(async (to) => {
  const auth = useAuthStore()
  // Pages that must stay reachable without a session: the sign-in screen and
  // the OAuth callback (which may be a bind for an already signed-in user).
  const publicPages = ['entrance', 'oauth-google']
  if (auth.accessToken) return to.name === 'entrance' ? navigateTo({ name: 'index' }) : undefined
  if (publicPages.includes(String(to.name))) return
  if (await auth.$restore()) return
  return navigateTo({ name: 'entrance' })
})
