package httputil

import (
	"fmt"
	"net/http"
)

// HTTPClient interface
type HTTPClient interface {
	Get(url string) (resp *http.Response, err error)
}

var (
	client HTTPClient
)

func init() {
	client = &http.Client{CheckRedirect: CheckRedirectFunc}
}

// Get sends a GET request to the URL
func Get(url string) (*http.Response, error) {
	return client.Get(url)
}

// CheckRedirectFunc sets the maximum number of redirects that will be followed to 5.
func CheckRedirectFunc(req *http.Request, via []*http.Request) error {
	if len(via) > 5 {
		return fmt.Errorf("too many redirects (5 max)")
	}
	return nil
}

// SetClient sets the value of the client.
func SetClient(httpClient HTTPClient) {
	client = httpClient
}
