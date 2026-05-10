export const useConf = defineStore('conf', {
  state: () => ({ theme: 'dark' }),
  persist: { storage: piniaPluginPersistedstate.localStorage() },
})
