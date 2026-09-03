# AGENTS.md

## Primary entrypoints

- **`./main.sh`** — unified CLI for all operations (not `go build`, not `make`)
- `./main.sh dev` — starts tmux session with 3 windows: Go hot-reload (gowatch), admin (3001), user (3000)
- `./main.sh build` — parallel: `go build -o bin/backend` + both `pnpm generate` (SSG, not `nuxt build`)
- `./main.sh init` — first-time: installs swag/gowatch, inits submodules, `go mod tidy`, `pnpm install`
- `./main.sh deploy [--skip-build] [--force-service]` — builds, zips, rsyncs to the server, and (re)starts the service. **Target mode (`tmux` | `systemctl`, default `systemctl`) is chosen by editing the `deploy_mode` variable at the top of `main.sh`** — there is no `deploy-tmux`/`deploy-systemctl` subcommand or `--use-systemctl` flag
- `./main.sh docs` — `swag fmt && swag init --parseDependency --parseInternal`, then regenerates the frontend SDK via `pnpm gen:api` (alova/wormhole)
- `./main.sh copyfile` — copies `assets/conf.toml` → `conf.toml` and `assets/nuxt.env` → each app's `.env`

## Architecture

```
Go backend (Fiber v3)         Nuxt 3 frontends (pnpm workspace)
├── internal/srv/std/  public ── port 8080, prefix /api/v1/
├── internal/srv/usr/  user   ── JWT, prefix /usr/api/v1/
├── internal/srv/adm/  admin  ── JWT, prefix /adm/api/v1/
├── internal/db/        GORM models (auto-migrate on startup)
├── internal/conf/      Viper config loading (init() runs LoadFlag+LoadApp)
├── internal/flag/      CLI flag scripts (--adm/--usr/--mock/...)
├── internal/auth/      JWT sign/verify
├── internal/pwd/       password hash/verify (only Go test lives here)
├── internal/mail/      SMTP mailer + templates
├── internal/task/      cron: rate sync at 0:00 and 12:00 daily
└── internal/client/    WeChat/Google/Rate API clients

web/apps/usr/  (port 3000) — shadcn-vue + TailwindCSS + Tauri v2
web/apps/adm/  (port 3001) — Vuetify + TailwindCSS (v4, CSS-first via `app/assets/styles/css/tailwind.css`; layer order declared in `public/layers.css`)
web/packages/natholdallas/ — git submodule (shared Nuxt modules). Used by apps: `alova`, `i18n`, `infra`, `tauri`, plus `vuetify` (adm) / `shadcn` (usr). Present but unused by any app: `unocss`, `watermark`, `pinia`, `tailwindcss`
```

## Go specifics

- **Module:** `webtplmst` — this name is baked into all imports, `main.sh`, and `assets/run.service`
- **Build output:** `bin/backend` (gitignored)
- **Config:** `conf.toml` (TOML, gitignored). Template at `assets/conf.toml`. Loaded via Viper.
- **Secrets required:** `secret.adm` and `secret.usr` (validated, 32-char strings). Generate with `./bin/backend --remake-secret`
- **Hot-reload:** `gowatch -o bin/backend` (gowatch installed via `go install github.com/silenceper/gowatch@latest`)
- **Go tabs** — 4-space tab indent (`.editorconfig`)
- **CLI flags:** `--adm`, `--usr`, `--rstdb`, `--migration`, `--sync`, `--mock`, `--remake-secret`
- **Auto-migrate:** runs `db.Migration()` on startup when `db.auto-migrate = true`
- **CORS prefix matching:** `AllowOriginsFunc` uses `strs.AnyPrefix`, not exact match — e.g. `http://localhost` also matches `http://localhost:3000`
- **Log rotation:** `lumberjack` writes to `<RLog>/app.log` (10MB, 7 backups, 28 days)
- **Lint:** after any backend code change, run `golangci-lint run --fix ./...` from the repo root (equivalent to running the frontend apps' per-app `pnpm lint`). golangci-lint v2 (2.13.2) is installed at `~/.local/share/go/bin/golangci-lint`. Config: `.golangci.yml` (revive `use-any` forces `any` over `interface{}`, gofumpt/goimports auto-format)
- **No `_ = expr` blank assignments** — never write code that discards a value/error with `_ = xxx` (e.g. `_ = db.AutoMigrate(...)`). Handle the error/result properly instead (check it, log it, or propagate it).
- **Formatting:** after writing/finishing any Go code, always run `goimports -w ./internal/*` from the repo root before committing
- **Go tests are minimal** — the only test file is `internal/pwd/pwd_test.go`; run it with `go test ./internal/pwd/`. There is no wider test suite.
- **`go mod` has a `replace`** directive for telegram-bot-api (redirects to a fork)

## GORM model conventions (internal/db/)

Every GORM model in `internal/db/` must follow these rules. Reference: `assets/exp.md`.

- **Full gotags on every field — no bare fields.** Each field must declare both tags:
  - `gorm` tag: always starts with `column:<snake_case>`; string columns add `size:<n>`; **every column adds `comment:<English description>`**; append constraints as needed (`primaryKey`, `unique`, `not null`, `default`, `index`, ...)
  - `json` tag: snake_case matching the column; `json:"-"` for fields never exposed (e.g. `Password`)
- **Trailing comment on every field** — `// <description>` matching the gorm `comment` value (see `internal/db/admin.go:12`).
- **Embed a base model instead of hand-writing ID/timestamps** — `orms.Model[uint]` (ID + created/updated at) or `orms.IDModel[uint]` (ID only). Do not redeclare `ID`, `CreatedAt`, `UpdatedAt` yourself.
- **Many-to-one association** (from `assets/exp.md`): embed the referenced struct plus the FK column, with the FK constraint and a comment:

  ```go
  type User struct {
      User   User   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"user"` // User
      UserID uint   `gorm:"column:user_id;comment:User" json:"user_id"`               // User ID
  }
  ```

- **JSON-persisted fields** (`orms.Dict[T]` / `orms.List[T]`): declare the column as `gorm:"type:json;comment:..."` (per `assets/exp.md`, a `Dict`'s type must be `json`).
- **Register new models** in `internal/db/db.go` → `Migration()` → `tx.AutoMigrate(&...{})`.

Example matching existing style:

```go
type User struct {
	orms.Model[uint]
	Username string `gorm:"column:username;size:50;unique;comment:Username" json:"username"` // Username
	Password string `gorm:"column:password;size:255;comment:Password" json:"-"`              // Password
} //	@name	User
```

## Frontend specifics

- **Package manager:** pnpm@11.25.0 (enforced in root `package.json`)
- **Workspace packages:** `apps/*`, `packages/*`, `packages/natholdallas/*`
- **Prod build:** `pnpm generate` (SSG / static generation via `nuxt generate --dotenv .env.production`), not `nuxt build`
- **SSR disabled:** both apps set `ssr: false`
- **Formatting:** `pnpm format` (Prettier) — run from `web/` dir. Config: no semis, single quotes, 120 print width, 2-space indent
- **ESLint:** auto-generated by Nuxt (`.nuxt/eslint.config.mjs`), imported by root config. Check per-app directory.
- **Lint:** after writing frontend code, run `pnpm lint` (or `pnpm lint:fix` to autofix) from `web/` — the root script is `pnpm -r --parallel run lint`, so it runs each app's own `lint`/`lint:fix` (`eslint .` / `eslint . --fix`) in parallel across every app under `web/apps/` (adm, usr, and any others like `xxx`, `xxx2`); no per-app loop needed, and packages without a lint script are skipped. Each app has its own `eslint.config.mjs` importing that app's generated `.nuxt/eslint.config.mjs` (adm additionally disables `vue/valid-v-slot` for Vuetify dotted slot names, and both apps ignore the generated `app/lib/sdk`); `.nuxt/` must exist first (`postinstall` runs `nuxt prepare`).
- **Env vars:** `NUXT_PUBLIC_API_BASE` (backend URL), `NUXT_PUBLIC_SITE_URL`, `ENABLE_PWA`, `ENABLE_SEO`
- **Env files:** `.env` (dev), `.env.production` (prod). Both copied from `assets/nuxt.env` by `./main.sh copyfile`.
- **Shared packages:** git submodule at `web/packages/natholdallas` → `https://github.com/natholdallas/nuxt-modules.git`
- **Test utils:** configured via `.nuxtrc`: `setups.@nuxt/test-utils="4.2.0"` — test with `nuxi test` from app dir

## Swagger → Frontend SDK (alova/wormhole)

The frontend SDKs are **generated** from the backend swagger doc, not hand-written. Flow: `swag init` produces `docs/swagger.json` → `web/packages/apiclient/alova.config.ts` runs `@alova/wormhole` → emits each app's `app/lib/sdk/*` directly (gitignored, no `gen/` subdir).

- **Entrypoint:** `./main.sh docs` regenerates swagger **and** the SDK. `./main.sh build` also runs `pnpm gen:api` before `pnpm generate`.
- **Config:** `web/packages/apiclient/alova.config.ts` — two generators (usr → `/usr/api/v1`+`/api/v1`, adm → `/adm/api/v1`+`/api/v1`), driven by plugins in `web/packages/apiclient/codegen/plugins.ts`. The root `pnpm gen:api` runs it via `pnpm --dir packages/apiclient exec alova gen -f`.
- **Generated files** per app (`app/lib/sdk/`): `index.ts` (alova runtime: `useRuntimeConfig().public.apiBase`, Bearer from `useAuth()`, `Api.NewEvent` — plus each app's wiring injected by `customIndex` in `alova.config.ts`), `createApis.ts`, `apiDefinitions.ts`, `globals.d.ts` (types + `declare global Apis`), `models.ts` (`type X = G.X` re-export from `globals` + `const X = {...}` model factories, including shared `PageQueries`/`SortQueries`/`BaseQueries`/`Page` helpers).
- **Runtime behavior** (`api.NewEvent`): 200 → success toast (non-GET), 401 → `useAuth().$signOut()`, else → error fallback by `code`/`message`. Wired in each app's generated `app/lib/sdk/index.ts` (edit the `usrWiring`/`admWiring` templates in `alova.config.ts`, then regenerate).

### Swagger annotation conventions (the ONLY manual cost)

Changing/adding an API = edit the swag annotations + run `./main.sh docs`.

- **`@ID` must be globally unique** across the whole doc. Naming: `{app}{Verb}{Entity}` e.g. `usrSignIn`, `admListUsers`; std (public) has **no** prefix (`findRate`) and must not collide with a stripped app name. Codegen strips the `usr`/`adm` prefix **and** the tag-matching entity suffix → concise methods under the tag group (`admListUsers` → `Apis.User.list`, `admCreateAdmin` → `Apis.Admin.create`). Reserve-word verbs (`delete`) are aliased (`deleteUser` → `Apis.User.remove`).
- **`@Tags` carries the app prefix** (`usrAuth`, `admAdmin`, `admUser`); std uses bare (`Rate`). Codegen strips it → SDK group `Apis.User`, `Apis.Auth`, ... The swagger UI (`internal/srv/swg/srv.go`) also strips it for display.
- **`@name` must be globally unique.** Every app-owned DTO gets a `Usr`/`Adm` prefix (`UsrUser`, `AdmAuthIn`, `AdmAdminsPage`); shared/global models (`User`, `Admin`, `Rate`, `Media`, `BaseQueries`, `Fail`) stay unprefixed. Codegen's `stripSchemaPrefix` strips the `Usr`/`Adm` prefix automatically — **no manual map needed** when adding a new DTO.
- **`@Success`/`@Param` should reference the Go type name**, not the `@name` (e.g. `{object} User` even if `@name UsrUser`).
- **Response models are assumed fully populated**: `requiredResponses` marks every property of 2xx-response schemas as required. So response DTOs must not have genuinely-missing fields (avoid `omitempty`-nullable response fields).
- **Error DTOs**: any handler returning `fext.Fail` should declare `@Failure 400/500 {object} Fail`.
- **Security**: `main.go` declares a global `ApiKeyAuth` header scheme; protected routes add `@Security ApiKeyAuth`.

## Project scaffolding

- **`./main.sh renewal <name>`** — renames the entire project. Replaces all `webtplmst` strings, deletes `.git`, removes/re-adds submodule, inits fresh git repo. Copies config/env files first. **Irreversible** — prompts for confirmation.

## Order of operations

```
init → copyfile → (edit conf.toml) → dev
build → deploy
```

## Key gotchas

- `conf.toml` is gitignored — must run `./main.sh copyfile` or manually copy `assets/conf.toml` before running
- `secret.adm` and `secret.usr` are validated as required — the app won't start without them
- The swagger UI is served at `/doc/api/v1` (mounted in `internal/srv/srv.go:41`) only when `app.swagger = true` in `conf.toml` (default `false`); the spec is still written to `docs/` regardless
- Frontend `nuxt.config.ts` imports from `@natholdallas/*` workspace packages — these come from the git submodule
- The `internal/conf` package `init()` function runs on import (loads flags and app config before `main()`)
- The generated SDK (`web/apps/*/app/lib/sdk/`) is **gitignored** — always regenerate via `./main.sh docs`, never commit it
- `swag` is pinned to **v2.0.0-rc5** in `go.mod` (`github.com/swaggo/swag/v2`) and installed by `./main.sh init` via `go install github.com/swaggo/swag/v2/cmd/swag@v2.0.0-rc5`. **Gotcha:** the `docs()` failure message inside `main.sh` still says `@v1.16.6` — trust `go.mod`, not that message. If `swag --version` differs, reinstall with the v2 command above
- `stripSchemaPrefix` only strips names whose remainder starts uppercase, so the shared `Admin` model (whose name happens to start with the `Adm` prefix) is preserved; don't introduce app DTOs whose stripped name would collide with a shared model used in the same app
- Every new DB model must have **complete gotags** (`gorm:"column:...;comment:..."` + `json` tag) and a trailing comment on **every field** — see the "GORM model conventions" section above and `assets/exp.md`
