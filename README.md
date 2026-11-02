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
> ./bin/backend --remake-secret
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
> service) — set `deploy_mode="tmux"` at the top of `main.sh` to switch to the
> tmux-based deploy. Both are triggered through the single `deploy` command.

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

All project configuration lives at the **top of `main.sh`** (the `CONFIGURATION (edit only here)`
block) — the rest is derived state and logic. Everything below is documented here instead of
as verbose comments in the script.

### Registered apps (`apps[]`)

Format: `name|dir|build_cmd|out_dir`

- `name` — app id (also the key prefix in `app_env`)
- `dir` — app directory (relative to repo root)
- `build_cmd` — build command run inside `dir` (may contain arguments)
- `out_dir` — build output dir relative to `dir` (what gets packaged & deployed)

```bash
apps=(
  "adm|web/apps/adm|pnpm generate|dist"  # run `pnpm generate` in web/apps/adm, package web/apps/adm/dist
  "usr|web/apps/usr|pnpm generate|dist"
)
```

### Target servers (`servers[]`)

Format: `name|ssh_user|host`. Root users skip sudo automatically (`SUDO=""`).

```bash
servers=(
  "prod|deploy-user|203.0.113.72"  # deploys as deploy-user@203.0.113.72 using sudo
)
```

**The default server** (used by `build` / `synconf` / `webconf` / `serverlog` when no
`[server]` arg is given) is the **first entry** — reorder the array to change it.
`./main.sh deploy` ships to **every** entry in `servers[]` (frontends rebuilt per-server env).

### Env overrides (`app_env`)

Per-server per-app env overrides, keyed `server:app:VAR`. Exported during that app's
frontend build. Nuxt/c12 gives existing env vars precedence over `--dotenv .env.production`,
so `.env.production` stays a fallback.

```bash
declare -A app_env
app_env[prod:adm:NUXT_PUBLIC_API_BASE]=https://api.example.com
app_env[prod:adm:NUXT_PUBLIC_SITE_URL]=https://mgr.example.com
app_env[prod:usr:NUXT_PUBLIC_API_BASE]=https://api.example.com
app_env[prod:usr:NUXT_PUBLIC_SITE_URL]=https://example.com
```

> `NUXT_PUBLIC_API_BASE` / `NUXT_PUBLIC_SITE_URL` are **auto-derived** from `domains[]` +
> `app_bind[]` when the bound endpoint domain is set — explicit `app_env` entries win.

### Web server (`web_server`) & domains (`domains[]`)

- `none` — backend serves static itself (default, no site config generated/installed)
- `nginx` — renders `assets/nginx/*`, installs `/etc/nginx/conf.d/$pname.conf`, reloads
- `caddy` — renders `assets/caddy/*`, installs `/etc/caddy/conf.d/$pname.caddy` (+ an
  `import conf.d/*.caddy` line in the main Caddyfile), reloads

`domains[]` binds endpoints to domains (`endpoint|domain`, empty = endpoint disabled):

```bash
domains=(
  "usr|example.com"          # user-facing site → usr frontend's dist/public
  "adm|mgr.example.com"      # admin site       → adm frontend's dist/public
  "api|api.example.com"      # backend API      → reverse proxy to $web_api_addr
  "media|media.example.com" # media files      → $deploy_dir/media (fallback: api site /media route)
)
```

When `web_server != none` the remote `conf.toml` is also patched: `app.nginx = true` plus
`resources.media-url` (only when the media domain is set).

### Frontend ↔ endpoint binding (`app_bind[]`)

Maps each app to its `api` and `site` endpoints — drives the env auto-derivation above.

```bash
app_bind=(
  "usr|api|usr"
  "adm|api|adm"
)
```

### Other switches

| Config          | Default            | Meaning                                                          |
| :-------------- | :----------------- | :--------------------------------------------------------------- |
| `deploy_mode`   | `systemctl`        | systemd deploys into `/srv/http/$pname`; `tmux` into `~/tasks/$pname` |
| `use_rsync`     | `true` (env `USE_RSYNC`) | file transfer: `rsync` vs `scp`                             |
| `web_api_addr`  | `127.0.0.1:8080`   | api reverse-proxy target — must match `conf.toml app.port`       |
| `client_max_body` | `200m`           | nginx body limit — must match `conf.toml app.body-limit`         |
| `sudoers_name`  | `webtplmstx`       | shared fixed sudoers file name (merged per-user) — do not rename |

### How `deploy` works

1. Backend + SDK are built **once** (`build_core` — server-independent)
2. For **each** server: rebuild every app with that server's env (`build_front`) → zip →
   rsync/scp → install & (re)start the service
3. After each server, `install_webconf` runs (when `web_server != none`)

- `--skip-build` reuses previous artifacts — each app's `dist/.built-for` marker warns when
  they were built for a different server
- `--force-service` overwrites an existing systemd unit file (systemctl mode only)
- `./main.sh webconf [server]` installs/reloads the web config without redeploying

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
./bin/backend --remake-secret # Regenerate JWT secrets
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
