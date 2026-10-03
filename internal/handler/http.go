package handler

import (
	"encoding/json"
	"log"
	"os"

	"pay-as-you-use/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/stripe/stripe-go/v86"
	"github.com/stripe/stripe-go/v86/webhook"
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

func (h *PaymentHandler) CreateCheckout(c *fiber.Ctx) error {
	userID := c.Params("user_id")

	priceID := "price_1UMTP0Gb0JTyNFJi82HjL8br"

	url, err := h.paymentService.CreateCheckoutSession(c.Context(), userID, priceID)
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

	if event.Type == "checkout.session.completed" {
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			log.Printf("Error parsing webhook JSON: %v", err)
			return c.Status(fiber.StatusBadRequest).SendString("Invalid payload")
		}

		userID := session.ClientReferenceID
		log.Printf("Payment success for user: %s", userID)

		err = h.paymentService.AddDeposit(c.Context(), userID, 1000)
		if err != nil {
			log.Printf("Failed to process deposit for user %s: %v", userID, err)
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to save deposit")
		}

		log.Printf("Successfully added deposit of 1000 tokens for user: %s", userID)
	}
	// TODO: add else
	return c.SendStatus(fiber.StatusOK)
}

func (h *PaymentHandler) SetupRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	api.Get("/balance/:user_id", h.GetBalance)
	api.Post("/checkout/:user_id", h.CreateCheckout)
	api.Post("/webhook", h.HandleWebhook)
}
