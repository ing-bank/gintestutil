package gintestutil

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type MyObject struct {
	ID int
}

type MyController struct{}

func (m *MyController) Get(e *gin.Context) {
	// [...]
}

func (m *MyController) Post(e *gin.Context) {
	// [...]
}

func ExamplePrepareRequest() {
	// Arrange
	t := new(mockT)
	myController := &MyController{}

	context, writer := PrepareRequest(t,
		WithMethod(http.MethodGet),
		WithUrl("https://ing.net"),
		WithJsonBody(t, MyObject{ID: 5}),
		WithUrlParams(map[string]any{"one": "two", "three": []string{"four", "five"}}),
		WithQueryParams(map[string]any{"id": map[string]any{"a": "b"}}),
	)

	// Act
	myController.Post(context)

	// Assert
	assert.Equal(t, "...", writer.Body.String())
}

func ExampleResponse() {
	// Arrange
	t := new(mockT)
	expected := []MyObject{{ID: 5}}
	myController := &MyController{}

	context, writer := PrepareRequest(t)
	defer writer.Result().Body.Close()

	// Act
	myController.Get(context)

	var result []MyObject

	// Assert
	ok := Response(t, &result, http.StatusOK, writer.Result())

	if assert.True(t, ok) {
		assert.Equal(t, expected[0].ID, result[0].ID)
	}
}

// without arguments expect called assumes that the endpoint is only called once and
// creates a new expectation
func ExampleExpectCalled_withoutVarargs() {
	t := new(testing.T)

	ginContext := gin.Default()

	// create expectation
	expectation := ExpectCalled(t, ginContext, "/hello-world")

	// create endpoints on ginContext
	ginContext.GET("/hello-world", func(context *gin.Context) {
		context.Status(http.StatusOK)
	})

	// create webserver
	ts := httptest.NewServer(ginContext)

	// Send request to webserver path
	_, _ = http.Get(fmt.Sprintf("%s/hello-world", ts.URL))

	// Wait for expectation in bounded time
	if ok := EnsureCompletion(t, expectation); !ok {
		// do something
	}
}

// arguments can configure the expected amount of times and endpoint is called or
// re-use an existing expectation
func ExampleExpectCalled_withVarargs() {
	t := new(testing.T)

	ginContext := gin.Default()

	// create expectation
	expectation := ExpectCalled(t, ginContext, "/hello-world", TimesCalled(2))
	expectation = ExpectCalled(t, ginContext, "/other-path", Expectation(expectation))

	// create endpoints on ginContext
	for _, endpoint := range []string{"/hello-world", "/other-path"} {
		ginContext.GET(endpoint, func(context *gin.Context) {
			context.Status(http.StatusOK)
		})
	}

	// create webserver
	ts := httptest.NewServer(ginContext)

	// Send request to webserver path
	_, _ = http.Get(fmt.Sprintf("%s/hello-world", ts.URL))
	_, _ = http.Get(fmt.Sprintf("%s/hello-world", ts.URL))
	_, _ = http.Get(fmt.Sprintf("%s/other-path", ts.URL))

	// Wait for expectation in bounded time
	if ok := EnsureCompletion(t, expectation); !ok {
		// do something
	}
}
