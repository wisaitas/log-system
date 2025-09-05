package pkg

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	HeaderTraceID      = "Trace-Id"
	HeaderErrSignature = "X-Error-Signature"
	HeaderInternal     = "X-Internal-Call"
	HeaderSource       = "X-Source"
)

// Pre-allocated buffers for better performance
var (
	headerBufferPool = sync.Pool{
		New: func() interface{} {
			return make(map[string]string, 16)
		},
	}
)

func NewLogger(serviceName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		traceID := c.Get(HeaderTraceID)
		if traceID == "" {
			tid, _ := uuid.NewV7()
			traceID = tid.String()
		}
		c.Set(HeaderTraceID, traceID)

		// Optimize content type check
		contentType := c.Get("Content-Type")
		if strings.HasPrefix(contentType, "application/json") {
			return HandleJSON(c, serviceName)
		}
		return c.Next()
	}
}

type Log struct {
	Timestamp  string `json:"timestamp"`
	DurationMs string `json:"duration_ms"`

	Current *LogBlock `json:"current"`
	Source  *LogBlock `json:"source,omitempty"`
}

type LogBlock struct {
	Service    string   `json:"service"`
	Method     string   `json:"method"`
	Path       string   `json:"path"`
	StatusCode string   `json:"status_code"`
	Code       string   `json:"code"`
	Request    *BodyLog `json:"request"`
	Response   *BodyLog `json:"response"`
	File       *string  `json:"file"`
}

type BodyLog struct {
	Headers map[string]string `json:"headers"`
	Body    map[string]any    `json:"body,omitempty"`
}

func HandleJSON(c *fiber.Ctx, serviceName string) error {
	start := time.Now()

	// Optimize JSON parsing with buffer reuse
	payload := readJSONMapOptimized(c.Body())

	// Reuse header map from pool
	requestHeaders := headerBufferPool.Get().(map[string]string)
	// Clear the map for reuse
	for k := range requestHeaders {
		delete(requestHeaders, k)
	}
	defer headerBufferPool.Put(requestHeaders)

	// Optimize header processing
	c.Request().Header.VisitAll(func(key, value []byte) {
		requestHeaders[unsafe.String(unsafe.SliceData(key), len(key))] = unsafe.String(unsafe.SliceData(value), len(value))
	})

	if err := c.Next(); err != nil {
		return err
	}

	responseBody := c.Response().Body()
	responsePayload := readJSONMapOptimized(responseBody)

	// Reuse response header map
	responseHeaders := headerBufferPool.Get().(map[string]string)
	for k := range responseHeaders {
		delete(responseHeaders, k)
	}
	defer headerBufferPool.Put(responseHeaders)

	// Optimize response header processing
	c.Response().Header.VisitAll(func(key, value []byte) {
		keyStr := unsafe.String(unsafe.SliceData(key), len(key))
		if keyStr != HeaderTraceID && keyStr != HeaderSource {
			responseHeaders[keyStr] = unsafe.String(unsafe.SliceData(value), len(value))
		}
	})

	var filePath *string
	statusCode := c.Response().StatusCode()
	if !checkStatusCode2xx(statusCode) {
		if filePathFromLocals, ok := c.Locals("filePath").(string); ok {
			filePath = Ptr(filePathFromLocals)
		}
	}

	// Pre-allocate strings to avoid repeated allocations
	method := c.Method()
	path := c.Hostname() + c.Path()
	statusCodeStr := strconv.Itoa(statusCode)

	current := &LogBlock{
		Code:       errorCodes[statusCode],
		Service:    serviceName,
		Method:     method,
		Path:       path,
		StatusCode: statusCodeStr,
		Request:    &BodyLog{Headers: requestHeaders, Body: payload},
		Response:   &BodyLog{Headers: responseHeaders, Body: responsePayload},
		File:       filePath,
	}

	// Optimize timestamp formatting
	timestamp := start.Format(time.RFC3339)
	durationMs := strconv.Itoa(int(time.Since(start).Milliseconds()))

	logInfo := Log{
		Timestamp:  timestamp,
		DurationMs: durationMs,
		Current:    current,
	}

	// Optimize source handling
	source := &LogBlock{}
	sourceHeader := c.Response().Header.Peek(HeaderSource)
	if len(sourceHeader) > 0 {
		if err := json.Unmarshal(sourceHeader, source); err != nil {
			log.Printf("[middleware] : %s", err.Error())
		}
	} else {
		// Reuse the same data for source
		source = &LogBlock{
			Code:       errorCodes[statusCode],
			Service:    serviceName,
			Method:     method,
			Path:       path,
			StatusCode: statusCodeStr,
			Request:    &BodyLog{Headers: requestHeaders, Body: payload},
			Response:   &BodyLog{Headers: responseHeaders, Body: responsePayload},
			File:       filePath,
		}

		// Use buffer pool for JSON marshaling
		buf := jsonBufferPool.Get().([]byte)
		buf = buf[:0] // Reset length
		defer jsonBufferPool.Put(buf)

		if jsonResp, err := json.Marshal(source); err != nil {
			log.Printf("[middleware] : %s", err.Error())
		} else {
			c.Response().Header.Set(HeaderSource, string(jsonResp))
		}
	}
	logInfo.Source = source

	if c.Get(HeaderInternal) != "true" {
		c.Response().Header.Del(HeaderSource)
	}

	// Use buffer pool for final JSON marshaling
	buf := jsonBufferPool.Get().([]byte)
	buf = buf[:0]
	defer jsonBufferPool.Put(buf)

	if jsonResp, err := json.Marshal(logInfo); err != nil {
		log.Printf("[middleware] : %s", err.Error())
	} else {
		fmt.Println(string(jsonResp))
	}

	return nil
}

func checkStatusCode2xx(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}

// Optimized JSON parsing with better error handling and memory usage
func readJSONMapOptimized(b []byte) map[string]any {
	if len(b) == 0 {
		return nil
	}

	// Limit size to prevent memory issues
	const maxSize = 64 << 10 // 64KB
	if len(b) > maxSize {
		b = b[:maxSize]
	}

	// Quick validation before parsing
	if !json.Valid(b) {
		return nil
	}

	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil
	}
	return m
}
