package main

import (
	"log-system/pkg/httpx"
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

	app.Use(httpx.NewLogger("server"))

	app.Post("/do/:id", func(c *fiber.Ctx) error {
		var req Request
		if err := c.BodyParser(&req); err != nil {
			return httpx.NewErrorResponse[any](c, fiber.StatusBadRequest, err)
		}

		paramID := c.Params("id")

		var resp httpx.StandardResponse[*ProcessorResponse]
		if err := httpx.Client(c, http.MethodPost, "http://localhost:8082/do/"+paramID, ProcessorRequest(req), &resp); err != nil {
			return httpx.NewErrorResponse[any](c, fiber.StatusInternalServerError, err)
		}

		return c.Status(resp.StatusCode).JSON(resp)
	})

	app.Listen(":8081")
}
