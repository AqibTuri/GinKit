package health

// StatusOut — simple { "status": "ok|ready" } for probes.

type StatusOut struct {
	Status string `json:"status" example:"ok"`
}
