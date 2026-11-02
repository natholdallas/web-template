---
name: Exp
description: The web-template project's backend code patterns and unique use cases — GORM model conventions (orms Model/IDModel/Dict/List, embedded prefixed structs, nullable pointers), generic orms CRUD helpers, sorting/pagination list queries, Fiber handler + swagger annotation conventions, JWT auth with Redis-backed refresh rotation, media upload, OAuth (Google/Telegram) login, cron scheduling, config access, and CLI flag migration scripts. Load this when writing backend code that touches internal/db, internal/srv, internal/task, internal/auth, internal/oauth, or internal/conf so new code matches the existing style instead of inventing new shapes.
---

# Exp — Backend Code Patterns

Work through this skill whenever you write or modify backend code in this
project (Go, Fiber v3 + GORM). It collects the unique code shapes used across
`internal/`, distilled from real code — write new code that mirrors these
patterns instead of inventing your own.

**Scope:** this skill covers the generic backend stack (`internal/db`,
`internal/srv`, `internal/task`, `internal/auth`, `internal/oauth`,
`internal/conf`, `internal/client`). Repo-wide rules (English-only comments, no
`_ =` blank assignments, Go tabs, file ordering) are in AGENTS.md and still
apply.

## 1. GORM model conventions (`internal/db/`)

### Base models

Embed a generic base model instead of hand-writing ID/timestamps:

- `orms.Model[uint]` — ID + `created_at`/`updated_at` (most rows: `Admin`,
  `User`, `Media`, `SysConf`)
- `orms.IDModel[uint]` — ID only, no timestamps (rows with no meaningful
  timestamps: `Rate`, a cache table rebuilt every 12h)
- Also available in `orms` but **not used here**: `orms.SoftModel[T]` (adds
  `deleted_at`) — do not introduce soft delete unless asked

### Full gotags on every field, trailing comment on every field

Every field declares both a `gorm` tag (starts with `column:<snake_case>`,
`size:<n>` on strings, `comment:<English description>`, plus constraints) and a
`json` tag (`json:"-"` when never exposed). A trailing `// <description>`
comment mirrors the gorm comment. The struct itself ends with the
`//	@name	<Name>` swagger annotation (app-owned DTOs get a `Usr`/`Adm` prefix,
shared/global models stay unprefixed).

```go
type User struct {
	orms.Model[uint]
	Username string `gorm:"column:username;size:50;unique;comment:Username" json:"username"` // Username
	Password string `gorm:"column:password;size:255;comment:Password" json:"-"`              // Password
} //	@name	User
```

### Embedded prefixed structs (the association pattern in this project)

Many-to-one struct embedding (embedding a referenced row + FK column) is **not
used here**. When a row carries a group of related columns, model them as a
plain struct embedded with a column prefix (see `internal/db/user.go`):

```go
type User struct {
	orms.Model[uint]
	Username     string       `gorm:"column:username;size:50;unique;comment:Username" json:"username"` // Username
	GoogleInfo   GoogleInfo   `gorm:"embedded;embeddedPrefix:google_" json:"-"`                        // Google login info
	TelegramInfo TelegramInfo `gorm:"embedded;embeddedPrefix:telegram_" json:"-"`                      // Telegram login info
} //	@name	User

// GoogleInfo holds the Google login bindings (columns prefixed google_).
type GoogleInfo struct {
	Email *string `gorm:"column:email;unique;size:256;comment:Google email" json:"email"` // Google email
} //	@name	GoogleInfo
```

Rules: the embedded struct gets its own `// @name` annotation; fields that can
be absent are nullable pointers; `unique`/`index` constraints live on the inner
columns; the outer struct marks the embedded field `json:"-"` and exposes the
flattened columns instead (the frontend sees `google_email` etc.).

### JSON-persisted fields — `orms.List[T]` / `orms.Dict[T]`

Not used in the template yet, but available: a `Dict`'s (or `List`'s) column
type **must be `json`** — the generic Scan/Value pair JSON-serializes any
round-trippable element type `T`.

```go
Modules orms.List[string] `gorm:"column:modules;type:json;comment:Enabled modules" json:"modules"` // Enabled modules
```

### Nullable columns are pointers

A column that can be absent in the API is a pointer type: `*string`, `*int`,
`*time.Time`. Examples: `GoogleInfo.Email *string` (no Google binding yet),
`TelegramInfo.ID *string`. When writing, drop the pointer (`values["col"] =
nil`) or use `gorm.Expr("NULL")`; when reading, nil means "not set".

### Registration

Register new models in `internal/db/db.go` → add to `var Models = []any{...}`.
This single list drives `db.Migrate()` (`orms.MustAutoMigrate`), `--db-sync`
(`orms.SyncDB`), and `--db-reset-table` (`orms.ResetTables`) — the model set is
the single source of truth for the schema.

## 2. `orms` generic CRUD helpers

Prefer the generic helpers over hand-rolled GORM calls. They all take the
`*gorm.DB` session as the first argument — pass `db.Tx`, or `t` inside a
transaction.

| Helper | Use |
|---|---|
| `orms.First[T](tx, conds...)` | one row by PK or arbitrary conds: `orms.First[User](tx, "username = ?", username)` |
| `orms.Take[T]` / `orms.Last[T]` | one row without order assumptions |
| `orms.Find[T](tx, conds...)` | a list of rows |
| `orms.IFirst[T]` / `ITake` / `ILast` / `IFind` | same, but swallow the error and return a zero value |
| `orms.Create[T](tx, v)` | insert (`&v` to get the ID back) |
| `orms.Save[T](tx, v)` | upsert by PK |
| `orms.Delete[T](tx, conds...)` | delete by PK or conds |
| `orms.UpdateByID[T](tx, id, column, value)` | single-column update |
| `orms.UpdatesByID[T](tx, id, values)` | multi-column update; **pass a map** so zero values (`0`, `false`, `""`) actually write |
| `orms.QE[T](tx)` | fluent wrapper with T as the model: `.QE[Media](tx).Where("user_id = ?", id).Scopes(...).IPaginate(...)` |
| `orms.Q[T](tx)` / `orms.QM[T, M](tx)` | fluent wrapper without model / mapping model M into result type T |
| `orms.Transaction(tx, func(t *gorm.DB) error)` | wrap multi-step writes |
| `orms.Count[T]` / `orms.Exists[T]` / `orms.PluckStrings[T]` | quick aggregates |

Conventions around them:

- **Ownership scoping.** Every handler-facing finder that touches a user-owned
  row filters by the caller: `Where("user_id = ?", claims.ID)`, and
  read-then-verify before mutating (`usr.RemoveMedia`: fetch, check
  `v.UserID != claims.ID` → 403, then delete).
- **Passwords.** Hash with `pwd.Hash`/`pwd.TryHash`, verify with `pwd.Verify`
  (argon2id with transparent legacy-plaintext upgrade — `db.AuthUser`/
  `db.AuthAdmin` also re-hash and persist the plaintext match).
- **Clear-to-NULL.** To null a column, pass `gorm.Expr("NULL")` in the values
  map (see `usr.UnbindPlatform`); a Go `nil` inside a `map[string]any` also
  works for pointer columns.

## 3. Query patterns

### List endpoints: sorting + pagination (the canonical list shape)

Queries DTOs embed `orms.Sorter` + `orms.Pagination`; the handler parses them
and chains `Scopes(q.Sorter.Scope)` + `IPaginate(q.Pagination)`:

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
	v := orms.QE[db.User](db.Tx).
		Scopes(q.Sorter.Scope).
		IPaginate(q.Pagination)
	return c.JSON(v)
}
```

- `IPaginate` returns `orms.Page[T]` (`{total, page, content}`); page/size are
  sanitized (size defaults 20, capped 100) and `Sorter.Scope` validates the
  column name against injection (`validColumn`).
- Declare a per-endpoint named Page DTO (`AdmUsersPage`, `UsrMediaPage`, ...)
  with the **same shape** (`Total`/`Page`/`Content`) purely for swagger naming
  (`// @name AdmUsersPage`) — the handler returns the `IPaginate` result
  directly since the JSON shapes match.

### Row → DTO mapping (`QM`)

Map a db model into a flat response DTO in one query:

```go
v := orms.QM[Profile, db.User](db.Tx).IFirst(claims.ID)
return c.JSON(v)
```

`Profile` is a plain DTO with `json` tags only (no gorm tags needed); `QM[T, M]`
sets `M` as the GORM model and scans into `T`.

### Filter accumulation & user scoping

Build the query layer by layer on the fluent wrapper, adding a filter only when
needed:

```go
v := orms.QE[db.Media](db.Tx).
	Where("user_id = ?", claims.ID).
	Scopes(q.Sorter.Scope).
	IPaginate(q.Pagination)
```

### Simple counts

Use `db.Tx.Model(&X{}).Count(&n)` for quick aggregates (see `adm.GetStats`).

## 4. Fiber handlers (`internal/srv/{std,usr,adm}`)

### Route registration

Groups carry the middleware; auth endpoints live on a bare group:

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

- `/api/v1` (`std`, public) — no JWT. `/adm/api/v1` + `/usr/api/v1` — protected
  by their own `jwt.Middleware` (secrets `secret.adm` / `secret.usr`).
- Groups mount with a log prefix in `srv.Setup()`: `internal.Log("GEN"/"ADM"/"USR"/"DOC")`.

### Parsing inputs

- Body: `fext.BodyVarser[T](c)` — binds **and validates** (`validate` tags on
  the DTO). Return `&fext.Fail{Code: internal.InvalidData, Message: err.Error()}`
  on error.
- Query: `fext.QueryParser[T](c)` — binds query params (incl. embedded
  `orms.Sorter`/`orms.Pagination`) + 400 `InvalidData` on malformed input.
- URI params: `c.Params("id")` directly, or bind into a DTO with
  `fext.ParamsParser[T]` / `fext.FormData[T]` (URI+query+body in one pass);
  `fext.IdentityParam[uint]` (`ID uint \`param:"id"\``) is the standard embed
  for path IDs.

### Errors

Handlers return `*fext.Fail{Code: <const>, Status: <http>, Message: <msg>,
System: <err>}` and let `fext.ErrorHandler` render `{code, message}` with the
status (default 400). `System` is logged via narder and only surfaced in the
body when debug mode is on. Error-code constants live in
`internal/srv/internal/txtcode.go` (`InvalidData`, `CreateFailed`,
`SignInFailed`, `RecordNotFound`, `OauthFailed`, `UploadFailed`,
`InvalidMedia`, ...). Success is `return c.JSON(v)` or plain `return nil` for
empty 200s.

### Middleware identity

`jwt.Claims(c)` returns `*jwt.RegisteredClaims`; the numeric subject is
`claims.ID` (a **string**) — convert with `strs.ParseUint[uint](claims.ID)`,
format back with `strs.FormatUint`. `claims` may be `nil` on unauthenticated
requests, so only call it behind `jwt.Middleware`.

### JWT + refresh rotation (`internal/auth`)

Each app gets one `authMgr` built from its secret and configured TTLs:

```go
var authMgr = auth.New(conf.App.SecretUsr, "usr",
	time.Duration(conf.App.AccessMinutes)*time.Minute,
	time.Duration(conf.App.RefreshHours)*time.Hour)
```

`authMgr.GenPair(userID)` → `Pair{AccessToken, RefreshToken}` (refresh token
recorded in Redis for revocation, namespaced `usr:refresh:<jti>`).
`VerifyRefresh(token)` validates + checks Redis; `RevokeRefresh(token)`
invalidates. Refresh tokens are rotated on every `/auth/refresh` (revoke old,
issue new). The `jti` random comes from `rands.Char(32)`.

### Media upload (`internal/srv/internal/media.go`)

The shared upload pipeline, used by both `adm` and `usr`:

```go
file, err := internal.ParseUploadFile(c)          // multipart "file" field
if err != nil {
	return err
}
kind := internal.Kind(strings.ToLower(c.FormValue("kind"))) // "image"/"video"/"file", default file
v, err := internal.SaveFile(c, file, kind, userID)          // validates + stores + creates db.Media
if err != nil {
	return err
}
return c.JSON(v)
```

- Kinds are validated against extension + **sniffed MIME** (content, not
  filename); `KindFile` blocks XSS-capable MIME types (`safeFileMIME`).
- Files land at `<RMedia>/<YYYY>/<MM>/<rands.Char(20)><ext>`; the row's `URL` is
  `conf.App.MediaBase() + "/" + path` and the file is removed if the DB write
  fails (and on delete via `Media.AfterDelete`).

### OAuth login (`internal/oauth`, `internal/srv/usr/auth.go`)

- **Google:** `oauth.GoogleExchange(code)` → `client.GoogleUser`; on
  `gorm.ErrRecordNotFound` from `db.FindUserByGoogle`, auto-provision via
  `createGoogleUser` (unique `username` derived by `oauthUsername`, random
  password, nullable `GoogleInfo` pointers).
- **Telegram:** bind `oauth.TelegramAuth` (the Login Widget payload) as the
  body DTO, verify with `oauth.CheckTelegramAuth` (HMAC signature + 5-min
  freshness), then `db.FindUserByTelegram` / `createTelegramUser`.
- Bind/unbind endpoints on `/usr/api/v1/profile/bind/...` reuse the same
  exchange + `gorm.Expr("NULL")` to clear columns.

### Swagger annotations (the ONLY manual cost of an API)

```go
// Login godoc
//
//	@Summary		Admin sign in
//	@Description	Sign in with username and password, returns access and refresh tokens
//	@Tags			admAuth
//	@ID				admSignIn
//	@Accept			json
//	@Produce		json
//	@Param			body	body		AuthIn	true	"Credentials"
//	@Success		200		{object}	Auth
//	@Failure		400		{object}	Fail
//	@Router			/adm/api/v1/auth/login [post]
```

Rules: `@ID` is globally unique (`usrSignIn`, `admListUsers`; std has **no**
prefix: `findRate`); `@Tags` carries the app prefix (`admAuth`, `usrMedia`;
std uses bare `Rate`); DTO `@name`s are prefixed (`AdmAuthIn`, `UsrProfile`);
`@Success`/`@Param` reference the **Go type name**; protected routes add
`@Security ApiKeyAuth`; uploads declare `@Accept multipart/form-data` + a
`formData` param. After changing annotations run `./main.sh docs` (swagger +
frontend SDK regeneration).

## 5. Config & runtime (`internal/conf/`)

- Read everything through the `conf.App` singleton (`conf.App.Port`,
  `conf.App.RedisAddr`, `conf.App.MediaBase()`, ...). Viper keys use dot
  notation matching `conf.toml` (`db.name` → `[db] name =`).
- `conf.init()` runs `LoadFlag()` + `LoadApp()` on **package import** — before
  `main()` — so config is ready for any consumer. `conf.Validate()`
  (`vipers.MustValidate`) enforces `validate:"required"` keys
  (`secret.adm`, `secret.usr`, `db.*`, `resources.web/log/cache`).
- `vipers.Watch(conf.Reload, db.Reload)` hot-reloads on file change;
  `conf.Reload` re-runs `LoadApp` and dumps the new config in debug mode.
- Mandatory values use `vipers.String(...)`; optional ones
  `vipers.Get("app.port", "8080")` / `vipers.StringSlice(...)`.
- `conf.toml` is gitignored — copy from `assets/conf.toml` via
  `./main.sh copyfile` before running.

## 6. Scheduled tasks (`internal/task/`)

Cron runs with seconds precision (6-field specs, seconds first):

```go
func Setup() {
	schedule := cron.New(cron.WithSeconds())
	schedule.AddFunc("0 0 0,12 * * ?", Rate) // 00:00 and 12:00 daily
	schedule.Start()
}
```

The rate sync shows the **atomic full-rebuild** pattern: fetch every currency
externally first (failing fast before touching the DB), then swap the table in
one transaction:

```go
if err := db.Tx.Transaction(func(tx *gorm.DB) error {
	if err := tx.Exec("DELETE FROM rates").Error; err != nil { // DELETE, not TRUNCATE:
		return err // TRUNCATE is DDL → implicit commit, defeats the transaction
	}
	for _, v := range all {
		if err := tx.Create(&v).Error; err != nil {
			return err
		}
	}
	return nil
}); err != nil {
	log.Info("caching rates failed: ", err)
}
```

## 7. CLI flags & migration layers (`internal/flag/`)

Flags are defined in `internal/conf/flag.go`, dispatched in
`internal/flag/flag.go` via `flags.Run(flag, script)` (runs the script when the
flag is set, then `os.Exit`). `PreRun()` runs before `db.Connect()`
(`--gensec` prints 32-char secrets to paste into `conf.toml`); `Run()` runs
after connect: `--adm`/`--usr` (interactive create), `--mock` (admin/123456,
user/123456), `--sync` (runs the cron task scripts once).

Schema evolution — all driven by the `db.Models` list:

| Flag             | Script                                        | Behavior                                                                                          | Data     |
| ---------------- | --------------------------------------------- | ------------------------------------------------------------------------------------------------- | -------- |
| `--db-migrate`   | `db.Migrate()` → `orms.MustAutoMigrate`       | Incremental, add-only (create missing tables/columns)                                             | Safe     |
| `--db-sync`      | `orms.SyncDB(db.Tx, db.Models...)`            | Reconcile per table: add missing, **drop extra**, reorder columns (MySQL in place, others rebuild) | Keeps    |
| `--db-reset-table` | `orms.ResetTables(db.Tx, db.Models...)`     | Drops + recreates each table from its struct                                                      | All lost |
| `--db-reset`     | `db.Reset()`                                  | Drops + recreates the whole database                                                              | All lost |

## Gotchas

- Never `_ = expr`; comments are English-only; tabs per `.editorconfig` —
  enforced by golangci-lint.
- `orms.UpdatesByID` / `Updates` with a struct skips zero values — pass
  `map[string]any` whenever zero is meaningful (reset/clear operations).
- Never pass raw user input into `Order` — always go through `Sorter.Scope`
  (validated), not `Order(c.Params(...))`.
- After any backend change: run `golangci-lint run --fix ./...` from the repo
  root; run `./main.sh docs` when swagger/SDK are touched (the generated SDK
  and `docs/` are gitignored — regenerate, never hand-edit).
- `conf.toml` is gitignored; `secret.adm` and `secret.usr` are validated as
  required — the app won't start without them (generate with `--gensec`).
