export const useConf = defineStore('conf', {
  state: () => useLocalStorage('conf', { isSidebarOpen: false }),
})
