package handler

import (
	"errors"
	"log"

	"pay-as-you-use/internal/domain"
	"pay-as-you-use/internal/service"

	"github.com/gofiber/fiber/v2"
)

type BusinessHandler struct {
	businessService service.BusinessService
}

func NewBusinessHandler(svc service.BusinessService) *BusinessHandler {
	return &BusinessHandler{
		businessService: svc,
	}
}

type ConsumeRequest struct {
	ActionType string `json:"action_type"`
}

func (h *BusinessHandler) Consume(c *fiber.Ctx) error {
	userID := c.Params("user_id")

	var req ConsumeRequest
	if err := c.BodyParser(&req); err != nil || req.ActionType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid action_type"})
	}

	// HARDCODE FOR MVP
	cost := 0
	switch req.ActionType {
	case "generate_image":
		cost = 5
	case "translate_text":
		cost = 2
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "unknown action_type"})
	}

	err := h.businessService.ExecuteAction(c.Context(), userID, req.ActionType, cost)
	if err != nil {
		if errors.Is(err, domain.ErrInsufficientFunds) {
			return c.Status(fiber.StatusPaymentRequired).JSON(fiber.Map{
				"error": "insufficient funds",
			})
		}

		log.Printf("Failed to execute action for user %s: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal server error",
		})
	}
	return c.JSON(fiber.Map{
		"message": "Action executed successfully",
		"cost":    cost,
	})
}

func (h *BusinessHandler) SetupRoutes(app *fiber.App) {
	api := app.Group("/api/v1/business")
	api.Post("/consume/:user_id", h.Consume)
}
