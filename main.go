package main

import (
	"backend-pedika-fiber/database"
	"backend-pedika-fiber/migration"
	"backend-pedika-fiber/routes"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Buat instance Fiber
	app := fiber.New()

	// Inisialisasi database dan jalankan migrasi
	database.GetDBInstance()
	migration.RunMigration()

	// Atur routing
	routes.SetAuthRoutes(app)
	routes.SetAdminRoutes(app)
	routes.SetMasyarakatRoutes(app)
	routes.RoutesWithOutLogin(app)

	// Ambil PORT dari environment variable (fallback ke 8080)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Listening on %s...", addr)

	// Jalankan aplikasi di PORT yang ditentukan
	log.Fatal(app.Listen(addr))
}
