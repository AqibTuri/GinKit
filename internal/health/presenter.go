package health

// Live / Ready return fixed payloads for handler.go (swap Ready body when DB checks exist).

func Live() StatusOut {
	return StatusOut{Status: "ok"}
}

func Ready() StatusOut {
	return StatusOut{Status: "ready"}
}
