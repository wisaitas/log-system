package main

import (
	"log-system/pkg"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type Request struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type Response struct {
	FullName string `json:"full_name"`
}

type ProcessorRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type ProcessorResponse struct {
	FullName string `json:"full_name"`
}

func main() {
	app := fiber.New()

	app.Use(pkg.NewLogger("server"))

	app.Post("/do", func(c *fiber.Ctx) error {
		var req Request
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		var resp ProcessorResponse
		if err := pkg.DownStreamHttp(c, http.MethodPost, "http://localhost:8082/do/b", ProcessorRequest(req), &resp); err != nil {

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.Status(fiber.StatusOK).JSON(resp)
	})

	app.Listen(":8081")
}
