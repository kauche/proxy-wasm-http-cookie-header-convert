package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	echoserverpb "github.com/110y/echoserver/echoserver/api/v1"
)

func TestE2E(t *testing.T) {
	t.Parallel()

	t.Run("with token_cache_duration and original_authorization_propagation_header", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		for name, test := range map[string]struct {
			host        string
			cookies     []*http.Cookie
			wantHeaders map[string]string
		}{
			"should populate a header with a given prefix": {
				host: "upstream-1",
				cookies: []*http.Cookie{
					{
						Name:  "access_token",
						Value: "87cc5f79-35f4-46f5-b482-d3b1a52d6c98",
					},
				},
				wantHeaders: map[string]string{
					"authorization": "bearer 87cc5f79-35f4-46f5-b482-d3b1a52d6c98",
				},
			},
			"should populate headers if a request has multiple cookies": {
				host: "upstream-2",
				cookies: []*http.Cookie{
					{
						Name:  "cookie1",
						Value: "cooki-val-1",
					},
					{
						Name:  "cookie2",
						Value: "cooki-val-2",
					},
				},
				wantHeaders: map[string]string{
					"cookie-1": "cooki-val-1",
					"cookie-2": "cooki-val-2",
				},
			},
		} {
			test := test
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				req, err := createHTTPRequest(ctx, test.host)
				if err != nil {
					t.Errorf("failed to create a http request: %s", err)
					return
				}

				for _, c := range test.cookies {
					req.AddCookie(c)
				}

				res, err := http.DefaultClient.Do(req)
				if err != nil {
					t.Errorf("failed to send the http request: %s", err)
					return
				}
				defer res.Body.Close()

				if res.StatusCode != 200 {
					t.Errorf("invalid http status code for the http request: %s", res.Status)
					return
				}

				echores := new(echoserverpb.EchoResponse)
				if err = json.NewDecoder(res.Body).Decode(echores); err != nil {
					t.Errorf("failed to marshal the response to json: %s", err)
					return
				}

				for wantKey, wantVal := range test.wantHeaders {
					actualVals, ok := echores.Headers[wantKey]
					if !ok {
						t.Errorf("expected header not found: %s", err)
						return
					}

					if len(actualVals.Value) != 1 {
						t.Errorf("expected header has unexpected values: %s", err)
						return
					}

					if actualVals.Value[0] != wantVal {
						t.Errorf("want %s, but got %s", wantVal, actualVals.Value[0])
						return
					}
				}
			})
		}
	})
}

func createHTTPRequest(ctx context.Context, host string) (*http.Request, error) {
	addr := os.Getenv("ENVOY_ADDRESS")
	if addr == "" {
		addr = "localhost:8080"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s/", addr), bytes.NewBuffer([]byte(`{"message":"hello"}`)))
	if err != nil {
		return nil, err
	}

	req.Host = host
	req.Header.Set("content-type", "application/json")

	return req, nil
}
