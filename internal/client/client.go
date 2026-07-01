// Package client use resty to make http request
package client

import (
	"fmt"
	"net/http"
	"strings"

	"resty.dev/v3"
)

var client = resty.New()

func init() {
	// client.SetTLSClientConfig(&tls.Config{
	// 	InsecureSkipVerify: true,
	// })
}

func CookiesToString(cookies []*http.Cookie) string {
	var parts []string
	for _, c := range cookies {
		parts = append(parts, fmt.Sprintf("%s=%s", c.Name, c.Value))
	}
	return strings.Join(parts, "; ")
}
