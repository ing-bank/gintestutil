package gintestutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testObject struct {
	Name string
}

func TestSuccessResponse_ReturnsExpectedResult(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		code         int
		response     *http.Response
		responseBody any
		expected     bool
	}{
		"no content": {
			code:     http.StatusNoContent,
			response: &http.Response{StatusCode: http.StatusNoContent},
			expected: true,
			// This should not be read
			responseBody: "{{{}{}{{[][}",
		},
		"not modified": {
			code:     http.StatusNotModified,
			response: &http.Response{StatusCode: http.StatusNotModified},
			expected: true,
			// This should not be read
			responseBody: "{{{}{}{{[][}",
		},
		"accepted": {
			code:     http.StatusAccepted,
			response: &http.Response{StatusCode: http.StatusAccepted},
			expected: true,
		},
		"bad request": {
			code:     http.StatusBadRequest,
			response: &http.Response{StatusCode: http.StatusOK},
			expected: false,
		},
		"broken json": {
			code:         http.StatusOK,
			response:     &http.Response{StatusCode: http.StatusOK},
			responseBody: "{{{}{}{{[][}",
			expected:     false,
		},
	}

	for name, testData := range tests {
		testData := testData
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Arrange
			testingObject := new(mockT)
			jsonData, _ := json.Marshal(testData.responseBody)
			testData.response.Body = io.NopCloser(bytes.NewBuffer(jsonData))

			// Act
			ok := Response(testingObject, &testObject{}, testData.code, testData.response)

			// Assert
			assert.Equal(t, testData.expected, ok)
		})
	}
}

func TestSuccessResponse_ReturnsData(t *testing.T) {
	t.Parallel()

	// Arrange
	testingObject := new(mockT)

	type testObject struct {
		Name  string  `json:"name"`
		Other float64 `json:"other"`
	}

	object := testObject{Name: "abc", Other: 23}

	jsonData, _ := json.Marshal(object)
	response := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBuffer(jsonData)),
	}

	var result testObject

	// Act
	ok := Response(testingObject, &result, http.StatusOK, response)

	// Assert
	assert.Equal(t, object, result)
	assert.True(t, ok)
}

func TestSuccessResponse_FailsOnNoBody(t *testing.T) {
	t.Parallel()
	// Arrange
	testingObject := new(mockT)
	response := &http.Response{
		//nolint:mirror // Used for a test
		Body:       io.NopCloser(bytes.NewBuffer([]byte(""))),
		StatusCode: http.StatusOK,
	}

	var result testObject

	// Act
	ok := Response(testingObject, result, http.StatusOK, response)

	// Assert
	assert.Empty(t, result)
	assert.False(t, ok)
}

func TestSuccessResponse_FailsOnUnmarshall(t *testing.T) {
	t.Parallel()
	// Arrange
	testingObject := new(mockT)
	response := &http.Response{
		StatusCode: http.StatusOK,
		//nolint:mirror // Used for a test
		Body: io.NopCloser(bytes.NewBuffer([]byte("test"))),
	}

	var result testObject

	// Act
	ok := Response(testingObject, result, http.StatusOK, response)

	// Assert
	assert.Empty(t, result)
	assert.False(t, ok)
}
