package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
	"white-label-crm/app/database"
	"white-label-crm/app/middleware/auth"
	"white-label-crm/app/middleware/brand"
	"white-label-crm/app/services"
)

type ApiService interface {
	RegisterRoutes(router *fiber.App)
}

func main() {
	database.NewConnection(
		options.Client().
			SetAuth(
				options.Credential{
					Username: "root",
					Password: "root",
				},
			).
			SetConnectTimeout(5 * time.Second),
	)

	http := fiber.New()
	// Logging
	http.Use(
		logger.New(
			logger.Config{
				Format: "[${ip}:${port}] ${status} - ${method} ${path}\n",
			},
		),
	)
	// Brand detection
	http.Use(
		brand.New(
			brand.Config{
				Brands: map[string]map[string]string{
					"alpha": {
						"Name": "Brand Alpha",
					},
					"bravo": {
						"Name": "Brand Bravo",
					},
				},
			},
		),
	)
	// Global authentication
	http.Use(
		auth.New(
			auth.Config{
				ExcludePaths: []string{"/login", "/register"},
			},
		),
	)

	apiServices := []ApiService{
		services.NewAuthService(),
		services.NewUserService(),
	}

	for _, service := range apiServices {
		service.RegisterRoutes(http)
	}

	if err := http.Listen(":42069"); err != nil {
		log.Fatalf("[http.Listen] %v\n", err)
	}
}
