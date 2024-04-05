package http

import (
	"TokenHoldersAnalyse/internal/delivery"
	"github.com/gofiber/fiber/v2"
)

func InitHandlers(app *fiber.App) {
	app.Get("/:tokenHash", delivery.FetchTransfersData)
}
