import { Auth } from '~/sdk'

export const useAuth = defineStore('auth', {
  state: () => inst(Auth),
  actions: {
    $signIn({ data }: { data: Auth }) {
      this.$patch(data)
      navigateTo({ name: 'index' })
    },
    $signOut() {
      this.$reset()
      navigateTo({ name: 'entrance' })
    },
  },
  persist: { storage: piniaPluginPersistedstate.localStorage() },
})
