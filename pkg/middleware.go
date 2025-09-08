package pkg

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	HeaderTraceID      = "Trace-Id"
	HeaderErrSignature = "X-Error-Signature"
	HeaderInternal     = "X-Internal-Call"
	HeaderSource       = "X-Source"
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

	Current *LogBlock `json:"current"`
	Source  *LogBlock `json:"source,omitempty"`
}

type LogBlock struct {
	Service      string   `json:"service"`
	Method       string   `json:"method"`
	ErrorMessage *string  `json:"error_message,omitempty"`
	Path         string   `json:"path"`
	StatusCode   string   `json:"status_code"`
	Code         string   `json:"code"`
	File         *string  `json:"file,omitempty"`
	Request      *BodyLog `json:"request"`
	Response     *BodyLog `json:"response"`
}

type BodyLog struct {
	Headers map[string]string `json:"headers"`
	Body    map[string]any    `json:"body,omitempty"`
}

func HandleJSON(c *fiber.Ctx, serviceName string) error {
	start := time.Now()
	payload := readJSONMapLimited(c.Body(), 64<<10)
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
	responsePayload := readJSONMapLimited(responseBody, 64<<10)
	responseHeaders := make(map[string]string)
	c.Response().Header.VisitAll(func(key, value []byte) {
		if string(key) != HeaderTraceID && string(key) != HeaderSource {
			responseHeaders[string(key)] = string(value)
		}
	})

	errorContext := &ErrorContext{}
	if !checkStatusCode2xx(c.Response().StatusCode()) {
		errorContextLocal, ok := c.Locals("errorContext").(ErrorContext)
		if !ok {
			log.Printf("[middleware] : errorContext not found")
		}
		errorContext = &errorContextLocal
	}

	current := &LogBlock{
		Service:      serviceName,
		Method:       c.Method(),
		Path:         c.Hostname() + c.Path(),
		StatusCode:   strconv.Itoa(c.Response().StatusCode()),
		Request:      &BodyLog{Headers: requestHeaders, Body: payload},
		Response:     &BodyLog{Headers: responseHeaders, Body: responsePayload},
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
		source := new(LogBlock)
		if err := json.Unmarshal(c.Response().Header.Peek(HeaderSource), source); err != nil {
			log.Printf("[middleware] : %s", err.Error())
		}

		logInfo.Source = source
	} else if string(c.Response().Header.Peek(HeaderSource)) == "" {
		source := &LogBlock{
			Service:      serviceName,
			Method:       c.Method(),
			Path:         c.Hostname() + c.Path(),
			StatusCode:   strconv.Itoa(c.Response().StatusCode()),
			File:         errorContext.FilePath,
			ErrorMessage: &errorContext.ErrorMessage,
			Request:      &BodyLog{Headers: requestHeaders, Body: payload},
			Response:     &BodyLog{Headers: responseHeaders, Body: responsePayload},
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

func checkStatusCode2xx(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}

func readJSONMapLimited(b []byte, limit int) map[string]any {
	if len(b) > limit {
		b = b[:limit]
	}
	return tryParseJSON(b)
}

func tryParseJSON(b []byte) map[string]any {
	if len(b) == 0 {
		return nil
	}
	var m map[string]any
	if json.Valid(b) && json.Unmarshal(b, &m) == nil {
		return m
	}
	return nil
}
