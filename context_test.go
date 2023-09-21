package gintestutil

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type testStringer struct {
	input string
}

func (s testStringer) String() string {
	return s.input
}

func TestPrepareRequest_CreatesExpectedContext(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		options []RequestOption

		expectedBody   string
		expectedUrl    string
		expectedMethod string
		expectedParams gin.Params
		expectedError  error
	}{
		"empty request": {
			options: []RequestOption{},

			expectedMethod: http.MethodGet,
			expectedUrl:    "https://example.com",
		},
		"method post": {
			options: []RequestOption{WithMethod(http.MethodPost)},

			expectedMethod: http.MethodPost,
			expectedUrl:    "https://example.com",
		},
		"with url params": {
			options: []RequestOption{
				WithUrlParams(
					map[string]any{
						"one":   testStringer{input: "two"},
						"three": []string{"four", "five"},
					}),
			},

			expectedMethod: http.MethodGet,
			expectedUrl:    "https://example.com",
			expectedParams: []gin.Param{
				{
					Key:   "one",
					Value: "two",
				},
				{
					Key:   "three",
					Value: "four",
				},
				{
					Key:   "three",
					Value: "five",
				},
			},
		},
		"with query params": {
			options: []RequestOption{
				WithUrl("https://maarten.dev"),
				WithQueryParams(map[string]any{
					"one":   testStringer{input: "two"},
					"three": []string{"four", "five"},
				}),
			},

			expectedMethod: http.MethodGet,
			expectedUrl:    "https://maarten.dev?one=two&three=four&three=five",
		},
		"maarten.dev url": {
			options: []RequestOption{WithUrl("https://maarten.dev")},

			expectedUrl:    "https://maarten.dev",
			expectedMethod: http.MethodGet,
		},
		"json body": {
			options: []RequestOption{WithJsonBody(t, map[string]any{"one": "two", "three": 4})},

			expectedMethod: http.MethodGet,
			expectedUrl:    "https://example.com",
			expectedBody:   `{"one": "two", "three": 4}`,
		},
		"raw body": {
			options: []RequestOption{WithBody([]byte("a b c"))},

			expectedMethod: http.MethodGet,
			expectedUrl:    "https://example.com",
			expectedBody:   "a b c",
		},
		"expected error on nonsensical request": {
			options: []RequestOption{WithUrl("://://::::///::::")},

			//nolint:goerr113 // Not relevant
			expectedError: &url.Error{Op: "parse", URL: "://://::::///::::", Err: errors.New("missing protocol scheme")},
		},
	}

	for name, testData := range tests {
		testData := testData
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Arrange
			mockT := new(mockT)

			// Act
			context, writer := PrepareRequest(mockT, testData.options...)

			// Assert
			assert.NotNil(t, writer)

			if testData.expectedError != nil {
				assert.Equal(t, testData.expectedError, mockT.ErrorCalls[0])

				return
			}

			assert.Len(t, mockT.ErrorCalls, 0)

			assert.Equal(t, testData.expectedMethod, context.Request.Method)
			assert.Equal(t, testData.expectedUrl, context.Request.URL.String())
			assert.ElementsMatch(t, testData.expectedParams, context.Params)

			if testData.expectedBody != "" {
				if body, err := io.ReadAll(context.Request.Body); err != nil {
					assert.Equal(t, []byte(testData.expectedBody), body)
				}
			}
		})
	}
}

func TestWithJsonBody_CallsErrorOnFaultyJson(t *testing.T) {
	t.Parallel()
	// Arrange
	mock := &mockT{}

	input := map[bool]string{
		true:  "A boolean map is not a thing in json",
		false: "so it won't work :-)",
	}

	// Act
	_ = WithJsonBody(mock, input)

	// Assert
	if assert.Len(t, mock.ErrorCalls, 1) {
		assert.IsType(t, mock.ErrorCalls[0], &json.UnsupportedTypeError{})
	}
}

func TestApplyQueryParams_SetsExpectedValues(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		input    map[string]any
		expected url.Values
	}{
		"empty": {
			input:    map[string]any{},
			expected: map[string][]string{},
		},
		"simple": {
			input: map[string]any{
				"a": "b",
				"c": "d",
			},
			expected: map[string][]string{
				"a": {"b"},
				"c": {"d"},
			},
		},
		"multi": {
			input: map[string]any{
				"a": []string{"a", "b"},
				"c": []string{"c", "d"},
			},
			expected: map[string][]string{
				"a": {"a", "b"},
				"c": {"c", "d"},
			},
		},
		"level 1": {
			input: map[string]any{
				"a": map[string]any{"aa": "bb"},
				"c": map[string]any{"cc": "dd"},
			},
			expected: map[string][]string{
				"a[aa]": {"bb"},
				"c[cc]": {"dd"},
			},
		},
		"level 2": {
			input: map[string]any{
				"a": map[string]any{
					"aa": map[string]any{
						"aaa": "bbb",
					},
				},
				"c": map[string]any{
					"cc": map[string]any{
						"ccc": "ddd",
					},
				},
			},
			expected: map[string][]string{
				"a[aa][aaa]": {"bbb"},
				"c[cc][ccc]": {"ddd"},
			},
		},
		"level 6m ": {
			input: map[string]any{
				"a": map[string]any{
					"aa": map[string]any{
						"aaa": map[string]any{
							"aaaa": map[string]any{
								"aaaaa": map[string]any{
									"aaaaaa": "bbbbbb",
								},
							},
						},
					},
				},
				"c": map[string]any{
					"cc": map[string]any{
						"ccc": map[string]any{
							"cccc": map[string]any{
								"ccccc": map[string]any{
									"cccccc": "dddddd",
								},
							},
						},
					},
				},
			},
			expected: map[string][]string{
				"a[aa][aaa][aaaa][aaaaa][aaaaaa]": {"bbbbbb"},
				"c[cc][ccc][cccc][ccccc][cccccc]": {"dddddd"},
			},
		},
	}

	for name, testData := range tests {
		testData := testData
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Arrange
			params := url.Values{}

			// Act
			applyQueryParams(testData.input, params, "")

			// Assert
			assert.Equal(t, testData.expected, params)
		})
	}
}
