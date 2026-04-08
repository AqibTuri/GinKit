// Package auth is the HTTP surface for registration, login, and current user (JWT).
//
// Files:
//   - dto.go      — JSON request/response structs (API contract).
//   - presenter.go — maps domain.User (or claims) → outbound JSON (never expose password hash).
//   - handler.go  — Gin handlers; call AuthService then response.* helpers.
package auth
