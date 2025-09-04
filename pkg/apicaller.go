package pkg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"

	"github.com/gofiber/fiber/v2"
)

func DownStreamHttp[T any](c *fiber.Ctx, method string, url string, req any, resp *StandardResponse[T]) error {
	client := &http.Client{}
	reqJson, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("[apicaller] : %w", err)
	}
	body := bytes.NewReader(reqJson)

	reqHttp, err := http.NewRequestWithContext(c.UserContext(), method, url, body)
	if err != nil {
		return fmt.Errorf("[apicaller] : %w", err)
	}

	for key, values := range c.GetReqHeaders() {
		reqHttp.Header.Add(HeaderInternal, "true")
		for _, value := range values {
			reqHttp.Header.Add(key, value)
		}
	}

	respHttp, err := client.Do(reqHttp)
	if err != nil {
		return fmt.Errorf("[apicaller] : %w", err)
	}
	defer respHttp.Body.Close()

	for key, values := range respHttp.Header {
		for _, value := range values {
			if key != HeaderTraceID {
				c.Response().Header.Add(key, value)
			}
		}
	}

	if err = json.NewDecoder(respHttp.Body).Decode(resp); err != nil {
		return fmt.Errorf("[apicaller] : %w", err)
	}

	if respHttp.StatusCode != http.StatusOK {
		resp.Data = new(T)
	}

	_, file, line, ok := runtime.Caller(1)
	if !ok {
		return fmt.Errorf("[apicaller] : %w", errors.New("runtime.Caller failed"))
	}

	filePath := Ptr(fmt.Sprintf("%s:%d", file, line))

	fmt.Println(*resp.Data)

	c.Locals("filePath", filePath)
	return nil
}
