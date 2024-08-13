package services

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"time"
	"white-label-crm/app/models"
	"white-label-crm/database"
	"white-label-crm/hash"
	"white-label-crm/utils"
)

type AuthService struct {
	limiter *utils.ThroughputLimiter
}

type AuthOptions struct {
	Throughput uint
}

func NewAuthService(opts *AuthOptions) *AuthService {
	return &AuthService{
		limiter: utils.NewThroughputLimiter(opts.Throughput),
	}
}

func (s *AuthService) RegisterRoutes(router *fiber.App) {
	router.Post("/login", s.login)
	router.Post("/register", s.register)
}

type loginRequest struct {
	Email    string `json:"email" form:"email"`
	Password string `json:"password" form:"password"`
}

func (s *AuthService) login(ctx *fiber.Ctx) error {
	// Limit requests
	lock, err := s.limiter.Acquire(5 * time.Second)
	if err != nil {
		log.Printf("[AuthService.login] %v\n", err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}
	defer s.limiter.Release(lock)

	// Parse body
	var data loginRequest
	if err := ctx.BodyParser(&data); err != nil {
		log.Printf("[AuthService.login] %v\n", err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	// Find user
	user, err := database.FindOne[models.User](
		database.GetBrandDb(ctx),
		context.TODO(),
		bson.M{"email": data.Email},
	)
	if err != nil {
		log.Printf("[AuthService.login] %v\n", err)
		return ctx.SendStatus(fiber.StatusForbidden)
	}

	// Check password
	if err := hash.Compare(data.Password, user.Password); err != nil {
		log.Printf("[AuthService.login] %v\n", err)
		return ctx.SendStatus(fiber.StatusForbidden)
	}

	// Login success
	log.Printf("[AuthService.login] Success: %v\n", user)
	return ctx.SendStatus(fiber.StatusNoContent)
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *AuthService) register(ctx *fiber.Ctx) error {
	// Limit requests
	lock, err := s.limiter.Acquire(5 * time.Second)
	if err != nil {
		log.Printf("[AuthService.login] %v\n", err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}
	defer s.limiter.Release(lock)

	// Parse body
	var data registerRequest
	if err := ctx.BodyParser(&data); err != nil {
		log.Printf("[AuthService.register] %v\n", err)
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	// Hash password
	password, err := hash.Hash(
		data.Password,
		&hash.Argon2Options{
			Time:       hash.PasswordTime,
			Memory:     hash.PasswordMemory,
			Threads:    hash.PasswordThreads,
			SaltLength: hash.PasswordSaltLength,
			KeyLength:  hash.PasswordKeyLength,
		},
	)
	if err != nil {
		log.Printf("[AuthService.register] %v\n", err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	// Create user
	user := models.User{
		Model:    database.NewModel(ctx),
		Email:    data.Email,
		Password: password,
	}

	_, err = database.InsertOne[*models.User](database.GetBrandDb(ctx), context.TODO(), &user)
	if err != nil {
		log.Printf("[AuthService.register] %v\n", err)
		return ctx.SendStatus(fiber.StatusBadRequest)
	}

	// Registration complete
	log.Printf("[AuthService.register] Complete: %v\n", user)
	return ctx.SendStatus(fiber.StatusNoContent)
}
