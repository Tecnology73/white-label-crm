package auth

import (
	"github.com/gofiber/fiber/v2"
	"slices"
)

type Config struct {
	ExcludePaths []string
}

func New(config Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()
		if slices.Contains(config.ExcludePaths, path) {
			return c.Next()
		}

		// TODO
		return c.Next()
	}
}
