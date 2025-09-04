package pkg

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func DownStreamHttp[T any](c *fiber.Ctx, method string, url string, req any, resp T) error {
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
		for _, value := range values {
			reqHttp.Header.Add(key, value)
		}
	}

	respHttp, err := client.Do(reqHttp)
	if err != nil {
		return err
	}
	defer respHttp.Body.Close()

	if err = json.NewDecoder(respHttp.Body).Decode(resp); err != nil {
		return err
	}

	return nil
}
