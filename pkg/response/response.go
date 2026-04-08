// Package response writes the unified JSON envelope for every HTTP response (success + errors).
// Always use OK/Fail/Error/ValidationError so clients see consistent { success, message, status_code, data }.
package response

import (
	"net/http"

	"gin-api/internal/apperrors"
	"github.com/gin-gonic/gin"
)

// Body is the JSON envelope for every API response. status_code matches the HTTP status on the wire.
type Body struct {
	Success    bool   `json:"success" example:"true"`
	Message    string `json:"message" example:"OK"`
	StatusCode int    `json:"status_code" example:"200"`
	Data       any    `json:"data" swaggertype:"object"`
}

// OK sends a success response. Pass a struct, slice, or map for data; nil becomes {}.
func OK(c *gin.Context, status int, message string, data any) {
	if data == nil {
		data = map[string]any{}
	}
	c.JSON(status, Body{Success: true, Message: message, StatusCode: status, Data: data})
}

// Fail sends an error-style response (success: false). Put machine-readable extras (e.g. code, fields) inside data.
func Fail(c *gin.Context, status int, message string, data any) {
	if data == nil {
		data = map[string]any{}
	}
	c.JSON(status, Body{Success: false, Message: message, StatusCode: status, Data: data})
}

// Error maps errors to HTTP status and puts apperrors.Code inside data.code when known.
func Error(c *gin.Context, err error) {
	if ae, ok := apperrors.IsAppError(err); ok {
		Fail(c, ae.HTTPStatus, ae.Message, gin.H{"code": ae.Code})
		return
	}
	Fail(c, http.StatusInternalServerError, "Something went wrong", gin.H{"code": "INTERNAL_ERROR"})
}

// ValidationError sends 400 with field errors under data.fields.
func ValidationError(c *gin.Context, fields map[string]string) {
	Fail(c, http.StatusBadRequest, "Validation failed", gin.H{
		"code":   "VALIDATION_ERROR",
		"fields": fields,
	})
}

func TooManyRequests(c *gin.Context) {
	c.Header("Retry-After", "1")
	Fail(c, http.StatusTooManyRequests, "Too many requests", gin.H{"code": "RATE_LIMITED"})
}
