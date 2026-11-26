import { createAlova } from 'alova'
import type { Method } from 'alova'
import adapterFetch from 'alova/fetch'
import VueHook from 'alova/vue'

/**
 * The single alova instance shared by every generated app SDK.
 *
 * This module is framework-agnostic on purpose: it lives in a workspace package
 * and must not reach for Nuxt/Pinia composables (they do not exist at module
 * scope, and each app exposes different toast primitives). The app injects
 * its own baseURL / session / toast / i18n through `setupSdk()` from a Nuxt
 * plugin, which runs before any request is made.
 */
export type Notify = (message: string) => void

/** The subset of the app's auth store the SDK drives. */
export interface SdkAuth {
  accessToken: string
  $refresh: () => Promise<boolean>
  $clear: () => void
  $signOut: (silent?: boolean) => void
}

export interface SdkHooks {
  /** API origin resolved at request time (the app's `runtimeConfig.public.apiBase`). */
  baseURL?: () => string
  auth: () => SdkAuth
  /** i18n lookup for the backend's message keys; keys pass through untouched when absent. */
  t?: (key: string) => string
  notify?: { success: Notify; warn: Notify; error: Notify }
}

let hooks: SdkHooks | null = null

export function setupSdk(next: SdkHooks): void {
  hooks = next
}

function sdk(): SdkHooks {
  if (!hooks) throw new Error('[sdk] setupSdk() must run before any API call')
  return hooks
}

function t(key: string): string {
  return hooks?.t?.(key) ?? key
}

const retried = new Set<string>()
// Bound the retry bookkeeping: keys are per token+URL+params, so an unbounded
// set grows for the whole session and a stale key forces an early sign-out.
const RETRY_TTL = 30000
const MAX_RETRIED = 200

function retryKey(method: Method) {
  return `${method.type}:${method.url}:${JSON.stringify(method.config.params ?? {})}`
}

function markRetried(key: string) {
  if (retried.size >= MAX_RETRIED) retried.clear()
  retried.add(key)
  setTimeout(() => retried.delete(key), RETRY_TTL)
}

function ok(_response: Response, method: Method, data: unknown) {
  // Auth calls are not user-visible actions, and silent calls opt out.
  if (method.type !== 'GET' && !method.url.includes('/auth/') && !method.meta?.silent) hooks?.notify?.success(t('success'))
  return data
}

async function unauthorized(_response: Response, method: Method) {
  const auth = sdk().auth()
  // The refresh request itself was rejected: the session is gone. Clear the
  // in-memory token without firing another request (that recursed before).
  if (method.meta?.isRefresh) {
    auth.$clear()
    throw Object.assign(new Error('refresh failed'), { status: 401 })
  }
  // Logout and session probes must never trigger a refresh of their own.
  if (method.meta?.noRefresh) throw Object.assign(new Error('unauthorized'), { status: 401 })
  const failed = String(method.config.headers?.Authorization ?? '')
  // Another request already refreshed while this one was in flight: the failed
  // token is stale, so replaying once is enough.
  if (failed && failed !== 'Bearer ' + auth.accessToken) return method.send(true)
  const key = `${failed}:${retryKey(method)}`
  if (retried.has(key)) {
    auth.$signOut(true)
    throw Object.assign(new Error('retry failed'), { status: 401 })
  }
  if (!(await auth.$refresh())) {
    markRetried(key)
    auth.$signOut(true)
    throw Object.assign(new Error('refresh failed'), { status: 401 })
  }
  return method.send(true)
}

function fallback(response: Response, method: Method, v: unknown) {
  if (!method.meta?.silent && v && typeof v === 'object') {
    const { code, message } = v as { code?: string; message?: string }
    if (code) hooks?.notify?.warn(t(code))
    else if (message) hooks?.notify?.error(message)
  }
  // A real Error carrying the status, so callers branch on the HTTP code
  // instead of matching message strings.
  throw Object.assign(new Error(response.statusText), { status: response.status })
}

function network(err: unknown, method: Method) {
  // A rejected fetch (offline, DNS, blocked by CORS) never reaches onSuccess.
  const hasStatus = err instanceof Error && 'status' in err && typeof err.status === 'number'
  if (!method.meta?.silent && !hasStatus) hooks?.notify?.error(t('err.network'))
  return Promise.reject(err)
}

type FetchEvent = (response: Response, method: Method, data: unknown) => unknown

const events: Record<number, FetchEvent> = {}

events[200] = ok
// 401 is the only status with session semantics; everything else is reported.
events[401] = unauthorized
// `-1` is the documented "not a 2xx" slot used by the dispatch below.
events[-1] = fallback

/**
 * An empty config map: the generated `globals.d.ts` reads it as the source of
 * per-endpoint transform overrides, and no app defines any. Keeping the export
 * here (rather than in generated code) is what makes `index.ts` app-agnostic.
 */
export const $$userConfigMap = {}

// No explicit `Alova<AlovaGenerics>` annotation: the inferred generics
// (request config, response, VueHook states) are what the generated
// `globals.d.ts` reads back through `typeof`, and collapsing them to `any`
// makes every `useRequest` / `useWatcher` in the apps return `any`.
export const alovaInstance = createAlova({
  baseURL: '',
  cacheFor: null,
  statesHook: VueHook,
  // The refresh token is an HttpOnly cookie: without credentials the browser
  // neither stores nor sends it cross-origin (Tauri webview, dev server).
  requestAdapter: adapterFetch({ customFetch: (input, init) => fetch(input, { ...init, credentials: 'include' }) }),
  beforeRequest(method) {
    if (!method.baseURL) method.baseURL = hooks?.baseURL?.() ?? ''
    const { accessToken } = sdk().auth()
    if (accessToken) {
      method.config.headers = { ...method.config.headers, Authorization: 'Bearer ' + accessToken }
    }
  },
  responded: {
    async onSuccess(response, method) {
      const data = await fromData(response)
      // Unmapped 2xx statuses fall through to the success handler instead of the
      // error one: 201/204 are successes, not failures.
      const fn = events[response.status] ?? (response.ok ? ok : events[-1])
      if (fn) return fn(response, method, data)
      return data
    },
    onError(err, method) {
      // A rejected fetch (offline, DNS, blocked by CORS) never reaches
      // onSuccess; report it here so callers are not left with a silent failure.
      return network(err, method)
    },
  },
})

async function fromData(response: Response) {
  const ct = response.headers.get('Content-Type')
  if (response.status === 204 || !ct) return
  if (ct.includes('application/json')) return response.json()
  if (
    ct.includes('image/') ||
    ct.includes('audio/') ||
    ct.includes('video/') ||
    ct.includes('application/pdf') ||
    ct.includes('application/octet-stream')
  )
    return response.blob()
  if (ct.includes('application/zip') || ct.includes('arraybuffer')) return response.arrayBuffer()
  return response.text()
}
