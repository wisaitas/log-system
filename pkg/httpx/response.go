package httpx

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
)

type StandardResponse[T any] struct {
	Timestamp     string      `json:"timestamp"`
	StatusCode    int         `json:"status_code"`
	Code          string      `json:"code"`
	Data          *T          `json:"data"`
	Pagination    *Pagination `json:"pagination"`
	PublicMessage *string     `json:"public_message"`
}

type Pagination struct {
	Page          int  `json:"page"`
	Limit         int  `json:"limit"`
	TotalElements int  `json:"total_elements"`
	HasNext       bool `json:"has_next"`
	HasPrevious   bool `json:"has_previous"`
	IsLastPage    bool `json:"is_last_page"`
}

type ErrorContext struct {
	FilePath     *string `json:"file_path"`
	ErrorMessage string  `json:"error_message"`
}

func NewErrorResponse[T any](c *fiber.Ctx, statusCode int, err error) error {
	if err == nil {
		return nil
	}

	var code string

	switch statusCode {
	case 304:
		code = "E30400"
	case 400:
		code = "E40000"
	case 401:
		code = "E40002"
	case 403:
		code = "E40003"
	case 404:
		code = "E40004"
	case 500:
		code = "E50000"
	default:
		code = "E50000"
	}

	_, file, line, ok := runtime.Caller(1)
	if !ok {
		log.Println("[response] : runtime.Caller failed")
	}

	filePath := fmt.Sprintf("%s:%d", file, line)

	c.Locals("errorContext", ErrorContext{
		FilePath:     &filePath,
		ErrorMessage: err.Error(),
	})

	return c.Status(statusCode).JSON(&StandardResponse[T]{
		Timestamp:  time.Now().Format(time.RFC3339),
		StatusCode: statusCode,
		Data:       new(T),
		Code:       code,
		Pagination: nil,
	})
}

func NewSuccessResponse[T any](data *T, statusCode int, pagination *Pagination, publicMessage ...string) StandardResponse[T] {
	var msg *string
	if len(publicMessage) > 0 {
		msg = &publicMessage[0]
	}

	var code string

	switch statusCode {
	case 200:
		code = "E20000"
	case 201:
		code = "E20001"
	case 204:
		code = "E20004"
	default:
		code = "E20000"
	}

	return StandardResponse[T]{
		Timestamp:     time.Now().Format(time.RFC3339),
		StatusCode:    statusCode,
		Data:          data,
		Code:          code,
		Pagination:    pagination,
		PublicMessage: msg,
	}
}
