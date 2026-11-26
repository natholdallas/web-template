import { api } from '@sdk/client/adm'
import type { Auth as AuthType } from '@sdk/client/adm'

// One in-flight refresh shared by every caller: a page load, a poller and a
// user action must not each rotate the same cookie.
let refreshing: Promise<boolean> | null = null

export const useAuthStore = defineStore('auth', () => {
  // The access token lives in memory only. The refresh token is an HttpOnly
  // cookie the browser attaches to /auth/*, so nothing usable is readable from
  // localStorage by an injected script.
  const id = ref(0)
  const accessToken = ref('')
  let restored = false

  // Builds before the cookie migration kept the whole pair in localStorage.
  // Drop it once so a stale refresh token cannot be replayed from disk.
  if (import.meta.client) localStorage.removeItem('auth')

  function $clear() {
    id.value = 0
    accessToken.value = ''
    useProfileStore().$reset()
  }

  function $signIn({ data }: { data: AuthType }) {
    id.value = data.id
    accessToken.value = data.accessToken
    useProfileStore().$reset()
    navigateTo({ name: 'index' })
  }

  async function $refresh(): Promise<boolean> {
    if (refreshing) return refreshing
    refreshing = (async () => {
      try {
        const data = await api.auth.refresh({ meta: { isRefresh: true } })
        id.value = data.id
        accessToken.value = data.accessToken
        return true
      } catch (err) {
        // Only a definitive rejection means the session is gone. A network or
        // 5xx failure must not sign the admin out during a transient outage.
        const status = err instanceof Error && 'status' in err && typeof err.status === 'number' ? err.status : undefined
        if (status === 400 || status === 401) $clear()
        return false
      }
    })()
    try {
      return await refreshing
    } finally {
      refreshing = null
    }
  }

  // Bootstrap: the page has no token after a reload, only the cookie, so the
  // first guarded navigation trades it for a fresh access token.
  async function $restore(): Promise<boolean> {
    // Never during `nuxt generate`: prerendering has no browser session.
    if (!import.meta.client) return false
    if (accessToken.value) return true
    if (restored) return false
    restored = true
    return $refresh()
  }

  async function $signOut(skipServer = false) {
    // Clear first: the logout request must not be able to re-enter the 401
    // handler with a token this store no longer holds.
    $clear()
    if (!skipServer) {
      try {
        await api.auth.logout({ meta: { noRefresh: true } })
      } catch {
        // The cookie is cleared by the response; revocation is best effort.
      }
    }
    navigateTo({ name: 'entrance' })
  }

  return { id, accessToken, $signIn, $refresh, $restore, $clear, $signOut }
})
