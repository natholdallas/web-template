export const useConf = defineStore('conf', {
  state: () => useLocalStorage('conf', { theme: 'dark' }),
})
