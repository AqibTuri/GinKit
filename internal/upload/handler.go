package upload

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gin-api/internal/config"
	"gin-api/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler: multipart parse → validate size/ext → write under cfg.UploadDir → PresentResult → response.
type Handler struct {
	cfg *config.Config
}

func NewHandler(cfg *config.Config) *Handler {
	return &Handler{cfg: cfg}
}

var allowedExt = map[string]struct{}{
	".jpg": {}, ".jpeg": {}, ".png": {}, ".gif": {}, ".webp": {},
	".pdf": {}, ".txt": {},
}

// Upload godoc
// @Summary      Upload a file
// @Description  Multipart form field name: file. Size and extensions are restricted.
// @Tags         upload
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file  formData  file  true  "File to upload"
// @Success      201   {object}  response.Body{data=upload.UploadResultOut}
// @Failure      400   {object}  response.Body
// @Router       /api/v1/upload [post]
func (h *Handler) Upload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.cfg.MaxUploadBytes)
	if err := c.Request.ParseMultipartForm(h.cfg.MaxUploadBytes); err != nil {
		response.ValidationError(c, map[string]string{"file": "file too large or invalid multipart"})
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.ValidationError(c, map[string]string{"file": "file is required"})
		return
	}
	f, err := file.Open()
	if err != nil {
		response.Error(c, err)
		return
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if _, ok := allowedExt[ext]; ext == "" || !ok {
		response.ValidationError(c, map[string]string{"file": "extension not allowed"})
		return
	}

	if err := os.MkdirAll(h.cfg.UploadDir, 0o750); err != nil {
		response.Error(c, err)
		return
	}
	name := fmt.Sprintf("%s%s", uuid.NewString(), ext)
	dst := filepath.Join(h.cfg.UploadDir, name)
	out, err := os.Create(dst)
	if err != nil {
		response.Error(c, err)
		return
	}
	defer out.Close()
	written, err := io.Copy(out, f)
	if err != nil {
		_ = os.Remove(dst)
		response.Error(c, err)
		return
	}
	if written > h.cfg.MaxUploadBytes {
		_ = os.Remove(dst)
		response.ValidationError(c, map[string]string{"file": "file too large"})
		return
	}

	response.OK(c, http.StatusCreated, "File uploaded", PresentResult(name, written))
}
