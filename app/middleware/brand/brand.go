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
		hostname := ctx.Hostname()
		// Remove port from hostname (dev environment)
		if idx := strings.Index(hostname, ":"); idx > -1 {
			hostname = hostname[:idx]
			if len(strings.TrimSpace(hostname)) == 0 {
				log.Printf("[brand middleware] Invalid hostname | Raw(%v) | Parsed(%v)\n", ctx.Hostname(), hostname)
				return ctx.SendStatus(fiber.StatusNotFound)
			}
		}

		// Lookup brand
		data, err := redis.Client.HGetAll(context.Background(), fmt.Sprintf("brands:%s", hostname)).Result()
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
			Model:  database.Model{ID: id},
			Name:   data["name"],
			Slug:   data["slug"],
			Domain: data["domain"],
		}

		ctx.Locals("dbName", fmt.Sprintf("brand_%s", brand.Slug))
		ctx.Locals("brand", brand)
		return ctx.Next()
	}
}
