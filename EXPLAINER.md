# Explainer — this whole template in simple steps

Hi. This file is here so **you can open the project and not feel lost**. If you are young, new, or just tired, read it in order from top to bottom. It tells the **same story the code tells**, but with small words and a clear path.

---

## 1. What is this project?

Imagine a **restaurant**.

- The **customer** is someone using a phone app or a website.
- The **waiter** is this Go program (the “API”). It takes orders and brings answers back.
- The **kitchen** is the **database** (PostgreSQL). That is where we remember things, like who signed up.

This template is the **waiter + kitchen rules**: how to take an order safely, how to check ID (login), and how to answer always in the **same simple shape**.

---

## 2. The answer shape (every time)

Every answer from the API uses the same keys: `success`, `message`, `status_code` (same as HTTP status), and `data`.

**When things work:**

```json
{
  "success": true,
  "message": "short human words",
  "status_code": 200,
  "data": { }
}
```

`status_code` is the same number as the HTTP status (200, 201, 400, …) so clients can read one JSON body and know both outcome and status.

`data` can be an **object** `{}` or a **list** `[]`, depending on the endpoint.

**When something is wrong:**

```json
{
  "success": false,
  "message": "what went wrong in plain language",
  "status_code": 400,
  "data": { }
}
```

Extra hints (like `code` or `fields`) live **inside** `data` so the outside shape never changes.

The code that does this lives in `pkg/response/response.go`. If you add a new endpoint, use `response.OK` or `response.Fail` / `response.Error` so you do not break the pattern (including `status_code`).

---

## 3. The folders (who does what)

Think of **modules** as **mini apps** inside one big app.

For each feature area you will usually see **three friends in the same folder**:

| File kind | Job (kid-friendly) |
|-----------|--------------------|
| `dto.go` | The **forms**: what JSON we allow in, and the **shape** we send out. |
| `presenter.go` | The **translator**: turns database models into safe JSON (never leaks passwords). |
| `handler.go` | The **worker** for HTTP: reads the request, calls services, writes the answer. |

Example folders:

- `internal/auth` — sign up, log in, “who am I?”
- `internal/user` — read a user profile (with permission rules)
- `internal/upload` — save a file
- `internal/health` — “is the server alive?”
- `internal/admin` — example of “only admins allowed”

Shared stuff that is not one feature:

- `internal/router` — connects URLs to handlers (traffic signs)
- `internal/middleware` — checks that run **before** your handler (rate limit, JWT)
- `internal/service` — business rules (the smart part)
- `internal/repository` — talks to the database with GORM
- `internal/domain` — the database row shape in Go
- `migrations/` — SQL files that build tables (the real blueprint)

---

## 4. The full journey of one request (start → end)

Read this like a **comic strip** in order.

### Step A — The program starts

1. You run `go run ./cmd/api` (or the built `.exe`).
2. `cmd/api/main.go` loads settings from the environment (`.env` helps on your laptop).
3. `internal/app` opens the database connection and builds the Gin `Engine` (the web server).
4. The server listens on a port (default `8080`).

### Step B — A request hits the server

1. The request goes into **middleware** first (like airport security):
   - **Request ID** — tag the request so logs make sense.
   - **Recovery** — if code panics, we still return JSON instead of crashing ugly.
   - **Logger** — write a line about the request.
2. If the path is under `/api/v1`, **rate limiting** runs (stop spam).

### Step C — The router picks a handler

`internal/router/router.go` groups URLs:

- `/api/v1/auth/...` → `internal/auth`
- `/api/v1/users/...` → `internal/user`
- `/api/v1/upload` → `internal/upload`
- `/health/...` → `internal/health`

Each group can also say: “JWT required” or “admin only”.

### Step D — Inside the handler

1. **Bind JSON** into a struct from `dto.go` (Gin checks basic rules).
2. Sometimes **extra validation** runs in `internal/validate`.
3. Call the **service** (the rules: “can this happen?”).
4. The service calls the **repository** (database).
5. Use **`presenter.go`** to build safe JSON for the client.
6. Return with **`response.OK`** (success) or **`response.Error`** / **`response.ValidationError`** (failure).

### Step E — The answer goes back

The client always gets `{ success, message, status_code, data }`.

---

## 5. Login and JWT (the “hall pass”)

1. `POST /api/v1/auth/login` checks email + password.
2. Passwords are stored as **bcrypt hashes** (one-way; we cannot “read back” the password).
3. If good, we return a **JWT** string. That is a signed note that says “this user is allowed”.
4. For protected routes, the client sends header: `Authorization: Bearer <token>`.
5. `middleware/JWTAuth` verifies the signature using `JWT_SECRET` from the environment.

**Secret rule:** `JWT_SECRET` must be long and private, like a real secret diary key.

---

## 6. Database and migrations (the real table layout)

1. SQL files in `migrations/` describe tables.
2. You run `migrate up` to apply them (see `Makefile` / `ARCHITECTURE.md`).
3. GORM reads/writes rows using models in `internal/domain`.

Why not only GORM auto-migrate? Because SQL files are easy to review in a team, like architectural drawings.

---

## 7. Swagger (try the API in the browser)

When the server runs, open:

`http://localhost:8080/swagger/index.html`

(Use your real port if you changed `HTTP_PORT`.)

That page lists routes and lets you **try requests**. For protected routes, use the **Authorize** button and paste: `Bearer <your token>`.

If you change comments above handlers, regenerate docs:

```bash
go run github.com/swaggo/swag/cmd/swag@v1.16.4 init -g cmd/api/main.go -o docs --parseInternal
```

---

## 8. How you add a new “thing” (example: “books”)

Do these steps in order. Copy the style of `internal/user` or `internal/auth`.

1. **Migration** — add SQL for table `books`.
2. **Domain** — `internal/domain/book.go` model.
3. **Repository** — `internal/repository/book_repository.go`.
4. **Service** — `internal/service/book_service.go` rules.
5. **Module folder** — `internal/book/dto.go`, `presenter.go`, `handler.go`.
6. **Routes** — `internal/router/book_routes.go` + register in `router.go`.
7. **Swagger comments** on handler methods, then run `swag init` again.

If you follow that order, you rarely get lost.

---

## 9. If you feel stuck

- Read **`ARCHITECTURE.md`** for grown-up detail and command cheat sheets.
- Search the code for `response.OK` to see happy paths.
- Search for `response.Error` to see failure paths.

You are not supposed to memorize everything. You are supposed to know **where the story lives**: handlers → services → repositories → SQL.

Welcome aboard. You can contribute. Go one file at a time.
