package gintestutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
)

// RequestOption are functions used in PrepareRequest to configure a request using the Functional Option pattern.
type RequestOption func(*requestConfig)

// requestConfig is the internal config of PrepareRequest
type requestConfig struct {
	method      string
	url         string
	body        io.ReadCloser
	urlParams   map[string]any
	queryParams map[string]any
}

// applyQueryParams turn a map of string/[]string/maps into query parameter names as expected from the user. Check
// out the unit-tests for a more in-depth explanation.
func applyQueryParams(params map[string]any, query url.Values, keyPrefix string) {
	for key, value := range params {
		newKey := key
		if keyPrefix != "" {
			newKey = fmt.Sprintf("%s[%s]", keyPrefix, key)
		}

		switch resultValue := value.(type) {
		case string:
			query.Add(newKey, resultValue)
		case []string:
			for _, valueString := range resultValue {
				query.Add(newKey, valueString)
			}
		case fmt.Stringer:
			query.Add(newKey, resultValue.String())

		case map[string]any:
			applyQueryParams(resultValue, query, newKey)
		}
	}
}

// PrepareRequest Formulate a request with optional properties. This returns a *gin.Context which can be used
// in controller unit-tests. Use the returned *httptest.ResponseRecorder to perform assertions on the response.
func PrepareRequest(t TestingT, options ...RequestOption) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	config := &requestConfig{
		method: http.MethodGet,
		url:    "https://example.com",
	}

	for _, option := range options {
		option(config)
	}

	writer := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(writer)

	var err error
	if context.Request, err = http.NewRequest(config.method, config.url, config.body); err != nil {
		t.Error(err)
		return context, writer
	}

	query := context.Request.URL.Query()
	applyQueryParams(config.queryParams, query, "")
	context.Request.URL.RawQuery = query.Encode()

	for key, value := range config.urlParams {
		switch resultValue := value.(type) {
		case string:
			context.Params = append(context.Params, gin.Param{Key: key, Value: resultValue})

		case []string:
			for _, valueString := range resultValue {
				context.Params = append(context.Params, gin.Param{Key: key, Value: valueString})
			}

		case fmt.Stringer:
			context.Params = append(context.Params, gin.Param{Key: key, Value: resultValue.String()})
		}
	}

	return context, writer
}

// WithMethod specifies the method to use, defaults to Get
func WithMethod(method string) RequestOption {
	return func(config *requestConfig) {
		config.method = method
	}
}

// WithUrl specifies the url to use, defaults to https://example.com
func WithUrl(url string) RequestOption {
	return func(config *requestConfig) {
		config.url = url
	}
}

// WithJsonBody specifies the request body using json.Marshal, will report an error on marshal failure
func WithJsonBody(t TestingT, object any) RequestOption {
	data, err := json.Marshal(object)
	if err != nil {
		t.Error(err)
	}

	return func(config *requestConfig) {
		config.body = io.NopCloser(bytes.NewBuffer(data))
	}
}

// WithBody allows you to define a custom body for the request
func WithBody(data []byte) RequestOption {
	return func(config *requestConfig) {
		config.body = io.NopCloser(bytes.NewBuffer(data))
	}
}

// WithUrlParams adds url parameters to the request. The value can be either:
// - string
// - []string
// - fmt.Stringer (anything with a String() method)
func WithUrlParams(parameters map[string]any) RequestOption {
	return func(config *requestConfig) {
		config.urlParams = parameters
	}
}

// WithQueryParams adds query parameters to the request. The value can be either:
// - string
// - []string
// - fmt.Stringer (anything with a String() method)
// - map[string]any
func WithQueryParams(queryParams map[string]any) RequestOption {
	return func(config *requestConfig) {
		config.queryParams = queryParams
	}
}
