package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Tahseen-Zaman/url_shortener/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)
func setupRoutes(app *fiber.App) {
	app.Get("/:url", routes.ResolveURL)
	app.Post("/api/v1", routes.ShortenURL)
}

func main() {
	err:= godotenv.Load()
	if err != nil {
		fmt.Print("Error loading .env file")
	}
	app := fiber.New()
	app.Use(logger.New())
	setupRoutes(app)
	log.Fatal(app.Listen(os.Getenv("APP_PORT")))
}