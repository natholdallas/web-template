import { createPlugin } from '@alova/wormhole/plugin'
import type { OpenAPIDocument, SchemaObject } from '@alova/wormhole'
import fs from 'node:fs'
import path from 'node:path'

const METHODS = ['get', 'post', 'put', 'delete', 'patch', 'head', 'options'] as const

const isSchema = (s: unknown): s is SchemaObject => !!s && typeof s === 'object' && !Array.isArray(s)

const getRefName = (ref: string) => ref.split('/').pop() ?? ''

function primitiveDefault(t: unknown): string {
  if (t === 'object') return '{}'
  if (t === 'array') return '[]'
  if (t === 'boolean') return 'false'
  if (t === 'number' || t === 'integer') return '0'
  return "''"
}

type Ctx = {
  pad: (n: number) => string
  schemas: Record<string, SchemaObject>
}

/**
 * Build a runtime default expression for a field/root schema.
 * Mirrors the previous `const X = { field: default }` SDK style so
 * `inst(model)` / `defineModel({ default: model })` keep working as-is.
 */
function defaultExpr(field: string, schema: SchemaObject | undefined, ctx: Ctx, indent: number): string {
  const pad = ctx.pad(indent)
  void field
  if (!schema) return 'undefined'
  if (isSchema(schema) && '$ref' in schema) {
    const refName = getRefName((schema as any).$ref)
    return ctx.schemas[refName] ? refName : 'undefined'
  }
  if (!isSchema(schema)) return 'undefined'
  if (schema.default !== undefined) return JSON.stringify(schema.default)
  if (schema.enum?.length) return JSON.stringify(schema.enum[0])

  const eff = Array.isArray(schema.type) ? schema.type.find((t) => t !== 'null') : schema.type

  if (eff === 'object' && schema.properties && Object.keys(schema.properties).length > 0) {
    const body = Object.entries(schema.properties)
      .map(([k, v]) => `${pad}  ${k}: ${defaultExpr(k, v as SchemaObject, ctx, indent + 1)}`)
      .join(',\n')
    return `{${body ? `\n${body}\n${pad}` : ''}}`
  }
  if (eff === 'array') return '[]'
  return primitiveDefault(eff)
}

/**
 * Strip the per-service operationId prefix so generated method names are clean
 * and camelCased (e.g. `admSignIn` / `usrSignIn` -> `signIn` in each app).
 *
 * Also strips the entity carried in the operationId when it matches the group's
 * tag, yielding concise methods: `admListUsers` -> `listUsers` -> `list` under
 * `Apis.User`, `admCreateAdmin` -> `createAdmin` -> `create` under `Apis.Admin`.
 *
 * Because the tag already names the entity, the verb-only method stays unique
 * within its group. Since the result is a valid JS identifier, wormhole keeps
 * it as-is (its operationIdSet dedup only kicks in for non-identifier ids), so
 * two groups may both have `list` (`Apis.User.list`, `Apis.Admin.list`).
 */
export const stripOperationIdPrefix = createPlugin((prefix: string) => ({
  afterOpenapiParse(document: OpenAPIDocument) {
    const lower = (id: string) => id.replace(/^(\w)/, (_m, c: string) => c.toLowerCase())
    for (const [, item] of Object.entries(document.paths ?? {})) {
      for (const method of METHODS) {
        const op = item?.[method]
        if (!op || typeof op.operationId !== 'string') continue
        const stripped =
          prefix && op.operationId.startsWith(prefix) ? op.operationId.slice(prefix.length) : op.operationId

        // derive the tag (post app-prefix) to find the entity to drop
        let tag = (Array.isArray(op.tags) ? op.tags[0] : '') as string
        if (prefix && tag.startsWith(prefix)) tag = tag.slice(prefix.length)

        let id = lower(stripped)
        if (tag) {
          const suffixes = [tag, tag + 's', tag + 'es'].filter((x) => x.length > 0)
          for (const suf of suffixes.sort((a, b) => b.length - a.length)) {
            if (id.length > suf.length && id.toLowerCase().endsWith(suf.toLowerCase())) {
              id = id.slice(0, id.length - suf.length)
              break
            }
          }
        }
        // JS reserved words can't be method names (wormhole would emit
        // `delete_`/`delete_1`); give them a clean alias instead.
        const reserved: Record<string, string> = { delete: 'remove' }
        op.operationId = reserved[id] ?? id
      }
    }
  },
}))

// Tags are also app-scoped in the single merged doc (`usrUser`, `admAdmin`, ...)
// but should surface as clean groups in the generated SDK (`Apis.User`, ...).
// Unlike operationIds, tags keep their original casing after stripping.
export const stripTagPrefix = createPlugin((prefix: string) => ({
  afterOpenapiParse(document: OpenAPIDocument) {
    for (const [, item] of Object.entries(document.paths ?? {})) {
      for (const method of METHODS) {
        const op = item?.[method]
        if (!op || !Array.isArray(op.tags)) continue
        op.tags = op.tags.map((t) => (prefix && t.startsWith(prefix) ? t.slice(prefix.length) : t))
      }
    }
  },
}))

// Deep-rewrite every schema `$ref` (components.schemas, parameters, request
// bodies, responses, nested properties, etc.) after a rename.
function rewriteSchemaRefs(node: any, map: Record<string, string>) {
  if (Array.isArray(node)) {
    node.forEach((n) => rewriteSchemaRefs(n, map))
    return
  }
  if (node && typeof node === 'object') {
    if (typeof node.$ref === 'string' && node.$ref.startsWith('#/components/schemas/')) {
      const name = node.$ref.split('/').pop()
      if (map[name]) node.$ref = '#/components/schemas/' + map[name]
    }
    for (const v of Object.values(node)) rewriteSchemaRefs(v, map)
  }
}

/**
 * Strip a per-app prefix from OpenAPI schema (definition) names.
 *
 * `swag init` requires type names to be globally-unique across the SINGLE
 * backend swagger doc, so every app-owned DTO carries a `Usr`/`Adm` prefix in
 * `@name`. Each frontend generates its OWN sdk/globals, so there is no real
 * collision on the frontend — this plugin removes the prefix and rewrites all
 * refs, yielding clean names (`UsrUser -> User`, `AdmAuthIn -> AuthIn`, ...).
 *
 * Only names whose stripped remainder starts with an UPPERCASE letter are
 * touched, so an accidental prefix match against an unprefixed global model is
 * left alone (e.g. `Adm` is a prefix of the shared `Admin` model, but stripping
 * it would yield lowercase `in` — so `Admin` is preserved).
 *
 * MUST run before `modelDefaults`, so that the model file/type refs line up.
 */
export const stripSchemaPrefix = createPlugin((prefix: string) => ({
  afterOpenapiParse(document: OpenAPIDocument) {
    const schemas = document.components?.schemas
    if (!schemas || !prefix) return

    const map: Record<string, string> = {}
    for (const name of Object.keys(schemas)) {
      if (!name.startsWith(prefix)) continue
      const to = name.slice(prefix.length)
      // only strip real type names (PascalCase remainder); skip coincidental
      // prefix matches on lowercase remainders like `Admin` -> `in`.
      if (!to || to[0] === to[0].toLowerCase()) continue
      map[name] = to
    }

    for (const [from, to] of Object.entries(map)) {
      if (schemas[to] && schemas[to] !== schemas[from]) {
        // a prefixed app DTO shadows an unprefixed global of the same name;
        // the global is unused by this app's routes and gets pruned anyway.
        console.warn(`[stripSchemaPrefix] "${to}" replaced by prefixed "${from}"`)
      }
      schemas[to] = schemas[from]
      delete schemas[from]
    }

    rewriteSchemaRefs(document, map)
  },
}))

type SchemaWalkerOptions = { responses?: boolean; requestBody?: boolean; parameters?: boolean }

// Collect all schema names transitively referenced by this app's routes.
// `responses` only walks 2xx responses, `requestBody` only request bodies,
// `parameters` only (path/query) parameter schemas.
function collectUsedSchemas(
  document: OpenAPIDocument,
  prefixes: string[],
  options: SchemaWalkerOptions = {},
): Set<string> {
  const schemas = (document.components?.schemas ?? {}) as Record<string, SchemaObject>
  const used = new Set<string>()

  const visit = (s: unknown) => {
    if (!isSchema(s)) return
    if (isSchema(s) && '$ref' in s) {
      const name = getRefName((s as any).$ref)
      if (!used.has(name) && schemas[name]) {
        used.add(name)
        visit(schemas[name])
      }
      return
    }
    for (const key of ['allOf', 'oneOf', 'anyOf'] as const) (s as any)[key]?.forEach(visit)
    if ((s as any).items) visit((s as any).items)
    for (const p of Object.values((s as any).properties ?? {})) visit(p)
    if ((s as any).additionalProperties && typeof (s as any).additionalProperties === 'object')
      visit((s as any).additionalProperties)
  }

  const requestBodies = (document.components?.requestBodies ?? {}) as Record<string, any>

  const resolveRefName = (s: any): string | undefined => {
    if (s && typeof s === 'object' && !Array.isArray(s) && '$ref' in s && typeof s.$ref === 'string') {
      return s.$ref.split('/').pop()
    }
    return undefined
  }

  const walk = {
    parameters: options.parameters !== false,
    requestBody: options.requestBody !== false,
    responses: options.responses !== false,
  }

  for (const [path, item] of Object.entries(document.paths ?? {})) {
    if (!prefixes.some((pre) => path.startsWith(pre))) continue
    for (const method of METHODS) {
      const op = item?.[method]
      if (!op) continue
      if (walk.parameters) for (const param of op.parameters ?? []) visit((param as any).schema)
      if (walk.requestBody) {
        const rb = (op.requestBody as any) ?? undefined
        if (rb) {
          const rbName = resolveRefName(rb)
          const rbObj = rbName ? requestBodies[rbName] : rb
          for (const media of Object.values(rbObj?.content ?? {})) visit((media as any).schema)
        }
      }
      if (walk.responses) {
        for (const [status, r] of Object.entries(op.responses ?? {})) {
          if (!/^2/.test(status) || !r) continue
          for (const media of Object.values((r as any).content ?? {})) visit((media as any).schema)
        }
      }
    }
  }
  return used
}

/**
 * swag only emits `required` for REQUEST bodies, never for responses. So every
 * response model lands in the SDK as fully optional (`field?: number`) even
 * though the backend always returns the full object. This plugin marks every
 * property of schemas used in 2xx responses as required — request bodies keep
 * swag's own required list, and error schemas (e.g. `Fail` on 4xx) stay
 * optional. MUST run after `renameSchemas`.
 */
export const requiredResponses = createPlugin((prefixes: string[]) => ({
  afterOpenapiParse(document: OpenAPIDocument) {
    const schemas = (document.components?.schemas ?? {}) as Record<string, SchemaObject>
    const used = collectUsedSchemas(document, prefixes, { responses: true, requestBody: false, parameters: false })
    for (const name of used) {
      const s = schemas[name]
      if (!isSchema(s) || s.type !== 'object' || !s.properties) continue
      s.required = Object.keys(s.properties)
    }
  },
}))

/**
 * Generate `defaults.ts` with const model factories referencing the generated
 * interfaces, so form models / local storage / table crud keep working.
 * Only schemas used by this app's routes (matching `prefixes`) are generated,
 * mirroring the pruned `globals.d.ts`.
 */
export const modelDefaults = createPlugin((outputDir: string, prefixes: string[]) => {
  const outFile = path.join(outputDir, 'models.ts')
  return {
    afterOpenapiParse(document: OpenAPIDocument) {
      const used = collectUsedSchemas(document, prefixes)
      const schemas = (document.components?.schemas ?? {}) as Record<string, SchemaObject>
      const entries = Object.entries(schemas).filter(([n, s]) => used.has(n) && isSchema(s) && !s.enum)
      if (entries.length === 0) return

      const ctx: Ctx = {
        schemas,
        pad: (n) => '  '.repeat(n),
      }

      const body = entries
        .sort(([a], [b]) => a.localeCompare(b))
        .map(([name, schema]) => {
          const s = schema as SchemaObject
          const props = s.properties ?? {}
          const lines = Object.entries(props).map(([k, v]) => `  ${k}: ${defaultExpr(k, v as SchemaObject, ctx, 1)}`)
          return `export type ${name} = G.${name}\nexport const ${name}: ${name} = {\n${lines.join(',\n')}\n}`
        })
        .join('\n\n')

      const code = `/* tslint:disable */\n/* eslint-disable */\nimport type * as G from './globals'\n\n${body}\n`
      fs.mkdirSync(path.dirname(outFile), { recursive: true })
      fs.writeFileSync(outFile, code)
    },
  }
})

/**
 * Override the default index.ts with an app-agnostic runtime: apiBase from
 * `useRuntimeConfig`, Bearer token from `useAuth()`, and a minimal event
 * registry (`Api.NewEvent`) so each app wires its own toast/session handlers.
 */
export const customIndex = createPlugin(() => ({
  async beforeCodeGenerate(_data: unknown, outputFile: string) {
    if (!outputFile.endsWith('index.ts')) return
    return `/* eslint-disable */
import { createAlova } from 'alova'
import type { Method } from 'alova'
import adapterFetch from 'alova/fetch'
import VueHook from 'alova/vue'
import { createApis, mountApis, withConfigType } from './createApis'

type FetchEvent = (response: Response, method: Method, data: unknown) => void
const events: Record<number, FetchEvent> = {}

export const alovaInstance = createAlova({
  baseURL: useRuntimeConfig().public.apiBase,
  cacheFor: null,
  statesHook: VueHook,
  requestAdapter: adapterFetch(),
  beforeRequest(method) {
    const { token } = useAuth()
    if (token) {
      method.config.headers = { ...method.config.headers, Authorization: 'Bearer ' + token }
    }
  },
  responded: {
    async onSuccess(response, method) {
      const data = await fromData(response)
      const fn = events[response.status] ?? events[-1]
      if (fn) fn(response, method, data)
      return data
    },
  },
})

export const Api = {
  NewEvent(status: number, fn: FetchEvent) {
    events[status] = fn
  },
}

export const $$userConfigMap = withConfigType({})

const Apis = createApis(alovaInstance, $$userConfigMap)

mountApis(Apis)

export { Apis }
export default Apis

async function fromData(response: Response) {
  const ct = response.headers.get('Content-Type')
  if (response.status === 204 || !ct) return
  if (ct.includes('application/json')) return response.json()
  if (ct.includes('image/') || ct.includes('audio/') || ct.includes('video/') || ct.includes('application/pdf') || ct.includes('application/octet-stream')) return response.blob()
  if (ct.includes('application/zip') || ct.includes('arraybuffer')) return response.arrayBuffer()
  return response.text()
}
`
  },
}))
