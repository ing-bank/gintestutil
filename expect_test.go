package gintestutil

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

const expectTimeout = 5 * time.Second

func TestExpectCalled_SingleCallReturnsSuccess(t *testing.T) {
	t.Parallel()
	// Arrange
	testObject := new(testing.T)
	ginContext := gin.Default()

	// Act
	expectation := ExpectCalled(testObject, ginContext, "/ping")

	// Assert
	ginContext.GET("/ping", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})

	ts := httptest.NewServer(ginContext)

	_, err := http.Get(fmt.Sprintf("%s%s", ts.URL, "/ping"))

	assert.NoError(t, err)

	completionChannel := make(chan struct{})

	go func() {
		expectation.Wait()
		completionChannel <- struct{}{}
	}()

	select {
	case <-completionChannel:
		assert.False(t, testObject.Failed())
	case <-time.After(expectTimeout):
		t.Error("did not complete")
	}
}

func TestExpectCalled_ZeroCallsFails(t *testing.T) {
	t.Parallel()
	// Arrange
	testObject := new(testing.T)
	ginContext := gin.Default()

	// Act
	expectation := ExpectCalled(testObject, ginContext, "/ping")

	// Assert
	ginContext.GET("/ping", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})

	ts := httptest.NewServer(ginContext)

	_, err := http.Get(fmt.Sprintf("%s%s", ts.URL, "/pong"))

	// Assert
	assert.NoError(t, err)

	completionChannel := make(chan struct{})

	go func() {
		expectation.Wait()
		completionChannel <- struct{}{}
	}()

	select {
	case <-completionChannel:
		t.Error("should not have completed")

	case <-time.After(expectTimeout):
		// Success!
	}
}

func TestExpectCalled_NilGinContextReturnsError(t *testing.T) {
	t.Parallel()
	// Arrange
	testObject := new(testing.T)
	var ginContext *gin.Engine

	// Act
	expectation := ExpectCalled(testObject, ginContext, "")

	// Assert
	assert.True(t, testObject.Failed())
	assert.Nil(t, expectation)
}

func TestExpectCalled_CalledTooOftenReturnsError(t *testing.T) {
	t.Parallel()
	// Arrange
	testObject := new(testing.T)

	ginContext := gin.Default()

	// Act
	expectation := ExpectCalled(testObject, ginContext, "/ping")

	// Assert
	ginContext.GET("/ping", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})

	ts := httptest.NewServer(ginContext)

	_, _ = http.Get(fmt.Sprintf("%s%s", ts.URL, "/ping"))
	_, _ = http.Get(fmt.Sprintf("%s%s", ts.URL, "/ping"))

	completionChannel := make(chan struct{})

	go func() {
		expectation.Wait()
		completionChannel <- struct{}{}
	}()

	select {
	case <-completionChannel:
		assert.True(t, testObject.Failed())
	case <-time.After(expectTimeout):
		t.Error("did not complete")
	}
}

func TestExpectCalled_TwoTimesWithTwoEndpointsSucceeds(t *testing.T) {
	t.Parallel()
	// Arrange
	testObject := new(testing.T)

	ginContext := gin.Default()

	// Act
	expectation := ExpectCalled(testObject, ginContext, "/ping", TimesCalled(2))
	expectation = ExpectCalled(testObject, ginContext, "/pong", Expectation(expectation))

	// Assert
	for _, endpoint := range []string{"/ping", "/pong"} {
		ginContext.GET(endpoint, func(context *gin.Context) {
			context.Status(http.StatusOK)
		})
	}

	ts := httptest.NewServer(ginContext)

	_, _ = http.Get(fmt.Sprintf("%s%s", ts.URL, "/ping"))
	_, _ = http.Get(fmt.Sprintf("%s%s", ts.URL, "/ping"))
	_, _ = http.Get(fmt.Sprintf("%s%s", ts.URL, "/pong"))

	completionChannel := make(chan struct{})

	go func() {
		expectation.Wait()
		completionChannel <- struct{}{}
	}()

	select {
	case <-completionChannel:
		assert.False(t, testObject.Failed())
	case <-time.After(expectTimeout):
		t.Error("did not complete")
	}
}
