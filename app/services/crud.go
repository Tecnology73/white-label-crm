package services

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"reflect"
	"strings"
	"white-label-crm/app/models"
	"white-label-crm/database"
)

type CrudService struct{}

func NewCrudService() *CrudService {
	return &CrudService{}
}

func (c *CrudService) RegisterRoutes(router *fiber.App) {
	router.Get("/test", c.list)
	router.Put("/test/:record", c.update)
}

func (c *CrudService) list(ctx *fiber.Ctx) error {
	return ctx.SendStatus(204)
}

type updateRequest struct {
	Field    string      `json:"field"`
	NewValue interface{} `json:"newValue"`
	OldValue interface{} `json:"oldValue"`
}

func (c *CrudService) update(ctx *fiber.Ctx) error {
	// Parse body
	var data updateRequest
	if err := ctx.BodyParser(&data); err != nil {
		log.Printf("[CrudService.update] %v\n", err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	// Fetch record
	id, err := primitive.ObjectIDFromHex(ctx.Params("record"))
	if err != nil {
		return ctx.SendStatus(fiber.StatusNotFound)
	}

	user, err := database.FindOne[models.User](
		database.GetBrandDb(ctx),
		context.TODO(),
		bson.M{"_id": id},
	)
	if err != nil {
		return ctx.SendStatus(fiber.StatusNotFound)
	}

	// Update field
	field := getFieldBsonName(*user, data.Field)
	_, err = database.NewQuery(ctx).
		Set(field, data.NewValue).
		UpdateOne(context.TODO(), user)
	if err != nil {
		log.Printf("[CrudService.update] %v\n", err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	// Success
	return ctx.SendStatus(204)
}

func getFieldBsonName(record interface{}, field string) string {
	t := reflect.TypeOf(record)
	if t.Kind() != reflect.Struct {
		return field
	}

	f, exists := t.FieldByName(field)
	if !exists {
		return field
	}

	tag, ok := f.Tag.Lookup("bson")
	if !ok {
		return field
	}

	parts := strings.Split(tag, ",")
	if len(parts) < 1 || len(parts[0]) == 0 {
		return field
	}

	return parts[0]
}
