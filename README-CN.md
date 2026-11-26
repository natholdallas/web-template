<div align="center">

# 🚀 Web Templates

### 生产就绪的全栈脚手架

[English](./README.md) | 简体中文

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Vue](https://img.shields.io/badge/Vue-3.x-4FC08D?style=flat&logo=vue.js)](https://vuejs.org/)
[![Nuxt](https://img.shields.io/badge/Nuxt-3.x-00DC82?style=flat&logo=nuxt.js)](https://nuxt.com/)
[![MySQL](https://img.shields.io/badge/MySQL-8.x-4479A1?style=flat&logo=mysql)](https://www.mysql.com/)
[![Redis](https://img.shields.io/badge/Redis-7.x-DC382D?style=flat&logo=redis)](https://redis.io/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

</div>

---

## ✨ 概览

**Web Templates** 是一个面向现代 Web 开发的生产就绪全栈脚手架。它自带完整的基础设施——后端 API、双前端门户、数据库、缓存以及第三方集成——让你能直接开始下一个项目。

> 🎯 **一次配置，长期迭代** —— 跳过样板代码，专注真正重要的事。

---

## 🏗️ 架构

|          层          | 技术栈                            | 用途                     |
| :------------------: | :-------------------------------- | :----------------------- |
|     ⚙️ **后端**      | Go 1.25 + Fiber v3                | 高性能 HTTP 框架         |
|       🗄️ **ORM**     | GORM + MySQL/MariaDB              | 自动迁移与查询构建器     |
|      ⚡ **缓存**     | Redis (go-redis/v9)               | 会话与热数据缓存         |
| 🎨 **前端（用户）**  | Nuxt 3 + Tailwind + shadcn-vue    | 现代 UI 组件             |
| 🖥️ **前端（管理）**  | Nuxt 3 + TailwindCSS v4 + Vuetify | Material Design + 工具类 CSS |
|      📦 **状态**     | Pinia + Alova                     | 响应式状态与请求钩子     |
|       🔐 **认证**    | JWT (golang-jwt/v5)               | 用户/管理员双密钥体系    |
|      📧 **邮件**     | SMTP (QQ/Gmail)                   | HTML 验证模板            |
|    💰 **微信支付**   | wechatpay-go v3                   | 支付回调                 |
|     📝 **API 文档**  | Swaggo + Scalar UI                | 自动生成 OpenAPI         |
|     🖥️ **桌面端**    | Tauri v2                          | 跨平台桌面客户端         |
|    ⏰ **定时任务**   | robfig/cron v3                    | 后台任务（汇率同步）     |
|      🚢 **部署**     | Systemd（默认）/ tmux             | 一键部署脚本             |

---

## 🌟 特性

```
🛡️  JWT 认证         ──  用户与管理员双通道认证
👥  用户管理         ──  登录、重置密码、个人资料
🛠️  管理后台         ──  用户与管理员 CRUD
💱  汇率             ──  定时同步 + API 查询
📤  文件上传         ──  图片与视频媒体管理
📧  邮件服务         ──  HTML 模板 + 验证码
💳  微信生态         ──  登录、用户信息、支付回调
🌐  国际化           ──  英文 / 中文切换
🌗  主题切换         ──  亮色 / 暗色模式
📱  PWA              ──  渐进式 Web 应用支持
🖥️  Tauri 桌面端     ──  跨平台桌面应用
```

---

## 🚀 快速开始

### 环境要求

- **Go** ≥ 1.25
- **Node.js** ≥ 20 + **pnpm**
- **MySQL** / **MariaDB**
- **Redis**（可选）

### 初始化

```bash
# First-time setup (install dependencies, init submodules)
./main.sh init

# Copy config/env templates into place (conf.toml is gitignored)
./main.sh copyfile
```

> ⚠️ **运行前：** 编辑 `conf.toml`，填写 `secret.adm` / `secret.usr`
> （32 字符字符串，必填——缺失时应用无法启动）。生成方式：
>
> ```bash
> ./bin/backend --gensec
> ```

### 脚手架新项目

```bash
# Rename the template to your project name.
# --no-backup skips the pre-renewal git bundle + conf.toml copy (the clone's
# template history is not worth keeping; use it only on a throwaway clone).
./main.sh renewal my-awesome-project --no-backup
```

### 开发

```bash
# Start all services (backend + user + admin)
./main.sh dev

# Stop development services
./main.sh dev stop
```

> `./main.sh dev` 会启动一个包含三个窗口的 `tmux` 会话：
>
> - **窗口 0**：Go 后端热重载（gowatch）
> - **窗口 1**：管理端前端（端口 3001）
> - **窗口 2**：用户端前端（端口 3000）

### 构建与部署

```bash
./main.sh build                # Build backend + both frontends (SSG)
./main.sh docs                 # Regenerate Swagger + frontend SDK (alova/wormhole)
./main.sh deploy               # Build, sync & (re)start the remote service
./main.sh deploy --skip-build  # Deploy without rebuilding
./main.sh deploy --force-service # Overwrite an existing systemd service file
```

> 🚢 `deploy` **默认使用 Systemd**（部署到 `/srv/http/<name>` 并注册为服务）——在
> `deploy.conf` 里设置 `deploy_mode=tmux` 可切换到基于 tmux 的部署方式。两者都通过同一个
> `deploy` 命令触发。

### 工作流

```
init → copyfile → (edit conf.toml) → dev
build → deploy
```

### 更多命令

```bash
./main.sh synconf        # Sync conf.toml to the server
./main.sh serverlog      # Tail remote service logs (journalctl -f)
./main.sh serverlog once # Print the log once, then exit
./main.sh clean          # Remove node_modules / dist / .nuxt / .output
./main.sh push "message" # Commit & push everything
```

---

## ⚙️ main.sh 配置

所有部署配置都写在仓库根目录的 **`deploy.conf`** 里（纯 bash，由 `main.sh` source）；
`main.sh` 只保留内部常量与逻辑。每条命令都会先校验该文件，校验失败立即退出并指出出错条目。

### 项目、部署目标与 Web 服务器

```bash
pname=webgx        # systemd unit / tmux session / remote dir name
deploy_mode=systemctl  # systemctl | tmux
web_server=none        # none | nginx | caddy
use_rsync=true         # true=rsync, false=scp (env USE_RSYNC wins)
```

`pname` 必须匹配 `^[a-z0-9][a-z0-9_-]{0,62}$`（它会被用作单元名、tmux 会话名
和远程目录名）。

- `none` —— 后端自己提供静态文件（默认，不生成也不安装站点配置）
- `nginx` —— 渲染 `assets/nginx/*`，安装 `/etc/nginx/conf.d/$pname.conf` 并重载
- `caddy` —— 渲染 `assets/caddy/*`，安装 `/etc/caddy/conf.d/$pname.conf`（并在主
  Caddyfile 里加一行 `import conf.d/*.conf`），然后重载

当新配置无法通过 `nginx -t` / `caddy validate` 时，两者都会回滚到上一份站点配置。

### 注册的应用（`apps[]`）

格式：`name|dir|build_cmd|out_dir`

- `name` —— 应用标识；同时也是该应用的**站点端点**（`usr`、`adm` 或你新增的任何名称）以及 `app_env` 的键前缀。必须匹配 `[a-z0-9][a-z0-9_-]*`；`api` 与 `media` 为保留名
- `dir` —— 应用目录（相对仓库根目录）
- `build_cmd` —— 在 `dir` 内执行的构建命令（可以带参数）
- `out_dir` —— 相对 `dir` 的构建输出目录（被打包部署的内容）

```bash
apps=(
  "adm|web/apps/adm|pnpm generate|dist"  # run `pnpm generate` in web/apps/adm, package web/apps/adm/dist
  "usr|web/apps/usr|pnpm generate|dist"
)
```

字段数不对、使用绝对路径、`dir` 不存在或名称重复的条目会在任何操作执行前被拒绝。

### 目标服务器（`servers[]`）

格式：`name|ssh_user|host`。root 用户会自动跳过 sudo（`SUDO=""`）。

```bash
servers=(
  "prod|deploy-user|203.0.113.72"  # deploys as deploy-user@203.0.113.72 using sudo
)
```

**默认服务器**（`build` / `synconf` / `webconf` / `serverlog` 未指定 `[server]` 参数时使用）
是**第一条**条目——调整数组顺序即可更改。`./main.sh deploy` 会部署到 `servers[]` 里的
**每一个**条目（前端按各服务器环境重建）；`./main.sh build [server]` 只为一个服务器构建。

### 域名

每个端点一个变量：`domain_api`（后端）、`domain_media`（媒体目录），以及
`apps[]` 中每个应用的 `domain_<app>`。

```bash
domain_usr=""    # user site    → usr app's <out_dir>/public
domain_adm=""    # admin site   → adm app's <out_dir>/public
domain_api=""    # backend API  → reverse proxy to 127.0.0.1:<conf.toml [app] port>
domain_media=""  # media files  → <deploy dir>/media
# domain_mgr=""  # any app you add to apps[] gets its own endpoint for free
```

只写裸 `hostname` 或 `hostname:port`——不带协议、不带路径。留空表示该端点关闭。
`domain_<x>` 要求存在名为 `<x>` 的应用（拼错会校验失败，而不是被静默忽略）。

因此新增一个前端只需两处修改，无需改动代码：

```bash
apps=(... "mgr|web/apps/mgr|pnpm generate|dist")
domain_mgr="mgr.hnath.club"
```

它的站点块会优先从 `assets/nginx/mgr.conf` / `assets/caddy/mgr.conf` 渲染（若该文件存在），
否则使用内置 SPA 模板块（`listen 80`，root = `<deploy dir>/<dir>/<out_dir>/public`，
`/_nuxt/` 缓存，SPA 回退）。自定义模板中可用的占位符：`@DOMAIN@`、
`@SRV_DIR@`、`@MEDIA_DIR@`、`@API_ADDR@`、`@CLIENT_MAX_BODY@`，以及每个应用的
`@<APP>_DIR@`（名称大写，例如 `@MGR_DIR@`）。

有两个值**只做派生，从不配置**：

| 派生值 | 来源 | 覆盖方式 |
| :-- | :-- | :-- |
| api 上游（`@API_ADDR@`） | `conf.toml` `[app] port` | 环境变量 `WEB_API_ADDR` |
| API 请求体上限（`@CLIENT_MAX_BODY@`） | `conf.toml` `[app] body-limit`（MB） | 环境变量 `CLIENT_MAX_BODY` |

当 `web_server != none` 时，远端 `conf.toml` 也会被自动改写，因此填好域名就是唯一需要做的事：

| 被改写的键 | 值 |
| :-- | :-- |
| `app.nginx` | `true` —— 后端不再自己提供 SPA/媒体文件，改由 Web 服务器负责 |
| `resources.media-url` | `https://<domain_media>`（键不存在时插入）；为空则前端继续使用相对 `/media/...` URL |
| `[cors] prd` | 为每个**站点**域名追加 `http://<site>` 与 `https://<site>`（`domain_api` / `domain_media` 会被跳过——从它们不会提供任何页面） |

CORS 是与浏览器 `Origin` 头**精确**匹配的（`internal.AllowOriginsFunc`），
因此真正授权前端的是这些 `prd` 条目；到 `127.0.0.1` 的反向代理跳转与授权无关。
若 `[app] debug = true`，后端会改用 `[cors] dev` 并让 `deploy` 打印警告——生产环境请设置
`debug = false`。

每个站点块还会把 `/api/` 反代到后端并提供 `/media/`，因此
`NUXT_PUBLIC_API_BASE=same-origin`（以及相对媒体 URL）无需额外虚拟主机即可工作。

### 环境变量覆盖（`app_env`）

按服务器、按应用覆盖构建期环境变量，键为 `server:app:VAR`（服务器和应用都必须
存在于注册表中）。在该应用的前端构建过程中导出。Nuxt/c12 让已存在的环境变量
优先于 `--dotenv .env.production`，因此 `.env.production` 始终只是回退。

```bash
declare -A app_env
app_env[prod:adm:NUXT_PUBLIC_API_BASE]=https://api.example.com
app_env[prod:adm:NUXT_PUBLIC_SITE_URL]=https://mgr.example.com
```

`NUXT_PUBLIC_API_BASE` 由 `domain_api` 自动派生，`NUXT_PUBLIC_SITE_URL` 由该应用自己的
站点域名派生——显式的 `app_env` 条目优先。

### 占位符守卫

`deploy` 拒绝发布 `NUXT_PUBLIC_API_BASE` 或 `NUXT_PUBLIC_SITE_URL` 仍指向
`localhost` / `127.0.0.1` / `example.com|org|net`（或未设置）的产物。解决办法：设置
`domain_api` / `domain_usr` / `domain_adm`，添加一条 `app_env` 覆盖，或在前端与 API
共用同一主机时使用 `NUXT_PUBLIC_API_BASE=same-origin`。`--allow-placeholder-env` 可跳过该
检查；`build` 只会告警。

### 其他开关

| 配置          | 默认值              | 含义                                                              |
| :------------ | :------------------ | :---------------------------------------------------------------- |
| `deploy_mode` | `systemctl`         | systemd 部署到 `/srv/http/$pname`；`tmux` 部署到 `~/tasks/$pname`  |
| `use_rsync`   | `true`（环境变量 `USE_RSYNC`） | 文件传输方式：`rsync` 或 `scp`                        |

各种根目录（`tmux_root`、`systemd_root`、`sys`）以及 sudoers 文件名后缀是
**`main.sh` 里的内部常量**，不属于配置。

### `deploy` 的工作方式

1. 后端 + SDK 只构建**一次**（`build_core`——与服务器无关）
2. 对**每个**服务器：用该服务器的环境重建所有应用（`build_front`）→ 打包 →
   rsync/scp → `install_webconf` → 改写远端 `conf.toml` → 解包、安装单元并
   （重新）启动服务

Web 服务器配置与 `conf.toml` 都在后端重启**之前**安装完成，因此进程绝不会在
指向旧目录树的站点配置后面启动。每个远端步骤都会被检查：传输失败、zip 损坏、
nginx/caddy 配置被拒或服务未能激活，都会让（该服务器的）部署失败，而不是报告成功。

- `--skip-build` 复用上一次的产物——当产物是为其他服务器构建时，各应用的 `dist/.built-for`
  标记会给出警告
- `--force-service` 覆盖已存在的 systemd 单元文件（仅 systemctl 模式）
- `./main.sh webconf [server]` 只安装/重载 Web 配置，不重新部署

> **sudoers：** 生成的 `/etc/sudoers.d/<user>-webdeploy` 只放行 `systemctl`、
> 向 `/etc/...` 执行 `install`、`journalctl -u *`、`nginx`/`caddy` 检查以及 Caddy 的
> `tee`。`install` + `systemctl` 实际上等价于 root —— 请在你掌控的机器上使用专用的部署
> 账号，或以 `root` 部署（这会完全跳过 sudo）。

### `initremote`

创建一个 GitHub 仓库（默认私有），绑定 `origin`，并设置 upstream——它从不推送。

```bash
./main.sh initremote                  # private repo named $pname
./main.sh initremote my-app --public  # public repo named my-app
# later: ./main.sh push "commit message"
```

---

## 📂 项目结构

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

## 🔧 CLI 工具箱

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

## 📡 API 端点

| 分组        | 前缀           | 认证 | 示例端点                                                               |
| :---------- | :------------- | :--: | :--------------------------------------------------------------------- |
| 🔓 **公开**     | `/api/v1/`     |  —   | `GET /rate/:code`, `ALL /webhook/wechat`                               |
| 👤 **用户**     | `/usr/api/v1/` | JWT  | `POST /auth/in`, `GET /user`, `PUT /user`, `POST /user/reset/password` |
| 🛡️ **管理**     | `/adm/api/v1/` | JWT  | `POST /auth/in`, `CRUD /user`, `CRUD /admin`, `GET /stat`              |

> 📖 交互式 API 文档（Scalar UI）位于 `/swg/doc/v1`——仅在 `conf.toml` 里
> `app.swagger = true` 时可用。规范文件无论如何都会写入 `docs/`，
> 而 `./main.sh docs` 会同时重新生成文档**和**前端 SDK。

---

## 📄 许可证

[MIT](LICENSE) © Natholdallas

---

<div align="center">

### ⭐ 如果这个项目对你有帮助，请点个 Star！

</div>
