import { useLocalStorage } from '@vueuse/core'

export type ThemeMode = 'system' | 'dark' | 'light'

type Conf = { theme: ThemeMode }

function isSystemDark() {
  return typeof window !== 'undefined' && window.matchMedia('(prefers-color-scheme: dark)').matches
}

function resolveTheme(theme: ThemeMode): Exclude<ThemeMode, 'system'> {
  return theme === 'system' ? (isSystemDark() ? 'dark' : 'light') : theme
}

export const useConfStore = defineStore('conf', () => {
  const conf = useLocalStorage<Conf>('conf', { theme: 'system' }, { mergeDefaults: true })

  function $switchTheme() {
    conf.value.theme = resolveTheme(conf.value.theme) === 'dark' ? 'light' : 'dark'
  }

  return { ...toRefs(conf.value), $switchTheme }
})
