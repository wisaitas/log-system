package httpx

import "encoding/json"

func CheckStatusCode2xx(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}

func ReadJSONMapLimited(b []byte, limit int) map[string]any {
	if len(b) > limit {
		b = b[:limit]
	}
	return TryParseJSON(b)
}

func TryParseJSON(b []byte) map[string]any {
	if len(b) == 0 {
		return nil
	}
	var m map[string]any
	if json.Valid(b) && json.Unmarshal(b, &m) == nil {
		return m
	}
	return nil
}
