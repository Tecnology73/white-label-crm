package auth

import (
	"github.com/gofiber/fiber/v2"
	"slices"
	"white-label-crm/database"
)

type Config struct {
	ExcludePaths []string
}

func New(config Config) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		path := ctx.Path()
		if slices.Contains(config.ExcludePaths, path) {
			ctx.Locals(
				"user",
				database.UserRelation{
					ID:   0,
					Name: "System",
				},
			)

			return ctx.Next()
		}

		ctx.Locals(
			"user",
			database.UserRelation{
				ID:   1,
				Name: "Admin",
			},
		)

		return ctx.Next()
	}
}
