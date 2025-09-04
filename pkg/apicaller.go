package pkg

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type ErrorResponse struct {
	Code  string `json:"code"`
	Error string `json:"error"`
}

func DownStreamHttp[T any](c *fiber.Ctx, method string, url string, req any, resp *T) error {
	client := &http.Client{}
	reqJson, err := json.Marshal(req)
	if err != nil {
		return err
	}
	body := bytes.NewReader(reqJson)

	reqHttp, err := http.NewRequestWithContext(c.UserContext(), method, url, body)
	if err != nil {
		return err
	}

	for key, values := range c.GetReqHeaders() {
		reqHttp.Header.Add(HeaderInternal, "true")
		for _, value := range values {
			reqHttp.Header.Add(key, value)
		}
	}

	respHttp, err := client.Do(reqHttp)
	if err != nil {
		return err
	}
	defer respHttp.Body.Close()

	if checkStatus2xx(respHttp.StatusCode) {
		errorResponse := ErrorResponse{}
		if err = json.NewDecoder(respHttp.Body).Decode(&errorResponse); err != nil {
			return err
		}
		return errors.New(errorResponse.Error)
	}

	for key, values := range respHttp.Header {
		for _, value := range values {
			if key != HeaderTraceID {
				c.Response().Header.Add(key, value)
			}
		}
	}

	if err = json.NewDecoder(respHttp.Body).Decode(resp); err != nil {
		return err
	}

	return nil
}

func checkStatus2xx(statusCode int) bool {
	if statusCode >= 200 && statusCode < 300 {
		return true
	}
	return false
}
