package brand

import (
	"github.com/gofiber/fiber/v2"
	"white-label-crm/app/models"
)

type Config struct {
	Brands map[string]models.Brand
}

func New(config Config) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		brandSlug := ctx.Query("brand")
		brand, exists := config.Brands[brandSlug]
		if !exists {
			return ctx.SendStatus(fiber.StatusNotFound)
		}

		ctx.Locals("dbName", brand.Slug)
		ctx.Locals("brand", brand)
		return ctx.Next()
	}
}
