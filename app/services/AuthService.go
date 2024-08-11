package services

import (
	"github.com/gofiber/fiber/v2"
)

type AuthService struct {
}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) RegisterRoutes(router *fiber.App) {
	router.Post("/login", s.login)
	router.Post("/register", s.register)
}

func (s *AuthService) login(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func (s *AuthService) register(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}
