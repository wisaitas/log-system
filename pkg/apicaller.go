package pkg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Reuse HTTP client with optimized settings
var (
	httpClient = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// Buffer pool for JSON marshaling
	jsonBufferPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 1024)
		},
	}
)

func DownStreamHttp[T any](c *fiber.Ctx, method string, url string, req any, resp *StandardResponse[T]) error {
	// Use buffer pool for JSON marshaling
	buf := jsonBufferPool.Get().([]byte)
	buf = buf[:0]
	defer jsonBufferPool.Put(buf)

	reqJson, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("[apicaller] : %w", err)
	}

	body := bytes.NewReader(reqJson)
	ctx := c.UserContext()

	reqHttp, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return fmt.Errorf("[apicaller] : %w", err)
	}

	// Optimize header copying
	reqHttp.Header.Set(HeaderInternal, "true")
	for key, values := range c.GetReqHeaders() {
		for _, value := range values {
			reqHttp.Header.Add(key, value)
		}
	}

	respHttp, err := httpClient.Do(reqHttp)
	if err != nil {
		return fmt.Errorf("[apicaller] : %w", err)
	}
	defer respHttp.Body.Close()

	// Optimize response header copying
	for key, values := range respHttp.Header {
		if key != HeaderTraceID {
			for _, value := range values {
				c.Response().Header.Add(key, value)
			}
		}
	}

	// Use streaming JSON decoder for better memory usage
	if err = json.NewDecoder(respHttp.Body).Decode(resp); err != nil {
		return fmt.Errorf("[apicaller] : %w", err)
	}

	if respHttp.StatusCode != http.StatusOK {
		resp.Data = new(T)
	}

	// Optimize file path generation
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		return fmt.Errorf("[apicaller] : %w", errors.New("runtime.Caller failed"))
	}

	// Use string builder for better performance
	var filePathBuilder strings.Builder
	filePathBuilder.Grow(len(file) + 10) // Pre-allocate capacity
	filePathBuilder.WriteString(file)
	filePathBuilder.WriteByte(':')
	filePathBuilder.WriteString(strconv.Itoa(line))

	filePath := filePathBuilder.String()
	c.Locals("filePath", filePath)

	return nil
}
