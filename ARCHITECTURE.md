# Gin API — architecture and operations

This repository is a **production-style** Go backend template: layered packages, explicit SQL migrations, JWT auth, bcrypt password hashing, strict request validation, modular routers per resource, rate limiting, a CI pipeline, and **Swagger UI**. The module path is `gin-api` (`go.mod`). For a public repo, rename to e.g. `github.com/you/gin-api` and update all imports.

**New here?** Read **[EXPLAINER.md](EXPLAINER.md)** first — same architecture explained in plain language and request order.

---

## API JSON envelope (all responses)

Every response uses the same top-level keys:

```json
{ "success": true, "message": "…", "status_code": 200, "data": {} }
```

Failures:

```json
{ "success": false, "message": "…", "status_code": 400, "data": { "code": "…", … } }
```

`status_code` duplicates the HTTP status on the response line so clients can rely on the JSON body alone.

Implementation: `pkg/response` (`OK`, `Fail`, `Error`, `ValidationError`). Validation field errors are under `data.fields`.

---

## Directory layout (layered / “clean” style)

| Path | Responsibility |
|------|----------------|
| `cmd/api/main.go` | Process entry: load config, build `app`, start HTTP server with graceful shutdown; hosts **swag** general API comments. |
| `docs/` | Generated OpenAPI (`docs.go`, `swagger.json`, `swagger.yaml`) — run `make swagger` after changing handler comments. |
| `internal/config` | 12-factor env configuration; **secrets only from environment** (`.env` locally via `godotenv`, secret manager in prod). |
| `internal/domain` | Core entities (`User`); table mapping tags for GORM; no HTTP concerns. |
| `internal/auth`, `internal/user`, `internal/upload`, `internal/health`, `internal/admin` | **Feature modules**: each folder has `dto.go` (in/out JSON shapes), `presenter.go` (domain → safe JSON), `handler.go` (HTTP + swag annotations). |
| `internal/repository` | Data access (GORM); one repository per aggregate/table is the usual pattern. |
| `internal/service` | Business rules: register, login, authorization helpers (`CanView`). |
| `internal/middleware` | Cross-cutting: request ID, panic recovery (JSON), JWT auth, **role guards**, rate limit. |
| `internal/router` | Composes Gin engine; **one file per module** registering routes (`auth_routes.go`, `user_routes.go`, …); mounts **`/swagger/*any`**. |
| `internal/database` | GORM + Postgres driver; connection tuning. |
| `internal/app` | Dependency injection: wires repos → services → handlers → router. |
| `internal/apperrors` | Typed errors with HTTP status + stable `code` for clients (surfaced inside `data`). |
| `internal/validate` | Extra `validator/v10` pass after binding when you want shared rules. |
| `pkg/password` | Bcrypt hashing (**passwords are hashed, not “encrypted”**). |
| `pkg/jwtutil` | HS256 access tokens (claims: user id, email, role). |
| `pkg/response` | Unified `{ success, message, status_code, data }` envelope for every endpoint. |
| `migrations/` | Versioned SQL (**source of truth** for schema). |
| `.github/workflows/ci.yml` | CI: Postgres service, `migrate up`, `go test`. |

**Why `internal/`?** Go tooling hides `internal` from other modules, which matches “this is app code, not a public library.” **`pkg/`** holds small, reusable utilities you might extract later.

---

## Request pipeline (what happens per HTTP call)

1. **Gin engine** (`internal/router/router.go`): global stack — `RequestID` → `Recovery` → `gin.Logger()`.
2. **Versioned API** `/api/v1`: **rate limit** per client IP (in-memory token bucket).
3. **Route group**: JSON binding on DTOs → optional **second validation** via `internal/validate` → **service** → **repository** (GORM).
4. **Errors**: services return `*apperrors.AppError` or wrapped errors; `pkg/response.Error` maps them to HTTP + JSON with `success: false` and `data.code`. Unknown errors become `500` without leaking internals.

Protected routes add `middleware.JWTAuth` (access JWT from HttpOnly cookie or `Authorization: Bearer`, with silent refresh when configured). Admin-only routes add `middleware.AdminOnly()` (role guard; `role` comes from the DB via JWT claims).

---

## Authentication and authorization

- **Register / login**: password hashed with **bcrypt** (`pkg/password`); never store plaintext.
- **JWT**: HS256 signed with `JWT_SECRET` (minimum 32 characters enforced in config). Claims include `uid`, `email`, `role`. **Treat `JWT_SECRET` like a symmetric key**: rotate by issuing new secrets and validating with a key set if you need zero-downtime rotation.
- **Authorization**: example rule in `UserService.CanView` — admin can read any user; normal users only themselves. Extend with policy structs or Casbin if roles grow.

**Session secrets**: this template uses **stateless JWT** for APIs. If you add cookie sessions, store a separate `SESSION_SECRET` in env and use a server-side session store (Redis) for revocation and scale-out.

---

## Database and migrations

- **ORM**: GORM for queries; **schema** is owned by SQL files in `migrations/`, not `AutoMigrate`, so reviews and rollbacks stay explicit.
- **Tool**: [golang-migrate](https://github.com/golang-migrate/migrate) CLI (`migrate up` / `down`).

The **full vertical slice** (migrations → domain → repository → service → HTTP module → router → `app.go` → Swagger) is documented in the next section — **mirrored from [LEARNING.md](LEARNING.md)** so architecture and onboarding stay in sync.

---

## How to add a new feature module (complete checklist)

Here **“module”** means a full vertical slice: **database + domain + repository + (optional) service + HTTP package + routes + app wiring + Swagger**. Follow these steps **in order**. Example names below use **`book` / `books`** — replace with your resource.

### 0) Choose names and URLs

| Choice | Example |
|--------|---------|
| HTTP path prefix | `/api/v1/books` |
| DB table | `books` |
| Go package for HTTP | `internal/book/` (folder name must match import path segment) |
| Domain struct | `Book` in `internal/domain/book.go` |
| Repository | `BookRepository` in `internal/repository/book_repository.go` |
| Service (recommended) | `BookService` in `internal/service/book_service.go` |

---

### 1) SQL migrations (schema first)

1. Add **`migrations/000002_create_books.up.sql`** — `CREATE TABLE books (...)` (align column types with Postgres + GORM).
2. Add **`migrations/000002_create_books.down.sql`** — `DROP TABLE IF EXISTS books;`.
3. Apply locally:  
   `migrate -path migrations -database "$DATABASE_URL" up`  
   (PowerShell: set `$env:DATABASE_URL` first.)

**CI:** `.github/workflows/ci.yml` already runs `migrate up`; new files are picked up automatically.

---

### 2) Domain model

- Create **`internal/domain/book.go`**.
- Define `type Book struct` with `gorm` tags matching the migration (`primaryKey`, `type:uuid`, sizes, `not null`, etc.).
- Implement **`TableName() string`** if the table name differs from pluralized struct name.

---

### 3) Repository (GORM access)

- Create **`internal/repository/book_repository.go`**.
- `NewBookRepository(db *gorm.DB) *BookRepository`.
- Methods take **`context.Context`** as first argument; use `r.db.WithContext(ctx)`.
- For “not found”, follow **`user_repository`**: return **`nil, nil`** (no row) vs real DB error — services map nil to **`apperrors.ErrNotFound`**.

---

### 4) Service (business rules) — recommended

- Create **`internal/service/book_service.go`**.
- Constructor: **`NewBookService(repo *repository.BookRepository) *BookService`**.
- Put validations and authorization rules here (not in handlers). Return **`apperrors.New` / `apperrors.Wrap`** or plain errors as appropriate.
- Handlers stay thin: bind → call service → presenter → **`response.OK` / `response.Error`**.

*Skip the service only for trivial read-only proxies; even then a thin service keeps the pattern consistent.*

---

### 5) HTTP feature package `internal/book/`

Create four files (copy **`internal/user/`** as a minimal template if you only need CRUD + JWT):

| File | Responsibility |
|------|------------------|
| **`doc.go`** | Short package comment: what the feature does; list dto / presenter / handler. |
| **`dto.go`** | Request structs (`binding` tags) + response structs (JSON field names). |
| **`presenter.go`** | **`domain.Book` → outward DTO** (never expose internal or secret fields). |
| **`handler.go`** | Gin handlers: bind JSON → optional **`validate.Struct`** → service → presenter → **`gin-api/pkg/response`**. |

**Handler flow (every method):**

1. Parse path/query/body into DTOs.  
2. On bad input: **`response.ValidationError`**.  
3. Call service with **`c.Request.Context()`**.  
4. On error: **`response.Error(c, err)`**.  
5. On success: **`response.OK(c, http.Status…, "message", presenter.…)`**.

---

### 6) Router: new route file

- Create **`internal/router/book_routes.go`**:

```go
package router

import (
    "gin-api/internal/book"
    "github.com/gin-gonic/gin"
)

func RegisterBookRoutes(g *gin.RouterGroup, h *book.Handler) {
    g.GET("", h.List)           // example
    g.GET("/:id", h.GetByID)
    g.POST("", h.Create)
    // g.PUT("/:id", h.Update) — add as needed
}
```

Adjust methods to match your handler.

---

### 7) Router: register groups in `internal/router/router.go`

1. Import **`gin-api/internal/book`**.
2. Add **`Book *book.Handler`** to the **`Handlers`** struct.
3. Mount routes under **`/api/v1`**:

**Public routes (no JWT):**

```go
books := v1.Group("/books")
RegisterBookRoutes(books, h.Book)
```

**JWT-protected (same pattern as users):**

```go
books := v1.Group("/books", middleware.JWTAuth(cfg, authSvc))
RegisterBookRoutes(books, h.Book)
```

**Admin-only:**

```go
books := v1.Group("/books", middleware.JWTAuth(cfg, authSvc), middleware.AdminOnly())
RegisterBookRoutes(books, h.Book)
```

---

### 8) App wiring: `internal/app/app.go`

1. **`bookRepo := repository.NewBookRepository(db)`**
2. **`bookSvc := service.NewBookService(bookRepo)`** (if you added a service)
3. In **`router.Handlers`**, set **`Book: book.NewHandler(bookSvc)`** (or pass repo if you skipped service — not recommended).

---

### 9) Swagger (OpenAPI)

On each exported handler method, add **`// @Summary`**, **`// @Tags`**, **`// @Router`** (full path like **`/api/v1/books/{id}`**), **`// @Success`** / **`// @Failure`**, and **`// @Security BearerAuth`** when JWT is required.

Then run:

```bash
make swagger
# or: go run github.com/swaggo/swag/cmd/swag@v1.16.4 init -g cmd/api/main.go -o docs --parseInternal
```

---

### 10) Errors (optional)

- Reuse **`internal/apperrors`** sentinels when they fit (**`ErrNotFound`**, **`ErrConflict`**, …).
- Add new **`var Err… = New("CODE", "message", http.Status…)`** only when you need a stable client **`data.code`**.

---

### 11) Config / env (optional)

- If the feature needs new settings (feature flags, buckets, etc.), add fields to **`internal/config/config.go`**, document them in **`.env.example`**, and pass **`cfg`** into **`NewHandler`** (see **`internal/upload`**).

---

### 12) Finish checklist

- [ ] `migrate up` succeeds on a clean DB.  
- [ ] `go build ./...` and `go test ./...` pass.  
- [ ] Hit new endpoints with curl or Swagger UI.  
- [ ] Add a row to the **“Every file explained”** table in **[LEARNING.md](LEARNING.md)** for your new files (keeps onboarding accurate).

*(Same content as LEARNING.md § “How to add a new feature module”; update both places if you change the recipe.)*

---

## Swagger / OpenAPI

- **UI**: with the server running, open `http://localhost:8080/swagger/index.html` (adjust host/port).
- **Regenerate** after editing `// @Summary` / `// @Router` on handlers: `make swagger` or the `go run … swag init …` line in the Makefile.
- `cmd/api/main.go` holds `@title`, `@BasePath`, and `@securityDefinitions.apikey` for Bearer JWT.

---

## File uploads

- `POST /api/v1/upload` (multipart field `file`), JWT required.
- Size cap: `MAX_UPLOAD_MB`; extension allowlist in `internal/upload/handler.go`.
- Files land under `UPLOAD_DIR` with a random UUID filename. For production, move objects to **S3** (or similar) and return URLs.

---

## Scaling out

| Concern | In this template | Typical next step |
|--------|-------------------|-------------------|
| API processes | Single binary | Horizontal replicas behind a load balancer |
| Rate limiting | Per-process map | Redis-backed limiter shared across instances |
| Uploads | Local disk | Object storage + CDN |
| JWT validation | Shared secret on all nodes | Same; or asymmetric keys (RS256) with JWKS |
| DB | One Postgres | Read replicas, connection pool tuning, PgBouncer |

---

## One-liner commands (what each does)

Run these from the repo root (PowerShell: use `;` instead of `&&` where needed).

| Command | What it does |
|---------|----------------|
| `docker compose up -d postgres` | Starts local Postgres (see `docker-compose.yml`). |
| `copy .env.example .env` (Windows) / `cp .env.example .env` (Unix) | Creates env file; **edit** `JWT_SECRET` and `DATABASE_URL`. |
| `go install -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1` | Installs the migrate CLI once (add `$(go env GOPATH)/bin` to `PATH`). |
| `migrate -path migrations -database "%DATABASE_URL%" up` | Applies all pending SQL migrations (Windows cmd; in PowerShell use `$env:DATABASE_URL`). |
| `go mod tidy` | Cleans `go.mod` / `go.sum` to match imports. |
| `go run ./cmd/api` | Builds and runs the HTTP API (loads `.env` if present). |
| `go test ./... -count=1` | Runs tests without cache reuse. |
| `go build -o bin/api ./cmd/api` | Compiles the server binary to `bin/api` (or `bin/api.exe` on Windows). |
| `make swagger` | Regenerates `docs/` OpenAPI from handler comments (same as `go run … swag init …`). |

`Makefile` wraps several of these for Unix-like shells (`make docker-up`, `make migrate-up`, `make run`, `make test`).

---

## HTTP surface (v1)

| Method | Path | Auth | Notes |
|--------|------|------|--------|
| GET | `/health/live` | No | Liveness. |
| GET | `/health/ready` | No | Extend with DB ping. |
| POST | `/api/v1/auth/register` | No | JSON: `email`, `password` (min 12). |
| POST | `/api/v1/auth/login` | No | JSON `data`: `access_token`, `expires_in_seconds` (inside `{ success, message, data }`). |
| GET | `/api/v1/auth/me` | JWT | Current user from token. |
| GET | `/api/v1/users/:id` | JWT | Self or **admin**. |
| POST | `/api/v1/upload` | JWT | Multipart `file`. |
| GET | `/api/v1/admin/ping` | JWT + **admin** role | Example guard. |

**Bootstrap an admin user**: register a user, then in SQL: `UPDATE users SET role = 'admin' WHERE email = 'you@example.com';` (or add a guarded CLI later).

---

## CI pipeline

`.github/workflows/ci.yml` starts Postgres, runs migrations, then `go test ./...` with `DATABASE_URL` and `JWT_SECRET` set. Extend with `golangci-lint` when you add a config file.

---

## Security checklist (short)

- Strong `JWT_SECRET` in production; never commit `.env`.
- TLS termination at the load balancer or reverse proxy.
- CORS: configure explicitly if you add browser clients (`github.com/gin-contrib/cors`).
- Audit uploads (MIME sniffing, antivirus) before production.
- Log structured fields (`slog`); avoid logging tokens or passwords.

This structure keeps **transport**, **use-cases**, and **persistence** separate so teams can grow each layer and test services without HTTP.
