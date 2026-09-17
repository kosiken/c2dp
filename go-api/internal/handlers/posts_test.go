package handlers

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"c2dp/go-api/internal/config"
	appmiddleware "c2dp/go-api/internal/middleware"
	"c2dp/go-api/internal/models"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestSignedUploads(t *testing.T) {
	dsn, tool := os.Getenv("TEST_DATABASE_URL"), os.Getenv("TEST_C2PATOOL_PATH")
	if dsn == "" || tool == "" {
		t.Skip("set TEST_DATABASE_URL and TEST_C2PATOOL_PATH")
	}
	tool, err := filepath.Abs(tool)
	if err != nil {
		t.Fatal(err)
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	defer tx.Rollback()
	if err := tx.AutoMigrate(&models.User{}, &models.Post{}); err != nil {
		t.Fatal(err)
	}
	user := models.User{Username: uuid.NewString(), Email: uuid.NewString() + "@example.com", PasswordHash: "unused"}
	if err := tx.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	cfg := config.Config{UploadDir: filepath.Join(root, "uploads"), C2PAToolPath: tool, PublicBaseURL: "http://localhost:8080"}
	handler := NewPostHandler(tx, cfg)
	upload := func(data []byte) *httptest.ResponseRecorder {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, err := writer.CreateFormFile("image", "untrusted-name.html")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write(data); err != nil {
			t.Fatal(err)
		}
		writer.WriteField("caption", "test")
		writer.Close()
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/posts", &body)
		req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(appmiddleware.UserIDKey, user.ID)
		if err := handler.Create(c); err != nil {
			e.HTTPErrorHandler(err, c)
		}
		return rec
	}
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	var raw bytes.Buffer
	if err := png.Encode(&raw, img); err != nil {
		t.Fatal(err)
	}
	var previous []byte
	for i := 0; i < 2; i++ {
		input := raw.Bytes()
		if i > 0 {
			input = previous
		}
		rec := upload(input)
		if rec.Code != 201 {
			t.Fatalf("upload %d: %d %s", i, rec.Code, rec.Body.String())
		}
		var post models.Post
		if err := json.Unmarshal(rec.Body.Bytes(), &post); err != nil {
			t.Fatal(err)
		}
		filename := filepath.Join(cfg.UploadDir, filepath.Base(post.ImageURL))
		if filepath.Ext(filename) != ".png" {
			t.Fatal("extension must follow content")
		}
		previous, err = os.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}
		// Exercise the same static-serving mechanism used by the API.
		e := echo.New()
		e.Static("/uploads", cfg.UploadDir)
		served := httptest.NewRecorder()
		e.ServeHTTP(served, httptest.NewRequest(http.MethodGet, "/uploads/"+filepath.Base(filename), nil))
		if served.Code != 200 || !bytes.Equal(served.Body.Bytes(), previous) {
			t.Fatal("download changed signed bytes")
		}
		downloaded := filepath.Join(root, "download.png")
		if err := os.WriteFile(downloaded, served.Body.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
		output, err := exec.Command(tool, downloaded).Output()
		if err != nil {
			t.Fatalf("inspect signed image: %v", err)
		}
		var report struct {
			Manifests         map[string]json.RawMessage `json:"manifests"`
			ValidationResults struct {
				ActiveManifest struct {
					Success []struct {
						Code string `json:"code"`
					} `json:"success"`
					Failure []struct {
						Code string `json:"code"`
					} `json:"failure"`
				} `json:"activeManifest"`
			} `json:"validation_results"`
		}
		if err := json.Unmarshal(output, &report); err != nil {
			t.Fatal(err)
		}
		if len(report.Manifests) != i+1 {
			t.Fatalf("expected %d manifests, got %d", i+1, len(report.Manifests))
		}
		found := map[string]bool{}
		for _, s := range report.ValidationResults.ActiveManifest.Success {
			found[s.Code] = true
		}
		if !found["claimSignature.validated"] || !found["assertion.dataHash.match"] {
			t.Fatal("signature or asset binding not validated")
		}
		for _, f := range report.ValidationResults.ActiveManifest.Failure {
			if f.Code != "signingCredential.untrusted" {
				t.Fatalf("unexpected validation failure: %s", f.Code)
			}
		}
	}
	for _, tc := range []struct {
		name   string
		data   []byte
		tool   string
		status int
	}{
		{"non-image", []byte("<html>not an image</html>"), tool, 400},
		{"missing signer", raw.Bytes(), filepath.Join(root, "missing-tool"), 500},
		{"invalid image", []byte("\x89PNG\r\n\x1a\ninvalid"), tool, 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handler.cfg.C2PAToolPath = tc.tool
			before, _ := os.ReadDir(cfg.UploadDir)
			rec := upload(tc.data)
			after, _ := os.ReadDir(cfg.UploadDir)
			if rec.Code != tc.status {
				t.Fatalf("got %d want %d", rec.Code, tc.status)
			}
			if len(before) != len(after) {
				t.Fatal("failed upload left public file")
			}
			var count int64
			if err := tx.Model(&models.Post{}).Where("user_id = ?", user.ID).Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if count != 2 {
				t.Fatal("failed upload created post")
			}
			entries, _ := os.ReadDir(root)
			for _, entry := range entries {
				if entry.IsDir() && entry.Name() != "uploads" {
					t.Fatal("staging directory leaked")
				}
			}
		})
	}
}
