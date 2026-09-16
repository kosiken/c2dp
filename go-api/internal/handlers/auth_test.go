package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"c2dp/go-api/internal/config"
	"c2dp/go-api/internal/models"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Run with TEST_DATABASE_URL pointing to a test PostgreSQL database.
// All schema changes and fixtures are rolled back.
func TestLoginIdentifiers(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
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
	name := "Case" + strings.ReplaceAll(uuid.NewString(), "-", "")[:16]
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	user := models.User{Username: name, Email: strings.ToLower(name) + "@example.com", PasswordHash: string(hash)}
	if err := tx.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	handler := NewAuthHandler(tx, config.Config{JWTSecret: "integration-test-secret"})
	tests := []struct {
		name   string
		input  map[string]string
		status int
	}{
		{"email", map[string]string{"identifier": user.Email}, 200},
		{"username", map[string]string{"identifier": user.Username}, 200},
		{"normalized email", map[string]string{"identifier": "  " + strings.ToUpper(user.Email) + "  "}, 200},
		{"trimmed username", map[string]string{"identifier": "  " + user.Username + "  "}, 200},
		{"legacy email", map[string]string{"email": user.Email}, 200},
		{"username field", map[string]string{"username": user.Username}, 200},
		{"wrong password", map[string]string{"identifier": user.Username, "password": "wrong"}, 401},
		{"unknown identifier", map[string]string{"identifier": "missing-" + name}, 401},
		{"username case", map[string]string{"identifier": strings.ToLower(user.Username)}, 401},
		{"missing identifier", map[string]string{}, 400},
		{"missing password", map[string]string{"identifier": user.Email, "password": ""}, 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, ok := tt.input["password"]; !ok {
				tt.input["password"] = "password123"
			}
			body, _ := json.Marshal(tt.input)
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(string(body)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			if err := handler.Login(c); err != nil {
				e.HTTPErrorHandler(err, c)
			}
			if rec.Code != tt.status {
				t.Fatalf("status %d, want %d: %s", rec.Code, tt.status, rec.Body.String())
			}
			if tt.status == 200 {
				var result authResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
					t.Fatal(err)
				}
				if result.User.ID != user.ID || result.Token == "" {
					t.Fatal("expected original user and session token")
				}
			}
		})
	}
	t.Run("ambiguous identifier", func(t *testing.T) {
		other := models.User{Username: user.Email, Email: "other-" + user.Email, PasswordHash: string(hash)}
		if err := tx.Create(&other).Error; err != nil {
			t.Fatal(err)
		}
		body, _ := json.Marshal(map[string]string{"identifier": user.Email, "password": "password123"})
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(string(body)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		if err := handler.Login(c); err != nil {
			e.HTTPErrorHandler(err, c)
		}
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("ambiguous login returned %d", rec.Code)
		}
	})
}
