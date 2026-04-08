package admin

// PresentPing builds the admin ping JSON body.
func PresentPing() PingOut {
	return PingOut{Scope: "admin", Status: "ok"}
}
