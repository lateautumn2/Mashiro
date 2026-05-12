package db

import (
	"log"

	"backend/config"
	"backend/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"golang.org/x/crypto/bcrypt"
)

var DB *gorm.DB

func InitDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open(config.DatabasePath()), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}

	// Auto Migrate the schema
	err = DB.AutoMigrate(
		&models.User{},
		&models.Server{},
		&models.AgentData{},
		&models.LatencyTask{},
		&models.LatencyTaskServer{},
		&models.LatencyResult{},
		&models.AppConfig{},
		&models.NotificationPref{},
	)
	if err != nil {
		log.Fatal("Failed to auto migrate database:", err)
	}

	// Seed default admin user
	seedAdmin()
}

func seedAdmin() {
	var count int64
	DB.Model(&models.User{}).Count(&count)
	if count == 0 {
		password := config.AdminPassword()
		if password == "" {
			log.Fatal("MASHIRO_ADMIN_PASSWORD is required when seeding the first admin user")
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal("Failed to hash default admin password:", err)
		}

		admin := models.User{
			Username: config.AdminUsername(),
			Password: string(hash),
		}
		DB.Create(&admin)
		log.Printf("Created default admin user (%s)", admin.Username)
	}
}
