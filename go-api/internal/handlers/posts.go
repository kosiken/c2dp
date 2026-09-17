package handlers

import (
	"encoding/json"
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
	cfg         config.Config
	db          *gorm.DB
	broadcaster *Broadcaster
}

func NewPostHandler(db *gorm.DB, cfg config.Config) *PostHandler {
	return &PostHandler{cfg: cfg, db: db, broadcaster: NewBroadcaster()}
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
	// Keep unsigned inputs outside the publicly served upload directory.
	uploadDir, err := filepath.Abs(h.cfg.UploadDir)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not prepare upload directory")
	}
	stage, err := os.MkdirTemp("", "c2dp-upload-")
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
	manifest, err := signing.Inspect(c.Request().Context(), h.cfg.C2PAToolPath, signed)
	if err != nil {
		c.Logger().Errorf("inspect signed upload: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "could not read signed image manifest; no post was created")
	}
	imagePath := filepath.Join(uploadDir, uuid.NewString()+ext)
	if err := publishSigned(signed, imagePath); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not publish signed image")
	}

	post := models.Post{
		Caption:      strings.TrimSpace(c.FormValue("caption")),
		ImagePath:    imagePath,
		UserID:       userID,
		ManifestData: models.C2PAManifest(manifest),
	}
	if err := h.db.Create(&post).Error; err != nil {
		_ = os.Remove(imagePath)
		return echo.NewHTTPError(http.StatusInternalServerError, "could not create post")
	}

	if err := h.db.Preload("User").First(&post, post.ID).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not load post")
	}
	h.attachImageURL(&post)
	h.broadcaster.Publish(post)

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

// Stream pushes each newly created post over SSE as it happens, for viewers
// such as the presentation's real-time slide. It sends no backlog; callers
// that want existing posts too should also call List on connect.
func (h *PostHandler) Stream(c echo.Context) error {
	resp := c.Response()
	resp.Header().Set(echo.HeaderContentType, "text/event-stream")
	resp.Header().Set("Cache-Control", "no-cache")
	resp.Header().Set("Connection", "keep-alive")
	resp.WriteHeader(http.StatusOK)
	resp.Flush()

	ch := h.broadcaster.Subscribe()
	defer h.broadcaster.Unsubscribe(ch)

	ctx := c.Request().Context()
	for {
		select {
		case <-ctx.Done():
			return nil
		case post, ok := <-ch:
			if !ok {
				return nil
			}
			data, err := json.Marshal(post)
			if err != nil {
				continue
			}
			if _, err := fmt.Fprintf(resp, "data: %s\n\n", data); err != nil {
				return nil
			}
			resp.Flush()
		}
	}
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

// Copy into the destination filesystem before renaming, since uploads may be a
// Docker volume. Only signed bytes enter the public directory.
func publishSigned(source, destination string) error {
	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.CreateTemp(filepath.Dir(destination), ".signed-")
	if err != nil {
		return err
	}
	defer os.Remove(dst.Name())
	_, copyErr := io.Copy(dst, src)
	closeErr := dst.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(dst.Name(), destination)
}
