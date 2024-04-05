package main

import (
	"TokenHoldersAnalyse/internal/delivery/http"
	"TokenHoldersAnalyse/internal/redisClient"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"log"
)

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	app := fiber.New()
	redisClient.InitRedis()
	http.InitHandlers(app)

	err := app.Listen(":3000")
	if err != nil {
		return
	}
}
