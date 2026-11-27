# AGENTS.md

## Workflow and gates

- `./main.sh` is the only entrypoint; `./main.sh --help` is the authoritative command list. Do not replace its build/deploy/lint workflows with raw `go`/`make`; frontend scripts (`pnpm lint`, `pnpm generate`) are expected.
- `./main.sh dev [restart|stop]` starts tmux: window 0 is `gowatch -o bin/backend`, then one window per app (`pnpm dev`).
- `./main.sh build [server]` = `swag`, then `go build` and `gen:api` in parallel, then the frontends. Never run `gen:api` concurrently with `pnpm generate`: the generated SDK is an input to the frontend build.
- **Backend Go change:** run `./main.sh lint --fix`, then `./main.sh lint`; it must be clean. The gate is standalone `gofumpt`/`gci` (authoritative; golangci-lint's gofumpt formatter is deliberately disabled), `go build ./...`, `go vet ./...`, `staticcheck ./...`, `golangci-lint run ./...`, and `shellcheck -x main.sh`. Also run `gopls check <edited .go files>` manually; it matches the editor buffer.
- **API/swagger change:** run `./main.sh docs` (swag + SDK generation).
- **Frontend change:** run `pnpm lint` from `web/` (or `pnpm lint:fix`); fix every warning and error.
- **Never hand-edit generated files:** `docs/`, `web/packages/sdk/src/{usr,adm}/`, `.nuxt/`, `dist/`, `bin/backend`. The generated SDK directories and `docs/` are gitignored.
- **Never write `_ = expr`** to discard a value/error. Handle, log, or propagate it.
- **Do not edit `main.sh` to add servers/apps.** Edit `deploy.conf`.
- Order of operations: `init → copyfile → (edit conf.toml) → dev`; then `build → deploy`.

## Project map

```
Go backend (Fiber v3, module webgx → bin/backend)
├── internal/srv/{std,usr,adm}  /api/v1, /usr/api/v1, /adm/api/v1
├── internal/srv/{internal,swg}
├── internal/db                 GORM models (auto-migrate on start)
├── internal/conf               Viper; init() loads flags + app config
├── internal/auth               JWT + refresh rotation
├── internal/flag               CLI flags and scripts
├── internal/task               cron (rate sync 0:00/12:00)
├── internal/oauth              Google/Telegram
├── internal/client             WeChat/Google/Rate clients
├── internal/mail               SMTP + templates
└── internal/app                process lifecycle / wiring

web/apps/usr    port 3000 — shadcn-vue + Tailwind + Tauri v2
web/apps/adm    port 3001 — Vuetify 4 + Tailwind utilities + Tauri v2
web/packages/apiclient  swagger → SDK generator (not a Nuxt module)
web/packages/sdk        @sdk/client: hand-written runtime + generated SDK
web/packages/natholdallas  git submodule: optional shared Nuxt modules
```

## Configuration (`deploy.conf`)

- `deploy.conf` is the single source of truth for deploy: `pname` (systemd unit / tmux session / remote dir), `deploy_mode`, `web_server`, `use_rsync`, `apps[]`, `servers[]`, `domain_*`, `app_env[]`. `main.sh` holds logic only.
- `apps[]` = `name|dir|build_cmd|out_dir`. `name` is also the site endpoint and must match `[a-z0-9][a-z0-9_-]*`; `api`/`media` are reserved. Adding an app needs no `main.sh` edit: append to `apps[]`, set `domain_<name>`, optionally add `assets/<kind>/<name>.conf`.
- `servers[]` = `name|ssh_user|host`; the first is the default for `[server]` commands; `deploy` ships to every entry (frontends rebuilt per-server env). Root users skip sudo.
- `app_env[]` = `server:app:VAR`; explicit values beat `.env.production`. `NUXT_PUBLIC_API_BASE` is derived from `domain_api`, `NUXT_PUBLIC_SITE_URL` from `domain_<app>`.
- `domain_<endpoint>` = bare `hostname[:port]`; empty disables it. Endpoints are `domain_api`, `domain_media`, and one `domain_<app>` per app. Nginx sites listen on `:80` (use certbot for TLS); Caddy does auto-HTTPS.
- `web_server` = `none|nginx|caddy` (default `none`). When set, `deploy`/`webconf` render `assets/<kind>/*`, install/reload the remote config, and **rewrite remote `conf.toml`**: `debug=false`, `swagger=false`, `nginx=true`, `resources.media-url` when fronted, and rebuild `[cors] prd` (keep non-URL schemes like Tauri, drop stale http(s) origins, add http+https for each site). `webconf` runs before the backend restart; a failed web config fails the deploy.
- `deploy_mode` = `systemctl` (default) or `tmux`; there is no CLI flag or subcommand. `systemctl` installs `assets/run.service` when changed, runs as the deploy user with hardening, stops the service before unzip (ETXTBSY), restarts with a health check, and uses the shared sudoers file `/etc/sudoers.d/<user>-webdeploy`. `tmux` restarts a remote tmux session running `./bin/backend` after `unzip -tq`.
- `use_rsync`: `true`=rsync, `false`=scp; the tool must exist locally.
- Derived, never configured: API upstream from `conf.toml` `[app] port`; nginx body limit from `[app] body-limit` (MB). Override with `WEB_API_ADDR`/`CLIENT_MAX_BODY`.
- `--skip-build` refuses artifacts whose `dist/.built-for` marker does not match the target server/API base/site URL; `--allow-placeholder-env` bypasses the guard that rejects `localhost`/`example.com`.
- `./main.sh renewal <name>` renames the whole project, rewrites text files/directories, re-inits git, and is irreversible; it refuses on a dirty worktree and backs up by default (`--no-backup` skips).

## Go specifics

- Module `webgx`; build output `bin/backend`; config `conf.toml` (gitignored, template `assets/conf.toml`). HTTP port comes from `[app] port` (template: 8080).
- `secret.adm` / `secret.usr` are required 32-char strings. `./bin/backend --gensec` prints them; paste into `conf.toml`.
- Hot reload: `gowatch -o bin/backend`. Go uses 4-space tab indent.
- CLI flags: `--adm`, `--usr`, `--db-reset`, `--db-create`, `--db-migrate`, `--db-sync`, `--db-reset-table`, `--sync` (run the cron task once), `--mock`, `--gensec`. Defined in `internal/conf/flag.go`, dispatched in `internal/flag/flag.go`; `PreRun()` before `db.Connect()`, `Run()` after. Scripts in `internal/flag/runner.go` return `error`; no `os.Exit`/`fail()`.
- Auto-migrate: `db.Connect()` runs `db.Migrate()` when `db.auto-migrate=true`.
- CORS: exact match against `cors.dev`/`cors.prd` (`slices.Contains`); no prefix matching. List each allowed origin in full.
- Logs: lumberjack writes `<RLog>/app.log` (10MB, 7 backups, 28 days).
- No Go tests in this repo; verification is lint + build. `natools4go` has tests; run them there.
- `go.mod` has no `replace` directives; add one only with a documented reason.
- Swagger UI: `/doc/api/v1` when `app.swagger=true` (code default `false`; `assets/conf.toml` template is `true`; deploy forces `false`). The spec is always written to `docs/`.
- `swag` is pinned in `go.mod`; `./main.sh init` installs that version. Do not hand-install a different one.

### Database migration layers

| Flag | Behavior | Data |
|---|---|---|
| `--db-migrate` | `db.Migrate()` → `orms.MustAutoMigrate`: add-only, creates tables/columns, alters types/comments. Runs on startup when `db.auto-migrate=true`. | Safe |
| `--db-sync` | `orms.SyncDB`: add missing columns, **drop extra columns**, **reorder to struct order**. MySQL/MariaDB in place; other drivers rebuild + copy shared columns. | Dropped columns lose data |
| `--db-reset-table` | `orms.ResetTables`: drop + recreate each registered table from its struct (exact column order). | All data lost |
| `--db-reset` | `db.Reset()`: drop + recreate the whole database. | All data lost |

`--db-create` creates the database if missing. `db.Models` in `internal/db/db.go` is the single source of truth; `Migrate`, `SyncDB`, and `ResetTables` all iterate it.

### GORM models (`internal/db/`)

- Every field needs a full `gorm` tag starting with `column:<snake_case>`, string `size:<n>`, **every column** a `comment:<English description>`, plus constraints (`primaryKey`, `unique`, `not null`, `default`, `index`, ...). Add a `json` tag (snake_case; `json:"-"` when never exposed).
- Add a trailing `// <description>` comment on every field, matching the gorm `comment`.
- Embed `orms.Model[uint]` (ID + timestamps) or `orms.IDModel[uint]` (ID only); do not redeclare `ID`/`CreatedAt`/`UpdatedAt`.
- Group related columns into a plain embedded struct with `embedded;embeddedPrefix:<prefix>_` (see `GoogleInfo` / `TelegramInfo` in `internal/db/user.go`). Do not use many-to-one struct+FK embedding.
- JSON fields (`orms.Dict[T]` / `orms.List[T]`): `gorm:"type:json;comment:..."` (a `Dict` must be `json`).
- Register new models in `db.Models`; `Migrate()` picks them up.

```go
type User struct {
	orms.Model[uint]
	Username string `gorm:"column:username;size:50;unique;comment:Username" json:"username"` // Username
	Password string `gorm:"column:password;size:255;comment:Password" json:"-"`              // Password
} //	@name	User
```

### `orms` CRUD

| Helper | Purpose |
|---|---|
| `orms.First[T](tx, conds...)` / `Take` / `Last` / `Find` | Row/list queries; `I*` variants swallow errors and return zero values |
| `orms.Create[T](tx, v)` | Insert; pass `&v` to get the ID back |
| `orms.Save[T](tx, v)` | Upsert by PK |
| `orms.Delete[T](tx, conds...)` | Delete by PK or condition |
| `orms.UpdateByID[T](tx, id, column, value)` | Update one column |
| `orms.UpdatesByID[T](tx, id, values)` | Update multiple columns; pass a `map` when clearing/resetting |
| `orms.QE[T](tx)` | Chain scoped to model T |
| `orms.Q[T](tx)` / `orms.QM[T, M](tx)` | Chain without a model / map model M onto T |
| `orms.Transaction(tx, fn)` | Multi-step write |
| `orms.Count[T]` / `Exists[T]` / `PluckStrings[T]` | Aggregates |

- User-facing rows carry `Where("user_id = ?", claims.ID)`; read then verify before mutating (`usr.RemoveMedia`).
- Passwords: `pwd.Hash` / `pwd.MustHash` / `pwd.Verify` (argon2id only; no plaintext fallback).
- Clear to NULL with `gorm.Expr("NULL")` in the values map (or a Go `nil` pointer).

### List queries

```go
type UserQueries struct {
	orms.Sorter
	orms.Pagination
} //	@name	AdmUserQueries

func ListUser(c fiber.Ctx) error {
	q, err := fext.QueryParser[UserQueries](c)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	return c.JSON(orms.QE[db.User](db.Tx).
		Scopes(q.Sorter.Scope).
		IPaginate(q.Pagination))
}
```

`IPaginate` returns `orms.Page[T]` (`{total, page, content}`); size defaults to 20, caps at 100. `Sorter.Scope` validates column names; never put raw user input into `Order`. Each endpoint gets a same-shaped named Page DTO purely to name the swagger schema. Single-row mapping: `orms.QM[Profile, db.User](db.Tx).IFirst(claims.ID)`.

### Fiber handlers

- Mount groups in `srv.Setup()`; `/api/v1` (std) has no JWT, `/usr/api/v1` and `/adm/api/v1` use their own `jwt.Middleware`.
- Bodies: `fext.BodyVarser[T](c)`; queries: `fext.QueryParser[T](c)`; path IDs: `fext.IdentityParam[uint]` (or `ParamsParser`/`FormData`). Parse failures return `&fext.Fail{Code: internal.InvalidData, Message: err.Error()}`.
- Errors: `&fext.Fail{Code, Status, Message, System}`; `fext.ErrorHandler` renders `{code, message}` (400 default; `System` only in debug). Codes live in `internal/srv/internal/txtcode.go`. Success: `c.JSON(v)`; empty 200: `return nil`.
- Identity: `jwt.Claims(c)` returns `*jwt.RegisteredClaims`; `claims.ID` is a **string** — convert with `strs.ParseUint[uint]`/`strs.FormatUint`. Only call behind `jwt.Middleware`.

### JWT / refresh rotation

- One `authMgr` per app via `auth.New(secret, namespace, accessTTL, refreshTTL)`.
- `GenPair` records the refresh token in Redis (`<ns>:refresh:<jti>`). `Rotate` redeems once with a 30s `<ns>:refresh:used:<jti>` grace marker (racing tabs both succeed). `RevokeRefresh`/`RevokeAll` clear both.
- Access token in the response body; refresh token in an HttpOnly cookie (`usr_refresh`/`adm_refresh`, `Path=/{usr,adm}/api/v1/auth`, `SameSite=Lax`, `Secure` behind TLS, `SameSite=None` for the cross-site Tauri webview). `/auth/refresh` and `/auth/logout` take no body and read the cookie; logout is idempotent.
- `internal.OriginGuard` runs on both `/auth` groups and rejects a foreign `Origin` (not in `[cors]`, not same-origin). CORS allows credentials.

### Media / OAuth / conf / tasks

- Media: `internal.ParseUploadFile(c)` → `internal.SaveFile(c, file, kind, userID)`. `kind` is `image`/`video`/`file` (default `file`), validated by extension + MIME sniffing (`KindFile` rejects XSS-capable types). Files land at `<RMedia>/<YYYY>/<MM>/<rands.Char(20)><ext>`; `URL = conf.App.MediaBase() + "/" + path`. Failed DB write deletes the file; `Media.AfterDelete` deletes on row removal.
- OAuth: Google exchange (`oauth.GoogleExchange`) provisions on `ErrRecordNotFound`; Telegram Login Widget payload is verified by `oauth.CheckTelegramAuth` (HMAC + 5-minute freshness). Bind/unbind lives under `/usr/api/v1/profile/bind/...` and clears columns with `gorm.Expr("NULL")`.
- Config: everything via `conf.App`; Viper keys map dotted paths (`db.name`). `conf` `init()` loads flags + app config at import time, before `main()`; `Validate()` enforces required keys. Hot reload via `vipers.Watch(conf.Reload, db.Reload)`.
- Tasks: six-field cron (seconds first), e.g. `"0 0 0,12 * * ?"`. Rate sync fetches externally first, then one transaction does `DELETE FROM rates` + inserts. **Use `DELETE`, never `TRUNCATE`** (DDL implicit commit breaks the transaction).

## Frontend

- pnpm; version pinned in `web/package.json` `packageManager` (enforced by corepack). Workspace: `apps/*`, `packages/*`, `packages/natholdallas/*`.
- Prod build: `pnpm generate` (SSG, `--dotenv .env.production`), not `nuxt build`. Root `pnpm gen` runs both apps in parallel; `pnpm gen:api` regenerates the SDK. Both apps set `ssr: false`.
- Prettier: no semis, single quotes, 120 print width, 2 spaces. Run `pnpm format` from `web/`.
- ESLint is generated by Nuxt (`.nuxt/eslint.config.mjs`) and imported per app. `pnpm lint` = `pnpm check:locale && pnpm -r --no-bail run lint`; fix every warning/error. `pnpm lint:fix` autofixes.
- Env: `NUXT_PUBLIC_API_BASE`, `NUXT_PUBLIC_SITE_URL`, `ENABLE_PWA`, `ENABLE_SEO`. `.env`/`.env.production` come from `assets/nuxt.env*` via `copyfile` (only when absent; `--force` overwrites). Deploy rejects `localhost`/`example.com` API bases unless `--allow-placeholder-env`; use `NUXT_PUBLIC_API_BASE=same-origin` for a single-host reverse-proxied deployment.
- Auth store: `{ id, accessToken }` in memory only; refresh token is the HttpOnly cookie. `$restore()` trades the cookie once per page load (never during `nuxt generate`), `$refresh()` is single-flight, `$clear()` drops state, `$signOut(skipServer?)` clears then revokes. `app/middleware/auth.global.ts` guards every route; public pages are `entrance` and `oauth-google`.
- Shared modules are optional; an app loads only what it lists. Modules must not assign `runtimeConfig.public.*` (it overrides the app) and must declare every import. `infra` installs `@nuxtjs/seo`/`@vite-pwa/nuxt` only when `ENABLE_SEO`/`ENABLE_PWA`; do not install them disabled (the SEO prerender pipeline breaks SPA/SSG). The i18n base dict lives in the submodule; do not edit it.
- Tauri: `src-tauri/` per app (crate `webgx`/`webgx-adm`, id `com.webgx.<app>`); `frontendDist: ../dist/public`, `devUrl` = app port. Allow webview origins in `[cors]`: `tauri://localhost`, `http://tauri.localhost`, `https://tauri.localhost` (exact). `window.open` is a no-op in the webview (adm uses `tauri-plugin-opener`; usr hides Google/Telegram OAuth). Android in `gen/android` is committed; cleartext HTTP is debug-only; build with `ANDROID_HOME` set (`pnpm exec tauri android dev`, add `TAURI_DEV_HOST=<lan-ip>` for a device); iOS needs macOS/Xcode. `target/` and `gen/schemas/` are ignored; `Cargo.lock` and `gen/android` are committed.
- Tests: `nuxi test` from the app dir; per-app `.nuxtrc` pins `@nuxt/test-utils`.

### Locale files

- `web/apps/<app>/app/locale/<lang>.ts` (`zh_cn.ts`, `en_us.ts`, ...). Each imports and spreads the shared base first, then app keys. Never add app keys to `web/packages/natholdallas/i18n/locale/*.ts`.
- All languages of one app must define the same keys in the same order; `zh_cn` is the reference. Add `ja_jp`/`zh_tw` only when also listed in that app's `i18n.locales`.
- `pnpm check:locale` reports missing/extra keys, order divergence, unused keys, missing referenced keys, missing base spread, and Vuetify key-prop misuse. `--delete-unused` cleans an app; pass app names to narrow.
- Vuetify props resolved through vue-i18n (`no-data-text`, `loading-text`, `items-per-page-text`, `title` on `VFileUpload`, `label` on `VBadge`/`VOtpInput`) take the **key**, not `$t(...)`.

## Swagger → SDK

- Flow: `swag init` → `docs/swagger.json` → `web/packages/apiclient` (`alova.config.ts` + `codegen/`) → `web/packages/sdk/src/{usr,adm}/`. `./main.sh docs` does both.
- Generated per app: `apiDefinitions.ts`, `globals.d.ts`, `models.ts`, `api.ts`, `index.ts`, and a `createApis.ts` stub. The real proxy builder is `web/packages/sdk/src/createApis.ts`. Never edit generated files.
- Usage: `import { api, Model } from '@sdk/client/usr'` (or `/adm`); `api.<group>.<verb>(pathArg, data, params, config?)` — `config` last. Do not reintroduce an `Apis` proxy in app code. Apps install the runtime via `app/plugins/sdk.ts` → `setupSdk({ baseURL, auth, t, notify })`.
- Runtime (`web/packages/sdk/src/runtime.ts`, hand-written): one alova instance; `fetch` uses `credentials: 'include'`; 200 → success, 401 → refresh once per failed token (concurrent callers share it), then replay; second failure signs out. `meta.isRefresh` clears and throws; `meta.noRefresh` throws; `meta.silent` suppresses reports; other statuses/rejected fetches go to `fallback`/`onError`.
- Annotations: `@ID` globally unique (`{app}{Verb}{Entity}`; std no prefix). `@Tags` carries the app prefix (`usrAuth`); std bare (`Rate`). `@name` globally unique; app DTOs get `Usr`/`Adm`, shared models stay unprefixed. `@Success`/`@Param` reference the Go type name. 2xx response schemas are fully required — avoid `omitempty`-nullable response fields. Return `Fail` with `@Failure 400/500`; protected routes add `@Security ApiKeyAuth`.
- `stripSchemaPrefix` only strips when the remainder starts uppercase, so shared `Admin` survives; avoid stripped-name collisions.
