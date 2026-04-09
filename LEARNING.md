# LEARNING guide — Gin API template

Deep dive: **reading order** for source files, **startup and request runtime flow**, and **how to add a feature module**. For a high-level overview, start with [README.md](README.md); for plain-language concepts, see [EXPLAINER.md](EXPLAINER.md); for operations and security, see [ARCHITECTURE.md](ARCHITECTURE.md).

---

## Which doc should I read?

| Document | Purpose |
|----------|---------|
| [README.md](README.md) | **Overview:** features, architecture summary, installation, tech stack, contribution link. |
| **This guide** | **Reading order** for files, **runtime flow** (which file runs when), feature checklist. |
| [EXPLAINER.md](EXPLAINER.md) | Same ideas in very plain language (good if concepts are new). |
| [ARCHITECTURE.md](ARCHITECTURE.md) | Deeper layout, scaling, security checklist, command reference. |

Source files also include **package comments** and **inline flow notes** (especially `cmd/api/main.go`, `internal/router/router.go`, feature `doc.go` files, and handlers).

---

## Every file explained (read in this sequence)

Use this as a checklist. Each line is **one file** and what it does in the request/startup flow.

| # | File | What it does |
|---|------|----------------|
| 1 | `go.mod` / `go.sum` | Module name `gin-api` and dependency versions. |
| 2 | `.env.example` | Documents required env vars (copy to `.env` locally). |
| 3 | `cmd/api/main.go` | **Entry:** loads config, builds `app`, starts `http.Server`, graceful shutdown; blank-imports `gin-api/docs` so Swagger registers; swag `@` comments for OpenAPI. |
| 4 | `internal/config/config.go` | Reads env + `.env`; validates `DATABASE_URL`, `JWT_SECRET`; builds `Config`. |
| 5 | `internal/app/app.go` | **Wiring:** `database.Open` → `UserRepository` → `AuthService` / `UserService` → feature handlers → `router.NewEngine`. |
| 6 | `internal/database/postgres.go` | Opens one shared `*gorm.DB` to Postgres (UTC timestamps, optional SQL log in dev). |
| 7 | `internal/router/router.go` | Builds Gin engine: global middleware → `/swagger` → `/health` → `/api/v1` + rate limit → sub-routes + JWT on protected groups. |
| 8 | `internal/router/auth_routes.go` | Maps `/auth/register`, `/login`, `/me` → `auth.Handler`. |
| 9 | `internal/router/user_routes.go` | Maps `/users/:id` → `user.Handler` (JWT on parent group). |
| 10 | `internal/router/upload_routes.go` | Maps `POST /upload` → `upload.Handler`. |
| 11 | `internal/router/admin_routes.go` | Maps `/admin/ping` → `admin.Handler`. |
| 12 | `internal/middleware/requestid.go` | Sets/propagates `X-Request-ID`. |
| 13 | `internal/middleware/recovery.go` | Catches panics → JSON 500 via `response.Error`. |
| 14 | `internal/middleware/ratelimit.go` | Per-IP token bucket on `/api/v1`. |
| 15 | `internal/middleware/auth.go` | `JWTAuth`, `RequireRole` / `AdminOnly`; helpers `MustUserID` / `MustEmail` / `MustRole`. |
| 16 | `migrations/000001_create_users.up.sql` | **Schema:** `users` table (source of truth with `.down.sql`). |
| 17 | `migrations/000001_create_users.down.sql` | Drops `users` for rollback. |
| 18 | `internal/domain/user.go` | GORM model aligned with `users` table + role constants. |
| 19 | `internal/repository/user_repository.go` | `Create`, `FindByEmail`, `FindByID` (nil user = not found). |
| 20 | `internal/service/doc.go` | Package doc: services = use-cases without HTTP. |
| 21 | `internal/service/auth_service.go` | Register (hash + insert), Login (verify + JWT). |
| 22 | `internal/service/user_service.go` | GetByID + `CanView` authorization helper. |
| 23 | `internal/apperrors/errors.go` | Typed errors + HTTP status; `response.Error` maps them. |
| 24 | `internal/validate/validate.go` | Second validation pass on DTOs after Gin binding. |
| 25 | `pkg/response/response.go` | **All JSON:** `{ success, message, status_code, data }`. |
| 26 | `pkg/password/password.go` | Bcrypt hash/verify. |
| 27 | `pkg/jwtutil/jwtutil.go` | Sign/parse HS256 JWT with user claims. |
| 28 | `internal/auth/doc.go` | Describes auth package layout (dto / presenter / handler). |
| 29 | `internal/auth/dto.go` | Auth request/response structs + binding tags. |
| 30 | `internal/auth/presenter.go` | Domain user / claims → `UserOut`, `TokenOut`. |
| 31 | `internal/auth/handler.go` | HTTP handlers + Swagger annotations → `AuthService` + `response.*`. |
| 32 | `internal/user/doc.go` | User feature package overview. |
| 33 | `internal/user/dto.go` | `UserOut` shape. |
| 34 | `internal/user/presenter.go` | `domain.User` → `UserOut`. |
| 35 | `internal/user/handler.go` | GET user by id with permission check. |
| 36 | `internal/upload/doc.go` | Upload feature overview. |
| 37 | `internal/upload/dto.go` | Upload result JSON shape. |
| 38 | `internal/upload/presenter.go` | Build upload success payload. |
| 39 | `internal/upload/handler.go` | Multipart save to disk. |
| 40 | `internal/health/doc.go` | Health package overview. |
| 41 | `internal/health/dto.go` / `presenter.go` | Probe JSON shapes + builders. |
| 42 | `internal/health/handler.go` | `/health/live` and `/ready`. |
| 43 | `internal/admin/doc.go` | Admin example overview. |
| 44 | `internal/admin/dto.go` / `presenter.go` / `handler.go` | Admin ping JSON + handler. |
| 45 | `docs/docs.go` + `swagger.json` / `.yaml` | **Generated** OpenAPI (`make swagger`); do not hand-edit `docs.go`. |
| 46 | `docker-compose.yml` | Local Postgres for dev. |
| 47 | `Makefile` | `run`, `test`, `swagger`, migrate helpers. |
| 48 | `.github/workflows/ci.yml` | CI: Postgres, migrate, `go test`. |

---

## Read the codebase in this order (first day)

Follow this sequence once. After that, jump by feature using the sections below.

### Phase 1 — Program boots

1. [`cmd/api/main.go`](cmd/api/main.go) — entry: config, build app, HTTP server, graceful shutdown. Swagger metadata comments live here.
2. [`internal/config/config.go`](internal/config/config.go) — env vars, validation (`JWT_SECRET` length, `DATABASE_URL`).
3. [`internal/app/app.go`](internal/app/app.go) — **wiring**: DB → repositories → services → handlers → `router.NewEngine`.
4. [`internal/database/postgres.go`](internal/database/postgres.go) — GORM + Postgres connection.

### Phase 2 — HTTP wiring

5. [`internal/router/router.go`](internal/router/router.go) — global middleware, `/swagger`, `/health/*`, `/api/v1` groups, attaches rate limit + JWT where needed.
6. [`internal/router/auth_routes.go`](internal/router/auth_routes.go) — example: how URLs map to handler methods.
7. [`internal/router/user_routes.go`](internal/router/user_routes.go)
8. [`internal/router/upload_routes.go`](internal/router/upload_routes.go)
9. [`internal/router/admin_routes.go`](internal/router/admin_routes.go)

### Phase 3 — Cross-cutting HTTP

10. [`internal/middleware/requestid.go`](internal/middleware/requestid.go)
11. [`internal/middleware/recovery.go`](internal/middleware/recovery.go)
12. [`internal/middleware/ratelimit.go`](internal/middleware/ratelimit.go)
13. [`internal/middleware/auth.go`](internal/middleware/auth.go) — JWT + role guards (`AdminOnly`).

### Phase 4 — One full feature (auth), top to bottom

14. [`migrations/000001_create_users.up.sql`](migrations/000001_create_users.up.sql) — schema source of truth.
15. [`internal/domain/user.go`](internal/domain/user.go) — GORM model matching that table.
16. [`internal/repository/user_repository.go`](internal/repository/user_repository.go) — DB reads/writes.
17. [`internal/service/auth_service.go`](internal/service/auth_service.go) — register/login rules.
18. [`internal/auth/dto.go`](internal/auth/dto.go) — JSON in/out shapes.
19. [`internal/auth/presenter.go`](internal/auth/presenter.go) — domain → safe JSON (no secrets).
20. [`internal/auth/handler.go`](internal/auth/handler.go) — HTTP + Swagger annotations.
21. [`internal/validate/validate.go`](internal/validate/validate.go) — extra validator pass after binding.
22. [`pkg/response/response.go`](pkg/response/response.go) — `{ success, message, status_code, data }` for every response.
23. [`pkg/password/password.go`](pkg/password/password.go) — bcrypt.
24. [`pkg/jwtutil/jwtutil.go`](pkg/jwtutil/jwtutil.go) — sign/parse JWT.
25. [`internal/apperrors/errors.go`](internal/apperrors/errors.go) — typed errors → HTTP + `data.code`.

### Phase 5 — Other modules (same pattern each)

26. [`internal/user/`](internal/user/) — `dto.go` → `presenter.go` → `handler.go` (+ `internal/service/user_service.go`).
27. [`internal/upload/`](internal/upload/) — file upload handler + config from [`internal/config/config.go`](internal/config/config.go).
28. [`internal/health/`](internal/health/) — liveness/readiness.
29. [`internal/admin/`](internal/admin/) — JWT + admin-only example.

### Phase 6 — API contract & automation

30. [`docs/swagger.yaml`](docs/swagger.yaml) or [`docs/swagger.json`](docs/swagger.json) — generated OpenAPI (run `make swagger` after changing handler comments).
31. [`.github/workflows/ci.yml`](.github/workflows/ci.yml) — Postgres, migrate, `go test`.
32. [`Makefile`](Makefile) — `run`, `test`, `swagger`, `migrate-*`, `docker-*`.

---

## Runtime flow: one HTTP request (file sequence)

When a client calls e.g. `POST /api/v1/auth/login`, execution order is:

1. **`cmd/api/main.go`** — `http.Server` receives the connection (handler is the Gin `Engine`).
2. **`internal/router/router.go`** — Gin runs middleware **in registration order**:
   - `middleware.RequestID`
   - `middleware.Recovery`
   - `gin.Logger`
3. Same file — route group `/api/v1` runs **`middleware.RateLimit`**.
4. **`internal/router/auth_routes.go`** — route matches `POST .../login` → **`internal/auth/handler.go`** `Login`.
5. **`internal/auth/handler.go`** — `ShouldBindJSON` into **`internal/auth/dto.go`**; optional **`internal/validate/validate.go`**; calls **`internal/service/auth_service.go`**.
6. **`internal/service/auth_service.go`** — **`internal/repository/user_repository.go`** (GORM) + **`pkg/password/password.go`**; on success **`pkg/jwtutil/jwtutil.go`**.
7. **`internal/repository/user_repository.go`** — uses **`internal/domain/user.go`** model and Postgres via **`internal/database/postgres.go`** (opened in **`internal/app/app.go`**).
8. Back in handler — **`internal/auth/presenter.go`** builds outbound DTO; **`pkg/response/response.go`** writes JSON.

If the route is **JWT-protected** (e.g. `GET /api/v1/auth/me`), steps 3–4 include **`internal/middleware/auth.go`** `JWTAuth` **before** the handler runs.

If something fails with a known app error, **`pkg/response/response.go`** `Error` uses **`internal/apperrors/errors.go`** to set status and `data.code`.

---

## Flow: “understand login only” (minimal path)

Read only these, in order:

`router.go` (see `/api/v1` + auth group) → `auth_routes.go` → `internal/auth/handler.go` (`Login`) → `internal/auth/dto.go` → `internal/service/auth_service.go` → `internal/repository/user_repository.go` → `internal/domain/user.go` → `migrations/000001_create_users.up.sql` → `pkg/password/password.go` + `pkg/jwtutil/jwtutil.go` → `pkg/response/response.go`.

---

## How to add a new feature module (complete checklist)

This section is **mirrored in [ARCHITECTURE.md](ARCHITECTURE.md)** (after “Database and migrations”); edit both if you change the recipe.

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
- [ ] Add a row to the **“Every file explained”** table in this guide for your new files (keeps onboarding accurate).

---

## Response shape (every endpoint)

Success: `{ "success": true, "message": "…", "status_code": 200, "data": { } }`  
Failure: `{ "success": false, "message": "…", "status_code": 400, "data": { } }` — `status_code` matches the HTTP status line (e.g. 401, 404, 429). Extra hints like `code` / `fields` stay inside `data`.

Implemented in [`pkg/response/response.go`](pkg/response/response.go).
