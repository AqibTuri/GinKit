package user

// UserOut — public user JSON for GET /users/:id (built by presenter.PresentPublic).
type UserOut struct {
	ID    string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email string `json:"email" example:"you@example.com"`
	Role  string `json:"role" example:"user"`
}
