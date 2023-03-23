package gintestutil

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// setup prepares the tests
func setup(endpoint string, varargs ...ExpectOption) (*testing.T, chan struct{}, *gin.Engine, *sync.WaitGroup, *httptest.Server) {
	testObject := new(testing.T)

	// create gin context
	ginContext := gin.Default()

	// create expectation
	expectation := ExpectCalled(testObject, ginContext, endpoint, varargs...)

	// create endpoints on ginContext
	ginContext.GET(endpoint, func(context *gin.Context) {
		context.Status(http.StatusOK)
	})

	// channel for go-routine to signal completion
	c := make(chan struct{})

	// create webserver
	ts := httptest.NewServer(ginContext)

	return testObject, c, ginContext, expectation, ts
}

func TestExpectCalled_SingleCallWithDefaultArgumentsReturnsSuccess(t *testing.T) {
	t.Parallel()

	// arrange
	path := "/hello-world"
	testObject, c, _, expectation, ts := setup(path)

	// Act
	_, err := http.Get(fmt.Sprintf("%s%s", ts.URL, path))

	// Assert
	assert.Nil(t, err)

	go func() {
		defer close(c)
		expectation.Wait()
	}()

	select {
	case <-c:
		assert.False(t, testObject.Failed())
	case <-time.After(15 * time.Second):
		t.FailNow()
	}
}

func TestExpectCalled_ZeroCallsWithDefaultArgumentsTimesOutAndFails(t *testing.T) {
	t.Parallel()

	// arrange
	path := "/hello-world"
	_, c, _, expectation, ts := setup(path)

	// Act
	// Make a call to and endpoint which is _NOT_ path such that path is never called
	_, err := http.Get(fmt.Sprintf("%s%s", ts.URL, "/something-other-than-path"))

	// Assert
	assert.Nil(t, err)

	go func() {
		defer close(c)
		expectation.Wait()
	}()

	select {
	case <-c:
		t.FailNow()
	case <-time.After(15 * time.Second):
		// test is bounded to accept after 15 seconds
	}
}

func TestExpectCalled_NilGinContextReturnsError(t *testing.T) {
	t.Parallel()

	// arrange
	testObject := new(testing.T)
	var ginContext *gin.Engine

	// Act
	expectation := ExpectCalled(testObject, ginContext, "")

	// Assert
	assert.True(t, testObject.Failed())
	assert.Nil(t, expectation)
}

func TestExpectCalled_CalledToOftenReturnsError(t *testing.T) {
	t.Parallel()

	// arrange
	path := "/hello-world"
	testObject, c, _, expectation, ts := setup(path)

	// Act
	_, _ = http.Get(fmt.Sprintf("%s%s", ts.URL, path))
	_, _ = http.Get(fmt.Sprintf("%s%s", ts.URL, path))

	// Assert
	go func() {
		defer close(c)
		expectation.Wait()
	}()

	select {
	case <-c:
		assert.True(t, testObject.Failed())
	case <-time.After(15 * time.Second):
		t.FailNow()
	}
}

func TestExpectCalled_TwoTimesWithTwoEndpointsSucceeds(t *testing.T) {
	t.Parallel()

	// arrange
	testObject := new(testing.T)

	// create gin context
	ginContext := gin.Default()

	// endpoints
	endpointCalledTwice := "/hello-world"
	endpointCalledOnce := "/other-path"

	// create expectation
	expectation := ExpectCalled(testObject, ginContext, endpointCalledTwice, TimesCalled(2))
	expectation = ExpectCalled(testObject, ginContext, endpointCalledOnce, Expectation(expectation))

	// create endpoints on ginContext
	for _, endpoint := range []string{endpointCalledTwice, endpointCalledOnce} {
		ginContext.GET(endpoint, func(context *gin.Context) {
			context.Status(http.StatusOK)
		})
	}

	// channel for go-routine to signal completion
	c := make(chan struct{})

	// create webserver
	ts := httptest.NewServer(ginContext)

	// act
	_, _ = http.Get(fmt.Sprintf("%s%s", ts.URL, endpointCalledTwice))
	_, _ = http.Get(fmt.Sprintf("%s%s", ts.URL, endpointCalledTwice))
	_, _ = http.Get(fmt.Sprintf("%s%s", ts.URL, endpointCalledOnce))

	// Assert
	go func() {
		defer close(c)
		expectation.Wait()
	}()

	select {
	case <-c:
		assert.False(t, testObject.Failed())
	case <-time.After(15 * time.Second):
		t.FailNow()
	}
}
