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

|          Layer          | Stack                          | Purpose                         |
| :---------------------: | :----------------------------- | :------------------------------ |
|     ⚙️ **Backend**      | Go 1.25 + Fiber v3             | High-performance HTTP framework |
|       🗄️ **ORM**        | GORM + MySQL/MariaDB           | Auto-migration & query builder  |
|      ⚡ **Cache**       | Redis v9                       | Session & hot data caching      |
| 🎨 **Frontend (User)**  | Nuxt 3 + Tailwind + shadcn-vue | Modern UI components            |
| 🖥️ **Frontend (Admin)** | Nuxt 3 + UnoCSS + Vuetify      | Material Design                 |
|      📦 **State**       | Pinia + Alova                  | Reactive state & request hooks  |
|       🔐 **Auth**       | JWT (golang-jwt/v5)            | Dual-secret system (user/admin) |
|      📧 **Email**       | SMTP (QQ/Gmail)                | HTML verification templates     |
|    💰 **WeChat Pay**    | wechatpay-go v3                | Payment callbacks               |
|     📝 **API Docs**     | Swaggo + Scalar UI             | Auto-generated OpenAPI          |
|     🖥️ **Desktop**      | Tauri v2                       | Cross-platform desktop client   |
|    ⏰ **Scheduler**     | robfig/cron v3                 | Background jobs (rate sync)     |
|      🚢 **Deploy**      | Systemd + tmux                 | One-click deployment script     |

---

## 🌟 Features

```
🛡️  JWT Authentication   ──  Dual-channel auth for users & admins
👥  User Management      ──  Sign-up, sign-in, reset password, profile
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
```

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
./main.sh build          # Build everything
./main.sh docs           # Generate API docs
./main.sh deploy         # One-click deploy to remote
```

---

## 📂 Project Structure

```
web-templates/
├── 📄 main.go                  # Backend entry point
├── 📄 conf.toml                # App configuration (TOML)
├── 📄 main.sh                  # Unified management script ⭐
├── 📁 internal/                # Backend Go source
│   ├── 📁 srv/                 #  HTTP routes & handlers
│   │   ├── 📁 std/             #   Public endpoints (rate, webhook)
│   │   ├── 📁 usr/             #   User endpoints (JWT)
│   │   └── 📁 adm/             #   Admin endpoints (JWT)
│   ├── 📁 db/                  #  Database models (GORM)
│   ├── 📁 client/              #  External API clients
│   │   ├── wechat.go           #   WeChat login/payment
│   │   ├── google.go           #   Google OAuth
│   │   └── rate.go             #   Exchange rate API
│   ├── 📁 task/                #  Scheduled tasks (cron)
│   ├── 📁 mail/                #  Email service
│   └── 📁 conf/                #  Config loader (Viper)
├── 📁 web/                     # Frontend monorepo (pnpm)
│   ├── 📁 apps/
│   │   ├── 📁 usr/             #  User portal (Nuxt 3)
│   │   │   ├── 📁 app/         #    Pages, components, stores
│   │   │   └── 📁 tauri/       #    Desktop app (Tauri v2)
│   │   └── 📁 adm/             #  Admin portal (Nuxt 3)
│   │       └── 📁 app/         #    Pages, components, stores
│   └── 📁 packages/            # Shared packages
├── 📁 assets/                  # Static assets & systemd
│   ├── run.service             # Systemd service unit
│   ├── conf.toml               # Default config template
│   └── nuxt.env                # Default env vars
└── 📁 docs/                    # Swagger docs
```

---

## 🔧 CLI Toolbox

```bash
# Admin operations
./bin/backend --adm          # Create admin interactively
./bin/backend --usr          # Create user interactively
./bin/backend --mock         # Seed test data

# Database operations
./bin/backend --migration    # Run auto-migration
./bin/backend --rstdb        # Reset database

# Utility
./bin/backend --sync         # Sync exchange rates
./bin/backend --remake-secret # Regenerate JWT secrets
```

---

## 📡 API Endpoints

| Group         | Prefix         | Auth | Sample Endpoints              |
| :------------ | :------------- | :--: | :---------------------------- |
| 🔓 **Public** | `/api/v1/`     |  —   | `GET /rate/:code`             |
| 👤 **User**   | `/usr/api/v1/` | JWT  | `POST /auth/in`, `GET /user`  |
| 🛡️ **Admin**  | `/adm/api/v1/` | JWT  | `POST /auth/in`, `CRUD /user` |

> 📖 API docs are served via Scalar UI at `/docs/*`. Run `./main.sh docs` to generate.

---

## 📄 License

[MIT](LICENSE) © Natholdallas

---

<div align="center">

### ⭐ If this project helps you, please give it a Star!

</div>
