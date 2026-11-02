import Apis, { Config } from '~/lib/sdk'

export const useAppStore = defineStore('app', () => {
  const cfg = ref<Config>(inst(Config))
  const loaded = ref(false)

  // Fetch the public backend config once (login providers, media base).
  async function $load() {
    if (loaded.value) return
    const data = await Apis.Config.find({ meta: { silent: true } })
    Object.assign(cfg.value, data)
    loaded.value = true
  }

  // Resolve a stored media path (or already-absolute URL) to a full URL using
  // the backend-provided media base (custom domain or same-origin route).
  function mediaUrl(path: string): string {
    if (!path) return ''
    if (/^https?:\/\//.test(path)) return path
    const base = (cfg.value.media || '/media').trim()
    const p = path.replace(/^\/+/, '')
    if (/^https?:\/\//.test(base)) return base + '/' + p
    const apiBase = useRuntimeConfig().public.apiBase || ''
    const route = base.startsWith('/') ? base : '/' + base
    return apiBase + route + '/' + p
  }

  // Build the Google authorization URL. state distinguishes login vs bind.
  function googleAuthURL(mode: 'login' | 'bind'): string {
    const g = cfg.value.oauth?.google
    if (!g?.clientId || !g?.redirectUri) return ''
    const q = new URLSearchParams({
      client_id: g.clientId,
      redirect_uri: g.redirectUri,
      response_type: 'code',
      scope: 'openid email profile',
      prompt: 'select_account',
      state: mode,
    })
    return 'https://accounts.google.com/o/oauth2/v2/auth?' + q.toString()
  }

  return { cfg, loaded, $load, mediaUrl, googleAuthURL }
})