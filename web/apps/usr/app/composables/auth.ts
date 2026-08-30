import { Auth } from '~/lib/sdk/gen/models'

function loadAuth(): Auth {
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

function $signIn({ data }: { data: Auth }) {
  Object.assign(auth, data)
  navigateTo({ name: 'index' })
}

function $signOut() {
  Object.assign(auth, inst(Auth))
  navigateTo({ name: 'entrance' })
}

const auth = reactive<Auth & { $signIn: typeof $signIn; $signOut: typeof $signOut }>({
  ...loadAuth(),
  $signIn,
  $signOut,
})

watch(auth, persist, { deep: true })

export function useAuth() {
  return auth
}
