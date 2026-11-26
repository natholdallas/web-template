// True only inside the Tauri desktop/mobile webview. Google rejects embedded
// webviews outright and both OAuth flows finish on a web origin that cannot
// hand control back to the shell, so the shell degrades to password sign-in.
//
// Initialised synchronously (the app is ssr:false, so `window` exists by the
// time any component runs): an onMounted flip would render the OAuth buttons
// for one frame inside the shell.
export function useIsTauri() {
  return ref(typeof window !== 'undefined' && '__TAURI_INTERNALS__' in window)
}
