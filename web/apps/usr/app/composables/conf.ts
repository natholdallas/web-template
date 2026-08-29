type Conf = { isSidebarOpen: boolean }

function loadConf(): Conf {
  if (typeof window === 'undefined') return { isSidebarOpen: false }
  try {
    const raw = window.localStorage.getItem('conf')
    if (raw) return { isSidebarOpen: false, ...JSON.parse(raw) }
  } catch {
    // ignore
  }
  return { isSidebarOpen: false }
}

function persist() {
  if (typeof window === 'undefined') return
  try {
    window.localStorage.setItem('conf', JSON.stringify(conf))
  } catch {
    // ignore
  }
}

const conf = reactive<Conf>(loadConf())

watch(conf, persist, { deep: true })

export function useConf() {
  return conf
}