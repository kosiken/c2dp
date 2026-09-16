package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"c2dp/go-api/internal/config"
	"c2dp/go-api/internal/models"
)

type AuthHandler struct {
	cfg config.Config
	db  *gorm.DB
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

type loginRequest struct {
	Identifier string `json:"identifier"`
	Email      string `json:"email"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

type authResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

func NewAuthHandler(db *gorm.DB, cfg config.Config) *AuthHandler {
	return &AuthHandler{cfg: cfg, db: db}
}

func (h *AuthHandler) Signup(c echo.Context) error {
	var req authRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Username = strings.TrimSpace(req.Username)
	if req.Email == "" || req.Username == "" || len(req.Password) < 8 {
		return echo.NewHTTPError(http.StatusBadRequest, "email, username, and password with at least 8 characters are required")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not hash password")
	}

	user := models.User{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: string(passwordHash),
	}
	if err := h.db.Create(&user).Error; err != nil {
		return echo.NewHTTPError(http.StatusConflict, "email or username already exists")
	}

	token, err := h.issueToken(user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not issue token")
	}

	return c.JSON(http.StatusCreated, authResponse{Token: token, User: user})
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	identifier := strings.TrimSpace(req.Identifier)
	// Preserve existing email clients and accept username-only API requests.
	if identifier == "" {
		identifier = strings.TrimSpace(req.Email)
	}
	if identifier == "" {
		identifier = strings.TrimSpace(req.Username)
	}
	if identifier == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email or username and password are required")
	}

	var users []models.User
	// Usernames retain their existing case-sensitive uniqueness semantics.
	// Reject ambiguous matches instead of selecting an arbitrary account.
	if err := h.db.Where("email = ? OR username = ?", strings.ToLower(identifier), identifier).Limit(2).Find(&users).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not login")
	}
	if len(users) != 1 {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid email, username, or password")
	}
	user := users[0]
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid email, username, or password")
	}

	token, err := h.issueToken(user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not issue token")
	}

	return c.JSON(http.StatusOK, authResponse{Token: token, User: user})
}

func (h *AuthHandler) issueToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"user_id": userID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.cfg.JWTSecret))
}
