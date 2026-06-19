package main

import (
	"backend-elearning/config"
	"fmt"
	"log"

	"backend-elearning/database"

	"backend-elearning/routes"

	loggermw "github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	cfg := config.LoadConfig()

	database.ConnectDB(cfg)

	app := fiber.New()

	// Middleware: CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "https://nesyasal.github.io,http://localhost:5500,http://127.0.0.1:5500x",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	app.Use(loggermw.New(loggermw.Config{
		// Format: [Waktu] IP [WarnaStatus]Status[ResetWarna] - Latensi Method Path
		Format: "[${time}] ${ip} ${color}${status}${reset} - ${latency} ${method} ${path}\n",
	}))

	routes.SetupRoutes(app)

	log.Fatal(app.Listen(fmt.Sprintf(":%s", cfg.AppPort)))
}
