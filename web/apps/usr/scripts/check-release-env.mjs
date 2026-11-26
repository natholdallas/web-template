#!/usr/bin/env node
// Guard for packaged Tauri builds: a desktop/mobile bundle must never embed a
// localhost or http origin, because the shell has no reverse proxy in front of
// it. Runs before Nuxt so the failure is immediate, and reads .env.production
// itself — Nuxt has not loaded that file yet, and an exported env var (how
// main.sh passes domain_api) still wins, exactly as it does for Nuxt.
import fs from 'node:fs'
import path from 'node:path'

const envFile = path.resolve(process.cwd(), '.env.production')
let contents = ''
try {
  contents = fs.readFileSync(envFile, 'utf8')
} catch {
  // A missing file is treated as "unset": the check below fails either way.
}

function valueOf(name) {
  for (const raw of contents.split(/\r?\n/)) {
    const line = raw.trim()
    if (!line || line.startsWith('#')) continue
    const eq = line.indexOf('=')
    if (eq < 0 || line.slice(0, eq) !== name) continue
    return line.slice(eq + 1).replace(/^["']|["']$/g, '')
  }
  return ''
}

const base = process.env.NUXT_PUBLIC_API_BASE || valueOf('NUXT_PUBLIC_API_BASE')
if (!/^https:\/\//.test(base)) {
  console.error(
    `[check-release-env] NUXT_PUBLIC_API_BASE must be an https:// origin for a packaged build (got "${base || '<unset>'}").`,
  )
  console.error('  Set domain_api in deploy.conf, or export NUXT_PUBLIC_API_BASE=https://api.example.com before packaging.')
  process.exit(1)
}
