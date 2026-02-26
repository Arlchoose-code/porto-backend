package main

import (
	"arlchoose/backend-api/config"
	"arlchoose/backend-api/database"
	"arlchoose/backend-api/routes"
)

func main() {

	// Load config .env
	config.LoadEnv()

	// Inisialisasi database
	database.InitDB()

	// Setup router
	r := routes.SetupRouter()

	// Serve static files dari folder uploads/
	r.Static("/uploads", "./uploads")

	// Mulai server
	r.Run(":" + config.GetEnv("APP_PORT", "3000"))
}
