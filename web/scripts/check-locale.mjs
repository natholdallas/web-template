#!/usr/bin/env node
/**
 * Locale consistency checker for every frontend app under web/apps.
 *
 * For each app (auto-discovered, so a new app needs no script change):
 *   1. compares the key set across the app's language files (app/locale/*.ts)
 *      and reports keys MISSING from a language (reference: zh_cn when present,
 *      else the first file alphabetically);
 *   2. reports keys that are EXTRA in a language;
 *   3. checks the KEY ORDER of each language against the reference;
 *   4. reports UNUSED keys: keys defined in the app's locale files that no app
 *      source file references (direct `$t('key')`, a quoted constant used via
 *      `$t(obj.key)`, or a dynamic `\`prefix.${x}\`` skeleton);
 *   5. reports keys the source references but no language defines (shared
 *      @natholdallas/i18n base keys count as defined);
 *   6. reports TEXT passed where Vuetify 4 expects a TRANSLATION KEY
 *      (`:no-data-text="$t('no.data')"`): Vuetify resolves those props through
 *      the app's vue-i18n instance, so the translated value becomes the key and
 *      logs `Not found '暂无数据' key in 'zh' locale messages`.
 *
 * Only the app's OWN keys are inspected: locale files spread the shared base
 * dict from @natholdallas/i18n (the submodule), which is out of scope here.
 *
 * Usage: node web/scripts/check-locale.mjs [app...] [--delete-unused]
 *        (no app args checks every app; pass names like `adm usr` to narrow;
 *         --delete-unused removes the unused keys from the app's locale files)
 */
import { readdirSync, readFileSync, statSync, writeFileSync } from 'node:fs'
import { dirname, join, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const args = process.argv.slice(2)
const deleteUnused = args.includes('--delete-unused')
const appArgs = args.filter((a) => !a.startsWith('--'))

const here = dirname(fileURLToPath(import.meta.url))
const repo = resolve(here, '..', '..')
const appsDir = join(repo, 'web', 'apps')
// Every app in this template follows the Nuxt `restructureDir: 'app'` layout.
const localeRel = ['app', 'locale']
const sourceRoots = ['app']
const baseLocaleDir = join(repo, 'web', 'packages', 'natholdallas', 'i18n', 'locale')

const ignoreDirNames = new Set(['node_modules', '.nuxt', '.output', 'dist', 'sdk'])
const skipRelPrefix = [join(...localeRel), join('app', 'lib', 'sdk')]
// Keys may also be referenced from the Go backend (e.g. `err.*` codes returned
// by the API and rendered client-side via `t(code)` in the generated SDK
// wiring), so the backend source is scanned too. The hand-written SDK runtime
// is shared by every app, so keys it translates count as used everywhere.
const extraRoots = [
  { dir: join(repo, 'internal'), ext: /\.go$/ },
  { dir: join(repo, 'web', 'packages', 'sdk', 'src'), ext: /\.ts$/ },
]

const SOURCE_EXT_RE = /\.(vue|ts|tsx|js|jsx|mjs|json)$/

// Extract the app's own translation keys from a locale file: the entries of the
// `export default { ... }` object, skipping spread (`...zh`) and shorthand
// members (`$vuetify`).
function parseLocaleKeys(content) {
  const keys = []
  let inDefault = false
  for (const raw of content.split('\n')) {
    const line = raw.trim()
    if (!inDefault) {
      if (line.startsWith('export default {')) inDefault = true
      continue
    }
    if (line === '}' || line === '};') break
    if (!line || line.startsWith('//') || line.startsWith('...') || line.startsWith('$')) continue
    const m = line.match(/^('([^']*)'|"([^"]*)"|([\w.$-]+))\s*:/)
    if (m) keys.push(m[2] ?? m[3] ?? m[4])
  }
  return keys
}

function* walkFiles(dir, base = dir, extRe = SOURCE_EXT_RE) {
  let entries
  try {
    entries = readdirSync(dir, { withFileTypes: true })
  } catch {
    return
  }
  for (const entry of entries) {
    if (entry.name.startsWith('.') || ignoreDirNames.has(entry.name)) continue
    const p = join(dir, entry.name)
    if (entry.isDirectory()) {
      yield* walkFiles(p, base, extRe)
    } else if (extRe.test(entry.name)) {
      yield p
    }
  }
}

// The app's own source only (locale files excluded): keys here are "implied"
// as translation keys.
function appFiles(appDir) {
  const files = []
  for (const root of sourceRoots) {
    const dir = join(appDir, root)
    if (!statSync(dir, { throwIfNoEntry: false })) continue
    for (const f of walkFiles(dir)) {
      const rel = relative(appDir, f)
      if (skipRelPrefix.some((p) => rel.startsWith(p))) continue
      files.push(f)
    }
  }
  return files
}

// Everything that may reference a key: the app plus the shared Go backend and
// SDK runtime.
function sourceFiles(appDir) {
  const files = appFiles(appDir)
  for (const { dir, ext } of extraRoots) {
    if (!statSync(dir, { throwIfNoEntry: false })) continue
    for (const f of walkFiles(dir, dir, ext)) files.push(f)
  }
  return files
}

function readText(f) {
  try {
    return readFileSync(f, 'utf8')
  } catch {
    return null
  }
}

// Static prefixes used to build keys dynamically: `$t(\`faq.${cat}.${idx}.q\`)`,
// `labelKey: \`script.modules.${key}\``, or `$t('script.modules.' + m)`. Any key
// starting with one of these prefixes is considered used, since its full key is
// assembled at runtime.
function extractDynamicPrefixes(content) {
  const prefixes = new Set()
  for (const m of content.matchAll(/`([^`$]*)\$\{/g)) {
    if (m[1]) prefixes.add(m[1])
  }
  for (const m of content.matchAll(/(?:\$i18n\.|\bi18n\.)?(?:\$t|\bt)\s*\(\s*['"]([^'"]+)['"]\s*\+/g)) {
    if (m[1]) prefixes.add(m[1])
  }
  return prefixes
}

// Concrete keys the source references: `$t('key')` / `t('key')` literals, the
// values of properties that conventionally hold keys (`k:`, `key:`, `label:`,
// `title:`, `desc:`, ...) and keys implied by template-literal skeletons plus
// `k:` constants. Only literals that already exist in SOME locale of the
// project are added from the property form, so ordinary strings like
// `label: 'primary'` cannot create false positives.
function collectImpliedKeys(appDir, knownKeys) {
  const constKeys = new Set()
  const direct = new Set()
  const props = new Set()
  const skeletons = []
  for (const f of appFiles(appDir)) {
    const content = readText(f)
    if (!content) continue
    for (const m of content.matchAll(/\bk:\s*'([^']+)'/g)) constKeys.add(m[1])
    for (const m of content.matchAll(/(?:\$i18n\.|\bi18n\.)?(?:\$t|\bt)\s*\(\s*'([^']+)'\s*\)/g)) direct.add(m[1])
    for (const m of content.matchAll(/(?:\$i18n\.|\bi18n\.)?(?:\$t|\bt)\s*\(\s*"([^"]+)"\s*\)/g)) direct.add(m[1])
    for (const m of content.matchAll(/\b(?:k|key|label|title|desc|description|placeholder|text)\s*:\s*['"]([^'"]+)['"]/g)) {
      if (knownKeys.has(m[1])) props.add(m[1])
    }
    for (const m of content.matchAll(/`([^`$]*)\$\{[^}]*\}([^`$]*)`/g)) {
      if (m[1] && (m[1].includes('.') || m[2].includes('.'))) {
        skeletons.push({ prefix: m[1], suffix: m[2] })
      }
    }
  }
  const implied = new Set([...direct, ...props])
  for (const { prefix, suffix } of skeletons) {
    for (const k of constKeys) {
      if (k.startsWith(prefix)) {
        implied.add(k)
        if (suffix) implied.add(k + suffix)
      }
    }
  }
  return implied
}

// Whether a key is "used": quoted somewhere in the app source (covers
// `$t('key')`, `t('key')` and `k: 'key'` / `label: 'key'` maps that feed dynamic
// `$t(x.k)` lookups), or matching a dynamic prefix.
function scanUsage(appDir, keys) {
  const used = new Set()
  const prefixes = new Set()
  for (const f of sourceFiles(appDir)) {
    const content = readText(f)
    if (!content) continue
    for (const p of extractDynamicPrefixes(content)) prefixes.add(p)
    for (const k of keys) {
      if (used.has(k)) continue
      if (content.includes(`'${k}'`) || content.includes(`"${k}"`)) used.add(k)
    }
  }
  for (const k of keys) {
    if (used.has(k)) continue
    for (const p of prefixes) {
      if (k.startsWith(p)) {
        used.add(k)
        break
      }
    }
  }
  return used
}

// Vuetify 4 props whose value is a TRANSLATION KEY resolved through the app's
// vue-i18n instance (their default is `'$vuetify.<key>'`): `noDataText`,
// `loadingText`, `itemsPerPageText`, ... Discovered from the installed Vuetify
// lib so the list cannot drift, and scoped to the owning component family so an
// ambiguous name such as `title`/`label` is only checked on the components that
// really translate it (e.g. `VFileUpload`, `VBadge`) — not on `VBtn`/`VListItem`.
function vuetifyKeyProps() {
  const families = new Map()
  const lib = join(repo, 'web', 'node_modules', 'vuetify', 'lib')
  if (!statSync(lib, { throwIfNoEntry: false })) return families
  for (const f of walkFiles(lib, lib, /\.js$/)) {
    const content = readText(f)
    if (!content || (!content.includes("'$vuetify.") && !content.includes('"$vuetify.'))) continue
    const family = relative(lib, f)
      .split(/[\\/]/)
      .find((p) => /^V[A-Z]/.test(p))
    if (!family) continue
    let props = families.get(family)
    if (!props) families.set(family, (props = new Map()))
    for (const m of content.matchAll(/([\w$]+)\s*:\s*\{[^{}]*?default:\s*['"]\$vuetify\.([^'"]*)/gs)) {
      props.set(m[1], m[2])
    }
  }
  return families
}

function kebab(name) {
  return name.replace(/[A-Z]/g, (c) => `-${c.toLowerCase()}`)
}

// `<VDataTable ... :no-data-text="$t('no.data')">`: Vuetify resolves that prop
// through vue-i18n, so the translated value becomes the key.
const TAG_RE = /<([A-Za-z][\w.-]*)\b((?:"[^"]*"|'[^']*'|[^>"'])*)>/g

function scanDoubleTranslation(appDir, families) {
  const findings = []
  if (!families.size) return findings
  const scopes = [...families].map(([tag, props]) => ({ tag, kebabTag: kebab(tag), props }))
  for (const f of appFiles(appDir)) {
    if (!f.endsWith('.vue')) continue
    const content = readText(f)
    if (!content) continue
    for (const tag of content.matchAll(TAG_RE)) {
      const name = tag[1]
      const scope = scopes.find((s) => name === s.tag || name === s.kebabTag || name.startsWith(s.tag))
      if (!scope) continue
      for (const prop of scope.props.keys()) {
        const kprop = kebab(prop)
        const m = tag[2].match(new RegExp(`(?:^|\\s)(?::|v-bind:)${kprop}\\s*=\\s*"([^"]*)"`))
        if (!m) continue
        if (!/(?:\$i18n\.|\bi18n\.)?(?:\$t|\bt)\s*\(/.test(m[1])) continue
        findings.push({
          file: relative(repo, f),
          line: lineOf(content, tag.index),
          prop: kprop,
          value: m[1].trim(),
        })
      }
    }
  }
  return findings
}

function lineOf(content, index) {
  let line = 1
  for (let i = 0; i < index; i++) if (content[i] === '\n') line++
  return line
}

function firstDiff(a, b) {
  const len = Math.max(a.length, b.length)
  for (let i = 0; i < len; i++) if (a[i] !== b[i]) return i
  return -1
}

function readLocaleDir(dir) {
  const dict = {}
  if (!statSync(dir, { throwIfNoEntry: false })) return dict
  const files = readdirSync(dir)
    .filter((f) => f.endsWith('.ts') && !f.endsWith('.d.ts'))
    .sort()
  for (const f of files) dict[f.replace(/\.ts$/, '')] = parseLocaleKeys(readFileSync(join(dir, f), 'utf8'))
  return dict
}

let problems = 0
const apps = appArgs.length > 0 ? appArgs : readdirSync(appsDir).filter((n) => statSync(join(appsDir, n)).isDirectory())
const baseDict = readLocaleDir(baseLocaleDir)
const knownKeys = new Set(Object.values(baseDict).flat())
const keyProps = vuetifyKeyProps()

for (const app of apps) {
  const appDir = join(appsDir, app)
  const localeDir = join(appDir, ...localeRel)
  if (!statSync(localeDir, { throwIfNoEntry: false })) continue

  const dict = readLocaleDir(localeDir)
  for (const keys of Object.values(dict)) for (const k of keys) knownKeys.add(k)

  const langs = Object.keys(dict)
  if (langs.length < 2) {
    console.log(`[${app}] only ${langs.length} locale file(s), cross-language check skipped`)
  }

  const refLang = langs.includes('zh_cn') ? 'zh_cn' : langs[0]
  const ref = dict[refLang]

  console.log(`\n=== ${app} (languages: ${langs.join(', ')}; reference: ${refLang}) ===`)

  // Languages the shared base provides must be spread into the app file, or
  // that language silently loses every shared key.
  for (const lang of langs) {
    if (!baseDict[lang]) continue
    if (!readFileSync(join(localeDir, `${lang}.ts`), 'utf8').includes(`@natholdallas/i18n/locale/${lang}`)) {
      problems++
      console.log(`  [base-spread] app/locale/${lang}.ts does not import the shared base '@natholdallas/i18n/locale/${lang}'`)
    }
  }

  for (const lang of langs) {
    if (lang === refLang) continue
    const keys = dict[lang]
    const missing = ref.filter((k) => !keys.includes(k))
    const extra = keys.filter((k) => !ref.includes(k))
    if (missing.length) {
      problems++
      console.log(`  [missing] ${lang} lacks ${missing.length}:`)
      for (const k of missing) console.log(`    - '${k}'`)
    }
    if (extra.length) {
      problems++
      console.log(`  [extra] ${lang} has ${extra.length} key(s) not in ${refLang}:`)
      for (const k of extra) console.log(`    - '${k}'`)
    }
    const idx = firstDiff(ref, keys)
    if (idx >= 0) {
      problems++
      console.log(`  [order] ${lang} diverges from ${refLang} at #${idx}: '${ref[idx]}' vs '${keys[idx] ?? '(end)'}'`)
    }
  }

  // Unused keys: union across every language file of the app.
  const allKeys = [...new Set(Object.values(dict).flat())]
  const used = scanUsage(appDir, allKeys)
  const unused = allKeys.filter((k) => !used.has(k))
  if (unused.length) {
    problems++
    console.log(`  [unused] ${unused.length} key(s) defined but never referenced in app source:`)
    for (const k of unused) console.log(`    - '${k}'`)
    if (deleteUnused) {
      const unusedSet = new Set(unused)
      for (const f of Object.keys(dict)) {
        const p = join(localeDir, `${f}.ts`)
        const lines = readFileSync(p, 'utf8').split('\n')
        const kept = []
        let dropContinuation = false
        for (const line of lines) {
          const m = line.trim().match(/^'([^']*)'|^"([^"]*)"/)
          const key = m?.[1] ?? m?.[2]
          if (key) {
            // A new entry starts: decide whether it is a deletion target and,
            // when the value is written on the following line(s), keep dropping
            // until the next key.
            dropContinuation = unusedSet.has(key)
            if (!dropContinuation) kept.push(line)
            continue
          }
          // A blank line or section comment ends an entry: keep it, otherwise
          // deleting a key would swallow the comment that introduces the next
          // section.
          if (dropContinuation && (!line.trim() || line.trim().startsWith('//'))) dropContinuation = false
          if (dropContinuation) continue
          kept.push(line)
        }
        writeFileSync(p, kept.join('\n'))
        console.log(`    removed ${lines.length - kept.length} line(s) from ${app}/app/locale/${f}.ts`)
      }
    }
  } else if (deleteUnused) {
    console.log('  (no unused keys to delete)')
  }

  // Keys the source references/implies but a language lacks. Shared base keys
  // count as available, so `create`, `dialog`, ... are not reported.
  const implied = collectImpliedKeys(appDir, knownKeys)
  for (const lang of langs) {
    const available = new Set([...dict[lang], ...(baseDict[lang] ?? [])])
    const missing = [...implied].filter((k) => !available.has(k))
    if (missing.length) {
      problems++
      console.log(`  [missing-from-code] ${lang} lacks ${missing.length} key(s) referenced/implied by the source:`)
      for (const k of missing) console.log(`    - '${k}'`)
    }
  }

  // Translated text handed to a Vuetify prop that expects a raw key.
  for (const { file, line, prop, value } of scanDoubleTranslation(appDir, keyProps)) {
    problems++
    console.log(`  [double-translation] ${file}:${line} :${prop} expects a key, got translated text: ${value}`)
  }

  if (!ref.length) console.log('  (no keys parsed)')
}

// The shared base dict (submodule) is read-only here, but consistency between
// its languages matters for every app that spreads it.
const baseLangs = Object.keys(baseDict)
if (baseLangs.length >= 2) {
  const refLang = baseLangs.includes('zh_cn') ? 'zh_cn' : baseLangs[0]
  const ref = baseDict[refLang]
  console.log(`\n=== shared base (languages: ${baseLangs.join(', ')}; reference: ${refLang}) ===`)
  for (const lang of baseLangs) {
    if (lang === refLang) continue
    const keys = baseDict[lang]
    const missing = ref.filter((k) => !keys.includes(k))
    const extra = keys.filter((k) => !ref.includes(k))
    if (missing.length) {
      problems++
      console.log(`  [missing] ${lang} lacks ${missing.length}:`)
      for (const k of missing) console.log(`    - '${k}'`)
    }
    if (extra.length) {
      problems++
      console.log(`  [extra] ${lang} has ${extra.length} key(s) not in ${refLang}:`)
      for (const k of extra) console.log(`    - '${k}'`)
    }
  }
}

console.log(`\n${problems ? `✖ ${problems} problem(s) found` : '✔ all locale files consistent'}`)
process.exit(problems ? 1 : 0)
