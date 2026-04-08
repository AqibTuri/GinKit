package admin

// PingOut — example admin JSON (scope + status).

type PingOut struct {
	Scope  string `json:"scope" example:"admin"`
	Status string `json:"status" example:"ok"`
}
