package httpx

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func NewLogger(serviceName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		traceID := c.Get(HeaderTraceID)
		if traceID == "" {
			tid, _ := uuid.NewV7()
			traceID = tid.String()
		}
		c.Request().Header.Set(HeaderTraceID, traceID)
		c.Set(HeaderTraceID, traceID)
		switch c.Get("Content-Type") {
		case "application/json":
			return HandleJSON(c, serviceName)
		default:
			return c.Next()
		}
	}
}

type Log struct {
	TraceID    string `json:"trace_id"`
	Timestamp  string `json:"timestamp"`
	DurationMs string `json:"duration_ms"`

	Current *Block `json:"current"`
	Source  *Block `json:"source,omitempty"`
}

type Block struct {
	Service      string  `json:"service"`
	Method       string  `json:"method"`
	ErrorMessage *string `json:"error_message,omitempty"`
	Path         string  `json:"path"`
	StatusCode   string  `json:"status_code"`
	Code         string  `json:"code"`
	File         *string `json:"file,omitempty"`
	Request      *Body   `json:"request"`
	Response     *Body   `json:"response"`
}

type Body struct {
	Headers map[string]string `json:"headers"`
	Body    map[string]any    `json:"body,omitempty"`
}

func HandleJSON(c *fiber.Ctx, serviceName string) error {
	start := time.Now()
	payload := ReadJSONMapLimited(c.Body(), 64<<10)
	requestHeaders := make(map[string]string)
	c.Request().Header.VisitAll(func(key, value []byte) {
		if string(key) != HeaderTraceID {
			requestHeaders[string(key)] = string(value)
		}
	})

	if err := c.Next(); err != nil {
		return err
	}

	responseBody := c.Response().Body()
	responsePayload := ReadJSONMapLimited(responseBody, 64<<10)
	responseHeaders := make(map[string]string)
	c.Response().Header.VisitAll(func(key, value []byte) {
		if string(key) != HeaderTraceID && string(key) != HeaderSource {
			responseHeaders[string(key)] = string(value)
		}
	})

	errorContext := &ErrorContext{}
	if !CheckStatusCode2xx(c.Response().StatusCode()) {
		errorContextLocal, ok := c.Locals("errorContext").(ErrorContext)
		if !ok {
			log.Printf("[middleware] : errorContext not found")
		}
		errorContext = &errorContextLocal
	}

	current := &Block{
		Service:      serviceName,
		Method:       c.Method(),
		Path:         c.Hostname() + c.Path(),
		StatusCode:   strconv.Itoa(c.Response().StatusCode()),
		Request:      &Body{Headers: requestHeaders, Body: payload},
		Response:     &Body{Headers: responseHeaders, Body: responsePayload},
		ErrorMessage: &errorContext.ErrorMessage,
		File:         errorContext.FilePath,
	}

	logInfo := Log{
		TraceID:    c.Get(HeaderTraceID),
		Timestamp:  start.Format(time.RFC3339),
		DurationMs: strconv.Itoa(int(time.Since(start).Milliseconds())),
		Current:    current,
	}

	if string(c.Response().Header.Peek(HeaderSource)) != "" {
		source := new(Block)
		if err := json.Unmarshal(c.Response().Header.Peek(HeaderSource), source); err != nil {
			log.Printf("[middleware] : %s", err.Error())
		}

		logInfo.Source = source
	} else if string(c.Response().Header.Peek(HeaderSource)) == "" {
		source := &Block{
			Service:      serviceName,
			Method:       c.Method(),
			Path:         c.Hostname() + c.Path(),
			StatusCode:   strconv.Itoa(c.Response().StatusCode()),
			File:         errorContext.FilePath,
			ErrorMessage: &errorContext.ErrorMessage,
			Request:      &Body{Headers: requestHeaders, Body: payload},
			Response:     &Body{Headers: responseHeaders, Body: responsePayload},
		}

		jsonResp, err := json.Marshal(source)
		if err != nil {
			log.Printf("[middleware] : %s", err.Error())
		}
		c.Response().Header.Set(HeaderSource, string(jsonResp))
	}

	if c.Get(HeaderInternal) != "true" {
		c.Response().Header.Del(HeaderSource)
	}

	jsonResp, err := json.Marshal(logInfo)
	if err != nil {
		log.Printf("[middleware] : %s", err.Error())
	}

	fmt.Println(string(jsonResp))
	return err
}
