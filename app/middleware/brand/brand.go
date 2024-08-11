package brand

import (
	"github.com/gofiber/fiber/v2"
)

type Config struct {
	Brands map[string]map[string]string
}

func New(config Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		brandSlug := c.Query("brand")
		brand, exists := config.Brands[brandSlug]
		if !exists {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		c.Locals("brand", brand)
		return c.Next()
	}
}
