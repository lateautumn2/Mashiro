package main

import (
	"log"

	"backend/config"
	"backend/db"
	"backend/notifier"
	"backend/routes"
)

func main() {
	db.InitDB()

	notifier.StartChecker()

	r := routes.SetupRouter()
	port := config.ServerPort()

	log.Printf("Backend server is running on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
