package health

import (
	"net/http"

	"gin-api/pkg/response"
	"github.com/gin-gonic/gin"
)

// Handler: minimal JSON for orchestrator probes (extend Ready with DB ping later).
type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// Live godoc
// @Summary      Liveness probe
// @Tags         health
// @Produce      json
// @Success      200  {object}  response.Body{data=health.StatusOut}
// @Router       /health/live [get]
func (h *Handler) Live(c *gin.Context) {
	response.OK(c, http.StatusOK, "OK", Live())
}

// Ready godoc
// @Summary      Readiness probe
// @Tags         health
// @Produce      json
// @Success      200  {object}  response.Body{data=health.StatusOut}
// @Router       /health/ready [get]
func (h *Handler) Ready(c *gin.Context) {
	response.OK(c, http.StatusOK, "OK", Ready())
}
