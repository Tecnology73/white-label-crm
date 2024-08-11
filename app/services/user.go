package services

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"white-label-crm/app/models"
	"white-label-crm/database"
)

type UserService struct {
}

func NewUserService() *UserService {
	return &UserService{}
}

func (u *UserService) RegisterRoutes(router *fiber.App) {
	api := router.Group("/users")

	api.Get("/", u.list)
	api.Post("/", u.create)
	api.Get("/:user", u.read)
}

func (u *UserService) list(ctx *fiber.Ctx) error {
	users, err := database.Find[models.User](
		database.GetBrandDb(ctx),
		context.TODO(),
		bson.M{},
		options.Find().SetLimit(10),
	)
	if err != nil {
		log.Print(err)
		return ctx.SendStatus(500)
	}

	return ctx.JSON(users)
}

func (u *UserService) create(ctx *fiber.Ctx) error {
	user := &models.User{
		Model: database.NewModel(ctx),
		Name:  "John Doe",
		Email: "john.doe@mail.com",
	}

	_, err := database.InsertOne[*models.User](database.GetBrandDb(ctx), context.TODO(), user)
	if err != nil {
		return ctx.SendStatus(500)
	}

	return ctx.JSON(user)

	/*users := make([]*models.User, 5)
	for i := range users {
		users[i] = &models.User{
			Name:  "John Doe",
			Email: "john.doe@mail.com",
		}
	}

	err := database.InsertMany(context.TODO(), users)
	if err != nil {
		log.Printf("[UserService.create] %v\n", err)
		return ctx.SendStatus(500)
	}

	return ctx.JSON(users)*/
}

func (u *UserService) read(ctx *fiber.Ctx) error {
	id, err := primitive.ObjectIDFromHex(ctx.Params("user"))
	if err != nil {
		return ctx.SendStatus(404)
	}

	user, err := database.FindOne[models.User](
		database.GetBrandDb(ctx),
		context.TODO(),
		bson.M{"_id": id},
	)
	if err != nil {
		return ctx.SendStatus(404)
	}

	return ctx.JSON(user)
}
