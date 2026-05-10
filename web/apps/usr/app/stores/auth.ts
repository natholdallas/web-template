import { Auth } from '~/lib/sdk'

export const useAuth = defineStore('auth', {
  state: () => inst(Auth),
  actions: {
    $signIn({ data }: { data: Auth }) {
      this.$patch(data)
      navigateTo({ name: 'home' })
    },
    $signOut() {
      this.$reset()
      navigateTo({ name: 'entrance' })
    },
  },
  persist: { storage: piniaPluginPersistedstate.localStorage() },
})
