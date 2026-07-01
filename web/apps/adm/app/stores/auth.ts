import { Auth } from '~/lib/sdk'

export const useAuth = defineStore('auth', {
  state: () => useLocalStorage('auth', inst(Auth)),
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
})
