package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"c2dp/go-api/internal/config"
	"c2dp/go-api/internal/database"
	"c2dp/go-api/internal/handlers"
	appmiddleware "c2dp/go-api/internal/middleware"
	"c2dp/go-api/internal/models"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.Post{}); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.Static("/uploads", cfg.UploadDir)
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	authHandler := handlers.NewAuthHandler(db, cfg)
	postHandler := handlers.NewPostHandler(db, cfg)
	authMiddleware := appmiddleware.JWT(cfg.JWTSecret)

	v1 := e.Group("/api/v1")
	auth := v1.Group("/auth")
	auth.POST("/signup", authHandler.Signup)
	auth.POST("/login", authHandler.Login)

	posts := v1.Group("/posts")
	posts.GET("", postHandler.List)
	posts.GET("/:id", postHandler.Get)
	posts.POST("", postHandler.Create, authMiddleware)

	e.Logger.Fatal(e.Start(":" + cfg.Port))
}
