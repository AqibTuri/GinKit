package user

import (
	"net/http"

	"gin-api/internal/apperrors"
	"gin-api/internal/middleware"
	"gin-api/internal/service"
	"gin-api/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler: parse :id → UserService.GetByID → CanView (actor from JWT context) → PresentPublic → response.
type Handler struct {
	svc *service.UserService
}

func NewHandler(svc *service.UserService) *Handler {
	return &Handler{svc: svc}
}

// GetByID godoc
// @Summary      Get user by ID
// @Description  Caller may read self or any user if admin.
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User UUID"
// @Success      200  {object}  response.Body{data=user.UserOut}
// @Failure      400  {object}  response.Body
// @Failure      403  {object}  response.Body
// @Failure      404  {object}  response.Body
// @Router       /api/v1/users/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.ValidationError(c, map[string]string{"id": "must be a valid UUID"})
		return
	}
	actorID, ok := middleware.MustUserID(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}
	actorRole := middleware.MustRole(c)

	u, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	if !h.svc.CanView(actorID, actorRole, u) {
		response.Error(c, apperrors.ErrForbidden)
		return
	}
	response.OK(c, http.StatusOK, "OK", PresentPublic(u))
}
