import { openUrl } from '@tauri-apps/plugin-opener'

// `window.open` is a no-op inside the Tauri webview (it has no browser tabs), so
// every external link must go through the opener plugin when running in the
// shell and fall back to a plain tab when running as a normal web app.
export function useOpener() {
  function isTauri() {
    return typeof window !== 'undefined' && '__TAURI_INTERNALS__' in window
  }
  async function openExternal(url: string) {
    // Only web URLs are ever handed to the OS: the native side is scoped to
    // http(s), and a non-web scheme must not reach system handlers.
    if (!/^https?:\/\//i.test(url)) return
    if (isTauri()) {
      try {
        await openUrl(url)
      } catch {
        // A failed hand-off must not break the caller's click handler.
      }
      return
    }
    window.open(url, '_blank', 'noopener')
  }
  return { isTauri, openExternal }
}
