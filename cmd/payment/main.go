package main

import (
	"log"
	"os"

	"pay-as-you-use/internal/database"
	"pay-as-you-use/internal/handler"
	"pay-as-you-use/internal/repository"
	"pay-as-you-use/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system envs")
	}

	dbPool, err := database.NewPool()
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	defer dbPool.Close()

	ledgerRepo := repository.NewLedgerRepository(dbPool)
	tokenPackageRepo := repository.NewTokenPackageRepository(dbPool)
	depositReqRepo := repository.NewDepositRequestRepository(dbPool)
	depositStatRepo := repository.NewDepositStatusRepository(dbPool)

	stripeConfig := service.StripeConfig{
		SecretKey:   os.Getenv("STRIPE_SECRET_KEY"),
		BaseURL:     os.Getenv("BASE_URL"),
		FrontendURL: os.Getenv("FRONTEND_URL"),
	}
	paymentSvc := service.NewPaymentService(ledgerRepo, tokenPackageRepo, depositReqRepo, depositStatRepo, stripeConfig)

	paymentHandler := handler.NewPaymentHandler(paymentSvc)

	app := fiber.New()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "payment-service",
		})
	})

	paymentHandler.SetupRoutes(app)

	port := os.Getenv("PAYMENT_PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("Payment service is running on port %s...", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
