package handler

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"pay-as-you-use/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/stripe/stripe-go/v86"
	"github.com/stripe/stripe-go/v86/webhook"
)

type CreateCheckoutRequest struct {
	PackageID string `json:"package_id"`
}

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

func (h *PaymentHandler) CreateCheckout(c *fiber.Ctx) error {
	userID := c.Params("user_id")

	var req CreateCheckoutRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Валідація на порожнє значення
	if req.PackageID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "package_id is required",
		})
	}

	url, err := h.paymentService.CreateCheckoutSession(c.Context(), userID, req.PackageID)
	if err != nil {
		log.Printf("Error creating checkout: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create checkout session"})
	}

	return c.JSON(fiber.Map{"checkout_url": url})
}

func (h *PaymentHandler) HandleWebhook(c *fiber.Ctx) error {
	payload := c.Body()
	signature := c.Get("Stripe-Signature")
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")

	event, err := webhook.ConstructEvent(payload, signature, webhookSecret)
	if err != nil {
		log.Printf("Webhook signature verification failed: %v", err)
		return c.Status(fiber.StatusBadRequest).SendString("Invalid payload")
	}
	if event.Type == "checkout.session.completed" ||
		event.Type == "checkout.session.expired" ||
		event.Type == "checkout.session.async_payment_failed" {

		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			log.Printf("Error parsing webhook JSON: %v", err)
			return c.Status(fiber.StatusBadRequest).SendString("Invalid payload")
		}

		// userID := session.ClientReferenceID
		requestID := session.Metadata["request_id"]

		if requestID == "" {
			log.Printf("Warning: no request_id found in session metadata for event %s", requestID)
			return c.SendStatus(fiber.StatusOK)
		}

		switch event.Type {
		case "checkout.session.completed":
			log.Printf("Payment success for request: %s", requestID)
			err = h.paymentService.ProcessSuccessfulPayment(c.Context(), requestID)
			if err != nil {
				log.Printf("Failed to process successful deposit: %v", err)
				if strings.Contains(err.Error(), "business error") {
					return c.SendStatus(fiber.StatusOK)
				}
				return c.Status(fiber.StatusInternalServerError).SendString("DB error")
			}
			log.Printf("Successfully added deposit for request: %s", requestID)
		case "checkout.session.expired", "checkout.session.async_payment_failed":
			log.Printf("Payment failed/expired for request: %s. Event: %s", requestID, event.Type)
			err := h.paymentService.ProcessFailedPayment(c.Context(), requestID)
			if err != nil {
				log.Printf("Failed to process failed deposit status: %v", err)
				return c.Status(fiber.StatusInternalServerError).SendString("DB error")
			}
		}
	}
	return c.SendStatus(fiber.StatusOK)
}

func (h *PaymentHandler) SetupRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	api.Get("/balance/:user_id", h.GetBalance)
	api.Post("/checkout/:user_id", h.CreateCheckout)
	api.Post("/webhook", h.HandleWebhook)
}
