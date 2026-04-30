// Package client use resty to make http request
package client

import (
	"resty.dev/v3"
)

var client = resty.New()
