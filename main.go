package main

import (
	"backend-elearning/config"
	"fmt"
	"log"

	"backend-elearning/database"

	"github.com/gofiber/fiber/v2"
	"backend-elearning/routes"
)

func main() {
	cfg := config.LoadConfig()

	database.ConnectDB(cfg)

	app := fiber.New()

	routes.SetupRoutes(app)

	log.Fatal(app.Listen(fmt.Sprintf(":%s", cfg.AppPort)))
}
