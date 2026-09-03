import { useLocalStorage } from '@vueuse/core'
import type { Auth as AuthType } from '~/lib/sdk/models'
import Apis from '~/lib/sdk'
import { Auth, RefreshIn } from '~/lib/sdk/models'

let refreshing: Promise<boolean> | null = null

export const useAuthStore = defineStore('auth', () => {
  const auth = useLocalStorage('auth', inst(Auth), { mergeDefaults: true })

  function $signIn({ data }: { data: AuthType }) {
    Object.assign(auth.value, data)
    navigateTo({ name: 'index' })
  }

  async function $refresh(): Promise<boolean> {
    const token = auth.value.refreshToken
    if (!token) {
      return false
    }
    if (refreshing) {
      return refreshing
    }
    refreshing = (async () => {
      try {
        const data = await Apis.Auth.refresh({
          data: inst(RefreshIn, { refreshToken: token }),
          meta: { isRefresh: true },
        })
        Object.assign(auth.value, { id: data.id, accessToken: data.accessToken, refreshToken: data.refreshToken })
        return true
      } catch {
        return false
      }
    })()
    try {
      return await refreshing
    } finally {
      refreshing = null
    }
  }

  async function $signOut() {
    try {
      if (auth.value.refreshToken) {
        await Apis.Auth.logout({ data: inst(RefreshIn, { refreshToken: auth.value.refreshToken }) })
      }
    } catch {
      // ignore
    }
    Object.assign(auth.value, inst(Auth))
    navigateTo({ name: 'entrance' })
  }

  return { ...toRefs(auth.value), $signIn, $signOut, $refresh }
})
