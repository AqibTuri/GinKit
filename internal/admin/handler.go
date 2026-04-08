package admin

import (
	"net/http"

	"gin-api/pkg/response"
	"github.com/gin-gonic/gin"
)

// Handler: example admin-only JSON (real admin checks happen in router middleware chain).
type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// Ping godoc
// @Summary      Admin ping
// @Description  Example route protected by JWT + admin role.
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Body{data=admin.PingOut}
// @Failure      403  {object}  response.Body
// @Router       /api/v1/admin/ping [get]
func (h *Handler) Ping(c *gin.Context) {
	response.OK(c, http.StatusOK, "OK", PresentPing())
}
