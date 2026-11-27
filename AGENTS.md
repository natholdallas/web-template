# AGENTS.md

## Primary entrypoints

- **`./main.sh`** — unified CLI for all operations (not `go build`, not `make`)
- **Deploy configuration lives in `deploy.conf`** (repo root, bash, sourced by `main.sh`; validated on every command): `pname`, `deploy_mode`, `web_server`, `use_rsync`, `apps[]`, `servers[]`, `domain_usr/adm/api/media`, `app_env`. `main.sh` holds only internal constants (`tmux_root`, `systemd_root`, `sys`, and the shared sudoers file suffix) and logic — edit `deploy.conf`, never the script, to add/remove servers or frontend apps (see "Configuration" below).
- `./main.sh dev [restart|stop]` — starts tmux session: window 0 = Go hot-reload (gowatch), then one window per app in `apps[]` (`pnpm dev`); `restart`/`stop` kill the session.
- `./main.sh build [server]` — backend + SDK (`build_core`, sequential: swag → `go build` ∥ `gen:api`) and **then** the frontends for `[server]` (default: first entry in `servers[]`, `build_front`). They are sequential on purpose: `pnpm generate` imports the generated SDK, so building while `gen:api` rewrites `web/packages/sdk/src/{usr,adm}` races
- `./main.sh deploy [--skip-build] [--force-service] [--allow-placeholder-env]` — builds backend+SDK **once**, then for **EACH server in `servers[]`**: rebuilds every app with that server's env, zips, rsyncs, and (re)starts the service. **Target mode (`tmux` | `systemctl`, default `systemctl`) is chosen by editing `deploy_mode` in `deploy.conf`** — there is no `deploy-tmux`/`deploy-systemctl` subcommand or `--use-systemctl` flag. `--skip-build` reuses artifacts and **refuses** them (each app's `dist/.built-for` marker records server + API base + site URL) when they no longer match the target or the current env; the placeholder guard runs again in that mode. `--allow-placeholder-env` bypasses the guard that otherwise aborts when an app would ship `localhost`/`example.com` as its API base
- `./main.sh servers` — list available servers/apps/domains from `deploy.conf`
- `./main.sh docs` — `swag fmt && swag init --parseDependency --parseInternal`, then regenerates the frontend SDK via `pnpm gen:api` (alova/wormhole)
- `./main.sh lint [--fix]` — the lint + format gate for the Go backend and `main.sh` itself (gofumpt/gci, `go build`, `go vet`, staticcheck, golangci-lint, shellcheck); `--fix` applies gofumpt/gci first. Fails loudly when a tool is missing
- `./main.sh copyfile [--force]` — copies `assets/conf.toml` → `conf.toml` **only when absent** (an existing `conf.toml` holds the JWT secrets; `--force` overwrites it), then `assets/nuxt.env` → each app's `.env` and `assets/nuxt.env.production` → each app's `.env.production`
- `./main.sh init` — first-time: installs gowatch/swag, inits submodules, `pnpm install`, `go mod tidy`, then `docs` (swag needs the workspace installed and the submodule present)
- `./main.sh synconf [server]` — rsync `conf.toml` to the mode-appropriate remote dir of `[server]` (default: first server in `servers[]`)
- `./main.sh webconf [server]` — when `web_server != none`, generate & install the nginx/caddy site config on `[server]` from the `domain_*` variables and reload it (no redeploy; the previous config is restored if the new one fails validation). `deploy` runs this per server **before** restarting the backend, so a failed web config fails that deploy instead of silently succeeding.
- `./main.sh serverlog [server] [once|lines]` — tail remote `journalctl -u <name>` (default follows; requires limited sudo)
- `./main.sh push "msg"` — `git add -A && git commit -m "msg" && git push` (defaults to `upd`)
- `./main.sh renewal <name> [--no-backup]` — renames the entire project (irreversible, see "Project scaffolding"); `name` must satisfy the `pname` pattern and every `*.sh`/`*.conf`/`*.toml`/`*.md`/… file (including `deploy.conf`) is rewritten. `--no-backup` skips the pre-renewal git bundle + `conf.toml` copy (used by the `webgx` scaffolder, which starts from a throwaway template clone)
- `./main.sh initremote [name] [--public|--private]` — create a GitHub repo (private by default), bind `origin`, set upstream (never pushes)

## Configuration (`deploy.conf`)

- `pname` — project name used as the systemd unit, tmux session and remote dir name; must match `^[a-z0-9][a-z0-9_-]{0,62}$`.
- `apps[]` — frontend apps, `name|dir|build_cmd|out_dir`. `build_cmd` runs inside `dir` (e.g. `pnpm generate`); `out_dir` is the build output relative to `dir`. The app `name` is also its **site endpoint** (`usr`/`adm`, plus any app you add) — it must match `[a-z0-9][a-z0-9_-]*`, and `api`/`media` are reserved.
- `servers[]` — deploy targets, `name|ssh_user|host`; the **default server** (for `build`/`synconf`/`webconf`/`serverlog` when no `[server]` is given) is the **first entry**. **`./main.sh deploy` ships to every entry in `servers[]`** (frontends rebuilt per-server env). Root users skip sudo automatically (`SUDO=""`, sudoers setup skipped).
- `app_env[]` — per-server per-app env overrides keyed `"server:app:VAR"` (server and app must exist in the registries), exported during each app's frontend build. Nuxt/c12 gives existing env vars precedence over `--dotenv .env.production`, so `.env.production` stays a fallback. **`NUXT_PUBLIC_API_BASE` is auto-derived from `domain_api` and `NUXT_PUBLIC_SITE_URL` from the app's own `domain_<app>`** — explicit `app_env` entries still win.
- `web_server` — `none` (default) | `nginx` | `caddy`. When not `none`, deploy/webconf render `assets/<kind>/*` for every **non-empty** `domain_*` and install+reload on the remote (`/etc/nginx/conf.d/$pname.conf`, or `/etc/caddy/conf.d/$pname.conf` + an `import conf.d/*.conf` line in the main Caddyfile). It also **rewrites** the remote `conf.toml` on every deploy (`patch_conf_toml` → `assets/patch-conf.awk`, fetched, patched locally, uploaded back): `app.debug = false` and `app.swagger = false` always; `app.nginx = true` and `resources.media-url` when fronted/the media domain is set; and `[cors] prd` is rebuilt — every non-URL scheme (Tauri) is kept, stale `http(s)://` origins are dropped, and each **site** domain gets `http://` + `https://` (CORS is an exact match on the browser `Origin`; `domain_api`/`domain_media` are skipped). A structure it cannot patch (missing `[app] debug`/`swagger`/`nginx` or `[cors]`) fails the deploy. The rendered site blocks also proxy `/api/` and serve `/media/`, which is what makes `NUXT_PUBLIC_API_BASE=same-origin` work.
- `domain_<endpoint>` — one variable per endpoint, bare `hostname[:port]` (empty = that site is skipped): `domain_api` (backend reverse proxy), `domain_media` (media files), and `domain_<app>` for every entry in `apps[]`. Endpoints are **derived from `apps[]`** (`endpoints()` in `main.sh`), and `validate_config` walks `compgen -A variable domain_`, so a `domain_<x>` with no matching app is an error. Static roots derive from `apps[]`: `<deploy_dir>/<app dir>/<out_dir>/public`, `<deploy_dir>/media` for media. Nginx sites listen on `:80` (run `certbot --nginx` for TLS); caddy does auto-HTTPS.
- **Adding a frontend needs no `main.sh` edit**: append to `apps[]` and set `domain_<name>`. Its site block comes from `assets/<kind>/<name>.conf` when present, else from the built-in SPA block in `default_site()` (`root` = `<deploy dir>/<app dir>/<out_dir>/public`, `/_nuxt/` cache, SPA fallback). Template tokens: `@DOMAIN@`, `@SRV_DIR@`, `@MEDIA_DIR@`, `@API_ADDR@`, `@CLIENT_MAX_BODY@`, `@<APP>_DIR@` (upper-cased app name).
- **Derived, never configured:** the API upstream comes from `conf.toml` `[app] port` and the nginx body limit from `[app] body-limit` (MB). Override per machine with env `WEB_API_ADDR` / `CLIENT_MAX_BODY`.
- `deploy_mode` — `systemctl` (default) or `tmux`:
  - `systemctl` renders `assets/run.service` (`@PNAME@`/`@SRV_DIR@`/`@RUN_USER@`) and installs it under `/etc/systemd/system` when the rendered unit differs from the installed one (or `--force-service`), then enables/restarts with a health check. The unit runs as the deploy user (`User=@RUN_USER@`, never root) with `NoNewPrivileges`, `ProtectSystem=strict`, `ProtectHome`, `PrivateTmp`, `RestrictAddressFamilies`, `ReadWritePaths=@SRV_DIR@`, and `CacheDirectory`/`XDG_CACHE_HOME` so `os.UserCacheDir()` still resolves. The service is **stopped before unzip** (ETXTBSY) and started afterwards. Sudoers is a **shared, fixed** file `/etc/sudoers.d/<user>-webdeploy` (do not rename) with exact, wildcard-free commands for this project (systemctl lifecycle for `<pname>`, `install` of `<deploy_dir>/<pname>.service` + `<deploy_dir>/.webconf`, `install -d` of `/srv/http[/<pname>]`, journalctl, nginx/caddy validate+reload, `tee -a /etc/caddy/Caddyfile`); binaries are resolved on the remote. The sudoers candidate is a `mktemp` file in `$HOME` (umask 077), validated with `visudo -cf`, then installed 0440. `install` + `systemctl` is still root-equivalent in practice — use a dedicated deploy account or deploy as root.
  - `tmux` kills/restarts a remote tmux session running `./bin/backend`; the zip is integrity-checked (`unzip -tq`) before the session is touched.
- `use_rsync` (`USE_RSYNC`) — transfer method: `true`=rsync, `false`=scp. The tool must exist locally or the transfer fails loudly.

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
├── internal/mail/      SMTP mailer + templates (password hashing lives in the external natools4go/pwd)
├── internal/task/      cron: rate sync at 0:00 and 12:00 daily
└── internal/client/    WeChat/Google/Rate API clients

web/apps/usr/  (port 3000) — shadcn-vue + TailwindCSS + Tauri v2
web/apps/adm/  (port 3001) — Vuetify 4 (`@natholdallas/vuetify` + `vuetify-nuxt-module`) + TailwindCSS utilities + Tauri v2
web/packages/natholdallas/ — git submodule of shared, optional Nuxt modules (see "Shared Nuxt modules" below)
web/packages/apiclient/   — build-time SDK codegen tool, not a Nuxt module (see "Swagger → Frontend SDK" below)
web/packages/sdk/         — `@sdk/client`: the alova runtime (hand-written) + the generated per-app SDK (see "Swagger → Frontend SDK" below)
```

## Go specifics

- **Module:** `webgx` — this name is baked into all imports, `main.sh`, and `assets/run.service`
- **Build output:** `bin/backend` (gitignored)
- **Config:** `conf.toml` (TOML, gitignored). Template at `assets/conf.toml`. Loaded via Viper.
- **Secrets required:** `secret.adm` and `secret.usr` (validated, 32-char strings). Generate with `./bin/backend --gensec`, which **prints** new secrets — you must paste them into `conf.toml` yourself
- **Hot-reload:** `gowatch -o bin/backend` (gowatch installed via `go install github.com/silenceper/gowatch@latest`)
- **Go tabs** — 4-space tab indent (`.editorconfig`)
- **CLI flags** (defined in `internal/conf/flag.go`, dispatched in `internal/flag/flag.go`): `--adm`, `--usr`, `--db-reset`, `--db-create`, `--db-migrate`, `--db-sync`, `--db-reset-table`, `--sync` (task script), `--mock`, `--gensec`. `--db-sync` maps to `orms.SyncTablesHard` (add missing columns, drop removed ones, reorder); `--db-reset-table` maps to `orms.ResetTables`
- **Auto-migrate:** `db.Connect()` runs `db.Migrate()` on startup when `db.auto-migrate = true`

### Database migration layers

Four distinct CLI flags handle schema evolution (`--sync` is unrelated — it runs the cron task scripts). All are dispatched from `internal/flag/runner.go`:

| Flag             | Source                                             | Behavior                                                                                                                                                                                                                                                             | Data                            |
| ---------------- | -------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------- |
| `--db-migrate`   | `db.Migrate()` → `orms.MustAutoMigrate`            | Incremental, add-only: creates missing tables/columns, alters types/comments. Runs on startup when `db.auto-migrate`                                                                                                                                                 | Safe                            |
| `--db-sync`      | `orms.SyncDB(db.Tx, db.Models...)`                 | Data-preserving reconciliation per table: add missing columns, **drop extra ones**, **reorder columns to struct order**. MySQL/MariaDB reorder in place via `ALTER TABLE ... MODIFY COLUMN ... FIRST/AFTER`; other drivers rebuild the table and copy shared columns | Dropped columns lose their data |
| `--db-reset-table` | `orms.ResetTables(db.Tx, db.Models...)`            | Drops + recreates each registered table from its struct so column order exactly matches declaration order                                                                                                                                                            | All data lost                   |
| `--db-reset`     | `db.Reset()`                                       | Drops + recreates the whole database                                                                                                                                                                                                                                 | All data lost                   |

`--db-create` (`db.AutoCreate()`) creates the database if missing. The GORM model set is the single source of truth — registered in `db.Models` (`internal/db/db.go`), which `Migrate`, `SyncDB`, and `ResetTables` all iterate. GORM's own `AutoMigrate` only ever appends new columns (at the end) and never drops or reorders, which is why `--db-sync`/`--db-reset-table` exist.

- **CORS origin matching:** `AllowOriginsFunc` compares origins **exactly** against `cors.dev` / `cors.prd` (`slices.Contains`). Prefix matching was removed because it accepted `https://example.com.attacker.net` for a whitelisted `https://example.com`. List every allowed origin in full.
- **Log rotation:** `lumberjack` writes to `<RLog>/app.log` (10MB, 7 backups, 28 days)
- **Lint — must run after any backend Go change and be completely clean before reporting done** (the team's nvim — AstroNvim/astrolsp with `gopls` + `staticcheck = true` — surfaces exactly these diagnostics, so a clean run here equals a zero-diagnostic editor buffer). **Run `./main.sh lint`**, which runs:
  - `gofumpt -l .` — must print nothing; `gci diff -s standard -s default -s 'Prefix(webgx)' .` — must print nothing. These standalone binaries are the **authoritative** format gate: golangci-lint 2.14's built-in `gofumpt` formatter disagrees with them (it flags already-formatted files, and the flagged set changes between runs), so `formatters` is deliberately absent from `.golangci.yml`. `./main.sh lint --fix` applies both.
  - `go build ./...` — must compile
  - `go vet ./...`
  - `~/.local/share/go/bin/staticcheck ./...`
  - `~/.local/share/go/bin/golangci-lint run ./...` — analysis linters only (govet/staticcheck/unused/revive/misspell); `revive use-any` forces `any` over `interface{}`
  - `shellcheck -x main.sh` — the deploy script is code too. `main.sh lint` **fails** when any of these tools (or `go`/`shellcheck`) is missing instead of reporting a false green.
  - `gopls check <each edited .go file>` (run manually — not part of `main.sh lint`) — exact parity with the editor buffer; this is what reports "should use range loop", "`err` shadows declaration at ...", "use `min`/`max`", "use `strings.Builder`/`WriteString`", and similar hints. If `gopls check` is clean, the nvim buffer will be too.
  - Fix everything each tool reports and re-run until the output is clean; never finish a change with leftover diagnostics.
- **No `_ = expr` blank assignments** — never write code that discards a value/error with `_ = xxx` (e.g. `_ = db.AutoMigrate(...)`). Handle the error/result properly instead (check it, log it, or propagate it).
- **No Go tests exist** — there is no `*_test.go` anywhere in *this* repo (password helpers come from `natools4go/pwd`). Don't assume a test suite here; the verification is the lint suite (see **Lint** above) + `go build`. (`natools4go` itself does have tests — run `go test ./...` there when changing it.)
- **`go mod` has no `replace` directives** — the old telegram-bot-api fork override was unused (the module is not in the dependency graph) and has been removed. Add one only when a dependency genuinely needs patching, and document why.

## GORM model conventions (internal/db/)

Every GORM model in `internal/db/` must follow these rules.

- **Full gotags on every field — no bare fields.** Each field must declare both tags:
  - `gorm` tag: always starts with `column:<snake_case>`; string columns add `size:<n>`; **every column adds `comment:<English description>`**; append constraints as needed (`primaryKey`, `unique`, `not null`, `default`, `index`, ...)
  - `json` tag: snake_case matching the column; `json:"-"` for fields never exposed (e.g. `Password`)
- **Trailing comment on every field** — `// <description>` matching the gorm `comment` value (see `internal/db/admin.go:12`).
- **Embed a base model instead of hand-writing ID/timestamps** — `orms.Model[uint]` (ID + created/updated at) or `orms.IDModel[uint]` (ID only). Do not redeclare `ID`, `CreatedAt`, `UpdatedAt` yourself.
- **Embedded prefixed structs** (this project does **not** use the many-to-one struct+FK embedding pattern): group related columns into a plain struct embedded with a column prefix — see `internal/db/user.go` (`GoogleInfo`/`TelegramInfo`):

  ```go
  type User struct {
      orms.Model[uint]
      Username   string       `gorm:"column:username;size:50;unique;comment:Username" json:"username"` // Username
      GoogleInfo GoogleInfo   `gorm:"embedded;embeddedPrefix:google_" json:"-"`                        // Google login info
  }

  type GoogleInfo struct {
      Email *string `gorm:"column:email;unique;size:256;comment:Google email" json:"email"` // Google email
  } //	@name	GoogleInfo
  ```

- **JSON-persisted fields** (`orms.Dict[T]` / `orms.List[T]`): declare the column as `gorm:"type:json;comment:..."` (a `Dict`'s type must be `json`).
- **Register new models** in `internal/db/db.go` → add to the `Models` var → `Migrate()` runs `orms.MustAutoMigrate` over them.

Example matching existing style:

```go
type User struct {
	orms.Model[uint]
	Username string `gorm:"column:username;size:50;unique;comment:Username" json:"username"` // Username
	Password string `gorm:"column:password;size:255;comment:Password" json:"-"`              // Password
} //	@name	User
```

## Backend code patterns (`internal/srv`, `auth`, `conf`, `task`, `flag`)

When writing backend code, copy these existing shapes; do not invent new ones.

### `orms` generic CRUD

Every helper takes the `*gorm.DB` session as its first argument — pass `db.Tx`, or `t` inside a transaction.

| Helper | Purpose |
|---|---|
| `orms.First[T](tx, conds...)` | One row by PK or any condition: `orms.First[User](tx, "username = ?", username)` |
| `orms.Take[T]` / `orms.Last[T]` | One row, no ordering assumption |
| `orms.Find[T](tx, conds...)` | List rows |
| `orms.IFirst[T]` / `ITake` / `ILast` / `IFind` | Same, but swallow the error and return the zero value |
| `orms.Create[T](tx, v)` | Insert (pass `&v` to get the ID back) |
| `orms.Save[T](tx, v)` | Upsert by PK |
| `orms.Delete[T](tx, conds...)` | Delete by PK or condition |
| `orms.UpdateByID[T](tx, id, column, value)` | Update a single column |
| `orms.UpdatesByID[T](tx, id, values)` | Update multiple columns; **pass a map**, otherwise zero values (`0`/`false`/`""`) are not written |
| `orms.QE[T](tx)` | Chain wrapper scoped to model T: `.QE[Media](tx).Where(...).Scopes(...).IPaginate(...)` |
| `orms.Q[T](tx)` / `orms.QM[T, M](tx)` | Chain without a model / map model M onto result type T |
| `orms.Transaction(tx, fn)` | Wrap a multi-step write |
| `orms.Count[T]` / `Exists[T]` / `PluckStrings[T]` | Quick aggregates |

- **Ownership filtering**: user-facing row queries carry `Where("user_id = ?", claims.ID)`; read then verify before mutating (`usr.RemoveMedia`: fetch, compare `v.UserID != claims.ID` → 403, then delete).
- **Passwords**: write with `pwd.Hash` / `pwd.MustHash`, verify with `pwd.Verify` (argon2id only; there is no plaintext fallback).
- **Clearing to NULL**: pass `gorm.Expr("NULL")` in the values map (see `usr.UnbindPlatform`); passing Go `nil` for a pointer column works too.

### List queries (sorting + pagination)

A query DTO embeds `orms.Sorter` + `orms.Pagination`; the handler parses it, then chains `Scopes(q.Sorter.Scope)` + `IPaginate(q.Pagination)`:

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

- `IPaginate` returns `orms.Page[T]` (`{total, page, content}`); page/size are sanitised (size defaults to 20, capped at 100) and `Sorter.Scope` validates column names against injection.
- Each endpoint gets a same-shaped named Page DTO (`AdmUsersPage`…) purely to name the swagger schema; the handler returns the `IPaginate` result directly.
- A single-row query maps the row onto a flat DTO: `orms.QM[Profile, db.User](db.Tx).IFirst(claims.ID)`.
- Simple count: `db.Tx.Model(&X{}).Count(&n)`.

### Fiber handler

```go
var jwt = fext.NewJWT(conf.App.SecretAdm)

func Setup(api fiber.Router) {
	api.Group("/auth").
		Post("/login", Login).
		Post("/refresh", Refresh).
		Post("/logout", Logout)
	api.Group("/profile", jwt.Middleware).
		Get("/me", FindProfile).
		Put("/me", UpdateProfile)
}
```

- `/api/v1` (`std`, public) has no JWT; `/adm/api/v1` + `/usr/api/v1` each use their own `jwt.Middleware` (`secret.adm`/`secret.usr`). Groups are mounted in `srv.Setup()` by `internal.Log("GEN"/"ADM"/"USR"/"DOC")`.
- Input: bodies via `fext.BodyVarser[T](c)` (bind + validate against the `validate` tags); queries via `fext.QueryParser[T](c)`; path IDs via `fext.IdentityParam[uint]` (``ID uint `param:"id"` ``) or `fext.ParamsParser`/`FormData`. Every parse failure returns `&fext.Fail{Code: internal.InvalidData, Message: err.Error()}`.
- Errors: return `&fext.Fail{Code, Status, Message, System}`; `fext.ErrorHandler` renders `{code, message}` (status defaults to 400; `System` only appears in the response in debug mode). Error-code constants live in `internal/srv/internal/txtcode.go`. Success is `return c.JSON(v)`; an empty 200 is a bare `return nil`.
- Identity: `jwt.Claims(c)` returns `*jwt.RegisteredClaims`, whose numeric subject is `claims.ID` (a **string**) — convert with `strs.ParseUint[uint]` / `strs.FormatUint`. On unauthenticated requests `claims` may be `nil`, so only call it behind `jwt.Middleware`.

### JWT and refresh rotation (`internal/auth`)

One `authMgr` per app:

```go
var authMgr = auth.New(conf.App.SecretUsr, "usr",
	time.Duration(conf.App.AccessMinutes)*time.Minute,
	time.Duration(conf.App.RefreshHours)*time.Hour)
```

`GenPair(userID)` → `Pair{AccessToken, RefreshToken}` (the refresh token is recorded in Redis for revocation, namespace `usr:refresh:<jti>`). `Rotate` redeems it exactly once, keeping a 30 s `usr:refresh:used:<jti>` grace marker so two tabs racing on the same cookie both succeed; `RevokeRefresh`/`RevokeAll` clear both keys.

Transport: the access token is returned in the response body (`Auth{ID, AccessToken}`) and the refresh token is an **HttpOnly cookie** (`usr_refresh` / `adm_refresh`, `Path=/{usr,adm}/api/v1/auth`, `SameSite=Lax`, `Secure` behind TLS, `SameSite=None` only for the cross-site Tauri webview). `/auth/refresh` and `/auth/logout` take **no body** and read the cookie; logout is idempotent (always 200 + a clearing `Set-Cookie`). Access tokens are deliberately not revoked — they expire after `jwt.access-minutes`.

CSRF: `internal.OriginGuard` runs on both `/auth` groups and rejects a request whose `Origin` is neither in `[cors]` nor same-origin with the request host. CORS sets `AllowCredentials` (origins are still matched exactly).

### Media upload (`internal/srv/internal/media.go`)

Shared by `adm`/`usr`: `internal.ParseUploadFile(c)` → `kind := internal.Kind(c.FormValue("kind"))` (`image`/`video`/`file`, default file) → `internal.SaveFile(c, file, kind, userID)`. `kind` is validated by extension plus **MIME sniffing** (`KindFile` rejects XSS-capable MIME types). Files land at `<RMedia>/<YYYY>/<MM>/<rands.Char(20)><ext>` with the row's `URL` = `conf.App.MediaBase() + "/" + path`; a failed DB write deletes the file, and deleting the row deletes the file via `Media.AfterDelete`.

### OAuth (`internal/oauth`, `internal/srv/usr/auth.go`)

- Google: `oauth.GoogleExchange(code)` → `client.GoogleUser`; on `gorm.ErrRecordNotFound`, `createGoogleUser` provisions the account automatically (`oauthUsername` derives a unique username, random password, nullable `GoogleInfo` pointer).
- Telegram: bind `oauth.TelegramAuth` (the Login Widget payload) as the body DTO, verify it with `oauth.CheckTelegramAuth` (HMAC + 5-minute freshness), then `db.FindUserByTelegram` / `createTelegramUser`.
- Bind/unbind endpoints live at `/usr/api/v1/profile/bind/...` and reuse the same exchange, clearing the column with `gorm.Expr("NULL")`.

### Configuration and runtime (`internal/conf/`)

Everything goes through `conf.App` (`conf.App.Port`, `conf.App.MediaBase()`…); Viper keys map to `conf.toml` by dotted path (`db.name` → `[db] name =`). `conf.init()` calls `LoadFlag()` + `LoadApp()` **at package import time**, before `main()`; `Validate()` enforces the `validate:"required"` keys. `vipers.Watch(conf.Reload, db.Reload)` hot-reloads. Required values use `vipers.String(...)`; optional ones use `vipers.Get("app.port", "8080")` / `vipers.StringSlice(...)`.

### Scheduled tasks (`internal/task/`)

Six-field cron (seconds first): `schedule.AddFunc("0 0 0,12 * * ?", Rate)`. The rate sync is a full atomic rebuild: fetch everything externally first (return without touching the DB on failure), then in one transaction `DELETE FROM rates` + `Create` each row. **Use `DELETE`, never `TRUNCATE`** — the latter is DDL, and its implicit commit destroys the transaction.

### CLI flags (`internal/flag/`)

Flags are defined in `internal/conf/flag.go` and dispatched in `internal/flag/flag.go` by `flags.RunE(flag, script)` (a hit runs the script, then exits 0 — or 1 after printing the script's error to stderr). `PreRun()` runs before `db.Connect()` (`--gensec`); `Run()` runs after connecting: `--adm`/`--usr` (interactive account creation), `--mock` (admin/123456, user/123456), `--sync` (run the cron task once). Every script in `internal/flag/runner.go` returns `error`; there is no `os.Exit`/`fail()` in this repo's own code.

### Backend conventions (beyond lint)

- `orms.UpdatesByID` / `Updates` skip zero values when passed a struct — pass a `map[string]any` when you need to clear/reset.
- Never put raw user input into `Order` — go through `Sorter.Scope` only.
- After swagger/SDK changes run `./main.sh docs`; for backend changes run `./main.sh lint --fix` first, then `./main.sh lint`.

## Frontend specifics

- **Package manager:** pnpm@12.21.0 (enforced in `web/package.json` `packageManager`)
- **Workspace packages:** `apps/*` (Nuxt apps), `packages/*` (build-time tooling — `apiclient`, the SDK codegen, and `sdk`, the `@sdk/client` runtime + generated SDK both apps import), `packages/natholdallas/*` (shared Nuxt modules, git submodule)
- **Prod build:** `pnpm generate` (SSG / static generation via `nuxt generate --dotenv .env.production`), not `nuxt build`. Root script `pnpm gen` (`pnpm -F adm -F usr --parallel generate`) runs both apps' SSG in parallel — distinct from `pnpm gen:api` (SDK generation)
- **SSR disabled:** both apps set `ssr: false`
- **Formatting:** `pnpm format` (Prettier) — run from `web/` dir. Config: no semis, single quotes, 120 print width, 2-space indent
- **ESLint:** auto-generated by Nuxt (`.nuxt/eslint.config.mjs`), imported by root config. Check per-app directory.
- **Lint:** after writing frontend code, run `pnpm lint` (or `pnpm lint:fix` to autofix) from `web/` — the root script is `pnpm check:locale && pnpm -r --no-bail run lint`, so it first checks the locale files (see "Locale files" below) and then runs each app's own `lint`/`lint:fix` (`eslint .` / `eslint . --fix`) across every app under `web/apps/` (usr, adm); packages without a lint script are skipped. **All `warn` and `error` findings must be fixed before finishing** (run `pnpm lint` after writing code, fix every warning/error — `pnpm lint:fix` for auto-fixables, then manually fix the rest; re-run until clean). Each app has its own `eslint.config.mjs` importing that app's generated `.nuxt/eslint.config.mjs` (usr ignores the vendored `app/components/ui`; adm ignores `src-tauri/**` and disables `vue/valid-v-slot`, because Vuetify's static dotted slot names such as `item.actions` are misread as modifiers); `.nuxt/` must exist first (`postinstall` runs `nuxt prepare`).
- **Env vars:** `NUXT_PUBLIC_API_BASE` (backend URL), `NUXT_PUBLIC_SITE_URL`, `ENABLE_PWA`, `ENABLE_SEO`
- **Auth state:** each app's `useAuthStore` keeps `{ id, accessToken }` **in memory only** (the refresh token is the HttpOnly cookie described above; the legacy `localStorage` entry is deleted on store creation). `$restore()` trades the cookie for an access token once per page load (never during `nuxt generate`), `$refresh()` shares one in-flight request, `$clear()` drops state without navigating, and `$signOut(skipServer?)` clears first, then revokes server-side. A single `app/middleware/auth.global.ts` guards every route: public pages are `entrance` and `oauth-google`, everything else calls `$restore()` and redirects to `entrance` — there are no per-page `middleware: 'auth'` metas.
- **Env files:** `.env` (dev) from `assets/nuxt.env`, `.env.production` (prod) from `assets/nuxt.env.production`, copied by `./main.sh copyfile` **only when absent** (`copyfile --force` overwrites; it never clobbers local overrides silently). `deploy` **refuses** to ship a bundle whose `NUXT_PUBLIC_API_BASE` is still `localhost`/`example.com` unless `--allow-placeholder-env` is passed; set the override to `same-origin` for a single-host (reverse-proxied) deployment.
- **Shared packages:** git submodule at `web/packages/natholdallas` → `https://github.com/natholdallas/nuxt-modules.git`
- **Test utils:** configured via `.nuxtrc`: `setups.@nuxt/test-utils="4.2.0"` — test with `nuxi test` from app dir
- **Tauri v2 shells:** every app carries `src-tauri/` (crate `webgx` / `webgx-adm`, identifier `com.webgx.<app>`); `tauri.conf.json` uses `frontendDist: ../dist/public`, `devUrl` = that app's port, `beforeDevCommand: pnpm dev`, `beforeBuildCommand: pnpm generate`. Scripts: `pnpm dev:tauri`, `pnpm package` (release), `pnpm package:debug`. `@tauri-apps/cli` is a devDependency of each app, and the `@natholdallas/tauri` module supplies the build-time wiring.
  - The backend must allow the webview origins in `[cors]`: `tauri://localhost` (Linux/macOS/iOS), `http://tauri.localhost` (Windows), `https://tauri.localhost` (Android with `useHttpsScheme`). Matching is exact, so they are listed verbatim in `assets/conf.toml`/`conf.toml`.
  - `window.open` is a no-op inside the webview: adm's `useOpener().openExternal()` routes through `tauri-plugin-opener` in the shell and falls back to a browser tab on the web. usr hides Google/Telegram OAuth in the shell (`useIsTauri()`) because those flows need a real browser and a web-origin callback.
  - Android projects live in `src-tauri/gen/android` (committed). Cleartext HTTP is enabled for **debug builds only** via `manifestPlaceholders["usesCleartextTraffic"]` in each generated `app/build.gradle.kts`; release stays `false`. Build/run with `ANDROID_HOME` set: `pnpm exec tauri android dev` (add `TAURI_DEV_HOST=<lan-ip>` so the device reaches the dev server) or `pnpm exec tauri android build --apk --target aarch64`.
  - iOS needs macOS + Xcode (`pnpm exec tauri ios init` there); the shared `tauri.conf.json` is already iOS-compatible.
  - `src-tauri/target/` and `gen/schemas/` stay gitignored; `Cargo.lock`, `icons/` and `gen/android` are committed.

### Locale files

Every app owns flat, dotted translation keys in `web/apps/<app>/app/locale/<lang>.ts` (`zh_cn.ts`, `en_us.ts`; add `ja_jp.ts`/`zh_tw.ts` only when the language is also added to that app's `i18n.locales`). Each file imports **and spreads** the shared base dict first, then declares app keys:

```ts
import zh from '@natholdallas/i18n/locale/zh_cn'

export default {
  ...zh,

  'app.name': 'webgx',
}
```

- Never add app keys to `web/packages/natholdallas/i18n/locale/*.ts` — that is the submodule's global base set shared by every project (changes belong in `nuxt-modules.git`). Shared keys already cover `create`, `cancel`, `dialog`, `err.network`, ... . Add the key to each app locale file that uses it.
- **All languages of one app must define the same keys in the same order**; `zh_cn` is the reference (else the first file alphabetically).
- **Check:** `pnpm check:locale` from `web/` (`node scripts/check-locale.mjs`; runs as the first step of `pnpm lint`). Per app it reports keys missing/extra vs the reference, order divergence, keys defined but unreferenced, keys referenced/implied by source but missing (shared base keys count as available), a locale file that no longer spreads its shared base, and translated text passed to Vuetify props that expect a **key**. It also checks the shared base dict itself for missing/extra keys between its four languages (its key order is not enforced — it is maintained in the submodule). `--delete-unused` removes the unused entries from every language file of that app; pass app names (`pnpm check:locale adm`) to narrow.
- Source scan: app `app/**` (locale dir and generated SDK excluded), plus `internal/**/*.go` and `web/packages/sdk/src` for keys referenced by backend/SDK error codes. `$t('key')`, `t('key')`, `k:`/`label:`/`title:`/... constants that feed `$t(x.k)`, and `` $t(`a.${b}`) `` prefixes count as uses.
- **Adding an app needs no script edit:** the checker auto-discovers directories under `web/apps/`. The app must use the template's `restructureDir: 'app'` layout (`app/locale/<lang>.ts`), and its locale filenames must match the shared base filenames (`en_us`, `zh_cn`, `ja_jp`, `zh_tw`) for the spread check to line up. An app with one locale file skips cross-language checks but still gets unused/missing checks.
- **Vuetify key props:** Vuetify 4 resolves some prop values through the app's vue-i18n instance (`no-data-text`, `loading-text`, `items-per-page-text`, `title` on `VFileUpload`, `label` on `VBadge`/`VOtpInput`, ...). Pass the **key**, not the translation: `no-data-text="no.data"`, **not** `:no-data-text="$t('no.data')"` (the translated value becomes the lookup key and logs `Not found '暂无数据' key in 'zh' locale messages`). The checker derives this prop list from the installed Vuetify lib and scopes it per component family, so ordinary `:title`/`:label` props are not flagged.


## Shared Nuxt modules (web/packages/natholdallas/)

Every package under `web/packages/natholdallas/` is a **Nuxt module** (each exposes a `defineNuxtModule` in its `index.ts`). They are **optional** — an app only loads the ones it lists in its `nuxt.config.ts` `modules` and declares as `@natholdallas/*` `workspace:*` deps. This directory is a **git submodule** (shared across projects) pointing at `nuxt-modules.git`. Changes are committed inside the submodule and the pointer is bumped in this repo — keep both commits together so a checkout is never left with a dangling pointer. Two rules keep apps healthy:

- **Modules must not assign `runtimeConfig.public.*`.** Module setup runs after `nuxt.config.ts`, so doing so silently overrides the app. `apiBase` is declared by each app and fed by `main.sh` from `domain_api`/`app_env`.
- **Modules declare everything they import.** Packages used only by app-level components belong in the app; otherwise apps end up relying on hoisted phantom dependencies.

- **`alova`** — registers the alova runtime auto-imports (`lib/`: `useRequest`, `useWatcher`, `Method`, ...). The instance and event handling live in `@sdk/client` (`web/packages/sdk/src/runtime.ts`), not here. Used by both apps.
- **`i18n`** — wraps `@nuxtjs/i18n`. Ships the **globally stored shared translation keys** in `locale/*.ts` (`zh_cn`, `en_us`, `zh_tw`, `ja_jp`). **Do NOT modify these** — they are the global key set. To change or add translations, edit the **project's own i18n files** (`web/apps/<app>/app/locale/*.ts`), which import and spread the shared dict (e.g. `import zh from '@natholdallas/i18n/locale/zh_cn'` then `...zh`) and layer app-specific keys on top. The submodule locales should stay untouched so every project gets the same base keys.
- **`infra`** — bundles the common infra modules: `@nuxt/icon`, `@nuxtjs/device`, `@vueuse/nuxt`, `dayjs-nuxt`, `nuxt-og-image`, `@nuxt/eslint`, `@nuxt/test-utils`. `@nuxtjs/seo` and `@vite-pwa/nuxt` are installed **only when `ENABLE_SEO=true` / `ENABLE_PWA=true`** (`installModule` in setup): @vite-pwa/nuxt 2 dropped its `disable` option, and the SEO umbrella's prerender pipeline breaks an SPA/SSG build when it is installed but disabled. Sets typed pages and dayjs defaults; bundle analysis is opt-in via `ANALYZE=true`, and console/debugger calls are dropped through `vite.esbuild` (Nuxt minifies with esbuild, not terser). It does **not** touch `runtimeConfig.public.apiBase` — each app declares that itself. Does **not** force `future.compatibilityVersion` or `experimental.viteEnvironmentApi` (that is the app's Nuxt-5 opt-in). Also auto-imports shared components (`PwaProvider`) and composables (`crud`, `routes`, `interval`). Used by both apps (usr via `shadcn`, adm via `vuetify`).
- **`pinia`** — wraps `@pinia/nuxt`; enables Pinia stores (`useAuthStore`, ...). Used by both apps (the SDK's `beforeRequest` reads the access token from `useAuthStore()`).
- **`shadcn`** — the shadcn-vue module (usr app): depends on `@natholdallas/i18n`, bundles `shadcn-nuxt` + Tailwind v4 (`@tailwindcss/vite`) + reka-ui + vee-validate/zod, and registers `Uix`-prefixed module components (`Form`, `DataTable`, `Field`, `Modal`, ...). The vendored `app/components/ui/*` primitives are copied into the app, so optional packages (`@unovis/vue`, `embla-carousel-vue`, `vue-input-otp`, `@internationalized/date`) are declared by the **app** that actually uses them.
- **`tailwindcss`** — Tailwind v4 CSS-first module (postcss/vite plugin + sass), for apps that don't use the shadcn/vuetify bundles. Optional, not currently used.
- **`tauri`** — Tauri v2 build wiring for both apps' `src-tauri/` shells: ignores `**/src-tauri/**` in Nuxt/Vite, adds the `TAURI_` env prefix, sets `vite.server.strictPort` (so `devUrl` in `tauri.conf.json` cannot silently drift), and maps `TAURI_DEV_HOST` to `devServer.host` for device testing. It never touches `runtimeConfig`.
- **`unocss`** — UnoCSS module (wraps `@unocss/nuxt` + presets, ships an example `uno.config.ts`). Optional, not currently used.
- **`vuetify`** — the Vuetify module (adm app): depends on `@natholdallas/i18n` + `@natholdallas/infra`, wraps `vuetify-nuxt-module` with theme/component defaults, wires Tailwind postcss, and registers `Vx`-prefixed module components (`Form`, `Dialog`, `Upload`, `Drawer`, ...). It reads the app's `app/assets/styles/sass/settings.scss` at that fixed relative path. Icons come from `@mdi/font` (a dependency of adm) so `vuetify-nuxt-module` imports the font CSS locally instead of the jsDelivr CDN — important for the Tauri shell.
- **`watermark`** — adds a `Watermark` component + runtime config. Optional, not currently used.

## Swagger → Frontend SDK (alova/wormhole)

The frontend SDKs are **generated** from the backend swagger doc, not hand-written. Flow: `swag init` produces `docs/swagger.json` → `web/packages/apiclient/alova.config.ts` runs `@alova/wormhole` → emits both apps' typed SDK into the shared `web/packages/sdk/src/{usr,adm}/` (gitignored, no `gen/` subdir).

**`web/packages/apiclient`** is **not a Nuxt module** — it's the build-time SDK generator. It holds the `@alova/wormhole` config (`alova.config.ts`) plus the custom codegen plugins (`codegen/plugins.ts`) that turn `docs/swagger.json` into the typed API SDK of each app (`web/packages/sdk/src/{usr,adm}/`). It also owns the plugins that emit the positional facade and the package entry (`codegen/emitApi.ts`). It's only ever driven by `pnpm gen:api` (`pnpm --dir packages/apiclient exec alova gen -f`) at build time; app code never imports it at runtime.

- **Entrypoint:** `./main.sh docs` regenerates swagger **and** the SDK. `./main.sh build` also runs `pnpm gen:api` before `pnpm generate`.
- **Config:** `web/packages/apiclient/alova.config.ts` — two generators (usr → `/usr/api/v1`+`/api/v1`, adm → `/adm/api/v1`+`/api/v1`), driven by plugins in `web/packages/apiclient/codegen/` (`plugins.ts` renames + `modelDefaults` + `exposeApis`, `emitApi.ts` emits `api.ts` and `index.ts`). The root `pnpm gen:api` runs it via `pnpm --dir packages/apiclient exec alova gen -f`.
- **Generated files** per app (`web/packages/sdk/src/{usr,adm}/`, structurally identical in both): `apiDefinitions.ts` (`<METHOD> <url>` per endpoint), `globals.d.ts` (wormhole's types; `exposeApis` rewrites its `declare global` block into `export interface Apis`, so nothing is global), `models.ts` (`type X = G.X` re-export from `globals` + `const X = {...}` model factories; nullable response fields emit `null`, and the file is written only after a successful generation), `api.ts` (the positional facade `api.<group>.<verb>(pathArg, data, params, config?)`, emitted by `codegen/emitApi.ts`), `index.ts` (re-exports `models`, `api` and the runtime's `alovaInstance`/`$$userConfigMap`; wormhole only writes `index.ts` when it is absent, so it is rewritten on every run). `createApis.ts` is rewritten by `codegen/emitApi.ts` to a forwarding stub (wormhole unlinks the file name it recorded on the next run) — the proxy builder is hand-written in `web/packages/sdk/src/createApis.ts`.
- **App usage:** `import { api, Model } from '@sdk/client/usr'` (adm: `@sdk/client/adm`). Call the facade with positional arguments — `api.user.update(id, { username })`, `api.media.list({ page, size })`, `api.auth.refresh({ meta: { isRefresh: true } })` — and keep `config` (meta/headers/transform) as the last optional argument. Do not re-introduce the `Apis` proxy in app code.
- **Runtime behaviour** (`web/packages/sdk/src/runtime.ts`, hand-written and committed): one alova instance plus the event table (200 → `ok`, 401 → `unauthorized`, everything else → `fallback`). `fetch` uses `credentials: 'include'` so the refresh cookie travels; `baseURL` and the Bearer token come from the injected hooks. Any unmapped 2xx → success handler; 401 → `$refresh()` at most once per failed token, replaying via `method.send(true)` (concurrent callers reuse the same in-flight refresh; a second failure marks the key and calls `$signOut(true)`); `meta.isRefresh` → `$clear()` and throw; `meta.noRefresh` → throw (logout probes never recurse); other statuses and rejected fetches → `fallback`/`onError`, which report `code`/`message`/`err.network` unless `meta.silent` and throw an `Error` carrying `status`. The package never imports Nuxt/Pinia composables — each app installs them via `app/plugins/sdk.ts` → `setupSdk({ baseURL, auth, t, notify })` (usr → `vue-sonner`, adm → `useSnackBar()`), which runs before the auth middleware restores the session.

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

- **`./main.sh renewal <name> [--no-backup]`** — renames the entire project. Refuses to start on a dirty worktree, bundles git history (`<pname>-pre-renewal-<ts>.bundle`) and `conf.toml` next to the repo (**`--no-backup` skips both**: nothing is recoverable and `[db] name` follows the rename), then replaces all `webgx` strings (text files by suffix, plus `*.lock`, **and directories** such as the Android package path `com/<pname>/…`), deletes `.git`, removes/re-adds submodule, inits a fresh git repo, and keeps `[db] name` in `conf.toml` so the existing database stays reachable. Prunes build output (`target`, `build`, `.gradle`, `.kotlin`, `dist`, `.nuxt`, …). Copies env files from `assets/`. **Irreversible** — prompts for confirmation.

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
- The generated SDK (`web/packages/sdk/src/usr`, `web/packages/sdk/src/adm`) and the swagger spec (`docs/`) are both **gitignored** — always regenerate via `./main.sh docs`, never commit them. The hand-written runtime next to them (`web/packages/sdk/src/runtime.ts`, `createApis.ts`) **is** committed.
- `swag` is pinned to **v2.0.0-rc5** in `go.mod` (`github.com/swaggo/swag/v2`) and installed by `./main.sh init` via `go install github.com/swaggo/swag/v2/cmd/swag@v2.0.0-rc5`. If `swag --version` differs, reinstall with the v2 command above
- `stripSchemaPrefix` only strips names whose remainder starts uppercase, so the shared `Admin` model (whose name happens to start with the `Adm` prefix) is preserved; don't introduce app DTOs whose stripped name would collide with a shared model used in the same app
- Every new DB model must have **complete gotags** (`gorm:"column:...;comment:..."` + `json` tag) and a trailing comment on **every field** — see the "GORM model conventions" section above
