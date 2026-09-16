package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"c2dp/go-api/internal/config"
	appmiddleware "c2dp/go-api/internal/middleware"
	"c2dp/go-api/internal/models"
	"c2dp/go-api/internal/signing"
)

type PostHandler struct {
	cfg config.Config
	db  *gorm.DB
}

func NewPostHandler(db *gorm.DB, cfg config.Config) *PostHandler {
	return &PostHandler{cfg: cfg, db: db}
}

func (h *PostHandler) Create(c echo.Context) error {
	userID, ok := c.Get(appmiddleware.UserIDKey).(uint)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "missing user")
	}

	// Bound multipart parsing as well as the uploaded image itself.
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, 20<<20)
	if err := c.Request().ParseMultipartForm(4 << 20); err != nil {
		if c.Request().MultipartForm != nil {
			defer c.Request().MultipartForm.RemoveAll()
		}
		return echo.NewHTTPError(http.StatusBadRequest, "invalid multipart upload; maximum request size is 20 MiB")
	}
	defer c.Request().MultipartForm.RemoveAll()
	fileHeader, err := c.FormFile("image")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "image is required")
	}
	src, err := fileHeader.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "could not read image")
	}
	defer src.Close()
	header := make([]byte, 512)
	n, err := src.Read(header)
	if err != nil && err != io.EOF {
		return echo.NewHTTPError(http.StatusBadRequest, "could not read image")
	}
	contentType := http.DetectContentType(header[:n])
	ext := extensionForContentType(contentType)
	if ext == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "image must be jpeg, png, webp, or gif")
	}
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "could not read image")
	}

	if err := os.MkdirAll(h.cfg.UploadDir, 0755); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not prepare upload directory")
	}
	// Stage outside the publicly served upload directory, on the same filesystem
	// so publishing the complete signed image can use an atomic rename.
	uploadDir, err := filepath.Abs(h.cfg.UploadDir)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not prepare upload directory")
	}
	stage, err := os.MkdirTemp(filepath.Dir(uploadDir), ".c2dp-upload-")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not stage upload")
	}
	defer os.RemoveAll(stage)
	source := filepath.Join(stage, "source"+ext)
	dst, err := os.Create(source)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not save image")
	}
	_, copyErr := io.Copy(dst, src)
	closeErr := dst.Close()
	if copyErr != nil || closeErr != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not save image")
	}
	signed := filepath.Join(stage, "signed"+ext)
	if err := signing.Sign(c.Request().Context(), h.cfg.C2PAToolPath, source, signed); err != nil {
		c.Logger().Errorf("sign upload: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "could not sign image; no post was created")
	}
	imagePath := filepath.Join(uploadDir, uuid.NewString()+ext)
	if err := os.Rename(signed, imagePath); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not publish signed image")
	}

	post := models.Post{
		Caption:   strings.TrimSpace(c.FormValue("caption")),
		ImagePath: imagePath,
		UserID:    userID,
	}
	if err := h.db.Create(&post).Error; err != nil {
		_ = os.Remove(imagePath)
		return echo.NewHTTPError(http.StatusInternalServerError, "could not create post")
	}

	if err := h.db.Preload("User").First(&post, post.ID).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not load post")
	}
	h.attachImageURL(&post)

	return c.JSON(http.StatusCreated, post)
}

func (h *PostHandler) List(c echo.Context) error {
	var posts []models.Post
	if err := h.db.Preload("User").Order("created_at DESC").Find(&posts).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not list posts")
	}

	for i := range posts {
		h.attachImageURL(&posts[i])
	}

	return c.JSON(http.StatusOK, posts)
}

func (h *PostHandler) Get(c echo.Context) error {
	var post models.Post
	if err := h.db.Preload("User").First(&post, c.Param("id")).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "post not found")
	}

	h.attachImageURL(&post)
	return c.JSON(http.StatusOK, post)
}

func (h *PostHandler) attachImageURL(post *models.Post) {
	filename := filepath.Base(post.ImagePath)
	post.ImageURL = fmt.Sprintf("%s/uploads/%s", strings.TrimRight(h.cfg.PublicBaseURL, "/"), filename)
}

func extensionForContentType(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ""
	}
}
