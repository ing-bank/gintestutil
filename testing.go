package gintestutil

import (
	"fmt"
	"testing"
)

// Compile-time interface checks
var _ TestingT = new(testing.T)
var _ TestingT = new(mockT)

// TestingT is an interface representing testing.T in our tests, allows for verifying Errorf calls. It's perfectly
// compatible with the normal testing.T, but we use an interface for mocking.
type TestingT interface {
	Helper()
	Error(...any)
	Errorf(string, ...any)
}

// mockT is the mock version of the TestingT interface, used to verify Errorf calls
type mockT struct {
	ErrorCalls  []any
	ErrorfCalls []string
}

// Helper does nothing
func (m *mockT) Helper() {}

// Errorf saves Errorf calls in an error for verification
func (m *mockT) Errorf(format string, args ...any) {
	m.ErrorfCalls = append(m.ErrorfCalls, fmt.Sprintf(format, args...))
}

// Error saves Error calls in an error for verification
func (m *mockT) Error(args ...any) {
	m.ErrorCalls = append(m.ErrorCalls, args...)
}
