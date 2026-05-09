package config

import (
	"log"
	"os"
	"strings"
)

const (
	defaultPort          = "8080"
	defaultDatabasePath  = "mashiro.db"
	defaultAdminUsername = "admin"
)

func ServerPort() string {
	port := strings.TrimSpace(os.Getenv("MASHIRO_PORT"))
	if port == "" {
		return defaultPort
	}
	return port
}

func DatabasePath() string {
	path := strings.TrimSpace(os.Getenv("MASHIRO_DB_PATH"))
	if path == "" {
		return defaultDatabasePath
	}
	return path
}

func JWTSecret() []byte {
	secret := strings.TrimSpace(os.Getenv("MASHIRO_JWT_SECRET"))
	if secret == "" {
		log.Fatal("MASHIRO_JWT_SECRET is required")
	}
	return []byte(secret)
}

func AdminUsername() string {
	username := strings.TrimSpace(os.Getenv("MASHIRO_ADMIN_USERNAME"))
	if username == "" {
		return defaultAdminUsername
	}
	return username
}

func AdminPassword() string {
	return strings.TrimSpace(os.Getenv("MASHIRO_ADMIN_PASSWORD"))
}
