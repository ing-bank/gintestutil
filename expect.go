package gintestutil

import (
	"sync"

	"github.com/gin-gonic/gin"
)

// ExpectOption allows various options to be supplied to Expect* functions
type ExpectOption func(*calledConfig)

// TimesCalled is used to expect an invocation an X amount of times
func TimesCalled(times int) ExpectOption {
	return func(config *calledConfig) {
		config.Times = times
	}
}

// Expectation is used to have a global wait group to wait for
// when asserting multiple calls made
func Expectation(expectation *sync.WaitGroup) ExpectOption {
	return func(config *calledConfig) {
		config.Expectation = expectation
	}
}

type calledConfig struct {
	Times       int
	Expectation *sync.WaitGroup
}

// ExpectCalled can be used on a gin endpoint to express an expectation that the endpoint will
// be called some time in the future. In combination with a test
// can wait for this expectation to be true or fail after some predetermined amount of time
func ExpectCalled(t TestingT, ctx *gin.Engine, path string, options ...ExpectOption) *sync.WaitGroup {
	t.Helper()

	if ctx == nil {
		t.Errorf("context cannot be nil")

		return nil
	}

	config := &calledConfig{
		Times:       1,
		Expectation: &sync.WaitGroup{},
	}

	for _, option := range options {
		option(config)
	}

	// Set waitgroup for amount of times
	config.Expectation.Add(config.Times)

	// Add middleware for provided route
	var timesCalled int
	ctx.Use(func(c *gin.Context) {
		c.Next()
		if c.FullPath() != path {
			return
		}

		timesCalled++
		if timesCalled <= config.Times {
			config.Expectation.Done()

			return
		}

		t.Errorf("%s hook asserts called %d times but called at least %d times\n", path, config.Times, timesCalled)
	})

	return config.Expectation
}
