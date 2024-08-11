package services

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"white-label-crm/app/database"
	"white-label-crm/app/models"
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

func (u *UserService) list(c *fiber.Ctx) error {
	users, err := database.Find[*models.User](
		context.TODO(),
		bson.M{},
		options.Find().SetLimit(10),
	)
	if err != nil {
		return c.SendStatus(500)
	}

	return c.JSON(users)
}

func (u *UserService) create(c *fiber.Ctx) error {
	user := &models.User{
		Name:  "John Doe",
		Email: "john.doe@mail.com",
	}

	err := database.Insert(context.TODO(), user)
	if err != nil {
		return c.SendStatus(500)
	}

	return c.JSON(user)

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
		return c.SendStatus(500)
	}

	return c.JSON(users)*/
}

func (u *UserService) read(c *fiber.Ctx) error {
	id, err := primitive.ObjectIDFromHex(c.Params("user"))
	if err != nil {
		return c.SendStatus(404)
	}

	user, err := database.FindOne[*models.User](
		context.TODO(),
		bson.M{"_id": id},
	)
	if err != nil {
		return c.SendStatus(404)
	}

	return c.JSON(user)
}
