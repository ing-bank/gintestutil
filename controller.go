package gintestutil

import (
	"encoding/json"
	"io"
	"net/http"
)

// statusHasBody is used to determine whether a response is allowed to have a body
func statusHasBody(status int) bool {
	switch {
	case status >= http.StatusContinue && status <= 199:
		return false

	case status == http.StatusNoContent:
		return false

	case status == http.StatusNotModified:
		return false
	}

	return true
}

// Response checks the status code and unmarshalls it to the given type.
// If you don't care about the response, Use nil. If the return code is 204 or 304, the response body is not converted.
func Response(t TestingT, result any, code int, res *http.Response) bool {
	t.Helper()

	if code != res.StatusCode {
		t.Errorf("Status code %d is not %d", res.StatusCode, code)

		return false
	}

	response, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("failed to read body of response")

		return false
	}

	// Some status codes don't allow bodies
	if result == nil || !statusHasBody(code) {
		return true
	}

	if err := json.Unmarshal(response, &result); err != nil {
		t.Errorf("Failed to unmarshall '%s' into '%T': %v", response, result, err)

		return false
	}

	return true
}
