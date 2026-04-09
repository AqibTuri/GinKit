package auth

import (
	"net/http"

	"gin-api/internal/apperrors"
	"gin-api/internal/authcookie"
	"gin-api/internal/config"
	"gin-api/internal/middleware"
	"gin-api/internal/service"
	"gin-api/internal/validate"
	"gin-api/pkg/response"
	"github.com/gin-gonic/gin"
)

// Handler wires Gin to AuthService.
type Handler struct {
	svc *service.AuthService
	cfg *config.Config
}

func NewHandler(svc *service.AuthService, cfg *config.Config) *Handler {
	return &Handler{svc: svc, cfg: cfg}
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates an account with email and password (bcrypt). Password min 12 chars.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      RegisterRequest  true  "Register payload"
// @Success      201   {object}  response.Body{data=auth.UserOut}
// @Failure      400   {object}  response.Body
// @Failure      409   {object}  response.Body
// @Router       /api/v1/auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, map[string]string{"body": "invalid json or binding failed"})
		return
	}
	if d := validate.Struct(&req); len(d) > 0 {
		response.ValidationError(c, d)
		return
	}
	u, err := h.svc.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, http.StatusCreated, "Account created", PresentUser(u))
}

// Login godoc
// @Summary      Login
// @Description  Sets HttpOnly cookies (access + refresh). Optional JSON body for credentials.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      LoginRequest  true  "Login payload"
// @Success      200   {object}  response.Body{data=auth.SessionOut}
// @Failure      400   {object}  response.Body
// @Failure      401   {object}  response.Body
// @Router       /api/v1/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, map[string]string{"body": "invalid json or binding failed"})
		return
	}
	if d := validate.Struct(&req); len(d) > 0 {
		response.ValidationError(c, d)
		return
	}
	access, refresh, aSec, rSec, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		response.Error(c, err)
		return
	}
	authcookie.Set(c, h.cfg.AuthCookieSettings(), access, refresh, int(aSec), int(rSec))
	response.OK(c, http.StatusOK, "Logged in", PresentSession(aSec, rSec))
}

// Refresh godoc
// @Summary      Refresh session
// @Description  Reads refresh cookie, rotates tokens, sets new HttpOnly cookies.
// @Tags         auth
// @Produce      json
// @Success      200  {object}  response.Body{data=auth.SessionOut}
// @Failure      401  {object}  response.Body
// @Router       /api/v1/auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	raw, err := c.Cookie(h.cfg.CookieRefreshName)
	if err != nil || raw == "" {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}
	access, refresh, aSec, rSec, err := h.svc.Refresh(c.Request.Context(), raw)
	if err != nil {
		response.Error(c, err)
		return
	}
	authcookie.Set(c, h.cfg.AuthCookieSettings(), access, refresh, int(aSec), int(rSec))
	response.OK(c, http.StatusOK, "Refreshed", PresentSession(aSec, rSec))
}

// Logout godoc
// @Summary      Logout
// @Description  Clears auth cookies (client session ends).
// @Tags         auth
// @Produce      json
// @Success      200  {object}  response.Body
// @Router       /api/v1/auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	authcookie.Clear(c, h.cfg.AuthCookieSettings())
	response.OK(c, http.StatusOK, "Logged out", nil)
}

// Me godoc
// @Summary      Current user
// @Description  Profile from JWT (no DB hit).
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Body{data=auth.UserOut}
// @Failure      401  {object}  response.Body
// @Router       /api/v1/auth/me [get]
func (h *Handler) Me(c *gin.Context) {
	uid, ok := middleware.MustUserID(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}
	out := PresentMe(uid, middleware.MustEmail(c), middleware.MustRole(c))
	response.OK(c, http.StatusOK, "OK", out)
}
