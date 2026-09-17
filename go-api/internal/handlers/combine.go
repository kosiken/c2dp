package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"c2dp/go-api/internal/models"
)

// Combine picks the four most recently uploaded posts and runs
// scripts/combine_signed_images.py to fold their signed images and C2PA
// manifests into one new signed composite, returning its public URL
// alongside the composite's own manifest report.
func (h *PostHandler) Combine(c echo.Context) error {
	var posts []models.Post
	if err := h.db.Order("created_at DESC").Limit(4).Find(&posts).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not select images to combine")
	}
	if len(posts) < 4 {
		return echo.NewHTTPError(http.StatusBadRequest, "need at least 4 uploaded posts to combine")
	}

	pythonPath, err := exec.LookPath(h.cfg.PythonPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "python3 not found; set PYTHON_PATH")
	}
	scriptPath, err := filepath.Abs(h.cfg.CombineScriptPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not resolve combine script path")
	}
	if _, err := os.Stat(scriptPath); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "combine script not found; set COMBINE_SCRIPT_PATH")
	}
	toolPath, err := exec.LookPath(h.cfg.C2PAToolPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "c2patool not found")
	}
	toolPath, err = filepath.Abs(toolPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not resolve c2patool path")
	}

	uploadDir, err := filepath.Abs(h.cfg.UploadDir)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not prepare upload directory")
	}
	outputName := "combined-" + uuid.NewString() + ".png"
	outputPath := filepath.Join(uploadDir, outputName)
	manifestPath := strings.TrimSuffix(outputPath, ".png") + ".manifest.json"
	recipePath := strings.TrimSuffix(outputPath, ".png") + ".recipe.json"
	// Sidecar reports are only scratch space; the manifest is read into the
	// response below and the recipe isn't served at all.
	defer os.Remove(manifestPath)
	defer os.Remove(recipePath)

	args := []string{scriptPath}
	for _, post := range posts {
		args = append(args, post.ImagePath)
	}
	args = append(args, "-o", outputPath, "--c2patool", toolPath)

	ctx, cancel := context.WithTimeout(c.Request().Context(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, pythonPath, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		c.Logger().Errorf("combine images: %v: %s", err, strings.TrimSpace(stderr.String()))
		if ctx.Err() != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "combining images timed out")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "could not generate combined image")
	}

	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "combined image is missing its manifest report")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"url":      fmt.Sprintf("%s/uploads/%s", strings.TrimRight(h.cfg.PublicBaseURL, "/"), outputName),
		"manifest": json.RawMessage(manifestBytes),
	})
}
