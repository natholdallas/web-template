export type ThemeMode = 'system' | 'dark' | 'light'

type Conf = { theme: ThemeMode }

function loadConf(): Conf {
  if (typeof window === 'undefined') {
    return { theme: 'system' }
  }
  try {
    const raw = window.localStorage.getItem('conf')
    if (raw) {
      return { theme: 'system', ...JSON.parse(raw) }
    }
  } catch {
    // ignore
  }
  return { theme: 'system' }
}

function persist() {
  if (typeof window === 'undefined') {
    return
  }
  try {
    window.localStorage.setItem('conf', JSON.stringify(conf))
  } catch {
    // ignore
  }
}

function isSystemDark() {
  return (
    typeof window !== 'undefined' &&
    window.matchMedia('(prefers-color-scheme: dark)').matches
  )
}

function resolveTheme(theme: ThemeMode): Exclude<ThemeMode, 'system'> {
  return theme === 'system' ? (isSystemDark() ? 'dark' : 'light') : theme
}

function $switchTheme() {
  conf.theme = resolveTheme(conf.theme) === 'dark' ? 'light' : 'dark'
}

const conf = reactive<Conf & { $switchTheme: typeof $switchTheme }>({
  ...loadConf(),
  $switchTheme,
})

watch(conf, persist, { deep: true })

export function useConf() {
  return conf
}