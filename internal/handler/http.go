package handler

import (
	"log"

	"pay-as-you-use/internal/service"

	"github.com/gofiber/fiber/v2"
)

type PaymentHandler struct {
	paymentService service.PaymentService
}

func NewPaymentHandler(svc service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: svc,
	}
}

func (h *PaymentHandler) GetBalance(c *fiber.Ctx) error {
	userID := c.Params("user_id")

	balance, err := h.paymentService.GetUserBalance(c.Context(), userID)
	if err != nil {
		log.Printf("Error getting balance: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve balance",
		})
	}

	return c.JSON(fiber.Map{
		"user_id": userID,
		"balance": balance,
	})
}

func (h *PaymentHandler) SetupRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	api.Get("/balance/:user_id", h.GetBalance)
}
