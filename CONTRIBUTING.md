# Contributing

Thanks for helping improve this template. This document describes how we expect contributions to fit the project.

## Before you start

- Read [README.md](README.md) for scope and setup.
- For file-by-file onboarding and how to add a feature end-to-end, see [Learning.md](Learning.md) and [ARCHITECTURE.md](ARCHITECTURE.md).

## Ways to contribute

- **Issues:** Report bugs, unclear docs, or missing checks with enough detail to reproduce (OS, Go version, commands run, relevant logs).
- **Pull requests:** Small, focused changes are easiest to review. One logical change per PR is ideal.

## Pull request checklist

1. **Format and build:** `go fmt ./...` and `go build ./...` succeed.
2. **Tests:** `go test ./...` passes locally (CI runs against Postgres and migrations).
3. **Migrations:** If you change the database schema, add paired `.up.sql` / `.down.sql` files and ensure `migrate up` works on a clean database.
4. **API docs:** If you add or change handler Swag comments, run `make swagger` and commit the generated `docs/` updates.
5. **Docs:** If you add notable files or flows, add a row to the **“Every file explained”** table in [Learning.md](Learning.md) and keep [ARCHITECTURE.md](ARCHITECTURE.md) in sync when you touch the mirrored “How to add a new feature module” section.

## Code style

- Match existing layout: feature folders under `internal/<feature>/` with `doc.go`, `dto.go`, `presenter.go`, `handler.go` where applicable.
- Keep handlers thin; put rules in `internal/service` and data access in `internal/repository`.
- Use [`pkg/response`](pkg/response/response.go) for JSON and [`internal/apperrors`](internal/apperrors/errors.go) for typed HTTP errors.
- Do not commit secrets, real `.env` files, or generated upload directories with user data.

## Questions

Open an issue for design or scope questions so decisions stay visible to future contributors.
