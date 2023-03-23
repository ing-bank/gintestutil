# 🦁 Gin Test Utils

[![Go package](https://github.com/ing-bank/gintestutil/actions/workflows/test.yaml/badge.svg)](https://github.com/ing-bank/gintestutil/actions/workflows/test.yaml)
![GitHub](https://img.shields.io/github/license/ing-bank/gintestutil)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/ing-bank/gintestutil)

Small utility functions for testing Gin-related code.
Such as the creation of a gin context and wait groups with callbacks.

## ⬇️ Installation

`go get github.com/ing-bank/gintestutil`

## 📋 Usage

### Context creation

```go
package main

import (
	"net/http"
	"testing"
	"github.com/ing-bank/gintestutil"
)

type TestObject struct {
	Name string
}

func TestProductController_Post_CreatesProducts(t *testing.T) {
	// Arrange
	context, writer := gintestutil.PrepareRequest(t,
		gintestutil.WithJsonBody(t, TestObject{Name: "test"}),
		gintestutil.WithMethod(http.MethodPost),
		gintestutil.WithUrl("https://my-website.com"),
		gintestutil.WithUrlParams(map[string]any{"category": "barbecue"}),
	    gintestutil.WithQueryParams(map[string]any{"force": "true"}))

	// [...]
}
```

### Response Assertions

```go
package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"github.com/ing-bank/gintestutil"
)

type TestObject struct {
	Name string
}

func TestProductController_Index_ReturnsAllProducts(t *testing.T) {
	// Arrange
	context, writer := gintestutil.PrepareRequest(t)

	// [...]

	// Assert
	var actual []TestObject
	if gintestutil.Response(t, &actual, http.StatusOK, writer.Result()) {
		assert.Equal(t, []TestObject{}, actual)
	}
}
```

### Hooks

```go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/ing-bank/gintestutil"
	"net/http"
	"net/http/httptest"
	"time"
	"testing"
)

func TestHelloController(t *testing.T) {
	// Arrange
	ginContext := gin.Default()

	// create expectation
	expectation := gintestutil.ExpectCalled(t, ginContext, "/hello-world")

	ginContext.GET("/hello-world", func(context *gin.Context) {
		context.Status(http.StatusOK)
	})

	// create webserver
	ts := httptest.NewServer(ginContext)

	// Send request to webserver path
	_, _ = http.Get(fmt.Sprintf("%s/hello-world", ts.URL))

	// Wait for expectation in bounded time
	if ok := gintestutil.EnsureCompletion(t, expectation); !ok {
		// do something
	}
}
```

## 🚀 Development

1. Clone the repository
2. Run `make t` to run unit tests
3. Run `make fmt` to format code

You can run `make` to see a list of useful commands.

## 🔭 Future Plans

Nothing here yet!
