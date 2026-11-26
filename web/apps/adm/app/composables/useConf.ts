import { useConfStore } from '~/stores/conf'

// App-side adapter for the shared vuetify module's Error provider, which
// expects a `useConf()` composable exposing the resolved theme.
export function useConf() {
  return useConfStore()
}