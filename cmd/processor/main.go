package main

import (
	"errors"
	"log-system/pkg"

	"github.com/gofiber/fiber/v2"
)

type Request struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type Response struct {
	FullName string `json:"full_name"`
}

func main() {
	app := fiber.New()

	app.Use(pkg.NewLogger("processor"))

	app.Post("/do/:id", func(c *fiber.Ctx) error {
		request := Request{}
		if err := c.BodyParser(&request); err != nil {
			return pkg.NewErrorResponse[any](c, fiber.StatusBadRequest, err)
		}

		param := c.Params("id")
		if param == "b" {
			return pkg.NewErrorResponse[any](c, fiber.StatusBadRequest, errors.New("b is not allowed"))
		}

		return c.Status(fiber.StatusOK).JSON(
			pkg.NewSuccessResponse(&Response{
				FullName: request.FirstName + " " + request.LastName,
			}, fiber.StatusOK, nil),
		)
	})

	app.Listen(":8082")
}
