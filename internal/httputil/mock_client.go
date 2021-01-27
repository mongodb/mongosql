package httputil

import "net/http"

// MockClient is the mock client
type MockClient struct {
	GetFunc func(url string) (*http.Response, error)
}

// Get is the mock client's 'Get' function
func (m *MockClient) Get(url string) (*http.Response, error) {
	return m.GetFunc(url)
}
