package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pay-as-you-use/internal/database"
	"pay-as-you-use/internal/handler"
	"pay-as-you-use/internal/repository"
	"pay-as-you-use/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	dbPool, err := database.NewPool()
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	defer dbPool.Close()

	userRepo := repository.NewUserRepository(dbPool)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, OPTIONS",
	}))
	app.Use(logger.New())

	userHandler.SetupRoutes(app)

	port := os.Getenv("USER_PORT")
	if port == "" {
		port = "8002"
	}
	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down user service...")
	app.ShutdownWithTimeout(5 * time.Second)
}
