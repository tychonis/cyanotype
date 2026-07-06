package catalog

import (
	"net/http"
	"os"
)

type authTransport struct {
	token string
	base  http.RoundTripper
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.base == nil {
		t.base = http.DefaultTransport
	}

	// Clone so we don't mutate the caller's request.
	req2 := req.Clone(req.Context())
	req2.Header.Set("Authorization", "Bearer "+t.token)

	return t.base.RoundTrip(req2)
}

func NewClient(token string) *http.Client {
	if token == "" {
		token = os.Getenv("BOMHUB_TOKEN")
	}
	if token == "" {
		return http.DefaultClient
	}

	return &http.Client{
		Transport: &authTransport{
			base:  http.DefaultTransport,
			token: token,
		},
	}
}
