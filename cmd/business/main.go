package main

import (
	"log"
	"os"
	"os/signal"
	"pay-as-you-use/internal/client"
	"pay-as-you-use/internal/database"
	"pay-as-you-use/internal/handler"
	"pay-as-you-use/internal/repository"
	"pay-as-you-use/internal/service"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	paymentURL := os.Getenv("PAYMENT_SERVICE_URL")
	if paymentURL == "" {
		paymentURL = "http://localhost:8000"
	}

	dbPool, err := database.NewPool()
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	defer dbPool.Close()

	businessRepo := repository.NewBusinessRepository(dbPool)
	paymentClient := client.NewPaymentClient(paymentURL)
	businessService := service.NewBusinessService(businessRepo, paymentClient)
	businessHandler := handler.NewBusinessHandler(businessService)

	app := fiber.New()
	app.Use(logger.New())

	businessHandler.SetupRoutes(app)

	port := os.Getenv("BUSINESS_PORT")
	if port == "" {
		port = "8001"
	}

	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down business service...")
	app.ShutdownWithTimeout(5 * time.Second)
}
