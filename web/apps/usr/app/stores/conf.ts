import { useLocalStorage } from '@vueuse/core'

type Conf = { isSidebarOpen: boolean }

export const useConfStore = defineStore('conf', () => {
  const conf = useLocalStorage<Conf>('conf', { isSidebarOpen: false }, { mergeDefaults: true })

  return { ...toRefs(conf.value) }
})
