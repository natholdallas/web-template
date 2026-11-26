<div align="center">

# 🚀 Web Templates

### Production-Ready Full-Stack Scaffold

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Vue](https://img.shields.io/badge/Vue-3.x-4FC08D?style=flat&logo=vue.js)](https://vuejs.org/)
[![Nuxt](https://img.shields.io/badge/Nuxt-3.x-00DC82?style=flat&logo=nuxt.js)](https://nuxt.com/)
[![MySQL](https://img.shields.io/badge/MySQL-8.x-4479A1?style=flat&logo=mysql)](https://www.mysql.com/)
[![Redis](https://img.shields.io/badge/Redis-7.x-DC382D?style=flat&logo=redis)](https://redis.io/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

</div>

---

[简体中文](./README-CN.md)

## ✨ Overview

**Web Templates** is a production-ready, full-stack scaffold for modern web development. Kickstart your next project with a complete infrastructure — backend API, dual frontend portals, database, caching, and third-party integrations — all in one place.

> 🎯 **Configure once, build forever** — skip the boilerplate and focus on what matters.

---

## 🏗️ Architecture

|          Layer          | Stack                             | Purpose                         |
| :---------------------: | :-------------------------------- | :------------------------------ |
|     ⚙️ **Backend**      | Go 1.25 + Fiber v3                | High-performance HTTP framework |
|       🗄️ **ORM**        | GORM + MySQL/MariaDB              | Auto-migration & query builder  |
|      ⚡ **Cache**       | Redis (go-redis/v9)               | Session & hot data caching      |
| 🎨 **Frontend (User)**  | Nuxt 3 + Tailwind + shadcn-vue    | Modern UI components            |
| 🖥️ **Frontend (Admin)** | Nuxt 3 + TailwindCSS v4 + Vuetify | Material Design + utility CSS   |
|      📦 **State**       | Pinia + Alova                     | Reactive state & request hooks  |
|       🔐 **Auth**       | JWT (golang-jwt/v5)               | Dual-secret system (user/admin) |
|      📧 **Email**       | SMTP (QQ/Gmail)                   | HTML verification templates     |
|    💰 **WeChat Pay**    | wechatpay-go v3                   | Payment callbacks               |
|     📝 **API Docs**     | Swaggo + Scalar UI                | Auto-generated OpenAPI          |
|     🖥️ **Desktop**      | Tauri v2                          | Cross-platform desktop client   |
|    ⏰ **Scheduler**     | robfig/cron v3                    | Background jobs (rate sync)     |
|      🚢 **Deploy**      | Systemd (default) / tmux          | One-click deployment script     |

---

## 🌟 Features

```
🛡️  JWT Authentication   ──  Dual-channel auth for users & admins
👥  User Management      ──  Sign-in, reset password, profile
🛠️  Admin Panel          ──  CRUD for users & admins
💱  Exchange Rates       ──  Scheduled sync + API query
📤  File Upload          ──  Image & video media management
📧  Email Service        ──  HTML templates + verification codes
💳  WeChat Ecosystem     ──  Login, user info, payment callbacks
🌐  i18n                 ──  English / Chinese switch
🌗  Theme Toggle         ──  Light / Dark mode
📱  PWA                  ──  Progressive Web App support
🖥️  Tauri Desktop        ──  Cross-platform desktop app
```

---

## 🚀 Quick Start

### Prerequisites

- **Go** ≥ 1.25
- **Node.js** ≥ 20 + **pnpm**
- **MySQL** / **MariaDB**
- **Redis** (optional)

### Initialize

```bash
# First-time setup (install dependencies, init submodules)
./main.sh init

# Copy config/env templates into place (conf.toml is gitignored)
./main.sh copyfile
```

> ⚠️ **Before running:** edit `conf.toml` and fill in `secret.adm` / `secret.usr`
> (32-char strings, required — the app won't start without them). Generate them with:
>
> ```bash
> ./bin/backend --gensec
> ```

### Scaffold a New Project

```bash
# Rename the template to your project name
./main.sh renewal my-awesome-project
```

### Development

```bash
# Start all services (backend + user + admin)
./main.sh dev

# Stop development services
./main.sh dev stop
```

> `./main.sh dev` spawns a `tmux` session with three windows:
>
> - **Window 0**: Go backend hot-reload (gowatch)
> - **Window 1**: Admin frontend (port 3001)
> - **Window 2**: User frontend (port 3000)

### Build & Deploy

```bash
./main.sh build                # Build backend + both frontends (SSG)
./main.sh docs                 # Regenerate Swagger + frontend SDK (alova/wormhole)
./main.sh deploy               # Build, sync & (re)start the remote service
./main.sh deploy --skip-build  # Deploy without rebuilding
./main.sh deploy --force-service # Overwrite an existing systemd service file
```

> 🚢 `deploy` targets **Systemd by default** (into `/srv/http/<name>`, registered as a
> service) — set `deploy_mode=tmux` in `deploy.conf` to switch to the tmux-based
> deploy. Both are triggered through the single `deploy` command.

### Workflow

```
init → copyfile → (edit conf.toml) → dev
build → deploy
```

### More Commands

```bash
./main.sh synconf        # Sync conf.toml to the server
./main.sh serverlog      # Tail remote service logs (journalctl -f)
./main.sh serverlog once # Print the log once, then exit
./main.sh clean          # Remove node_modules / dist / .nuxt / .output
./main.sh push "message" # Commit & push everything
```

---

## ⚙️ main.sh Configuration

All deploy configuration lives in **`deploy.conf`** at the repo root (plain bash, sourced by
`main.sh`); `main.sh` keeps only internal constants and logic. Every command validates that
file first and fails fast, naming the offending entry.

### Project, deploy target & web server

```bash
pname=webgx        # systemd unit / tmux session / remote dir name
deploy_mode=systemctl  # systemctl | tmux
web_server=none        # none | nginx | caddy
use_rsync=true         # true=rsync, false=scp (env USE_RSYNC wins)
```

`pname` must match `^[a-z0-9][a-z0-9_-]{0,62}$` (it becomes a unit name, a tmux session name
and a remote directory name).

- `none` — the backend serves static itself (default, no site config generated/installed)
- `nginx` — renders `assets/nginx/*`, installs `/etc/nginx/conf.d/$pname.conf`, reloads
- `caddy` — renders `assets/caddy/*`, installs `/etc/caddy/conf.d/$pname.conf` (+ an
  `import conf.d/*.conf` line in the main Caddyfile), reloads

Both roll back to the previous site config when the new one fails `nginx -t` / `caddy validate`.

### Registered apps (`apps[]`)

Format: `name|dir|build_cmd|out_dir`

- `name` — app id; also the app's **site endpoint** (`usr`, `adm`, or anything you add) and the `app_env` key prefix. Must match `[a-z0-9][a-z0-9_-]*`; `api` and `media` are reserved
- `dir` — app directory (relative to repo root)
- `build_cmd` — build command run inside `dir` (may contain arguments)
- `out_dir` — build output dir relative to `dir` (what gets packaged & deployed)

```bash
apps=(
  "adm|web/apps/adm|pnpm generate|dist"  # run `pnpm generate` in web/apps/adm, package web/apps/adm/dist
  "usr|web/apps/usr|pnpm generate|dist"
)
```

Entries with the wrong field count, absolute paths, a missing `dir` or a duplicate name are
rejected before anything runs.

### Target servers (`servers[]`)

Format: `name|ssh_user|host`. Root users skip sudo automatically (`SUDO=""`).

```bash
servers=(
  "prod|deploy-user|203.0.113.72"  # deploys as deploy-user@203.0.113.72 using sudo
)
```

**The default server** (used by `build` / `synconf` / `webconf` / `serverlog` when no
`[server]` arg is given) is the **first entry** — reorder the array to change it.
`./main.sh deploy` ships to **every** entry in `servers[]` (frontends rebuilt per-server env);
`./main.sh build [server]` builds for one.

### Domains

One variable per endpoint: `domain_api` (backend), `domain_media` (media dir), and
`domain_<app>` for every entry in `apps[]`.

```bash
domain_usr=""    # user site    → usr app's <out_dir>/public
domain_adm=""    # admin site   → adm app's <out_dir>/public
domain_api=""    # backend API  → reverse proxy to 127.0.0.1:<conf.toml [app] port>
domain_media=""  # media files  → <deploy dir>/media
# domain_mgr=""  # any app you add to apps[] gets its own endpoint for free
```

Write bare `hostname` or `hostname:port` — no scheme, no path. Empty means the endpoint is
off. `domain_<x>` requires an app named `<x>` (a typo fails validation instead of being
ignored).

Adding a frontend is therefore two edits and no code change:

```bash
apps=(... "mgr|web/apps/mgr|pnpm generate|dist")
domain_mgr="mgr.hnath.club"
```

Its site block is rendered from `assets/nginx/mgr.conf` / `assets/caddy/mgr.conf` if that file
exists, otherwise from a built-in SPA block (`listen 80`, root = `<deploy dir>/<dir>/<out_dir>/public`,
`/_nuxt/` cache, SPA fallback). Tokens available in a custom template: `@DOMAIN@`,
`@SRV_DIR@`, `@MEDIA_DIR@`, `@API_ADDR@`, `@CLIENT_MAX_BODY@` and `@<APP>_DIR@` for every app
(upper-cased name, e.g. `@MGR_DIR@`).

Two values are **derived, never configured**:

| Derived value | Source | Override |
| :-- | :-- | :-- |
| api upstream (`@API_ADDR@`) | `conf.toml` `[app] port` | env `WEB_API_ADDR` |
| API body limit (`@CLIENT_MAX_BODY@`) | `conf.toml` `[app] body-limit` (MB) | env `CLIENT_MAX_BODY` |

When `web_server != none` the remote `conf.toml` is also patched so filling in the domains is
the only step needed:

| patched key | value |
| :-- | :-- |
| `app.nginx` | `true` — the backend stops serving the SPA/media, the web server does |
| `resources.media-url` | `https://<domain_media>` (inserted if the key is missing); empty → the frontend keeps relative `/media/...` URLs |
| `[cors] prd` | `http://<site>` and `https://<site>` appended for every **site** domain (`domain_api` / `domain_media` are skipped — no page is ever served from them) |

CORS is matched **exactly** against the browser's `Origin` header (`internal.AllowOriginsFunc`),
so those `prd` entries are what actually authorise the frontend; the reverse-proxy hop to
`127.0.0.1` is irrelevant. If `[app] debug = true`, the backend uses `[cors] dev` instead and
`deploy` prints a warning — set `debug = false` for production.

Each site block also proxies `/api/` to the backend and serves `/media/`, so
`NUXT_PUBLIC_API_BASE=same-origin` (and relative media URLs) work without an extra vhost.

### Env overrides (`app_env`)

Per-server per-app env overrides, keyed `server:app:VAR` (both the server and the app must
exist in the registries). Exported during that app's frontend build. Nuxt/c12 gives existing
env vars precedence over `--dotenv .env.production`, so `.env.production` stays a fallback.

```bash
declare -A app_env
app_env[prod:adm:NUXT_PUBLIC_API_BASE]=https://api.example.com
app_env[prod:adm:NUXT_PUBLIC_SITE_URL]=https://mgr.example.com
```

`NUXT_PUBLIC_API_BASE` is auto-derived from `domain_api` and `NUXT_PUBLIC_SITE_URL` from the
app's own site domain — explicit `app_env` entries win.

### Placeholder guard

`deploy` refuses to ship a bundle whose `NUXT_PUBLIC_API_BASE` or `NUXT_PUBLIC_SITE_URL` still
points at `localhost` / `127.0.0.1` / `example.com|org|net` (or is unset). Fix it by setting
`domain_api` / `domain_usr` / `domain_adm`, by adding an `app_env` override, or by using
`NUXT_PUBLIC_API_BASE=same-origin` when the frontend and API share one host.
`--allow-placeholder-env` bypasses the check; `build` only warns.

### Other switches

| Config        | Default             | Meaning                                                           |
| :------------ | :------------------ | :---------------------------------------------------------------- |
| `deploy_mode` | `systemctl`         | systemd deploys into `/srv/http/$pname`; `tmux` into `~/tasks/$pname` |
| `use_rsync`   | `true` (env `USE_RSYNC`) | file transfer: `rsync` vs `scp`                              |

Repo roots (`tmux_root`, `systemd_root`, `sys`) and the sudoers file suffix are **internal
constants in `main.sh`**, not configuration.

### How `deploy` works

1. Backend + SDK are built **once** (`build_core` — server-independent)
2. For **each** server: rebuild every app with that server's env (`build_front`) → zip →
   rsync/scp → `install_webconf` → patch remote `conf.toml` → extract, install the unit and
   (re)start the service

Web-server config and `conf.toml` are installed **before** the backend restarts, so the
process never comes up behind a site config that points at the old tree. Every remote step is
checked: a failed transfer, a corrupt zip, a rejected nginx/caddy config or a service that
does not become active fails the deploy (per server) instead of reporting success.

- `--skip-build` reuses previous artifacts — each app's `dist/.built-for` marker warns when
  they were built for a different server
- `--force-service` overwrites an existing systemd unit file (systemctl mode only)
- `./main.sh webconf [server]` installs/reloads the web config without redeploying

> **sudoers:** the generated `/etc/sudoers.d/<user>-webdeploy` whitelists `systemctl`,
> `install` into `/etc/...`, `journalctl -u *`, `nginx`/`caddy` checks and the Caddy
> `tee`. `install` + `systemctl` is effectively root-equivalent — use a dedicated deploy
> account on a machine you control, or deploy as `root` (which skips sudo entirely).

### `initremote`

Creates a GitHub repo (private by default), binds `origin`, and sets upstream — it never pushes.

```bash
./main.sh initremote                  # private repo named $pname
./main.sh initremote my-app --public  # public repo named my-app
# later: ./main.sh push "commit message"
```

---

## 📂 Project Structure

```
web-templates/
├── 📄 main.go                  # Backend entry point
├── 📄 conf.toml                # App configuration (TOML, gitignored)
├── 📄 main.sh                  # Unified management script ⭐
├── 📁 internal/                # Backend Go source
│   ├── 📁 srv/                 #  HTTP routes & handlers
│   │   ├── 📁 std/             #   Public endpoints (rate, webhook)
│   │   ├── 📁 usr/             #   User endpoints (JWT)
│   │   ├── 📁 adm/             #   Admin endpoints (JWT)
│   │   └── 📁 swg/             #   Scalar UI doc server
│   ├── 📁 db/                  #  Database connection & models (GORM)
│   ├── 📁 client/              #  External API clients (WeChat/Google/Rate)
│   ├── 📁 task/                #  Scheduled tasks (cron: rate sync)
│   ├── 📁 mail/                #  SMTP mailer + HTML templates
│   ├── 📁 flag/                #  CLI flag scripts (--adm/--usr/--mock/...)
│   └── 📁 conf/                #  Config loader (Viper)
├── 📁 web/                     # Frontend monorepo (pnpm)
│   ├── 📁 apps/
│   │   ├── 📁 usr/             #  User portal (Nuxt 3 + shadcn-vue + Tauri)
│   │   │   └── 📁 app/         #    Pages, components, stores, lib/sdk
│   │   └── 📁 adm/             #  Admin portal (Nuxt 3 + Vuetify + Tailwind v4)
│   │       └── 📁 app/         #    Pages, components, stores, lib/sdk
│   └── 📁 packages/            # Shared workspace packages
│       ├── 📁 apiclient/       #  alova/wormhole SDK generator (swagger → SDK)
│       └── 📁 natholdallas/    #  Git submodule of shared Nuxt modules
├── 📁 assets/                  # Templates & service units
│   ├── run.service             # Systemd service unit template
│   ├── conf.toml               # Default config template
│   └── nuxt.env                # Default frontend env vars
└── 📁 docs/                    # Swagger docs (swagger.json)
```

---

## 🔧 CLI Toolbox

```bash
# Admin operations
./bin/backend --adm          # Create admin interactively
./bin/backend --usr          # Create user interactively
./bin/backend --mock         # Seed test data

# Database operations
./bin/backend --migration    # Run auto-migration (add-only, safe)
./bin/backend --sync-db      # Sync table schemas with models (add/drop/reorder, keeps data)
./bin/backend --rstable      # Reset table structures to match models (drops data)
./bin/backend --rstdb        # Reset database

# Task scripts
./bin/backend --sync         # Run cron task scripts (e.g. sync exchange rates)
./bin/backend --gensec # Regenerate JWT secrets
```

---

## 📡 API Endpoints

| Group         | Prefix         | Auth | Sample Endpoints                                                       |
| :------------ | :------------- | :--: | :--------------------------------------------------------------------- |
| 🔓 **Public** | `/api/v1/`     |  —   | `GET /rate/:code`, `ALL /webhook/wechat`                               |
| 👤 **User**   | `/usr/api/v1/` | JWT  | `POST /auth/in`, `GET /user`, `PUT /user`, `POST /user/reset/password` |
| 🛡️ **Admin**  | `/adm/api/v1/` | JWT  | `POST /auth/in`, `CRUD /user`, `CRUD /admin`, `GET /stat`              |

> 📖 The interactive API docs (Scalar UI) are served at `/swg/doc/v1` — only when
> `app.swagger = true` in `conf.toml`. The spec is still written to `docs/` regardless,
> and `./main.sh docs` regenerates both the docs **and** the frontend SDK.

---

## 📄 License

[MIT](LICENSE) © Natholdallas

---

<div align="center">

### ⭐ If this project helps you, please give it a Star!

</div>
