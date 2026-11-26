import { api, Config } from '@sdk/client/adm'

export const useAppStore = defineStore('app', () => {
  const cfg = ref<Config>(inst(Config))
  const loaded = ref(false)

  // Fetch the public backend config once (media base URL). Failures are
  // non-fatal: the page still renders, media just resolves against the default.
  async function $load() {
    if (loaded.value) return
    try {
      Object.assign(cfg.value, await api.config.find({ meta: { silent: true } }))
      loaded.value = true
    } catch {
      loaded.value = false
    }
  }

  // Resolve a stored media path (or already-absolute URL) to a full URL. A
  // leading slash means the path already carries the media route, so it must
  // not be prefixed with the media base again.
  function mediaUrl(path: string): string {
    if (!path) return ''
    if (/^https?:\/\//.test(path)) return path
    const apiBase = (useRuntimeConfig().public.apiBase || '').replace(/\/+$/, '')
    if (path.startsWith('/')) return `${apiBase}${path}`
    const base = (cfg.value.media || '/media').trim().replace(/\/+$/, '')
    if (/^https?:\/\//.test(base)) return `${base}/${path.replace(/^\/+/, '')}`
    const route = base.startsWith('/') ? base : `/${base}`
    return `${apiBase}${route}/${path.replace(/^\/+/, '')}`
  }

  return { cfg, loaded, $load, mediaUrl }
})
