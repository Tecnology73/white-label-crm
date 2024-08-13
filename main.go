package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	redis2 "github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
	"white-label-crm/app/middleware/auth"
	"white-label-crm/app/middleware/brand"
	"white-label-crm/app/services"
	"white-label-crm/database"
	"white-label-crm/redis"
)

type ApiService interface {
	RegisterRoutes(router *fiber.App)
}

func initDatabase() {
	dbClient := database.NewConnection(
		options.Client().
			SetAuth(
				options.Credential{
					Username: "root",
					Password: "root",
				},
			).
			SetConnectTimeout(5 * time.Second),
	)

	watcher := database.NewWatcher(dbClient)
	err := watcher.Start()
	if err != nil {
		log.Fatalf("[watcher.Start] %v\n", err)
	}
}

func initRedis() {
	redis.NewConnection(
		&redis2.Options{
			Addr:     "127.0.0.1:6379",
			Username: "",
			Password: "",
			DB:       0,
		},
	)
}

func main() {
	initDatabase()
	initRedis()

	http := fiber.New()
	http.Use(pprof.New())
	// Logging
	/*http.Use(
		logger.New(
			logger.Config{
				Format: "[${ip}:${port}] ${status} - ${method} ${path}\n",
			},
		),
	)*/
	// Brand detection
	http.Use(brand.New())
	// Global authentication
	http.Use(
		auth.New(
			auth.Config{
				ExcludePaths: []string{"/login", "/register"},
			},
		),
	)

	apiServices := []ApiService{
		services.NewAuthService(&services.AuthOptions{Throughput: 10}),
		services.NewUserService(),
		services.NewCrudService(),
	}

	for _, service := range apiServices {
		service.RegisterRoutes(http)
	}

	if err := http.Listen(":42069"); err != nil {
		log.Fatalf("[http.Listen] %v\n", err)
	}
}
