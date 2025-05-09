package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	"backend-pedika-fiber/database"
	"backend-pedika-fiber/migration"
	"backend-pedika-fiber/routes"
)

func main() {
	// 1. Load .env (jika ada)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env vars")
	}

	// 2. Inisialisasi Fiber
	app := fiber.New()

	// 3. Koneksi DB, migrasi, routes
	database.GetDBInstance()
	migration.RunMigration()
	routes.SetAuthRoutes(app)
	routes.SetAdminRoutes(app)
	routes.SetMasyarakatRoutes(app)
	routes.RoutesWithOutLogin(app)

	// 4. Start server di PORT dari env
	port := os.Getenv("PORT")
	if port == "" {
		port = "3060" // fallback
	}
	app.Listen(":" + port)
}
