import { api, Config } from '@sdk/client/usr'

export const useAppStore = defineStore('app', () => {
  const cfg = ref<Config>(inst(Config))
  const loaded = ref(false)

  // Fetch the public backend config once (login providers, media base).
  // Failures are non-fatal: the page still renders, OAuth buttons just stay hidden.
  async function $load() {
    if (loaded.value) return
    try {
      const data = await api.config.find({ meta: { silent: true } })
      Object.assign(cfg.value, data)
      loaded.value = true
    } catch {
      loaded.value = false
    }
  }

  // Resolve a stored media path (or already-absolute URL) to a full URL using
  // the backend-provided media base (custom domain or same-origin route).
  function mediaUrl(path: string): string {
    if (!path) return ''
    if (/^https?:\/\//.test(path)) return path
    const apiBase = (useRuntimeConfig().public.apiBase || '').replace(/\/+$/, '')
    // Media.URL already carries the media route (e.g. /media/2026/07/x.png), so
    // an absolute path must not be prefixed with the media base again.
    if (path.startsWith('/')) return `${apiBase}${path}`
    const base = (cfg.value.media || '/media').trim().replace(/\/+$/, '')
    if (/^https?:\/\//.test(base)) return `${base}/${path.replace(/^\/+/, '')}`
    const route = base.startsWith('/') ? base : `/${base}`
    return `${apiBase}${route}/${path.replace(/^\/+/, '')}`
  }

  // Build the Google authorization URL. state carries a per-request nonce so a
  // forged callback (login CSRF) cannot bind an attacker's Google account.
  function googleAuthURL(mode: 'login' | 'bind'): string {
    const g = cfg.value.oauth?.google
    if (!g?.clientId || !g?.redirectUri) return ''
    const nonce =
      crypto.randomUUID?.() ??
      Array.from(crypto.getRandomValues(new Uint8Array(16)), (b) => b.toString(16).padStart(2, '0')).join('')
    sessionStorage.setItem('oauth:' + mode, nonce)
    const q = new URLSearchParams({
      client_id: g.clientId,
      redirect_uri: g.redirectUri,
      response_type: 'code',
      scope: 'openid email profile',
      prompt: 'select_account',
      state: `${mode}:${nonce}`,
    })
    return 'https://accounts.google.com/o/oauth2/v2/auth?' + q.toString()
  }

  return { cfg, loaded, $load, mediaUrl, googleAuthURL }
})