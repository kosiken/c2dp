package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL       string
	JWTSecret         string
	Port              string
	PublicBaseURL     string
	UploadDir         string
	C2PAToolPath      string
	PythonPath        string
	CombineScriptPath string
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/c2dp_go_api?sslmode=disable"),
		JWTSecret:         getEnv("JWT_SECRET", "dev-secret"),
		Port:              getEnv("PORT", "8080"),
		PublicBaseURL:     getEnv("PUBLIC_BASE_URL", "http://localhost:8080"),
		UploadDir:         getEnv("UPLOAD_DIR", "uploads"),
		C2PAToolPath:      getEnv("C2PATOOL_PATH", "bin/c2patool"),
		PythonPath:        getEnv("PYTHON_PATH", "python3"),
		CombineScriptPath: getEnv("COMBINE_SCRIPT_PATH", "../scripts/combine_signed_images.py"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
