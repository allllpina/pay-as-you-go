package handler

import (
	"pay-as-you-use/internal/service"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{userService: svc}
}

type CreateUserRequest struct {
	Email string `json:"email"`
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil || req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request or missing email"})
	}

	id, err := h.userService.CreateUser(c.Context(), req.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":    id,
		"email": req.Email,
	})
}

func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")
	user, err := h.userService.GetUser(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}
	return c.JSON(user)
}

func (h *UserHandler) SetupRoutes(app *fiber.App) {
	api := app.Group("/api/v1/users")
	api.Post("/", h.CreateUser)
	api.Get("/:id", h.GetUser)
}
