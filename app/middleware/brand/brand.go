package brand

import (
	"context"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	redis2 "github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"strings"
	"white-label-crm/app/models"
	"white-label-crm/database"
	"white-label-crm/redis"
)

func New() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		brandSlug := strings.TrimSpace(ctx.Query("brand"))
		// TODO: Restrict slug to a min & max length.
		// The slug should probably be something like `maxLength(brand.Slug, 16) + "-" + strings.random(8)`
		if len(brandSlug) == 0 {
			log.Printf("[brand middleware] Invalid slug | %s\n", ctx.Query("brand"))
			return ctx.SendStatus(fiber.StatusNotFound)
		}

		// Lookup brand
		data, err := redis.Client.HGetAll(context.Background(), fmt.Sprintf("brands:%s", brandSlug)).Result()
		if err != nil {
			if !errors.Is(err, redis2.Nil) {
				log.Printf("[brand middleware] %v\n", err)
			}

			return ctx.SendStatus(fiber.StatusNotFound)
		}

		id, err := primitive.ObjectIDFromHex(data["_id"])
		if err != nil {
			log.Printf("[brand middleware] %v\n", err)
			return ctx.SendStatus(fiber.StatusNotFound)
		}

		// Store the brand on the context
		brand := models.Brand{
			Model: database.Model{ID: id},
			Slug:  data["slug"],
			Name:  data["Name"],
		}

		ctx.Locals("dbName", brand.Slug)
		ctx.Locals("brand", brand)
		return ctx.Next()
	}
}
