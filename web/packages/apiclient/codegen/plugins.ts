import { createPlugin } from '@alova/wormhole/plugin'
import type { OpenAPIDocument, SchemaObject } from '@alova/wormhole'
import fs from 'node:fs'
import path from 'node:path'

const METHODS = ['get', 'post', 'put', 'delete', 'patch', 'head', 'options'] as const

// Reserved words and literals that cannot be emitted as a method name verbatim.
const RESERVED_WORDS = new Set([
  'break',
  'case',
  'catch',
  'class',
  'const',
  'continue',
  'debugger',
  'default',
  'delete',
  'do',
  'else',
  'enum',
  'export',
  'extends',
  'false',
  'finally',
  'for',
  'function',
  'if',
  'import',
  'in',
  'instanceof',
  'new',
  'null',
  'return',
  'super',
  'switch',
  'this',
  'throw',
  'true',
  'try',
  'typeof',
  'var',
  'void',
  'while',
  'with',
  'yield',
  'let',
  'static',
  'await',
  'implements',
  'interface',
  'package',
  'private',
  'protected',
  'public',
  'arguments',
  'eval',
  'undefined',
])

const isSchema = (s: unknown): s is SchemaObject => !!s && typeof s === 'object' && !Array.isArray(s)

const getRefName = (ref: string) => ref.split('/').pop() ?? ''

/**
 * A schema that can legitimately be `null` in a response. swag marks pointer
 * fields with the `x-nullable` extension; wormhole turns OpenAPI `nullable`
 * and a `null` union member into a null type, so all three shapes count.
 */
const isNullable = (s: unknown): boolean => {
  if (!isSchema(s)) return false
  const schema = s as SchemaObject & { 'nullable'?: boolean; 'x-nullable'?: boolean; 'anyOf'?: unknown[]; 'oneOf'?: unknown[] }
  if (schema['x-nullable'] === true || schema.nullable === true) return true
  if (Array.isArray(schema.type) && schema.type.includes('null')) return true
  for (const key of ['anyOf', 'oneOf'] as const) {
    if (schema[key]?.some((m) => isSchema(m) && (m as { type?: unknown }).type === 'null')) return true
  }
  return false
}

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
  if (!schema) return 'null'
  // A nullable response field must default to null, not to the type's zero
  // value: `googleEmail: ''` is a real address-shaped lie.
  if (isNullable(schema)) return 'null'
  const comp = schema as SchemaObject & { allOf?: any; oneOf?: any; anyOf?: any }
  // swag sometimes wraps a plain $ref in `allOf: [{$ref}]`; resolve it to the
  // referenced model factory so nested objects get real default values.
  const aloneRef = (comp.allOf ?? comp.oneOf ?? comp.anyOf)?.[0]?.$ref as string | undefined
  const aloneName = aloneRef ? getRefName(aloneRef) : undefined
  if (isSchema(schema) && '$ref' in schema) {
    const refName = getRefName((schema as any).$ref)
    // A pruned $ref (the model is not emitted for this app) has no factory.
    return ctx.schemas[refName] ? refName : 'null'
  }
  if (aloneName && ctx.schemas[aloneName]) return aloneName
  if (!isSchema(schema)) return 'null'
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
        const stripped = prefix && op.operationId.startsWith(prefix) ? op.operationId.slice(prefix.length) : op.operationId

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
        // `delete` reads better as `remove`; every other JS keyword gets a
        // trailing underscore (wormhole would otherwise suffix it with `_1`).
        op.operationId = id === 'delete' ? 'remove' : RESERVED_WORDS.has(id) ? id + '_' : id
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
        // Iteration order must not decide which shape survives: two schemas
        // and one generated name is a duplicate `@name` in the swag
        // annotations, and silently keeping one produces wrong types.
        throw new Error(`[stripSchemaPrefix] "${from}" and an existing "${to}" both map to "${to}"`)
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
function collectUsedSchemas(document: OpenAPIDocument, prefixes: string[], options: SchemaWalkerOptions = {}): Set<string> {
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
    if ((s as any).additionalProperties && typeof (s as any).additionalProperties === 'object') visit((s as any).additionalProperties)
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
      // x-nullable is swag's spelling; wormhole only understands `nullable`.
      for (const prop of Object.values(s.properties) as SchemaObject[]) {
        if (isSchema(prop) && (prop as { 'x-nullable'?: boolean })['x-nullable'] === true) {
          ;(prop as { nullable?: boolean }).nullable = true
        }
      }
      // A field that may be null is present-but-null, so it cannot be a
      // required non-null property in the generated type.
      s.required = Object.keys(s.properties).filter((k) => !isNullable(s.properties[k]))
    }
  },
}))

// Collect $ref targets referenced by a schema. Used to order model factories
// so dependencies are declared before their consumers (alphabetical order
// breaks `const` block-scoped references like `Config.oauth: OauthConfig`).
function schemaDeps(name: string, schemas: Record<string, SchemaObject>): string[] {
  const deps = new Set<string>()
  const visit = (v: unknown) => {
    if (!isSchema(v)) return
    if (isSchema(v) && '$ref' in v) {
      const refName = getRefName((v as any).$ref)
      if (schemas[refName]) deps.add(refName)
      return
    }
    for (const key of ['allOf', 'oneOf', 'anyOf'] as const) (v as any)[key]?.forEach(visit)
    if ((v as any).items) visit((v as any).items)
    for (const p of Object.values((v as any).properties ?? {})) visit(p)
  }
  visit(schemas[name])
  return [...deps]
}

/**
 * Generate `defaults.ts` with const model factories referencing the generated
 * interfaces, so form models / local storage / table crud keep working.
 * Only schemas used by this app's routes (matching `prefixes`) are generated,
 * mirroring the pruned `globals.d.ts`.
 */
export const modelDefaults = createPlugin((outputDir: string, prefixes: string[]) => {
  const outFile = path.join(outputDir, 'models.ts')
  // The text is built while the document is in memory but written only after
  // the generator has written every other file: a failed generation must leave
  // the previous models.ts untouched instead of half-updated.
  let pending: string | null = null
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

      // Depth-first, dependency-before-consumer emission order.
      const byName = new Map(entries)
      const state = new Map<string, number>() // 0=unseen 1=visiting 2=done
      const order: string[] = []
      const visit = (name: string) => {
        const st = state.get(name) ?? 0
        if (st === 2 || st === 1) return // done or cycle-guard
        state.set(name, 1)
        for (const dep of schemaDeps(name, schemas).sort()) {
          if (byName.has(dep)) visit(dep)
        }
        state.set(name, 2)
        order.push(name)
      }
      for (const [name] of entries.slice().sort(([a], [b]) => a.localeCompare(b))) visit(name)

      const body = order
        .map((name) => {
          const s = schemas[name] as SchemaObject
          const props = s.properties ?? {}
          const lines = Object.entries(props).map(([k, v]) => `  ${k}: ${defaultExpr(k, v as SchemaObject, ctx, 1)}`)
          return `export type ${name} = G.${name}\nexport const ${name}: ${name} = {\n${lines.join(',\n')}\n}`
        })
        .join('\n\n')

      pending = `/* tslint:disable */\n/* eslint-disable */\nimport type * as G from './globals'\n\n${body}\n`
    },
    afterCodeGenerate(error?: Error) {
      if (error || !pending) return
      const code = pending
      pending = null
      fs.mkdirSync(path.dirname(outFile), { recursive: true })
      fs.writeFileSync(outFile, code)
    },
  }
})

/**
 * Turn the generated `declare global { interface <G> {...} var <G>: <G> }`
 * block of `globals.d.ts` into a module-local `export interface Apis`.
 *
 * wormhole declares the API surface globally and mounts the proxy on
 * `globalThis`; the SDK package imports it by name from `./globals` instead,
 * which is what keeps `api.ts` / `index.ts` importable without a global.
 */
export const exposeApis = createPlugin((globalName = 'Apis') => ({
  async beforeCodeGenerate(_data: unknown, outputFile: string, ctx: { renderTemplate: () => Promise<string> }): Promise<string | undefined> {
    if (!outputFile.endsWith('globals.d.ts')) return
    const code = await ctx.renderTemplate()

    const marker = 'declare global {'
    const start = code.indexOf(marker)
    if (start === -1) throw new Error('[exposeApis] declare global block not found in globals.d.ts')

    const lines = code.slice(start).trimEnd().split('\n')
    const end = lines.length - 1
    if (lines[0].trim() !== marker) throw new Error('[exposeApis] malformed declare global block in globals.d.ts')
    if (lines[end].trim() !== '}') throw new Error('[exposeApis] unterminated declare global block in globals.d.ts')

    const body = lines
      .slice(1, end)
      // the global `var` mount is what the package replaces with its own export
      .filter((line) => !/^\s*var\s+\w+\s*:\s*\w+\s*;?\s*$/.test(line))
      .map((line) => (line.startsWith('  ') ? line.slice(2) : line))
      .join('\n')
      .replace(/^interface\s+\w+\s*\{/m, 'export interface Apis {')

    const out = `${code.slice(0, start).trimEnd()}\n\n${body.trimEnd()}\n`
    if (!out.includes('export interface Apis')) {
      throw new Error('[exposeApis] could not rewrite the declare global block into "export interface Apis"')
    }
    return out
  },
}))
