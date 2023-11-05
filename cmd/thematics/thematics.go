package thematics

import (
	"context"
	"errors"
	"log"
	"time"

	"invisible-tir-go/cmd/user"
	"invisible-tir-go/infrastructure"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

const (
	address = "localhost:50051"
)

func ListThematics(c *fiber.Ctx, db *gorm.DB) error {

	accessToken := c.Get("x-access-token")
	if accessToken == "" {
		return missingHeaderResponse(c)
	}

	user := new(user.User)

	result := db.Where("api_key = ?", accessToken).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return missingHeaderResponse(c)
		}
	}

	if user.ApiKey != accessToken {
		return missingHeaderResponse(c)
	}

	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := infrastructure.NewTirServiceClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())

	return c.SendStatus(fiber.StatusOK)
}

func missingHeaderResponse(c *fiber.Ctx) error {
	missingHeaderResponse := user.ErrorResponse{
		Code:    fiber.StatusUnauthorized,
		Type:    fiber.ErrUnauthorized.Message,
		Message: "Missing x-access-token header variable",
	}
	return c.Status(fiber.StatusUnauthorized).JSON(missingHeaderResponse)
}
