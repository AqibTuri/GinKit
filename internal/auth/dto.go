package auth

// dto.go — API JSON shapes for auth. Inbound structs use `binding` tags (Gin validator).
// Outbound structs are filled via presenter.go for Swagger and stable field names.

// Inbound JSON bodies (requests).

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email,max=255" example:"you@example.com"`
	Password string `json:"password" binding:"required,min=12,max=72" example:"long-safe-password"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"you@example.com"`
	Password string `json:"password" binding:"required" example:"long-safe-password"`
}

// Outbound JSON shapes (responses) — use via presenter functions for safe, consistent output.

type UserOut struct {
	ID    string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email string `json:"email" example:"you@example.com"`
	Role  string `json:"role" example:"user"`
}

type TokenOut struct {
	AccessToken      string `json:"access_token"`
	ExpiresInSeconds int64  `json:"expires_in_seconds" example:"3600"`
}
