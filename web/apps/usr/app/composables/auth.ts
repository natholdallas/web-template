import Apis from '~/lib/sdk'
import { Auth, RefreshIn } from '~/lib/sdk/models'
import type { Auth as AuthType } from '~/lib/sdk/models'

let refreshing = false

function loadAuth(): AuthType {
  if (typeof window === 'undefined') return inst(Auth)
  try {
    const raw = window.localStorage.getItem('auth')
    if (raw) return { ...inst(Auth), ...JSON.parse(raw) }
  } catch {
    // ignore
  }
  return inst(Auth)
}

function persist() {
  if (typeof window === 'undefined') return
  try {
    window.localStorage.setItem('auth', JSON.stringify(auth))
  } catch {
    // ignore
  }
}

function $signIn({ data }: { data: AuthType }) {
  Object.assign(auth, data)
  navigateTo({ name: 'index' })
}

async function $refresh(): Promise<boolean> {
  if (refreshing || !auth.refreshToken) return false
  refreshing = true
  try {
    const data = await Apis.Auth.refresh({
      data: inst(RefreshIn, { refreshToken: auth.refreshToken }),
    })
    Object.assign(auth, { id: data.id, accessToken: data.accessToken, refreshToken: data.refreshToken })
    return true
  } catch {
    return false
  } finally {
    refreshing = false
  }
}

async function $signOut() {
  try {
    if (auth.refreshToken) {
      await Apis.Auth.logout({ data: inst(RefreshIn, { refreshToken: auth.refreshToken }) })
    }
  } catch {
    // ignore
  }
  Object.assign(auth, inst(Auth))
  navigateTo({ name: 'entrance' })
}

const auth = reactive<AuthType & { $signIn: typeof $signIn; $signOut: typeof $signOut; $refresh: typeof $refresh }>({
  ...loadAuth(),
  $signIn,
  $signOut,
  $refresh,
})

watch(auth, persist, { deep: true })

export function useAuth() {
  return auth
}
