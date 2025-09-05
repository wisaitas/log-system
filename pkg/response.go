package pkg

import (
	"log"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Pre-allocated error code maps for better performance
var (
	errorCodes = map[int]string{
		304: "E30400",
		400: "E40000",
		401: "E40002",
		403: "E40003",
		404: "E40004",
		500: "E50000",
	}

	successCodes = map[int]string{
		200: "E20000",
		201: "E20001",
		204: "E20004",
	}

	// String builder pool for file path construction
	filePathPool = sync.Pool{
		New: func() interface{} {
			return &strings.Builder{}
		},
	}
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

func NewErrorResponse[T any](c *fiber.Ctx, statusCode int, err error) error {
	if err == nil {
		return nil
	}

	// Use pre-allocated error codes
	code, exists := errorCodes[statusCode]
	if !exists {
		code = "E50000"
	}

	// Optimize file path generation
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		log.Println("[response] : runtime.Caller failed")
	}

	// Use string builder pool for better performance
	builder := filePathPool.Get().(*strings.Builder)
	builder.Reset()
	defer filePathPool.Put(builder)

	builder.WriteString(file)
	builder.WriteByte(':')
	builder.WriteString(strconv.Itoa(line))

	filePath := builder.String()
	c.Locals("filePath", filePath)

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
		msg = Ptr(publicMessage[0])
	}

	// Use pre-allocated success codes
	code, exists := successCodes[statusCode]
	if !exists {
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
