import { Method } from 'alova'
import type { Alova, AlovaGenerics, AlovaMethodCreateConfig, MethodType } from 'alova'

/**
 * Build the `Apis` proxy from the generated `apiDefinitions` map.
 *
 * `$apis.<Tag>.<method>(config)` resolves to the `<METHOD> <url>` pair recorded
 * for `"<Tag>.<method>"`, fills the `{placeholder}`s of the url from
 * `config.pathParams`, and hands everything to alova as a `Method`.
 *
 * Proxies are cached per access path, so a repeated call reuses one instance
 * and alova's own identity checks (dedup, cache keys) keep working.
 */
export function createApis<T>(alovaInstance: Alova<AlovaGenerics>, definitions: Record<string, string[]>, configMap: Record<string, unknown>): T {
  const cache: Record<string, unknown> = Object.create(null)

  const createFunctionalProxy = (array: (string | symbol)[]): unknown => {
    const apiPathKey = array.join('.')
    if (cache[apiPathKey]) return cache[apiPathKey]

    const proxy = new Proxy(function () {}, {
      // Record the accessed property so nested paths keep accumulating.
      get(_, property) {
        return createFunctionalProxy([...array, property])
      },
      apply(_, __, [config]) {
        const apiItem = definitions[apiPathKey]
        if (!apiItem) throw new Error(`the api path of \`${apiPathKey}\` is not found`)

        const mergedConfig: Record<string, unknown> = { ...(configMap[apiPathKey] as object), ...(config as object) }
        const [method, url] = apiItem
        if (!method || !url) throw new Error(`the definition of \`${apiPathKey}\` is malformed`)
        const pathParams = mergedConfig.pathParams as Record<string, string | number> | undefined
        const urlReplaced = url.replace(/\{([^}]+)\}/g, (_, key: string) => {
          const pathParam = pathParams?.[key]
          // A missing parameter would otherwise request a literal `{id}` path.
          if (pathParam === undefined) throw new Error(`[\`${apiPathKey}\`] missing path parameter \`${key}\``)
          return String(pathParam)
        })
        delete mergedConfig.pathParams

        let data = mergedConfig.data
        // A plain object body becomes multipart as soon as it carries a Blob, so
        // file uploads work through the same `{ file, kind }` call shape.
        if (Object.prototype.toString.call(data) === '[object Object]' && typeof FormData !== 'undefined') {
          let hasBlobData = false
          const formData = new FormData()
          for (const [key, value] of Object.entries(data as Record<string, unknown>)) {
            for (const ele of Array.isArray(value) ? value : [value]) {
              formData.append(key, ele as string | Blob)
              if (ele instanceof Blob) hasBlobData = true
            }
          }
          data = hasBlobData ? formData : data
        }

        return new Method(
          method.toUpperCase() as MethodType,
          alovaInstance,
          urlReplaced,
          mergedConfig as AlovaMethodCreateConfig<AlovaGenerics, any, any>,
          data as string | FormData | Blob | ArrayBuffer | URLSearchParams | Record<string, any> | undefined,
        )
      },
    })

    cache[apiPathKey] = proxy
    return proxy
  }

  return new Proxy({}, { get: (_, property) => createFunctionalProxy([property]) }) as T
}
