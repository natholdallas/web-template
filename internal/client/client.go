// Package client use resty to make http request
package client

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"resty.dev/v3"

	"webgx/internal/conf"
)

var (
	httpOnce   sync.Once
	httpClient *resty.Client
)

// requestClient returns the shared resty client, building it on first use so
// the process never reads the configuration during package initialization.
func requestClient() *resty.Client {
	httpOnce.Do(func() {
		httpClient = resty.New()
		httpClient.SetTLSClientConfig(&tls.Config{
			InsecureSkipVerify: conf.App.RestyInsecureSkipVerify,
		})
		// Without a timeout a stalled Google/WeChat/rate endpoint hangs the
		// caller (and its request) forever.
		httpClient.SetTimeout(10 * time.Second)
	})
	return httpClient
}

func CookiesToString(cookies []*http.Cookie) string {
	var parts []string
	for _, c := range cookies {
		parts = append(parts, fmt.Sprintf("%s=%s", c.Name, c.Value))
	}
	return strings.Join(parts, "; ")
}
