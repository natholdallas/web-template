type Conf = { theme: string }

function loadConf(): Conf {
  if (typeof window === 'undefined') return { theme: 'dark' }
  try {
    const raw = window.localStorage.getItem('conf')
    if (raw) return { theme: 'dark', ...JSON.parse(raw) }
  } catch {
    // ignore
  }
  return { theme: 'dark' }
}

function persist() {
  if (typeof window === 'undefined') return
  try {
    window.localStorage.setItem('conf', JSON.stringify(conf))
  } catch {
    // ignore
  }
}

function $switchTheme() {
  conf.theme = conf.theme === 'dark' ? 'light' : 'dark'
}

const conf = reactive<Conf & { $switchTheme: typeof $switchTheme }>({
  ...loadConf(),
  $switchTheme,
})

watch(conf, persist, { deep: true })

export function useConf() {
  return conf
}