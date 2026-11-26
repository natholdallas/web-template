import { setupSdk } from '@sdk/client'

// The SDK package is framework-agnostic, so the app injects its primitives
// here. Plugins run before the global route middleware, which is what restores
// the session through `api.auth.refresh()` — so the hooks are always installed
// before the first request.
export default defineNuxtPlugin(() => {
  setupSdk({
    baseURL: () => useRuntimeConfig().public.apiBase,
    auth: () => useAuthStore(),
    t: (key) => useNuxtApp().$i18n.t(key) as string,
    notify: {
      success: (message) => useSnackBar().success(message),
      warn: (message) => useSnackBar().warn(message),
      error: (message) => useSnackBar().error(message),
    },
  })
})
